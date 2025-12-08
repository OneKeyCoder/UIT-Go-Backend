package routes

import (
	"net/http"
	"user-service/internal/handlers"

	commonMiddleware "github.com/OneKeyCoder/UIT-Go-Backend/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRoutes(router *gin.Engine, userHandlers *handlers.Handlers) {
	// Apply shared metrics middleware so labels match Grafana dashboards
	router.Use(commonMiddleware.GinPrometheusMetrics("user-service"))

	// Add OpenTelemetry instrumentation for Gin (skip /metrics and health endpoints)
	router.Use(otelgin.Middleware(
		"user-service",
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
