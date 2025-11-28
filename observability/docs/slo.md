# Module D: Thiết kế cho Observability - SLO/SLI Documentation

## Tổng quan

Document này định nghĩa Service Level Objectives (SLOs) và Service Level Indicators (SLIs) cho hệ thống UIT-Go ride-hailing platform.

## Định nghĩa thuật ngữ

| Thuật ngữ | Định nghĩa |
|-----------|------------|
| **SLI** (Service Level Indicator) | Metric đo lường chất lượng service cụ thể |
| **SLO** (Service Level Objective) | Mục tiêu cần đạt được cho SLI, có khoảng thời gian |
| **SLA** (Service Level Agreement) | Cam kết với khách hàng, có hậu quả nếu vi phạm |
| **Error Budget** | Lượng lỗi được phép trong khoảng thời gian SLO |

---

## Service Level Objectives

### 1. API Gateway

#### SLO-001: Availability
```yaml
Service: api-gateway
SLI: Success Rate = (requests with status < 500) / (total requests)
Target: 99.9%
Window: 30 days rolling
Error Budget: 0.1% = 43.2 minutes/month
```

**PromQL:**
```promql
# SLI
sum(rate(http_requests_total{service="api-gateway", status!~"5.."}[30d]))
/ sum(rate(http_requests_total{service="api-gateway"}[30d]))

# Error Budget Remaining
1 - (
  sum(increase(http_requests_total{service="api-gateway", status=~"5.."}[30d]))
  / (sum(increase(http_requests_total{service="api-gateway"}[30d])) * 0.001)
)
```

#### SLO-002: Latency
```yaml
Service: api-gateway
SLI: P95 Response Time
Target: < 500ms
Window: 30 days rolling
```

**PromQL:**
```promql
histogram_quantile(0.95, 
  sum(rate(http_request_duration_seconds_bucket{service="api-gateway"}[5m])) by (le)
) < 0.5
```

---

### 2. Authentication Service

#### SLO-003: Authentication Latency
```yaml
Service: authentication-service
SLI: P95 Response Time for login/register
Target: < 100ms
Window: 30 days rolling
Rationale: Auth phải nhanh vì blocking user flow
```

**PromQL:**
```promql
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket{service="authentication"}[5m])) by (le)
) < 0.1
```

#### SLO-004: Token Validation
```yaml
Service: authentication-service
SLI: Token validation success rate
Target: 99.99%
Window: 30 days rolling
Rationale: Invalid token validation = security risk
```

---

### 3. Trip Service

#### SLO-005: Trip Booking Success
```yaml
Service: trip-service
SLI: Successful trip creations / Total trip creation attempts
Target: 99.9%
Window: 30 days rolling
Rationale: Core business operation
```

**PromQL:**
```promql
# SLI (requires custom metric)
sum(rate(trip_bookings_total{status="success"}[30d]))
/ sum(rate(trip_bookings_total[30d]))
```

#### SLO-006: Trip Matching Latency
```yaml
Service: trip-service + location-service
SLI: Time from trip request to driver match (P95)
Target: < 30 seconds
Window: 30 days rolling
Rationale: User experience critical
```

---

### 4. Location Service

#### SLO-007: Driver Search Latency
```yaml
Service: location-service
SLI: P95 time to find nearby drivers
Target: < 200ms
Window: 30 days rolling
Rationale: Affects trip matching speed
```

**PromQL:**
```promql
histogram_quantile(0.95,
  sum(rate(driver_search_duration_seconds_bucket[5m])) by (le)
) < 0.2
```

#### SLO-008: Location Update Processing
```yaml
Service: location-service
SLI: Success rate of location updates
Target: 99.9%
Window: 30 days rolling
Rationale: Driver location accuracy
```

---

## Error Budget Policy

### Budget Calculation
```
Monthly Error Budget = SLO Target Gap × Total Requests

Example for 99.9% availability:
- Error Budget = 0.1% of requests
- If 1,000,000 requests/month → 1,000 errors allowed
- If 43,200 seconds/month → ~43 seconds downtime allowed
```

### Budget Consumption Rate

| Burn Rate | Description | Alert |
|-----------|-------------|-------|
| 14.4x | Exhausts 30-day budget in 2 days | Page immediately |
| 6x | Exhausts 30-day budget in 5 days | Page immediately |
| 3x | Exhausts 30-day budget in 10 days | Ticket |
| 1x | Exhausts 30-day budget in 30 days | Monitor |

### Error Budget Actions

```yaml
Budget > 50%:
  - Normal feature development
  - Can take risks with deployments

Budget 25-50%:
  - Cautious deployments
  - Focus on reliability improvements

Budget < 25%:
  - Feature freeze consideration
  - All hands on reliability
  - No risky deployments

Budget < 10%:
  - Incident mode
  - Only bug fixes deployed
  - All efforts on recovery
```

---

## Multi-Window Multi-Burn-Rate Alerts

### Alert Configuration

Sử dụng approach từ Google SRE Book Chapter 5:

```yaml
# Fast burn - short window
- alert: SLO_HighErrorRate_FastBurn
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[1h]))
      / sum(rate(http_requests_total[1h]))
    ) > 14.4 * 0.001  # 14.4x burn rate, 0.1% error budget
    AND
    (
      sum(rate(http_requests_total{status=~"5.."}[5m]))
      / sum(rate(http_requests_total[5m]))
    ) > 14.4 * 0.001
  for: 2m
  severity: critical
  action: page

# Slow burn - long window  
- alert: SLO_HighErrorRate_SlowBurn
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[6h]))
      / sum(rate(http_requests_total[6h]))
    ) > 3 * 0.001  # 3x burn rate
    AND
    (
      sum(rate(http_requests_total{status=~"5.."}[30m]))
      / sum(rate(http_requests_total[30m]))
    ) > 3 * 0.001
  for: 15m
  severity: warning
  action: ticket
```

### Tại sao Multi-Window?

```
Single Window Problem:
- Window quá ngắn → false positives (noise)
- Window quá dài → chậm phát hiện vấn đề

Multi-Window Solution:
- Long window: Đảm bảo có significant impact
- Short window: Đảm bảo vấn đề đang xảy ra (not historical)
```

---

## Dashboard Design

### SLO Overview Panel
```
┌─────────────────────────────────────────────────────────────┐
│                    SLO Status Overview                       │
├───────────────┬──────────┬───────────┬─────────────────────┤
│ Service       │ Current  │ Target    │ Error Budget        │
├───────────────┼──────────┼───────────┼─────────────────────┤
│ API Gateway   │ 99.95%   │ 99.9%     │ ████████░░ 80%     │
│ Auth Service  │ 99.99%   │ 99.9%     │ █████████░ 95%     │
│ Trip Service  │ 99.85%   │ 99.9%     │ ██░░░░░░░░ 20% ⚠️  │
│ Location Svc  │ 99.92%   │ 99.9%     │ ███████░░░ 70%     │
└───────────────┴──────────┴───────────┴─────────────────────┘
```

### Error Budget Burn-Down
```
Error Budget Remaining (30-day)
100% │█
     │██
 75% │███
     │████
 50% │█████         ← Current
     │██████
 25% │███████       ← Warning threshold
     │████████
  0% │█████████████████████████████████
    Day 1                          Day 30
```

---

## Implementation Checklist

### Metrics Required

- [ ] `http_requests_total` với labels: service, status, method, path
- [ ] `http_request_duration_seconds` histogram với le buckets
- [ ] `trip_bookings_total` với labels: status (success/failure)
- [ ] `driver_search_duration_seconds` histogram
- [ ] `grpc_server_handled_total` với labels: grpc_code

### Alerting Rules

- [ ] SLO_HighErrorRate_FastBurn
- [ ] SLO_HighErrorRate_SlowBurn
- [ ] SLO_ErrorBudgetLow
- [ ] SLO_HighLatency_P95
- [ ] ServiceDown

### Runbooks

- [ ] [High Error Rate](../runbooks/high-error-rate.md)
- [ ] [High Latency](../runbooks/high-latency.md)
- [ ] [Service Down](../runbooks/service-down.md)

### Dashboards

- [ ] SLO Overview Dashboard
- [ ] Error Budget Dashboard
- [ ] Service Detail Dashboard

---

## References

1. [Google SRE Book - Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
2. [Google SRE Workbook - Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/)
3. [Prometheus - Recording Rules for SLOs](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)
4. [Grafana - SLO Dashboards](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/best-practices/)
