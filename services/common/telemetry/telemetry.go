package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Tracer is the global tracer instance
	Tracer trace.Tracer
	// Logger is the global slog logger that exports to OTLP
	Logger *slog.Logger
)

// InitTracer initializes OpenTelemetry tracing with support for OTLP or stdout export
//
// Environment variables:
//   - OTEL_EXPORTER: "otlp" for OTLP, anything else for stdout
//   - OTEL_COLLECTOR_ENDPOINT: endpoint URL or host:port
//   - OTEL_EXPORTER_OTLP_HEADERS: optional headers for auth (e.g., "Authorization=Basic xxx")
//   - OTEL_INSECURE: "true" to disable TLS (for local development)
//   - REPLICA_ID: replica/pod identifier (e.g., pod name in Kubernetes)
func InitTracer(serviceName, serviceVersion string) (func(context.Context) error, error) {
	ctx := context.Background()
	
	// Get replica ID from environment (hostname in containers, pod name in K8s)
	replicaID := os.Getenv("REPLICA_ID")
	if replicaID == "" {
		// Fallback to hostname for local development
		hostname, err := os.Hostname()
		if err == nil {
			replicaID = hostname
		} else {
			replicaID = "unknown"
		}
	}
	
	// Create resource with service information + replica ID
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			// Custom attributes for Loki labels
			attribute.String("replica.id", replicaID),
			attribute.String("deployment.environment", getEnv("ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		return nil, err
	}

	// Determine exporter type from environment
	exporterType := os.Getenv("OTEL_EXPORTER")
	var tp *sdktrace.TracerProvider
	var lp *sdklog.LoggerProvider

	if exporterType == "otlp" {
		endpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")
		if endpoint == "" {
			endpoint = "alloy:4317" // Default endpoint for Alloy in docker-compose
		}

		// Check if endpoint is HTTPS (Grafana Cloud) or plain gRPC (local)
		if strings.HasPrefix(endpoint, "https://") {
			// Use HTTP exporter for Grafana Cloud (HTTPS endpoint)
			traceExporter, err := createHTTPTraceExporter(ctx, endpoint)
			if err != nil {
				return nil, err
			}
			tp = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter),
				sdktrace.WithResource(res),
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
			)

			// Create HTTP log exporter
			logExporter, err := createHTTPLogExporter(ctx, endpoint)
			if err != nil {
				return nil, err
			}
			lp = sdklog.NewLoggerProvider(
				sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
				sdklog.WithResource(res),
			)
		} else {
			// Use gRPC exporter for local Alloy/Jaeger
			traceExporter, err := createGRPCTraceExporter(ctx, endpoint)
			if err != nil {
				return nil, err
			}
			tp = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(traceExporter),
				sdktrace.WithResource(res),
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
			)

			// Create gRPC log exporter
			logExporter, err := createGRPCLogExporter(ctx, endpoint)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[TELEMETRY INIT] Failed to create gRPC log exporter: %v\n", err)
				return nil, err
			}
			fmt.Fprintf(os.Stderr, "[TELEMETRY INIT] Created gRPC log exporter for endpoint=%s\n", endpoint)
			lp = sdklog.NewLoggerProvider(
				sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
				sdklog.WithResource(res),
			)
		}
	} else {
		// Use stdout exporter for development
		exporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		// No log provider for stdout mode - use default slog
		lp = nil
	}

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagators for trace context propagation (CRITICAL for distributed tracing)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	Tracer = tp.Tracer(serviceName)

	// Set global log provider and create slog logger
	if lp != nil {
		global.SetLoggerProvider(lp)
		Logger = otelslog.NewLogger(serviceName)
		fmt.Fprintf(os.Stderr, "[TELEMETRY INIT] Set global LoggerProvider for service=%s\n", serviceName)
	} else {
		// Use default slog for stdout mode
		Logger = slog.Default()
		fmt.Fprintf(os.Stderr, "[TELEMETRY INIT] Using default slog (LoggerProvider is nil) for service=%s\n", serviceName)
	}

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if lp != nil {
			if err := lp.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	}

	return shutdown, nil
}

// createGRPCTraceExporter creates a gRPC OTLP trace exporter for local Alloy/Jaeger
func createGRPCTraceExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	// Check if insecure mode (local development)
	if os.Getenv("OTEL_INSECURE") == "true" || !strings.Contains(endpoint, ":443") {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPTraceExporter creates an HTTP OTLP trace exporter for Grafana Cloud
func createHTTPTraceExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	// Remove https:// prefix for the endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}

	// Add headers if provided (for Grafana Cloud auth)
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if headers != "" {
		headerMap := parseHeaders(headers)
		opts = append(opts, otlptracehttp.WithHeaders(headerMap))
	}

	// Check if insecure (http:// instead of https://)
	if os.Getenv("OTEL_INSECURE") == "true" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.New(ctx, opts...)
}

// createGRPCLogExporter creates a gRPC OTLP log exporter for local Alloy
func createGRPCLogExporter(ctx context.Context, endpoint string) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
	}

	// Check if insecure mode (local development)
	if os.Getenv("OTEL_INSECURE") == "true" || !strings.Contains(endpoint, ":443") {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	return otlploggrpc.New(ctx, opts...)
}

// createHTTPLogExporter creates an HTTP OTLP log exporter for Grafana Cloud
func createHTTPLogExporter(ctx context.Context, endpoint string) (sdklog.Exporter, error) {
	// Remove https:// prefix for the endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
	}

	// Add headers if provided (for Grafana Cloud auth)
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if headers != "" {
		headerMap := parseHeaders(headers)
		opts = append(opts, otlploghttp.WithHeaders(headerMap))
	}

	// Check if insecure (http:// instead of https://)
	if os.Getenv("OTEL_INSECURE") == "true" {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	return otlploghttp.New(ctx, opts...)
}

// parseHeaders parses header string like "Key1=Value1,Key2=Value2"
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	if Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return Tracer.Start(ctx, spanName)
}

// GetTraceID returns the trace ID from context if available
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
