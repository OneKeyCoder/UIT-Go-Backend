package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	location_service "location-service/internal"


	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	pb "github.com/OneKeyCoder/UIT-Go-Backend/proto/location"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const grpcPort = "50053"

type LocationServer struct {
	pb.UnimplementedLocationServiceServer
	service *location_service.LocationService
}

func (s *LocationServer) SetLocation(ctx context.Context, req *pb.SetLocationRequest) (*pb.SetLocationResponse, error) {
	logger.Info("gRPC SetLocation called", "user_id", strconv.Itoa(int(req.UserId)))

	location := &location_service.CurrentLocation{
		UserID:    int(req.UserId),
		Role:      req.Role,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Speed:     req.Speed,
		Heading:   req.Heading,
		Timestamp: req.Timestamp,
	}

	err := s.service.SetCurrentLocation(ctx, location)
	if err != nil {
		logger.Error("Failed to set location", "error", err)
		return &pb.SetLocationResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.SetLocationResponse{
		Success: true,
		Message: "Location updated successfully",
		Location: &pb.Location{
			UserId:    int32(location.UserID),
			Role:      location.Role,
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
			Speed:     location.Speed,
			Heading:   location.Heading,
			Timestamp: location.Timestamp,
		},
	}, nil
}

func (s *LocationServer) GetLocation(ctx context.Context, req *pb.GetLocationRequest) (*pb.GetLocationResponse, error) {
	logger.Info("gRPC GetLocation called", "user_id", strconv.Itoa(int(req.UserId)))

	location, err := s.service.GetCurrentLocation(ctx, int(req.UserId))
	if err != nil {
		logger.Error("Failed to get location", "error", err)
		return &pb.GetLocationResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	if location == nil {
		return &pb.GetLocationResponse{
			Success: false,
			Message: "Location not found",
		}, nil
	}

	return &pb.GetLocationResponse{
		Success: true,
		Message: "Location retrieved successfully",
		Location: &pb.Location{
			UserId:    int32(location.UserID),
			Role:      location.Role,
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
			Speed:     location.Speed,
			Heading:   location.Heading,
			Timestamp: location.Timestamp,
			Distance:  location.Distance,
		},
	}, nil
}

func (s *LocationServer) FindNearestUsers(ctx context.Context, req *pb.FindNearestUsersRequest) (*pb.FindNearestUsersResponse, error) {
	logger.Info("gRPC FindNearestUsers called",
		"user_id", strconv.Itoa(int(req.UserId)),
		"top_n", req.TopN,
		"radius", req.Radius)

	topN := int(req.TopN)
	if topN <= 0 {
		topN = 10
	}

	radius := req.Radius
	if radius <= 0 {
		radius = 10.0
	}

	locations, err := s.service.FindTopNearestUsers(ctx, int(req.UserId), topN, radius)
	if err != nil {
		logger.Error("Failed to find nearest users", "error", err)
		return &pb.FindNearestUsersResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	pbLocations := make([]*pb.Location, 0, len(locations))
	for _, loc := range locations {
		pbLocations = append(pbLocations, &pb.Location{
			UserId:    int32(loc.UserID),
			Role:      loc.Role,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Speed:     loc.Speed,
			Heading:   loc.Heading,
			Timestamp: loc.Timestamp,
			Distance:  loc.Distance,
		})
	}

	return &pb.FindNearestUsersResponse{
		Success:   true,
		Message:   fmt.Sprintf("Found %d nearest users", len(pbLocations)),
		Locations: pbLocations,
	}, nil
}

func (s *LocationServer) GetAllLocations(ctx context.Context, req *pb.GetAllLocationsRequest) (*pb.GetAllLocationsResponse, error) {
	logger.Info("gRPC GetAllLocations called")

	locations, err := s.service.GetAllLocations(ctx)
	if err != nil {
		logger.Error("Failed to get all locations", "error", err)
		return &pb.GetAllLocationsResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	pbLocations := make([]*pb.Location, 0, len(locations))
	for _, loc := range locations {
		pbLocations = append(pbLocations, &pb.Location{
			UserId:    int32(loc.UserID),
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Speed:     loc.Speed,
			Heading:   loc.Heading,
			Timestamp: loc.Timestamp,
		})
	}

	return &pb.GetAllLocationsResponse{
		Success:    true,
		Message:    "All locations retrieved successfully",
		Locations:  pbLocations,
		TotalCount: int32(len(pbLocations)),
	}, nil
}

func startGRPCServer(locationService *location_service.LocationService) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", "error", err)
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterLocationServiceServer(s, &LocationServer{
		service: locationService,
	})

	logger.Info("Starting gRPC server", "port", grpcPort)

	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", "error", err)
	}
}
