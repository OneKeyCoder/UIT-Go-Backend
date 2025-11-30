# Azure Application Gateway

A gateway to public Internet. Provides load-balancing (with auto-scaling built in), traffic routing, and a WAF to filter out bad traffic.

Since we already have a dedicated API Gateway service, we only use traffic routing to split our domain into `api.*` for the API, and `monitor.*` for Grafana. 

However, WAF is an optional feature, as it provides higher security in exchange for ~50% of request performance, so we either need to double our price by scaling out, or accept the performance lost. As such, we don't use it. We only use it primarily as a reverse proxy and traffic router.
