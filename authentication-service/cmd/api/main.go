package main

import (
	"authentication-service/data"
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const webPort = "80"

var counts int64

type Config struct {
	DB            *sql.DB
	Models        data.Models
	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration
	RabbitConn    *amqp.Connection
}

func main() {
	// Initialize logger
	logger.InitDefault("authentication-service")
	defer logger.Sync()

	logger.Info("Starting authentication service")

	// Initialize tracing
	shutdown, err := telemetry.InitTracer("authentication-service", "1.0.0")
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

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		logger.Fatal("Cannot connect to database")
	}

	// Get JWT configuration from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
		logger.Warn("Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	jwtExpiry := 24 * time.Hour
	if expiry := os.Getenv("JWT_EXPIRY"); expiry != "" {
		if d, err := time.ParseDuration(expiry); err == nil {
			jwtExpiry = d
		}
	}

	refreshExpiry := 7 * 24 * time.Hour
	if expiry := os.Getenv("REFRESH_TOKEN_EXPIRY"); expiry != "" {
		if d, err := time.ParseDuration(expiry); err == nil {
			refreshExpiry = d
		}
	}

	// Connect to RabbitMQ
	rabbitConn, err := connectToRabbitMQ()
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ, continuing without events", zap.Error(err))
	} else {
		logger.Info("Connected to RabbitMQ")
		defer func() {
			if err := rabbitConn.Close(); err != nil {
				logger.Error("Error closing RabbitMQ connection", zap.Error(err))
			}
		}()
	}

	// set up config
	app := Config{
		DB:            conn,
		Models:        data.New(conn),
		JWTSecret:     jwtSecret,
		JWTExpiry:     jwtExpiry,
		RefreshExpiry: refreshExpiry,
		RabbitConn:    rabbitConn,
	}

	logger.Info("Starting HTTP server",
		zap.String("port", webPort),
		zap.Duration("jwt_expiry", jwtExpiry),
		zap.Duration("refresh_expiry", refreshExpiry),
	)

	go func() {
		err := app.StartGRPCServer()
		if err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Connection pooling configuration
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			logger.Warn("Postgres not yet ready, retrying...",
				zap.Int64("attempt", counts+1),
				zap.Error(err),
			)
			counts++
		} else {
			logger.Info("Connected to Postgres successfully")
			return connection
		}

		if counts > 10 {
			logger.Error("Failed to connect to Postgres after 10 attempts", zap.Error(err))
			return nil
		}

		logger.Debug("Backing off for two seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			logger.Info("RabbitMQ not yet ready...")
			counts++
		} else {
			logger.Info("Connected to rabbitmq")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		logger.Info("backing off...")
		time.Sleep(backOff)
		continue
	}
	return connection, nil
}
