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
		"rabbitmq": app.checkRabbitMQ(),
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
	if app.DB == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := app.DB.PingContext(ctx)
	return err == nil
}

func (app *Config) checkRabbitMQ() bool {
	if app.RabbitConn == nil {
		return false
	}
	// go-amqp doesn't have IsClosed(), so we check by trying to create a session
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	session, err := app.RabbitConn.NewSession(ctx, nil)
	if err != nil {
		return false
	}
	session.Close(ctx)
	return true
}
