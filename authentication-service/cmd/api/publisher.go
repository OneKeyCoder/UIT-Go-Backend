package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type EventMessage struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func PublishEvent(conn *amqp.Connection, eventName, eventData string) error {
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open RabbitMQ channel", zap.Error(err))
		return err
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		logger.Error("Failed to declare exchange", zap.Error(err))
		return err
	}

	// Create event message
	event := EventMessage{
		Name: eventName,
		Data: eventData,
	}

	body, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	// Publish message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		"logs_topic", // exchange
		"log.INFO",   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		logger.Error("Failed to publish event", zap.Error(err))
		return err
	}

	logger.Info("Published event to RabbitMQ",
		zap.String("name", eventName),
		zap.String("data", eventData))

	return nil
}
