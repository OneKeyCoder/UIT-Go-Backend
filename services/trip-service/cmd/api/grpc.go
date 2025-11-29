package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"trip-service/internal/models"
	"trip-service/internal/repository"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/grpcutil"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	pb "github.com/OneKeyCoder/UIT-Go-Backend/proto/trip"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TripServer struct {
	pb.UnimplementedTripServiceServer
	Config *Config
}

func (s *TripServer) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	logger.Info("Create Trip via gRPC",
		"userID", strconv.Itoa(int(req.PassengerId)),
	)
	newTrip := repository.NewTripDTO{
		PassengerID:   int(req.PassengerId),
		OriginLat:     req.OriginLat,
		OriginLng:     req.OriginLng,
		DestLat:       req.DestLat,
		DestLng:       req.DestLng,
		PaymentMethod: req.PaymentMethod,
	}
	tripRecord, duration, err := s.Config.TripService.CreateTrip(newTrip)
	if err != nil {
		logger.Error("Failed to create trip via gRPC", "error", err)
		return nil, err
	}
	var status pb.TripStatus
	if v, ok := pb.TripStatus_value[string(tripRecord.Status)]; ok {
		status = pb.TripStatus(v)
	} else {
		status = pb.TripStatus(0)
	}
	var driverID int32
	if tripRecord.DriverID.Valid {
		driverID = int32(tripRecord.DriverID.Int32)
	} else {
		driverID = int32(0)
	}
	return &pb.CreateTripResponse{
		Trip: &pb.Trip{
			Id:            int32(tripRecord.ID),
			PassengerId:   int32(tripRecord.PassengerID),
			DriverId:      driverID,
			OriginLat:     tripRecord.OriginLat,
			OriginLng:     tripRecord.OriginLng,
			DestLat:       tripRecord.DestLat,
			DestLng:       tripRecord.DestLng,
			Distance:      tripRecord.Distance,
			Status:        status,
			Fare:          tripRecord.Fare,
			PaymentMethod: tripRecord.PaymentMethod,
		},
		Duration: float32(duration),
	}, nil
}

func (s *TripServer) AcceptTrip(ctx context.Context, req *pb.AcceptTripRequest) (*pb.MessageResponse, error) {
	logger.Info("Accept Trip via gRPC",
		"driverID", strconv.Itoa(int(req.DriverId)),
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	err := s.Config.TripService.AcceptTrip(int(req.DriverId), int(req.TripId))
	if err != nil {
		logger.Error("Failed to accept trip via gRPC", "error", err)
		return nil, err
	}
	return &pb.MessageResponse{
		Success: true,
		Message: fmt.Sprintf("Trip %d accepted by driver %d successfully", req.TripId, req.DriverId),
	}, nil
}

func (s *TripServer) RejectTrip(ctx context.Context, req *pb.RejectTripRequest) (*pb.MessageResponse, error) {
	logger.Info("Reject Trip via gRPC",
		"driverID", strconv.Itoa(int(req.DriverId)),
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	err := s.Config.TripService.RejectTrip(int(req.PassengerId), int(req.DriverId), int(req.TripId))
	if err != nil {
		logger.Error("Failed to reject trip via gRPC", "error", err)
		return nil, err
	}
	return &pb.MessageResponse{
		Success: true,
		Message: fmt.Sprintf("Trip %d rejected by driver %d successfully", req.TripId, req.DriverId),
	}, nil
}

func (s *TripServer) GetSuggestedDriver(ctx context.Context, req *pb.TripIDRequest) (*pb.GetSuggestedDriverResponse, error) {
	logger.Info("Get Suggested Driver via gRPC",
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	driverID, err := s.Config.TripService.GetSuggestedDriver(int(req.TripId))
	if err != nil {
		logger.Error("Failed to get suggested driver via gRPC", "error", err)
		return nil, err
	}
	return &pb.GetSuggestedDriverResponse{
		DriverId: int32(driverID),
	}, nil
}

func (s *TripServer) GetTripDetail(ctx context.Context, req *pb.TripIDRequest) (*pb.GetTripDetailResponse, error) {
	logger.Info("Get Trip via gRPC",
		"userID", strconv.Itoa(int(req.PassengerId)),
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	tripRecord, err := s.Config.TripService.GetTrip(int(req.PassengerId), int(req.TripId))
	if err != nil {
		logger.Error("Failed to get trip via gRPC", "error", err)
		return nil, err
	}
	var driver int
	if tripRecord.DriverID.Valid {
		driver = int(tripRecord.DriverID.Int32)
	} else {
		driver = 0
	}
	return &pb.GetTripDetailResponse{
		Trip: &pb.Trip{
			Id:            int32(tripRecord.ID),
			PassengerId:   int32(tripRecord.PassengerID),
			DriverId:      int32(driver),
			OriginLat:     tripRecord.OriginLat,
			OriginLng:     tripRecord.OriginLng,
			DestLat:       tripRecord.DestLat,
			DestLng:       tripRecord.DestLng,
			Status:        pb.TripStatus(pb.TripStatus_value[string(tripRecord.Status)]),
			Distance:      tripRecord.Distance,
			Fare:          tripRecord.Fare,
			PaymentMethod: tripRecord.PaymentMethod,
			CreatedAt:     timestamppb.New(tripRecord.CreatedAt),
			UpdatedAt:     timestamppb.New(tripRecord.UpdatedAt),
		},
	}, nil
}

func (s *TripServer) GetTripsByPassenger(ctx context.Context, req *pb.GetTripsByUserIDRequest) (*pb.TripsResponse, error) {
	logger.Info("Get Trips By Passenger via gRPC",
		"passengerID", strconv.Itoa(int(req.UserId)),
	)
	trips, err := s.Config.TripService.GetTripsByPassenger(int(req.UserId))
	if err != nil {
		logger.Error("Failed to get trips by passenger via gRPC", "error", err)
		return nil, err
	}
	var pbTrips []*pb.Trip
	for _, tripRecord := range trips {
		var driver int
		if tripRecord.DriverID.Valid {
			driver = int(tripRecord.DriverID.Int32)
		} else {
			driver = 0
		}
		pbTrip := &pb.Trip{
			Id:            int32(tripRecord.ID),
			PassengerId:   int32(tripRecord.PassengerID),
			DriverId:      int32(driver),
			OriginLat:     tripRecord.OriginLat,
			OriginLng:     tripRecord.OriginLng,
			DestLat:       tripRecord.DestLat,
			DestLng:       tripRecord.DestLng,
			Status:        pb.TripStatus(pb.TripStatus_value[string(tripRecord.Status)]),
			Distance:      tripRecord.Distance,
			Fare:          tripRecord.Fare,
			PaymentMethod: tripRecord.PaymentMethod,
			CreatedAt:     timestamppb.New(tripRecord.CreatedAt),
			UpdatedAt:     timestamppb.New(tripRecord.UpdatedAt),
		}
		pbTrips = append(pbTrips, pbTrip)
	}
	return &pb.TripsResponse{
		Trips: pbTrips,
	}, nil
}

func (s *TripServer) GetTripsByDriver(ctx context.Context, req *pb.GetTripsByUserIDRequest) (*pb.TripsResponse, error) {
	logger.Info("Get Trips By Driver via gRPC",
		"driverID", strconv.Itoa(int(req.UserId)),
	)
	trips, err := s.Config.TripService.GetTripsByDriver(int(req.UserId))
	if err != nil {
		logger.Error("Failed to get trips by driver via gRPC", "error", err)
		return nil, err
	}
	var pbTrips []*pb.Trip
	for _, tripRecord := range trips {
		var driver int
		if tripRecord.DriverID.Valid {
			driver = int(tripRecord.DriverID.Int32)
		} else {
			driver = 0
		}
		pbTrip := &pb.Trip{
			Id:            int32(tripRecord.ID),
			PassengerId:   int32(tripRecord.PassengerID),
			DriverId:      int32(driver),
			OriginLat:     tripRecord.OriginLat,
			OriginLng:     tripRecord.OriginLng,
			DestLat:       tripRecord.DestLat,
			DestLng:       tripRecord.DestLng,
			Status:        pb.TripStatus(pb.TripStatus_value[string(tripRecord.Status)]),
			Fare:          tripRecord.Fare,
			Distance:      tripRecord.Distance,
			PaymentMethod: tripRecord.PaymentMethod,
			CreatedAt:     timestamppb.New(tripRecord.CreatedAt),
			UpdatedAt:     timestamppb.New(tripRecord.UpdatedAt),
		}
		pbTrips = append(pbTrips, pbTrip)
	}
	return &pb.TripsResponse{
		Trips: pbTrips,
	}, nil
}

func (s *TripServer) GetAllTrips(ctx context.Context, req *pb.GetAllTripsRequest) (*pb.PageResponse, error) {
	logger.Info("Get All Trips via gRPC",
		"page", strconv.Itoa(int(req.Page)),
		"limit", strconv.Itoa(int(req.Limit)),
	)
	trips, err := s.Config.TripService.GetAllTrips(int(req.Page), int(req.Limit))
	if err != nil {
		logger.Error("Failed to get all trips via gRPC", "error", err)
		return nil, err
	}
	var pbTrips []*pb.Trip
	for _, tripRecord := range trips {
		var driver int
		if tripRecord.DriverID.Valid {
			driver = int(tripRecord.DriverID.Int32)
		} else {
			driver = 0
		}
		pbTrip := &pb.Trip{
			Id:            int32(tripRecord.ID),
			PassengerId:   int32(tripRecord.PassengerID),
			DriverId:      int32(driver),
			OriginLat:     tripRecord.OriginLat,
			OriginLng:     tripRecord.OriginLng,
			DestLat:       tripRecord.DestLat,
			DestLng:       tripRecord.DestLng,
			Status:        pb.TripStatus(pb.TripStatus_value[string(tripRecord.Status)]),
			Distance:      tripRecord.Distance,
			Fare:          tripRecord.Fare,
			PaymentMethod: tripRecord.PaymentMethod,
		}
		pbTrips = append(pbTrips, pbTrip)
	}
	return &pb.PageResponse{
		Trips: pbTrips,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (s *TripServer) UpdateTripStatus(ctx context.Context, req *pb.UpdateTripStatusRequest) (*pb.MessageResponse, error) {
	logger.Info("Update Trip Status via gRPC",
		"tripID", strconv.Itoa(int(req.TripId)),
		"tripStatus", req.Status.String(),
	)
	var status models.TripStatus
	if v, ok := pb.TripStatus_value[req.Status.String()]; ok {
		status = models.TripStatus(v)
	} else {
		err := fmt.Errorf("invalid trip status: %s", req.Status.String())
		logger.Error("Failed to update trip status via gRPC", "error", err)
		return nil, err
	}
	err := s.Config.TripService.UpdateTripStatus(status, int(req.TripId), int(req.DriverId))
	if err != nil {
		logger.Error("Failed to update trip status via gRPC", "error", err)
		return nil, err
	}
	return &pb.MessageResponse{
		Success: true,
		Message: fmt.Sprintf("Trip %d status updated to %s successfully", req.TripId, req.Status.String()),
	}, nil
}

func (s *TripServer) CancelTrip(ctx context.Context, req *pb.CancelTripRequest) (*pb.MessageResponse, error) {
	logger.Info("Cancel Trip via gRPC",
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	err := s.Config.TripService.CancelTrip(int(req.UserId), int(req.TripId))
	if err != nil {
		logger.Error("Failed to cancel trip via gRPC", "error", err)
		return nil, err
	}
	return &pb.MessageResponse{
		Message: fmt.Sprintf("Trip %d cancelled successfully", req.TripId),
	}, nil
}

func (s *TripServer) ReviewTrip(ctx context.Context, req *pb.SubmitReviewRequest) (*pb.MessageResponse, error) {
	logger.Info("Review Trip via gRPC",
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	review := repository.ReviewDTO{
		PassengerID: int(req.UserId),
		Comment:     req.Review.Comment,
		Rating:      int(req.Review.Rating),
	}
	err := s.Config.TripService.ReviewTrip(int(req.UserId), int(req.TripId), review)
	if err != nil {
		logger.Error("Failed to review trip via gRPC", "error", err)
		return nil, err
	}
	return &pb.MessageResponse{
		Success: true,
		Message: fmt.Sprintf("Trip %d reviewed successfully", req.TripId),
	}, nil
}

func (s *TripServer) GetReview(ctx context.Context, req *pb.TripIDRequest) (*pb.GetTripReviewResponse, error) {
	logger.Info("Get Review via gRPC",
		"userID", strconv.Itoa(int(req.PassengerId)),
		"tripID", strconv.Itoa(int(req.TripId)),
	)
	review, err := s.Config.TripService.GetReview(int(req.TripId), int(req.PassengerId))
	if err != nil {
		logger.Error("Failed to get review via gRPC", "error", err)
		return nil, err
	}
	return &pb.GetTripReviewResponse{
		Review: &pb.Review{
			Comment: review.Comment,
			Rating:  int32(review.Rating),
		},
	}, nil
}

func (app *Config) StartGRPCServer() error {
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		logger.Error("Failed to listen for gRPC", "error", err)
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcutil.UnaryServerInterceptor()),
	)

	tripServer := &TripServer{
		Config: app,
	}

	pb.RegisterTripServiceServer(grpcServer, tripServer)

	logger.Info("gRPC server started", "port", "50054")

	return grpcServer.Serve(lis)
}
