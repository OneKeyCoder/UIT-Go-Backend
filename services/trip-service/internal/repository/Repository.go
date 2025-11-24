package repository

import (
	"context"
	"database/sql"
	"trip-service/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	PingContext(ctx context.Context) error
	CreateTrip(tripDTO NewTripDTO, distance float64, fare float64) (models.Trip, error)
	AcceptTrip(tripID int, driverID int) error
	GetTrip(tripID int) (models.Trip, error)
	GetTripsByPassenger(passengerID int) ([]models.Trip, error)
	GetTripsByDriver(driverID int) ([]models.Trip, error)
	UpdateTripStatus(status models.TripStatus, tripID int) error
	GetTrips(page int, limit int) ([]models.Trip, error)
	CancelTrip(userID int, tripID int) error
	ReviewTrip(tripID int, review ReviewDTO) error
	GetReview(tripID int) (ReviewDTO, error)
}
