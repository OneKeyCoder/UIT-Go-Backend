# Modernized Architecture Overview

## Current Microservices Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Client Applications                         │
│                     (Web, Mobile, Third-party)                       │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ HTTP/HTTPS
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Broker Service (API Gateway)                  │
│                           Port: 8080                                 │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ • Request routing & orchestration                           │   │
│  │ • CORS handling                                              │   │
│  │ • Logging & recovery middleware                             │   │
│  │ • Request validation                                         │   │
│  └─────────────────────────────────────────────────────────────┘   │
└───┬─────────────┬──────────────┬─────────────┬────────────────────┘
    │             │              │             │
    │             │              │             │
    ▼             ▼              ▼             ▼
┌─────────┐  ┌─────────┐  ┌──────────┐  ┌─────────────┐
│  Auth   │  │ Logger  │  │ Listener │  │   Future    │
│ Service │  │ Service │  │ Service  │  │  Services   │
│:8081    │  │         │  │          │  │             │
└────┬────┘  └────┬────┘  └────┬─────┘  └─────────────┘
     │            │            │
     │            │            │
     ▼            ▼            ▼
┌─────────────────────────────────────────────────────────┐
│                    Data Layer                            │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
│  │Postgres │  │ MongoDB │  │  Redis  │  │RabbitMQ │   │
│  │  :5432  │  │ :27017  │  │  :6379  │  │  :5672  │   │
│  │         │  │         │  │         │  │         │   │
│  │  Users  │  │  Logs   │  │ Caching │  │  Events │   │
│  │   DB    │  │   DB    │  │  & Geo  │  │  Queue  │   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Service Details

### 1. Broker Service (API Gateway)

**Port:** 8080  
**Technology:** Go, Chi Router, gRPC, RabbitMQ  
**Purpose:** Central entry point for all client requests

**Features:**

-   Request routing to appropriate microservices
-   CORS configuration
-   Common middleware (logging, recovery)
-   Request validation using validator/v10
-   Integration with RabbitMQ for async messaging
-   gRPC client for logger service

**Endpoints:**

-   `POST /` - Health check
-   `POST /handle` - Main request handler (routes based on action)
-   `POST /log-grpc` - gRPC logging

### 2. Authentication Service

**Port:** 8081  
**Technology:** Go, Chi Router, JWT, PostgreSQL  
**Purpose:** User authentication and JWT token management

**Features:**

-   JWT-based authentication (access + refresh tokens)
-   Password hashing with bcrypt
-   Token validation for other services
-   Environment-based JWT configuration
-   Input validation

**Endpoints:**

-   `POST /authenticate` - Login and get JWT tokens
-   `POST /refresh` - Refresh access token
-   `POST /validate` - Validate JWT token
-   `GET /ping` - Health check

**Database:** PostgreSQL (users table)

### 3. Logger Service

**Port:** Internal  
**Technology:** Go, Chi Router, gRPC, MongoDB  
**Purpose:** Centralized logging for all services

**Features:**

-   HTTP REST API for logging
-   gRPC server for high-performance logging
-   RPC interface for legacy support
-   MongoDB for log persistence
-   Structured log storage

**Endpoints:**

-   `POST /log` - Write log entry
-   gRPC service for remote logging

**Database:** MongoDB (logs collection)

### 4. Listener Service

**Technology:** Go, RabbitMQ  
**Purpose:** Event-driven message consumer

**Features:**

-   Consumes messages from RabbitMQ
-   Event-driven architecture support
-   Async processing
-   Integration with logger service

## Common Library Package

```
common/
├── request/
│   └── request.go      # Request parsing & validation
├── response/
│   └── response.go     # Standardized API responses
├── middleware/
│   └── middleware.go   # Logger & Recovery middleware
└── jwt/
    └── jwt.go          # JWT generation & validation
```

**Shared Across All Services:**

-   Standardized request/response handling
-   Input validation using go-playground/validator
-   JWT token utilities
-   Common middleware (logging, panic recovery)
-   Reduces code duplication

## Data Stores

### PostgreSQL

-   **Purpose:** Relational data (users, rides, drivers)
-   **Version:** 16-alpine
-   **Features:** ACID compliance, complex queries
-   **Future Use:** User accounts, ride data, driver profiles

### MongoDB

-   **Purpose:** Document storage (logs, events)
-   **Version:** 7-jammy
-   **Features:** Flexible schema, fast writes
-   **Current Use:** Application logs

### Redis

-   **Purpose:** Caching & real-time data
-   **Version:** 7-alpine
-   **Features:** In-memory, geo-spatial queries
-   **Future Use:**
    -   Real-time driver location tracking
    -   Session management
    -   Rate limiting
    -   Cache frequently accessed data

### RabbitMQ

-   **Purpose:** Message queue for async communication
-   **Version:** 3.13-management-alpine
-   **Features:** Reliable messaging, pub/sub
-   **Current Use:** Event-driven communication between services

## Communication Patterns

### 1. Synchronous HTTP/REST

```
Client → Broker Service → Authentication Service
         ↓
         Response with JWT tokens
```

### 2. Asynchronous Messaging (RabbitMQ)

```
Service A → RabbitMQ → Listener Service → Process Event
```

### 3. gRPC (High Performance)

```
Broker Service → gRPC → Logger Service (fast logging)
```

### 4. RPC (Remote Procedure Call)

```
Service → RPC Client → Logger Service (legacy support)
```

## Security Features

### Authentication

-   ✅ JWT-based authentication
-   ✅ Access tokens (24h expiry)
-   ✅ Refresh tokens (168h expiry)
-   ✅ Password hashing (bcrypt)
-   ✅ Token validation endpoint for other services

### Middleware

-   ✅ CORS configuration
-   ✅ Request size limits (1MB)
-   ✅ Panic recovery
-   ✅ Structured logging

### Input Validation

-   ✅ Validator/v10 for struct validation
-   ✅ Email format validation
-   ✅ Password length requirements
-   ✅ Detailed validation error responses

## Health Checks & Monitoring

### Docker Health Checks

```yaml
PostgreSQL: pg_isready (10s interval)
MongoDB: mongosh ping (10s interval)
Redis: redis-cli ping (10s interval)
RabbitMQ: rabbitmq-diagnostics (30s interval)
Services: HTTP endpoint checks (30s interval)
```

### Service Health Endpoints

-   All services expose `/ping` endpoint
-   Returns HTTP 200 if healthy
-   Used by load balancers and monitoring tools

## Scalability Considerations

### Horizontal Scaling

-   All services are stateless (except databases)
-   Can run multiple instances behind a load balancer
-   RabbitMQ ensures message delivery to only one consumer

### Database Scaling

-   PostgreSQL: Read replicas for queries
-   MongoDB: Replica sets for high availability
-   Redis: Redis Cluster for distributed cache

### Caching Strategy (with Redis)

```
Request → Check Redis → Cache Hit? → Return cached data
                      → Cache Miss → Query DB → Cache result → Return
```

## Future Architecture (Uber-like Features)

```
┌────────────────────────────────────────────────────────────┐
│                   New Services to Add                      │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐          │
│  │   Rider    │  │   Driver   │  │     Ride     │          │
│  │  Service   │  │  Service   │  │   Service    │          │
│  │            │  │            │  │              │          │
│  │ • Profile  │  │ • Profile  │  │ • Matching   │          │
│  │ • History  │  │ • Vehicle  │  │ • Tracking   │          │
│  │ • Ratings  │  │ • Earnings │  │ • Fares      │          │
│  └────────────┘  └────────────┘  └──────────────┘          │
│                                                            │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐          │
│  │ Location   │  │  Payment   │  │Notification  │          │
│  │  Service   │  │  Service   │  │   Service    │          │
│  │            │  │            │  │              │          │
│  │ • GPS      │  │ • Stripe   │  │ • WebSocket  │          │
│  │ • Redis    │  │ • Invoices │  │ • Push       │          │
│  │   Geo      │  │ • Wallet   │  │ • SMS/Email  │          │
│  └────────────┘  └────────────┘  └──────────────┘          │
└────────────────────────────────────────────────────────────┘
```

### Recommended Tech Stack for New Services

**Location Service:**

-   Redis Geo-spatial commands (GEOADD, GEORADIUS)
-   WebSocket for real-time updates
-   Sub-second response times

**Payment Service:**

-   Stripe/PayPal integration
-   Idempotency for payment operations
-   Event sourcing for transaction history

**Notification Service:**

-   WebSocket for real-time notifications
-   Firebase Cloud Messaging for push
-   Queue-based for reliability

## Development Workflow

```
1. Code Changes
   ↓
2. Local Testing (go test)
   ↓
3. Docker Build (make up_build)
   ↓
4. Integration Testing
   ↓
5. Git Commit & Push
   ↓
6. CI/CD Pipeline (future)
   ↓
7. Deploy to Staging
   ↓
8. Deploy to Production
```

## Best Practices Implemented

✅ **Microservices Patterns**

-   Single responsibility per service
-   Shared common library
-   API Gateway pattern

✅ **Modern Go Practices**

-   Struct tags for validation
-   Context for request cancellation
-   Error wrapping with %w

✅ **Security**

-   JWT authentication
-   Password hashing
-   Input validation
-   CORS configuration

✅ **Observability**

-   Structured logging
-   Health checks
-   Centralized logging

✅ **Docker Best Practices**

-   Health checks
-   Restart policies
-   Volume management
-   Network isolation

---

This architecture is production-ready and scalable for building your Uber-like ride-hailing application! 🚀
