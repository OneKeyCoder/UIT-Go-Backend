package main

import (
	"context"
	"fmt"
	"logger-service/data"
	"net"


	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	pb "github.com/OneKeyCoder/UIT-Go-Backend/proto/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const grpcPort = "50052"

type LoggerServer struct {
	pb.UnimplementedLoggerServiceServer
	Models data.Models
}

// StartGRPCServer starts the gRPC server for logger-service
func (app *Config) StartGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	pb.RegisterLoggerServiceServer(s, &LoggerServer{Models: app.Models})
	reflection.Register(s)

	logger.Info("Starting gRPC server", "port", grpcPort)
	return s.Serve(lis)
}

// WriteLog implements the WriteLog RPC method
func (l *LoggerServer) WriteLog(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	logger.Info("WriteLog called via gRPC",
		"name", req.GetName(),
		"data", req.GetData())

	logEntry := data.LogEntry{
		Name: req.GetName(),
		Data: req.GetData(),
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		logger.Error("Failed to insert log", "error", err)
		return &pb.LogResponse{
			Success: false,
			Message: "Failed to write log",
		}, err
	}

	return &pb.LogResponse{
		Success: true,
		Message: "Log written successfully",
	}, nil
}

// GetLogs implements the GetLogs RPC method
func (l *LoggerServer) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	logger.Info("GetLogs called via gRPC", "limit", req.GetLimit())

	logs, err := l.Models.LogEntry.All()
	if err != nil {
		logger.Error("Failed to get logs", "error", err)
		return nil, err
	}

	var pbLogs []*pb.LogEntry
	for _, log := range logs {
		pbLogs = append(pbLogs, &pb.LogEntry{
			Id:        log.ID,
			Name:      log.Name,
			Data:      log.Data,
			CreatedAt: log.CreatedAt.String(),
			UpdatedAt: log.UpdatedAt.String(),
		})
	}

	return &pb.GetLogsResponse{Logs: pbLogs}, nil
}
