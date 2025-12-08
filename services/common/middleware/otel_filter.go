package middleware

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// skipTracePaths are paths that should not be traced (health checks, metrics)
var skipTracePaths = []string{
	"/metrics",
	"/health",
	"/healthz",
	"/ready",
	"/readiness",
	"/live",
	"/liveness",
}

// ShouldSkipTrace returns true if the path should not be traced
func ShouldSkipTrace(path string) bool {
	for _, skip := range skipTracePaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// OTelHTTPFilter is a custom span filter for OpenTelemetry HTTP instrumentation
// It prevents creating spans for health/metrics endpoints
type OTelHTTPFilter struct{}

// Filter returns true if the span should be created
func (f *OTelHTTPFilter) Filter(r *http.Request) bool {
	// Return false to skip creating span for metrics/health endpoints
	return !ShouldSkipTrace(r.URL.Path)
}

// SpanNameFormatter returns a custom span name
func (f *OTelHTTPFilter) SpanNameFormatter(operation string, r *http.Request) string {
	// For non-skipped requests, use default formatting
	return operation
}

// AttributesFromRequest extracts attributes from HTTP request
func AttributesFromRequest(r *http.Request) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPMethodKey.String(r.Method),
		semconv.HTTPTargetKey.String(r.URL.Path),
		semconv.HTTPSchemeKey.String(r.URL.Scheme),
	}

	if r.URL.RawQuery != "" {
		attrs = append(attrs, semconv.HTTPURLKey.String(r.URL.String()))
	}

	if ua := r.UserAgent(); ua != "" {
		attrs = append(attrs, semconv.HTTPUserAgentKey.String(ua))
	}

	return attrs
}

// ShouldSampleSpan determines if a span should be sampled
// Always sample non-metrics requests, never sample metrics
func ShouldSampleSpan(s trace.Span, r *http.Request) bool {
	if ShouldSkipTrace(r.URL.Path) {
		// Mark span as not sampled
		s.SetAttributes(attribute.Bool("sampled", false))
		return false
	}
	return true
}
