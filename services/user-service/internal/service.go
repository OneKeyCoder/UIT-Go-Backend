package user_service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbTimeout          = time.Second * 3
	usersCollection    = "users"
	vehiclesCollection = "vehicles"
)

type UserService struct {
	mongoClient *mongo.Client
}

func NewUserService(mongoClient *mongo.Client) *UserService {
	return &UserService{
		mongoClient: mongoClient,
	}
}

func (us *UserService) getUsersCollection() *mongo.Collection {
	return us.mongoClient.Database("mongo").Collection(usersCollection)
}

func (us *UserService) getVehiclesCollection() *mongo.Collection {
	return us.mongoClient.Database("mongo").Collection(vehiclesCollection)
}

func (us *UserService) GetUserById(ctx context.Context, userId int) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var user User
	collection := us.getUsersCollection()

	err := collection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, fmt.Errorf("user not found with id: %d", userId)
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (us *UserService) GetAllUsers(ctx context.Context) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getUsersCollection()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (us *UserService) CreateUser(ctx context.Context, userRequest UserRequest) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getUsersCollection()

	// Check if user with email already exists
	count, err := collection.CountDocuments(ctx, bson.M{"email": userRequest.Email})
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("user with email %s already exists", userRequest.Email)
	}

	// Get the next user_id
	opts := options.FindOne().SetSort(bson.D{{Key: "user_id", Value: -1}})
	var lastUser User
	err = collection.FindOne(ctx, bson.M{}, opts).Decode(&lastUser)
	nextUserId := 1
	if err == nil {
		nextUserId = lastUser.UserId + 1
	} else if err != mongo.ErrNoDocuments {
		return fmt.Errorf("failed to get last user: %w", err)
	}

	user := User{
		UserId:           nextUserId,
		Email:            userRequest.Email,
		FirstName:        userRequest.FirstName,
		LastName:         userRequest.LastName,
		Role:             userRequest.Role,
		DriverStatus:     userRequest.DriverStatus,
		DriverTotalTrip:  0,
		DriverRevenue:    0,
		DriverAvgRating:  0,
		DriverVerifiedAt: time.Time{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Set defaults if not provided
	if user.Role == "" {
		user.Role = "user"
	}
	if user.DriverStatus == "" {
		user.DriverStatus = "inactive"
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (us *UserService) DeleteUserById(ctx context.Context, userId int) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getUsersCollection()

	result, err := collection.DeleteOne(ctx, bson.M{"user_id": userId})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found with id: %d", userId)
	}

	return nil
}

func (us *UserService) UpdateUserById(ctx context.Context, userId int, userRequest UserRequest) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getUsersCollection()

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Only update fields that are provided
	if userRequest.Email != "" {
		update["$set"].(bson.M)["email"] = userRequest.Email
	}
	if userRequest.FirstName != "" {
		update["$set"].(bson.M)["first_name"] = userRequest.FirstName
	}
	if userRequest.LastName != "" {
		update["$set"].(bson.M)["last_name"] = userRequest.LastName
	}
	if userRequest.Role != "" {
		update["$set"].(bson.M)["role"] = userRequest.Role
		if userRequest.Role == "driver" {
			update["$set"].(bson.M)["driver_status"] = "pending"
		}
	}
	if userRequest.DriverStatus != "" {
		update["$set"].(bson.M)["driver_status"] = userRequest.DriverStatus
		if userRequest.DriverStatus == "verified" {
			update["$set"].(bson.M)["driver_verified_at"] = time.Now()
		}
	}

	result, err := collection.UpdateOne(ctx, bson.M{"user_id": userId}, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found with id: %d", userId)
	}

	return nil
}

func (us *UserService) GetVehiclesByUserId(ctx context.Context, userId int) ([]Vehicle, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getVehiclesCollection()
	cursor, err := collection.Find(ctx, bson.M{"driver_id": userId})
	if err != nil {
		return nil, fmt.Errorf("failed to find vehicles: %w", err)
	}
	defer cursor.Close(ctx)

	var vehicles []Vehicle
	if err := cursor.All(ctx, &vehicles); err != nil {
		return nil, fmt.Errorf("failed to decode vehicles: %w", err)
	}

	return vehicles, nil
}

func (us *UserService) GetVehicleById(ctx context.Context, vehicleId int) (Vehicle, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var vehicle Vehicle
	collection := us.getVehiclesCollection()

	err := collection.FindOne(ctx, bson.M{"vehicle_id": vehicleId}).Decode(&vehicle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Vehicle{}, fmt.Errorf("vehicle not found with id: %d", vehicleId)
		}
		return Vehicle{}, fmt.Errorf("failed to get vehicle: %w", err)
	}

	return vehicle, nil
}

func (us *UserService) GetAllVehicles(ctx context.Context) ([]Vehicle, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getVehiclesCollection()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find vehicles: %w", err)
	}
	defer cursor.Close(ctx)

	var vehicles []Vehicle
	if err := cursor.All(ctx, &vehicles); err != nil {
		return nil, fmt.Errorf("failed to decode vehicles: %w", err)
	}

	return vehicles, nil
}

func (us *UserService) CreateVehicle(ctx context.Context, driverId int, vehicleRequest VehicleRequest) (Vehicle, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getVehiclesCollection()

	// Get the next vehicle_id
	opts := options.FindOne().SetSort(bson.D{{Key: "vehicle_id", Value: -1}})
	var lastVehicle Vehicle
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&lastVehicle)
	nextVehicleId := 1
	if err == nil {
		nextVehicleId = lastVehicle.VehicleId + 1
	} else if err != mongo.ErrNoDocuments {
		return Vehicle{}, fmt.Errorf("failed to get last vehicle: %w", err)
	}

	vehicle := Vehicle{
		VehicleId:    nextVehicleId,
		DriverId:     driverId,
		LicensePlate: vehicleRequest.LicensePlate,
		VehicleType:  vehicleRequest.VehicleType,
		Seats:        vehicleRequest.Seats,
		Status:       vehicleRequest.Status,
		VerifiedAt:   time.Time{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = collection.InsertOne(ctx, vehicle)
	if err != nil {
		return Vehicle{}, fmt.Errorf("failed to create vehicle: %w", err)
	}

	return vehicle, nil
}

func (us *UserService) UpdateVehicle(ctx context.Context, vehicleId int, vehicleRequest VehicleRequest) (Vehicle, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getVehiclesCollection()

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Only update fields that are provided
	if vehicleRequest.LicensePlate != "" {
		update["$set"].(bson.M)["license_plate"] = vehicleRequest.LicensePlate
	}
	if vehicleRequest.VehicleType != "" {
		update["$set"].(bson.M)["vehicle_type"] = vehicleRequest.VehicleType
	}
	if vehicleRequest.Seats > 0 {
		update["$set"].(bson.M)["seats"] = vehicleRequest.Seats
	}
	if vehicleRequest.Status != "" {
		update["$set"].(bson.M)["status"] = vehicleRequest.Status
	}

	result, err := collection.UpdateOne(ctx, bson.M{"vehicle_id": vehicleId}, update)
	if err != nil {
		return Vehicle{}, fmt.Errorf("failed to update vehicle: %w", err)
	}

	if result.MatchedCount == 0 {
		return Vehicle{}, fmt.Errorf("vehicle not found with id: %d", vehicleId)
	}

	// Fetch and return the updated vehicle
	var vehicle Vehicle
	err = collection.FindOne(ctx, bson.M{"vehicle_id": vehicleId}).Decode(&vehicle)
	if err != nil {
		return Vehicle{}, fmt.Errorf("failed to fetch updated vehicle: %w", err)
	}

	return vehicle, nil
}

func (us *UserService) DeleteVehicleById(ctx context.Context, vehicleId int) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	collection := us.getVehiclesCollection()

	result, err := collection.DeleteOne(ctx, bson.M{"vehicle_id": vehicleId})
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("vehicle not found with id: %d", vehicleId)
	}

	return nil
}
