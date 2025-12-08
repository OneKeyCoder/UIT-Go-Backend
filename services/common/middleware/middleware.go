package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ContextKey type for request context keys
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
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

// RequestID middleware generates a unique request ID and adds it to context
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header (from upstream)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add request ID to response header for debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
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
		
		// Extract trace ID and request ID from context
		traceID := ""
		spanID := ""
		if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		}
		
		requestID := GetRequestID(r.Context())
		
		// Log based on status code level using Request log type with request_id
		if wrapped.statusCode >= 500 {
			logger.RequestError("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"trace_id", traceID,
				"span_id", spanID,
				"request_id", requestID,
			)
		} else if wrapped.statusCode >= 400 {
			logger.RequestWarn("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"trace_id", traceID,
				"span_id", spanID,
				"request_id", requestID,
			)
		} else {
			logger.RequestInfo("HTTP request",
				"method", r.Method,
				"path", r.RequestURI,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"trace_id", traceID,
				"span_id", spanID,
				"request_id", requestID,
			)
		}
	})
}

// Recovery middleware recovers from panics with structured logging
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.AppError("Panic recovered",
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
