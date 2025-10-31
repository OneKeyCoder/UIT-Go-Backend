package repository

import (
	"context"
	"database/sql"
	"time"
	"trip-service/internal/models"
)

const dbTimeout = time.Second * 3

type PostgresDBRepo struct {
	DB *sql.DB
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) PingContext(ctx context.Context) error {
	return m.DB.PingContext(ctx)
}

func (m *PostgresDBRepo) CreateTrip(tripDTO NewTripDTO, distance float64, fare float64) (models.Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `insert into trips (passenger_id, origin_lat, origin_lng, dest_lat, dest_lng, status, distance, fare, 
				payment_method, created_at, updated_at) values
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id, passenger_id, origin_lat, origin_lng, dest_lat, dest_lng, status,
				distance, fare, payment_method`
	var trip models.Trip
	err := m.DB.QueryRowContext(ctx, query,
		tripDTO.PassengerID,
		tripDTO.OriginLat,
		tripDTO.OriginLng,
		tripDTO.DestLat,
		tripDTO.DestLng,
		models.StatusRequested,
		distance,
		fare,
		tripDTO.PaymentMethod,
		time.Now(),
		time.Now(),
	).Scan(
		&trip.ID,
		&trip.PassengerID,
		&trip.OriginLat,
		&trip.OriginLng,
		&trip.DestLat,
		&trip.DestLng,
		&trip.Status,
		&trip.Distance,
		&trip.Fare,
		&trip.PaymentMethod,
	)
	if err != nil {
		return trip, err
	}
	return trip, nil
}

func (m *PostgresDBRepo) AcceptTrip(tripID int, driverID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update trips set driver_id = $1, status = $2, updated_at = $3 where id = $4`

	_, err := m.DB.ExecContext(ctx, query,
		driverID,
		models.StatusAccepted,
		time.Now(),
		tripID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresDBRepo) GetTrip(tripID int) (models.Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, passenger_id, driver_id, origin_lat, origin_lng, dest_lat, dest_lng, status,
		distance, fare, payment_method, rating, review, created_at, updated_at, started_at, completed_at, cancelled_at, cancel_by_user_id
		from trips where id = $1`

	var trip models.Trip
	err := m.DB.QueryRowContext(ctx, query, tripID).Scan(
		&trip.ID,
		&trip.PassengerID,
		&trip.DriverID,
		&trip.OriginLat,
		&trip.OriginLng,
		&trip.DestLat,
		&trip.DestLng,
		&trip.Status,
		&trip.Distance,
		&trip.Fare,
		&trip.PaymentMethod,
		&trip.Rating,
		&trip.Review,
		&trip.CreatedAt,
		&trip.UpdatedAt,
		&trip.StartedAt,
		&trip.CompletedAt,
		&trip.CancelledAt,
		&trip.CancelByUserID,
	)
	if err != nil {
		return trip, err
	}

	return trip, nil
}

func (m *PostgresDBRepo) GetTrips(page int, limit int) ([]models.Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	offset := (page - 1) * limit
	rows, err := m.DB.QueryContext(ctx, `
	SELECT id, passenger_id, driver_id, origin_lat, origin_lng, dest_lat, dest_lng, status,
	       distance, fare, payment_method, rating, review, created_at, updated_at, started_at, 
	       completed_at, cancelled_at, cancel_by_user_id
	FROM trips
	LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return []models.Trip{}, err
	}
	defer rows.Close()

	var trips = []models.Trip{}
	for rows.Next() {
		var trip models.Trip
		if err = rows.Scan(
			&trip.ID,
			&trip.PassengerID,
			&trip.DriverID,
			&trip.OriginLat,
			&trip.OriginLng,
			&trip.DestLat,
			&trip.DestLng,
			&trip.Status,
			&trip.Distance,
			&trip.Fare,
			&trip.PaymentMethod,
			&trip.Rating,
			&trip.Review,
			&trip.CreatedAt,
			&trip.UpdatedAt,
			&trip.StartedAt,
			&trip.CompletedAt,
			&trip.CancelledAt,
			&trip.CancelByUserID,
		); err != nil {
			return []models.Trip{}, err
		}
		trips = append(trips, trip)
	}
	return trips, nil
}

func (m *PostgresDBRepo) UpdateTripStatus(status models.TripStatus, tripID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update trips set status = $1, updated_at = $2 where id = $3`
	_, err := m.DB.ExecContext(ctx, query,
		status,
		time.Now(),
		tripID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) GetTripsByPassenger(passengerID int) ([]models.Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, passenger_id, driver_id, origin_lat, origin_lng, dest_lat, dest_lng, status,
		distance, fare, payment_method, rating, review, created_at, updated_at, started_at, completed_at, cancelled_at, cancel_by_user_id
		from trips where passenger_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, passengerID)
	if err != nil {
		return nil, err
	}
	trips := []models.Trip{}
	for rows.Next() {
		var trip models.Trip
		if err = rows.Scan(
			&trip.ID,
			&trip.PassengerID,
			&trip.DriverID,
			&trip.OriginLat,
			&trip.OriginLng,
			&trip.DestLat,
			&trip.DestLng,
			&trip.Status,
			&trip.Distance,
			&trip.Fare,
			&trip.PaymentMethod,
			&trip.Rating,
			&trip.Review,
			&trip.CreatedAt,
			&trip.UpdatedAt,
			&trip.StartedAt,
			&trip.CompletedAt,
			&trip.CancelledAt,
			&trip.CancelByUserID,
		); err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	defer rows.Close()
	return trips, nil
}
func (m *PostgresDBRepo) GetTripsByDriver(driverID int) ([]models.Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, passenger_id, driver_id, origin_lat, origin_lng, dest_lat, dest_lng, status,
		distance, fare, payment_method, rating, review, created_at, updated_at, started_at, completed_at, cancelled_at, cancel_by_user_id
		from trips where driver_id = $1`
	rows, err := m.DB.QueryContext(ctx, query, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	trips := []models.Trip{}
	for rows.Next() {
		var trip models.Trip
		if err = rows.Scan(
			&trip.ID,
			&trip.PassengerID,
			&trip.DriverID,
			&trip.OriginLat,
			&trip.OriginLng,
			&trip.DestLat,
			&trip.DestLng,
			&trip.Status,
			&trip.Distance,
			&trip.Fare,
			&trip.PaymentMethod,
			&trip.Rating,
			&trip.Review,
			&trip.CreatedAt,
			&trip.UpdatedAt,
			&trip.StartedAt,
			&trip.CompletedAt,
			&trip.CancelledAt,
			&trip.CancelByUserID,
		); err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	defer rows.Close()
	return trips, nil
}
func (m *PostgresDBRepo) CancelTrip(userID int, tripID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update trips set status = $1, cancel_by_user_id = $2, updated_at = $3, cancelled_at = $4 where id = $5`
	_, err := m.DB.ExecContext(ctx, query,
		models.StatusCancelled,
		userID,
		time.Now(),
		time.Now(),
		tripID,
	)
	if err != nil {
		return err
	}
	return nil
}
func (m *PostgresDBRepo) ReviewTrip(tripID int, review ReviewDTO) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update trips set rating = $1, review = $2, updated_at = $3 where id = $4`
	_, err := m.DB.ExecContext(ctx, query,
		review.Rating,
		review.Comment,
		time.Now(),
		tripID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) GetReview(tripID int) (ReviewDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select passenger_id, rating, review from trips where id = $1`
	var review ReviewDTO
	err := m.DB.QueryRowContext(ctx, query, tripID).Scan(
		&review.PassengerID,
		&review.Rating,
		&review.Comment,
	)
	if err != nil {
		return ReviewDTO{}, err
	}
	return review, nil
}
