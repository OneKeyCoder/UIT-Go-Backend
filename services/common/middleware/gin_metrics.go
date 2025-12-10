package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// GinPrometheusMetrics instruments Gin handlers with OpenTelemetry metrics
func GinPrometheusMetrics(serviceName string) gin.HandlerFunc {
	initMetrics() // Ensure metrics are initialized

	return func(c *gin.Context) {
		if shouldSkipMetrics(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip if metrics not available
		if httpRequestsTotal == nil {
			c.Next()
			return
		}

		start := time.Now()
		ctx := c.Request.Context()

		routePattern := c.FullPath()
		if routePattern == "" {
			routePattern = c.Request.URL.Path
		}

		// Common attributes
		attrs := []attribute.KeyValue{
			attribute.String("service.name", serviceName),
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", routePattern),
		}

		// Track in-flight requests
		httpRequestsInFlight.Add(ctx, 1, metric.WithAttributes(attrs...))
		defer httpRequestsInFlight.Add(ctx, -1, metric.WithAttributes(attrs...))

		c.Next()

		// Record metrics with status code
		status := strconv.Itoa(c.Writer.Status())
		attrsWithStatus := append(attrs, attribute.String("http.status_code", status))

		httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrsWithStatus...))

		duration := time.Since(start).Seconds()
		httpRequestDuration.Record(ctx, duration, metric.WithAttributes(attrsWithStatus...))
	}
}
