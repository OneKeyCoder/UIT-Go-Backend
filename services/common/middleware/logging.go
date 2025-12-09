package middleware

import (
	"context"
	"log/slog"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
)

// RequestLoggerKey is the context key for storing request-scoped logger
const RequestLoggerKey contextKey = "request_logger"

// GetRequestLogger retrieves logger from request context
// Returns *traceContextLogger which has all logging methods
func GetRequestLogger(ctx context.Context) interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	With(args ...any) *slog.Logger
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
} {
	if reqLogger, ok := ctx.Value(RequestLoggerKey).(interface {
		InfoContext(ctx context.Context, msg string, args ...any)
		ErrorContext(ctx context.Context, msg string, args ...any)
		WarnContext(ctx context.Context, msg string, args ...any)
		DebugContext(ctx context.Context, msg string, args ...any)
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
		With(args ...any) *slog.Logger
		Info(msg string, args ...any)
		Error(msg string, args ...any)
		Warn(msg string, args ...any)
		Debug(msg string, args ...any)
	}); ok {
		return reqLogger
	}
	return logger.WithContext(ctx) // Fallback to basic logger
}
