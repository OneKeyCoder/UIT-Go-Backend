package main

import (
	"context"
	"strconv"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	authpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/auth"
	locationpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/location"
	loggerpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/logger"
	trippb "github.com/OneKeyCoder/UIT-Go-Backend/proto/trip"
	userpb "github.com/OneKeyCoder/UIT-Go-Backend/proto/user"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients holds all gRPC client connections
type GRPCClients struct {
	AuthClient     authpb.AuthServiceClient
	LoggerClient   loggerpb.LoggerServiceClient
	LocationClient locationpb.LocationServiceClient
	TripClient     trippb.TripServiceClient
	UserClient     userpb.UserServiceClient
}

// InitGRPCClients initializes all gRPC client connections
func InitGRPCClients() (*GRPCClients, error) {
	// Connect to authentication service
	authConn, err := grpc.NewClient(
		"authentication-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to authentication service", "error", err)
		return nil, err
	}

	// Connect to logger service
	loggerConn, err := grpc.NewClient(
		"logger-service:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to logger service", "error", err)
		return nil, err
	}

	// Connect to location service
	locationConn, err := grpc.NewClient(
		"location-service:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to location service", "error", err)
		return nil, err
	}

	tripConn, err := grpc.NewClient(
		"trip-service:50054",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to trip service", "error", err)
		return nil, err
	}

	// Connect to user service
	userConn, err := grpc.NewClient(
		"user-service:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logger.Error("Failed to connect to user service", "error", err)
		return nil, err
	}

	logger.Info("gRPC clients initialized",
		"auth_addr", "authentication-service:50051",
		"logger_addr", "logger-service:50052",
		"location_addr", "location-service:50053",
		"trip_addr", "trip-service:50054",
		"user_addr", "user-service:50055",
	)

	return &GRPCClients{
		AuthClient:     authpb.NewAuthServiceClient(authConn),
		LoggerClient:   loggerpb.NewLoggerServiceClient(loggerConn),
		LocationClient: locationpb.NewLocationServiceClient(locationConn),
		TripClient:     trippb.NewTripServiceClient(tripConn),
		UserClient:     userpb.NewUserServiceClient(userConn),
	}, nil
}

// RegisterViaGRPC registers a new user via gRPC
func (app *Config) RegisterViaGRPC(ctx context.Context, email, password, firstName, lastName string) (*authpb.RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authpb.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	logger.Info("Calling authentication service Register via gRPC",
		"email", email,
	)

	resp, err := app.GRPCClients.AuthClient.Register(ctx, req)
	if err != nil {
		logger.Error("gRPC Register failed", "error", err)
		return nil, err
	}

	return resp, nil
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
		"email", email,
	)

	resp, err := app.GRPCClients.AuthClient.Authenticate(ctx, req)
	if err != nil {
		logger.Error("gRPC authentication failed", "error", err)
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
		"token", token,
	)

	resp, err := app.GRPCClients.AuthClient.ValidateToken(ctx, req)
	if err != nil {
		logger.Error("gRPC ValidateToken failed", "error", err)
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
		"name", name,
	)

	_, err := app.GRPCClients.LoggerClient.WriteLog(ctx, req)
	if err != nil {
		logger.Error("gRPC logging failed", "error", err)
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
		"user_id", strconv.Itoa(userID),
	)

	resp, err := app.GRPCClients.LocationClient.SetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC SetLocation failed", "error", err)
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
		"user_id", strconv.Itoa(userID),
	)

	resp, err := app.GRPCClients.LocationClient.GetLocation(ctx, req)
	if err != nil {
		logger.Error("gRPC GetLocation failed", "error", err)
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
		"user_id", strconv.Itoa(userID),
		"top_n", topN,
		"radius", radius,
	)

	resp, err := app.GRPCClients.LocationClient.FindNearestUsers(ctx, req)
	if err != nil {
		logger.Error("gRPC FindNearestUsers failed", "error", err)
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
		logger.Error("gRPC GetAllLocations failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) CreateTripViaGRPC(ctx context.Context, passengerID int, originLat float64, originLng float64, DestLat float64, DestLng float64, PaymentMethod string) (*trippb.CreateTripResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.CreateTripRequest{
		PassengerId:   int32(passengerID),
		OriginLat:     originLat,
		OriginLng:     originLng,
		DestLat:       DestLat,
		DestLng:       DestLng,
		PaymentMethod: PaymentMethod,
	}

	logger.Info("Calling trip service CreateTrip via gRPC",
		"user_id", strconv.Itoa(passengerID),
	)

	resp, err := app.GRPCClients.TripClient.CreateTrip(ctx, req)
	if err != nil {
		logger.Error("gRPC CreateTrip failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) AcceptTripViaGRPC(ctx context.Context, driverID int, tripID int) (*trippb.MessageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.AcceptTripRequest{
		DriverId: int32(driverID),
		TripId:   int32(tripID),
	}
	logger.Info("Calling trip service AcceptTrip via gRPC",
		"driver_id", strconv.Itoa(driverID),
		"trip_id", strconv.Itoa(tripID),
	)

	resp, err := app.GRPCClients.TripClient.AcceptTrip(ctx, req)
	if err != nil {
		logger.Error("gRPC AcceptTrip failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) RejectTripViaGRPC(ctx context.Context, driverID int, tripID int) (*trippb.MessageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.RejectTripRequest{
		DriverId: int32(driverID),
		TripId:   int32(tripID),
	}
	logger.Info("Calling trip service RejectTrip via gRPC",
		"driver_id", strconv.Itoa(driverID),
		"trip_id", strconv.Itoa(tripID),
	)

	resp, err := app.GRPCClients.TripClient.RejectTrip(ctx, req)
	if err != nil {
		logger.Error("gRPC RejectTrip failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) GetSuggestedDriverViaGRPC(ctx context.Context, tripID int) (*trippb.GetSuggestedDriverResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.TripIDRequest{
		TripId: int32(tripID),
	}
	logger.Info("Calling trip service GetSuggestedDriver via gRPC",
		"trip_id", strconv.Itoa(tripID),
	)

	resp, err := app.GRPCClients.TripClient.GetSuggestedDriver(ctx, req)
	if err != nil {
		logger.Error("gRPC GetSuggestedDriver failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) GetTripDetailViaGRPC(ctx context.Context, tripID int, userID int) (*trippb.GetTripDetailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.TripIDRequest{
		TripId:      int32(tripID),
		PassengerId: int32(userID),
	}
	logger.Info("Calling trip service GetTripDetail via gRPC",
		"trip_id", strconv.Itoa(tripID),
		"user_id", strconv.Itoa(userID),
	)

	resp, err := app.GRPCClients.TripClient.GetTripDetail(ctx, req)
	if err != nil {
		logger.Error("gRPC GetTripDetail failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) GetTripsByPassengerViaGRPC(ctx context.Context, passengerID int) (*trippb.TripsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.GetTripsByUserIDRequest{
		UserId: int32(passengerID),
	}
	logger.Info("Calling trip service GetTripsByPassenger via gRPC",
		"passenger_id", strconv.Itoa(passengerID),
	)

	resp, err := app.GRPCClients.TripClient.GetTripsByPassenger(ctx, req)
	if err != nil {
		logger.Error("gRPC GetTripsByPassenger failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) GetTripsByDriverViaGRPC(ctx context.Context, driverID int) (*trippb.TripsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.GetTripsByUserIDRequest{
		UserId: int32(driverID),
	}
	logger.Info("Calling trip service GetTripsByDriver via gRPC",
		"driver_id", strconv.Itoa(driverID),
	)

	resp, err := app.GRPCClients.TripClient.GetTripsByDriver(ctx, req)
	if err != nil {
		logger.Error("gRPC GetTripsByDriver failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) GetAllTripsViaGRPC(ctx context.Context, page int, limit int) (*trippb.PageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.GetAllTripsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	}
	logger.Info("Calling trip service GetAllTrips via gRPC")

	resp, err := app.GRPCClients.TripClient.GetAllTrips(ctx, req)
	if err != nil {
		logger.Error("gRPC GetAllTrips failed", "error", err)
		return nil, err
	}

	return resp, nil
}

func (app *Config) UpdateTripStatusViaGRPC(ctx context.Context, tripID int, driverID int, newStatus string) (*trippb.MessageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Convert string status to trippb.TripStatus enum; unknown strings map to zero value.
	statusEnum := trippb.TripStatus(trippb.TripStatus_value[newStatus])
	req := &trippb.UpdateTripStatusRequest{
		TripId:   int32(tripID),
		DriverId: int32(driverID),
		Status:   statusEnum,
	}
	logger.Info("Calling trip service UpdateTripStatus via gRPC",
		"trip_id", strconv.Itoa(tripID),
		"driver_id", strconv.Itoa(driverID),
		"new_status", newStatus,
	)
	resp, err := app.GRPCClients.TripClient.UpdateTripStatus(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateTripStatus failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (app *Config) CancelTripViaGRPC(ctx context.Context, tripID int, userID int) (*trippb.MessageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.CancelTripRequest{
		TripId: int32(tripID),
		UserId: int32(userID),
	}
	logger.Info("Calling trip service CancelTrip via gRPC",
		"trip_id", strconv.Itoa(tripID),
		"user_id", strconv.Itoa(userID),
	)

	resp, err := app.GRPCClients.TripClient.CancelTrip(ctx, req)
	if err != nil {
		logger.Error("gRPC CancelTrip failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (app *Config) SubmitReviewViaGRPC(ctx context.Context, tripID int, passengerID int, rating int, comment string) (*trippb.MessageResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.SubmitReviewRequest{
		TripId: int32(tripID),
		UserId: int32(passengerID),
		Review: &trippb.Review{
			Rating:  int32(rating),
			Comment: comment,
		},
	}
	logger.Info("Calling trip service SubmitReview via gRPC",
		"trip_id", strconv.Itoa(tripID),
		"passenger_id", strconv.Itoa(passengerID),
	)

	resp, err := app.GRPCClients.TripClient.SubmitReview(ctx, req)
	if err != nil {
		logger.Error("gRPC SubmitReview failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (app *Config) GetReviewViaGRPC(ctx context.Context, tripID int, userID int) (*trippb.GetTripReviewResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &trippb.TripIDRequest{
		PassengerId: int32(userID),
		TripId:      int32(tripID),
	}
	logger.Info("Calling trip service GetReview via gRPC",
		"trip_id", strconv.Itoa(tripID),
		"user_id", strconv.Itoa(userID),
	)

	resp, err := app.GRPCClients.TripClient.GetTripReview(ctx, req)
	if err != nil {
		logger.Error("gRPC GetReview failed", "error", err)
		return nil, err
	}
	return resp, nil
}

// I ain't touching all that
// ============================================
// User Service gRPC Client Methods
// ============================================

func (app *Config) GetAllUsersViaGRPC(ctx context.Context) (*userpb.GetAllUsersResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.GetAllUsersRequest{}
logger.Info("Calling user service GetAllUsers via gRPC")

resp, err := app.GRPCClients.UserClient.GetAllUsers(ctx, req)
if err != nil {
logger.Error("gRPC GetAllUsers failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) GetUserByIdViaGRPC(ctx context.Context, userID int) (*userpb.GetUserByIdResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.GetUserByIdRequest{
UserId: int32(userID),
}
logger.Info("Calling user service GetUserById via gRPC", "user_id", userID)

resp, err := app.GRPCClients.UserClient.GetUserById(ctx, req)
if err != nil {
logger.Error("gRPC GetUserById failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) CreateUserViaGRPC(ctx context.Context, email string) (*userpb.CreateUserResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.CreateUserRequest{
Email: email,
}
logger.Info("Calling user service CreateUser via gRPC", "email", email)

resp, err := app.GRPCClients.UserClient.CreateUser(ctx, req)
if err != nil {
logger.Error("gRPC CreateUser failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) UpdateUserViaGRPC(ctx context.Context, userID int, email, role, driverStatus string) (*userpb.UpdateUserResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.UpdateUserRequest{
UserId:       int32(userID),
Email:        email,
Role:         role,
DriverStatus: driverStatus,
}
logger.Info("Calling user service UpdateUser via gRPC", "user_id", userID)

resp, err := app.GRPCClients.UserClient.UpdateUser(ctx, req)
if err != nil {
logger.Error("gRPC UpdateUser failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) DeleteUserViaGRPC(ctx context.Context, userID int) (*userpb.DeleteUserResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.DeleteUserRequest{
UserId: int32(userID),
}
logger.Info("Calling user service DeleteUser via gRPC", "user_id", userID)

resp, err := app.GRPCClients.UserClient.DeleteUser(ctx, req)
if err != nil {
logger.Error("gRPC DeleteUser failed", "error", err)
return nil, err
}
return resp, nil
}

// ============================================
// Vehicle Service gRPC Client Methods
// ============================================

func (app *Config) GetAllVehiclesViaGRPC(ctx context.Context) (*userpb.GetAllVehiclesResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.GetAllVehiclesRequest{}
logger.Info("Calling user service GetAllVehicles via gRPC")

resp, err := app.GRPCClients.UserClient.GetAllVehicles(ctx, req)
if err != nil {
logger.Error("gRPC GetAllVehicles failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) GetVehicleByIdViaGRPC(ctx context.Context, vehicleID int) (*userpb.GetVehicleByIdResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.GetVehicleByIdRequest{
VehicleId: int32(vehicleID),
}
logger.Info("Calling user service GetVehicleById via gRPC", "vehicle_id", vehicleID)

resp, err := app.GRPCClients.UserClient.GetVehicleById(ctx, req)
if err != nil {
logger.Error("gRPC GetVehicleById failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) GetVehiclesByUserIdViaGRPC(ctx context.Context, userID int) (*userpb.GetVehiclesByUserIdResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.GetVehiclesByUserIdRequest{
UserId: int32(userID),
}
logger.Info("Calling user service GetVehiclesByUserId via gRPC", "user_id", userID)

resp, err := app.GRPCClients.UserClient.GetVehiclesByUserId(ctx, req)
if err != nil {
logger.Error("gRPC GetVehiclesByUserId failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) CreateVehicleViaGRPC(ctx context.Context, driverID int, licensePlate, vehicleType string, seats int, status string) (*userpb.CreateVehicleResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.CreateVehicleRequest{
DriverId:     int32(driverID),
LicensePlate: licensePlate,
VehicleType:  vehicleType,
Seats:        int32(seats),
Status:       status,
}
logger.Info("Calling user service CreateVehicle via gRPC", "driver_id", driverID)

resp, err := app.GRPCClients.UserClient.CreateVehicle(ctx, req)
if err != nil {
logger.Error("gRPC CreateVehicle failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) UpdateVehicleViaGRPC(ctx context.Context, vehicleID int, licensePlate, vehicleType string, seats int, status string) (*userpb.UpdateVehicleResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.UpdateVehicleRequest{
VehicleId:    int32(vehicleID),
LicensePlate: licensePlate,
VehicleType:  vehicleType,
Seats:        int32(seats),
Status:       status,
}
logger.Info("Calling user service UpdateVehicle via gRPC", "vehicle_id", vehicleID)

resp, err := app.GRPCClients.UserClient.UpdateVehicle(ctx, req)
if err != nil {
logger.Error("gRPC UpdateVehicle failed", "error", err)
return nil, err
}
return resp, nil
}

func (app *Config) DeleteVehicleViaGRPC(ctx context.Context, vehicleID int) (*userpb.DeleteVehicleResponse, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req := &userpb.DeleteVehicleRequest{
VehicleId: int32(vehicleID),
}
logger.Info("Calling user service DeleteVehicle via gRPC", "vehicle_id", vehicleID)

resp, err := app.GRPCClients.UserClient.DeleteVehicle(ctx, req)
if err != nil {
logger.Error("gRPC DeleteVehicle failed", "error", err)
return nil, err
}
return resp, nil
}
