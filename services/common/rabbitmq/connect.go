package rabbitmq

import (
	"context"
	"math"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"go.uber.org/zap"
)

// ConnectOptions configures the connection behavior
type ConnectOptions struct {
	// MaxRetries is the maximum number of connection attempts (default: 5)
	MaxRetries int
	// EnsureQueues if true, will create default queues after connecting
	EnsureQueues bool
	// ManagementURL is required if EnsureQueues is true
	ManagementURL string
	// Username for RabbitMQ authentication (default: guest)
	Username string
	// Password for RabbitMQ authentication (default: guest)
	Password string
}

// DefaultConnectOptions returns sensible defaults
func DefaultConnectOptions() ConnectOptions {
	return ConnectOptions{
		MaxRetries:    5,
		EnsureQueues:  true,
		ManagementURL: "http://rabbitmq:15672",
		Username:      "guest",
		Password:      "guest",
	}
}

// Connect establishes connection to RabbitMQ with retry logic and optional queue creation
// This is the recommended way to connect to RabbitMQ in this project
func Connect(ctx context.Context, amqpURL string, opts *ConnectOptions) (*amqp.Conn, error) {
	if opts == nil {
		defaultOpts := DefaultConnectOptions()
		opts = &defaultOpts
	}

	var connection *amqp.Conn
	var lastErr error
	backOff := 1 * time.Second

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		c, err := amqp.Dial(dialCtx, amqpURL, &amqp.ConnOptions{
			IdleTimeout: 30 * time.Second,
		})
		cancel()

		if err != nil {
			logger.Info("RabbitMQ not yet ready...",
				zap.Int("attempt", attempt),
				zap.Error(err))
			lastErr = err

			if attempt >= opts.MaxRetries {
				break
			}

			backOff = time.Duration(math.Pow(float64(attempt), 2)) * time.Second
			logger.Info("Backing off...", zap.Duration("duration", backOff))
			time.Sleep(backOff)
			continue
		}

		logger.Info("Connected to RabbitMQ")
		connection = c
		break
	}

	if connection == nil {
		return nil, lastErr
	}

	// Ensure queues exist if requested
	if opts.EnsureQueues && opts.ManagementURL != "" {
		if err := ensureDefaultQueues(ctx, opts); err != nil {
			logger.Warn("Failed to ensure queues exist, they may need to be created manually",
				zap.Error(err))
			// Don't fail - queues might already exist or be created by another service
		}
	}

	return connection, nil
}

// ensureDefaultQueues creates all default queues using Management HTTP API
func ensureDefaultQueues(ctx context.Context, opts *ConnectOptions) error {
	client := &Client{
		config: ConnectionConfig{
			ManagementURL: opts.ManagementURL,
			Username:      opts.Username,
			Password:      opts.Password,
			VHost:         "/",
		},
	}

	queues := DefaultQueues()
	for _, q := range queues {
		if err := client.EnsureQueue(ctx, q); err != nil {
			return err
		}
		logger.Info("Ensured queue exists", zap.String("queue", q.Name))
	}

	return nil
}

// ConnectSimple is a convenience function that connects with default options
// and ensures all default queues exist
func ConnectSimple(amqpURL string) (*amqp.Conn, error) {
	return Connect(context.Background(), amqpURL, nil)
}
