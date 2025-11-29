package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-amqp"
)

// QueueConfig represents queue configuration
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Arguments  map[string]interface{}
}

// ConnectionConfig holds RabbitMQ connection details
type ConnectionConfig struct {
	// AMQP URL for go-amqp connection (amqp://guest:guest@localhost:5672/)
	AMQPURL string
	// Management API URL (http://localhost:15672)
	ManagementURL string
	// Username for authentication
	Username string
	// Password for authentication
	Password string
	// Virtual host (default "/")
	VHost string
}

// Client wraps AMQP connection with queue management capabilities
type Client struct {
	Conn   *amqp.Conn
	config ConnectionConfig
}

// NewClient creates a new RabbitMQ client with queue management
func NewClient(ctx context.Context, config ConnectionConfig) (*Client, error) {
	if config.VHost == "" {
		config.VHost = "/"
	}

	// Connect using go-amqp (AMQP 1.0)
	conn, err := amqp.Dial(ctx, config.AMQPURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &Client{
		Conn:   conn,
		config: config,
	}, nil
}

// EnsureQueue creates queue if it doesn't exist using Management HTTP API
// This is necessary because AMQP 1.0 doesn't support queue declaration
func (c *Client) EnsureQueue(ctx context.Context, queue QueueConfig) error {
	// URL encode the vhost (/ becomes %2F)
	encodedVHost := url.PathEscape(c.config.VHost)

	apiURL := fmt.Sprintf("%s/api/queues/%s/%s",
		c.config.ManagementURL,
		encodedVHost,
		url.PathEscape(queue.Name),
	)

	// Prepare request body
	body := map[string]interface{}{
		"durable":     queue.Durable,
		"auto_delete": queue.AutoDelete,
	}
	if queue.Arguments != nil {
		body["arguments"] = queue.Arguments
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal queue config: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call management API: %w", err)
	}
	defer resp.Body.Close()

	// 201 = created, 204 = already exists (no change)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to create queue: HTTP %d", resp.StatusCode)
	}

	return nil
}

// EnsureQueues creates multiple queues
func (c *Client) EnsureQueues(ctx context.Context, queues ...QueueConfig) error {
	for _, q := range queues {
		if err := c.EnsureQueue(ctx, q); err != nil {
			return fmt.Errorf("failed to ensure queue %s: %w", q.Name, err)
		}
	}
	return nil
}

// Close closes the AMQP connection
func (c *Client) Close(ctx context.Context) error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

// GetConnection returns the underlying AMQP connection
func (c *Client) GetConnection() *amqp.Conn {
	return c.Conn
}
