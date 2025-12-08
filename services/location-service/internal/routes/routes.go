package routes

import (
	"location-service/internal/handlers"
	"net/http"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRoutes(router *gin.Engine, locationHandlers *handlers.Handlers) {
	// Apply shared metrics middleware to emit standard http_request_* metrics
	router.Use(commonMiddleware.GinPrometheusMetrics("location-service"))

	// Add OpenTelemetry instrumentation for Gin (skip /metrics and health endpoints)
	router.Use(otelgin.Middleware(
		"location-service",
		otelgin.WithFilter(func(req *http.Request) bool {
			return !commonMiddleware.ShouldSkipTrace(req.URL.Path)
		}),
	))

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
