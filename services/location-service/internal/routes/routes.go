package routes

import (
	"location-service/internal/handlers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "location_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "location_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// metricsMiddleware wraps Gin handlers with Prometheus metrics
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
	}
}

func SetupRoutes(router *gin.Engine, locationHandlers *handlers.Handlers) {
	// Apply metrics middleware to all routes
	router.Use(metricsMiddleware())

	// Health check endpoints
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Metrics endpoint - use promhttp.Handler()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Location endpoints
	router.POST("/location", func(c *gin.Context) {
		locationHandlers.SetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/location", func(c *gin.Context) {
		locationHandlers.GetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/location/nearest", func(c *gin.Context) {
		locationHandlers.FindNearestUsers(c.Writer, c.Request)
	})

	router.GET("/location/all", func(c *gin.Context) {
		locationHandlers.GetAllLocations(c.Writer, c.Request)
	})
}
