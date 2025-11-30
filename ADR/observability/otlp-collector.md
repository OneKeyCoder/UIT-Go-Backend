# OTLP Collector for Traces and Logs

## Context

Services need to send traces and logs to backends (Jaeger for traces, Loki for logs) for:

-   Distributed tracing across microservices
-   Centralized logging with trace correlation
-   Debugging specific request flows when SLO alerts fire

Two approaches:

1. Services send directly to each backend
2. Services send to a collector (Grafana Alloy) which routes to backends

## Decision

Use **Grafana Alloy** as OTLP collector.

Services send traces and logs via OTLP gRPC to Alloy. Alloy routes to appropriate backends.

## Rationale

### Why a collector is necessary

**Problem**: Loki only supports OTLP over HTTP, not gRPC. Jaeger supports both.

If services send directly to backends, they would need two different exporters configured (gRPC for Jaeger, HTTP for Loki). This means more configuration, more failure points, and more code to maintain.

With Alloy as intermediary:

-   Services only need one exporter (gRPC to Alloy)
-   Alloy handles protocol conversion (gRPC → HTTP for Loki)
-   Backend changes only require Alloy config updates, not service code changes

### Why this matters for SLO debugging

When an SLO alert fires (e.g., "Trip booking error rate exceeds threshold"):

1. Check Grafana alert → see logs from that time window
2. Logs contain `trace_id`
3. Click trace_id → Jaeger shows full request path across services
4. Identify which service/function failed
5. Follow runbook to remediate

This workflow requires traces and logs to share the same `trace_id`. OTLP protocol handles this automatically - both TracerProvider and LoggerProvider use the same trace context.

### OTLP Protocol

OTLP (OpenTelemetry Protocol) defines separate RPC methods for traces, logs, and metrics. When Alloy receives data, the RPC method indicates the data type. No ambiguity, automatic separation.

This is a CNCF standard, so switching backends later (e.g., from Jaeger to Tempo) only requires changing Alloy configuration.

### Why Alloy over OpenTelemetry Collector?

Both accomplish the same goal. We chose Alloy because:

-   Configuration syntax is more readable for our use case
-   Better integration with Grafana ecosystem
-   Functionally equivalent for our requirements

## Trade-offs

| Aspect                  | With Collector        | Direct to Backends       |
| ----------------------- | --------------------- | ------------------------ |
| Service config          | Single endpoint       | Multiple endpoints       |
| Protocol conversion     | Handled centrally     | Each service handles     |
| Backend changes         | Collector config only | All services must update |
| Network hops            | +1 hop                | Direct                   |
| Single point of failure | Collector             | None                     |
