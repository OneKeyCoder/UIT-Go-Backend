package main

import (
	"authentication-service/data"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Azure/go-amqp"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/env"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/rabbitmq"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
	userpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/user"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const webPort = "80"

var counts int64

type Config struct {
	DB            *sql.DB
	Models        data.Models
	JWTSecret     string
	JWTExpiry     time.Duration
	RefreshExpiry time.Duration
	RabbitConn    *amqp.Conn
	RabbitSession *amqp.Session // Reusable session for publishing
	UserClient    userpb.UserServiceClient
}

func main() {
	// Initialize tracing FIRST (sets up OTLP LoggerProvider)
	shutdown, err := telemetry.InitTracer("authentication-service", "1.0.0")
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
	logger.InitDefault("authentication-service")
	logger.Info("Starting authentication service")

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
	rabbitConn, err := rabbitmq.ConnectSimple(env.RabbitMQURL())
	var rabbitSession *amqp.Session
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ, continuing without events", "error", err)
	} else {
		logger.Info("Connected to RabbitMQ")

		// Create a reusable session for publishing (reduce overhead)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		rabbitSession, err = rabbitConn.NewSession(ctx, nil)
		cancel()
		if err != nil {
			logger.Error("Failed to create RabbitMQ session", "error", err)
		} else {
			logger.Info("Created RabbitMQ session for publishing")
		}

		defer func() {
			if rabbitSession != nil {
				if err := rabbitSession.Close(context.Background()); err != nil {
					logger.Error("Error closing RabbitMQ session", "error", err)
				}
			}
			if err := rabbitConn.Close(); err != nil {
				logger.Error("Error closing RabbitMQ connection", "error", err)
			}
		}()
	}

	// Connect to user-service via gRPC
	userConn, err := grpc.NewClient(
		"user-service:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to user-service", "error", err)
	} else {
		logger.Info("Connected to user-service gRPC")
		defer userConn.Close()
	}

	userClient := userpb.NewUserServiceClient(userConn)

	// set up config
	app := Config{
		DB:            conn,
		Models:        data.New(conn),
		JWTSecret:     jwtSecret,
		JWTExpiry:     jwtExpiry,
		RefreshExpiry: refreshExpiry,
		RabbitConn:    rabbitConn,
		RabbitSession: rabbitSession,
		UserClient:    userClient,
	}

	logger.Info("Starting HTTP server",
		"port", webPort,
		"jwt_expiry", jwtExpiry,
		"refresh_expiry", refreshExpiry,
	)

	go func() {
		err := app.StartGRPCServer()
		if err != nil {
			logger.Fatal("gRPC server failed", "error", err)
		}
	}()

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

func openDB(dsn string) (*sql.DB, error) {
	// Use otelsql for automatic PostgreSQL tracing
	db, err := otelsql.Open("pgx", dsn,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		return nil, err
	}

	// Register DB stats metrics
	if err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL)); err != nil {
		logger.Warn("Failed to register DB stats", "error", err)
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
				"attempt", counts+1,
				"error", err,
			)
			counts++
		} else {
			logger.Info("Connected to Postgres successfully")
			return connection
		}

		if counts > 10 {
			logger.Error("Failed to connect to Postgres after 10 attempts", "error", err)
			return nil
		}

		logger.Debug("Backing off for two seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
