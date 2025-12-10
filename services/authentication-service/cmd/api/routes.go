package main

import (
	"net/http"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// Request ID must be first
	mux.Use(commonMiddleware.RequestID)
	
	// OpenTelemetry HTTP instrumentation BEFORE Logger
	mux.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			next,
			"authentication-service.http",
			otelhttp.WithFilter(func(req *http.Request) bool {
				return !commonMiddleware.ShouldSkipTrace(req.URL.Path)
			}),
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	})
	
	// Logger AFTER otelhttp
	mux.Use(commonMiddleware.Logger)
	mux.Use(commonMiddleware.Recovery)
	mux.Use(commonMiddleware.PrometheusMetrics("authentication-service"))

	// CORS configuration
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	// Health check endpoints for Kubernetes
	mux.Get("/health/live", app.Liveness)
	mux.Get("/health/ready", app.Readiness)

	// Metrics endpoint REMOVED - using OTLP push to Alloy
	// mux.Handle("/metrics", promhttp.Handler())

	// Authentication routes
	mux.Post("/register", app.Register)
	mux.Post("/authenticate", app.Authenticate)
	mux.Post("/refresh", app.RefreshToken)
	mux.Post("/validate", app.ValidateToken)
	mux.Patch("/change-password", app.ChangePassword)

	return mux
}
