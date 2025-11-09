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
	"go.uber.org/zap"
)

// "github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
// "github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"

const webPort = "80"

type Config struct {
	TripService *TripService
}

func main() {
	var app Config
	logger.InitDefault("trip-service")
	defer logger.Sync()
	logger.Info("Starting trip service")
	shutdown, err := telemetry.InitTracer("trip-service", "1.0.0")
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
	service := TripService{}
	service.InitializeServices()
	app.TripService = &service
	defer app.TripService.DB.Connection().Close()
	defer func() {
		if err := app.TripService.RabbitConn.Close(); err != nil {
			logger.Error("Error closing RabbitMQ connection", zap.Error(err))
		}
	}()
	go func() {
		err := app.StartGRPCServer()
		if err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	logger.Info("Starting HTTP server",
		zap.String("port", webPort),
	)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
