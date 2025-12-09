package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GinPrometheusMetrics instruments Gin handlers with the shared HTTP metrics vectors.
func GinPrometheusMetrics(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipMetrics(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		httpRequestsInFlight.WithLabelValues(serviceName).Inc()
		defer httpRequestsInFlight.WithLabelValues(serviceName).Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		routePattern := c.FullPath()
		if routePattern == "" {
			routePattern = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			routePattern,
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			serviceName,
			c.Request.Method,
			routePattern,
			status,
		).Observe(duration)
	}
}
