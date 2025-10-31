package repository

type NewTripDTO struct {
	PassengerID   int     `json:"passenger_id"`
	OriginLat     float64 `json:"origin_lat"`
	OriginLng     float64 `json:"origin_lng"`
	DestLat       float64 `json:"dest_lat"`
	DestLng       float64 `json:"dest_lng"`
	PaymentMethod string  `json:"payment_method"`
}

type ReviewDTO struct {
	PassengerID int    `json:"passenger_id"`
	Comment     string `json:"comment,omitempty"`
	Rating      int    `json:"rating"`
}
