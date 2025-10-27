package main

import (
	"context"
	"strconv"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	authpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/auth"
	locationpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/location"
	loggerpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients holds all gRPC client connections
type GRPCClients struct {
	AuthClient     authpb.AuthServiceClient
	LoggerClient   loggerpb.LoggerServiceClient
	LocationClient locationpb.LocationServiceClient
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

	// Connect to location service
	locationConn, err := grpc.NewClient(
		"location-service:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcutil.UnaryClientInterceptor()),
	)
	if err != nil {
		logger.Error("Failed to connect to location service", zap.Error(err))
		return nil, err
	}

	logger.Info("gRPC clients initialized",
		zap.String("auth_addr", "authentication-service:50051"),
		zap.String("logger_addr", "logger-service:50052"),
		zap.String("location_addr", "location-service:50053"),
	)

	return &GRPCClients{
		AuthClient:     authpb.NewAuthServiceClient(authConn),
		LoggerClient:   loggerpb.NewLoggerServiceClient(loggerConn),
		LocationClient: locationpb.NewLocationServiceClient(locationConn),
	}, nil
}

// AuthenticateViaGRPC authenticates a user via gRPC
func (app *Config) AuthenticateViaGRPC(ctx context.Context, email, password string) (*authpb.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
func (app *Config) ValidateTokenViaGRPC(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authpb.ValidateTokenRequest{
		Token: token,
	}
	logger.Info("Calling authentication service ValidateToken via gRPC",
		zap.String("token", token),
	)

	resp, err := app.GRPCClients.AuthClient.ValidateToken(ctx, req)
	if err != nil {
		logger.Error("gRPC ValidateToken failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

// LogViaGRPC logs via gRPC to logger service
func (app *Config) LogViaGRPC(ctx context.Context, name, data string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

// SetLocationViaGRPC sets a user's location via gRPC
func (app *Config) SetLocationViaGRPC(ctx context.Context, userID int, role string, lat, lon, speed float64, heading, timestamp string) (*locationpb.SetLocationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.SetLocationRequest{
		UserId:    int32(userID),
		Role:      role,
		Latitude:  lat,
		Longitude: lon,
		Speed:     speed,
		Heading:   heading,
		Timestamp: timestamp,
	}

	logger.Info("Calling location service SetLocation via gRPC",
		zap.String("user_id", strconv.Itoa(userID)),
	)

	resp, err := app.GRPCClients.LocationClient.SetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC SetLocation failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

// GetLocationViaGRPC gets a user's location via gRPC
func (app *Config) GetLocationViaGRPC(ctx context.Context, userID int) (*locationpb.GetLocationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.GetLocationRequest{
		UserId: int32(userID),
	}

	logger.Info("Calling location service GetLocation via gRPC",
		zap.String("user_id", strconv.Itoa(userID)),
	)

	resp, err := app.GRPCClients.LocationClient.GetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC GetLocation failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

// FindNearestUsersViaGRPC finds nearest users via gRPC
func (app *Config) FindNearestUsersViaGRPC(ctx context.Context, userID int, topN int32, radius float64) (*locationpb.FindNearestUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.FindNearestUsersRequest{
		UserId: int32(userID),
		TopN:   topN,
		Radius: radius,
	}

	logger.Info("Calling location service FindNearestUsers via gRPC",
		zap.String("user_id", strconv.Itoa(userID)),
		zap.Int32("top_n", topN),
		zap.Float64("radius", radius),
	)

	resp, err := app.GRPCClients.LocationClient.FindNearestUsers(ctx, req)
	if err != nil {
		logger.Error("gRPC FindNearestUsers failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

// GetAllLocationsViaGRPC gets all locations via gRPC
func (app *Config) GetAllLocationsViaGRPC(ctx context.Context) (*locationpb.GetAllLocationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.GetAllLocationsRequest{}

	logger.Info("Calling location service GetAllLocations via gRPC")

	resp, err := app.GRPCClients.LocationClient.GetAllLocations(ctx, req)
	if err != nil {
		logger.Error("gRPC GetAllLocations failed", zap.Error(err))
		return nil, err
	}

	return resp, nil
}
