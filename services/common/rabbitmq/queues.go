package rabbitmq

// Default queue names used across services
const (
	// LogsQueue is the queue for log messages
	LogsQueue = "logs"

	// Default AMQP 1.0 address format for queues
	// RabbitMQ expects /queues/<queue_name> format for AMQP 1.0
	LogsQueueAddress = "/queues/logs"
)

// DefaultQueues returns the list of queues that should be created on startup
func DefaultQueues() []QueueConfig {
	return []QueueConfig{
		{
			Name:       LogsQueue,
			Durable:    true,
			AutoDelete: false,
		},
		// Add more queues here as needed
		// {
		// 	Name:       "notifications",
		// 	Durable:    true,
		// 	AutoDelete: false,
		// },
	}
}
