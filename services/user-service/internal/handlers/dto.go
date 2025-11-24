package handlers

type ResponseDTO struct {
	StatusCode int
	Message    string
	Data       interface{}
}

type UserRequest struct {
	Email        string `json:"email"`
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
