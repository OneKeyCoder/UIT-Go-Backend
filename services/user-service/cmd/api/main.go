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
	"go.uber.org/zap"
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
	// Initialize logger
	logger.InitDefault(serviceName)
	logger.Info("Starting User Service", zap.String("version", serviceVersion))

	// Initialize telemetry
	shutdown, err := telemetry.InitTracer(serviceName, serviceVersion)
	if err != nil {
		logger.Error("Failed to initialize tracer", zap.Error(err))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown tracer", zap.Error(err))
			}
		}()
	}

	// Initialize MongoDB
	logger.Info("Connecting to MongoDB...")
	err = configs.InitMongo()
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
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
		logger.Info("Starting HTTP server", zap.String("port", webPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
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
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
