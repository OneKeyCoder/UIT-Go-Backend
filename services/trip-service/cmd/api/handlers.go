package main

// Viết cho giống mấy service khác chứ xài grpc rồi cần gì :v
import (
	"net/http"
	"strconv"
	"trip-service/internal/models"
	"trip-service/internal/repository"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/go-chi/chi/v5"
)

type TripResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type AcceptTripRequest struct {
	DriverID int `json:"driver_id" validate:"required"`
	TripID   int `json:"trip_id" validate:"required"`
}

type RejectTripRequest struct {
	PassengerID int `json:"passenger_id" validate:"required"`
	DriverID    int `json:"driver_id" validate:"required"`
	TripID      int `json:"trip_id" validate:"required"`
}

type UpdateTripDetailRequest struct {
	TripStatus models.TripStatus `json:"trip_status" validate:"required,oneof=REQUESTED ACCEPTED STARTED COMPLETED CANCELLED"`
	UserID     int               `json:"user_id" validate:"required"`
	TripID     int               `json:"trip_id" validate:"required"`
}

type ReviewRequest struct {
	TripID int                  `json:"trip_id" validate:"required"`
	UserID int                  `json:"user_id" validate:"required"`
	Review repository.ReviewDTO `json:"review" validate:"required"`
}

func (app *Config) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var tripRequest repository.NewTripDTO
	err := request.ReadAndValidate(w, r, &tripRequest)
	if request.HandleError(w, err) {
		return
	}

	tripRecord, duration, err := app.TripService.CreateTrip(r.Context(), tripRequest)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, 202, TripResponse{
		Error:   false,
		Message: "Trip created successfully",
		Data: map[string]interface{}{
			"trip":     tripRecord,
			"duration": duration,
		},
	})
}

func (app *Config) AcceptTrip(w http.ResponseWriter, r *http.Request) {
	var acceptRequest AcceptTripRequest
	err := request.ReadAndValidate(w, r, &acceptRequest)
	if request.HandleError(w, err) {
		return
	}

	err = app.TripService.AcceptTrip(acceptRequest.DriverID, acceptRequest.TripID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip accepted successfully",
	})
}

func (app *Config) RejectTrip(w http.ResponseWriter, r *http.Request) {
	var rejectRequest RejectTripRequest
	err := request.ReadAndValidate(w, r, &rejectRequest)
	if request.HandleError(w, err) {
		return
	}

	err = app.TripService.RejectTrip(rejectRequest.PassengerID, rejectRequest.DriverID, rejectRequest.TripID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip rejected successfully",
	})
}

func (app *Config) GetSuggestedDriver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tripID, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	driver, err := app.TripService.GetSuggestedDriver(tripID)
	if err != nil || driver == 0 {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Suggested driver retrieved successfully",
		Data:    driver,
	})
}

func (app *Config) GetTripDetail(w http.ResponseWriter, r *http.Request) {
	id1 := chi.URLParam(r, "user_id")
	tripID, err := strconv.Atoi(id1)
	if err != nil {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	id2 := chi.URLParam(r, "trip_id")
	userID, err := strconv.Atoi(id2)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}
	tripRecord, err := app.TripService.GetTrip(userID, tripID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip detail retrieved successfully",
		Data:    tripRecord,
	})
}
func (app *Config) GetTripsByPassenger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "passenger_id")
	passengerID, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, "Invalid passenger ID")
		return
	}
	trips, err := app.TripService.GetTripsByPassenger(passengerID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trips retrieved successfully",
		Data:    trips,
	})
}

func (app *Config) GetTripsByDriver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "driver_id")
	driverID, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, "Invalid driver ID")
		return
	}
	trips, err := app.TripService.GetTripsByDriver(driverID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trips retrieved successfully",
		Data:    trips,
	})
}

func (app *Config) GetAllTrips(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	trips, err := app.TripService.GetAllTrips(page, limit)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trips retrieved successfully",
		Data:    trips,
	})
}

func (app *Config) UpdateTripStatus(w http.ResponseWriter, r *http.Request) {
	var updateRequest UpdateTripDetailRequest
	err := request.ReadAndValidate(w, r, &updateRequest)
	if request.HandleError(w, err) {
		return
	}

	err = app.TripService.UpdateTripStatus(updateRequest.TripStatus, updateRequest.TripID, updateRequest.UserID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip status updated successfully",
	})
}

func (app *Config) CancelTrip(w http.ResponseWriter, r *http.Request) {
	var cancelRequest UpdateTripDetailRequest
	err := request.ReadAndValidate(w, r, &cancelRequest)
	if request.HandleError(w, err) {
		return
	}

	err = app.TripService.CancelTrip(cancelRequest.UserID, cancelRequest.TripID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip cancelled successfully",
	})
}

func (app *Config) ReviewTrip(w http.ResponseWriter, r *http.Request) {
	var reviewRequest ReviewRequest
	err := request.ReadAndValidate(w, r, &reviewRequest)
	if request.HandleError(w, err) {
		return
	}

	err = app.TripService.ReviewTrip(reviewRequest.UserID, reviewRequest.TripID, reviewRequest.Review)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Trip reviewed successfully",
	})
}

func (app *Config) GetReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "trip_id")
	tripID, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	id = chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		response.BadRequest(w, "Invalid trip ID")
		return
	}
	review, err := app.TripService.GetReview(tripID, userID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, TripResponse{
		Error:   false,
		Message: "Review retrieved successfully",
		Data:    review,
	})
}
