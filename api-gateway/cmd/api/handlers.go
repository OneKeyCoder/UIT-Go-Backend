package main

import (
	"net/http"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
)

type AuthPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogPayload struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" validate:"required"`
}

type LocationPayload struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
	Heading   string  `json:"heading,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
	TopN      int32   `json:"top_n,omitempty"`
	Radius    float64 `json:"radius,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "Hit the broker", nil)
}

// authenticateViaGRPC handles authentication using gRPC client
func (app *Config) authenticateViaGRPC(w http.ResponseWriter, r *http.Request, a AuthPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "authenticateViaGRPC")
	defer span.End()

	resp, err := app.AuthenticateViaGRPC(ctx, a.Email, a.Password)
	if err != nil {
		response.Unauthorized(w, "Authentication failed")
		return
	}

	if !resp.Success {
		response.Unauthorized(w, resp.Message)
		return
	}

	// Return the response with user info and tokens
	payload := map[string]interface{}{
		"message": resp.Message,
		"user": map[string]interface{}{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"first_name": resp.User.FirstName,
			"last_name":  resp.User.LastName,
			"active":     resp.User.Active,
		},
		"tokens": map[string]interface{}{
			"access_token":  resp.Tokens.AccessToken,
			"refresh_token": resp.Tokens.RefreshToken,
		},
	}

	response.Success(w, "Authenticated successfully", payload)
}

// logItemViaGRPCClient logs using the new gRPC client
func (app *Config) logItemViaGRPCClient(w http.ResponseWriter, r *http.Request, entry LogPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "logItemViaGRPCClient")
	defer span.End()

	err := app.LogViaGRPC(ctx, entry.Name, entry.Data)
	if err != nil {
		response.InternalServerError(w, "Failed to log via gRPC")
		return
	}

	response.Success(w, "Logged via gRPC", nil)
}

// setLocationViaGRPC sets a user's location using gRPC
func (app *Config) setLocationViaGRPC(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "setLocationViaGRPC")
	defer span.End()
	var loc LocationPayload
	err := request.ReadAndValidate(w, r, &loc)
	if request.HandleError(w, err) {
		return
	}
	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}
	userID := int(claims.UserID)
	resp, err := app.SetLocationViaGRPC(ctx, userID, loc.Latitude, loc.Longitude, loc.Speed, loc.Heading, loc.Timestamp)
	if err != nil {
		response.InternalServerError(w, "Failed to set location")
		return
	}

	if !resp.Success {
		response.BadRequest(w, resp.Message)
		return
	}

	payload := map[string]interface{}{
		"message": resp.Message,
		"location": map[string]interface{}{
			"user_id":   resp.Location.UserId,
			"latitude":  resp.Location.Latitude,
			"longitude": resp.Location.Longitude,
			"speed":     resp.Location.Speed,
			"heading":   resp.Location.Heading,
			"timestamp": resp.Location.Timestamp,
		},
	}

	response.Success(w, "Location set successfully", payload)
}

// getLocationViaGRPC gets a user's location using gRPC
func (app *Config) getLocationViaGRPC(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "getLocationViaGRPC")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	resp, err := app.GetLocationViaGRPC(ctx, int(claims.UserID))
	if err != nil {
		response.InternalServerError(w, "Failed to get location")
		return
	}

	if !resp.Success {
		response.NotFound(w, resp.Message)
		return
	}

	payload := map[string]interface{}{
		"message": resp.Message,
		"location": map[string]interface{}{
			"user_id":   resp.Location.UserId,
			"latitude":  resp.Location.Latitude,
			"longitude": resp.Location.Longitude,
			"speed":     resp.Location.Speed,
			"heading":   resp.Location.Heading,
			"timestamp": resp.Location.Timestamp,
			"distance":  resp.Location.Distance,
		},
	}

	response.Success(w, "Location retrieved successfully", payload)
}

// findNearestUsersViaGRPC finds nearest users using gRPC
func (app *Config) findNearestUsersViaGRPC(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "findNearestUsersViaGRPC")
	defer span.End()
	var loc LocationPayload
	err := request.ReadAndValidate(w, r, &loc)
	if request.HandleError(w, err) {
		return
	}
	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	topN := loc.TopN
	if topN <= 0 {
		topN = 10
	}

	radius := loc.Radius
	if radius <= 0 {
		radius = 10.0
	}

	resp, err := app.FindNearestUsersViaGRPC(ctx, int(claims.UserID), topN, radius)
	if err != nil {
		response.InternalServerError(w, "Failed to find nearest users")
		return
	}

	if !resp.Success {
		response.BadRequest(w, resp.Message)
		return
	}

	locations := make([]map[string]interface{}, 0, len(resp.Locations))
	for _, l := range resp.Locations {
		locations = append(locations, map[string]interface{}{
			"user_id":   l.UserId,
			"latitude":  l.Latitude,
			"longitude": l.Longitude,
			"speed":     l.Speed,
			"heading":   l.Heading,
			"timestamp": l.Timestamp,
			"distance":  l.Distance,
		})
	}

	payload := map[string]interface{}{
		"message":   resp.Message,
		"locations": locations,
		"count":     len(locations),
	}

	response.Success(w, "Nearest users found successfully", payload)
}

// getAllLocationsViaGRPC gets all locations using gRPC
func (app *Config) getAllLocationsViaGRPC(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "getAllLocationsViaGRPC")
	defer span.End()

	resp, err := app.GetAllLocationsViaGRPC(ctx)
	if err != nil {
		response.InternalServerError(w, "Failed to get all locations")
		return
	}

	if !resp.Success {
		response.BadRequest(w, resp.Message)
		return
	}

	locations := make([]map[string]interface{}, 0, len(resp.Locations))
	for _, l := range resp.Locations {
		locations = append(locations, map[string]interface{}{
			"user_id":   l.UserId,
			"latitude":  l.Latitude,
			"longitude": l.Longitude,
			"speed":     l.Speed,
			"heading":   l.Heading,
			"timestamp": l.Timestamp,
		})
	}

	payload := map[string]interface{}{
		"message":     resp.Message,
		"locations":   locations,
		"total_count": resp.TotalCount,
	}

	response.Success(w, "All locations retrieved successfully", payload)
}
