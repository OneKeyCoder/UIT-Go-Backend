package location_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"common_pkg/helpers"

	"github.com/redis/go-redis/v9"
)

var TIMETOLIVE = helpers.GetEnv("REDIS_TIME_TO_LIVE", "3600")

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

	if err := s.redisClient.Set(ctx, location.UserID, data, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set location in Redis: %w", err)
	}

	if err := s.redisClient.GeoAdd(ctx, "geo:users", &redis.GeoLocation{
		Name:      location.UserID,
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add location to geo index: %w", err)
	}

	return nil
}

func (s *LocationService) GetCurrentLocation(ctx context.Context, userID string) (*CurrentLocation, error) {
	data, err := s.redisClient.Get(ctx, userID).Result()
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

func (s *LocationService) FindTopNearestUsers(ctx context.Context, userID string, topN int, radius float64) ([]*CurrentLocation, error) {
	// Get the current user's position
	userPos, err := s.redisClient.GeoPos(ctx, "geo:users", userID).Result()
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
		if geoLoc.Name == userID {
			continue
		}

		// Get full location details from Redis
		location, err := s.GetCurrentLocation(ctx, geoLoc.Name)
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
