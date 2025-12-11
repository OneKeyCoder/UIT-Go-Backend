#!/bin/sh
# RabbitMQ Queue Initialization Script
# This script creates the required queues after RabbitMQ starts

echo "Waiting for RabbitMQ to be ready..."
sleep 10

echo "Creating queues..."

# Declare the logs queue (durable, not auto-delete)
rabbitmqadmin declare queue name=logs durable=true auto_delete=false || {
    echo "Failed to create logs queue, trying with rabbitmqctl..."
    # Fallback to rabbitmqctl if rabbitmqadmin is not available
    rabbitmqctl eval 'rabbit_amqqueue:declare({resource, <<"/">>, queue, <<"logs">>}, true, false, [], none, <<"guest">>).'
}

echo "Queue initialization complete!"
