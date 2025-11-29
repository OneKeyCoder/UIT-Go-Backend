package main

import (
	"net/http"
	"strconv"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	"github.com/go-chi/chi/v5"
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

type CreateTripRequest struct {
	OriginLat     float64 `json:"origin_lat" validate:"required"`
	OriginLng     float64 `json:"origin_lng" validate:"required"`
	DestLat       float64 `json:"dest_lat" validate:"required"`
	DestLng       float64 `json:"dest_lng" validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash card"`
}

type UpdateTripStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACCEPTED STARTED COMPLETED CANCELLED"`
}

type ReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment,omitempty"`
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
	resp, err := app.SetLocationViaGRPC(ctx, userID, claims.Role, loc.Latitude, loc.Longitude, loc.Speed, loc.Heading, loc.Timestamp)
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

func (app *Config) CreateTrip(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "CreateTrip")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	var tripReq CreateTripRequest
	err = request.ReadAndValidate(w, r, &tripReq)
	if request.HandleError(w, err) {
		response.BadRequest(w, "Invalid request payload: "+err.Error())
		return
	}
	resp, err := app.CreateTripViaGRPC(ctx,
		int(claims.UserID),
		tripReq.OriginLat,
		tripReq.OriginLng,
		tripReq.DestLat,
		tripReq.DestLng,
		tripReq.PaymentMethod,
	)
	if err != nil {
		response.InternalServerError(w, "Failed to create trip: "+err.Error())
		return
	}
	response.Success(w, "Trip created successfully", resp)
}

func (app *Config) AcceptTrip(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "AcceptTrip")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}
	resp, err := app.AcceptTripViaGRPC(ctx, int(claims.UserID), tripIDInt)
	if err != nil || !resp.Success {
		response.InternalServerError(w, "Failed to accept trip: "+err.Error())
		return
	}
	response.Success(w, "Trip accepted successfully", nil)
}

func (app *Config) RejectTrip(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "RejectTrip")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}
	resp, err := app.RejectTripViaGRPC(ctx, int(claims.UserID), tripIDInt)
	if err != nil || !resp.Success {
		response.InternalServerError(w, "Failed to reject trip: "+err.Error())
		return
	}
	response.Success(w, "Trip rejected successfully", nil)
}

func (app *Config) GetSuggestedDriver(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetSuggestedDriver")
	defer span.End()

	_, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	logger.Info("Trip ID:", "tripId from URL", tripID)
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	resp, err := app.GetSuggestedDriverViaGRPC(ctx, tripIDInt)
	if err != nil {
		response.InternalServerError(w, "Failed to get suggested driver: "+err.Error())
		return
	}
	response.Success(w, "Suggested driver retrieved successfully", resp)
}

func (app *Config) GetTripDetails(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetTripDetails")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}
	logger.Info("Request URL", "URL", r.URL.Path)
	for key := range chi.RouteContext(r.Context()).URLParams.Keys {
		logger.Info("Param key:", "key", chi.RouteContext(r.Context()).URLParams.Keys[key])
	}
	tripID := chi.URLParam(r, "tripID")
	logger.Info("Trip ID:", "tripId from URL", tripID)
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	resp, err := app.GetTripDetailViaGRPC(ctx, tripIDInt, int(claims.UserID))
	if err != nil {
		response.InternalServerError(w, "Failed to get trip details: "+err.Error())
		return
	}
	response.Success(w, "Trip details retrieved successfully", resp)
}

func (app *Config) GetTripsByPassenger(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetTripsByPassenger")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	resp, err := app.GetTripsByPassengerViaGRPC(ctx, int(claims.UserID))
	if err != nil {
		response.InternalServerError(w, "Failed to get trips by passenger: "+err.Error())
		return
	}
	response.Success(w, "Trips by passenger retrieved successfully", resp)
}

func (app *Config) GetTripsByDriver(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetTripsByDriver")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	resp, err := app.GetTripsByDriverViaGRPC(ctx, int(claims.UserID))
	if err != nil {
		response.InternalServerError(w, "Failed to get trips by driver: "+err.Error())
		return
	}
	response.Success(w, "Trips by driver retrieved successfully", resp)
}

func (app *Config) GetAllTrips(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetAllTrips")
	defer span.End()

	limit := r.URL.Query().Get("limit")
	page := r.URL.Query().Get("page")
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 10
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		pageInt = 1
	}
	_, err = app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}
	resp, err := app.GetAllTripsViaGRPC(ctx, pageInt, limitInt)
	if err != nil {
		response.InternalServerError(w, "Failed to get all trips: "+err.Error())
		return
	}
	response.Success(w, "All trips retrieved successfully", resp)
}

func (app *Config) UpdateTripStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "UpdateTripStatus")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	var statusReq UpdateTripStatusRequest
	err = request.ReadAndValidate(w, r, &statusReq)
	if request.HandleError(w, err) {
		response.BadRequest(w, "Invalid request payload: "+err.Error())
		return
	}

	resp, err := app.UpdateTripStatusViaGRPC(ctx,
		tripIDInt,
		int(claims.UserID),
		statusReq.Status,
	)
	if err != nil || !resp.Success {
		response.InternalServerError(w, "Failed to update trip status: "+err.Error())
		return
	}
	response.Success(w, "Trip status updated successfully", nil)
}

func (app *Config) CancelTrip(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "CancelTrip")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	resp, err := app.CancelTripViaGRPC(ctx, tripIDInt, int(claims.UserID))
	if err != nil || !resp.Success {
		response.InternalServerError(w, "Failed to cancel trip: "+err.Error())
		return
	}
	response.Success(w, "Trip cancelled successfully", nil)
}

func (app *Config) SubmitReview(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "SubmitReview")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	var reviewReq ReviewRequest
	err = request.ReadAndValidate(w, r, &reviewReq)
	if request.HandleError(w, err) {
		response.BadRequest(w, "Invalid request payload: "+err.Error())
		return
	}

	resp, err := app.SubmitReviewViaGRPC(ctx,
		tripIDInt,
		int(claims.UserID),
		reviewReq.Rating,
		reviewReq.Comment,
	)
	if err != nil || !resp.Success {
		response.InternalServerError(w, "Failed to submit review: "+err.Error())
		return
	}
	response.Success(w, "Review submitted successfully", nil)
}
func (app *Config) GetTripReview(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.StartSpan(r.Context(), "GetTripReview")
	defer span.End()

	claims, err := app.GetClaims(r.Context())
	if err != nil {
		response.Unauthorized(w, "Unauthorized: "+err.Error())
		return
	}

	tripID := chi.URLParam(r, "tripID")
	if tripID == "" {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	tripIDInt, err := strconv.Atoi(tripID)
	if err != nil {
		response.BadRequest(w, "Trip ID must be an integer")
		return
	}

	resp, err := app.GetReviewViaGRPC(ctx, tripIDInt, int(claims.UserID))
	if err != nil {
		response.InternalServerError(w, "Failed to get trip review: "+err.Error())
		return
	}
	response.Success(w, "Trip review retrieved successfully", resp)
}
