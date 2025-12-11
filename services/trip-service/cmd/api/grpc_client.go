package main

import (
	"context"
	"strconv"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	locationpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/location"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	LocationClient locationpb.LocationServiceClient
}

func (grpcClients *GRPCClients) InitGRPCClients() (*GRPCClients, error) {
	locationConn, err := grpc.NewClient(
		"location-service:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to location service", "error", err)
		return nil, err
	}

	logger.Info("gRPC clients initialized",
		"location_addr", "location-service:50053",
	)
	return &GRPCClients{
		LocationClient: locationpb.NewLocationServiceClient(locationConn),
	}, nil
}

// SetLocationViaGRPC sets a user's location via gRPC
func (grpcClients *GRPCClients) SetLocationViaGRPC(ctx context.Context, userID int, lat, lon, speed float64, heading, timestamp string) (*locationpb.SetLocationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.SetLocationRequest{
		UserId:    int32(userID),
		Latitude:  lat,
		Longitude: lon,
		Speed:     speed,
		Heading:   heading,
		Timestamp: timestamp,
	}

	logger.Info("Calling location service SetLocation via gRPC",
		"user_id", strconv.Itoa(userID),
	)

	resp, err := grpcClients.LocationClient.SetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC SetLocation failed", "error", err)
		return nil, err
	}

	return resp, nil
}

// GetLocationViaGRPC gets a user's location via gRPC
func (grpcClients *GRPCClients) GetLocationViaGRPC(ctx context.Context, userID int) (*locationpb.GetLocationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.GetLocationRequest{
		UserId: int32(userID),
	}

	logger.Info("Calling location service GetLocation via gRPC",
		"user_id", strconv.Itoa(userID),
	)

	resp, err := grpcClients.LocationClient.GetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC GetLocation failed", "error", err)
		return nil, err
	}

	return resp, nil
}

// FindNearestUsersViaGRPC finds nearest users via gRPC
func (grpcClients *GRPCClients) FindNearestUsersViaGRPC(ctx context.Context, userID int, topN int32, radius float64) (*locationpb.FindNearestUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.FindNearestUsersRequest{
		UserId: int32(userID),
		TopN:   topN,
		Radius: radius,
	}

	logger.Info("Calling location service FindNearestUsers via gRPC",
		"user_id", strconv.Itoa(userID),
		"top_n", topN,
		"radius", radius,
	)

	resp, err := grpcClients.LocationClient.FindNearestUsers(ctx, req)
	if err != nil {
		logger.Error("gRPC FindNearestUsers failed", "error", err)
		return nil, err
	}

	return resp, nil
}

// GetAllLocationsViaGRPC gets all locations via gRPC
func (grpcClients *GRPCClients) GetAllLocationsViaGRPC(ctx context.Context) (*locationpb.GetAllLocationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &locationpb.GetAllLocationsRequest{}

	logger.Info("Calling location service GetAllLocations via gRPC")

	resp, err := grpcClients.LocationClient.GetAllLocations(ctx, req)
	if err != nil {
		logger.Error("gRPC GetAllLocations failed", "error", err)
		return nil, err
	}

	return resp, nil
}
