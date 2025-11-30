# Observability Stack Selection

## Context

### Module D Requirements

1. Define SLOs/SLIs (e.g., trip booking success rate > 99.9%, driver search P95 < 200ms)
2. Define Error Budgets
3. Centralized logging, custom metrics, distributed tracing
4. Dashboard visualizing SLIs/SLOs/Error Budgets
5. Alerting with runbooks

As per Module D requirements we need OUR observability stack to be able to:

-   Track SLOs/SLIs for critical business flows
-   Alert when specific services/functions fail below defined thresholds
-   Provide runbooks for incident response
-   Enable distributed tracing across services

## Considered Options

### Option 1: Azure Monitor

Azure Monitor provides logs, metrics, alerts, distributed tracing (via Application Insights), and dashboards (via Workbooks). It integrates natively with ACA.

**Why we didn't choose it:**

| Requirement             | Azure Monitor Capability                            | Limitation                                                                                                                   |
| ----------------------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Custom SLI per endpoint | Application Insights tracks requests                | SLIs like "trip booking success rate" require custom KQL queries - no first-class SLI definition                             |
| Error Budget            | Can calculate via KQL                               | No native Error Budget concept - must build and maintain custom queries                                                      |
| Multi-burn-rate alerts  | Alerts support thresholds                           | Burn-rate patterns (Google SRE style) require manual implementation in alert rules                                           |
| Runbook linking         | Action Groups can link to Azure Automation Runbooks | Requires learning PowerShell/Python Runbook syntax or using graphical designer - not developer-friendly                      |
| SLO dashboard           | Workbooks can visualize anything                    | Workbooks use drag-and-drop designer or custom JSON schema - Grafana dashboards as JSON/YAML are more familiar to developers |

Azure Monitor can do all of this, but requires significant custom work to implement SLO-focused observability. The building blocks exist, but the SLO patterns aren't first-class citizens. Additionally, the tooling (Workbooks, Automation Runbooks) uses Azure-specific patterns rather than developer-familiar formats.

### Option 2: Azure Managed Prometheus + Managed Grafana

Azure offers managed versions that integrate with ACA:

-   Azure Monitor managed service for Prometheus
-   Azure Managed Grafana

**Network cost consideration:**

ACA services communicate within the VNet. When sending metrics/logs to Azure Managed Prometheus/Grafana:

-   If these services are in the same region and properly networked via Private Endpoints, egress costs may be minimized
-   However, the exact billing for telemetry data transfer between ACA and managed observability services needs verification with Azure pricing documentation
-   Data ingestion costs for Prometheus metrics and Grafana usage still apply

**Capability consideration:**

Azure Managed Prometheus supports PromQL, which enables proper SLO alerting patterns (burn-rate alerts, recording rules for error budgets). This is an advantage over Application Insights for SLO-focused monitoring.

### Option 3: Grafana Cloud

Managed SaaS outside Azure.

**Considerations:**

-   All telemetry data leaves Azure → definite egress costs
-   Latency for queries (data round-trip to external service)
-   Free tier exists and is sufficient for a student project scope

### Option 4: Self-hosted on ACA

Run Prometheus, Loki, Jaeger, Grafana as containers within ACA environment.

**Advantages:**

-   All communication stays within ACA VNet - no egress
-   Full control over configuration
-   Native PromQL support for SLO patterns
-   Native `runbook_url` annotation in Prometheus alerts

**Disadvantages:**

-   Must manage the containers ourselves
-   Uses ACA compute resources
-   No managed backups/HA unless we build it

## Decision

Use **self-hosted observability stack** (Prometheus + Loki + Jaeger + Grafana) running within ACA.

## Rationale

### Why not Azure Monitor?

Azure Monitor has everything needed: logs, metrics, alerts, tracing, dashboards. The issue isn't capability - it's how SLO patterns are supported.

Prometheus was designed with SLO concepts in mind. Recording rules, multi-window alerts, and error budget calculations are documented patterns with community examples. Azure Monitor can achieve the same results, but we'd be building these patterns ourselves in KQL rather than using established PromQL patterns.

For a team already familiar with Prometheus/Grafana ecosystem, self-hosted is faster to implement. For a team with Azure expertise, Azure Monitor would be equally valid.

### Why not Azure Managed Prometheus + Grafana?

This is a valid option and would reduce operational overhead. We chose self-hosted because:

-   **Uncertainty on egress costs**: Need to verify if ACA → Managed Prometheus/Grafana incurs egress charges even within same region
-   **Full control**: Can configure exactly what we need for SLO tracking
-   **Learning**: Understanding how the stack works internally

For a production system with dedicated ops team, Azure Managed Prometheus + Grafana would be worth evaluating.

### Why not Grafana Cloud?

-   Egress costs are certain (data leaves Azure)
-   For a student project, the free tier would likely be sufficient, but we chose to keep data within our infrastructure

### Why self-hosted works for this project

1. **Native SLO support**: Prometheus has first-class support for:

    - Recording rules for SLI calculations
    - Multi-window burn-rate alert patterns
    - Error budget tracking via recording rules

2. **Runbook integration**: Prometheus alerts support `runbook_url` annotation, displayed in Grafana

3. **No egress**: All telemetry stays within ACA VNet

4. **Specific alerting**: Can define alerts for exact service + endpoint + threshold combinations

## Trade-offs

| Aspect               | Self-hosted   | Azure Managed      | Grafana Cloud |
| -------------------- | ------------- | ------------------ | ------------- |
| Egress cost          | None          | Needs verification | Yes           |
| Operational overhead | High          | Low                | None          |
| SLO native support   | Full (PromQL) | Full (PromQL)      | Full          |
| Control              | Full          | Partial            | Partial       |
| Data location        | Within VNet   | Azure              | External      |

## References

-   Google SRE Book: Alerting on SLOs
-   Azure Monitor documentation
-   Prometheus recording rules documentation
