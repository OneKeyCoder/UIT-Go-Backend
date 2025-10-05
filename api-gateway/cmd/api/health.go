package main

import (
	"encoding/json"
	"net/http"
)

// Liveness probe - just check if the service is running
// Kubernetes uses this to know if it should restart the pod
func (app *Config) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness probe - check if service can handle requests
// Kubernetes uses this to know if it should send traffic to the pod
func (app *Config) Readiness(w http.ResponseWriter, r *http.Request) {
	// Check gRPC connections
	checks := map[string]bool{
		"auth_grpc_client":   app.checkAuthConnection(),
		"logger_grpc_client": app.checkLoggerConnection(),
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

func (app *Config) checkAuthConnection() bool {
	if app.GRPCClients == nil || app.GRPCClients.AuthClient == nil {
		return false
	}
	// Connection exists and client is initialized
	return true
}

func (app *Config) checkLoggerConnection() bool {
	if app.GRPCClients == nil || app.GRPCClients.LoggerClient == nil {
		return false
	}
	return true
}
