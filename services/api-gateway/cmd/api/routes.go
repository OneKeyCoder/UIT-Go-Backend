package main

import (
	"net/http"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/didip/tollbooth/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// Add common middleware
	mux.Use(commonMiddleware.Logger)
	mux.Use(commonMiddleware.Recovery)
	mux.Use(commonMiddleware.PrometheusMetrics("api-gateway"))

	// CORS configuration
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting: 100 requests per minute per IP
	// Applied to all routes after other middleware
	lmt := tollbooth.NewLimiter(100.0/60.0, nil)
	// Allow short bursts so probes or short spikes don't immediately get 429
	lmt.SetBurst(100)
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})
	mux.Use(func(next http.Handler) http.Handler {
		return tollbooth.LimitHandler(lmt, next)
	})

	// Simple ping endpoint
	mux.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// Health check endpoints for Kubernetes (or any orchestrator)
	// /health/live - Liveness probe (is the service running?)
	// /health/ready - Readiness probe (can the service handle requests?)
	mux.Get("/health/live", app.Liveness)
	mux.Get("/health/ready", app.Readiness)

	// Metrics endpoint for Prometheus
	mux.Handle("/metrics", promhttp.Handler())

	// Broker routes
	mux.Post("/", app.Broker)

	// gRPC-based routes (using persistent clients with interceptors)
	mux.Post("/grpc/register", func(w http.ResponseWriter, r *http.Request) {
		var regPayload RegisterPayload
		err := request.ReadAndValidate(w, r, &regPayload)
		if request.HandleError(w, err) {
			return
		}
		app.registerViaGRPC(w, r, regPayload)
	})

	mux.Post("/grpc/auth", func(w http.ResponseWriter, r *http.Request) {
		var authPayload AuthPayload
		err := request.ReadAndValidate(w, r, &authPayload)
		if request.HandleError(w, err) {
			return
		}
		app.authenticateViaGRPC(w, r, authPayload)
	})

	mux.Post("/grpc/log", func(w http.ResponseWriter, r *http.Request) {
		var logPayload LogPayload
		err := request.ReadAndValidate(w, r, &logPayload)
		if request.HandleError(w, err) {
			return
		}
		app.logItemViaGRPCClient(w, r, logPayload)
	})

	mux.Route("/location", func(r chi.Router) {
		r.Use(app.AuthRequired)
		r.Get("/me", app.getLocationViaGRPC)
		r.Post("/", app.setLocationViaGRPC)
		r.Get("/nearest", app.findNearestUsersViaGRPC)
		r.Get("/", app.getAllLocationsViaGRPC)
	})

	mux.Route("/trip", func(r chi.Router) {
		r.Use(app.AuthRequired)
		r.Post("/", app.CreateTrip)
		r.Put("/accept/{tripID}", app.AcceptTrip)
		r.Put("/reject/{tripID}", app.RejectTrip)
		r.Get("/", app.GetAllTrips)
		r.Get("/{tripID}", app.GetTripDetails)
		r.Get("/user", app.GetTripsByPassenger)
		r.Get("/driver", app.GetTripsByDriver)
		r.Get("/suggested/{tripID}", app.GetSuggestedDriver)
		r.Put("/status/{tripID}", app.UpdateTripStatus)
		r.Put("/cancel/{tripID}", app.CancelTrip)
		r.Put("/review/{tripID}", app.SubmitReview)
		r.Get("/review/{tripID}", app.GetTripReview)
	})

	// User and Vehicle routes
	mux.Route("/users", func(r chi.Router) {
		r.Use(app.AuthRequired)
		r.Get("/", app.GetAllUsers)
		r.Get("/{id}", app.GetUserById)
		r.Post("/", app.CreateUser)
		r.Put("/{id}", app.UpdateUser)
		r.Delete("/{id}", app.DeleteUser)
		r.Get("/{id}/vehicles", app.GetVehiclesByUserId)
	})

	mux.Route("/vehicles", func(r chi.Router) {
		r.Use(app.AuthRequired)
		r.Get("/", app.GetAllVehicles)
		r.Get("/{id}", app.GetVehicleById)
		r.Post("/", app.CreateVehicle)
		r.Put("/{id}", app.UpdateVehicle)
		r.Delete("/{id}", app.DeleteVehicle)
	})

	return mux
}
