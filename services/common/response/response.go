package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response structure
type Response struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorDetail provides detailed error information
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ErrorResponse represents an error response with details
type ErrorResponse struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	return json.NewEncoder(w).Encode(data)
}

// Success writes a successful JSON response
func Success(w http.ResponseWriter, message string, data interface{}) error {
	return WriteJSON(w, http.StatusOK, Response{
		Error:   false,
		Message: message,
		Data:    data,
	})
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, message string, data interface{}) error {
	return WriteJSON(w, http.StatusCreated, Response{
		Error:   false,
		Message: message,
		Data:    data,
	})
}

// BadRequest writes a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string, details ...ErrorDetail) error {
	return WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		Error:   true,
		Message: message,
		Details: details,
	})
}

// Unauthorized writes a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) error {
	return WriteJSON(w, http.StatusUnauthorized, Response{
		Error:   true,
		Message: message,
	})
}

// Forbidden writes a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) error {
	return WriteJSON(w, http.StatusForbidden, Response{
		Error:   true,
		Message: message,
	})
}

// NotFound writes a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) error {
	return WriteJSON(w, http.StatusNotFound, Response{
		Error:   true,
		Message: message,
	})
}

// InternalServerError writes a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) error {
	return WriteJSON(w, http.StatusInternalServerError, Response{
		Error:   true,
		Message: message,
	})
}

// ValidationError writes a validation error response
func ValidationError(w http.ResponseWriter, details []ErrorDetail) error {
	return WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Error:   true,
		Message: "Validation failed",
		Details: details,
	})
}
