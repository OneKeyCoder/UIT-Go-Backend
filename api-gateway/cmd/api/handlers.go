package main

import (
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	"net/http"
)

type RequestPayload struct {
	Action string       `json:"action" validate:"required"`
	Auth   *AuthPayload `json:"auth,omitempty"`
	Log    *LogPayload  `json:"log,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogPayload struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" validate:"required"`
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

