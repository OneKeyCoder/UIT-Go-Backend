package location_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/env"
	"github.com/redis/go-redis/v9"
)

var TIMETOLIVE = env.Get("REDIS_TIME_TO_LIVE", "3600")

const (
	GeoKeyDrivers    = "geo:drivers"
	GeoKeyPassengers = "geo:passengers"
)

type LocationService struct {
	redisClient *redis.Client
}

func NewLocationService(redisClient *redis.Client) *LocationService {
	return &LocationService{
		redisClient: redisClient,
	}
}

// getGeoKey returns the appropriate geo key based on user role
func (s *LocationService) getGeoKey(role string) string {
	if role == "driver" {
		return GeoKeyDrivers
	}
	return GeoKeyPassengers
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

	// Use separate geo keys for drivers and passengers
	geoKey := s.getGeoKey(location.Role)
	if err := s.redisClient.GeoAdd(ctx, geoKey, &redis.GeoLocation{
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
	// Get the current user's position from passengers geo key
	userPos, err := s.redisClient.GeoPos(ctx, GeoKeyPassengers, strconv.Itoa(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user position: %w", err)
	}

	if len(userPos) == 0 || userPos[0] == nil {
		return nil, fmt.Errorf("user position not found")
	}

	// Query only drivers geo key for optimal performance
	// This eliminates the need for application-layer filtering
	results, err := s.redisClient.GeoRadius(ctx, GeoKeyDrivers,
		userPos[0].Longitude,
		userPos[0].Latitude,
		&redis.GeoRadiusQuery{
			Radius:    radius,
			Unit:      "km",
			WithCoord: true,
			WithDist:  true,
			Count:     topN,
			Sort:      "ASC",
		}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}

	// Convert GeoLocation results to CurrentLocation
	locations := make([]*CurrentLocation, 0, len(results))
	for _, geoLoc := range results {
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
	}

	return locations, nil
}

// GetAllLocations retrieves all stored locations from Redis
// Uses SCAN instead of KEYS to avoid blocking Redis in production
func (s *LocationService) GetAllLocations(ctx context.Context) ([]*CurrentLocation, error) {
	var cursor uint64
	locations := make([]*CurrentLocation, 0)

	// Use SCAN instead of KEYS * to avoid blocking Redis server
	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys from Redis: %w", err)
		}

		// Process the keys in this batch
		for _, key := range keys {
			// Skip geo index keys
			if key == GeoKeyDrivers || key == GeoKeyPassengers {
				continue
			}

			tempGeo, err := strconv.Atoi(key)
			if err != nil {
				// Skip non-numeric keys
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

		// Update cursor for next iteration
		cursor = nextCursor
		if cursor == 0 {
			// Scan complete
			break
		}
	}

	return locations, nil
}

// RemoveLocation removes a user's location from Redis and geo index
func (s *LocationService) RemoveLocation(ctx context.Context, userID int, role string) error {
	userIDStr := strconv.Itoa(userID)

	// Remove from Redis key-value store
	if err := s.redisClient.Del(ctx, userIDStr).Err(); err != nil {
		return fmt.Errorf("failed to delete location from Redis: %w", err)
	}

	// Remove from geo index
	geoKey := s.getGeoKey(role)
	if err := s.redisClient.ZRem(ctx, geoKey, userIDStr).Err(); err != nil {
		return fmt.Errorf("failed to remove location from geo index: %w", err)
	}

	return nil
}

// UpdateUserRole updates a user's role and moves them between geo indexes
func (s *LocationService) UpdateUserRole(ctx context.Context, userID int, oldRole, newRole string) error {
	// Get current location
	location, err := s.GetCurrentLocation(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current location: %w", err)
	}
	if location == nil {
		return fmt.Errorf("location not found for user %d", userID)
	}

	userIDStr := strconv.Itoa(userID)

	// Remove from old geo index
	oldGeoKey := s.getGeoKey(oldRole)
	if err := s.redisClient.ZRem(ctx, oldGeoKey, userIDStr).Err(); err != nil {
		return fmt.Errorf("failed to remove from old geo index: %w", err)
	}

	// Add to new geo index
	newGeoKey := s.getGeoKey(newRole)
	if err := s.redisClient.GeoAdd(ctx, newGeoKey, &redis.GeoLocation{
		Name:      userIDStr,
		Longitude: location.Longitude,
		Latitude:  location.Latitude,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to new geo index: %w", err)
	}

	// Update the role in the location data
	location.Role = newRole
	data, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	ttl, err := strconv.Atoi(TIMETOLIVE)
	if err != nil {
		ttl = 3600
	}

	if err := s.redisClient.Set(ctx, userIDStr, data, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to update location in Redis: %w", err)
	}

	return nil
}
