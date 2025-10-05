package main

import (
	"context"
	"fmt"
	"logger-service/data"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const webPort = "80"

type Config struct {
	Models data.Models
}

var client *mongo.Client

func main() {
	logger.InitDefault("logger-service")
	defer logger.Sync()
	logger.Info("Starting logger service")

	shutdown, err := telemetry.InitTracer("logger-service", "1.0.0")
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

	mongoClient, err := connectToMongo()
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			logger.Error("Error disconnecting from MongoDB", zap.Error(err))
		}
	}()

	client = mongoClient

	app := Config{
		Models: data.New(client),
	}

	// Connect to RabbitMQ
	rabbitConn, err := connectToRabbitMQ()
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ, continuing without consumer", zap.Error(err))
	} else {
		logger.Info("Connected to RabbitMQ")
		// Start RabbitMQ consumer in goroutine
		go func() {
			err := app.ConsumeFromRabbitMQ(rabbitConn)
			if err != nil {
				logger.Error("RabbitMQ consumer error", zap.Error(err))
			}
		}()
		defer func() {
			if err := rabbitConn.Close(); err != nil {
				logger.Error("Error closing RabbitMQ connection", zap.Error(err))
			}
		}()
	}

	go func() {
		err := app.StartGRPCServer()
		if err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	logger.Info("Starting HTTP server", zap.String("port", webPort))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server gracefully...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func connectToMongo() (*mongo.Client, error) {
	mongoURL := "mongodb://mongo:27017"
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// Connection pooling configuration
	clientOptions.SetMaxPoolSize(50)                        // Maximum connections
	clientOptions.SetMinPoolSize(10)                        // Minimum idle connections
	clientOptions.SetMaxConnIdleTime(30 * time.Second)     // Close idle connections after 30s

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error("MongoDB connection failed", zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to MongoDB successfully")
	return c, nil
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	rabbitURL := "amqp://guest:guest@rabbitmq"
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
