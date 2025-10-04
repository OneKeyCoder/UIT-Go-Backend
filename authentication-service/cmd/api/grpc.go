package main

import (
	"context"
	"fmt"
	"net"

	pb "github.com/OneKeyCoder/UIT-Go-Backend/proto/auth"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/jwt"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the gRPC AuthService
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	Config *Config
}

// Authenticate handles user authentication
func (s *AuthServer) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	logger.Info("gRPC authentication request",
		zap.String("email", req.Email),
	)

	// Validate user credentials
	user, err := s.Config.Models.User.GetByEmail(req.Email)
	if err != nil {
		logger.Error("User not found", zap.String("email", req.Email), zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}

	// Check password
	valid, err := user.PasswordMatches(req.Password)
	if err != nil || !valid {
		logger.Error("Invalid password", zap.String("email", req.Email))
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}

	// Check if user is active
	if user.Active == 0 {
		logger.Error("Inactive user attempted login", zap.String("email", req.Email))
		return nil, status.Error(codes.PermissionDenied, "User account is inactive")
	}

	// Generate token pair
	tokens, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		"", // role - add if you have roles
		s.Config.JWTSecret,
		s.Config.JWTExpiry,
		s.Config.RefreshExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to generate tokens")
	}

	logger.Info("User authenticated successfully (gRPC)",
		zap.String("email", user.Email),
		zap.Int("user_id", user.ID),
	)

	return &pb.AuthResponse{
		Success: true,
		Message: "Authentication successful",
		User: &pb.User{
			Id:        int32(user.ID),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    int32(user.Active),
		},
		Tokens: &pb.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := jwt.ValidateToken(req.Token, s.Config.JWTSecret)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: fmt.Sprintf("%d", claims.UserID),
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	logger.Info("gRPC refresh token request")

	// Validate refresh token
	claims, err := jwt.ValidateToken(req.RefreshToken, s.Config.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}

	// Generate new token pair
	newTokens, err := jwt.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Role,
		s.Config.JWTSecret,
		s.Config.JWTExpiry,
		s.Config.RefreshExpiry,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate tokens")
	}

	return &pb.AuthResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Tokens: &pb.TokenPair{
			AccessToken:  newTokens.AccessToken,
			RefreshToken: newTokens.RefreshToken,
		},
	}, nil
}

// StartGRPCServer starts the gRPC server
func (app *Config) StartGRPCServer() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("Failed to listen for gRPC", zap.Error(err))
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcutil.UnaryServerInterceptor()),
	)

	authServer := &AuthServer{
		Config: app,
	}

	pb.RegisterAuthServiceServer(grpcServer, authServer)

	logger.Info("gRPC server started", zap.String("port", "50051"))

	return grpcServer.Serve(lis)
}
