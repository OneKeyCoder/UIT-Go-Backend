package env

import (
	"fmt"
	"strings"
)

// RabbitMQURL returns a fully-qualified AMQP(S) connection string.
// Priority order:
//  1. RABBITMQ_CONNECTION_STRING (new preferred variable)
//  2. RABBITMQ_URL (legacy single string)
//  3. Individual host/user/password env vars
func RabbitMQURL() string {
	for _, key := range []string{"RABBITMQ_CONNECTION_STRING", "RABBITMQ_URL"} {
		if url := strings.TrimSpace(Get(key, "")); url != "" {
			return url
		}
	}

	host := Get("RABBITMQ_HOST", "rabbitmq")
	port := Get("RABBITMQ_PORT", "5672")
	user := Get("RABBITMQ_USER", "guest")
	password := Get("RABBITMQ_PASSWORD", "guest")
	vhost := Get("RABBITMQ_VHOST", "/")

	// Normalize vhost so that "/" stays root and custom names do not double-prefix.
	trimmed := strings.TrimPrefix(vhost, "/")
	if trimmed == "" {
		return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, trimmed)
}
