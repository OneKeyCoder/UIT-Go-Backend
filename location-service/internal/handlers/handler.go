package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	service "location-service/internal"
)

type Handlers struct {
	ctx *context.Context
	ser *service.LocationService
}

func NewHandlers(ctx *context.Context, ser *service.LocationService) *Handlers {
	return &Handlers{
		ctx: ctx,
		ser: ser,
	}
}

// SetCurrentLocation handles POST requests to set a user's current location
func (h *Handlers) SetCurrentLocation(w http.ResponseWriter, r *http.Request) {
	var location service.CurrentLocation

	err := h.readJSON(w, r, &location)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if location.UserID == "" {
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

	location, err := h.ser.GetCurrentLocation(*h.ctx, userID)
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
	radius := 10.0 // default value in km
	if radiusStr != "" {
		var err error
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil || radius <= 0 {
			h.errorJSON(w, errors.New("radius must be a positive number"), http.StatusBadRequest)
			return
		}
	}

	locations, err := h.ser.FindTopNearestUsers(*h.ctx, userID, topN, radius)
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

// Helper methods

func (h *Handlers) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handlers) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var response ResponseDTO
	response.StatusCode = statusCode
	response.Message = err.Error()
	response.Data = nil

	return h.writeJSON(w, statusCode, response)
}
