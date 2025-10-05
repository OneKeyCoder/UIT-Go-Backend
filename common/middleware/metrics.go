package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"service"},
	)
)

// PrometheusMetrics returns a middleware that records HTTP metrics
func PrometheusMetrics(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Track in-flight requests
			httpRequestsInFlight.WithLabelValues(serviceName).Inc()
			defer httpRequestsInFlight.WithLabelValues(serviceName).Dec()

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(ww, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.statusCode)
			
			// Get route pattern from chi router
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			httpRequestsTotal.WithLabelValues(
				serviceName,
				r.Method,
				routePattern,
				status,
			).Inc()

			httpRequestDuration.WithLabelValues(
				serviceName,
				r.Method,
				routePattern,
				status,
			).Observe(duration)
		})
	}
}
