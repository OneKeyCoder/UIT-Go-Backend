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

	"github.com/OneKeyCoder/UIT-Go-Backend/common/env"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/rabbitmq"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
)

const webPort = "80"

type Config struct {
	Models data.Models
}

var client *mongo.Client

func main() {
	// Initialize telemetry FIRST (sets up OTLP LoggerProvider)
	shutdown, err := telemetry.InitTracer("logger-service", "1.0.0")
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
	logger.InitDefault("logger-service")
	logger.Info("Starting logger service")

	mongoClient, err := connectToMongo()
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			logger.Error("Error disconnecting from MongoDB", "error", err)
		}
	}()

	client = mongoClient

	app := Config{
		Models: data.New(client),
	}

	// Connect to RabbitMQ
	rabbitConn, err := rabbitmq.ConnectSimple(env.RabbitMQURL())
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ, continuing without consumer", "error", err)
	} else {
		logger.Info("Connected to RabbitMQ")
		// Start RabbitMQ consumer in goroutine
		go func() {
			err := app.ConsumeFromRabbitMQ(rabbitConn)
			if err != nil {
				logger.Error("RabbitMQ consumer error", "error", err)
			}
		}()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := rabbitConn.Close(); err != nil {
				logger.Error("Error closing RabbitMQ connection", "error", err)
			}
			_ = ctx // avoid unused variable
		}()
	}

	go func() {
		err := app.StartGRPCServer()
		if err != nil {
			logger.Fatal("gRPC server failed", "error", err)
		}
	}()

	logger.Info("Starting HTTP server", "port", webPort)

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
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

func connectToMongo() (*mongo.Client, error) {
	mongoURL, needsAuth := resolveMongoURL()
	clientOptions := options.Client().ApplyURI(mongoURL)
	if needsAuth {
		username := env.Get("MONGO_USERNAME", "admin")
		password := env.Get("MONGO_PASSWORD", "password")
		if username != "" || password != "" {
			clientOptions.SetAuth(options.Credential{
				Username: username,
				Password: password,
			})
		}
	}

	// Connection pooling configuration
	clientOptions.SetMaxPoolSize(50)                   // Maximum connections
	clientOptions.SetMinPoolSize(10)                   // Minimum idle connections
	clientOptions.SetMaxConnIdleTime(30 * time.Second) // Close idle connections after 30s

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error("MongoDB connection failed", "error", err)
		return nil, err
	}

	logger.Info("Connected to MongoDB successfully")
	return c, nil
}

func resolveMongoURL() (string, bool) {
	for _, key := range []string{"MONGO_CONNECTION_STRING", "MONGO_URL"} {
		if uri := env.Get(key, ""); uri != "" {
			return uri, false
		}
	}

	host := env.Get("MONGO_HOST", "mongo")
	port := env.Get("MONGO_PORT", "27017")
	return fmt.Sprintf("mongodb://%s:%s", host, port), true
}
