package main

import (
	"context"
	"time"

	authpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/auth"
	loggerpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients holds all gRPC client connections
type GRPCClients struct {
	AuthClient   authpb.AuthServiceClient
	LoggerClient loggerpb.LoggerServiceClient
}

// InitGRPCClients initializes all gRPC client connections
func InitGRPCClients() (*GRPCClients, error) {
	// Connect to authentication service
	authConn, err := grpc.NewClient(
		"authentication-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcutil.UnaryClientInterceptor()),
	)
	if err != nil {
		logger.Error("Failed to connect to authentication service", zap.Error(err))
		return nil, err
	}

	// Connect to logger service
	loggerConn, err := grpc.NewClient(
		"logger-service:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcutil.UnaryClientInterceptor()),
	)
	if err != nil {
		logger.Error("Failed to connect to logger service", zap.Error(err))
		return nil, err
	}

	logger.Info("gRPC clients initialized",
		zap.String("auth_addr", "authentication-service:50051"),
		zap.String("logger_addr", "logger-service:50052"),
	)

	return &GRPCClients{
		AuthClient:   authpb.NewAuthServiceClient(authConn),
		LoggerClient: loggerpb.NewLoggerServiceClient(loggerConn),
	}, nil
}

// AuthenticateViaGRPC authenticates a user via gRPC
func (app *Config) AuthenticateViaGRPC(email, password string) (*authpb.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &authpb.AuthRequest{
		Email:    email,
		Password: password,
	}

	logger.Info("Calling authentication service via gRPC",
		zap.String("email", email),
	)

	resp, err := app.GRPCClients.AuthClient.Authenticate(ctx, req)
	if err != nil {
		logger.Error("gRPC authentication failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

// LogViaGRPC logs via gRPC to logger service
func (app *Config) LogViaGRPC(name, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &loggerpb.LogRequest{
		Name: name,
		Data: data,
	}

	logger.Info("Calling logger service via gRPC",
		zap.String("name", name),
	)

	_, err := app.GRPCClients.LoggerClient.WriteLog(ctx, req)
	if err != nil {
		logger.Error("gRPC logging failed", zap.Error(err))
		return err
	}

	return nil
}
