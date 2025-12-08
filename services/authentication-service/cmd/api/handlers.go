package main

import (
	"authentication-service/data"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/jwt"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	userpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/user"
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
	User   interface{}    `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload AuthRequest

	// Use modern request validation
	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	// Get logger with request context
	reqLogger := commonMiddleware.GetRequestLogger(r.Context())

	// Validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		reqLogger.Warn("Failed authentication attempt",
			"email", requestPayload.Email,
			"error", err,
		)
		// Log failed authentication attempt for security monitoring
		app.logAuditEventAsync(r, "USER_LOGIN", requestPayload.Email, "failure", AuditMetadata{
			IP:        getClientIP(r),
			UserAgent: r.UserAgent(),
			Email:     requestPayload.Email,
			Action:    "Login attempt",
			Reason:    "User not found",
		})
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		reqLogger.Warn("Invalid password",
			"email", requestPayload.Email,
			"user_id", user.ID,
		)
		// Log failed password attempt for brute-force detection
		app.logAuditEventAsync(r, "USER_LOGIN", strconv.Itoa(user.ID), "failure", AuditMetadata{
			IP:        getClientIP(r),
			UserAgent: r.UserAgent(),
			Email:     requestPayload.Email,
			Action:    "Login attempt",
			Reason:    "Invalid password",
		})
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
		reqLogger.Error("Failed to generate tokens",
			"email", user.Email,
			"user_id", user.ID,
			"error", err,
		)
		response.InternalServerError(w, "Failed to generate authentication tokens")
		return
	}

	reqLogger.Info("User authenticated successfully",
		"email", user.Email,
		"user_id", user.ID,
	)

	// Log successful authentication with full context
	app.logAuditEventAsync(r, "USER_LOGIN", strconv.Itoa(user.ID), "success", AuditMetadata{
		IP:        getClientIP(r),
		UserAgent: r.UserAgent(),
		Email:     user.Email,
		Action:    "User logged in successfully",
	})

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

	// Create user in user-service via gRPC
	if app.UserClient != nil {
		userCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		_, err = app.UserClient.CreateUser(userCtx, &userpb.CreateUserRequest{
			Email:     requestPayload.Email,
			FirstName: requestPayload.FirstName,
			LastName:  requestPayload.LastName,
		})
		if err != nil {
			logger.Warn("Failed to create user in user-service",
				"email", requestPayload.Email,
				"error", err,
			)
			// Don't fail the registration if user-service is down
		}
	}

	// Log registration event with full context
	app.logAuditEventAsync(r, "USER_REGISTRATION", strconv.Itoa(userID), "success", AuditMetadata{
		IP:        getClientIP(r),
		UserAgent: r.UserAgent(),
		Email:     requestPayload.Email,
		Action:    "New user registered",
		Extra: map[string]interface{}{
			"first_name": requestPayload.FirstName,
			"last_name":  requestPayload.LastName,
		},
	})

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

	// Log password change event with full context
	app.logAuditEventAsync(r, "PASSWORD_CHANGE", strconv.Itoa(user.ID), "success", AuditMetadata{
		IP:        getClientIP(r),
		UserAgent: r.UserAgent(),
		Email:     user.Email,
		Action:    "Password changed successfully",
	})

	response.Success(w, "Password changed successfully", nil)
}

// getClientIP extracts the real client IP from request headers
// Checks X-Forwarded-For, X-Real-IP, and falls back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can be a comma-separated list, take the first one
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// logAuditEventAsync sends structured audit events to RabbitMQ asynchronously
// This provides comprehensive context: Who (actor), When (timestamp), What (event), Where (IP/UA)
func (app *Config) logAuditEventAsync(r *http.Request, eventName, actorID, status string, metadata AuditMetadata) {
	// Skip if RabbitMQ connection is not available
	if app.RabbitConn == nil {
		logger.Warn("RabbitMQ not available, skipping audit log", "event", eventName)
		return
	}

	// Run in goroutine to avoid blocking the request handler
	go func() {
		// Use reusable session if available (reduces connection overhead under load)
		var err error
		if app.RabbitSession != nil {
			err = PublishAuditEventWithSession(app.RabbitSession, app.RabbitConn, eventName, actorID, status, metadata)
		} else {
			err = PublishAuditEvent(app.RabbitConn, eventName, actorID, status, metadata)
		}

		if err != nil {
			// Log error but don't fail the request
			logger.Error("Failed to publish audit event to RabbitMQ",
				"event", eventName,
				"actor", actorID,
				"status", status,
				"error", err,
			)
		}
	}()
}
