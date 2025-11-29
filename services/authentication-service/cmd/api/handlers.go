package main

import (
	"authentication-service/data"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/jwt"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
)

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type AuthResponse struct {
	User   interface{}      `json:"user"`
	Tokens *jwt.TokenPair   `json:"tokens"`
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload AuthRequest

	// Use modern request validation
	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	// Validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		logger.Warn("Failed authentication attempt",
			"email", requestPayload.Email,
			"error", err,
		)
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		logger.Warn("Invalid password",
			"email", requestPayload.Email,
		)
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Generate JWT token pair
	tokens, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		"user", // Default role, extend this based on your needs
		app.JWTSecret,
		app.JWTExpiry,
		app.RefreshExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate tokens",
			"email", user.Email,
			"error", err,
		)
		response.InternalServerError(w, "Failed to generate authentication tokens")
		return
	}

	logger.Info("User authenticated successfully",
		"email", user.Email,
		"user_id", user.ID,
	)

	// Log to logger service (async via RabbitMQ would be better)
	_ = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))

	// Return user data with tokens
	authResponse := AuthResponse{
		User:   user,
		Tokens: tokens,
	}

	response.Success(w, "Authentication successful", authResponse)
}

// RefreshToken handles refresh token requests
func (app *Config) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	err := request.ReadAndValidate(w, r, &req)
	if request.HandleError(w, err) {
		return
	}

	// Validate refresh token and generate new access token
	claims, err := jwt.ValidateToken(req.RefreshToken, app.JWTSecret)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			response.Unauthorized(w, "Refresh token has expired")
			return
		}
		response.Unauthorized(w, "Invalid refresh token")
		return
	}

	// Generate new token pair
	tokens, err := jwt.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Role,
		app.JWTSecret,
		app.JWTExpiry,
		app.RefreshExpiry,
	)
	if err != nil {
		response.InternalServerError(w, "Failed to generate tokens")
		return
	}

	response.Success(w, "Token refreshed successfully", tokens)
}

// ValidateToken validates JWT token (for other services to call)
func (app *Config) ValidateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Unauthorized(w, "Missing authorization header")
		return
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Unauthorized(w, "Invalid authorization header format")
		return
	}

	claims, err := jwt.ValidateToken(parts[1], app.JWTSecret)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			response.Unauthorized(w, "Token has expired")
			return
		}
		response.Unauthorized(w, "Invalid token")
		return
	}

	response.Success(w, "Token is valid", claims)
}

// Register handles new user registration
func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	var requestPayload RegisterRequest

	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	// Check if user already exists
	existingUser, _ := app.Models.User.GetByEmail(requestPayload.Email)
	if existingUser != nil {
		response.BadRequest(w, "User with this email already exists")
		return
	}

	// Create new user
	newUser := data.User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password, // Will be hashed by Insert()
		Active:    1,
	}

	userID, err := app.Models.User.Insert(newUser)
	if err != nil {
		logger.Error("Failed to create user",
			"email", requestPayload.Email,
			"error", err,
		)
		response.InternalServerError(w, "Failed to create user account")
		return
	}

	logger.Info("New user registered",
		"email", requestPayload.Email,
		"user_id", userID,
	)

	// Log registration event
	_ = app.logRequest("registration", fmt.Sprintf("New user registered: %s", requestPayload.Email))

	// Get the newly created user (without password)
	createdUser, err := app.Models.User.GetOne(userID)
	if err != nil {
		// User was created but we couldn't fetch it - still success
		response.Success(w, "Registration successful", map[string]interface{}{
			"id":    userID,
			"email": requestPayload.Email,
		})
		return
	}

	response.Success(w, "Registration successful", createdUser)
}

// ChangePassword handles password change requests (requires old password verification)
func (app *Config) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var requestPayload ChangePasswordRequest

	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	// Get user by email
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		logger.Warn("Password change attempt for non-existent user",
			"email", requestPayload.Email,
		)
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Verify old password
	valid, err := user.PasswordMatches(requestPayload.OldPassword)
	if err != nil || !valid {
		logger.Warn("Invalid old password during password change",
			"email", requestPayload.Email,
		)
		response.Unauthorized(w, "Invalid old password")
		return
	}

	// Update to new password
	err = user.ResetPassword(requestPayload.NewPassword)
	if err != nil {
		logger.Error("Failed to change password",
			"email", user.Email,
			"error", err,
		)
		response.InternalServerError(w, "Failed to change password")
		return
	}

	logger.Info("Password changed successfully",
		"email", user.Email,
		"user_id", user.ID,
	)

	// Log password change event
	_ = app.logRequest("password_change", fmt.Sprintf("Password changed for user: %s", user.Email))

	response.Success(w, "Password changed successfully", nil)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
