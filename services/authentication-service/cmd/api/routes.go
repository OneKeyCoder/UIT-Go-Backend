package main

import (
	"net/http"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// Add common middleware
	// TODO: Re-enable LoggingMiddleware after debugging gRPC timeout
	// mux.Use(commonMiddleware.LoggingMiddleware) // Structured logging with request context
	mux.Use(commonMiddleware.Logger)
	mux.Use(commonMiddleware.Recovery)
	mux.Use(commonMiddleware.PrometheusMetrics("authentication-service"))

	// CORS configuration
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	// Health check endpoints for Kubernetes
	mux.Get("/health/live", app.Liveness)
	mux.Get("/health/ready", app.Readiness)

	// Metrics endpoint for Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	// Authentication routes
	mux.Post("/register", app.Register)
	mux.Post("/authenticate", app.Authenticate)
	mux.Post("/refresh", app.RefreshToken)
	mux.Post("/validate", app.ValidateToken)
	mux.Patch("/change-password", app.ChangePassword)

	return mux
}
