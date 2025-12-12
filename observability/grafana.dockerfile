# Custom Grafana image with baked-in provisioning for Azure Container Apps

FROM grafana/grafana:main-ubuntu

# Copy provisioning files (datasources, dashboards, alerting)
COPY grafana/provisioning /etc/grafana/provisioning
COPY grafana-config.ini /grafana/

ENV GF_SERVER_ROOT_URL=${GF_SERVER_ROOT_URL}
# ENV GF_PATHS_CONFIG=/grafana/grafana-config.ini

# Disable analytics
ENV GF_ANALYTICS_REPORTING_ENABLED=false
ENV GF_ANALYTICS_CHECK_FOR_UPDATES=false

# # Health check
# HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
#   CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1