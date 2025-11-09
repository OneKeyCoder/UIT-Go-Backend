package main

import (
	"logger-service/data"
	"net/http"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
)

type JSONPayload struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" validate:"required"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err = app.Models.LogEntry.Insert(event)
	if err != nil {
		response.InternalServerError(w, "Failed to insert log entry")
		return
	}

	response.Success(w, "Logged successfully", nil)
}
