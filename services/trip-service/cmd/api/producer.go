package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
)

type EventMessage struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func PublishEvent(conn *amqp.Conn, eventName, eventData string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new session for this publish operation
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		logger.Error("Failed to create AMQP 1.0 session", "error", err)
		return err
	}
	defer session.Close(ctx)

	// Create a sender to the logs queue
	sender, err := session.NewSender(ctx, "/queues/logs", nil)
	if err != nil {
		logger.Error("Failed to create RabbitMQ sender", "error", err)
		return err
	}
	defer sender.Close(ctx)

	// Create event message
	event := EventMessage{
		Name: eventName,
		Data: eventData,
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

	logger.Info("Published event to RabbitMQ",
		"name", eventName,
		"data", eventData)

	return nil
}

// Helper function to create string pointer
func to(s string) *string {
	return &s
}
