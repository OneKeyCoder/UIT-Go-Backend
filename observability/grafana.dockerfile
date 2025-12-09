# Custom Grafana image with baked-in provisioning for Azure Container Apps
# ACA doesn't support bind mounts, so we copy provisioning files into the image

FROM grafana/grafana:12.2.3

# Copy provisioning files (datasources, dashboards, alerting)
COPY grafana/provisioning /etc/grafana/provisioning

# Environment variables for Azure deployment
ENV GF_SERVER_ROOT_URL=https://grafana.yourdomain.com
ENV GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}

# Disable analytics
ENV GF_ANALYTICS_REPORTING_ENABLED=false
ENV GF_ANALYTICS_CHECK_FOR_UPDATES=false

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1

USER grafana
