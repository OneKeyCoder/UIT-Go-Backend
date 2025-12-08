package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Log is the global slog logger instance
	Log *slog.Logger
	// serviceName stores the service name for context
	serviceName string
)

// Init initializes the global logger with OTLP backend
// Must be called AFTER telemetry.InitTracer() to use OTLP export
func Init(service string, isDevelopment bool) error {
	serviceName = service

	// Check if OTLP LoggerProvider is available (set by telemetry.InitTracer)
	lp := global.GetLoggerProvider()

	// Create multi-handler: stdout (all levels) + OTLP (WARN+ only)
	var handlers []slog.Handler

	// Always add stdout handler for container logs visibility (all levels)
	stdoutOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	if isDevelopment {
		// Pretty text output for development
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, stdoutOpts))
	} else {
		// JSON output for production
		handlers = append(handlers, slog.NewJSONHandler(os.Stdout, stdoutOpts))
	}

	// Add OTLP handler if LoggerProvider is available (ALL log levels to Loki)
	if lp != nil {
		otelHandler := otelslog.NewHandler(service, otelslog.WithLoggerProvider(lp))
		handlers = append(handlers, otelHandler)
		// DEBUG: Log that OTLP handler was added
		fmt.Fprintf(os.Stderr, "[LOGGER INIT] OTLP handler ADDED for service=%s, handlers_count=%d\n", service, len(handlers))
	} else {
		// DEBUG: Log that OTLP handler was NOT added
		fmt.Fprintf(os.Stderr, "[LOGGER INIT] OTLP handler NOT added - LoggerProvider is nil for service=%s\n", service)
	}

	// Create multi-handler logger
	Log = slog.New(&multiHandler{handlers: handlers}).With("service", service)

	return nil
}

// InitDefault initializes with default development settings
func InitDefault(service string) {
	if err := Init(service, true); err != nil {
		// Fallback to basic logger
		Log = slog.Default().With("service", service)
	}
}


// WithContext returns logger with trace context if available
func WithContext(ctx context.Context) *slog.Logger {
	if Log == nil {
		return slog.Default()
	}
	
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return Log
	}
	
	spanCtx := span.SpanContext()
	return Log.With(
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	)
}

// InfoCtx logs an info message with trace context
func InfoCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).InfoContext(ctx, msg, args...)
}

// Info logs an info message (alias for backward compatibility - now auto-extracts trace if first arg is context)
func Info(msgOrCtx any, args ...any) {
	if ctx, ok := msgOrCtx.(context.Context); ok && len(args) > 0 {
		// New pattern: logger.Info(ctx, "message", "key", "value")
		if msg, ok := args[0].(string); ok {
			InfoCtx(ctx, msg, args[1:]...)
			return
		}
	}
	// Old pattern: logger.Info("message", "key", "value") - no context
	if msg, ok := msgOrCtx.(string); ok {
		if Log != nil {
			Log.Info(msg, args...)
		}
	}
}

// ErrorCtx logs an error message with trace context
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).ErrorContext(ctx, msg, args...)
}

// Error logs an error message (alias with auto-context extraction)
func Error(msgOrCtx any, args ...any) {
	if ctx, ok := msgOrCtx.(context.Context); ok && len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			ErrorCtx(ctx, msg, args[1:]...)
			return
		}
	}
	if msg, ok := msgOrCtx.(string); ok {
		if Log != nil {
			Log.Error(msg, args...)
		}
	}
}

// WarnCtx logs a warning message with trace context
func WarnCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).WarnContext(ctx, msg, args...)
}

// Warn logs a warning message (alias with auto-context extraction)
func Warn(msgOrCtx any, args ...any) {
	if ctx, ok := msgOrCtx.(context.Context); ok && len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			WarnCtx(ctx, msg, args[1:]...)
			return
		}
	}
	if msg, ok := msgOrCtx.(string); ok {
		if Log != nil {
			Log.Warn(msg, args...)
		}
	}
}

// DebugCtx logs a debug message with trace context
func DebugCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).DebugContext(ctx, msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if Log != nil {
		Log.Debug(msg, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...any) {
	if Log != nil {
		Log.Error(msg, args...)
	}
	os.Exit(1)
}

// Infof logs a formatted info message
func Infof(template string, args ...interface{}) {
	if Log != nil {
		Log.Info(fmt.Sprintf(template, args...))
	}
}

// Errorf logs a formatted error message
func Errorf(template string, args ...interface{}) {
	if Log != nil {
		Log.Error(fmt.Sprintf(template, args...))
	}
}

// Warnf logs a formatted warning message
func Warnf(template string, args ...interface{}) {
	if Log != nil {
		Log.Warn(fmt.Sprintf(template, args...))
	}
}

// Debugf logs a formatted debug message
func Debugf(template string, args ...interface{}) {
	if Log != nil {
		Log.Debug(fmt.Sprintf(template, args...))
	}
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(template string, args ...interface{}) {
	if Log != nil {
		Log.Error(fmt.Sprintf(template, args...))
	}
	os.Exit(1)
}

// multiHandler is a slog.Handler that writes to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				// Continue to other handlers even if one fails
				continue
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// Application log helpers - for startup, shutdown, errors
// These will have log_type="application" for filtering in Loki

func AppInfo(msg string, args ...any) {
	if Log != nil {
		Log.Info(msg, append(args, "log_type", "application")...)
	}
}

func AppError(msg string, args ...any) {
	if Log != nil {
		Log.Error(msg, append(args, "log_type", "application")...)
	}
}

func AppWarn(msg string, args ...any) {
	if Log != nil {
		Log.Warn(msg, append(args, "log_type", "application")...)
	}
}

// Request log helpers - for HTTP request/response logs
// These will have log_type="request" for filtering in Loki

func RequestInfo(msg string, args ...any) {
	if Log != nil {
		Log.Info(msg, append(args, "log_type", "request")...)
	}
}

func RequestError(msg string, args ...any) {
	if Log != nil {
		Log.Error(msg, append(args, "log_type", "request")...)
	}
}

func RequestWarn(msg string, args ...any) {
	if Log != nil {
		Log.Warn(msg, append(args, "log_type", "request")...)
	}
}
