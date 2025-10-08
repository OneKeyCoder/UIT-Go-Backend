# Kiến Trúc Hệ Thống UIT-Go-Backend

## Tổng Quan

UIT-Go-Backend là một hệ thống microservices được xây dựng bằng Go, áp dụng các best practices về kiến trúc phân tán, giao tiếp gRPC, event-driven architecture với RabbitMQ, và observability với OpenTelemetry, Prometheus, Grafana.

## Sơ Đồ Kiến Trúc

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Client/Browser                             │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ HTTP/REST (Port 8080)
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          API Gateway                                 │
│  - HTTP Server (Port 80)                                            │
│  - Rate Limiting (100 req/min)                                      │
│  - CORS Middleware                                                  │
│  - Metrics & Logging                                                │
│  - Health Checks (/health/live, /health/ready)                     │
└────────────┬────────────┬──────────────────┬─────────────────────────┘
             │ gRPC       │ gRPC             │ gRPC
             │ (50051)    │ (50052)          │ (50053)
             ▼            ▼                  ▼
┌──────────────────┐  ┌───────────────┐  ┌──────────────────────┐
│Authentication Svc│  │  Logger Svc   │  │  Location Service    │
│- HTTP (Port 80)  │  │- HTTP (80)    │  │  - HTTP (Port 80)    │
│- gRPC (50051)    │  │- gRPC (50052) │  │  - gRPC (Port 50053) │
│- JWT Tokens      │  │- MongoDB      │  │  - GeoSpatial Cache  │
│- Event Pub       │  │- RabbitMQ Sub │  │  - Real-time Tracking│
└────┬─────────────┘  └───┬───────────┘  └───────┬──────────────┘
     │                     │                      │
     │ SQL                 │ NoSQL                │ Cache
     ▼                     ▼                      ▼
┌─────────┐          ┌─────────┐            ┌─────────┐
│PostgreSQL│          │ MongoDB │            │  Redis  │
│  (5432) │          │ (27017) │            │ (6379)  │
└─────────┘          └─────────┘            └─────────┘
                          │                      
                          │ AMQP                 
                          ▼                      
                   ┌────────────────────┐
                   │     RabbitMQ       │
                   │  Exchange: logs_topic│
                   │  Routing: log.INFO │
                   │  (Ports 5672/15672)│
                   └────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                     Supporting Services                              │
├──────────────┬──────────────┬──────────────┬────────────────────────┤
│ Redis        │ Jaeger       │ Prometheus   │ Grafana                │
│ (Cache)      │ (Tracing)    │ (Metrics)    │ (Visualization)        │
│ Port: 6381   │ UI: 16686    │ Port: 9090   │ Port: 3000             │
│              │ OTLP: 4317   │              │                        │
└──────────────┴──────────────┴──────────────┴────────────────────────┘
```

## Chi Tiết Các Microservices

### 1. API Gateway (Port 8080)

**Công dụng:**
- Điểm vào duy nhất (Single Entry Point) cho toàn bộ hệ thống
- Định tuyến các request từ client đến các microservice tương ứng
- Áp dụng rate limiting để bảo vệ hệ thống khỏi quá tải
- Xử lý CORS cho phép các ứng dụng web từ domain khác truy cập API
- Thu thập metrics và logs cho tất cả request đi qua
- Health check endpoint cho monitoring

**Cổng giao tiếp:**
- HTTP Server: Port 80 (exposed qua 8080)
- gRPC Client: Kết nối đến Authentication (50051), Logger (50052), và Location (50053)

**Các endpoint chính:**
- `GET /ping` - Ping endpoint đơn giản
- `POST /` - Broker endpoint
- `POST /handle` - Handle submission với routing động
  - `action: "auth"` - Authentication qua gRPC
  - `action: "log"` - Logging qua gRPC
  - `action: "location"` - Location operations qua gRPC
- `POST /grpc/auth` - Authentication qua gRPC
- `POST /grpc/log` - Logging qua gRPC
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

**Luồng xử lý:**
1. Nhận HTTP request từ client
2. Validate payload với validation tags
3. Route request đến service tương ứng qua gRPC
4. Nhận response từ service backend
5. Trả về HTTP response cho client

**Dependencies:**
- Authentication Service (gRPC)
- Logger Service (gRPC)
- Location Service (gRPC)
- Common packages (logger, middleware, telemetry)

---

### 2. Authentication Service (Port 50051)

**Công dụng:**
- Quản lý authentication và authorization
- Tạo và validate JWT tokens (Access Token & Refresh Token)
- Quản lý thông tin user trong PostgreSQL
- Publish authentication events đến RabbitMQ
- Cung cấp các endpoint để đăng ký, đăng nhập, làm mới token

**Cổng giao tiếp:**
- HTTP Server: Port 80 (internal)
- gRPC Server: Port 50051

**Database:**
- PostgreSQL (Port 5432) - Lưu trữ thông tin users
- Schema: `users` table với các trường: id, email, first_name, last_name, password (hashed), user_active, created_at, updated_at

**gRPC Services:**
- `Authenticate(email, password)` → Returns User + JWT Tokens
- `ValidateToken(token)` → Returns validation result + claims
- `RefreshToken(refresh_token)` → Returns new token pair

**HTTP Endpoints:**
- `POST /authenticate` - Đăng nhập
- `POST /register` - Đăng ký user mới
- `POST /refresh` - Làm mới access token
- `POST /validate` - Validate JWT token
- `POST /change-password` - Đổi mật khẩu

**Luồng xử lý Authentication:**
1. Nhận request authentication (email + password)
2. Query user từ PostgreSQL database
3. Verify password với bcrypt
4. Generate JWT token pair (Access Token 24h, Refresh Token 7d)
5. Publish event "user.login" đến RabbitMQ
6. Trả về user info + tokens

**Event Publishing:**
- Exchange: `logs_topic` (type: topic)
- Routing Key: `log.INFO`
- Message format: JSON `{name: string, data: string}`

**Security:**
- Password hashing: bcrypt
- JWT signing: HS256
- JWT Secret: Configurable via environment variable
- Token expiry: Configurable (default: 24h access, 7d refresh)

---

### 3. Logger Service (Port 50052)

**Công dụng:**
- Centralized logging cho toàn bộ hệ thống
- Lưu trữ logs vào MongoDB
- Consumer events từ RabbitMQ (event-driven logging)
- Cung cấp API để query logs
- Real-time log ingestion

**Cổng giao tiếp:**
- HTTP Server: Port 80 (internal)
- gRPC Server: Port 50052
- RabbitMQ Consumer: Queue binding to `logs_topic`

**Database:**
- MongoDB (Port 27017) - Database: `logs`, Collection: `logs`
- Schema: `{_id, name, data, created_at, updated_at}`

**gRPC Services:**
- `WriteLog(name, data)` → Writes log to MongoDB
- `GetLogs(limit)` → Retrieves logs from MongoDB

**HTTP Endpoints:**
- `POST /log` - Write a log entry

**RabbitMQ Consumer:**
- Exchange: `logs_topic`
- Queue: Auto-generated exclusive queue
- Routing Key: `log.INFO`
- Auto-ack: true

**Luồng xử lý Logging:**

**Cách 1: Gọi trực tiếp qua gRPC**
1. Service gọi WriteLog gRPC method
2. Logger Service nhận request
3. Insert log vào MongoDB
4. Trả về response success/failure

**Cách 2: Event-driven qua RabbitMQ (Async)**
1. Service publish event đến RabbitMQ exchange `logs_topic`
2. RabbitMQ route message đến logger queue
3. Logger Service consumer nhận message
4. Parse JSON message
5. Insert log vào MongoDB
6. Acknowledge message

**Ưu điểm RabbitMQ approach:**
- Non-blocking: Service không cần chờ log được ghi
- Decoupling: Service không phụ thuộc trực tiếp vào Logger Service

---

### 4. Location Service (Port 50053)

**Công dụng:**
- Quản lý vị trí địa lý real-time của users
- Tính toán khoảng cách và tìm kiếm users gần nhau
- Cung cấp geospatial queries với Redis GEO
- Tracking user movement với timestamp
- Cache location data với TTL để tự động expire

**Cổng giao tiếp:**
- HTTP Server: Port 80 (internal)
- gRPC Server: Port 50053

**Database:**
- Redis (Port 6379) - In-memory cache cho high-performance geospatial queries
- Data structures:
  - String: `{user_id}` → JSON location data
  - GeoSet: `geo:users` → Geospatial index với coordinates

**gRPC Services:**
- `SetLocation(user_id, lat, lon, speed, heading, timestamp)` → Stores user location
- `GetLocation(user_id)` → Retrieves user's current location
- `FindNearestUsers(user_id, top_n, radius)` → Finds N nearest users within radius
- `GetAllLocations()` → Retrieves all stored locations

**HTTP Endpoints:**
- `POST /location` - Set user location
- `GET /location?user_id={id}` - Get user location
- `GET /location/nearest?user_id={id}&top_n={n}&radius={km}` - Find nearest users
- `GET /location/all` - Get all locations
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

**Location Model:**
```go
type CurrentLocation struct {
    UserID    string  `json:"user_id"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Distance  float64 `json:"distance,omitempty"` // Calculated distance
    Speed     float64 `json:"speed"`              // km/h
    Heading   string  `json:"heading"`            // N, NE, E, SE, S, SW, W, NW
    Timestamp string  `json:"timestamp"`          // ISO 8601
}
```

**Luồng xử lý SetLocation:**
1. Nhận location data từ client
2. Validate coordinates (latitude: -90 to 90, longitude: -180 to 180)
3. Store JSON data vào Redis với key `user_id` và TTL (default 3600s)
4. Add coordinates vào Redis GEO index `geo:users`
5. Trả về confirmation

**Luồng xử lý FindNearestUsers:**
1. Nhận user_id, top_n (default 10), radius (default 10km)
2. Get current user's position từ GEO index
3. Query Redis GEORADIUS để tìm users trong bán kính
4. Sort theo distance tăng dần
5. Exclude current user khỏi kết quả
6. Lấy location detail cho mỗi user từ Redis
7. Calculate và attach distance
8. Trả về top N results

**Redis Operations:**
- `SET {user_id} {json_data} EX {ttl}` - Store location with expiry
- `GEOADD geo:users {lon} {lat} {user_id}` - Add to geospatial index
- `GET {user_id}` - Retrieve location data
- `GEORADIUS geo:users {lon} {lat} {radius} km WITHDIST WITHCOORD ASC` - Find nearby

**Configuration:**
- `REDIS_HOST` - Redis hostname (default: "redis")
- `REDIS_PORT` - Redis port (default: "6379")
- `REDIS_PASSWORD` - Redis password (default: "redispassword")
- `REDIS_DB` - Redis database number (default: 0)
- `REDIS_TIME_TO_LIVE` - Location TTL in seconds (default: 3600)

**Use Cases:**
- Real-time location tracking cho ride-sharing apps
- Find nearby users cho social networking
- Geofencing và proximity alerts
- Location-based recommendations
- Fleet tracking và logistics

**Performance:**
- Redis GEO queries: O(N+log(M)) complexity
- In-memory operations: sub-millisecond response time
- Auto-expiry: TTL prevents stale data buildup
- Horizontal scaling: Redis Cluster support

- Reliability: Message được persist trong RabbitMQ nếu Logger Service down
- Scalability: Có thể có nhiều Logger Service consumers

---

## Common Packages (Shared Libraries)

### 1. `common/jwt`
**Công dụng:** Quản lý JWT token generation và validation

**Functions:**
- `GenerateTokenPair()` - Tạo cả access và refresh token
- `GenerateToken()` - Tạo một JWT token
- `ValidateToken()` - Validate và parse JWT claims
- `RefreshAccessToken()` - Tạo access token mới từ refresh token

**Token Claims:**
```go
{
  user_id: int
  email: string
  role: string
  exp: timestamp  // Expiration time
  iat: timestamp  // Issued at
  nbf: timestamp  // Not before
}
```

### 2. `common/logger`
**Công dụng:** Structured logging với zap

**Features:**
- Structured logging với JSON format (production)
- Color-coded console logging (development)
- Context-aware logging (trace_id, span_id)
- Service name tagging
- Log levels: Info, Error, Warn, Debug, Fatal

**Usage:**
```go
logger.InitDefault("service-name")
logger.Info("message", zap.String("key", "value"))
logger.Error("error message", zap.Error(err))
```

### 3. `common/middleware`
**Công dụng:** HTTP middleware cho logging, recovery, metrics

**Middleware:**
- `Logger`: Log mọi HTTP request với method, path, status, duration
- `Recovery`: Recover từ panic và log stack trace
- `PrometheusMetrics`: Thu thập HTTP metrics cho Prometheus

### 4. `common/telemetry`
**Công dụng:** Distributed tracing với OpenTelemetry

**Features:**
- Initialize tracer với OTLP exporter (Jaeger)
- Fallback to stdout exporter (development)
- Automatic trace context propagation
- Span creation helpers

**Usage:**
```go
shutdown, _ := telemetry.InitTracer("service-name", "version")
defer shutdown(ctx)

ctx, span := telemetry.StartSpan(ctx, "operation-name")
defer span.End()
```

### 5. `common/grpcutil`
**Công dụng:** gRPC interceptors cho logging và tracing

**Interceptors:**
- `UnaryServerInterceptor`: Server-side logging cho gRPC calls
- `UnaryClientInterceptor`: Client-side logging cho gRPC calls

### 6. `common/request` và `common/response`
**Công dụng:** Helpers cho HTTP request/response handling

**Request:**
- `ReadAndValidate()`: Đọc JSON và validate với struct tags
- `Validate()`: Validate struct với validator
- `HandleError()`: Xử lý validation errors

**Response:**
- `Success()`: Return JSON success response
- `BadRequest()`: Return 400 error
- `Unauthorized()`: Return 401 error
- `InternalServerError()`: Return 500 error

---

## Protocol Buffers (Proto Definitions)

### 1. `proto/auth/auth.proto`
Định nghĩa gRPC service và messages cho Authentication

**Services:**
- `AuthService.Authenticate`
- `AuthService.ValidateToken`
- `AuthService.RefreshToken`

### 2. `proto/logger/logger.proto`
Định nghĩa gRPC service và messages cho Logger

**Services:**
- `LoggerService.WriteLog`
- `LoggerService.GetLogs`

---

## Infrastructure Services

### 1. PostgreSQL (Port 5432)
- Database cho Authentication Service
- Stores: Users, credentials
- Init script: `init_db.sql`
- Connection pooling: Max 25 connections, 5 idle

### 2. MongoDB (Port 27017)
- Database cho Logger Service
- Stores: Log entries
- Connection pooling: Max 50, Min 10

### 3. Redis (Port 6381)
- Caching layer (sẵn sàng để sử dụng)
- Password: `redispassword`
- Có thể dùng cho: Session storage, Rate limiting cache, etc.

### 4. RabbitMQ (Ports 5672, 15672)
- Message broker cho event-driven architecture
- Exchange: `logs_topic` (type: topic)
- Management UI: http://localhost:15672
- Credentials: guest/guest

### 5. Jaeger (Ports 16686, 4317, 4318)
- Distributed tracing backend
- UI: http://localhost:16686
- OTLP Receiver: Port 4317 (gRPC), 4318 (HTTP)
- Visualize request flows across services

### 6. Prometheus (Port 9090)
- Metrics collection và storage
- Scrapes metrics from `/metrics` endpoints
- Retention: Configurable
- UI: http://localhost:9090

### 7. Grafana (Port 3000)
- Metrics visualization
- Pre-configured với Prometheus data source
- UI: http://localhost:3000
- Credentials: admin/admin

---

## Luồng Dữ Liệu Trong Hệ Thống

### Luồng 1: User Authentication

```
Client
  │
  │ POST /handle {"action": "auth", "auth": {...}}
  ▼
API Gateway (Port 8080)
  │
  │ 1. Validate request payload
  │ 2. Extract auth payload
  ▼
  │ gRPC: Authenticate(email, password)
  ▼
Authentication Service (Port 50051)
  │
  │ 1. Query user from PostgreSQL
  │ 2. Verify password with bcrypt
  │ 3. Generate JWT token pair
  │ 4. Publish event to RabbitMQ
  │    - Exchange: logs_topic
  │    - Routing: log.INFO
  │    - Message: {name: "user.login", data: "..."}
  │
  ▼
RabbitMQ
  │
  │ Route message to Logger queue
  ▼
Logger Service Consumer
  │
  │ 1. Receive message
  │ 2. Parse JSON
  │ 3. Insert to MongoDB
  │
  ▼
MongoDB (logs collection)

Response flow:
Authentication Service
  │ Return: {success, message, user, tokens}
  ▼
API Gateway
  │ Format HTTP response
  ▼
Client
  │ Receive: {message, user, tokens}
```

### Luồng 2: Direct Logging via gRPC

```
Client
  │
  │ POST /handle {"action": "log", "log": {...}}
  ▼
API Gateway
  │
  │ 1. Validate request
  │ 2. Extract log payload
  ▼
  │ gRPC: WriteLog(name, data)
  ▼
Logger Service (Port 50052)
  │
  │ 1. Receive gRPC request
  │ 2. Create LogEntry object
  │ 3. Insert to MongoDB
  │    - Database: logs
  │    - Collection: logs
  │    - Document: {name, data, created_at, updated_at}
  │
  ▼
MongoDB

Response:
Logger Service
  │ Return: {success, message}
  ▼
API Gateway
  │ Format HTTP response
  ▼
Client
  │ Receive: {message: "Logged via gRPC"}
```

### Luồng 3: Event-Driven Logging (Async)

```
Authentication Service
  │
  │ After successful login
  ▼
  │ Publish event
  │ - Channel: RabbitMQ channel
  │ - Exchange: logs_topic
  │ - Routing Key: log.INFO
  │ - Body: JSON {name, data}
  ▼
RabbitMQ Exchange (logs_topic)
  │
  │ Route based on routing key
  ▼
Logger Service Queue (auto-generated)
  │
  │ Consumer listening...
  ▼
Logger Service Consumer Goroutine
  │
  │ 1. Unmarshal JSON message
  │ 2. Create LogEntry
  │ 3. Insert to MongoDB
  │ 4. Auto-acknowledge message
  │
  ▼
MongoDB (logs collection)
```

### Luồng 4: Distributed Tracing

```
Request arrives at API Gateway
  │
  │ Trace ID: auto-generated or propagated
  │ Span: "HandleSubmission"
  ▼
API Gateway creates span
  │
  │ Context with trace info
  ▼
  │ gRPC call with trace context
  ▼
Authentication Service receives
  │
  │ Extract trace context
  │ Create child span: "Authenticate"
  ▼
Database query
  │
  │ Child span: "GetUserByEmail"
  ▼
All spans exported to Jaeger
  │
  │ OTLP gRPC (Port 4317)
  ▼
Jaeger Collector
  │
  │ Store and index traces
  ▼
Jaeger UI (Port 16686)
  │
  │ Visualize entire request flow
  │ - API Gateway → Auth Service → Database
  │ - Timing for each span
  │ - Error tracking
```

### Luồng 5: Metrics Collection

```
HTTP Request arrives
  │
  ▼
PrometheusMetrics Middleware
  │
  │ Increment counters:
  │ - http_requests_total{service, method, path, status}
  │ Record histograms:
  │ - http_request_duration_seconds{service, method, path}
  │
  ▼
Process request
  │
  ▼
Response sent
  │
  ▼
Metrics stored in memory
  │
  │ Exposed at /metrics endpoint
  ▼
Prometheus scrapes /metrics
  │
  │ Every 15s (configurable)
  ▼
Prometheus TSDB
  │
  │ Store time-series data
  ▼
Grafana queries Prometheus
  │
  │ Display dashboards:
  │ - Request rate
  │ - Error rate
  │ - Latency percentiles
  │ - Service health
```

---

## Observability Stack

### Logging (Structured)
- **Tool:** Zap logger
- **Format:** JSON (production), Console (dev)
- **Fields:** timestamp, level, service, message, trace_id, span_id
- **Storage:** Console output + MongoDB (via Logger Service)

### Tracing (Distributed)
- **Tool:** OpenTelemetry + Jaeger
- **Protocol:** OTLP gRPC
- **Sampling:** Always sample (development)
- **Features:** 
  - Request flow visualization
  - Latency analysis
  - Error tracking
  - Service dependency map

### Metrics (Time-series)
- **Tool:** Prometheus
- **Metrics types:**
  - Counter: Total requests, errors
  - Histogram: Request duration
  - Gauge: Active connections
- **Visualization:** Grafana dashboards
- **Alerting:** Prometheus Alertmanager (configured in `prometheus-alerts.yml`)

---

## Deployment

### Docker Compose
- File: `project/docker-compose.yml`
- Services: 10 containers
- Networks: Default bridge network (all services can communicate)
- Volumes: Persistent data for databases

### Build Commands (Makefile)
```bash
make up_build    # Build and start all services
make up          # Start services without rebuild
make down        # Stop all services
make logs        # View logs
make clean       # Remove all containers and volumes
make tidy        # Run go mod tidy on all services
make test        # Run tests
```

### Service Dependencies
```
API Gateway depends on:
  → Authentication Service
  → Logger Service
  → Jaeger

Authentication Service depends on:
  → PostgreSQL
  → RabbitMQ
  → Jaeger

Logger Service depends on:
  → MongoDB
  → RabbitMQ
  → Jaeger
```

---

## Security

### Authentication
- JWT-based authentication
- Secure password hashing (bcrypt cost 14)
- Token expiration (24h access, 7d refresh)
- Token validation on each protected request

### Network Security
- Internal services không expose port ra ngoài
- Chỉ API Gateway expose Port 8080
- gRPC communication internal only
- Database credentials trong environment variables

### Rate Limiting
- API Gateway: 100 requests/minute per IP
- Burst allowance: 100 requests
- IP detection: X-Forwarded-For, X-Real-IP, RemoteAddr

---

## Scalability Considerations

### Horizontal Scaling
- **Stateless services:** API Gateway, Auth Service, Logger Service có thể scale ra nhiều instance
- **Load balancing:** Cần thêm load balancer (nginx, traefik) trước API Gateway
- **Session management:** Sử dụng Redis cho distributed sessions

### Database Scaling
- **PostgreSQL:** Master-slave replication, connection pooling
- **MongoDB:** Replica set, sharding
- **Redis:** Redis Cluster cho high availability

### Message Queue
- **RabbitMQ:** Clustering, mirrored queues
- **Multiple consumers:** Logger Service có thể có nhiều consumers

### Monitoring
- Auto-scaling based on metrics (CPU, memory, request rate)
- Health checks cho Kubernetes/Docker Swarm
- Readiness và liveness probes đã implement

---

## Tổng Kết

Hệ thống UIT-Go-Backend là một kiến trúc microservices hiện đại với:

✅ **Separation of Concerns:** Mỗi service có trách nhiệm rõ ràng  
✅ **Scalability:** Stateless services, horizontal scaling ready  
✅ **Observability:** Full stack logging, tracing, metrics  
✅ **Resilience:** Health checks, graceful shutdown, error handling  
✅ **Event-Driven:** Async communication với RabbitMQ  
✅ **Performance:** gRPC cho inter-service communication, connection pooling  
✅ **Security:** JWT authentication, rate limiting, password hashing  
✅ **Developer Experience:** Shared common libraries, structured logging, easy deployment  

Hệ thống sẵn sàng cho production với các best practices về microservices architecture.
