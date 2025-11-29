package main

import (
	"context"
	"encoding/json"
	"logger-service/data"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"go.uber.org/zap"
)

// LogMessage represents the structure of messages received from RabbitMQ
type LogMessage struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type AMQPConsumer struct {
	conn    *amqp.Conn
	session *amqp.Session
}

func NewAMQPConsumer(conn *amqp.Conn) (*AMQPConsumer, error) {
	session, err := conn.NewSession(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &AMQPConsumer{conn: conn, session: session}, nil
}

// Close closes the AMQP consumer
func (c *AMQPConsumer) Close(ctx context.Context) error {
	if c.session != nil {
		return c.session.Close(ctx)
	}
	return nil
}

// ConsumeFromRabbitMQ sets up a RabbitMQ consumer that listens for log messages
func (app *Config) ConsumeFromRabbitMQ(conn *amqp.Conn) error {
	logger.Info("Setting up RabbitMQ consumer")

	consumer, err := NewAMQPConsumer(conn)
	if err != nil {
		return err
	}

	// Create a receiver on the logs queue
	receiver, err := consumer.session.NewReceiver(context.Background(), "/queues/logs", &amqp.ReceiverOptions{
		Credit: 10, // Prefetch count - number of messages to buffer
	})
	if err != nil {
		logger.Error("Failed to create RabbitMQ receiver", zap.Error(err))
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		receiver.Close(ctx)
	}()

	logger.Info("RabbitMQ consumer ready", zap.String("address", "/queues/logs"))

	// Process messages in a loop
	for {
		// Receive next message (blocks until message is available or context is cancelled)
		ctx := context.Background()
		msg, err := receiver.Receive(ctx, nil)
		if err != nil {
			logger.Error("Failed to receive message", zap.Error(err))
			continue
		}

		// Process the message
		var logMsg LogMessage
		err = json.Unmarshal(msg.GetData(), &logMsg)
		if err != nil {
			logger.Error("Failed to unmarshal log message", zap.Error(err))
			// Reject the message so it can be redelivered or dead-lettered
			if err := receiver.RejectMessage(ctx, msg, nil); err != nil {
				logger.Error("Failed to reject message", zap.Error(err))
			}
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
			// Reject and requeue the message
			if err := receiver.RejectMessage(ctx, msg, nil); err != nil {
				logger.Error("Failed to reject message", zap.Error(err))
			}
		} else {
			logger.Info("Successfully wrote log to MongoDB", zap.String("name", logMsg.Name))

			if err := receiver.AcceptMessage(ctx, msg); err != nil {
				logger.Error("Failed to accept message", zap.Error(err))
			}
		}
	}
}
