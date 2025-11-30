# SLO-based Alerting with Prometheus

## Context

Module D requires:

-   Alerts when SLO violation risk detected
-   Alerts for specific service/function failures
-   Runbooks linked to each alert

Two alerting approaches available:

1. **Prometheus Alerting**: Alert rules defined in YAML, evaluated by Prometheus, notifications via Alertmanager
2. **Grafana Alerting**: Alert rules defined in Grafana UI, Grafana queries datasources and evaluates conditions

## Decision

Use **Prometheus Alerting** with rules in configuration files, notifications via Alertmanager.

## Rationale

### Why Prometheus Alerting over Grafana Alerting?

**1. Evaluation happens at the data source**

Prometheus Alerting evaluates rules directly where metrics are stored. Grafana Alerting works differently - Grafana must query the datasource, pull data over network, then evaluate.

This matters for SLO alerts:

-   **Prometheus**: Evaluates expressions against local data every 30 seconds. No network latency. If Grafana goes down, alerts still fire.

-   **Grafana**: Must query Prometheus → wait for response → evaluate. If Grafana is overloaded or restarting, alert evaluation stops. Network issues between Grafana and Prometheus delay detection.

For SLO alerts that need to detect budget burn within minutes, having evaluation at the source eliminates a failure point.

**2. Recording rules enable complex SLO calculations**

Prometheus supports recording rules - pre-computed metrics stored as new time series. Grafana has no equivalent.

For multi-window burn-rate alerting (Google SRE pattern), we need to compare error rates across multiple time windows (1h, 6h, 24h, 72h) simultaneously. Without recording rules:

-   **Grafana approach**: Each alert evaluation runs 4+ expensive histogram queries across large time ranges
-   **Prometheus approach**: Recording rules pre-compute 5-minute error rates. Alert rules do simple arithmetic on pre-computed values

Example: "Alert if burning 14x budget over 1h AND 7x over 6h" - with recording rules, this is comparing two pre-computed numbers. Without, it's aggregating millions of histogram buckets at query time.

**3. Runbook linking is first-class**

Prometheus alert rules have a native `runbook_url` annotation field. When alert fires, this URL is part of the alert payload. Alertmanager passes it to notification templates. Grafana displays it in alert details.

Grafana Alerting requires workarounds - embedding URLs in description text or using custom labels. The URL is not a structured field, making it harder to template notifications consistently.

**4. Rules as code vs UI-driven configuration**

Prometheus alert rules are YAML files in version control. Rule changes go through PR review, have history, can be rolled back, can be deployed across environments consistently.

Grafana Alerting rules are stored in Grafana's database. To version control, you must export to JSON, manage separately, and re-import. This creates drift risk between what's in Git and what's actually running. Provisioning via configuration is possible but more complex than Prometheus's native file-based approach.

For Module D demonstration where reproducibility matters, file-based rules ensure the grader sees exactly what's defined in the repository.

**5. Alertmanager provides routing, grouping, silencing**

Prometheus ecosystem includes Alertmanager, designed specifically for alert delivery:

-   **Grouping**: Multiple related alerts (e.g., all instances of same service down) become one notification
-   **Routing**: Different alerts go to different channels (critical → PagerDuty, warning → Slack)
-   **Silencing**: Suppress known alerts during maintenance windows
-   **Inhibition**: If "cluster down" fires, suppress individual "service down" alerts

Grafana Alerting has notification policies with similar concepts, but Alertmanager's model is more mature and widely documented in SRE literature.

### Where Grafana Alerting wins

**Multi-datasource correlation**: Grafana can create alerts that combine Prometheus metrics + Loki log counts + external APIs in one rule. Prometheus can only alert on its own metrics.

For our use case, all SLOs are defined in terms of Prometheus metrics (request rates, latencies, error counts). We don't need to correlate with other datasources for alerting. Log-based alerts would be a different use case with different trade-offs.

### Defined SLO Alerts

Each alert targets a specific SLI with clear threshold:

| SLO Target                   | SLI Metric               | Alert Condition | Severity |
| ---------------------------- | ------------------------ | --------------- | -------- |
| Trip booking success > 99.9% | Error rate (5xx/total)   | > 0.1% for 5m   | critical |
| Driver search P95 < 200ms    | histogram_quantile(0.95) | > 200ms for 5m  | warning  |
| Authentication P95 < 100ms   | histogram_quantile(0.95) | > 100ms for 5m  | warning  |
| API Gateway P95 < 500ms      | histogram_quantile(0.95) | > 500ms for 5m  | warning  |
| Service availability         | up metric                | down for 1m     | critical |

Every alert includes `runbook_url` pointing to `observability/runbooks/` with troubleshooting steps.

### Multi-window burn-rate alerting

Following Google SRE book's recommendation, critical SLOs use burn-rate based detection:

-   Fast burn (14x) over 1h window → 2% error budget consumed → page immediately
-   Medium burn (7x) over 6h window → 5% consumed → page
-   Slow burn (1x) over 3 days → track in dashboard, no page

This catches both sudden outages (fast burn) and gradual degradation (slow burn) while avoiding alert fatigue from brief spikes.

## Consequences

### Benefits

-   Alerts continue firing even if Grafana is down
-   Pre-computed SLO metrics via recording rules
-   Alert rules version controlled alongside application code
-   Alertmanager provides mature routing and grouping

### Trade-offs accepted

-   Cannot create alerts that join Prometheus + Loki data
-   Requires learning PromQL for alert expressions
-   Alertmanager is additional component to operate
