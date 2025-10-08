package location_service

type CurrentLocation struct {
	UserID    string  `json:"user_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance,omitempty"`
	Speed     float64 `json:"speed"`
	Heading   string  `json:"heading"`
	Timestamp string  `json:"timestamp"`
}
