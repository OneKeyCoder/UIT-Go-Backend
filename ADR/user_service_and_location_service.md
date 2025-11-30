# User Service
- Which database should use? Why?

We chose **Azure DocumentDB** for the User Service because:
- Our user data may change in the future, and NoSQL provides flexibility.
- We don't need to store many tables and rules, only needing to store two tables: `user` and `vehicle`. This makes sharding simple and we do not many `join`s.
- User service team is familiar with MongoDB, using a MongoDB-compatible DB allowing us to develop quickly and focus more on the overall system architecture.
- DocumentDB can scale indepently in both computing and storage, unlike tiered scale of Mongo Atlas, so over-scaling is not an issue.
- Is **completely open-source**, unlike Mongo Atlas. Technically when we switch provider we can keep using CosmosDB.

# Location Service
- Why use Redis but not DynamoDb?

Location service will handle **very high load** and frequent DB read/write, with geospacial features, and **does not require persistent data**. As such, an in-memory DB would fit the bill. Redis is the first contender.

We could, and wanted to, use Valkey instead, since it's completely open-source and not bound to an enterprise entity (backed by Linux Foundation), but since unlike AWS, Azure does not provide a Valkey Managed service, we use Redis instead.
