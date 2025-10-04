# UIT-Go-Backend Microservices Architecture

## 📋 Project Overview

This is a microservices-based backend system built with Go, demonstrating modern distributed architecture patterns including service communication via HTTP, RPC, gRPC, and message queues (RabbitMQ).

### Architecture Summary
The system consists of **5 microservices** that work together:
- **Broker Service** - API Gateway/Entry point
- **Authentication Service** - User authentication
- **Logger Service** - Centralized logging
- **Listener Service** - Event consumer
- **Mail Service** - Email notifications (scaffold only)

---

## 🏗️ Services Deep Dive

### 1. BROKER SERVICE (API Gateway)
**Purpose**: Acts as the main entry point and orchestrator for all client requests. Routes requests to appropriate microservices.

**Port**: 8080 (external) → 80 (internal)

**Dependencies**: RabbitMQ

#### File Structure & Usage:

##### `cmd/api/main.go`
- **Purpose**: Application entry point
- **Functionality**:
  - Establishes RabbitMQ connection with retry logic
  - Initializes HTTP server on port 80
  - Creates Config struct with RabbitMQ connection
- **Key Features**:
  - Exponential backoff for RabbitMQ connection (max 5 retries)
  - Graceful connection handling

##### `cmd/api/routes.go`
- **Purpose**: HTTP route definitions
- **Endpoints**:
  - `POST /` - Health check endpoint (Broker handler)
  - `POST /handle` - Main request handler (routes to other services)
  - `POST /log-grpc` - Logging via gRPC
  - `GET /ping` - Heartbeat endpoint
- **Features**:
  - CORS enabled for all origins
  - Chi router with middleware

##### `cmd/api/handlers.go`
- **Purpose**: Request handling logic
- **Functions**:
  
  1. **`Broker()`**
     - Simple health check
     - Returns: `{"error": false, "message": "Hit the broker"}`
  
  2. **`HandleSubmission()`**
     - Main router based on action type
     - Supported actions:
       - `"auth"` → calls `authenticate()`
       - `"log"` → calls `logItemViaRPC()`
       - `"mail"` → calls `sendMail()`
  
  3. **`authenticate()`**
     - Forwards auth requests to Authentication Service
     - HTTP POST to `http://authentication-service/authenticate`
     - Validates credentials and returns user data
  
  4. **`logItem()`**
     - Logs via HTTP to Logger Service
     - HTTP POST to `http://logger-service/log`
  
  5. **`logItemViaRPC()`**
     - Logs via RPC to Logger Service
     - Connects to `logger-service:5001`
     - Calls `RPCServer.LogInfo`
  
  6. **`LogViaGRPC()`**
     - Logs via gRPC to Logger Service
     - Connects to `logger-service:50001`
     - Uses protocol buffers
  
  7. **`sendMail()`**
     - Sends email via Mail Service
     - HTTP POST to `http://mail-service/send`
  
  8. **`logEventViaRabbit()`**
     - Publishes log events to RabbitMQ
     - Uses event emitter to push to queue
  
  9. **`pushToQueue()`**
     - Helper to push messages to RabbitMQ
     - Publishes to "log.INFO" topic

##### `cmd/api/helpers.go`
- **Purpose**: Utility functions
- **Functions**:
  - `readJSON()` - Parse JSON from request body
  - `writeJson()` - Write JSON response
  - `errorJson()` - Return error as JSON

##### `event/emitter.go`
- **Purpose**: RabbitMQ event publisher
- **Functionality**:
  - Creates RabbitMQ channel
  - Declares "logs_topic" exchange
  - Publishes events with severity levels
- **Key Methods**:
  - `NewEventEmitter()` - Initialize emitter
  - `Push()` - Publish event to exchange
  - `setup()` - Declare exchange

##### `event/event.go`
- **Purpose**: RabbitMQ setup utilities
- **Functionality**:
  - Declares topic exchange for logs

##### `logs/*.proto`, `logs/*.pb.go`
- **Purpose**: Protocol buffer definitions for gRPC
- **Functionality**:
  - Defines LogService gRPC interface
  - Message structures for log entries

---

### 2. AUTHENTICATION SERVICE
**Purpose**: Handles user authentication and credential validation.

**Port**: 8081 (external) → 80 (internal)

**Dependencies**: PostgreSQL database

#### File Structure & Usage:

##### `cmd/api/main.go`
- **Purpose**: Application entry point
- **Functionality**:
  - Connects to PostgreSQL with retry logic (max 10 retries)
  - Initializes database models
  - Starts HTTP server on port 80
- **Key Features**:
  - Database connection pooling
  - 2-second retry interval for DB connection
  - Uses pgx PostgreSQL driver

##### `cmd/api/routes.go`
- **Purpose**: HTTP route definitions
- **Endpoints**:
  - `POST /authenticate` - User login endpoint
  - `GET /ping` - Heartbeat endpoint
- **Features**:
  - CORS enabled
  - Chi router

##### `cmd/api/handlers.go`
- **Purpose**: Authentication logic
- **Functions**:
  
  1. **`Authenticate()`**
     - Validates user credentials
     - Process:
       - Accepts JSON with email and password
       - Queries user from database
       - Verifies password using bcrypt
       - Logs authentication attempt via Logger Service
       - Returns user data if successful
     - Error handling for invalid credentials
  
  2. **`logRequest()`**
     - Helper to log authentication events
     - HTTP POST to `http://logger-service/log`

##### `cmd/api/helpers.go`
- **Purpose**: Utility functions (same as broker service)
- **Functions**:
  - JSON read/write helpers
  - Error response helpers

##### `data/models.go`
- **Purpose**: Database models and operations
- **Structures**:
  
  1. **`User` struct**
     - Fields: ID, Email, FirstName, LastName, Password, Active, CreatedAt, UpdatedAt
     - Password field excluded from JSON output (`json:"-"`)
  
  2. **`Models` struct**
     - Container for all data models
     - Provides access to User model
  
- **Methods**:
  
  1. **`New()`**
     - Factory function to create Models instance
     - Initializes database connection
  
  2. **`GetAll()`**
     - Retrieves all users sorted by last name
     - Returns slice of User pointers
  
  3. **`GetByEmail()`**
     - Finds user by email address
     - Used for authentication
  
  4. **`GetOne()`**
     - Retrieves single user by ID
  
  5. **`Update()`**
     - Updates user information
     - Updates timestamp automatically
  
  6. **`Delete()`**
     - Removes user from database
  
  7. **`DeleteByID()`**
     - Removes user by ID
  
  8. **`Insert()`**
     - Creates new user
     - Hashes password using bcrypt
     - Returns new user ID
  
  9. **`ResetPassword()`**
     - Updates user password
     - Hashes new password
  
  10. **`PasswordMatches()`**
      - Validates password against stored hash
      - Uses bcrypt comparison
      - Returns boolean for match status

---

### 3. LOGGER SERVICE
**Purpose**: Centralized logging system supporting multiple protocols (HTTP, RPC, gRPC).

**Port**: 80 (internal only)

**Additional Ports**: 
- RPC: 5001
- gRPC: 50001

**Dependencies**: MongoDB

#### File Structure & Usage:

##### `cmd/api/main.go`
- **Purpose**: Application entry point
- **Functionality**:
  - Connects to MongoDB with authentication
  - Registers RPC server
  - Starts three servers concurrently:
    - HTTP server (port 80)
    - RPC server (port 5001)
    - gRPC server (port 50001)
  - Handles graceful MongoDB disconnection
- **Connection Details**:
  - MongoDB URL: `mongodb://mongo:27017`
  - Username: admin
  - Password: password

##### `cmd/api/routes.go`
- **Purpose**: HTTP route definitions
- **Endpoints**:
  - `POST /log` - Write log entry via HTTP
  - `GET /ping` - Heartbeat endpoint

##### `cmd/api/handlers.go`
- **Purpose**: HTTP logging handler
- **Functions**:
  
  1. **`WriteLog()`**
     - Accepts log entry via HTTP POST
     - Payload: `{name: string, data: string}`
     - Inserts log into MongoDB
     - Returns success/error response

##### `cmd/api/rpc.go`
- **Purpose**: RPC server implementation
- **Structures**:
  
  1. **`RPCServer`**
     - Exported type for RPC methods
  
  2. **`RPCPayload`**
     - Fields: Name, Data
  
- **Methods**:
  
  1. **`LogInfo()`**
     - Receives log payload via RPC
     - Writes directly to MongoDB logs collection
     - Returns confirmation message
     - Format: "Processed payload via RPC: {name}"

##### `cmd/api/grpc.go`
- **Purpose**: gRPC server implementation
- **Structures**:
  
  1. **`LogServer`**
     - Implements UnimplementedLogServiceServer
     - Contains Models for database access
  
- **Methods**:
  
  1. **`WriteLog()`**
     - gRPC method for logging
     - Accepts LogRequest protobuf
     - Inserts log entry to MongoDB
     - Returns LogResponse with result status
  
  2. **`gRPCListen()`**
     - Starts gRPC server on port 50001
     - Registers LogServiceServer
     - Fatal error if startup fails

##### `data/models.go`
- **Purpose**: MongoDB data models
- **Structures**:
  
  1. **`LogEntry` struct**
     - Fields: ID, Name, Data, CreatedAt, UpdatedAt
     - BSON tags for MongoDB
     - JSON tags for API responses
  
  2. **`Models` struct**
     - Container for LogEntry model
  
- **Methods**:
  
  1. **`New()`**
     - Factory function
     - Initializes MongoDB client
  
  2. **`Insert()`**
     - Creates new log entry in MongoDB
     - Database: "logs"
     - Collection: "logs"
     - Auto-sets timestamps
  
  3. **`All()`**
     - Retrieves all log entries
     - Sorted by created_at (descending)
     - 15-second timeout
     - Returns slice of LogEntry pointers
  
  4. **`GetOne()`**
     - Fetches single log by ID
     - Converts hex string to ObjectID
  
  5. **`Drop()`**
     - Removes one log entry
     - Deletes by ObjectID
  
  6. **`Update()`**
     - Updates existing log entry
     - Updates timestamp automatically

##### `logs/*.proto`, `logs/*.pb.go`
- **Purpose**: Protocol buffer definitions
- **Messages**:
  - `Log` - Log entry structure
  - `LogRequest` - Request message
  - `LogResponse` - Response message
- **Service**: LogService with WriteLog RPC

---

### 4. LISTENER SERVICE
**Purpose**: Consumes messages from RabbitMQ queue and processes them asynchronously.

**Port**: None (background service)

**Dependencies**: RabbitMQ

#### File Structure & Usage:

##### `main.go`
- **Purpose**: Application entry point
- **Functionality**:
  - Connects to RabbitMQ with retry logic
  - Creates event consumer
  - Listens to multiple topics:
    - `log.INFO`
    - `log.WARNING`
    - `log.ERROR`
  - Runs continuously in background
- **Connection Details**:
  - RabbitMQ URL: `amqp://guest:guest@rabbitmq`
  - Max retries: 5
  - Exponential backoff

##### `event/consumer.go`
- **Purpose**: RabbitMQ message consumer
- **Structures**:
  
  1. **`Consumer`**
     - Fields: conn, queueName
  
  2. **`Payload`**
     - Fields: Name, Data
  
- **Methods**:
  
  1. **`NewConsumer()`**
     - Factory function
     - Initializes consumer with RabbitMQ connection
     - Calls setup()
  
  2. **`setup()`**
     - Creates RabbitMQ channel
     - Declares "logs_topic" exchange
  
  3. **`Listen()`**
     - Main consumer loop
     - Binds queue to topics
     - Consumes messages continuously
     - Processes messages in goroutines
     - Each message triggers `handlePayload()`
  
  4. **`handlePayload()`**
     - Processes received messages
     - Switch based on payload.Name:
       - `"log"` → calls `logEvent()`
       - Other cases can be added
  
  5. **`logEvent()`**
     - Forwards log to Logger Service
     - HTTP POST to `http://logger-service/log`
     - Converts payload to JSON
     - Handles HTTP errors

##### `event/event.go`
- **Purpose**: RabbitMQ utility functions
- **Functions**:
  
  1. **`declareExchange()`**
     - Declares topic exchange
     - Name: "logs_topic"
     - Type: "topic"
     - Durable: true
  
  2. **`declareRandomQueue()`**
     - Creates temporary queue
     - Auto-generated name
     - Non-durable
     - Auto-delete when unused

---

### 5. MAIL SERVICE (Scaffold Only)
**Purpose**: Sends email notifications via SMTP. Currently a scaffold/template.

**Port**: 80 (internal only)

**Dependencies**: SMTP server (MailHog for development)

**Status**: ⚠️ Commented out in docker-compose.yml

#### File Structure & Usage:

##### `cmd/api/main.go`
- **Purpose**: Application entry point
- **Functionality**:
  - Initializes mail configuration from environment variables
  - Creates Mail instance
  - Starts HTTP server on port 80
- **Environment Variables**:
  - MAIL_DOMAIN, MAIL_HOST, MAIL_PORT
  - MAIL_USERNAME, MAIL_PASSWORD
  - MAIL_ENCRYPTION
  - FROM_NAME, FROM_ADDRESS

##### `cmd/api/routes.go`
- **Purpose**: HTTP route definitions
- **Endpoints**:
  - `POST /send` - Send email
  - `GET /ping` - Heartbeat

##### `cmd/api/handlers.go`
- **Purpose**: Email handling logic
- **Functions**:
  
  1. **`SendMail()`**
     - Accepts email request
     - Payload: From, To, Subject, Message
     - Creates Message struct
     - Calls Mailer.SendSMTPMessage()
     - Returns success/error response

##### `cmd/api/mailer.go`
- **Purpose**: Email sending implementation
- **Structures**:
  
  1. **`Mail`**
     - Configuration: Domain, Host, Port, Username, Password
     - Encryption type
     - Default From address and name
  
  2. **`Message`**
     - Email details: From, To, Subject, Data
     - Supports attachments
     - Template data mapping
  
- **Methods**:
  
  1. **`SendSMTPMessage()`**
     - Main email sending function
     - Process:
       - Builds HTML and plain text versions
       - Configures SMTP client
       - Sets encryption (TLS/SSL/None)
       - Sends email with retries
       - Handles attachments
  
  2. **`buildHTMLMessage()`**
     - Renders HTML email template
     - Uses Go templates
     - Template: `templates/mail.html.gohtml`
     - Applies CSS inlining with premailer
  
  3. **`buildPlainTextMessage()`**
     - Renders plain text version
     - Template: `templates/mail.plain.gohtml`
  
  4. **`getEncryption()`**
     - Converts string to mail encryption type
     - Supports: TLS, SSL, None

##### `cmd/api/helpers.go`
- **Purpose**: Utility functions (standard JSON helpers)

##### `templates/mail.html.gohtml`
- **Purpose**: HTML email template
- **Features**:
  - Styled HTML layout
  - Dynamic message content
  - Responsive design

##### `templates/mail.plain.gohtml`
- **Purpose**: Plain text email template
- **Features**:
  - Simple text format
  - Fallback for non-HTML clients

---

## 🔄 Communication Patterns

### 1. HTTP/REST
- **Used by**: All services
- **Purpose**: Synchronous request/response
- **Example**: Broker → Authentication Service

### 2. RPC (Remote Procedure Call)
- **Used by**: Broker ↔ Logger
- **Port**: 5001
- **Purpose**: Fast, binary protocol for internal communication
- **Implementation**: Go's net/rpc package

### 3. gRPC
- **Used by**: Broker ↔ Logger
- **Port**: 50001
- **Purpose**: High-performance, type-safe communication
- **Implementation**: Protocol Buffers + gRPC

### 4. Message Queue (RabbitMQ)
- **Used by**: Broker → Listener → Logger
- **Purpose**: Asynchronous, decoupled event processing
- **Topics**: log.INFO, log.WARNING, log.ERROR
- **Pattern**: Publish/Subscribe with topic exchange

---

## 📊 Data Stores

### PostgreSQL
- **Used by**: Authentication Service
- **Database**: users
- **Tables**: users
- **Port**: 5432
- **Credentials**: postgres/password

### MongoDB
- **Used by**: Logger Service
- **Database**: logs
- **Collection**: logs
- **Port**: 27017
- **Credentials**: admin/password

### RabbitMQ
- **Used by**: Broker, Listener
- **Exchange**: logs_topic (type: topic)
- **Port**: 5672
- **Credentials**: guest/guest

---

## 🚀 Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.x (for development)

### Running the System

```bash
# Navigate to project directory
cd project

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Service URLs
- Broker Service: http://localhost:8080
- Authentication Service: http://localhost:8081

---

## 📝 API Examples

### Authenticate User
```bash
POST http://localhost:8080/handle
Content-Type: application/json

{
  "action": "auth",
  "auth": {
    "email": "user@example.com",
    "password": "password123"
  }
}
```

### Log via HTTP
```bash
POST http://localhost:8080/handle
Content-Type: application/json

{
  "action": "log",
  "log": {
    "name": "event",
    "data": "Something happened"
  }
}
```

### Send Email
```bash
POST http://localhost:8080/handle
Content-Type: application/json

{
  "action": "mail",
  "mail": {
    "from": "sender@example.com",
    "to": "recipient@example.com",
    "subject": "Test Email",
    "message": "This is a test message"
  }
}
```

---

## 🏗️ Project Structure Summary

```
UIT-Go-Backend/
├── authentication-service/          # User authentication
│   ├── cmd/api/
│   │   ├── main.go                 # Entry point, DB connection
│   │   ├── handlers.go             # Authentication logic
│   │   ├── helpers.go              # JSON utilities
│   │   └── routes.go               # HTTP routes
│   ├── data/
│   │   └── models.go               # User model, DB operations
│   └── authentication-service.dockerfile
│
├── broker-service/                  # API Gateway
│   ├── cmd/api/
│   │   ├── main.go                 # Entry point, RabbitMQ connection
│   │   ├── handlers.go             # Request routing, service calls
│   │   ├── helpers.go              # JSON utilities
│   │   └── routes.go               # HTTP routes
│   ├── event/
│   │   ├── emitter.go              # RabbitMQ publisher
│   │   ├── consumer.go             # RabbitMQ consumer (unused here)
│   │   └── event.go                # RabbitMQ utilities
│   ├── logs/                        # gRPC protocol buffers
│   │   ├── logs.proto              # Proto definition
│   │   ├── logs.pb.go              # Generated code
│   │   └── logs_grpc.pb.go         # Generated gRPC code
│   └── broker-service.dockerfile
│
├── logger-service/                  # Centralized logging
│   ├── cmd/api/
│   │   ├── main.go                 # Entry point, MongoDB, servers
│   │   ├── handlers.go             # HTTP log handler
│   │   ├── helpers.go              # JSON utilities
│   │   ├── routes.go               # HTTP routes
│   │   ├── rpc.go                  # RPC server
│   │   └── grpc.go                 # gRPC server
│   ├── data/
│   │   └── models.go               # LogEntry model, MongoDB ops
│   ├── logs/                        # gRPC protocol buffers
│   └── logger-service.dockerfile
│
├── listener-service/                # Event consumer
│   ├── main.go                      # Entry point, RabbitMQ listener
│   ├── event/
│   │   ├── consumer.go             # Message consumption logic
│   │   └── event.go                # RabbitMQ utilities
│   └── listener-service.dockerfile
│
├── mail-service (scaffold only)/    # Email service
│   ├── cmd/api/
│   │   ├── main.go                 # Entry point, mail config
│   │   ├── handlers.go             # Email sending handler
│   │   ├── helpers.go              # JSON utilities
│   │   ├── routes.go               # HTTP routes
│   │   └── mailer.go               # SMTP implementation
│   ├── templates/
│   │   ├── mail.html.gohtml        # HTML email template
│   │   └── mail.plain.gohtml       # Plain text template
│   └── mail-service.dockerfile
│
└── project/
    ├── docker-compose.yml           # Container orchestration
    ├── Makefile                     # Build commands
    └── swarm.yml                    # Docker Swarm configuration
```

---

## 🔧 Development Guidelines

### Adding a New Service
1. Create service directory with standard structure
2. Implement `main.go` with server initialization
3. Define routes in `routes.go`
4. Add handlers in `handlers.go`
5. Create Dockerfile
6. Add service to `docker-compose.yml`
7. Update this documentation

### Adding a New Endpoint
1. Add route in `routes.go`
2. Implement handler in `handlers.go`
3. Add any required models in `data/models.go`
4. Update API examples in documentation

### Common Patterns
- All services use Chi router
- CORS is enabled for all services
- JSON helpers are standardized across services
- Database connections use retry logic
- Error responses follow consistent format

---

## 🐛 Debugging

### View Service Logs
```bash
docker-compose logs -f [service-name]
```

### Check Service Health
```bash
# Broker
curl http://localhost:8080/ping

# Authentication
curl http://localhost:8081/ping
```

### Database Access

#### PostgreSQL
```bash
docker-compose exec postgres psql -U postgres -d users
```

#### MongoDB
```bash
docker-compose exec mongo mongosh -u admin -p password
```

#### RabbitMQ Management
- If management plugin enabled: http://localhost:15672

---

## 📚 Technology Stack

### Languages & Frameworks
- **Go 1.x** - Main programming language
- **Chi Router** - HTTP routing
- **GORM** - ORM (if extended)

### Databases
- **PostgreSQL** - Relational database for users
- **MongoDB** - Document database for logs

### Message Queue
- **RabbitMQ** - Message broker

### Communication Protocols
- **HTTP/REST** - Standard web APIs
- **RPC** - Go net/rpc package
- **gRPC** - Protocol Buffers

### Tools
- **Docker** - Containerization
- **Docker Compose** - Multi-container orchestration
- **Protocol Buffers** - Serialization for gRPC

---

## 🔐 Security Considerations

### Current Implementation
- Passwords hashed with bcrypt
- Database credentials in environment variables
- CORS enabled (currently allows all origins)

### Production Recommendations
- Implement JWT tokens for authentication
- Restrict CORS to specific origins
- Use secrets management (Vault, AWS Secrets Manager)
- Add rate limiting
- Implement request authentication between services
- Use TLS for all communications
- Enable RBAC (Role-Based Access Control)

---

## 📈 Scalability

### Horizontal Scaling
- All services are stateless (can be replicated)
- Docker Compose configured with replica count
- Load balancing can be added with nginx/traefik

### Performance Optimization
- Connection pooling for databases
- RabbitMQ for async operations
- gRPC for high-performance internal calls
- MongoDB for fast log writes

---

## 🧪 Testing

### Manual Testing
Use the API examples provided above with tools like:
- Postman
- cURL
- HTTPie

### Integration Testing
```bash
# Test authentication
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{"action":"auth","auth":{"email":"test@example.com","password":"pass"}}'

# Test logging
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{"action":"log","log":{"name":"test","data":"test data"}}'
```

---

## 📖 Additional Resources

### Go Packages Used
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/cors` - CORS middleware
- `github.com/rabbitmq/amqp091-go` - RabbitMQ client
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `golang.org/x/crypto/bcrypt` - Password hashing
- `google.golang.org/grpc` - gRPC framework
- `github.com/xhit/go-simple-mail/v2` - SMTP client

### Learning Path
1. Start with Authentication Service (simplest)
2. Understand Broker Service (routing patterns)
3. Learn Logger Service (multiple protocols)
4. Study Listener Service (async processing)
5. Explore Mail Service (external integration)

---

## 🤝 Contributing

When contributing to this project:
1. Follow the existing code structure
2. Maintain consistency across services
3. Add appropriate error handling
4. Update documentation
5. Test changes thoroughly

---

**Last Updated**: 2025-10-04
**Microservices Count**: 5
**Communication Protocols**: HTTP, RPC, gRPC, RabbitMQ
**Databases**: PostgreSQL, MongoDB
