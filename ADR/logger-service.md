# Logger Service

## Which database should use? Why?

We chose **Azure DocumentDB (MongoDB-compatible)** for the Logger Service because:
- Log data is unstructured and schema-less, making NoSQL a natural fit.
- We only need to store log entries with fields like `name`, `data`, `created_at`, `updated_at` - no complex relationships or joins required.
- DocumentDB supports auto-scaling and handles high write throughput for log ingestion.
- Is **completely open-source** (CosmosDB under the hood), avoiding vendor lock-in from MongoDB Atlas.

## Why use RabbitMQ/Azure Service Bus for log ingestion?

We use **event-driven architecture** with message queue for logging because:
- **Non-blocking**: Services don't need to wait for logs to be written, improving response time.
- **Decoupling**: Producer services don't need to know about Logger Service, they just publish events.
- **Reliability**: Messages are persisted in the queue, so logs won't be lost if Logger Service is temporarily down.
- **Scalability**: Can add multiple Logger Service consumers to handle high log volume.

Originally developed with RabbitMQ 3 (AMQP 0-9-1), we migrated to AMQP 1.0 to support both RabbitMQ 4 in local dev and **Azure Service Bus** in production, avoiding vendor lock-in.
