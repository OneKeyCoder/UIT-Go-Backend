package location_service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var TIMETOLIVE = getEnv("REDIS_TIME_TO_LIVE", "3600")

type LocationService struct {
	redisClient *redis.Client
}

func NewLocationService(redisClient *redis.Client) *LocationService {
	return &LocationService{
		redisClient: redisClient,
	}
}

func (s *LocationService) SetCurrentLocation(ctx context.Context, location *CurrentLocation) error {
	data, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	ttl, err := strconv.Atoi(TIMETOLIVE)
	if err != nil {
		ttl = 3600
	}

	if err := s.redisClient.Set(ctx, strconv.Itoa(location.UserID), data, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set location in Redis: %w", err)
	}

	if err := s.redisClient.GeoAdd(ctx, "geo:users", &redis.GeoLocation{
		Name:      strconv.Itoa(location.UserID),
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add location to geo index: %w", err)
	}

	return nil
}

func (s *LocationService) GetCurrentLocation(ctx context.Context, userID int) (*CurrentLocation, error) {
	data, err := s.redisClient.Get(ctx, strconv.Itoa(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No location found
		}
		return nil, fmt.Errorf("failed to get location from Redis: %w", err)
	}
	var location CurrentLocation
	if err := json.Unmarshal([]byte(data), &location); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location data: %w", err)
	}
	return &location, nil
}

func (s *LocationService) FindTopNearestUsers(ctx context.Context, userID int, topN int, radius float64) ([]*CurrentLocation, error) {
	// Get the current user's position
	userPos, err := s.redisClient.GeoPos(ctx, "geo:users", strconv.Itoa(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user position: %w", err)
	}

	if len(userPos) == 0 || userPos[0] == nil {
		return nil, fmt.Errorf("user position not found")
	}

	// Find nearby users within the specified radius using GeoRadius
	results, err := s.redisClient.GeoRadius(ctx, "geo:users",
		userPos[0].Longitude,
		userPos[0].Latitude,
		&redis.GeoRadiusQuery{
			Radius:    radius,
			Unit:      "km",
			WithCoord: true,
			WithDist:  true,
			Count:     topN + 1, // +1 to account for the user themselves
			Sort:      "ASC",
		}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to search nearby users: %w", err)
	}

	// Convert GeoLocation results to CurrentLocation
	locations := make([]*CurrentLocation, 0, topN)
	for _, geoLoc := range results {
		// Skip the user themselves
		if geoLoc.Name == strconv.Itoa(userID) {
			continue
		}
		tempGeoLoc, err := strconv.Atoi(geoLoc.Name)
		if err != nil {
			continue
		}
		// Get full location details from Redis
		location, err := s.GetCurrentLocation(ctx, tempGeoLoc)
		if err != nil {
			// Skip if we can't get the location details
			continue
		}

		if location != nil {
			// Add distance to the location
			location.Distance = geoLoc.Dist
			locations = append(locations, location)
		}

		// Stop if we have enough results
		if len(locations) >= topN {
			break
		}
	}

	return locations, nil
}

// GetAllLocations retrieves all stored locations from Redis
func (s *LocationService) GetAllLocations(ctx context.Context) ([]*CurrentLocation, error) {
	keys, err := s.redisClient.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from Redis: %w", err)
	}

	// Filter out the geo index key
	locations := make([]*CurrentLocation, 0, len(keys))
	for _, key := range keys {
		// Skip geo index key
		if key == "geo:users" {
			continue
		}
		tempGeo, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		location, err := s.GetCurrentLocation(ctx, tempGeo)
		if err != nil {
			continue
		}
		if location != nil {
			locations = append(locations, location)
		}
	}
	return locations, nil
}
