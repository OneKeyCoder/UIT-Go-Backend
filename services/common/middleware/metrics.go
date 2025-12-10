package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	metricsOnce          sync.Once
	httpRequestsTotal    metric.Int64Counter
	httpRequestDuration  metric.Float64Histogram
	httpRequestsInFlight metric.Int64UpDownCounter
)

// initMetrics initializes OpenTelemetry metrics (called once)
func initMetrics() {
	metricsOnce.Do(func() {
		if telemetry.Meter == nil {
			// Metrics not initialized - skip
			return
		}

		var err error
		httpRequestsTotal, err = telemetry.Meter.Int64Counter(
			"http.server.request.count",
			metric.WithDescription("Total number of HTTP requests"),
			metric.WithUnit("{request}"),
		)
		if err != nil {
			panic(err)
		}

		httpRequestDuration, err = telemetry.Meter.Float64Histogram(
			"http.server.request.duration",
			metric.WithDescription("HTTP request latency"),
			metric.WithUnit("s"),
		)
		if err != nil {
			panic(err)
		}

		httpRequestsInFlight, err = telemetry.Meter.Int64UpDownCounter(
			"http.server.request.active",
			metric.WithDescription("Number of active HTTP requests"),
			metric.WithUnit("{request}"),
		)
		if err != nil {
			panic(err)
		}
	})
}

// shouldSkipMetrics returns true if the path should not be recorded in metrics
func shouldSkipMetrics(path string) bool {
	skipPaths := []string{"/metrics", "/health", "/healthz", "/ready", "/live"}
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// PrometheusMetrics returns a middleware that records HTTP metrics
func PrometheusMetrics(serviceName string) func(next http.Handler) http.Handler {
	initMetrics() // Ensure metrics are initialized

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics recording for /metrics and /health endpoints
			if shouldSkipMetrics(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip if metrics not available
			if httpRequestsTotal == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ctx := r.Context()

			// Get route pattern from chi router
			routePattern := chi.RouteContext(ctx).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			// Common attributes
			attrs := []attribute.KeyValue{
				attribute.String("service.name", serviceName),
				attribute.String("http.method", r.Method),
				attribute.String("http.route", routePattern),
			}

			// Track in-flight requests
			httpRequestsInFlight.Add(ctx, 1, metric.WithAttributes(attrs...))
			defer httpRequestsInFlight.Add(ctx, -1, metric.WithAttributes(attrs...))

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(ww, r)

			// Record metrics with status code
			status := strconv.Itoa(ww.statusCode)
			attrsWithStatus := append(attrs, attribute.String("http.status_code", status))

			httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrsWithStatus...))
			
			duration := time.Since(start).Seconds()
			httpRequestDuration.Record(ctx, duration, metric.WithAttributes(attrsWithStatus...))
		})
	}
}
