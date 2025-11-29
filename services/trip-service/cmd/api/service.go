package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	"trip-service/internal"
	"trip-service/internal/models"
	"trip-service/internal/repository"

	"github.com/Azure/go-amqp"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/env"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/rabbitmq"
)

type TripService struct {
	DB          repository.DatabaseRepo
	grpcClients *GRPCClients
	RabbitConn  *amqp.Conn
}

var tripMap = make(map[int][]int)

func (trip *TripService) CreateTrip(newTrip repository.NewTripDTO) (models.Trip, float64, error) {
	origin := fmt.Sprintf("%f,%f", newTrip.OriginLat, newTrip.OriginLng)
	destination := fmt.Sprintf("%f,%f", newTrip.DestLat, newTrip.DestLng)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	routeSummary, err := internal.GetRouteSummary(ctx, origin, destination)
	if err != nil {
		logger.Error("Failed to get route summary from HERE API", "error", err)
		return models.Trip{}, 0, err
	}
	logger.Info("Route Summary", "route", routeSummary)
	tripRecord, err := trip.DB.CreateTrip(newTrip, routeSummary.Distance, routeSummary.Fare)
	if err != nil {
		logger.Error("Failed to create trip in database", "error", err)
		return models.Trip{}, 0, err
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("User %d requested a trip from (%f, %f) to (%f, %f)",
			newTrip.PassengerID,
			newTrip.OriginLat,
			newTrip.OriginLng,
			newTrip.DestLat,
			newTrip.DestLng,
		)
		go PublishEvent(trip.RabbitConn, "user.createTrip", eventData)
	}
	err = trip.getAllAvailableDrivers(tripRecord.ID, tripRecord.PassengerID)
	if err != nil {
		logger.Error("Failed to get available drivers", "error", err)
		return tripRecord, routeSummary.Duration, err
	}
	return tripRecord, routeSummary.Duration, nil
}
func (trip *TripService) AcceptTrip(driverID int, tripID int) error {
	if _, ok := tripMap[tripID]; !ok || len(tripMap[tripID]) == 0 {
		logger.Error("No available drivers for this trip", "trip_id", tripID)
		return errors.New("no available drivers for this trip")
	}
	if suggestID, err := trip.GetSuggestedDriver(tripID); err != nil || suggestID != driverID {
		logger.Error("Driver is not the suggested driver for this trip", "driver_id", driverID, "trip_id", tripID)
		return errors.New("driver is not the suggested driver for this trip")
	}
	err := trip.DB.AcceptTrip(tripID, driverID)
	if err != nil {
		logger.Error("Failed to accept trip in database", "error", err)
		return err
	}
	tripRecord, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return err
	}
	//Thông báo
	delete(tripMap, tripRecord.PassengerID)
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("Driver %d accepted trip %d", driverID, tripID)
		go PublishEvent(trip.RabbitConn, "driver.acceptTrip", eventData)
	}
	return nil
}
func (trip *TripService) GetTrip(userID int, tripID int) (models.Trip, error) {
	tripRecord, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return models.Trip{}, err
	}
	logger.Info("GetTrip", "user_id", userID, "trip_id", tripRecord.PassengerID)
	var driver int
	if tripRecord.DriverID.Valid {
		driver = int(tripRecord.DriverID.Int32)
	} else {
		driver = 0
	}
	if tripRecord.PassengerID != userID && driver != userID {
		logger.Error("User is not authorized to view this trip", "user_id", userID, "trip_id", tripID)
		return models.Trip{}, errors.New("user is not authorized to view this trip")
	}
	return tripRecord, nil
}

func (trip *TripService) GetTripsByPassenger(passengerID int) ([]models.Trip, error) {
	trips, err := trip.DB.GetTripsByPassenger(passengerID)
	if err != nil {
		logger.Error("Failed to get trips by passenger from database", "error", err)
		return nil, err
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("User %d requested their trip history", passengerID)
		go PublishEvent(trip.RabbitConn, "user.tripHistory", eventData)
	}
	return trips, nil
}
func (trip *TripService) GetTripsByDriver(driverID int) ([]models.Trip, error) {
	trips, err := trip.DB.GetTripsByDriver(driverID)
	if err != nil {
		logger.Error("Failed to get trips by driver from database", "error", err)
		return nil, err
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("Driver %d requested their trip history", driverID)
		go PublishEvent(trip.RabbitConn, "driver.tripHistory", eventData)
	}
	return trips, nil
}

func (trip *TripService) UpdateTripStatus(status models.TripStatus, tripID int, driverID int) error {
	tripRecord, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return err
	}
	var driver int
	if tripRecord.DriverID.Valid {
		driver = int(tripRecord.DriverID.Int32)
	} else {
		driver = 0
	}
	if driver != driverID {
		logger.Error("Driver is not authorized to update this trip", "driver_id", driverID, "trip_id", tripID)
		return errors.New("driver is not authorized to update this trip")
	}
	err = trip.DB.UpdateTripStatus(status, tripID)
	if err != nil {
		logger.Error("Failed to update trip status in database", "error", err)
		return err
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("Trip %d status updated to %s", tripID, status)
		go PublishEvent(trip.RabbitConn, "trip.updateStatus", eventData)
	}
	return nil
}

func (trip *TripService) GetAllTrips(page int, limit int) ([]models.Trip, error) {
	trips, err := trip.DB.GetTrips(page, limit)
	if err != nil {
		logger.Error("Failed to get app trips from database", "error", err)
		return nil, err
	}
	return trips, nil
}
func (trip *TripService) ReviewTrip(userID int, tripID int, review repository.ReviewDTO) error {
	record, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return errors.New("trip not found")
	}
	if record.PassengerID != userID {
		logger.Error("User is not authorized to review this trip", "user_id", userID, "trip_id", tripID)
		return errors.New("user is not authorized to review this trip")
	}
	err = trip.DB.ReviewTrip(tripID, review)
	if err != nil {
		logger.Error("Failed to review trip in database", "error", err)
		return err
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("User %d reviewed trip %d", userID, tripID)
		go PublishEvent(trip.RabbitConn, "user.tripHistory", eventData)
	}
	return nil
}

func (trip *TripService) GetReview(tripID int, userID int) (repository.ReviewDTO, error) {
	review, err := trip.DB.GetReview(tripID)
	if err != nil {
		logger.Error("Failed to get review from database", "error", err)
		return repository.ReviewDTO{}, err
	}
	if review.PassengerID != userID {
		logger.Error("User is not authorized to view this review", "user_id", userID, "trip_id", tripID)
		return repository.ReviewDTO{}, errors.New("user is not authorized to view this review")
	}
	return review, nil
}

func (trip *TripService) CancelTrip(userID int, tripID int) error {
	err := trip.DB.CancelTrip(userID, tripID)
	if err != nil {
		logger.Error("Failed to cancel trip in database", "error", err)
		return err
	}
	return nil
}

func (trip *TripService) getAllAvailableDrivers(tripID int, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	radiusList := []float64{5000.0, 10000.0, 15000.0}

	for _, radius := range radiusList {
		locations, err := trip.grpcClients.FindNearestUsersViaGRPC(ctx, userID, 5, radius)
		if err != nil {
			logger.Error("Failed to get nearest users via gRPC", "radius", radius, "error", err)
			continue
		}

		if len(locations.Locations) > 0 {
			seen := make(map[int]bool)
			for _, loc := range locations.Locations {
				if !seen[int(loc.UserId)] {
					tripMap[tripID] = append(tripMap[tripID], int(loc.UserId))
					seen[int(loc.UserId)] = true
				}
			}

			logger.Info("Found nearby drivers",
				"user_id", userID,
				"radius", radius,
				"found", len(tripMap[tripID]),
			)
			return nil
		}
	}
	logger.Warn("No drivers found within 15 km", "user_id", userID)
	return nil
}
func (trip *TripService) GetSuggestedDriver(tripID int) (int, error) {
	tripRecord, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return 0, err
	}
	if tripRecord.Status != models.StatusRequested {
		logger.Error("Trip is not in requested status", "trip_id", tripID, "status", string(tripRecord.Status))
		return 0, errors.New("trip is not in requested status")
	}
	if _, ok := tripMap[tripID]; !ok || len(tripMap[tripID]) == 0 {
		err := trip.getAllAvailableDrivers(tripID, tripID)
		if err != nil {
			return 0, err
		}
	}
	if len(tripMap[tripID]) == 0 {
		return 0, errors.New("no available drivers found")
	}
	suggestedDriverID := tripMap[tripID][0]
	return suggestedDriverID, nil
}

func (trip *TripService) RejectTrip(driverID int, passengerID int, tripID int) error {
	tripRecord, err := trip.DB.GetTrip(tripID)
	if err != nil {
		logger.Error("Failed to get trip from database", "error", err)
		return err
	}
	var driver int
	if tripRecord.DriverID.Valid {
		driver = int(tripRecord.DriverID.Int32)
	} else {
		driver = 0
	}
	if driver != driverID {
		logger.Error("Driver is not authorized to reject this trip", "driver_id", driverID, "trip_id", tripID)
		return errors.New("driver is not authorized to reject this trip")
	}
	if tripRecord.PassengerID != passengerID {
		logger.Error("Passenger ID does not match trip record", "passenger_id", passengerID, "trip_id", tripID)
		return errors.New("passenger ID does not match trip record")
	}
	//Thông báo
	if len(tripMap[passengerID]) > 0 {
		tripMap[passengerID] = tripMap[passengerID][1:]
	} else {
		logger.Warn("No more drivers available", "passenger_id", passengerID)
		return errors.New("no more drivers available")
	}
	if trip.RabbitConn != nil {
		eventData := fmt.Sprintf("Driver %d rejected trip %d", driverID, tripID)
		go PublishEvent(trip.RabbitConn, "driver.rejectTrip", eventData)
	}
	return nil
}
func (trip *TripService) InitializeServices() {
	conn, err := trip.connectToDB()
	if err != nil {
		logger.Fatal("Cannot connect to database", "error", err)
	}
	logger.Info("Connected to database")
	trip.grpcClients = &GRPCClients{}
	if trip.grpcClients, err = trip.grpcClients.InitGRPCClients(); err != nil {
		logger.Fatal("Cannot initialize gRPC clients", "error", err)
	}
	trip.DB = &repository.PostgresDBRepo{
		DB: conn,
	}
	rabbitConn, err := rabbitmq.ConnectSimple(env.RabbitMQURL())
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ, continuing without events", "error", err)
	} else {
		logger.Info("Connected to RabbitMQ")
	}
	trip.RabbitConn = rabbitConn
}

func (trip *TripService) connectToDB() (*sql.DB, error) {
	dsn := os.Getenv("DSN")
	connection, err := openDB(dsn)
	if err != nil {
		logger.Error("Cannot connect to database", "error", err)
		return nil, err
	}
	return connection, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
