package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	user_service "user-service/internal"
	"user-service/internal/configs"
	"user-service/internal/handlers"
	"user-service/internal/routes"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"

	"github.com/gin-gonic/gin"
)

const (
	serviceName    = "user-service"
	serviceVersion = "1.0.0"
	webPort        = "80"
)

type Config struct {
	Handlers *handlers.Handlers
}

func main() {
	// Initialize telemetry FIRST (sets up OTLP LoggerProvider)
	shutdown, err := telemetry.InitTracer(serviceName, serviceVersion)
	if err != nil {
		// Use basic println since logger not initialized yet
		fmt.Printf("Failed to initialize tracer: %v\n", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	// Initialize logger AFTER telemetry (to pick up OTLP provider)
	logger.InitDefault(serviceName)
	logger.Info("Starting User Service", "version", serviceVersion)

	// Initialize MongoDB
	logger.Info("Connecting to MongoDB...")
	err = configs.InitMongo()
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", "error", err)
	}
	defer configs.CloseMongo()

	// Initialize service and handlers
	ctx := context.Background()
	userService := user_service.NewUserService(configs.MongoClient)
	userHandlers := handlers.NewHandlers(&ctx, userService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Gin's default logger
	router.Use(gin.Logger())

	// Initialize routes
	routes.SetupRoutes(router, userHandlers)

	// Start gRPC server in goroutine
	go startGRPCServer(userService)

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
