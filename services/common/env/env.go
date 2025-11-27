package env

import "os"

// Get returns the environment variable for key or defaultValue when unset or empty.
func Get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
