package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"go.uber.org/zap"
)

// Logger middleware logs HTTP requests with structured logging
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		
		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.RequestURI),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

// Recovery middleware recovers from panics with structured logging
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", r.Method),
					zap.String("path", r.RequestURI),
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
