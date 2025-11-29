package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
)

const webPort = "80"

type Config struct {
	GRPCClients *GRPCClients
}

func main() {
	// Initialize tracing FIRST (sets up OTLP LoggerProvider)
	shutdown, err := telemetry.InitTracer("api-gateway", "1.0.0")
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
	logger.InitDefault("api-gateway")
	logger.Info("Starting API Gateway")

	// Initialize gRPC clients
	grpcClients, err := InitGRPCClients()
	if err != nil {
		logger.Fatal("Failed to initialize gRPC clients", "error", err)
		os.Exit(1)
	}

	app := Config{
		GRPCClients: grpcClients,
	}

	logger.Info("Starting HTTP server", "port", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server gracefully...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
