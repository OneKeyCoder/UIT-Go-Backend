# Module D: Observability Implementation

## Table of Contents

-   [Overview](#overview)
-   [Architecture](#architecture)
-   [Quick Start](#quick-start)
-   [Testing](#testing)
-   [Endpoints Reference](#endpoints-reference)
-   [SLO/SLI Definitions](#slosli-definitions)
-   [Troubleshooting](#troubleshooting)

---

## Overview

This module implements **production-grade observability** using the three pillars:

-   **Traces:** OpenTelemetry → Jaeger (distributed tracing)
-   **Metrics:** Prometheus (RED metrics: Rate, Errors, Duration)
-   **Logs:** Structured logging with Zap (correlated via trace_id)

### Technology Stack

| Component         | Version  | Purpose                            |
| ----------------- | -------- | ---------------------------------- |
| **Jaeger**        | 2.2.0    | Distributed tracing UI and storage |
| **Prometheus**    | 3.2.0    | Time-series metrics database       |
| **Grafana**       | 11.5.0   | Visualization and dashboarding     |
| **OpenTelemetry** | SDK v1.x | Trace instrumentation              |

---

## Architecture

### Observability Flow

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────────────────────┐
│              API Gateway (Port 8080)                     │
│  • OpenTelemetry spans created                          │
│  • Prometheus metrics (/metrics)                        │
│  • Context propagation to downstream services           │
└──────┬──────────────────────┬────────────────────────────┘
       │ gRPC + trace ctx     │ gRPC + trace ctx
       ▼                      ▼
┌──────────────────┐   ┌──────────────────┐
│  Auth Service    │   │  Logger Service  │
│  • Receive span  │   │  • Receive span  │
│  • Add metrics   │   │  • Add metrics   │
│  • Log w/traceID │   │  • Log w/traceID │
└──────────────────┘   └──────────────────┘
       │                      │
       ▼                      ▼
┌─────────────────────────────────────────┐
│         Jaeger (Port 4317 OTLP)         │
│  • Receives traces via OTLP gRPC        │
│  • UI on port 16686                     │
└─────────────────────────────────────────┘
       ▲
       │ Scrape /metrics every 15s
       │
┌──────────────────────────┐
│  Prometheus (Port 9090)  │
│  • Scrapes all services  │
│  • Stores time-series    │
└──────────────────────────┘
       ▲
       │ Query metrics
       │
┌──────────────────────────┐
│  Grafana (Port 3000)     │
│  • SLO Dashboard         │
│  • admin/admin           │
└──────────────────────────┘
```

### Why This Architecture?

**1. OpenTelemetry OTLP Export**

-   ✅ Vendor-neutral (can switch from Jaeger to other backends)
-   ✅ gRPC protocol (high performance)
-   ✅ Context propagation across service boundaries

**2. Prometheus Pull Model**

-   ✅ Service discovery via docker-compose (internal DNS)
-   ✅ No credential management needed for push
-   ✅ Target health monitoring built-in

**3. gRPC Interceptors**

-   ✅ Automatic trace context extraction
-   ✅ Centralized error handling
-   ✅ No manual instrumentation per RPC

---

## Quick Start

### 1. Start All Services

```powershell
cd "d:\I fucking hate my life\UIT-Go\project"
make up_build
```

Wait ~30 seconds for healthchecks to pass.

### 2. Verify Observability Stack

```powershell
# Check all containers running
docker ps

# Should see 11 containers:
# - api-gateway, authentication-service, logger-service
# - jaeger, prometheus, grafana
# - postgres, mongo, redis, rabbitmq
```

### 3. Access UIs

| Service    | URL                    | Credentials   |
| ---------- | ---------------------- | ------------- |
| Grafana    | http://localhost:3000  | admin / admin |
| Prometheus | http://localhost:9090  | -             |
| Jaeger     | http://localhost:16686 | -             |

---

## Testing

### 1. Generate Traffic

Use Postman or curl to send requests:

```powershell
# Authentication request (creates full trace)
curl -X POST http://localhost:8080/handle `
  -H "Content-Type: application/json" `
  -d '{
    "action": "auth",
    "auth": {
      "email": "admin@example.com",
      "password": "verysecret"
    }
  }'

# Logging request
curl -X POST http://localhost:8080/handle `
  -H "Content-Type: application/json" `
  -d '{
    "action": "log",
    "log": {
      "name": "test-event",
      "data": "observability test"
    }
  }'
```

### 2. View Traces in Jaeger

1. Open http://localhost:16686
2. Select service: **api-gateway**
3. Click "Find Traces"
4. Click on a trace to see:
    - Root span: `HandleSubmission`
    - Child span: `authenticateViaGRPC` or `logItemViaGRPCClient`
    - Timing breakdown
    - gRPC metadata

**Expected Auth Trace:**

```
HandleSubmission (400ms)
└── authenticateViaGRPC (395ms)
    └── AuthService.Authenticate (390ms)
        ├── bcrypt hashing (200ms)
        ├── DB query (100ms)
        └── JWT generation (50ms)
```

### 3. Check Metrics in Prometheus

1. Open http://localhost:9090
2. Try these queries:

```promql
# Request rate by service
sum(rate(http_requests_total[5m])) by (service)

# P95 latency for API Gateway
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="api-gateway"}[5m])) by (le))

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# In-flight requests
http_requests_in_flight{service="api-gateway"}
```

### 4. View Grafana Dashboard

1. Open http://localhost:3000 (admin/admin)
2. Navigate to: Dashboards → API Gateway SLO Dashboard
3. Panels show:
    - **SLO Compliance:** % of successful requests (target: 99.9%)
    - **P95 Latency:** Auth operations (target: <500ms)
    - **Error Budget:** Remaining error allowance
    - **Request Rate:** Traffic by endpoint
    - **Status Codes:** 2xx/4xx/5xx breakdown
    - **Latency Heatmap:** Distribution over time

---

## Endpoints Reference

### API Gateway (Port 8080)

| Endpoint     | Method | Purpose                |
| ------------ | ------ | ---------------------- |
| `/`          | GET    | Health check           |
| `/ping`      | GET    | Health check           |
| `/metrics`   | GET    | Prometheus metrics     |
| `/handle`    | POST   | Orchestration endpoint |
| `/grpc/auth` | POST   | Direct gRPC auth       |
| `/grpc/log`  | POST   | Direct gRPC log        |

### Authentication Service

**HTTP Endpoints (Port 8081):**
| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/authenticate` | POST | User login |
| `/register` | POST | User registration |
| `/validate` | POST | JWT validation |
| `/refresh` | POST | Token refresh |
| `/metrics` | GET | Prometheus metrics |

**gRPC (Port 50051):**

-   `AuthService.Authenticate`
-   `AuthService.ValidateToken`

### Logger Service

**HTTP Endpoints (Port 8082):**
| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/log` | POST | Write log entry |
| `/metrics` | GET | Prometheus metrics |

**gRPC (Port 50052):**

-   `LogService.WriteLog`

---

## SLO/SLI Definitions

### Service Level Objectives (SLOs)

#### 1. API Gateway Availability SLO

-   **Target:** 99.9% (three nines)
-   **Measurement Window:** 5 minutes
-   **Definition:** Percentage of non-5xx responses
-   **Query:**
    ```promql
    sum(rate(http_requests_total{service="api-gateway", status!~"5.."}[5m]))
    /
    sum(rate(http_requests_total{service="api-gateway"}[5m]))
    * 100
    ```

#### 2. API Gateway P95 Latency SLO

-   **Target:** <500ms
-   **Measurement Window:** 5 minutes
-   **Definition:** 95th percentile response time
-   **Query:**
    ```promql
    histogram_quantile(0.95,
      sum(rate(http_request_duration_seconds_bucket{service="api-gateway"}[5m]))
      by (le)
    ) * 1000
    ```

#### 3. Error Budget

-   **Calculation:** (1 - actual_error_rate - target_slo) / allowed_error_rate \* 100
-   **Target SLO:** 99.9% (0.1% allowed errors)
-   **Query:**
    ```promql
    (1
      - sum(rate(http_requests_total{service="api-gateway", status=~"5.."}[5m]))
        / sum(rate(http_requests_total{service="api-gateway"}[5m]))
      - 0.999
    ) / 0.001 * 100
    ```

### Service Level Indicators (SLIs)

| SLI              | Metric                     | Good/Bad                    |
| ---------------- | -------------------------- | --------------------------- |
| **Availability** | HTTP status codes          | Good: 2xx/3xx/4xx, Bad: 5xx |
| **Latency**      | Request duration histogram | Good: <500ms, Bad: ≥500ms   |
| **Throughput**   | Requests per second        | N/A (capacity indicator)    |
| **Saturation**   | In-flight requests         | Warning: >100               |

---

## Troubleshooting

### Traces Not Appearing

**Symptom:** Jaeger UI shows no traces for services

**Checks:**

1. Verify OTLP endpoint configured:

    ```powershell
    docker logs project-api-gateway-1 | Select-String "OTEL_COLLECTOR_ENDPOINT"
    # Should show: jaeger:4317
    ```

2. Check Jaeger receiving traces:

    ```powershell
    docker logs project-jaeger-1 | Select-String "OTLP"
    ```

3. Verify spans created:
    ```powershell
    docker logs project-api-gateway-1 | Select-String "span"
    ```

**Solution:**

-   Restart services: `docker-compose restart api-gateway authentication-service logger-service`

### Metrics Showing "No data"

**Symptom:** Prometheus queries return empty or Grafana shows "No data"

**Checks:**

1. Verify Prometheus scraping:

    ```
    http://localhost:9090/targets
    ```

    - All targets should show "UP" state
    - Check "Last Scrape" timestamp is recent

2. Test metrics endpoint directly:
    ```powershell
    curl http://localhost:8080/metrics
    ```

**Solution:**

-   Generate traffic (metrics only appear after requests)
-   Check Prometheus config: `project/prometheus.yml`

### Dashboard Shows NaN

**Symptom:** Histogram percentile queries show "NaN"

**Cause:** Not enough data points for histogram calculation (need multiple buckets)

**Solution:**

-   Send at least 10-20 requests to generate histogram data
-   Wait 15 seconds for Prometheus scrape
-   Refresh Grafana dashboard

### Authentication Takes ~400ms (Expected Behavior)

**This is NOT a bug!** Auth latency breakdown:

-   Bcrypt password hashing: ~200ms (intentionally slow for security)
-   Database query: ~100ms
-   JWT token generation: ~50ms
-   gRPC overhead: ~30ms

**Why slow?**

-   Bcrypt is designed to be computationally expensive (prevents brute-force attacks)
-   Industry standard: 200-300ms for bcrypt is considered secure
-   Logging is fast (~2ms) because it's async fire-and-forget

---

## Performance Characteristics

### Expected Latencies (P95)

| Operation        | Target | Typical | Notes             |
| ---------------- | ------ | ------- | ----------------- |
| `/handle` (auth) | <500ms | ~400ms  | Bcrypt dominates  |
| `/handle` (log)  | <50ms  | ~5ms    | Async, no DB wait |
| `/grpc/auth`     | <500ms | ~395ms  | Direct gRPC call  |
| `/grpc/log`      | <50ms  | ~2ms    | Fire-and-forget   |

### Throughput Capacity

**Single Instance:**

-   Auth: ~10 req/s (bcrypt bottleneck)
-   Logging: ~500 req/s (async)

**Horizontal Scaling:**

-   API Gateway: Stateless, linear scaling
-   Auth Service: Scale out for more CPU (bcrypt parallel)
-   Logger Service: Queue-based, scale consumers independently

---

## Development vs Production

### Current Setup (Development Mode)

**Port Exposure:**

-   ✅ All services expose HTTP ports (8080, 8081, 8082)
-   ✅ Databases accessible from host (5432, 27017, 6379)
-   ✅ Jaeger/Prometheus/Grafana UIs public

**Security:**

-   ⚠️ `otlptracegrpc.WithInsecure()` (no TLS)
-   ⚠️ Default passwords (JWT_SECRET, DB credentials)
-   ⚠️ CORS allows all origins

### Production Changes Required

See **[DEPLOYMENT.md](./DEPLOYMENT.md)** for complete checklist:

-   Remove internal service HTTP ports (8081, 8082)
-   Enable TLS for OTLP export
-   Use secrets management (HashiCorp Vault, AWS Secrets Manager)
-   Restrict CORS to allowed origins
-   Add rate limiting
-   Enable Prometheus remote write (long-term storage)

---

## Metrics Reference

### Exported Metrics

All services expose these standard metrics:

**HTTP Metrics:**

```
http_requests_total{service, path, method, status}
http_request_duration_seconds_bucket{service, path, method, le}
http_requests_in_flight{service}
```

**gRPC Metrics (via interceptors):**

```
grpc_server_handled_total{service, method, code}
grpc_server_handling_seconds{service, method}
```

**Process Metrics (Go runtime):**

```
go_goroutines
go_memstats_alloc_bytes
go_gc_duration_seconds
```

---

## Next Steps

1. **Add Custom Metrics:**

    - Business metrics (user signups, active sessions)
    - Database connection pool stats
    - Cache hit/miss rates

2. **Set Up Alerts:**

    - Prometheus Alertmanager
    - Alert on SLO violations
    - PagerDuty/Slack integration

3. **Long-term Storage:**

    - Prometheus remote write to Thanos/Cortex
    - Jaeger storage backend to Cassandra/Elasticsearch

4. **Synthetic Monitoring:**
    - Blackbox exporter for uptime checks
    - Distributed load testing (k6, Locust)

---

## Resources

-   **OpenTelemetry Go SDK:** https://opentelemetry.io/docs/languages/go/
-   **Prometheus Querying:** https://prometheus.io/docs/prometheus/latest/querying/basics/
-   **Grafana Dashboards:** https://grafana.com/docs/grafana/latest/dashboards/
-   **SLO Best Practices:** https://sre.google/workbook/implementing-slos/
