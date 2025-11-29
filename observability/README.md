# UIT-Go Observability Stack

## Tổng quan

Stack observability hoàn chỉnh cho hệ thống UIT-Go ride-sharing, bao gồm 3 trụ cột của observability:

| Pillar | Tool | Port | Mục đích |
|--------|------|------|----------|
| **Metrics** | Prometheus | 9090 | Thu thập và lưu trữ metrics từ các services |
| **Logs** | Loki + Promtail | 3100 | Centralized logging với label-based indexing |
| **Traces** | Jaeger | 16686 | Distributed tracing qua nhiều microservices |
| **Visualization** | Grafana | 3000 | Dashboard, alerts, và correlation giữa 3 pillars |

## Kiến trúc

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Grafana (3000)                              │
│        ┌────────────┬────────────┬────────────┬────────────┐           │
│        │ Dashboards │   Alerts   │   Explore  │   Traces   │           │
│        └────────────┴────────────┴────────────┴────────────┘           │
└────────────────┬────────────────────┬────────────────┬──────────────────┘
                 │                    │                │
        ┌────────▼────────┐  ┌───────▼───────┐  ┌─────▼─────┐
        │   Prometheus    │  │     Loki      │  │   Jaeger  │
        │    (9090)       │  │    (3100)     │  │  (16686)  │
        │                 │  │               │  │           │
        │  - Metrics      │  │  - Logs       │  │  - Traces │
        │  - Alerts       │  │  - Labels     │  │  - Spans  │
        │  - Rules        │  │  - LogQL      │  │  - Badger │
        └────────┬────────┘  └───────┬───────┘  └─────┬─────┘
                 │                   │                │
                 │           ┌───────▼───────┐        │
                 │           │   Promtail    │        │
                 │           │ (Log shipper) │        │
                 │           └───────┬───────┘        │
                 │                   │                │
        ┌────────▼───────────────────▼────────────────▼────────┐
        │                    Go Services                        │
        │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
        │  │ API Gateway │  │ Auth Svc    │  │ Trip Svc    │   │
        │  │ /metrics    │  │ /metrics    │  │ /metrics    │   │
        │  │ stdout logs │  │ stdout logs │  │ stdout logs │   │
        │  │ OTLP traces │  │ OTLP traces │  │ OTLP traces │   │
        │  └─────────────┘  └─────────────┘  └─────────────┘   │
        └──────────────────────────────────────────────────────┘
```

## SLOs/SLIs được định nghĩa

### API Gateway
| SLI | SLO | Error Budget |
|-----|-----|--------------|
| Availability (non-5xx) | 99.9% | 43.2 min/month |
| P95 Latency | < 500ms | - |
| P99 Latency | < 1s | - |

### Trip Service (Business Critical)
| SLI | SLO | Error Budget |
|-----|-----|--------------|
| Trip Booking Success Rate | 99.9% | 43.2 min/month |
| Driver Search P95 Latency | < 200ms | - |

### Authentication Service
| SLI | SLO | Error Budget |
|-----|-----|--------------|
| Auth Success Rate | 99.95% | 21.6 min/month |
| P95 Latency | < 100ms | - |

## Cách sử dụng

### 1. Khởi động stack

```bash
# Từ root directory
docker compose up -d

# Kiểm tra tất cả services healthy
docker compose ps
```

### 2. Truy cập các tools

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger UI**: http://localhost:16686
- **Loki**: http://localhost:3100 (thường truy cập qua Grafana)

### 3. Demo flow Observability

#### a) Tạo traffic
```bash
# Login để tạo traces
curl -X POST http://localhost:8080/grpc/auth \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Dùng token để tạo trip (sẽ tạo nhiều traces)
TOKEN="<access_token từ response trên>"
curl -X POST http://localhost:8080/trip \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"origin_lat":10.762622,"origin_lng":106.660172,"dest_lat":10.823099,"dest_lng":106.629664,"payment_method":"cash"}'
```

#### b) Xem traces trong Jaeger
1. Mở http://localhost:16686
2. Chọn Service: `api-gateway` hoặc `trip-service`
3. Click "Find Traces"
4. Click vào một trace để xem chi tiết spans

#### c) Correlate logs với traces
1. Trong Jaeger, copy `traceID` từ một trace
2. Mở Grafana → Explore → chọn Loki
3. Query: `{service="api-gateway"} |= "<traceID>"`
4. Hoặc click "View Logs" từ trace (nếu được cấu hình)

#### d) Xem SLO dashboard
1. Mở Grafana → Dashboards → UIT-Go
2. Chọn "UIT-Go SLO/SLI Dashboard"
3. Xem Error Budget, latency percentiles, request rates

## Cấu hình chi tiết

### Prometheus
- **File**: `prometheus.yml`
- **Alerts**: `prometheus-alerts.yml`
- **Retention**: 15 ngày
- **Scrape interval**: 15s

### Loki
- **File**: `loki-config.yml`
- **Retention**: 7 ngày (168h)
- **Storage**: Local filesystem (`/loki`)

### Jaeger
- **Storage**: Badger (embedded key-value store)
- **Retention**: 7 ngày
- **Protocol**: OTLP gRPC (port 4317)

### Promtail
- **File**: `promtail-config.yml`
- **Source**: Docker socket (đọc logs từ containers)
- **Labels extracted**: service, level, trace_id

## Alert Rules

Các alerts được định nghĩa trong `prometheus-alerts.yml`:

| Alert | Severity | Condition |
|-------|----------|-----------|
| SLO_HighErrorRate_FastBurn | critical | Error rate > 14.4x SLO trong 1h |
| SLO_HighErrorRate_SlowBurn | warning | Error rate > 3x SLO trong 6h |
| SLO_ErrorBudgetLow | critical | Error budget < 20% |
| SLO_HighLatency_APIGateway | warning | P95 > 500ms trong 5m |
| SLO_HighLatency_DriverSearch | warning | P95 > 200ms trong 5m |
| ServiceDown | critical | Service unreachable > 1m |
| HighTripBookingFailureRate | critical | Trip booking errors > 1% |

## Troubleshooting

### Loki không nhận logs
```bash
# Kiểm tra Promtail
docker logs uit-go-promtail-1

# Kiểm tra Docker socket mount
docker exec uit-go-promtail-1 ls -la /var/run/docker.sock
```

### Jaeger không hiển thị traces
```bash
# Kiểm tra services có gửi traces không
docker logs uit-go-api-gateway-1 | grep -i trace

# Kiểm tra Jaeger collector
docker logs uit-go-jaeger-1
```

### Grafana datasource lỗi
```bash
# Test Prometheus
curl http://localhost:9090/api/v1/query?query=up

# Test Loki
curl http://localhost:3100/ready

# Test Jaeger
curl http://localhost:16686/api/services
```

## Trade-offs & Decisions

### Tại sao chọn Loki thay vì ELK Stack?
- **Pro**: Nhẹ hơn nhiều, không cần index full-text, tích hợp Grafana native
- **Con**: Query chậm hơn cho full-text search, không mạnh bằng Elasticsearch cho log analytics

### Tại sao dùng Badger storage cho Jaeger?
- **Pro**: Zero-config, persistent, embedded, không cần Elasticsearch/Cassandra
- **Con**: Không scale được như Cassandra, giới hạn storage trên single node

### Tại sao structured logging với JSON?
- **Pro**: Machine-readable, dễ parse bởi Loki/Promtail, hỗ trợ correlation
- **Con**: Không human-friendly khi đọc raw logs

## Files trong thư mục này

```
infrastructure/observability/
├── docker-compose.observability.yml  # Standalone observability stack
├── prometheus.yml                     # Prometheus config
├── prometheus-alerts.yml              # Alert rules
├── loki-config.yml                    # Loki config
├── promtail-config.yml                # Promtail config
├── config-badger.yml                  # Jaeger BADGER config (reference)
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml        # Auto-configure datasources
│       └── dashboards/
│           ├── dashboards.yml         # Dashboard provider config
│           └── json/
│               └── slo-dashboard.json # SLO/SLI dashboard
└── README.md                          # This file
```

## Tài liệu tham khảo

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Prometheus Alerting Rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [Google SRE Book - SLOs](https://sre.google/sre-book/service-level-objectives/)
