package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
)

type EventMessage struct {
	EventName string `json:"event_name"` // What action was performed
	ActorID   string `json:"actor_id"`   // Who performed it
	Metadata  string `json:"metadata"`   // JSON with context (IP, User-Agent, etc.)
	Status    string `json:"status"`     // success, failure, error
}

// AuditMetadata contains contextual information for audit logs
// Implements the 4W principle for comprehensive audit trails:
// - WHO: ActorID (in EventMessage) + Email
// - WHEN: Timestamp (added by logger-service via MongoDB)
// - WHAT: EventName + Action + Status
// - WHERE: IP + UserAgent (browser/device info)
type AuditMetadata struct {
	IP        string                 `json:"ip"`         // WHERE: Client IP (from X-Forwarded-For, X-Real-IP, or RemoteAddr)
	UserAgent string                 `json:"user_agent"` // WHERE: Client User-Agent (Mobile/Desktop/Browser)
	Email     string                 `json:"email,omitempty"`
	Action    string                 `json:"action,omitempty"` // WHAT: Human-readable action description
	Reason    string                 `json:"reason,omitempty"` // WHY: Failure reason for security monitoring (brute-force detection)
	Extra     map[string]interface{} `json:"extra,omitempty"`  // Additional structured context data
}

// PublishAuditEvent publishes a structured audit event to RabbitMQ
// Uses a reusable session if provided, otherwise creates a new one
func PublishAuditEvent(conn *amqp.Conn, eventName, actorID, status string, metadata AuditMetadata) error {
	return PublishAuditEventWithSession(nil, conn, eventName, actorID, status, metadata)
}

// PublishAuditEventWithSession publishes using a specific session (for connection pooling)
func PublishAuditEventWithSession(session *amqp.Session, conn *amqp.Conn, eventName, actorID, status string, metadata AuditMetadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use provided session or create new one
	if session == nil {
		var err error
		session, err = conn.NewSession(ctx, nil)
		if err != nil {
			logger.Error("Failed to open RabbitMQ channel", "error", err)
			return err
		}
		defer session.Close(ctx)
	}

	// Create a sender to the logs queue
	sender, err := session.NewSender(ctx, "/queues/logs", nil)
	if err != nil {
		logger.Error("Failed to declare exchange", "error", err)
		return err
	}
	defer sender.Close(ctx)

	// Marshal metadata to JSON string
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		logger.Error("Failed to marshal metadata", "error", err)
		return err
	}

	// Create structured event message
	event := EventMessage{
		EventName: eventName,
		ActorID:   actorID,
		Metadata:  string(metadataJSON),
		Status:    status,
	}

	body, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", "error", err)
		return err
	}

	// Create message
	msg := &amqp.Message{
		Data: [][]byte{body},
		Properties: &amqp.MessageProperties{
			ContentType: to("application/json"),
		},
	}

	// Send the message
	err = sender.Send(ctx, msg, nil)
	if err != nil {
		logger.Error("Failed to publish event", "error", err)
		return err
	}

	logger.Info("Published audit event to RabbitMQ",
		"event_name", eventName,
		"actor_id", actorID,
		"status", status)

	return nil
}

// Legacy function for backward compatibility
func PublishEvent(conn *amqp.Conn, eventName, eventData string) error {
	// Convert to new format with minimal context
	metadata := AuditMetadata{
		Action: eventData,
	}
	return PublishAuditEvent(conn, eventName, "system", "success", metadata)
}

// Helper function to create string pointer
func to(s string) *string {
	return &s
}
