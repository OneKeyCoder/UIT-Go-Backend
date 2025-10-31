package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	location_service "location-service/internal"
)

type Handlers struct {
	ctx *context.Context
	ser *location_service.LocationService
}

func NewHandlers(ctx *context.Context, ser *location_service.LocationService) *Handlers {
	return &Handlers{
		ctx: ctx,
		ser: ser,
	}
}

// SetCurrentLocation handles POST requests to set a user's current location
func (h *Handlers) SetCurrentLocation(w http.ResponseWriter, r *http.Request) {
	var location location_service.CurrentLocation

	err := h.readJSON(w, r, &location)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if location.UserID == 0 {
		h.errorJSON(w, errors.New("user_id is required"), http.StatusBadRequest)
		return
	}

	err = h.ser.SetCurrentLocation(*h.ctx, &location)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Location updated successfully",
		Data:       location,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCurrentLocation handles GET requests to retrieve a user's current location
func (h *Handlers) GetCurrentLocation(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.errorJSON(w, errors.New("user_id query parameter is required"), http.StatusBadRequest)
		return
	}
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}
	location, err := h.ser.GetCurrentLocation(*h.ctx, tempUserID)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if location == nil {
		h.errorJSON(w, errors.New("location not found"), http.StatusNotFound)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Location retrieved successfully",
		Data:       location,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// FindNearestUsers handles GET requests to find the nearest users to a given user
func (h *Handlers) FindNearestUsers(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.errorJSON(w, errors.New("user_id query parameter is required"), http.StatusBadRequest)
		return
	}

	topNStr := r.URL.Query().Get("top_n")
	topN := 10 // default value
	if topNStr != "" {
		var err error
		topN, err = strconv.Atoi(topNStr)
		if err != nil || topN <= 0 {
			h.errorJSON(w, errors.New("top_n must be a positive integer"), http.StatusBadRequest)
			return
		}
	}

	radiusStr := r.URL.Query().Get("radius")
	radius := 5.0 // default value in km
	if radiusStr != "" {
		var err error
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil || radius <= 0 {
			h.errorJSON(w, errors.New("radius must be a positive number"), http.StatusBadRequest)
			return
		}
	}
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}
	locations, err := h.ser.FindTopNearestUsers(*h.ctx, tempUserID, topN, radius)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Nearest users retrieved successfully",
		Data:       locations,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handlers) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.ser.GetAllLocations(*h.ctx)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "All locations retrieved successfully",
		Data:       locations,
	}

	h.writeJSON(w, http.StatusOK, response)
}
