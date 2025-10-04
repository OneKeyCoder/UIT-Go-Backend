package main

import (
	"authentication-service/data"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/telemetry"
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

	// set up config
	app := Config{
		DB:            conn,
		Models:        data.New(conn),
		JWTSecret:     jwtSecret,
		JWTExpiry:     jwtExpiry,
		RefreshExpiry: refreshExpiry,
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

	err = srv.ListenAndServe()
	if err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

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
