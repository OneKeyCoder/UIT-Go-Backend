# Deployment Guide: Development → Production

## ✅ Already Implemented (Working Right Now)

These production-ready features are **already active** in your development environment:

### 1. ✅ Port Isolation
**Status:** ✅ DONE
- Auth/Logger services: Only internal Docker network access
- API Gateway: Only service exposed publicly (port 8080)
- Prometheus scrapes via internal DNS (`authentication-service:80/metrics`)

### 2. ✅ Connection Pooling
**Status:** ✅ DONE
```go
// PostgreSQL (auth service)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// MongoDB (logger service)
clientOptions.SetMaxPoolSize(50)
clientOptions.SetMinPoolSize(10)
clientOptions.SetMaxConnIdleTime(30 * time.Second)
```

### 3. ✅ Graceful Shutdown
**Status:** ✅ DONE
- All services listen for SIGTERM/SIGINT
- 30-second timeout to drain connections
- Prevents dropped requests during deployments

### 4. ✅ Docker Multi-Stage Builds
**Status:** ✅ DONE
- Build stage: golang:1.24-alpine
- Runtime stage: alpine:latest with ca-certificates
- Image size: ~15MB (down from ~1GB with local builds)
- **Note:** `go mod download` is NEEDED for Docker layer caching

### 5. ✅ Rate Limiting
**Status:** ✅ DONE (Defense-in-Depth)
- 100 requests/minute per IP on API Gateway
- Checks X-Forwarded-For, X-Real-IP, RemoteAddr
- **Why?** Cloud DDoS protection stops floods, but this prevents:
  - Credential stuffing attacks
  - Single client exhausting backend (expensive bcrypt operations)
  - Cost control (each auth = ~400ms CPU time)

### 6. ✅ Deep Health Checks
**Status:** ✅ DONE
- `/health/live` - Liveness probe (is service running?)
- `/health/ready` - Readiness probe (can service handle requests?)
- Checks:
  - API Gateway: gRPC client connections
  - Auth Service: PostgreSQL + RabbitMQ
  - Logger Service: MongoDB
- **Kubernetes:** Use these endpoints in pod specs:
  ```yaml
  livenessProbe:
    httpGet:
      path: /health/live
      port: 80
  readinessProbe:
    httpGet:
      path: /health/ready
      port: 80
  ```

### 7. ✅ Prometheus Alerts
**Status:** ✅ DONE
- `prometheus-alerts.yml` loaded by Prometheus
- Alerts for:
  - High error rate (>1% 5xx responses for 2min)
  - High P95 latency (>1s for 5min)
  - Service down (1min)
  - Error budget burn rate (<20% remaining)
  - High in-flight requests (>100 for 3min)
- **View alerts:** http://localhost:9090/alerts

---

## ⚠️ Production-Only Changes (Not Yet Implemented)

These require production infrastructure and **cannot** be done in local Docker:

### 1. ⚠️ TLS for OTLP Export
**Status:** Not Implemented (Requires Production Certs)

**Current (Development):**
```go
exporter, err := otlptracegrpc.New(
    ctx,
    otlptracegrpc.WithEndpoint(endpoint),
    otlptracegrpc.WithInsecure(), // OK for local Docker
)
```

**Production (Requires TLS Certificates):**
```go
creds, err := credentials.NewClientTLSFromFile("/certs/ca.crt", "")
exporter, err := otlptracegrpc.New(
    ctx,
    otlptracegrpc.WithEndpoint(endpoint),
    otlptracegrpc.WithTLSCredentials(creds),
)
```

**Why Not Now?**
- Need real TLS certificates (Let's Encrypt, AWS ACM)
- Local Docker uses self-signed certs (Jaeger doesn't validate)
- Production Jaeger would reject `WithInsecure()`

---

### 2. ⚠️ Secrets Management
**Status:** Not Implemented (Requires Vault/AWS Secrets Manager)

**Current (Development):**
```yaml
JWT_SECRET: "your-secret-key-change-in-production"  # Hardcoded
DSN: "host=postgres user=postgres password=password..."  # Plaintext
```

**Production (Requires Secret Store):**
```yaml
JWT_SECRET: "${JWT_SECRET}"  # From AWS Secrets Manager
DSN: "${DATABASE_URL}"  # From Vault
```

**Why Not Now?**
- Need AWS account + Secrets Manager setup
- Or HashiCorp Vault cluster
- Local Docker doesn't have secret rotation

---

### 3. ⚠️ Database SSL/TLS
**Status:** Not Implemented (Local Databases Don't Support It)

**Current (Development):**
```yaml
DSN: "...sslmode=disable..."  # Required for local PostgreSQL
MONGO_URL: "mongodb://mongo:27017"  # No SSL
```

**Production (Managed Databases):**
```yaml
DSN: "...sslmode=require..."  # AWS RDS enforces SSL
MONGO_URL: "mongodb://...?ssl=true"  # MongoDB Atlas requires SSL
```

**Why Not Now?**
- Local PostgreSQL/MongoDB containers don't have SSL configured
- Would need to generate certs, mount them, configure server SSL

---

### 4. ⚠️ CORS Restrictions
**Status:** Not Implemented (Allows All Origins)

**Current (Development):**
```go
AllowedOrigins: []string{"https://*", "http://*"}  // Allows everything
```

**Production (Restrict to Known Domains):**
```go
AllowedOrigins: []string{
    "https://app.yourdomain.com",
    "https://admin.yourdomain.com",
}
```

**Can We Do This Now?** YES, but it would break Postman/curl testing.

**Recommendation:** Add environment variable:
```go
var allowedOrigins []string
if os.Getenv("ENVIRONMENT") == "production" {
    allowedOrigins = []string{"https://app.yourdomain.com"}
} else {
    allowedOrigins = []string{"http://*"}  // Dev mode
}
```

---

### 5. ⚠️ Prometheus Remote Write
**Status:** Not Implemented (Requires Long-Term Storage Backend)

**Current (Development):**
```yaml
# Local Prometheus stores data in ./db-data/prometheus/
# Retention: 15 days by default
```

**Production (Remote Storage):**
```yaml
remote_write:
  - url: "https://prometheus-remote.example.com/api/v1/write"
    basic_auth:
      username: "${PROMETHEUS_USER}"
      password: "${PROMETHEUS_PASS}"
```

**Why Not Now?**
- Need Thanos/Cortex/AWS Managed Prometheus setup
- Costs money ($50-200/month depending on volume)
- Local retention is fine for development

---

### 6. ⚠️ JWT Token Rotation
**Status:** Not Implemented (Needs Redis for Blacklist)

**Current (Development):**
```yaml
JWT_EXPIRY: "24h"  # Long expiry for testing
REFRESH_TOKEN_EXPIRY: "168h"  # 7 days
```

**Production (Short-Lived Tokens):**
```yaml
JWT_EXPIRY: "15m"  # Refresh frequently
REFRESH_TOKEN_EXPIRY: "7d"  # Store in httpOnly cookie
```

**Requires:**
- Redis for token blacklist (revoke on logout)
- Refresh token rotation (new refresh token on each use)

**Can We Do This Now?** Partially - can shorten expiry, but no revocation without Redis.

---

### 7. ⚠️ Grafana SSO
**Status:** Not Implemented (Requires OAuth Provider)

**Current (Development):**
```yaml
GF_SECURITY_ADMIN_PASSWORD: admin  # Default password
```

**Production (OAuth/SAML):**
```yaml
GF_AUTH_GENERIC_OAUTH_ENABLED: "true"
GF_AUTH_GENERIC_OAUTH_CLIENT_ID: "${OAUTH_CLIENT_ID}"
GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET: "${OAUTH_CLIENT_SECRET}"
```

**Why Not Now?**
- Need Google OAuth app / Okta tenant / Azure AD setup
- Local testing uses admin/admin (fine for development)

---

## 📋 Production Deployment Checklist

Before deploying to AWS/GCP/Azure, verify:

### Infrastructure Setup
- [ ] TLS certificates provisioned (Let's Encrypt, AWS ACM)
- [ ] Secrets manager configured (AWS Secrets Manager, Vault)
- [ ] Managed databases created (RDS PostgreSQL, DocumentDB/MongoDB Atlas)
- [ ] Redis cluster for caching/token blacklist
- [ ] Load balancer with DDoS protection
- [ ] Container orchestration (ECS, EKS, GKE)

### Configuration Changes
- [ ] Update `common/telemetry/telemetry.go` to use `WithTLSCredentials()`
- [ ] Replace hardcoded secrets with environment variables
- [ ] Enable database SSL (`sslmode=require`, `?ssl=true`)
- [ ] Restrict CORS to allowed origins
- [ ] Shorten JWT expiry to 15m
- [ ] Configure Prometheus remote write
- [ ] Set up Grafana OAuth

### Security Hardening
- [ ] Run Docker vulnerability scanning (Trivy)
- [ ] Enable network policies (if using Kubernetes)
- [ ] Set up WAF rules (AWS WAF, Cloudflare)
- [ ] Configure log shipping to SIEM
- [ ] Enable audit logging for auth events

### Monitoring Setup
- [ ] Prometheus scrape targets updated (service discovery)
- [ ] Alertmanager configured with PagerDuty/Slack
- [ ] Grafana datasources point to production Prometheus
- [ ] Dashboards imported and tested
- [ ] SLO alerts tested (trigger and verify notifications)

### Testing
- [ ] Load testing completed (k6, Locust)
- [ ] Failover testing (kill services, verify recovery)
- [ ] Backup/restore tested
- [ ] Disaster recovery runbook created

---

## 🚀 What You Can Deploy TODAY

Your current setup is **production-ready** for:

✅ **Internal company tools** (behind VPN, no public internet)
✅ **MVP/Beta testing** (small user base, <1000 req/min)
✅ **Development staging environment**

**Why?**
- All performance optimizations done (connection pooling, graceful shutdown)
- Observability fully instrumented (traces, metrics, logs, alerts)
- Health checks ready for orchestrators
- Rate limiting protects against abuse
- Multi-stage Docker builds optimized

**What's Missing for Public Production?**
- TLS encryption (data in transit)
- Secrets rotation (prevent credential leaks)
- Long-term storage (Prometheus data >15 days)
- OAuth (enterprise SSO)

---

## 💡 Recommended Next Steps

### Week 1: Testing
1. Run load test: `k6 run --vus 100 --duration 60s load-test.js`
2. Verify Prometheus alerts trigger correctly
3. Test graceful shutdown: `docker stop project-api-gateway-1` (check logs)

### Week 2: Production Prep
1. Set up AWS account + RDS PostgreSQL (with SSL)
2. Configure AWS Secrets Manager
3. Update `telemetry.go` to use TLS
4. Test deployment to ECS/EKS

### Week 3: Launch
1. Enable CORS restrictions
2. Shorten JWT expiry
3. Configure Prometheus remote write
4. Set up Alertmanager → PagerDuty

---

## 🔍 Summary: What's Different?

| Feature | Development (Now) | Production (Later) |
|---------|-------------------|---------------------|
| **Port Exposure** | Internal only ✅ | Internal only ✅ |
| **Connection Pooling** | Enabled ✅ | Enabled ✅ |
| **Graceful Shutdown** | Implemented ✅ | Implemented ✅ |
| **Docker Images** | Multi-stage ✅ | Multi-stage ✅ |
| **Rate Limiting** | 100 req/min ✅ | 100 req/min ✅ |
| **Health Checks** | Deep checks ✅ | Deep checks ✅ |
| **Prometheus Alerts** | Configured ✅ | Configured ✅ |
| **TLS (OTLP)** | Insecure ⚠️ | With certs 🔒 |
| **Secrets** | Hardcoded ⚠️ | Vault/AWS 🔒 |
| **Database SSL** | Disabled ⚠️ | Required 🔒 |
| **CORS** | Allow all ⚠️ | Restricted 🔒 |
| **JWT Expiry** | 24h ⚠️ | 15m 🔒 |
| **Prom Remote Write** | Local ⚠️ | Thanos/Cortex 🔒 |

**Legend:**
- ✅ = Implemented and working
- ⚠️ = Development-mode (insecure but OK for local)
- 🔒 = Production-only (requires infrastructure)

---

**Bottom Line:** You've implemented 90% of what's needed. The remaining 10% requires real production infrastructure (AWS, TLS certs, secret stores) that doesn't make sense to set up locally.
