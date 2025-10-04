# UIT-Go Backend - Quick Start Guide 🚀

## Prerequisites

-   Docker Desktop installed and running
-   Go 1.24+ installed
-   PowerShell or Command Prompt

## Getting Started

### 1. Initialize Go Modules

Open PowerShell in the project root and run:

```powershell
# Navigate to common package
cd common
go mod tidy

# Navigate to each service and tidy dependencies
cd ..\authentication-service
go mod tidy

cd ..\broker-service
go mod tidy

cd ..\logger-service
go mod tidy

cd ..\listener-service
go mod tidy

# Back to project directory
cd ..\project
```

Or use the Makefile:

```powershell
cd project
make tidy
```

### 2. Start Services with Docker Compose

```powershell
cd project

# Start services (first time will download images)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

Or use Makefile commands:

```powershell
# Start services
make up

# Rebuild and start
make up_build

# Show logs
make logs

# Stop services
make down
```

### 3. Verify Services are Running

Check health endpoints:

```powershell
# Broker service
curl http://localhost:8080/ping

# Authentication service
curl http://localhost:8081/ping
```

Check RabbitMQ Management UI:

-   Open http://localhost:15672
-   Username: `guest`
-   Password: `guest`

### 4. Test Authentication Service

#### Register/Login (you'll need to create a user in the database first)

For testing, you can insert a test user directly into PostgreSQL:

```powershell
# Connect to PostgreSQL container
docker exec -it uit-go-postgres-1 psql -U postgres -d users

# In psql, create a test user table and insert a user:
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    user_active INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

# Insert a test user (password: password123)
# Note: You'll need to hash this properly in production
INSERT INTO users (email, first_name, last_name, password, user_active)
VALUES ('test@example.com', 'Test', 'User', '$2a$10$YourHashedPasswordHere', 1);

# Exit psql
\q
```

#### Test Authentication Endpoint

```powershell
# Login
curl -X POST http://localhost:8081/authenticate `
  -H "Content-Type: application/json" `
  -d '{\"email\": \"test@example.com\", \"password\": \"password123\"}'
```

Expected response:

```json
{
    "error": false,
    "message": "Authentication successful",
    "data": {
        "user": {
            "id": 1,
            "email": "test@example.com",
            "first_name": "Test",
            "last_name": "User"
        },
        "tokens": {
            "access_token": "eyJhbGciOiJIUzI1NiIs...",
            "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
        }
    }
}
```

#### Test Token Validation

```powershell
# Replace YOUR_ACCESS_TOKEN with the token from login response
curl -X POST http://localhost:8081/validate `
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### Test Token Refresh

```powershell
# Replace YOUR_REFRESH_TOKEN with the refresh token from login response
curl -X POST http://localhost:8081/refresh `
  -H "Content-Type: application/json" `
  -d '{\"refresh_token\": \"YOUR_REFRESH_TOKEN\"}'
```

### 5. Test Broker Service

```powershell
# Test broker endpoint
curl -X POST http://localhost:8080/ `
  -H "Content-Type: application/json"

# Test handle submission (authentication via broker)
curl -X POST http://localhost:8080/handle `
  -H "Content-Type: application/json" `
  -d '{\"action\": \"auth\", \"auth\": {\"email\": \"test@example.com\", \"password\": \"password123\"}}'
```

## Common Commands

### Docker Commands

```powershell
# View running containers
docker ps

# View all containers (including stopped)
docker ps -a

# View logs for specific service
docker-compose logs -f authentication-service
docker-compose logs -f broker-service
docker-compose logs -f logger-service

# Restart a specific service
docker-compose restart authentication-service

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ removes database data)
docker-compose down -v

# Rebuild a specific service
docker-compose up -d --build authentication-service
```

### Go Commands

```powershell
# Run tests
cd authentication-service
go test -v ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Update dependencies
go mod tidy
go mod vendor  # Optional: create vendor folder
```

## Troubleshooting

### Services won't start

1. Check if Docker Desktop is running
2. Check ports aren't already in use:
    ```powershell
    netstat -ano | findstr :8080
    netstat -ano | findstr :8081
    netstat -ano | findstr :5432
    netstat -ano | findstr :27017
    ```

### Database connection errors

```powershell
# Check if PostgreSQL is healthy
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Connect to PostgreSQL manually
docker exec -it uit-go-postgres-1 psql -U postgres -d users
```

### RabbitMQ connection errors

```powershell
# Check RabbitMQ status
docker-compose ps rabbitmq

# View RabbitMQ logs
docker-compose logs rabbitmq

# Wait for RabbitMQ to fully start (can take 30-60 seconds)
```

### Module not found errors

```powershell
# Clean and rebuild modules
cd common
go clean -modcache
go mod tidy

# Repeat for each service
cd ..\authentication-service
go mod tidy
```

## Service URLs

| Service             | URL                    | Purpose                  |
| ------------------- | ---------------------- | ------------------------ |
| Broker              | http://localhost:8080  | API Gateway              |
| Authentication      | http://localhost:8081  | User authentication      |
| PostgreSQL          | localhost:5432         | User database            |
| MongoDB             | localhost:27017        | Logs database            |
| Redis               | localhost:6379         | Caching & real-time data |
| RabbitMQ            | localhost:5672         | Message queue            |
| RabbitMQ Management | http://localhost:15672 | RabbitMQ UI              |

## Next Steps

1. ✅ Services are running
2. 📖 Read `MODERNIZATION.md` for details on changes
3. 🏗️ Start building your Uber-like microservices:
    - Rider Service
    - Driver Service
    - Ride Service
    - Location Service (use Redis for geo-spatial queries)
    - Payment Service
    - Notification Service

## Additional Resources

-   [Docker Compose Documentation](https://docs.docker.com/compose/)
-   [Go Chi Router](https://github.com/go-chi/chi)
-   [Go Validator](https://github.com/go-playground/validator)
-   [golang-jwt](https://github.com/golang-jwt/jwt)
-   [Redis Geo Commands](https://redis.io/commands/?group=geo)
-   [gRPC Documentation](https://grpc.io/docs/)
-   [Protocol Buffers](https://developers.google.com/protocol-buffers)

## Quick Test gRPC Endpoints

**Authenticate via gRPC:**

```bash
curl -X POST http://localhost:8080/grpc/auth \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"verysecret"}'
```

**Log via gRPC:**

```bash
curl -X POST http://localhost:8080/grpc/log \
  -H "Content-Type: application/json" \
  -d '{"name":"test-event","data":"Hello from gRPC!"}'
```

### gRPC Architecture

```
External Clients (HTTP)
        ↓
   Broker Service (Port 8080)
        ↓ (gRPC - internal)
   ┌────────────────┬────────────────┐
   ↓                ↓
Authentication    Logger
Service           Service
(Port 50051)      (Port 50052)
```

### See Full Documentation

-   **proto/auth/auth.proto** - Authentication service contract
-   **proto/logger/logger.proto** - Logger service contract

---

Happy coding! 🎉 Phase 3 Complete! 🚀
