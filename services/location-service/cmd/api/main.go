package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	location_service "location-service/internal"
	"location-service/internal/configs"
	"location-service/internal/handlers"
	"location-service/internal/routes"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"

	"github.com/gin-gonic/gin"
)

const (
	serviceName    = "location-service"
	serviceVersion = "1.0.0"
	webPort        = "80"
)

type Config struct {
	Handlers *handlers.Handlers
}

func main() {
	// Initialize logger
	logger.InitDefault(serviceName)
	logger.Info("Starting Location Service", "version", serviceVersion)

	// Initialize telemetry
	shutdown, err := telemetry.InitTracer(serviceName, serviceVersion)
	if err != nil {
		logger.Error("Failed to initialize tracer", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	logger.Info("Connecting to Redis...")
	redisClient, err := configs.ConnectRedis()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	// Initialize service and handlers
	ctx := context.Background()
	locationService := location_service.NewLocationService(redisClient)
	locationHandlers := handlers.NewHandlers(&ctx, locationService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Gin's default logger
	router.Use(gin.Logger())

	// Note: For HTTP middleware compatibility, we'll handle logging in routes

	// Initialize routes
	routes.SetupRoutes(router, locationHandlers)

	// Start gRPC server in goroutine
	go startGRPCServer(locationService)

	// HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", webPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

func InitRoute(router *gin.Engine, locationHandlers *handlers.Handlers) {
	router.POST("/", func(c *gin.Context) {
		locationHandlers.SetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		locationHandlers.GetCurrentLocation(c.Writer, c.Request)
	})

	router.GET("/nearest", func(c *gin.Context) {
		locationHandlers.FindNearestUsers(c.Writer, c.Request)
	})

	router.GET("/all", func(c *gin.Context) {
		locationHandlers.GetAllLocations(c.Writer, c.Request)
	})
}
