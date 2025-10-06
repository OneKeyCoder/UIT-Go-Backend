package helpers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadEnv() error {
	if err := godotenv.Load("../common/env/location-service.env"); err != nil {
		if err := godotenv.Load("../../common/env/location-service.env"); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
			return err
		}
	}
	return nil
}
