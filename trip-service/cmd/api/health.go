package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Liveness probe - just check if the service is running
func (app *Config) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness probe - check if service can handle requests
func (app *Config) Readiness(w http.ResponseWriter, r *http.Request) {
	checks := map[string]bool{
		"database": app.checkDatabase(),
	}

	allHealthy := true
	for _, healthy := range checks {
		if !healthy {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ready",
			"checks": checks,
		})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not ready",
			"checks": checks,
		})
	}
}

func (app *Config) checkDatabase() bool {
	if app.TripService.DB == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := app.TripService.DB.PingContext(ctx)
	return err == nil
}
