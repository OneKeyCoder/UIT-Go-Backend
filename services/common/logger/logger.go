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

// filteringOTLPHandler wraps OTLP handler with configurable level filtering
// Log Levels (from lowest to highest severity):
// - DEBUG: Detailed information for debugging (development/staging only)
// - INFO: Confirmation that things are working as expected
// - WARN: Something unexpected but system still functioning
// - ERROR: Error occurred, feature broken but system still running
// - FATAL: Critical error, application crash or cannot serve
type filteringOTLPHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func (f *filteringOTLPHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Filter based on configured minimum level
	return level >= f.minLevel
}

func (f *filteringOTLPHandler) Handle(ctx context.Context, r slog.Record) error {
	return f.handler.Handle(ctx, r)
}

func (f *filteringOTLPHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &filteringOTLPHandler{
		handler:  f.handler.WithAttrs(attrs),
		minLevel: f.minLevel,
	}
}

func (f *filteringOTLPHandler) WithGroup(name string) slog.Handler {
	return &filteringOTLPHandler{
		handler:  f.handler.WithGroup(name),
		minLevel: f.minLevel,
	}
}

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

	// Add OTLP handler if LoggerProvider is available
	// Production: Send INFO+ to Loki (DEBUG excluded to reduce noise)
	// Development: Send DEBUG+ to Loki (full visibility)
	if lp != nil {
		otelHandler := otelslog.NewHandler(service, otelslog.WithLoggerProvider(lp))
		minLevel := slog.LevelInfo // Production default: INFO+
		if isDevelopment {
			minLevel = slog.LevelDebug // Development: DEBUG+ for full visibility
		}
		handlers = append(handlers, &filteringOTLPHandler{
			handler:  otelHandler,
			minLevel: minLevel,
		})
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

// ErrorCtx logs an error message with trace context
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).ErrorContext(ctx, msg, args...)
}

// WarnCtx logs a warning message with trace context
func WarnCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).WarnContext(ctx, msg, args...)
}

// DebugCtx logs a debug message with trace context
func DebugCtx(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).DebugContext(ctx, msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	if Log != nil {
		Log.Info(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if Log != nil {
		Log.Error(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if Log != nil {
		Log.Warn(msg, args...)
	}
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if Log != nil {
		Log.Debug(msg, args...)
	}
}

// Fatal logs a fatal message at ERROR+4 level and exits
// FATAL indicates critical errors that cause application crash
func Fatal(msg string, args ...any) {
	if Log != nil {
		// Log at ERROR+4 level (equivalent to FATAL/CRITICAL)
		Log.Log(context.Background(), slog.LevelError+4, msg, args...)
	}
	os.Exit(1)
}

// FatalCtx logs a fatal message with trace context and exits
func FatalCtx(ctx context.Context, msg string, args ...any) {
	if Log != nil {
		WithContext(ctx).Log(ctx, slog.LevelError+4, msg, args...)
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

// RequestLogger creates a logger with HTTP request context
// Includes: method, path, remote_addr, user_agent, request_id
func RequestLogger(ctx context.Context, method, path, remoteAddr, userAgent, requestID string) *slog.Logger {
	logger := WithContext(ctx).With(
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"user_agent", userAgent,
		"request_id", requestID,
	)
	return logger
}

// GRPCLogger creates a logger with gRPC context
// Includes: grpc_method, peer_addr
func GRPCLogger(ctx context.Context, method, peerAddr string) *slog.Logger {
	logger := WithContext(ctx).With(
		"grpc_method", method,
		"peer_addr", peerAddr,
		"protocol", "grpc",
	)
	return logger
}

// ErrorWithStack logs error with full context and stack trace hint
func ErrorWithStack(ctx context.Context, msg string, err error, extraFields ...any) {
	fields := []any{"error", err.Error()}
	fields = append(fields, extraFields...)
	WithContext(ctx).ErrorContext(ctx, msg, fields...)
}
