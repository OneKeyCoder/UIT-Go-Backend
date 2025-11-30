# ADR-003: Production Deployment Options on Azure Container Apps

**Status**: Proposed  
**Date**: 2025-11-29  
**Deciders**: UIT-Go Team  
**Module**: D - Observability + Infrastructure

---

## Context

Team đang cân nhắc các options để deploy hệ thống UIT-Go lên **Azure Container Apps (ACA)**.

**Scope**: Chỉ focus vào **Production deployment**, không bàn về local development.

**Core Question**: Khi dùng ACA, observability stack nên được triển khai như thế nào?

1. **Dùng Azure Managed Services** (Application Insights + Log Analytics)
2. **Self-host observability trên ACA** (Loki, Prometheus, Grafana, Jaeger chạy như containers, nhận data qua OTLP)
3. **Hybrid với third-party SaaS** (Grafana Cloud)

---

## Data Flow Architecture với OTLP

Tất cả các options đều sử dụng **OTLP (OpenTelemetry Protocol)** để gửi telemetry data từ services:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Services                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │ api-gw  │ │  auth   │ │  trip   │ │ logger  │ │ location│          │
│  │         │ │         │ │         │ │         │ │         │          │
│  │ ┌─────┐ │ │ ┌─────┐ │ │ ┌─────┐ │ │ ┌─────┐ │ │ ┌─────┐ │          │
│  │ │OTLP │ │ │ │OTLP │ │ │ │OTLP │ │ │ │OTLP │ │ │ │OTLP │ │          │
│  │ │SDK  │ │ │ │SDK  │ │ │ │SDK  │ │ │ │SDK  │ │ │ │SDK  │ │          │
│  │ └──┬──┘ │ │ └──┬──┘ │ │ └──┬──┘ │ │ └──┬──┘ │ │ └──┬──┘ │          │
│  └────┼────┘ └────┼────┘ └────┼────┘ └────┼────┘ └────┼────┘          │
│       │           │           │           │           │                │
│       └───────────┴───────────┴─────┬─────┴───────────┘                │
│                                     │                                   │
│                              OTLP (gRPC/HTTP)                          │
│                                     │                                   │
│                                     ▼                                   │
│                    ┌────────────────────────────────┐                  │
│                    │      OTLP Collector            │                  │
│                    │   (OpenTelemetry Collector)    │                  │
│                    └────────────────┬───────────────┘                  │
│                                     │                                   │
└─────────────────────────────────────┼───────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ▼                 ▼                 ▼
            ┌───────────┐     ┌───────────┐     ┌───────────┐
            │   Logs    │     │  Metrics  │     │  Traces   │
            │  Backend  │     │  Backend  │     │  Backend  │
            └───────────┘     └───────────┘     └───────────┘
```

**Backends có thể là**:

-   **Azure**: Application Insights, Log Analytics, Azure Monitor
-   **Self-hosted trên ACA**: Loki, Prometheus, Jaeger (containers)
-   **SaaS**: Grafana Cloud, Datadog, New Relic

---

## Option A: ACA + Azure Native Services (Managed)

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Azure Container Apps Environment                          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Application Services (ACA)                       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │   │
│  │  │ api-gw  │ │  auth   │ │  trip   │ │ logger  │ │ location│      │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘      │   │
│  │       └───────────┴───────────┴─────┬─────┴───────────┘            │   │
│  └─────────────────────────────────────┼───────────────────────────────┘   │
│                                        │                                    │
│                                   OTLP/HTTP                                │
│                                        │                                    │
└────────────────────────────────────────┼────────────────────────────────────┘
                                         │
                                         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         Azure Managed Services                              │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                    Application Insights                               │ │
│  │                                                                       │ │
│  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐               │ │
│  │   │   Traces    │   │    Logs     │   │   Metrics   │               │ │
│  │   │ (Distributed│   │ (via OTLP)  │   │(Performance)│               │ │
│  │   │  Tracing)   │   │             │   │             │               │ │
│  │   └─────────────┘   └─────────────┘   └─────────────┘               │ │
│  │                                                                       │ │
│  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐               │ │
│  │   │ Application │   │   Smart     │   │   Alerts    │               │ │
│  │   │    Map      │   │  Detection  │   │ (AI-based)  │               │ │
│  │   └─────────────┘   └─────────────┘   └─────────────┘               │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                    Log Analytics Workspace                            │ │
│  │                                                                       │ │
│  │   • KQL Queries (Kusto Query Language)                               │ │
│  │   • 30-day default retention (configurable)                          │ │
│  │   • Container stdout/stderr logs (auto-collected by ACA)             │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                      Azure Monitor                                    │ │
│  │                                                                       │ │
│  │   • Container metrics (CPU, Memory, Network)                         │ │
│  │   • Alert rules & Action Groups                                      │ │
│  │   • Azure Workbooks (Dashboards)                                     │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

### Đặc điểm

| Aspect             | Details                                              |
| ------------------ | ---------------------------------------------------- |
| **Logs**           | Log Analytics Workspace, query bằng KQL              |
| **Metrics**        | Azure Monitor Metrics                                |
| **Traces**         | Application Insights Distributed Tracing             |
| **Dashboards**     | Azure Workbooks (hoặc Grafana với Azure data source) |
| **Alerts**         | Azure Monitor Alerts + Action Groups                 |
| **Query Language** | KQL (Kusto) - KHÔNG phải PromQL/LogQL                |

### Pros ✅

| Pro                        | Explanation                                       |
| -------------------------- | ------------------------------------------------- |
| **Zero ops overhead**      | Azure quản lý toàn bộ infrastructure              |
| **Auto-scaling**           | Không cần lo storage, compute cho observability   |
| **Built-in AI**            | Smart Detection tự động phát hiện anomalies       |
| **Deep Azure integration** | Native support cho ACA, Functions, etc.           |
| **Security**               | RBAC, Private endpoints, data encryption built-in |
| **Compliance**             | SOC2, ISO 27001, HIPAA ready                      |

### Cons ❌

| Con                       | Explanation                                               |
| ------------------------- | --------------------------------------------------------- |
| **Vendor lock-in CAO**    | KQL ≠ PromQL/LogQL, không portable                        |
| **Query language khác**   | Phải học KQL, dashboards không reuse được                 |
| **Cost unpredictable**    | Tính theo GB ingested, có thể spike                       |
| **Limited customization** | Không thể extend như open-source                          |
| **Data sovereignty**      | Data ở Azure region, có thể là issue                      |
| **No Grafana native**     | Phải dùng Azure Workbooks hoặc connect Grafana qua plugin |

### Những gì MẤT khi dùng Azure Native

```
┌─────────────────────────────────────────────────────────────────────┐
│                    WHAT YOU LOSE                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. PromQL Queries (Prometheus)                                     │
│     ────────────────────────────                                    │
│     Local:  rate(http_requests_total[5m])                           │
│     Azure:  requests | summarize count() by bin(timestamp, 5m)      │
│     → Phải viết lại TẤT CẢ queries                                  │
│                                                                      │
│  2. LogQL Queries (Loki)                                            │
│     ────────────────────────────                                    │
│     Local:  {service="api-gateway"} |= "error"                      │
│     Azure:  ContainerAppConsoleLogs                                 │
│             | where ContainerAppName == "api-gateway"               │
│             | where Log contains "error"                            │
│     → Phải viết lại TẤT CẢ queries                                  │
│                                                                      │
│  3. Grafana Dashboards                                              │
│     ────────────────────────────                                    │
│     → Không import được, phải tạo lại trong Azure Workbooks         │
│     → Hoặc dùng Grafana + Azure Monitor data source (thêm cost)     │
│                                                                      │
│  4. Alert Rules                                                      │
│     ────────────────────────────                                    │
│     → Prometheus alerting rules không dùng được                     │
│     → Phải tạo lại trong Azure Monitor Alerts                       │
│                                                                      │
│  5. Community Dashboards                                             │
│     ────────────────────────────                                    │
│     → Hàng ngàn dashboards trên grafana.com không dùng được         │
│                                                                      │
│  6. Portability                                                      │
│     ────────────────────────────                                    │
│     → Nếu muốn chuyển sang AWS/GCP, phải làm lại từ đầu            │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Cost Estimate

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Cost Breakdown (Low Traffic)                       │
│                    ~1000 requests/day, 5 services                     │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Application Services (ACA):                                         │
│  ├─ 5 services × $0.000016/vCPU-second                              │
│  │   (with scale-to-zero when no traffic)                           │
│  └─ Total: ~$20/month                                                │
│                                                                       │
│  Application Insights:                                                │
│  ├─ First 5GB/month: FREE                                            │
│  ├─ Additional: $2.30/GB ingested                                    │
│  ├─ Estimate: ~2GB/month logs + traces                               │
│  └─ Total: ~$0/month (within free tier)                              │
│                                                                       │
│  Log Analytics Workspace:                                             │
│  ├─ First 5GB/month: FREE                                            │
│  ├─ Additional: $2.76/GB ingested                                    │
│  ├─ Retention: First 31 days FREE                                    │
│  └─ Total: ~$0/month (within free tier)                              │
│                                                                       │
│  Managed Databases:                                                   │
│  ├─ PostgreSQL Flexible (B1ms): ~$13/month                          │
│  ├─ CosmosDB Serverless: ~$5/month                                   │
│  └─ Azure Cache for Redis (Basic): ~$16/month                        │
│                                                                       │
├──────────────────────────────────────────────────────────────────────┤
│  TOTAL (Low Traffic): ~$54/month                                     │
│                                                                       │
│  ⚠️  WARNING: Cost scales with traffic!                              │
│  High traffic (100k req/day) could be:                               │
│  ├─ ACA: ~$50/month (more compute time)                              │
│  ├─ App Insights: ~$23/month (10GB logs)                             │
│  └─ TOTAL: ~$100+/month                                              │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Option B: ACA + Self-hosted Observability Stack (Containers)

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Azure Container Apps Environment                          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Application Services (ACA)                       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │   │
│  │  │ api-gw  │ │  auth   │ │  trip   │ │ logger  │ │ location│      │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘      │   │
│  │       └───────────┴───────────┴─────┬─────┴───────────┘            │   │
│  └─────────────────────────────────────┼───────────────────────────────┘   │
│                                        │                                    │
│                                   OTLP/gRPC                                │
│                                        │                                    │
│  ┌─────────────────────────────────────┼───────────────────────────────┐   │
│  │                    OTLP Collector (ACA Container)                    │   │
│  │               ┌─────────────────────┴─────────────────────┐         │   │
│  │               │      OpenTelemetry Collector              │         │   │
│  │               │                                           │         │   │
│  │               │   receivers:                              │         │   │
│  │               │     otlp: (grpc:4317, http:4318)         │         │   │
│  │               │                                           │         │   │
│  │               │   exporters:                              │         │   │
│  │               │     loki: → Loki container                │         │   │
│  │               │     prometheus: → Prometheus container    │         │   │
│  │               │     otlp/jaeger: → Jaeger container       │         │   │
│  │               │                                           │         │   │
│  │               └───────────────────────────────────────────┘         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Observability Stack (ACA Containers)              │   │
│  │                                                                       │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐  │   │
│  │  │    Loki     │  │ Prometheus  │  │   Jaeger    │  │  Grafana  │  │   │
│  │  │   (Logs)    │  │  (Metrics)  │  │  (Traces)   │  │ (UI/Dash) │  │   │
│  │  │             │  │             │  │             │  │           │  │   │
│  │  │ Port: 3100  │  │ Port: 9090  │  │ Port: 16686 │  │ Port: 3000│  │   │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └─────┬─────┘  │   │
│  │         │                │                │               │         │   │
│  │         └────────────────┴────────────────┴───────────────┘         │   │
│  │                                   │                                  │   │
│  │                        Azure Files / Azure Disk                      │   │
│  │                         (Persistent Storage)                         │   │
│  │                                                                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### OTLP Collector Config Example

```yaml
# otel-collector-config.yaml
receivers:
    otlp:
        protocols:
            grpc:
                endpoint: 0.0.0.0:4317
            http:
                endpoint: 0.0.0.0:4318

processors:
    batch:
        timeout: 1s
        send_batch_size: 1024

exporters:
    # Logs → Loki
    loki:
        endpoint: http://loki:3100/loki/api/v1/push
        labels:
            attributes:
                service.name: "service"
                service.namespace: "namespace"

    # Metrics → Prometheus (via remote write)
    prometheusremotewrite:
        endpoint: http://prometheus:9090/api/v1/write

    # Traces → Jaeger
    otlp/jaeger:
        endpoint: jaeger:4317
        tls:
            insecure: true

service:
    pipelines:
        logs:
            receivers: [otlp]
            processors: [batch]
            exporters: [loki]
        metrics:
            receivers: [otlp]
            processors: [batch]
            exporters: [prometheusremotewrite]
        traces:
            receivers: [otlp]
            processors: [batch]
            exporters: [otlp/jaeger]
```

### Đặc điểm

| Aspect             | Details                         |
| ------------------ | ------------------------------- |
| **Logs**           | Loki (LogQL)                    |
| **Metrics**        | Prometheus (PromQL)             |
| **Traces**         | Jaeger                          |
| **Dashboards**     | Grafana (reuse từ local!)       |
| **Alerts**         | Prometheus Alertmanager         |
| **Query Language** | PromQL + LogQL (SAME as local!) |

### Pros ✅

| Pro                      | Explanation                             |
| ------------------------ | --------------------------------------- |
| **100% portable**        | Có thể move sang AWS/GCP/on-prem        |
| **Same as local dev**    | PromQL, LogQL, Grafana dashboards reuse |
| **Community dashboards** | Thousands of ready-made dashboards      |
| **Full control**         | Customize retention, sampling, etc.     |
| **Predictable cost**     | Fixed cost cho containers               |
| **No vendor lock-in**    | CNCF open-source stack                  |

### Cons ❌

| Con                         | Explanation                                |
| --------------------------- | ------------------------------------------ |
| **Ops overhead CAO**        | Phải manage 4-5 observability containers   |
| **Storage management**      | Phải configure Azure Files/Disk            |
| **Scaling manual**          | Loki, Prometheus không auto-scale          |
| **Higher base cost**        | Containers chạy 24/7 (không scale-to-zero) |
| **Security responsibility** | Phải tự configure TLS, auth                |
| **Updates manual**          | Phải tự update Loki, Prometheus versions   |

### Những gì ĐƯỢC khi self-host

```
┌─────────────────────────────────────────────────────────────────────┐
│                    WHAT YOU GAIN                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. Same Queries Everywhere                                         │
│     ────────────────────────────                                    │
│     Local:  rate(http_requests_total[5m])                           │
│     Prod:   rate(http_requests_total[5m])    ← SAME!                │
│                                                                      │
│  2. Dashboard Portability                                            │
│     ────────────────────────────                                    │
│     → Export JSON từ local Grafana                                   │
│     → Import vào production Grafana                                  │
│     → Done!                                                          │
│                                                                      │
│  3. Alert Rules Portability                                          │
│     ────────────────────────────                                    │
│     → prometheus-alerts.yml works everywhere                         │
│                                                                      │
│  4. Multi-cloud Ready                                                │
│     ────────────────────────────                                    │
│     → Có thể deploy lên AWS ECS, GCP Cloud Run                      │
│     → Observability stack y hệt                                      │
│                                                                      │
│  5. Community Resources                                              │
│     ────────────────────────────                                    │
│     → Thousands of Grafana dashboards                                │
│     → PromQL/LogQL tutorials everywhere                              │
│     → Large community support                                        │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Cost Estimate

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Cost Breakdown (Self-hosted on ACA)                │
│                    ~1000 requests/day, 5 services                     │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Application Services (ACA):                                         │
│  └─ Same as Option A: ~$20/month                                     │
│                                                                       │
│  Observability Stack (ACA - ALWAYS ON, no scale-to-zero):           │
│  ├─ OTLP Collector: 0.25 vCPU, 0.5GB RAM                            │
│  │   = ~$10/month                                                    │
│  │                                                                   │
│  ├─ Loki: 0.5 vCPU, 1GB RAM                                         │
│  │   = ~$20/month                                                    │
│  │                                                                   │
│  ├─ Prometheus: 0.5 vCPU, 1GB RAM                                   │
│  │   = ~$20/month                                                    │
│  │                                                                   │
│  ├─ Jaeger: 0.25 vCPU, 0.5GB RAM                                    │
│  │   = ~$10/month                                                    │
│  │                                                                   │
│  ├─ Grafana: 0.25 vCPU, 0.5GB RAM                                   │
│  │   = ~$10/month                                                    │
│  │                                                                   │
│  └─ Observability Total: ~$70/month                                  │
│                                                                       │
│  Storage (Azure Files):                                              │
│  ├─ Loki data: ~20GB = ~$2/month                                    │
│  ├─ Prometheus data: ~10GB = ~$1/month                              │
│  └─ Storage Total: ~$3/month                                         │
│                                                                       │
│  Managed Databases:                                                   │
│  └─ Same as Option A: ~$34/month                                     │
│                                                                       │
├──────────────────────────────────────────────────────────────────────┤
│  TOTAL: ~$127/month                                                  │
│                                                                       │
│  ⚠️  Cost is FIXED regardless of traffic                            │
│  (observability stack runs 24/7)                                     │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Option C: ACA + Azure Managed Grafana

### Tại sao có option này?

Azure và AWS đều có **Managed Grafana** service. Câu hỏi đặt ra: tại sao dùng third-party (Grafana Cloud) khi cloud provider đã có?

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Azure Container Apps Environment                          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Application Services (ACA)                       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │   │
│  │  │ api-gw  │ │  auth   │ │  trip   │ │ logger  │ │ location│      │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘      │   │
│  │       └───────────┴───────────┴─────┬─────┴───────────┘            │   │
│  └─────────────────────────────────────┼───────────────────────────────┘   │
│                                        │                                    │
│                                   OTLP/HTTP                                │
│                                        │                                    │
└────────────────────────────────────────┼────────────────────────────────────┘
                                         │
                                         ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                    Azure Managed Services                                   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                  Azure Managed Grafana                                │ │
│  │                                                                       │ │
│  │   ┌─────────────────────────────────────────────────────────────┐   │ │
│  │   │                    Grafana UI                                │   │ │
│  │   │                                                              │   │ │
│  │   │  ⚠️ CHỈ CÓ UI - KHÔNG CÓ LOKI/PROMETHEUS/TEMPO             │   │ │
│  │   │                                                              │   │ │
│  │   │  Backend phải dùng:                                         │   │ │
│  │   │  • Azure Monitor (metrics) → KQL, không phải PromQL        │   │ │
│  │   │  • Log Analytics (logs) → KQL, không phải LogQL            │   │ │
│  │   │  • App Insights (traces) → KQL                              │   │ │
│  │   │                                                              │   │ │
│  │   └─────────────────────────────────────────────────────────────┘   │ │
│  │                                                                       │ │
│  │   Cost: ~$108/month (Essential tier, 0.15$/hour × 24 × 30)           │ │
│  │                                                                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                    Log Analytics + Azure Monitor                      │ │
│  │                    (Same as Option A - backend)                       │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

### Vấn đề chính: Azure Managed Grafana ≠ Full Grafana Stack

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   What you THINK you get:          What you ACTUALLY get:                   │
│   ══════════════════════           ══════════════════════                   │
│                                                                              │
│   ┌─────────────────────┐          ┌─────────────────────┐                 │
│   │    Grafana UI       │          │    Grafana UI       │  ✅             │
│   └─────────────────────┘          └─────────────────────┘                 │
│   ┌─────────────────────┐          ┌─────────────────────┐                 │
│   │    Loki (LogQL)     │          │    NOT INCLUDED     │  ❌             │
│   └─────────────────────┘          └─────────────────────┘                 │
│   ┌─────────────────────┐          ┌─────────────────────┐                 │
│   │  Prometheus (PromQL)│          │    NOT INCLUDED     │  ❌             │
│   └─────────────────────┘          └─────────────────────┘                 │
│   ┌─────────────────────┐          ┌─────────────────────┐                 │
│   │   Tempo (TraceQL)   │          │    NOT INCLUDED     │  ❌             │
│   └─────────────────────┘          └─────────────────────┘                 │
│                                                                              │
│   → Must use Azure Log Analytics (KQL) as backend                           │
│   → Must use Azure Monitor (KQL) for metrics                                │
│   → Query language is DIFFERENT from local dev!                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### So sánh với các Managed Grafana khác

| Provider         | Service                | Cost        | Includes Backend?       |
| ---------------- | ---------------------- | ----------- | ----------------------- |
| **Grafana Labs** | Grafana Cloud          | FREE (50GB) | ✅ Loki + Mimir + Tempo |
| **Azure**        | Azure Managed Grafana  | ~$108/mo    | ❌ Only UI              |
| **AWS**          | Amazon Managed Grafana | ~$50-100/mo | ❌ Only UI              |
| **GCP**          | (No managed Grafana)   | N/A         | N/A                     |

### Tại sao Azure/AWS không bundle Loki/Prometheus?

**Business reason**:

-   Azure muốn bạn dùng Log Analytics ($2.76/GB) → revenue
-   AWS muốn bạn dùng CloudWatch → revenue
-   Nếu họ bundle Loki miễn phí → cannibalize own products

```
Azure's incentive:
══════════════════

"Grafana UI? Sure, we'll host it for $108/month.
 But you MUST use our Log Analytics backend.
 That's where we make money: $2.76/GB ingested."

vs

Grafana Labs' incentive:
════════════════════════

"Grafana UI? Free.
 Loki/Mimir/Tempo backend? Also free (50GB/month).
 We make money from Enterprise customers paying $thousands/month."
```

### Cost Estimate

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Cost Breakdown (Azure Managed Grafana)             │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Application Services (ACA): ~$20/month                              │
│                                                                       │
│  Azure Managed Grafana:                                              │
│  ├─ Essential tier: $0.15/hour × 24 × 30 = $108/month               │
│  └─ Standard tier: $0.36/hour × 24 × 30 = $259/month                │
│                                                                       │
│  Backend (same as Option A):                                         │
│  ├─ Log Analytics: ~$0 (free tier)                                  │
│  ├─ Azure Monitor: ~$0 (basic)                                      │
│  └─ App Insights: ~$0 (free tier)                                   │
│                                                                       │
│  Managed Databases: ~$34/month                                       │
│                                                                       │
├──────────────────────────────────────────────────────────────────────┤
│  TOTAL: ~$162/month                                                  │
│                                                                       │
│  ⚠️  You pay $108/month JUST for Grafana UI                          │
│  ⚠️  Backend still uses KQL (not PromQL/LogQL)                       │
│  ⚠️  Dashboards from local dev NOT reusable                          │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### Verdict: Azure Managed Grafana = Worst of Both Worlds

| Aspect              | Azure Managed Grafana | Grafana Cloud         |
| ------------------- | --------------------- | --------------------- |
| **Cost**            | $108/month (just UI)  | $0 (includes backend) |
| **Query Language**  | KQL (different)       | PromQL/LogQL (same)   |
| **Dashboard Reuse** | ❌ No                 | ✅ Yes                |
| **Vendor Lock-in**  | HIGH (Azure backend)  | LOW                   |
| **Value**           | Poor                  | Excellent             |

**Kết luận**: Azure Managed Grafana là option TỆ NHẤT vì:

-   Trả $108/month chỉ cho UI
-   Vẫn phải dùng KQL (không portable)
-   Không có lợi ích gì so với Option A (Azure Native)

---

## Option D: ACA + Grafana Cloud (Third-party SaaS)

### Tại sao Third-party thay vì Cloud-native?

Sau khi phân tích Option C (Azure Managed Grafana), ta thấy:

-   Azure Managed Grafana = **CHỈ CÓ UI** ($108/month)
-   Backend vẫn phải dùng **KQL** (không portable)

**Grafana Cloud** (third-party) khác biệt:

-   Grafana UI + Loki + Mimir + Tempo = **FULL STACK**
-   FREE tier generous
-   **Same query language** (PromQL/LogQL) as local dev

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Azure Container Apps Environment                          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Application Services (ACA)                       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │   │
│  │  │ api-gw  │ │  auth   │ │  trip   │ │ logger  │ │ location│      │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘      │   │
│  │       └───────────┴───────────┴─────┬─────┴───────────┘            │   │
│  └─────────────────────────────────────┼───────────────────────────────┘   │
│                                        │                                    │
│                                   OTLP/HTTP                                │
│                                        │                                    │
│  ┌─────────────────────────────────────┼───────────────────────────────┐   │
│  │               Grafana Alloy (ACA Container)                          │   │
│  │                                     │                                │   │
│  │   ┌─────────────────────────────────┴─────────────────────┐        │   │
│  │   │              Grafana Alloy                            │        │   │
│  │   │                                                       │        │   │
│  │   │   • Receives OTLP from services                       │        │   │
│  │   │   • Batches and compresses data                       │        │   │
│  │   │   • Pushes to Grafana Cloud (HTTPS)                   │        │   │
│  │   │                                                       │        │   │
│  │   └───────────────────────────────────────────────────────┘        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                         │
                                    HTTPS (push)
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Grafana Cloud (SaaS)                                 │
│                                                                              │
│  ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐         │
│  │    Grafana Loki   │ │ Grafana Mimir     │ │   Grafana Tempo   │         │
│  │      (Logs)       │ │    (Metrics)      │ │     (Traces)      │         │
│  │                   │ │                   │ │                   │         │
│  │  LogQL ✓          │ │  PromQL ✓         │ │  TraceQL ✓        │         │
│  └───────────────────┘ └───────────────────┘ └───────────────────┘         │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                         Grafana UI                                    │  │
│  │                                                                       │  │
│  │   • Same dashboards as local                                         │  │
│  │   • Same queries (PromQL, LogQL)                                     │  │
│  │   • Alert rules compatible                                           │  │
│  │   • Public dashboard links for demo                                  │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  Free Tier Limits:                                                          │
│  ├─ Logs: 50GB/month                                                        │
│  ├─ Metrics: 10,000 series                                                  │
│  ├─ Traces: 50GB/month                                                      │
│  └─ Retention: 14 days                                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Đặc điểm

| Aspect             | Details                            |
| ------------------ | ---------------------------------- |
| **Logs**           | Grafana Loki (hosted) - LogQL ✓    |
| **Metrics**        | Grafana Mimir (hosted) - PromQL ✓  |
| **Traces**         | Grafana Tempo (hosted) - TraceQL ✓ |
| **Dashboards**     | Grafana (same as local!)           |
| **Alerts**         | Grafana Alerting                   |
| **Query Language** | PromQL + LogQL (SAME as local!)    |

### Pros ✅

| Pro                            | Explanation                         |
| ------------------------------ | ----------------------------------- |
| **Same as local**              | PromQL, LogQL, Grafana UI identical |
| **Zero ops for observability** | Grafana manages everything          |
| **Free tier generous**         | 50GB logs, 10k metrics series       |
| **Dashboard portability**      | Export/import JSON works            |
| **Low cost**                   | Only pay for Alloy container        |
| **Public dashboards**          | Easy to share for demo/presentation |

### Cons ❌

| Con                        | Explanation                            |
| -------------------------- | -------------------------------------- |
| **Data leaves Azure**      | May have compliance/sovereignty issues |
| **14-day retention**       | Free tier limited, paid for longer     |
| **Rate limits**            | May hit limits during load testing     |
| **Third-party dependency** | Grafana Labs outage = no observability |
| **Need Alloy container**   | Extra container to manage              |

### Cost Estimate

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Cost Breakdown (Grafana Cloud)                     │
│                    ~1000 requests/day, 5 services                     │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Application Services (ACA):                                         │
│  └─ Same as Option A: ~$20/month                                     │
│                                                                       │
│  Grafana Alloy (ACA):                                                │
│  ├─ 0.25 vCPU, 0.5GB RAM                                            │
│  └─ Total: ~$10/month                                                │
│                                                                       │
│  Grafana Cloud:                                                       │
│  └─ FREE tier (within limits)                                        │
│                                                                       │
│  Managed Databases:                                                   │
│  └─ Same as Option A: ~$34/month                                     │
│                                                                       │
├──────────────────────────────────────────────────────────────────────┤
│  TOTAL: ~$64/month                                                   │
│                                                                       │
│  ⚠️  May need paid tier if:                                          │
│  ├─ Logs > 50GB/month                                                │
│  ├─ Metrics > 10k series                                             │
│  └─ Need > 14 days retention                                         │
│                                                                       │
│  Paid tier starts at $29/month for Pro                               │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Decision Matrix

| Criteria              | Weight | A: Azure Native  | B: Self-hosted    | C: Azure Managed Grafana | D: Grafana Cloud  |
| --------------------- | ------ | ---------------- | ----------------- | ------------------------ | ----------------- |
| **Monthly Cost**      | 20%    | ⭐⭐⭐⭐ ~$54    | ⭐⭐ ~$127        | ⭐ ~$162                 | ⭐⭐⭐⭐ ~$64     |
| **Ops Overhead**      | 20%    | ⭐⭐⭐⭐⭐ Zero  | ⭐⭐ High         | ⭐⭐⭐⭐ Low             | ⭐⭐⭐⭐ Low      |
| **Vendor Lock-in**    | 15%    | ⭐ Very High     | ⭐⭐⭐⭐⭐ None   | ⭐ Very High             | ⭐⭐⭐⭐ Low      |
| **Local/Prod Parity** | 15%    | ⭐ KQL≠PromQL    | ⭐⭐⭐⭐⭐ 100%   | ⭐ KQL≠PromQL            | ⭐⭐⭐⭐⭐ 100%   |
| **Dashboard Reuse**   | 10%    | ⭐ Recreate      | ⭐⭐⭐⭐⭐ Export | ⭐ Recreate              | ⭐⭐⭐⭐⭐ Export |
| **Scalability**       | 10%    | ⭐⭐⭐⭐⭐ Auto  | ⭐⭐ Manual       | ⭐⭐⭐⭐⭐ Auto          | ⭐⭐⭐⭐⭐ Auto   |
| **Data Sovereignty**  | 10%    | ⭐⭐⭐⭐⭐ Azure | ⭐⭐⭐⭐⭐ Azure  | ⭐⭐⭐⭐⭐ Azure         | ⭐⭐ External     |
| **Weighted Score**    | 100%   | 3.25             | 3.45              | **2.45**                 | **3.90**          |

### Option C là TỆ NHẤT vì:

-   Trả $108/month **CHỈ CHO UI**
-   Backend vẫn dùng KQL → **không portable**
-   Không có lợi ích gì so với Option A
-   "Worst of both worlds"

---

## Trade-off Summary Table

| Aspect              | A: Azure Native | B: Self-hosted | C: Azure Managed Grafana | D: Grafana Cloud |
| ------------------- | --------------- | -------------- | ------------------------ | ---------------- |
| **Cost**            | ~$54/mo         | ~$127/mo       | ~$162/mo ❌              | ~$64/mo          |
| **Query Language**  | KQL             | PromQL/LogQL   | KQL                      | PromQL/LogQL     |
| **Dashboards**      | Rebuild         | Reuse          | Rebuild                  | Reuse            |
| **Portability**     | ❌ Azure only   | ✅ Any cloud   | ❌ Azure only            | ✅ Any cloud     |
| **Ops Work**        | Zero            | High           | Low                      | Low              |
| **Data Location**   | Azure           | Azure          | Azure                    | External         |
| **Value for Money** | OK              | Good           | **POOR**                 | **Excellent**    |

---

## Cost Comparison Visual

```
Monthly Cost by Option (Production on ACA)
══════════════════════════════════════════════════════════════════════

Option A (Azure Native)      ████████████████████████░░░░░░░░░░  $54/mo
                             ↳ Cheapest, but vendor lock-in HIGH
                             ↳ Must learn KQL, recreate dashboards

Option C (Grafana Cloud)     ██████████████████████████████░░░░  $64/mo
                             ↳ Best balance: same tooling + low ops
                             ↳ Dashboards portable, PromQL/LogQL

Option B (Self-hosted)       ████████████████████████████████████████████████  $127/mo
                             ↳ Full control, highest cost
                             ↳ Must manage 5 observability containers

══════════════════════════════════════════════════════════════════════

Annual Cost:
├─ Option A: $648/year   (cheapest, but lock-in)
├─ Option C: $768/year   (+$120/year for portability)
└─ Option B: $1,524/year (+$876/year for full control)
```

---

## When to Choose What

### Choose Option A (Azure Native) when:

-   ✅ Already committed to Azure ecosystem
-   ✅ Team familiar with KQL
-   ✅ Compliance requires data stay in Azure
-   ✅ Want zero ops overhead
-   ✅ Budget is primary concern
-   ❌ Don't mind recreating dashboards

### Choose Option B (Self-hosted) when:

-   ✅ Need 100% same stack as local
-   ✅ Multi-cloud strategy planned
-   ✅ Have DevOps capacity
-   ✅ Need unlimited retention
-   ✅ Data must stay in Azure
-   ❌ Budget not a concern

### ❌ AVOID Option C (Azure Managed Grafana):

-   Pays $108/month ONLY for UI
-   Backend still uses KQL (not portable)
-   No advantage over Option A
-   **Worst value for money**

### Choose Option D (Grafana Cloud) when:

-   ✅ Want same Grafana experience as local
-   ✅ Don't have DevOps capacity for Option B
-   ✅ 14-day retention acceptable
-   ✅ Data leaving Azure is OK
-   ✅ Free tier sufficient for traffic
-   ✅ Easy demo với public dashboards

---

## Why Third-party (Grafana Cloud) over Cloud-native (Azure Managed Grafana)?

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   Question: Why use Grafana Cloud instead of Azure Managed Grafana?         │
│                                                                              │
│   ════════════════════════════════════════════════════════════════════════  │
│                                                                              │
│   Azure Managed Grafana:                                                    │
│   ─────────────────────                                                     │
│   • Cost: $108/month                                                        │
│   • Includes: ONLY Grafana UI                                               │
│   • Backend: Must use Azure Log Analytics (KQL)                             │
│   • Query Language: KQL (different from local dev)                          │
│   • Dashboards: Must recreate (not portable)                                │
│   • Value: POOR - pay for UI, still locked to Azure backend                 │
│                                                                              │
│   Grafana Cloud (Third-party):                                              │
│   ───────────────────────────                                               │
│   • Cost: FREE (50GB logs, 10k metrics)                                     │
│   • Includes: Grafana + Loki + Mimir + Tempo (FULL STACK)                  │
│   • Backend: PromQL/LogQL (same as local dev)                               │
│   • Query Language: Same as local dev ✅                                    │
│   • Dashboards: Export from local, import to cloud ✅                       │
│   • Value: EXCELLENT - free + portable + same tooling                       │
│                                                                              │
│   ════════════════════════════════════════════════════════════════════════  │
│                                                                              │
│   Business Reality:                                                          │
│   ─────────────────                                                          │
│   Azure/AWS want you to use THEIR backends:                                 │
│   • Azure → Log Analytics ($2.76/GB) → Revenue                              │
│   • AWS → CloudWatch → Revenue                                               │
│                                                                              │
│   Grafana Labs gives away free tier because:                                │
│   • Enterprise customers pay $thousands/month                               │
│   • Free tier = marketing + adoption                                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Recommendation

### For UIT-Go Project (Academic)

**Recommended: Option D (ACA + Grafana Cloud)**

Lý do:

1. **Same tooling as local** - không cần học KQL
2. **Dashboards reusable** - export từ local, import vào cloud
3. **Free tier đủ dùng** - academic project ít data
4. **Easy demo** - public dashboard link cho presentation
5. **Low ops** - chỉ manage 1 container (Alloy)
6. **Best value** - FREE với full stack

### If Compliance Requires Data in Azure

**Fallback: Option A (Azure Native)** - NOT Option C!

-   Option A: ~$54/month, KQL, zero ops
-   Option C: ~$162/month, KQL, still Azure backend

**Option C (Azure Managed Grafana) không có lý do để chọn** vì:

-   Đắt hơn Option A ($108 extra chỉ cho UI)
-   Vẫn dùng KQL như Option A
-   Không có thêm benefit gì

---

## References

-   [Azure Container Apps Pricing](https://azure.microsoft.com/en-us/pricing/details/container-apps/)
-   [Application Insights Pricing](https://azure.microsoft.com/en-us/pricing/details/monitor/)
-   [Log Analytics Pricing](https://azure.microsoft.com/en-us/pricing/details/monitor/)
-   [Grafana Cloud Pricing](https://grafana.com/pricing/)
-   [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
-   [Grafana Alloy](https://grafana.com/docs/alloy/latest/)
