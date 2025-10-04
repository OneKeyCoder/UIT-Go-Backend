package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// Add common middleware
	mux.Use(commonMiddleware.Logger)
	mux.Use(commonMiddleware.Recovery)
	
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

	// Broker routes
	mux.Post("/", app.Broker)
	mux.Post("/handle", app.HandleSubmission)
	
	// gRPC-based routes (using persistent clients with interceptors)
	mux.Post("/grpc/auth", func(w http.ResponseWriter, r *http.Request) {
		var authPayload AuthPayload
		err := request.ReadAndValidate(w, r, &authPayload)
		if request.HandleError(w, err) {
			return
		}
		app.authenticateViaGRPC(w, authPayload)
	})
	
	mux.Post("/grpc/log", func(w http.ResponseWriter, r *http.Request) {
		var logPayload LogPayload
		err := request.ReadAndValidate(w, r, &logPayload)
		if request.HandleError(w, err) {
			return
		}
		app.logItemViaGRPCClient(w, logPayload)
	})

	return mux
}
