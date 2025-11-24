package user_service

import "time"

type User struct {
	UserId           int       `json:"user_id"`
	Email            string    `json:"email"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
	Role             string    `json:"role"`
	DriverStatus     string    `json:"driver_status"`
	DriverTotalTrip  int       `json:"driver_total_trip"`
	DriverRevenue    float64   `json:"driver_revenue"`
	DriverAvgRating  float32   `json:"driver_avg_rating"`
	DriverVerifiedAt time.Time `json:"driver_verified_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Vehicle struct {
	VehicleId    int       `json:"vehicle_id"`
	DriverId     int       `json:"driver_id"`
	LicensePlate string    `json:"license_plate"`
	VehicleType  string    `json:"vehicle_type"`
	Seats        int       `json:"seats"`
	Status       string    `json:"status"`
	VerifiedAt   time.Time `json:"verified_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Request DTOs
type UserRequest struct {
	Email        string `json:"email"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Role         string `json:"role"`
	DriverStatus string `json:"driver_status"`
}

type VehicleRequest struct {
	VehicleId    int    `json:"vehicle_id"`
	DriverId     int    `json:"driver_id"`
	LicensePlate string `json:"license_plate"`
	VehicleType  string `json:"vehicle_type"`
	Seats        int    `json:"seats"`
	Status       string `json:"status"`
}
