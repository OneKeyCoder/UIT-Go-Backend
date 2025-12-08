package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
)

// skipLogPaths are paths that should not be logged (health checks, metrics)
var skipLogPaths = []string{
	"/metrics",
	"/health",
	"/healthz",
	"/ready",
	"/readiness",
	"/live",
	"/liveness",
}

// shouldSkipLog returns true if the path should not be logged
func shouldSkipLog(path string) bool {
	for _, skip := range skipLogPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// Logger middleware logs HTTP requests with structured logging
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		if shouldSkipLog(r.URL.Path) {
			return
		}

		duration := time.Since(start)

		// Log based on status code level
		if wrapped.statusCode >= 500 {
			logger.Error("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		} else if wrapped.statusCode >= 400 {
			logger.Warn("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		} else {
			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		}
	})
}

// Recovery middleware recovers from panics with structured logging
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.RequestURI,
				)
				response.InternalServerError(w, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter is a wrapper to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
