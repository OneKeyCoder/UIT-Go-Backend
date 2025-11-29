package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	user_service "user-service/internal"
)

type Handlers struct {
	ctx *context.Context
	ser *user_service.UserService
}

func NewHandlers(ctx *context.Context, ser *user_service.UserService) *Handlers {
	return &Handlers{
		ctx: ctx,
		ser: ser,
	}
}

// ============================================
// User Handlers
// ============================================

// GetAllUsers handles GET requests to retrieve all users
func (h *Handlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.ser.GetAllUsers(*h.ctx)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Users retrieved successfully",
		Data:       users,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetUserById handles GET requests to retrieve a user by ID
func (h *Handlers) GetUserById(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("user_id is required"), http.StatusBadRequest)
		return
	}

	userID := pathParts[2]
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}

	user, err := h.ser.GetUserById(*h.ctx, tempUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "User retrieved successfully",
		Data:       user,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CreateUser handles POST requests to create a new user
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload user_service.UserRequest

	err := h.readJSON(w, r, &payload)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.Email == "" {
		h.errorJSON(w, errors.New("email is required"), http.StatusBadRequest)
		return
	}

	err = h.ser.CreateUser(*h.ctx, payload)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.errorJSON(w, err, http.StatusConflict)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusCreated,
		Message:    "User created successfully",
		Data:       nil,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// UpdateUser handles PUT requests to update a user
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("user_id is required"), http.StatusBadRequest)
		return
	}

	userID := pathParts[2]
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}

	var userRequest user_service.UserRequest
	err = h.readJSON(w, r, &userRequest)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = h.ser.UpdateUserById(*h.ctx, tempUserID, userRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "User updated successfully",
		Data:       nil,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteUser handles DELETE requests to delete a user
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("user_id is required"), http.StatusBadRequest)
		return
	}

	userID := pathParts[2]
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}

	err = h.ser.DeleteUserById(*h.ctx, tempUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "User deleted successfully",
		Data:       nil,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ============================================
// Vehicle Handlers
// ============================================

// GetAllVehicles handles GET requests to retrieve all vehicles
func (h *Handlers) GetAllVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.ser.GetAllVehicles(*h.ctx)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Vehicles retrieved successfully",
		Data:       vehicles,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetVehicleById handles GET requests to retrieve a vehicle by ID
func (h *Handlers) GetVehicleById(w http.ResponseWriter, r *http.Request) {
	// Extract vehicle ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("vehicle_id is required"), http.StatusBadRequest)
		return
	}

	vehicleID := pathParts[2]
	tempVehicleID, err := strconv.Atoi(vehicleID)
	if err != nil {
		h.errorJSON(w, errors.New("vehicle_id must be an integer"), http.StatusBadRequest)
		return
	}

	vehicle, err := h.ser.GetVehicleById(*h.ctx, tempVehicleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Vehicle retrieved successfully",
		Data:       vehicle,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetVehiclesByUserId handles GET requests to retrieve all vehicles for a specific user/driver
func (h *Handlers) GetVehiclesByUserId(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("user_id is required"), http.StatusBadRequest)
		return
	}

	userID := pathParts[2]
	tempUserID, err := strconv.Atoi(userID)
	if err != nil {
		h.errorJSON(w, errors.New("user_id must be an integer"), http.StatusBadRequest)
		return
	}

	vehicles, err := h.ser.GetVehiclesByUserId(*h.ctx, tempUserID)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Vehicles retrieved successfully",
		Data:       vehicles,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CreateVehicle handles POST requests to create a new vehicle
func (h *Handlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DriverId int                         `json:"driver_id"`
		Vehicle  user_service.VehicleRequest `json:"vehicle"`
	}

	err := h.readJSON(w, r, &payload)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.DriverId == 0 {
		h.errorJSON(w, errors.New("driver_id is required"), http.StatusBadRequest)
		return
	}
	if payload.Vehicle.LicensePlate == "" {
		h.errorJSON(w, errors.New("license_plate is required"), http.StatusBadRequest)
		return
	}
	if payload.Vehicle.VehicleType == "" {
		h.errorJSON(w, errors.New("vehicle_type is required"), http.StatusBadRequest)
		return
	}

	vehicle, err := h.ser.CreateVehicle(*h.ctx, payload.DriverId, payload.Vehicle)
	if err != nil {
		h.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusCreated,
		Message:    "Vehicle created successfully",
		Data:       vehicle,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// UpdateVehicle handles PUT requests to update a vehicle
func (h *Handlers) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	// Extract vehicle ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("vehicle_id is required"), http.StatusBadRequest)
		return
	}

	vehicleID := pathParts[2]
	tempVehicleID, err := strconv.Atoi(vehicleID)
	if err != nil {
		h.errorJSON(w, errors.New("vehicle_id must be an integer"), http.StatusBadRequest)
		return
	}

	var vehicleRequest user_service.VehicleRequest
	err = h.readJSON(w, r, &vehicleRequest)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	vehicle, err := h.ser.UpdateVehicle(*h.ctx, tempVehicleID, vehicleRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Vehicle updated successfully",
		Data:       vehicle,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteVehicle handles DELETE requests to delete a vehicle
func (h *Handlers) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	// Extract vehicle ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.errorJSON(w, errors.New("vehicle_id is required"), http.StatusBadRequest)
		return
	}

	vehicleID := pathParts[2]
	tempVehicleID, err := strconv.Atoi(vehicleID)
	if err != nil {
		h.errorJSON(w, errors.New("vehicle_id must be an integer"), http.StatusBadRequest)
		return
	}

	err = h.ser.DeleteVehicleById(*h.ctx, tempVehicleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorJSON(w, err, http.StatusNotFound)
		} else {
			h.errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	response := ResponseDTO{
		StatusCode: http.StatusOK,
		Message:    "Vehicle deleted successfully",
		Data:       nil,
	}

	h.writeJSON(w, http.StatusOK, response)
}
