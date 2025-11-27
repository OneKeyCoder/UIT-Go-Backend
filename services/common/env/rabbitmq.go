package env

import (
	"fmt"
	"strings"
)

// RabbitMQURL returns the AMQP connection string built from split env vars.
// If RABBITMQ_URL is set, it takes precedence for backward compatibility.
func RabbitMQURL() string {
	if url := Get("RABBITMQ_URL", ""); url != "" {
		return url
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
