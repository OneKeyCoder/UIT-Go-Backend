package main

import (
	"context"
	"fmt"
	"logger-service/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
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

	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func connectToMongo() (*mongo.Client, error) {
	mongoURL := "mongodb://mongo:27017"
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error("MongoDB connection failed", zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to MongoDB successfully")
	return c, nil
}
