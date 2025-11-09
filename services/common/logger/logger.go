package logger

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
	// Sugar is the sugared logger for easier logging
	Sugar *zap.SugaredLogger
)

// Init initializes the global logger
func Init(serviceName string, isDevelopment bool) error {
	var config zap.Config
	
	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	
	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}
	
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}
	
	Log = logger
	Sugar = logger.Sugar()
	
	return nil
}

// InitDefault initializes with default development settings
func InitDefault(serviceName string) {
	if err := Init(serviceName, true); err != nil {
		// Fallback to basic logger
		Log = zap.NewExample()
		Sugar = Log.Sugar()
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// WithContext returns logger with trace context if available
func WithContext(ctx context.Context) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return Log
	}
	
	spanCtx := span.SpanContext()
	return Log.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}

// Infof logs a formatted info message (sugar)
func Infof(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Infof(template, args...)
	}
}

// Errorf logs a formatted error message (sugar)
func Errorf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Errorf(template, args...)
	}
}

// Warnf logs a formatted warning message (sugar)
func Warnf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Warnf(template, args...)
	}
}

// Debugf logs a formatted debug message (sugar)
func Debugf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Debugf(template, args...)
	}
}

// Fatalf logs a formatted fatal message and exits (sugar)
func Fatalf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Fatalf(template, args...)
	} else {
		os.Exit(1)
	}
}
