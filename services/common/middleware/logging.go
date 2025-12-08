package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey     contextKey = "request_id"
	RequestLoggerKey contextKey = "request_logger"
)

// LoggingMiddleware adds request context and logger to every request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Create logger with request context
		reqLogger := logger.RequestLogger(
			r.Context(),
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
			requestID,
		)

		// Inject into request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, RequestLoggerKey, reqLogger)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status code
		wrw := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request start
		reqLogger.Info("incoming request",
			"query", r.URL.RawQuery,
		)

		// Process request
		next.ServeHTTP(wrw, r)

		// Log request completion
		duration := time.Since(start)
		reqLogger.Info("request completed",
			"status_code", wrw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"bytes_written", wrw.bytesWritten,
		)
	})
}

// wrappedResponseWriter captures response status and size
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// GetRequestLogger retrieves logger from request context
func GetRequestLogger(ctx context.Context) *slog.Logger {
	if reqLogger, ok := ctx.Value(RequestLoggerKey).(*slog.Logger); ok {
		return reqLogger
	}
	return logger.WithContext(ctx) // Fallback to basic logger
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}
