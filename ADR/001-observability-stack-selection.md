# ADR-001: Observability Stack Selection

**Status**: Accepted  
**Date**: 2025-11-29  
**Deciders**: UIT-Go Team  
**Module**: D - Observability  

---

## Context

Hệ thống UIT-Go là một ứng dụng microservices đặt xe với các services: API Gateway, Authentication, Trip, Logger, Location. Chúng ta cần thiết kế hệ thống observability để:

1. Theo dõi health và performance của services
2. Debug và troubleshoot issues
3. Đảm bảo đạt được SLOs đã định nghĩa
4. Cảnh báo khi có sự cố

---

## Decision Drivers

- **Chi phí**: Budget giới hạn (student project)
- **Độ phức tạp vận hành**: Team 4 người, cần đơn giản
- **Khả năng học hỏi**: Hiểu sâu về observability concepts
- **Production-readiness**: Có thể scale lên production
- **Vendor lock-in**: Tránh phụ thuộc quá nhiều vào 1 cloud provider

---

## Considered Options

### Option 1: Self-hosted Open Source Stack (Loki + Prometheus + Grafana + Jaeger)
### Option 2: Azure Native (Application Insights + Azure Monitor)
### Option 3: Grafana Cloud (Managed)
### Option 4: AWS Native (CloudWatch + X-Ray)
### Option 5: Hybrid (Self-hosted local + Cloud production)

---

## Decision Matrix

| Criteria | Weight | Option 1 (Self-hosted) | Option 2 (Azure) | Option 3 (Grafana Cloud) | Option 4 (AWS) | Option 5 (Hybrid) |
|----------|--------|------------------------|------------------|--------------------------|----------------|-------------------|
| **Cost (Local Dev)** | 20% | ⭐⭐⭐⭐⭐ $0 | ⭐⭐ ~$50/mo | ⭐⭐⭐⭐ Free tier | ⭐⭐ ~$50/mo | ⭐⭐⭐⭐⭐ $0 |
| **Cost (Production)** | 15% | ⭐⭐⭐ ~$100/mo | ⭐⭐⭐⭐ Pay-per-use | ⭐⭐⭐⭐ Free tier | ⭐⭐⭐⭐ Pay-per-use | ⭐⭐⭐⭐ Flexible |
| **Setup Complexity** | 15% | ⭐⭐ High | ⭐⭐⭐⭐⭐ Low | ⭐⭐⭐⭐ Low | ⭐⭐⭐⭐ Medium | ⭐⭐⭐ Medium |
| **Learning Value** | 20% | ⭐⭐⭐⭐⭐ Very High | ⭐⭐⭐ Medium | ⭐⭐⭐⭐ High | ⭐⭐⭐ Medium | ⭐⭐⭐⭐⭐ Very High |
| **Vendor Lock-in** | 10% | ⭐⭐⭐⭐⭐ None | ⭐⭐ High | ⭐⭐⭐⭐ Low | ⭐⭐ High | ⭐⭐⭐⭐⭐ None |
| **Feature Completeness** | 10% | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good |
| **Scalability** | 10% | ⭐⭐⭐ Manual | ⭐⭐⭐⭐⭐ Auto | ⭐⭐⭐⭐⭐ Auto | ⭐⭐⭐⭐⭐ Auto | ⭐⭐⭐⭐ Good |
| **Weighted Score** | 100% | **3.85** | 3.55 | **3.90** | 3.45 | **4.00** |

---

## Decision

**Chọn Option 5: Hybrid Approach**

- **Local Development**: Self-hosted (Loki + Prometheus + Grafana + Jaeger)
- **Production (Azure ACA)**: Azure Application Insights hoặc Grafana Cloud

---

## Rationale

### Tại sao chọn Hybrid?

1. **Maximizes Learning**: 
   - Local dev với self-hosted giúp hiểu sâu về cách từng component hoạt động
   - Biết cách config Prometheus scraping, Loki indexing, Grafana dashboards

2. **Cost Effective**:
   - $0 cho local development
   - Production chỉ trả tiền khi cần

3. **No Vendor Lock-in**:
   - OpenTelemetry (OTLP) là standard, có thể switch giữa các backends
   - Code không cần thay đổi khi chuyển từ local → production

4. **Production Ready**:
   - Có thể demo trên Azure với Application Insights
   - Hoặc Grafana Cloud free tier

---

## Trade-offs Analysis

### Logs vs Metrics vs Traces

| Aspect | Logs | Metrics | Traces |
|--------|------|---------|--------|
| **What** | Event records | Numeric measurements | Request journey |
| **When to use** | Debugging, audit | Alerting, dashboards | Performance analysis |
| **Storage cost** | ⭐ High (verbose) | ⭐⭐⭐⭐⭐ Low (aggregated) | ⭐⭐⭐ Medium |
| **Query speed** | ⭐⭐ Slow (full scan) | ⭐⭐⭐⭐⭐ Fast (indexed) | ⭐⭐⭐ Medium |
| **Cardinality** | Unlimited | Limited (label explosion) | Per-request |
| **Retention** | Days-Weeks | Months-Years | Days-Weeks |

### Khi nào dùng gì?

```
┌─────────────────────────────────────────────────────────────────┐
│                    Observability Decision Tree                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  "Hệ thống có đang healthy không?"                              │
│       └─→ METRICS (up/down, error rate, latency percentiles)    │
│                                                                  │
│  "Tại sao request này chậm?"                                    │
│       └─→ TRACES (xem thời gian từng service, từng DB call)     │
│                                                                  │
│  "Chuyện gì đã xảy ra lúc 3:47 AM?"                             │
│       └─→ LOGS (chi tiết events, stack traces)                  │
│                                                                  │
│  "Có bao nhiêu users đang active?"                              │
│       └─→ METRICS (gauges, counters)                            │
│                                                                  │
│  "Request nào gây ra lỗi 500?"                                  │
│       └─→ TRACES + LOGS (correlate bằng trace_id)               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Trade-offs

### 1. Logging: Loki vs Elasticsearch vs CloudWatch Logs

| Criteria | Loki | Elasticsearch | CloudWatch Logs |
|----------|------|---------------|-----------------|
| **Cost** | Free (self-host) | High (RAM hungry) | Pay per GB ingested |
| **Query Language** | LogQL | KQL/Lucene | CloudWatch Insights |
| **Index Strategy** | Labels only | Full-text | Full-text |
| **Storage Efficiency** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐ Poor | ⭐⭐⭐ Good |
| **Query Flexibility** | ⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good |
| **Learning Curve** | Medium | High | Low |

**Decision**: Loki cho local (nhẹ, miễn phí), Application Insights cho Azure production.

**Trade-off accepted**: Loki không có full-text search như Elasticsearch, nhưng đủ dùng cho filtering by labels + grep patterns. Tiết kiệm được 80% storage cost.

### 2. Metrics: Prometheus vs Azure Monitor vs InfluxDB

| Criteria | Prometheus | Azure Monitor | InfluxDB |
|----------|------------|---------------|----------|
| **Cost** | Free | Pay per metric | Free (OSS) / Paid (Cloud) |
| **Pull vs Push** | Pull | Push | Push |
| **Query Language** | PromQL | KQL | InfluxQL/Flux |
| **Ecosystem** | ⭐⭐⭐⭐⭐ Huge | ⭐⭐⭐⭐ Good | ⭐⭐⭐ Medium |
| **Alerting** | AlertManager | Azure Alerts | Kapacitor |
| **Long-term Storage** | Needs Thanos/Cortex | Built-in | Built-in |

**Decision**: Prometheus cho local, Azure Monitor cho production.

**Trade-off accepted**: Prometheus cần thêm Thanos cho long-term storage. Chấp nhận vì local dev không cần retention dài.

### 3. Tracing: Jaeger vs Zipkin vs AWS X-Ray vs Azure App Insights

| Criteria | Jaeger | Zipkin | AWS X-Ray | Azure App Insights |
|----------|--------|--------|-----------|-------------------|
| **Cost** | Free | Free | Pay per trace | Pay per GB |
| **Protocol** | OTLP, Jaeger | Zipkin, OTLP | X-Ray SDK | OTLP, AI SDK |
| **UI Quality** | ⭐⭐⭐⭐ Good | ⭐⭐⭐ Basic | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent |
| **Service Map** | ⭐⭐⭐ Basic | ⭐⭐ Limited | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐⭐ Excellent |
| **Sampling** | Configurable | Configurable | Automatic | Automatic |
| **Vendor Lock-in** | None | None | High (AWS) | High (Azure) |

**Decision**: Jaeger cho local (OTLP compatible), Azure App Insights cho production (all-in-one).

**Trade-off accepted**: Jaeger UI không đẹp bằng App Insights, nhưng OTLP standard cho phép switch backends dễ dàng.

### 4. Dashboarding: Grafana vs Azure Portal vs Datadog

| Criteria | Grafana | Azure Portal | Datadog |
|----------|---------|--------------|---------|
| **Cost** | Free | Free (with Azure) | $$$$ Expensive |
| **Data Sources** | ⭐⭐⭐⭐⭐ 100+ | ⭐⭐⭐ Azure only | ⭐⭐⭐⭐⭐ Many |
| **Customization** | ⭐⭐⭐⭐⭐ Unlimited | ⭐⭐⭐ Limited | ⭐⭐⭐⭐⭐ Excellent |
| **Learning Curve** | Medium | Low | Medium |
| **Alerting** | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent |

**Decision**: Grafana cho cả local và production (có thể connect Azure Monitor).

---

## Cost Analysis

### Local Development (Docker Compose)
```
Loki:        $0
Prometheus:  $0
Grafana:     $0
Jaeger:      $0
─────────────────
Total:       $0/month
```

### Production Option A: Full Self-hosted on Azure AKS
```
AKS (2 nodes B2s):           ~$60/month
Azure Files (storage):       ~$10/month
Public IP:                   ~$5/month
─────────────────────────────────────────
Total:                       ~$75/month

Pros: Full control, no vendor lock-in
Cons: High ops overhead, need to manage scaling
```

### Production Option B: Azure Application Insights
```
App Insights (5GB logs):     ~$12/month
App Insights (traces):       ~$5/month  
Azure Monitor (metrics):     ~$3/month
─────────────────────────────────────────
Total:                       ~$20/month (low traffic)

Pros: Zero ops, auto-scale, great UI
Cons: Vendor lock-in, cost scales with usage
```

### Production Option C: Grafana Cloud Free Tier
```
Logs (50GB):                 $0
Metrics (10k series):        $0
Traces (50GB):               $0
─────────────────────────────────────────
Total:                       $0/month (within limits)

Pros: Free, same Grafana UI as local
Cons: Limited retention (14 days), rate limits
```

**Recommended for this project**: Start with Grafana Cloud free tier, migrate to App Insights if needed.

---

## Implementation

### Local Development Stack
```yaml
# docker-compose.yml
services:
  # Logs
  loki:
    image: grafana/loki:2.9.0
  promtail:
    image: grafana/promtail:2.9.0
  
  # Metrics  
  prometheus:
    image: prom/prometheus:latest
    
  # Traces
  jaeger:
    image: jaegertracing/jaeger:latest
    
  # Dashboard
  grafana:
    image: grafana/grafana:latest
```

### Production (Azure) - Using OTLP
```go
// Same code works for both local and production
// Just change OTEL_EXPORTER_OTLP_ENDPOINT environment variable

// Local:      OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
// Production: OTEL_EXPORTER_OTLP_ENDPOINT=<app-insights-endpoint>
```

---

## Consequences

### Positive
- Deep understanding of observability concepts
- No cost for development
- Flexible production options
- OTLP standard enables backend switching
- Great learning experience for team

### Negative
- More complex local setup (5 containers for observability)
- Need to maintain 2 configurations (local vs production)
- Grafana dashboards need to be recreated for Azure Portal (if using App Insights)

### Risks
- Local stack may consume significant resources (~2GB RAM)
- Production costs may increase with traffic
- Team needs to learn multiple tools

### Mitigations
- Document resource limits in docker-compose
- Set up billing alerts in Azure
- Create runbooks for common operations

---

## Related Decisions

- [ADR-002: SLO/SLI Definitions](./002-slo-sli-definitions.md)
- [ADR-003: Alerting Strategy](./003-alerting-strategy.md)
- [ADR-004: Log Retention Policy](./004-log-retention-policy.md)

---

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Azure Application Insights OTLP Support](https://learn.microsoft.com/en-us/azure/azure-monitor/app/opentelemetry-overview)
- [Google SRE Book - Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
