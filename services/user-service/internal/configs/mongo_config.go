package configs

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

var MongoClient *mongo.Client
var MongoDatabase *mongo.Database

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetMongoConfig() *MongoConfig {
	return &MongoConfig{
		Host:     getEnv("MONGO_HOST", "mongo"),
		Port:     getEnv("MONGO_PORT", "27017"),
		Database: getEnv("MONGO_DATABASE", "mongo"),
		Username: getEnv("MONGO_USERNAME", "admin"),
		Password: getEnv("MONGO_PASSWORD", "password"),
	}
}

func ConnectMongo() (*mongo.Client, error) {
	// Check if MONGO_URL is provided (for docker-compose)
	if mongoURL := os.Getenv("MONGO_URL"); mongoURL != "" {
		logger.Info("Using MONGO_URL from environment", zap.String("url", mongoURL))

		// Set client options
		clientOptions := options.Client().ApplyURI(mongoURL)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Connect to MongoDB
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
		}

		// Ping the database
		if err := client.Ping(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
		}

		logger.Info("Successfully connected to MongoDB via MONGO_URL")
		return client, nil
	}

	// Fallback to individual config
	config := GetMongoConfig()

	// Build connection URI
	var uri string
	if config.Username != "" && config.Password != "" {
		// With authentication
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	} else {
		// No authentication (default)
		uri = fmt.Sprintf("mongodb://%s:%s/%s",
			config.Host, config.Port, config.Database)
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("Successfully connected to MongoDB",
		zap.String("host", config.Host),
		zap.String("port", config.Port),
		zap.String("database", config.Database))

	return client, nil
}

func InitMongo() error {
	client, err := ConnectMongo()
	if err != nil {
		return err
	}

	config := GetMongoConfig()
	MongoClient = client
	MongoDatabase = client.Database(config.Database)

	return nil
}

func CloseMongo() error {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return MongoClient.Disconnect(ctx)
	}
	return nil
}

func GetMongoDatabase() *mongo.Database {
	return MongoDatabase
}
