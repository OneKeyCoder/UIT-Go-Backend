# Location Service

Location Service là một microservice trong hệ thống UIT-Go-Backend, chịu trách nhiệm quản lý vị trí địa lý real-time của users và cung cấp các tính năng geospatial.

## Tính năng

- **Real-time Location Tracking**: Lưu trữ và cập nhật vị trí của users
- **Geospatial Queries**: Tìm kiếm users gần nhau dựa trên khoảng cách
- **Auto-expiry**: Tự động xóa location data cũ với TTL
- **High Performance**: Sử dụng Redis GEO cho queries sub-millisecond
- **gRPC & HTTP**: Cung cấp cả gRPC và HTTP APIs

## Architecture

### Technology Stack
- **Language**: Go 1.25
- **Cache**: Redis 7 (với GEO support)
- **Protocol**: gRPC + HTTP/REST
- **Observability**: OpenTelemetry, Prometheus metrics
- **Framework**: Gin (HTTP), gRPC

### Data Model

```go
type CurrentLocation struct {
    UserID    string  `json:"user_id"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Distance  float64 `json:"distance,omitempty"`
    Speed     float64 `json:"speed"`
    Heading   string  `json:"heading"`
    Timestamp string  `json:"timestamp"`
}
```

## API Endpoints

### HTTP Endpoints

#### Set Location
```bash
POST /location
Content-Type: application/json

{
  "user_id": "user123",
  "latitude": 10.762622,
  "longitude": 106.660172,
  "speed": 45.5,
  "heading": "NE",
  "timestamp": "2025-10-08T10:30:00Z"
}
```

#### Get Location
```bash
GET /location?user_id=user123
```

#### Find Nearest Users
```bash
GET /location/nearest?user_id=user123&top_n=10&radius=5
# top_n: số lượng users muốn tìm (default: 10)
# radius: bán kính tìm kiếm tính bằng km (default: 10)
```

#### Get All Locations
```bash
GET /location/all
```

### gRPC Methods

- `SetLocation(SetLocationRequest) → SetLocationResponse`
- `GetLocation(GetLocationRequest) → GetLocationResponse`
- `FindNearestUsers(FindNearestUsersRequest) → FindNearestUsersResponse`
- `GetAllLocations(GetAllLocationsRequest) → GetAllLocationsResponse`

## Configuration

Environment Variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis hostname | `redis` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `redispassword` |
| `REDIS_DB` | Redis database number | `0` |
| `REDIS_TIME_TO_LIVE` | Location TTL (seconds) | `3600` |
| `OTEL_EXPORTER` | Telemetry exporter type | `otlp` |
| `OTEL_COLLECTOR_ENDPOINT` | OpenTelemetry endpoint | `jaeger:4317` |

## Usage Examples

### Via API Gateway

```bash
# Set location
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "location",
    "location": {
      "action": "set",
      "user_id": "user123",
      "latitude": 10.762622,
      "longitude": 106.660172,
      "speed": 45.5,
      "heading": "NE",
      "timestamp": "2025-10-08T10:30:00Z"
    }
  }'

# Get location
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "location",
    "location": {
      "action": "get",
      "user_id": "user123"
    }
  }'

# Find nearest users
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "location",
    "location": {
      "action": "nearest",
      "user_id": "user123",
      "top_n": 10,
      "radius": 5.0
    }
  }'

# Get all locations
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "location",
    "location": {
      "action": "all"
    }
  }'
```

## Development

### Build

```bash
cd location-service
go build -o locationServiceApp ./cmd/api
```

### Run Locally

```bash
# Set environment variables
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=redispassword

# Run
./locationServiceApp
```

### Docker Build

```bash
docker build -f location-service.dockerfile -t location-service .
```

## Redis Data Structure

### Location Data (String)
```
Key: {user_id}
Value: JSON {user_id, latitude, longitude, speed, heading, timestamp}
TTL: REDIS_TIME_TO_LIVE seconds
```

### Geospatial Index (Sorted Set)
```
Key: geo:users
Members: {user_id} with coordinates (longitude, latitude)
```

## Performance

- **Latency**: < 1ms for location operations (Redis in-memory)
- **Throughput**: 10,000+ req/s per instance
- **TTL Auto-cleanup**: Prevents memory bloat
- **Geospatial Queries**: O(N+log(M)) complexity

## Use Cases

1. **Ride-sharing Apps**: Find nearest drivers
2. **Social Networking**: Discover nearby users
3. **Delivery Services**: Track courier locations
4. **Fleet Management**: Monitor vehicle positions
5. **Location-based Marketing**: Target users by proximity

## Monitoring

### Health Checks
- Liveness: `GET /health/live`
- Readiness: `GET /health/ready`

### Metrics (Prometheus)
- `GET /metrics`
- HTTP request duration
- gRPC call latency
- Redis operation metrics

### Tracing (Jaeger)
- Distributed tracing enabled via OpenTelemetry
- View traces at http://localhost:16686

## Error Handling

- Invalid coordinates → 400 Bad Request
- User not found → 404 Not Found
- Redis connection error → 500 Internal Server Error
- TTL expired location → 404 Not Found (auto-deleted)
