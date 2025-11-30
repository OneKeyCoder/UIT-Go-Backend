# Metrics Collection Strategy

## Context

We need to collect metrics from services to calculate SLIs. Two main approaches exist:

1. **Pull model**: Prometheus scrapes `/metrics` endpoint from each service
2. **Push model**: Services push metrics via OTLP to a collector

## Decision

Use **Pull model** for metrics - Prometheus scrapes `/metrics` from services.

Use **Push model** for traces and logs - Services push via OTLP to Alloy collector.

## Rationale

### Why Pull for Metrics?

For SLO tracking, we need metrics labeled by service, endpoint, and status code. Prometheus pull model provides:

**1. Health detection**

When Prometheus cannot scrape a target, `up == 0`. This immediately indicates service is unreachable. With push model, silence is ambiguous - service could be dead or just have no traffic.

**2. Consistent timing**

Prometheus controls scrape intervals (e.g., every 15 seconds). This provides predictable data points for rate calculations in SLI queries.

**3. Backpressure handling**

If Prometheus is overloaded, it simply scrapes less frequently. With push, services might buffer metrics and risk memory issues.

**4. Debugging**

Operators can directly query a service's `/metrics` endpoint to verify what metrics are exposed. With push, need the collector running to see anything.

### Why not OTLP for Metrics?

OTLP supports metrics, and we could push metrics to Alloy like traces/logs. We chose not to because:

**1. Different semantics**

Prometheus metrics are cumulative counters. OTLP metrics can be delta-based. Converting between them requires care to avoid data loss or double-counting.

**2. PromQL ecosystem**

Alert rules, recording rules, and dashboards all use PromQL expecting Prometheus-style metrics. OTLP metrics would need translation layer.

**3. Separation of concerns**

-   Metrics → Prometheus (pull, optimized for aggregation and alerting)
-   Traces/Logs → Alloy → Jaeger/Loki (push, event-based)

This separation keeps each system doing what it's designed for.

### Implementation approach

Each Go service exposes a `/metrics` endpoint using the Prometheus client library. Prometheus is configured to scrape all service targets.

Metrics are labeled with service name, method, path, and status code - enabling SLI queries like "error rate for trip booking endpoint" or "P95 latency for driver search".

## Trade-offs

| Aspect            | Pull (Prometheus)        | Push (OTLP)            |
| ----------------- | ------------------------ | ---------------------- |
| Service discovery | Must configure targets   | Services self-register |
| Health detection  | `up == 0` works          | Silence is ambiguous   |
| Backpressure      | Prometheus handles       | Service must buffer    |
| Ecosystem         | Native PromQL            | Needs translation      |
| Best for          | Aggregated metrics, SLOs | Event-based telemetry  |
