package models

import (
	"database/sql"
	"time"
)

type Trip struct {
	ID             int            `json:"id"`
	PassengerID    int            `json:"passenger_id"`
	DriverID       sql.NullInt32  `json:"driver_id"`
	OriginLat      float64        `json:"origin_lat"`
	OriginLng      float64        `json:"origin_lng"`
	DestLat        float64        `json:"dest_lat"`
	DestLng        float64        `json:"dest_lng"`
	Status         TripStatus     `json:"status"`
	Distance       float64        `json:"distance"`
	Fare           float64        `json:"fare"`
	PaymentMethod  string         `json:"payment_method"`
	Rating         sql.NullInt32  `json:"rating,omitempty"`
	Review         sql.NullString `json:"review,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	StartedAt      sql.NullTime   `json:"started_at"`
	CompletedAt    sql.NullTime   `json:"completed_at"`
	CancelledAt    sql.NullTime   `json:"cancelled_at"`
	CancelByUserID sql.NullInt64  `json:"cancel_by_user_id,omitempty"`
}

type TripStatus string

const (
	StatusRequested TripStatus = "REQUESTED"
	StatusAccepted  TripStatus = "ACCEPTED"
	StatusStarted   TripStatus = "STARTED"
	StatusCompleted TripStatus = "COMPLETED"
	StatusCancelled TripStatus = "CANCELLED"
)
