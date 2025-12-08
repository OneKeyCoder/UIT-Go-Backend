package main

import (
	"net/http"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()
	
	// Request ID must be first to inject into all logs
	mux.Use(commonMiddleware.RequestID)
	mux.Use(commonMiddleware.Logger)
	mux.Use(commonMiddleware.Recovery)
	mux.Use(commonMiddleware.PrometheusMetrics("trip-service"))

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.Heartbeat("/ping"))

	// Add OpenTelemetry HTTP instrumentation with operation-specific span names
	mux.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "trip-service.http",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				// Create span name from HTTP method + path
				return r.Method + " " + r.URL.Path
			}),
		)
	})

	mux.Get("/health/live", app.Liveness)
	mux.Get("/health/ready", app.Readiness)

	mux.Handle("/metrics", promhttp.Handler())

	mux.Post("/trip/create", app.CreateTrip)
	mux.Put("/trip/accept", app.AcceptTrip)
	mux.Put("/trip/reject", app.RejectTrip)
	mux.Get("/trip/suggested/{trip_id}", app.GetSuggestedDriver)
	mux.Get("/trip/{trip_id}/{user_id}", app.GetTripDetail)
	mux.Get("/trip/passenger/{passenger_id}", app.GetTripsByPassenger)
	mux.Get("/trip/driver/{driver_id}", app.GetTripsByDriver)
	mux.Get("/trips/{page}/{limit}", app.GetAllTrips)
	mux.Put("/trip/update", app.UpdateTripStatus)
	mux.Put("/trip/cancel", app.CancelTrip)
	mux.Put("/trip/review", app.ReviewTrip)
	mux.Get("/trip/review/{trip_id}", app.GetReview)
	return mux
}
