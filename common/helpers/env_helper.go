package helpers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// GetEnv returns the environment variable value or default if not set
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LoadEnv loads .env file from common/env
func LoadEnv() error {
	if err := godotenv.Load("../common/env/location-service.env"); err != nil {
		if err := godotenv.Load("../../common/env/location-service.env"); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
			return err
		}
	}
	return nil
}
