package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	"go.uber.org/zap"
)

const webPort = "80"

type Config struct {
	Rabbit      *amqp.Connection
	GRPCClients *GRPCClients
}

func main() {
	// Initialize logger
	logger.InitDefault("broker-service")
	defer logger.Sync()

	logger.Info("Starting broker service")

	// Initialize tracing
	shutdown, err := telemetry.InitTracer("broker-service", "1.0.0")
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

	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// Initialize gRPC clients
	grpcClients, err := InitGRPCClients()
	if err != nil {
		logger.Fatal("Failed to initialize gRPC clients", zap.Error(err))
		os.Exit(1)
	}

	app := Config{
		Rabbit:      rabbitConn,
		GRPCClients: grpcClients,
	}

	logger.Info("Starting HTTP server", zap.String("port", webPort))

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start the server
	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			logger.Warn("RabbitMQ not yet ready, retrying...",
				zap.Int64("attempt", counts+1),
				zap.Error(err),
			)
			counts++
		} else {
			logger.Info("Connected to RabbitMQ successfully")
			connection = c
			break
		}

		if counts > 5 {
			logger.Error("Failed to connect to RabbitMQ after 5 attempts", zap.Error(err))
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		logger.Debug("Backing off", zap.Duration("duration", backOff))
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
