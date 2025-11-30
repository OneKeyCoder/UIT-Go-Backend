# Message queue - Azure Service Bus

Originally, the app is developed to use RabbitMQ 3, with AMQP 0-9-1, which is the older standard. We could run a cluster inside ACA, but since configuring such a cluster would require tremendous work both in intial config and in management (around 3 docs pages, each really really long) to ensure a healthy cluster, we decided against it.

So, we migrated the app to use AMQP 1.0, the newer protocol, which both RabbitMQ 4+, and more importantly Azure Service Bus supports, so that we can use RabbitMQ 4 in local dev `docker-compose`, and **Azure Service Bus** when deploying. This also mitigate vendor-lockin since when moving away from Azure we can just drop in an AMQP1.0-compatible service bus, or RabbitMQ 4 itself.
