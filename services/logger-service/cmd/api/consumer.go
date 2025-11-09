package main

import (
	"encoding/json"
	"logger-service/data"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// LogMessage represents the structure of messages received from RabbitMQ
type LogMessage struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// ConsumeFromRabbitMQ sets up a RabbitMQ consumer that listens for log messages
func (app *Config) ConsumeFromRabbitMQ(conn *amqp.Connection) error {
	logger.Info("Setting up RabbitMQ consumer")

	ch, err := conn.Channel()
	if err != nil {
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
		return err
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		"",    // name (empty = generate random name)
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Bind queue to exchange with routing key
	err = ch.QueueBind(
		q.Name,       // queue name
		"log.INFO",   // routing key
		"logs_topic", // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	logger.Info("RabbitMQ consumer ready", zap.String("queue", q.Name))

	// Process messages in a loop
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var logMsg LogMessage
			err := json.Unmarshal(d.Body, &logMsg)
			if err != nil {
				logger.Error("Failed to unmarshal log message", zap.Error(err))
				continue
			}

			logger.Info("Received log message from RabbitMQ",
				zap.String("name", logMsg.Name),
				zap.String("data", logMsg.Data))

			// Write to MongoDB
			logEntry := data.LogEntry{
				Name:      logMsg.Name,
				Data:      logMsg.Data,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = app.Models.LogEntry.Insert(logEntry)

			if err != nil {
				logger.Error("Failed to insert log entry", zap.Error(err))
			} else {
				logger.Info("Successfully wrote log to MongoDB", zap.String("name", logMsg.Name))
			}
		}
	}()

	<-forever
	return nil
}
