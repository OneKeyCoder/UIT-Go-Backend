package routes

import (
	"net/http"
	"time"
	"user-service/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_http_request_duration_seconds",
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

func SetupRoutes(router *gin.Engine, userHandlers *handlers.Handlers) {
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

	// User endpoints
	router.GET("/users", func(c *gin.Context) {
		userHandlers.GetAllUsers(c.Writer, c.Request)
	})

	router.GET("/users/:id", func(c *gin.Context) {
		userHandlers.GetUserById(c.Writer, c.Request)
	})

	router.POST("/users", func(c *gin.Context) {
		userHandlers.CreateUser(c.Writer, c.Request)
	})

	router.PUT("/users/:id", func(c *gin.Context) {
		userHandlers.UpdateUser(c.Writer, c.Request)
	})

	router.DELETE("/users/:id", func(c *gin.Context) {
		userHandlers.DeleteUser(c.Writer, c.Request)
	})

	// Vehicle endpoints
	router.GET("/vehicles", func(c *gin.Context) {
		userHandlers.GetAllVehicles(c.Writer, c.Request)
	})

	router.GET("/vehicles/:id", func(c *gin.Context) {
		userHandlers.GetVehicleById(c.Writer, c.Request)
	})

	router.GET("/users/:id/vehicles", func(c *gin.Context) {
		userHandlers.GetVehiclesByUserId(c.Writer, c.Request)
	})

	router.POST("/vehicles", func(c *gin.Context) {
		userHandlers.CreateVehicle(c.Writer, c.Request)
	})

	router.PUT("/vehicles/:id", func(c *gin.Context) {
		userHandlers.UpdateVehicle(c.Writer, c.Request)
	})

	router.DELETE("/vehicles/:id", func(c *gin.Context) {
		userHandlers.DeleteVehicle(c.Writer, c.Request)
	})
}
