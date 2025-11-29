# ADR-002: Logs vs Metrics vs Traces - When to Use What

**Status**: Accepted  
**Date**: 2025-11-29  
**Deciders**: UIT-Go Team  
**Module**: D - Observability

---

## Context

Observability có 3 "pillars" chính: Logs, Metrics, và Traces. Mỗi loại có use cases khác nhau và trade-offs riêng. Team cần quyết định khi nào sử dụng loại nào và cách correlate chúng với nhau.

---

## The Three Pillars Explained

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         OBSERVABILITY PILLARS                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐                 │
│  │   METRICS   │      │    LOGS     │      │   TRACES    │                 │
│  │             │      │             │      │             │                 │
│  │  "What is   │      │  "What      │      │  "Where     │                 │
│  │  happening?"│      │  happened?" │      │  did it go?"│                 │
│  │             │      │             │      │             │                 │
│  │  Aggregated │      │  Events     │      │  Request    │                 │
│  │  numbers    │      │  details    │      │  journey    │                 │
│  └─────────────┘      └─────────────┘      └─────────────┘                 │
│        │                    │                    │                         │
│        │                    │                    │                         │
│        ▼                    ▼                    ▼                         │
│   Dashboards           Debug/Audit          Performance                    │
│   Alerting             Troubleshoot         Bottlenecks                    │
│   Capacity             Compliance           Dependencies                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Detailed Comparison

### 1. METRICS

**Definition**: Numeric measurements aggregated over time

**Examples in UIT-Go**:

```promql
# Request rate
rate(http_requests_total{service="api-gateway"}[5m])

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# Latency percentiles
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Active connections
rabbitmq_connections_total
```

**Characteristics**:
| Property | Value |
|----------|-------|
| Data Type | Numbers (counters, gauges, histograms) |
| Cardinality | Low-Medium (limited by labels) |
| Storage | Very efficient (aggregated) |
| Query Speed | Very fast (pre-aggregated) |
| Retention | Long (months/years) |
| Cost | $0.01-0.10 per million data points |

**Best For**:

-   ✅ Alerting ("error rate > 1% for 5 minutes")
-   ✅ Dashboards (real-time visualization)
-   ✅ Capacity planning
-   ✅ SLO tracking
-   ✅ Trend analysis

**Not Good For**:

-   ❌ Debugging specific requests
-   ❌ Understanding "why" something happened
-   ❌ High-cardinality data (user_id, request_id)

**Trade-off**:

> Metrics tell you THAT something is wrong, but not WHY.

---

### 2. LOGS

**Definition**: Timestamped text records of events

**Examples in UIT-Go**:

```json
{
    "ts": "2025-11-29T10:30:00Z",
    "level": "ERROR",
    "service": "authentication-service",
    "msg": "Failed to authenticate user",
    "user_id": "12345",
    "error": "invalid password",
    "trace_id": "abc123",
    "span_id": "def456"
}
```

**Characteristics**:
| Property | Value |
|----------|-------|
| Data Type | Text (structured/unstructured) |
| Cardinality | Unlimited |
| Storage | High (verbose) |
| Query Speed | Slow (scan required) |
| Retention | Short-Medium (days/weeks) |
| Cost | $0.50-2.00 per GB ingested |

**Best For**:

-   ✅ Debugging specific issues
-   ✅ Audit trails
-   ✅ Stack traces
-   ✅ Understanding context
-   ✅ Post-mortem analysis

**Not Good For**:

-   ❌ Real-time alerting (too slow)
-   ❌ Trend analysis (not aggregated)
-   ❌ High-volume monitoring (expensive)

**Trade-off**:

> Logs give you details but are expensive to store and slow to query at scale.

---

### 3. TRACES

**Definition**: Records of request flow through distributed systems

**Example in UIT-Go**:

```
Trace ID: abc123

api-gateway (50ms total)
├── authentication-service (15ms)
│   └── PostgreSQL query (5ms)
├── trip-service (30ms)
│   ├── HERE API call (20ms)
│   └── PostgreSQL query (8ms)
└── logger-service (2ms)
    └── RabbitMQ publish (1ms)
```

**Characteristics**:
| Property | Value |
|----------|-------|
| Data Type | Spans with parent-child relationships |
| Cardinality | Per-request (can be sampled) |
| Storage | Medium-High |
| Query Speed | Medium (indexed by trace_id) |
| Retention | Short (days) |
| Cost | $1-5 per million spans |

**Best For**:

-   ✅ Finding performance bottlenecks
-   ✅ Understanding service dependencies
-   ✅ Debugging slow requests
-   ✅ Identifying which service failed
-   ✅ Latency breakdown

**Not Good For**:

-   ❌ Alerting (need metrics for that)
-   ❌ Log-level details (need logs)
-   ❌ Long-term analysis (short retention)

**Trade-off**:

> Traces show the request journey but require sampling at high scale (can't store every trace).

---

## Decision Matrix: Which to Use When?

| Question                                | Answer        | Tool                        |
| --------------------------------------- | ------------- | --------------------------- |
| "Is the system healthy right now?"      | Metrics       | Prometheus → Grafana        |
| "Why did this specific request fail?"   | Logs + Traces | Loki + Jaeger               |
| "Which service is the bottleneck?"      | Traces        | Jaeger                      |
| "What was the error rate yesterday?"    | Metrics       | Prometheus                  |
| "Show me all errors from user X"        | Logs          | Loki                        |
| "Alert me when error rate > 1%"         | Metrics       | Prometheus AlertManager     |
| "How long did the DB query take?"       | Traces        | Jaeger                      |
| "What requests happened at 3:47 AM?"    | Logs          | Loki                        |
| "Is latency p99 within SLO?"            | Metrics       | Prometheus                  |
| "Why is this request taking 5 seconds?" | Traces → Logs | Jaeger → Loki (by trace_id) |

---

## Correlation Strategy

The power of observability comes from **correlating** all three:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CORRELATION FLOW                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. METRICS alert fires: "Error rate > 1%"                                  │
│           │                                                                  │
│           ▼                                                                  │
│  2. Check METRICS dashboard: Which service? authentication-service          │
│           │                                                                  │
│           ▼                                                                  │
│  3. Query LOGS: {service="authentication-service"} |= "error"              │
│           │     Found: "Database connection timeout"                        │
│           │     trace_id: "abc123"                                          │
│           │                                                                  │
│           ▼                                                                  │
│  4. View TRACE: abc123                                                      │
│           │     Found: PostgreSQL query took 30s (timeout)                  │
│           │                                                                  │
│           ▼                                                                  │
│  5. ROOT CAUSE: Database connection pool exhausted                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Implementation in UIT-Go

```go
// Every log includes trace_id for correlation
logger.InfoCtx(ctx, "Processing request",
    zap.String("trace_id", span.SpanContext().TraceID().String()),
    zap.String("user_id", userID),
)
```

```yaml
# Grafana can link from metrics → traces → logs
# Using trace_id as the correlation key
```

---

## Cost Comparison (Real-world estimates)

### Scenario: 1000 requests/second, 30-day retention

| Pillar                    | Data Volume | Storage Cost | Query Cost | Total/month |
| ------------------------- | ----------- | ------------ | ---------- | ----------- |
| **Metrics**               | ~100MB/day  | ~$3          | ~$1        | **~$4**     |
| **Logs**                  | ~10GB/day   | ~$300        | ~$50       | **~$350**   |
| **Traces** (10% sampling) | ~1GB/day    | ~$30         | ~$10       | **~$40**    |

**Key insight**: Logs are 10x more expensive than metrics and traces combined!

---

## Recommendations for UIT-Go

### 1. Use Metrics For:

-   All SLO tracking (availability, latency, error rate)
-   Alerting (Prometheus AlertManager)
-   Real-time dashboards
-   Capacity planning

### 2. Use Logs For:

-   Error details and stack traces
-   Audit trails (authentication events)
-   Debugging specific issues
-   **Tip**: Use structured logging (JSON) for easier querying

### 3. Use Traces For:

-   Performance debugging
-   Service dependency mapping
-   Finding slow database queries
-   Cross-service request correlation

### 4. Sampling Strategy:

```yaml
# traces: Sample 10% in production, 100% in development
# logs: Log all errors, sample info/debug
# metrics: Always 100% (aggregated anyway)
```

---

## Trade-offs Accepted

| Decision                | Trade-off              | Rationale                                            |
| ----------------------- | ---------------------- | ---------------------------------------------------- |
| Loki over Elasticsearch | No full-text search    | 80% cost savings, labels sufficient for our use case |
| 10% trace sampling      | Miss some traces       | Cost control, can increase for debugging             |
| 7-day log retention     | Can't analyze old logs | Cost control, most debugging happens within days     |
| Structured JSON logs    | Slightly larger size   | Worth it for query flexibility                       |

---

## Summary Table

| Aspect           | Metrics          | Logs              | Traces            |
| ---------------- | ---------------- | ----------------- | ----------------- |
| **Purpose**      | What's happening | What happened     | Where did it go   |
| **Data type**    | Numbers          | Text              | Spans             |
| **Aggregation**  | Pre-aggregated   | Raw events        | Per-request       |
| **Query speed**  | Fast             | Slow              | Medium            |
| **Storage cost** | Low              | High              | Medium            |
| **Retention**    | Long             | Short             | Short             |
| **Alerting**     | ✅ Best          | ❌ Too slow       | ❌ Not suitable   |
| **Debugging**    | ❌ No details    | ✅ Best           | ✅ Good           |
| **Trends**       | ✅ Best          | ❌ Not aggregated | ❌ Not aggregated |

---

## References

-   [Google SRE Book - Practical Alerting](https://sre.google/sre-book/practical-alerting/)
-   [Observability 3 Ways - Cindy Sridharan](https://www.oreilly.com/library/view/distributed-systems-observability/9781492033431/)
-   [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otel/)
