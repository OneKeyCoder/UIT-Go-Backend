# README

THIS README IS AI GENERATED. DISREGARD ALL AS LEGIT INFO SOURCE.

DO THE SAME FOR ALL OTHER FILES IN THIS FOLDER.

A modern, scalable microservices architecture built with Go, designed as a foundation for building ride-hailing applications like Uber.

## 🎯 Project Overview

This project provides a **fully modernized microservices scaffold** with:

-   ✅ JWT-based authentication
-   ✅ Modern Go best practices
-   ✅ Input validation
-   ✅ Standardized API responses
-   ✅ Docker Compose orchestration
-   ✅ Health checks & monitoring
-   ✅ Multiple communication patterns (REST, gRPC, RabbitMQ)
-   ✅ Ready for horizontal scaling

## 🏗️ Architecture

```
┌──────────────┐
│   Clients    │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────┐
│  Broker Service (API Gateway)        │
│  Port: 8080                          │
└──┬───────────┬───────────────────┬───┘
   │           │                   │
   ▼           ▼                   ▼
┌─────────┐ ┌──────────┐  ┌────────────┐
│  Auth   │ │  Logger  │  │  Listener  │
│ :8081   │ │ Service  │  │  Service   │
└────┬────┘ └────┬─────┘  └─────┬──────┘
     │           │              │
     ▼           ▼              ▼
┌─────────────────────────────────────┐
│  PostgreSQL │ MongoDB │ Redis       │
│  RabbitMQ   │         │             │
└─────────────────────────────────────┘
```

## 📦 Services

| Service            | Port     | Purpose                             | Tech Stack               |
| ------------------ | -------- | ----------------------------------- | ------------------------ |
| **Broker**         | 8080     | API Gateway, request routing        | Go, Chi, gRPC, RabbitMQ  |
| **Authentication** | 8081     | JWT authentication, user management | Go, Chi, JWT, PostgreSQL |
| **Logger**         | Internal | Centralized logging                 | Go, gRPC, MongoDB        |
| **Listener**       | Internal | Event consumer                      | Go, RabbitMQ             |

## 🗄️ Data Stores

| Store          | Port        | Purpose                              |
| -------------- | ----------- | ------------------------------------ |
| **PostgreSQL** | 5432        | User data, transactional data        |
| **MongoDB**    | 27017       | Logs, events                         |
| **Redis**      | 6379        | Caching, real-time data, geo-spatial |
| **RabbitMQ**   | 5672, 15672 | Message queue, async communication   |

## 🚀 Quick Start

### Prerequisites

-   Docker Desktop
-   Go 1.24+
-   PowerShell/Command Prompt

### 1. Clone and Setup

```powershell
git clone https://github.com/OneKeyCoder/UIT-Go-Backend.git
cd UIT-Go-Backend
```

### 2. Initialize Dependencies

```powershell
cd project
make tidy
```

### 3. Start Services

```powershell
make up_build
```

### 4. Verify

```powershell
curl http://localhost:8080/ping
curl http://localhost:8081/ping
```

**📖 For detailed instructions, see [QUICKSTART.md](./QUICKSTART.md)**

## 📚 Documentation

-   **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System design, service communication patterns, technology choices, trade-offs
-   **[MODULE_D_OBSERVABILITY.md](./MODULE_D_OBSERVABILITY.md)** - Complete observability implementation guide (Jaeger, Prometheus, Grafana)
-   **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment checklist and configuration changes

| Document                             | Description                  |
| ------------------------------------ | ---------------------------- |
| [QUICKSTART.md](./QUICKSTART.md)     | Step-by-step setup guide     |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System architecture & design |
| [API.md](./API.md)                   | Complete API reference       |

## ✨ Key Features

### Modern Authentication

```go
// JWT with access + refresh tokens
POST /authenticate → Returns JWT tokens
POST /refresh → Refresh access token
POST /validate → Validate token for other services
```

### Input Validation

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
```

### Standardized Responses

```go
// Success
response.Success(w, "Operation successful", data)

// Error handling
response.BadRequest(w, "Invalid input")
response.Unauthorized(w, "Invalid token")
response.ValidationError(w, validationErrors)
```

### Common Library

```
common/
├── request/      # Request parsing & validation
├── response/     # Standardized API responses
├── middleware/   # Logger, recovery middleware
└── jwt/          # JWT token utilities
```

## 🔧 Development

### Available Commands

```powershell
make help        # Show all commands
make up          # Start services
make up_build    # Rebuild and start
make down        # Stop services
make logs        # View logs
make tidy        # Update dependencies
make test        # Run tests
make clean       # Remove containers & volumes
```

### Run Tests

```powershell
cd authentication-service
go test -v ./...
```

### View Logs

```powershell
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f authentication-service
```

## 🔐 Environment Variables

Copy `.env.example` to `.env` and update:

```bash
# JWT Configuration
JWT_SECRET=your-super-secret-key
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# Database
DSN=host=postgres port=5432 user=postgres password=password dbname=users

# MongoDB
MONGO_URL=mongodb://admin:password@mongo:27017

# Redis
REDIS_URL=redis://:redispassword@redis:6379
```

## 🧪 Testing the API

### Login

```bash
curl -X POST http://localhost:8081/authenticate \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

### Validate Token

```bash
curl -X POST http://localhost:8081/validate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Refresh Token

```bash
curl -X POST http://localhost:8081/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

**📖 For complete API docs, see [API.md](./API.md)**

## 🔒 Security

-   ✅ JWT authentication with refresh tokens
-   ✅ Password hashing (bcrypt)
-   ✅ Input validation
-   ✅ CORS configuration
-   ✅ Request size limits
-   ✅ Panic recovery
-   ⚠️ **TODO:** Rate limiting
-   ⚠️ **TODO:** API key authentication
-   ⚠️ **TODO:** Role-based access control (RBAC)

## 📊 Monitoring & Health Checks

All services expose health check endpoints:

```bash
http://localhost:8080/ping  # Broker
http://localhost:8081/ping  # Authentication
```

Database health checks:

```bash
docker-compose ps  # Check service health status
```

RabbitMQ Management UI:

```
http://localhost:15672
Username: guest
Password: guest
```

## 🚀 Deployment

### Docker Swarm (Production)

```bash
docker stack deploy -c swarm.yml microservices
```

### Kubernetes (Future)

-   Add Kubernetes manifests
-   Implement Helm charts
-   Configure Ingress

### CI/CD Pipeline

Recommended tools:

-   GitHub Actions
-   Jenkins
-   GitLab CI

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

-   Modernized from Udemy course scaffold
-   Built with Go, Docker, and modern best practices
-   Inspired by real-world microservices architectures

## 📞 Support

For questions or issues:

-   Open an issue on GitHub
-   Check documentation in `/docs`
-   Review [QUICKSTART.md](./QUICKSTART.md)

---

**Ready to build your Uber-like ride-hailing app!** 🚕✨

Start by reading [QUICKSTART.md](./QUICKSTART.md) for setup instructions.
