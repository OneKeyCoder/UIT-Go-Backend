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

// traceContextLogger wraps slog.Logger to inject trace_id into message body for Grafana correlation
type traceContextLogger struct {
	logger  *slog.Logger
	traceID string
}

func (l *traceContextLogger) With(args ...any) *slog.Logger {
	return l.logger.With(args...)
}

func (l *traceContextLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	// trace_id is already in structured metadata via logger.With()
	// No need to inject into message body
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *traceContextLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *traceContextLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *traceContextLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *traceContextLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.logger.Log(ctx, level, msg, args...)
}

// Non-context versions for backward compatibility
func (l *traceContextLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *traceContextLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *traceContextLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *traceContextLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// WithContext returns logger with trace context that auto-injects trace_id into messages
func WithContext(ctx context.Context) *traceContextLogger {
	if Log == nil {
		return &traceContextLogger{
			logger:  slog.Default(),
			traceID: "",
		}
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return &traceContextLogger{
			logger:  Log,
			traceID: "",
		}
	}

	spanCtx := span.SpanContext()
	traceID := spanCtx.TraceID().String()

	return &traceContextLogger{
		logger: Log.With(
			"trace_id", traceID,
			"span_id", spanCtx.SpanID().String(),
		),
		traceID: traceID,
	}
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
