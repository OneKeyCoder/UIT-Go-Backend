package main

import (
	"net/http"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
)

type RequestPayload struct {
	Action   string           `json:"action" validate:"required"`
	Auth     *AuthPayload     `json:"auth,omitempty"`
	Log      *LogPayload      `json:"log,omitempty"`
	Location *LocationPayload `json:"location,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogPayload struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" validate:"required"`
}

type LocationPayload struct {
	Action    string  `json:"action" validate:"required"` // set, get, nearest, all
	UserID    string  `json:"user_id" validate:"required"`
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

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "HandleSubmission")
	defer span.End()

	var requestPayload RequestPayload

	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	switch requestPayload.Action {
	case "auth":
		if requestPayload.Auth == nil {
			response.BadRequest(w, "auth payload is required for auth action")
			return
		}
		// Validate the auth payload
		if err := request.Validate(requestPayload.Auth); err != nil {
			request.HandleError(w, err)
			return
		}
		app.authenticateViaGRPC(w, r.WithContext(ctx), *requestPayload.Auth)
	case "log":
		if requestPayload.Log == nil {
			response.BadRequest(w, "log payload is required for log action")
			return
		}
		// Validate the log payload
		if err := request.Validate(requestPayload.Log); err != nil {
			request.HandleError(w, err)
			return
		}
		app.logItemViaGRPCClient(w, r.WithContext(ctx), *requestPayload.Log)
	case "location":
		if requestPayload.Location == nil {
			response.BadRequest(w, "location payload is required for location action")
			return
		}
		// Validate the location payload
		if err := request.Validate(requestPayload.Location); err != nil {
			request.HandleError(w, err)
			return
		}
		app.handleLocationRequest(w, r.WithContext(ctx), *requestPayload.Location)
	default:
		response.BadRequest(w, "Unknown action")
	}
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

// handleLocationRequest routes location-related requests to appropriate handlers
func (app *Config) handleLocationRequest(w http.ResponseWriter, r *http.Request, loc LocationPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "handleLocationRequest")
	defer span.End()

	switch loc.Action {
	case "set":
		app.setLocationViaGRPC(w, r.WithContext(ctx), loc)
	case "get":
		app.getLocationViaGRPC(w, r.WithContext(ctx), loc)
	case "nearest":
		app.findNearestUsersViaGRPC(w, r.WithContext(ctx), loc)
	case "all":
		app.getAllLocationsViaGRPC(w, r.WithContext(ctx))
	default:
		response.BadRequest(w, "Unknown location action. Use: set, get, nearest, or all")
	}
}

// setLocationViaGRPC sets a user's location using gRPC
func (app *Config) setLocationViaGRPC(w http.ResponseWriter, r *http.Request, loc LocationPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "setLocationViaGRPC")
	defer span.End()

	resp, err := app.SetLocationViaGRPC(ctx, loc.UserID, loc.Latitude, loc.Longitude, loc.Speed, loc.Heading, loc.Timestamp)
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
func (app *Config) getLocationViaGRPC(w http.ResponseWriter, r *http.Request, loc LocationPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "getLocationViaGRPC")
	defer span.End()

	resp, err := app.GetLocationViaGRPC(ctx, loc.UserID)
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
func (app *Config) findNearestUsersViaGRPC(w http.ResponseWriter, r *http.Request, loc LocationPayload) {
	ctx, span := telemetry.StartSpan(r.Context(), "findNearestUsersViaGRPC")
	defer span.End()

	topN := loc.TopN
	if topN <= 0 {
		topN = 10
	}

	radius := loc.Radius
	if radius <= 0 {
		radius = 10.0
	}

	resp, err := app.FindNearestUsersViaGRPC(ctx, loc.UserID, topN, radius)
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
