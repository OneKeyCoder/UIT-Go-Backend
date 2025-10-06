package main

import (
	"context"
	"log"

	"common_pkg/helpers"
	location_service "location-service/internal"
	"location-service/internal/configs"
	"location-service/internal/handlers"
	"location-service/internal/routes"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Handlers *handlers.Handlers
}

func main() {
	if err := helpers.LoadEnv(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	log.Println("Connecting to Redis...")
	redisClient, err := configs.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize service and handlers
	ctx := context.Background()
	locationService := location_service.NewLocationService(redisClient)
	locationHandlers := handlers.NewHandlers(&ctx, locationService)

	port := helpers.GetEnv("PORT", "8080")

	router := gin.Default()

	routes.InitRoute(router, locationHandlers)

	log.Printf("Starting Location Service on port %s", port)
	router.Run(":" + port)
}
