package telemetry

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Tracer is the global tracer instance
	Tracer trace.Tracer
)

// InitTracer initializes OpenTelemetry tracing with support for OTLP or stdout export
// Set OTEL_EXPORTER=otlp and OTEL_COLLECTOR_ENDPOINT=host:port to use OTLP collector (Jaeger)
// Defaults to stdout for development
func InitTracer(serviceName, serviceVersion string) (func(context.Context) error, error) {
	ctx := context.Background()
	
	// Create resource with service information
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Determine exporter type from environment
	exporterType := os.Getenv("OTEL_EXPORTER")
	var tp *sdktrace.TracerProvider

	if exporterType == "otlp" {
		// Use OTLP exporter for production (Jaeger/OpenTelemetry Collector)
		endpoint := os.Getenv("OTEL_COLLECTOR_ENDPOINT")
		if endpoint == "" {
			endpoint = "jaeger:4317" // Default endpoint for Jaeger in docker-compose
		}

		// Create OTLP gRPC exporter
		exporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(), // Disable TLS for docker-compose
		)
		if err != nil {
			return nil, err
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
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
	}

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Create tracer
	Tracer = tp.Tracer(serviceName)

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}

	return shutdown, nil
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
