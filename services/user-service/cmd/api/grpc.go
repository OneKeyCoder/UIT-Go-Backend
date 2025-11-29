package main

import (
	"context"
	"fmt"
	"net"

	user_service "user-service/internal"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	pb "github.com/OneKeyCoder/UIT-Go-Backend/proto/user"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const grpcPort = "50055"

type UserServer struct {
	pb.UnimplementedUserServiceServer
	service *user_service.UserService
}

// ============================================
// User gRPC Methods
// ============================================

func (s *UserServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	logger.Info("gRPC GetUserById called", zap.Int32("user_id", req.UserId))

	user, err := s.service.GetUserById(ctx, int(req.UserId))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		return &pb.GetUserByIdResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.GetUserByIdResponse{
		Success: true,
		Message: "User retrieved successfully",
		User: &pb.User{
			UserId:           int32(user.UserId),
			Email:            user.Email,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			Role:             user.Role,
			DriverStatus:     user.DriverStatus,
			DriverTotalTrip:  int32(user.DriverTotalTrip),
			DriverRevenue:    user.DriverRevenue,
			DriverAvgRating:  user.DriverAvgRating,
			DriverVerifiedAt: user.DriverVerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *UserServer) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	logger.Info("gRPC GetAllUsers called")

	users, err := s.service.GetAllUsers(ctx)
	if err != nil {
		logger.Error("Failed to get all users", zap.Error(err))
		return &pb.GetAllUsersResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	pbUsers := make([]*pb.User, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.User{
			UserId:           int32(user.UserId),
			Email:            user.Email,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			Role:             user.Role,
			DriverStatus:     user.DriverStatus,
			DriverTotalTrip:  int32(user.DriverTotalTrip),
			DriverRevenue:    user.DriverRevenue,
			DriverAvgRating:  user.DriverAvgRating,
			DriverVerifiedAt: user.DriverVerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetAllUsersResponse{
		Success:    true,
		Message:    "All users retrieved successfully",
		Users:      pbUsers,
		TotalCount: int32(len(pbUsers)),
	}, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	logger.Info("gRPC CreateUser called", zap.String("email", req.Email))

	userReq := user_service.UserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	err := s.service.CreateUser(ctx, userReq)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return &pb.CreateUserResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.CreateUserResponse{
		Success: true,
		Message: "User created successfully",
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	logger.Info("gRPC UpdateUser called", zap.Int32("user_id", req.UserId))

	userRequest := user_service.UserRequest{
		Email:        req.Email,
		Role:         req.Role,
		DriverStatus: req.DriverStatus,
	}

	err := s.service.UpdateUserById(ctx, int(req.UserId), userRequest)
	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return &pb.UpdateUserResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.UpdateUserResponse{
		Success: true,
		Message: "User updated successfully",
	}, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	logger.Info("gRPC DeleteUser called", zap.Int32("user_id", req.UserId))

	err := s.service.DeleteUserById(ctx, int(req.UserId))
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return &pb.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ============================================
// Vehicle gRPC Methods
// ============================================

func (s *UserServer) GetVehicleById(ctx context.Context, req *pb.GetVehicleByIdRequest) (*pb.GetVehicleByIdResponse, error) {
	logger.Info("gRPC GetVehicleById called", zap.Int32("vehicle_id", req.VehicleId))

	vehicle, err := s.service.GetVehicleById(ctx, int(req.VehicleId))
	if err != nil {
		logger.Error("Failed to get vehicle", zap.Error(err))
		return &pb.GetVehicleByIdResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.GetVehicleByIdResponse{
		Success: true,
		Message: "Vehicle retrieved successfully",
		Vehicle: &pb.Vehicle{
			VehicleId:    int32(vehicle.VehicleId),
			DriverId:     int32(vehicle.DriverId),
			LicensePlate: vehicle.LicensePlate,
			VehicleType:  vehicle.VehicleType,
			Seats:        int32(vehicle.Seats),
			Status:       vehicle.Status,
			VerifiedAt:   vehicle.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:    vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *UserServer) GetVehiclesByUserId(ctx context.Context, req *pb.GetVehiclesByUserIdRequest) (*pb.GetVehiclesByUserIdResponse, error) {
	logger.Info("gRPC GetVehiclesByUserId called", zap.Int32("user_id", req.UserId))

	vehicles, err := s.service.GetVehiclesByUserId(ctx, int(req.UserId))
	if err != nil {
		logger.Error("Failed to get vehicles by user id", zap.Error(err))
		return &pb.GetVehiclesByUserIdResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	pbVehicles := make([]*pb.Vehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		pbVehicles = append(pbVehicles, &pb.Vehicle{
			VehicleId:    int32(vehicle.VehicleId),
			DriverId:     int32(vehicle.DriverId),
			LicensePlate: vehicle.LicensePlate,
			VehicleType:  vehicle.VehicleType,
			Seats:        int32(vehicle.Seats),
			Status:       vehicle.Status,
			VerifiedAt:   vehicle.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:    vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetVehiclesByUserIdResponse{
		Success:    true,
		Message:    "Vehicles retrieved successfully",
		Vehicles:   pbVehicles,
		TotalCount: int32(len(pbVehicles)),
	}, nil
}

func (s *UserServer) GetAllVehicles(ctx context.Context, req *pb.GetAllVehiclesRequest) (*pb.GetAllVehiclesResponse, error) {
	logger.Info("gRPC GetAllVehicles called")

	vehicles, err := s.service.GetAllVehicles(ctx)
	if err != nil {
		logger.Error("Failed to get all vehicles", zap.Error(err))
		return &pb.GetAllVehiclesResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	pbVehicles := make([]*pb.Vehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		pbVehicles = append(pbVehicles, &pb.Vehicle{
			VehicleId:    int32(vehicle.VehicleId),
			DriverId:     int32(vehicle.DriverId),
			LicensePlate: vehicle.LicensePlate,
			VehicleType:  vehicle.VehicleType,
			Seats:        int32(vehicle.Seats),
			Status:       vehicle.Status,
			VerifiedAt:   vehicle.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:    vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetAllVehiclesResponse{
		Success:    true,
		Message:    "All vehicles retrieved successfully",
		Vehicles:   pbVehicles,
		TotalCount: int32(len(pbVehicles)),
	}, nil
}

func (s *UserServer) CreateVehicle(ctx context.Context, req *pb.CreateVehicleRequest) (*pb.CreateVehicleResponse, error) {
	logger.Info("gRPC CreateVehicle called",
		zap.Int32("driver_id", req.DriverId),
		zap.String("license_plate", req.LicensePlate))

	vehicleRequest := user_service.VehicleRequest{
		LicensePlate: req.LicensePlate,
		VehicleType:  req.VehicleType,
		Seats:        int(req.Seats),
		Status:       req.Status,
	}

	vehicle, err := s.service.CreateVehicle(ctx, int(req.DriverId), vehicleRequest)
	if err != nil {
		logger.Error("Failed to create vehicle", zap.Error(err))
		return &pb.CreateVehicleResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.CreateVehicleResponse{
		Success: true,
		Message: "Vehicle created successfully",
		Vehicle: &pb.Vehicle{
			VehicleId:    int32(vehicle.VehicleId),
			DriverId:     int32(vehicle.DriverId),
			LicensePlate: vehicle.LicensePlate,
			VehicleType:  vehicle.VehicleType,
			Seats:        int32(vehicle.Seats),
			Status:       vehicle.Status,
			VerifiedAt:   vehicle.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:    vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *UserServer) UpdateVehicle(ctx context.Context, req *pb.UpdateVehicleRequest) (*pb.UpdateVehicleResponse, error) {
	logger.Info("gRPC UpdateVehicle called", zap.Int32("vehicle_id", req.VehicleId))

	vehicleRequest := user_service.VehicleRequest{
		LicensePlate: req.LicensePlate,
		VehicleType:  req.VehicleType,
		Seats:        int(req.Seats),
		Status:       req.Status,
	}

	vehicle, err := s.service.UpdateVehicle(ctx, int(req.VehicleId), vehicleRequest)
	if err != nil {
		logger.Error("Failed to update vehicle", zap.Error(err))
		return &pb.UpdateVehicleResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.UpdateVehicleResponse{
		Success: true,
		Message: "Vehicle updated successfully",
		Vehicle: &pb.Vehicle{
			VehicleId:    int32(vehicle.VehicleId),
			DriverId:     int32(vehicle.DriverId),
			LicensePlate: vehicle.LicensePlate,
			VehicleType:  vehicle.VehicleType,
			Seats:        int32(vehicle.Seats),
			Status:       vehicle.Status,
			VerifiedAt:   vehicle.VerifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:    vehicle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    vehicle.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *UserServer) DeleteVehicle(ctx context.Context, req *pb.DeleteVehicleRequest) (*pb.DeleteVehicleResponse, error) {
	logger.Info("gRPC DeleteVehicle called", zap.Int32("vehicle_id", req.VehicleId))

	err := s.service.DeleteVehicleById(ctx, int(req.VehicleId))
	if err != nil {
		logger.Error("Failed to delete vehicle", zap.Error(err))
		return &pb.DeleteVehicleResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.DeleteVehicleResponse{
		Success: true,
		Message: "Vehicle deleted successfully",
	}, nil
}

// ============================================
// gRPC Server Initialization
// ============================================

func startGRPCServer(userService *user_service.UserService) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpcutil.UnaryServerInterceptor()),
	)

	pb.RegisterUserServiceServer(s, &UserServer{
		service: userService,
	})

	logger.Info("Starting gRPC server", zap.String("port", grpcPort))

	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}
