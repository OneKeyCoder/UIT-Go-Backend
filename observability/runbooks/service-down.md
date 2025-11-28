# Runbook: Service Down Alert

## Alert: ServiceDown

### Mô tả
Một hoặc nhiều service không phản hồi health check, có thể đã crash hoặc unresponsive.

### Mức độ nghiêm trọng
- **Critical**: Cần xử lý ngay lập tức

### Impact
- Service không available cho users
- Dependent services có thể bị cascade failure
- Data loss có thể xảy ra nếu transaction đang xử lý

---

## Quy trình xử lý

### Bước 1: Xác nhận service nào down (30 giây)

```bash
# Kiểm tra status tất cả containers
docker compose ps

# Query Prometheus
curl -s "http://localhost:9090/api/v1/query?query=up==0" | jq '.data.result[].metric.job'
```

**Expected output khi healthy:**
```
api-gateway        running
authentication     running
trip-service       running
location-service   running
logger-service     running
```

### Bước 2: Kiểm tra container state (1 phút)

```bash
# Chi tiết container
docker inspect uit-go-<service>-1 --format='{{.State.Status}} - {{.State.Error}}'

# Exit code (nếu crashed)
docker inspect uit-go-<service>-1 --format='{{.State.ExitCode}}'
```

**Exit codes:**
- `0`: Normal exit
- `1`: Application error
- `137`: OOM Killed (out of memory)
- `139`: Segfault
- `143`: SIGTERM (graceful shutdown)

### Bước 3: Thu thập logs (2 phút)

```bash
# Logs gần nhất trước khi crash
docker logs --tail 100 uit-go-<service>-1

# Loki query
{container_name="uit-go-<service>-1"} | json | line_format "{{.level}} {{.msg}}"
```

### Bước 4: Restart service (1 phút)

```bash
# Restart single service
docker compose restart <service-name>

# Ví dụ
docker compose restart api-gateway
docker compose restart authentication
docker compose restart trip-service
```

### Bước 5: Verify recovery (2 phút)

```bash
# Kiểm tra status
docker compose ps <service-name>

# Health check
curl -s http://localhost:8080/health/ready  # api-gateway

# Prometheus target
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health=="up") | .labels.job'
```

---

## Root Cause Scenarios

### Scenario A: OOM Kill (Exit code 137)

**Symptoms:**
- Exit code 137
- Container restart liên tục
- Memory usage cao trước khi crash

**Diagnosis:**
```bash
# Check docker events
docker events --filter container=uit-go-<service>-1 --since 1h

# Check system memory
free -h
docker stats --no-stream
```

**Mitigation:**
```yaml
# Tăng memory limit trong docker-compose.yml
services:
  api-gateway:
    deploy:
      resources:
        limits:
          memory: 1G  # Tăng từ 512M
```

### Scenario B: Application Error (Exit code 1)

**Symptoms:**
- Exit code 1
- Error logs trước khi exit
- Có thể do config sai hoặc dependency không available

**Diagnosis:**
```bash
# Check startup logs
docker logs uit-go-<service>-1 2>&1 | head -50

# Common issues:
# - Database connection failed
# - Config file not found
# - Port already in use
```

**Mitigation:**
```bash
# Fix config/dependency
# Restart với dependencies
docker compose up -d postgres mongo redis rabbitmq
sleep 10
docker compose up -d <service>
```

### Scenario C: Dependency Failure

**Symptoms:**
- Service healthy ban đầu rồi fail
- Logs show connection refused/timeout

**Diagnosis:**
```bash
# Check dependency health
docker compose ps postgres mongo redis rabbitmq

# Test connectivity từ trong container
docker exec uit-go-api-gateway-1 nc -zv postgres 5432
docker exec uit-go-api-gateway-1 nc -zv redis 6379
```

**Mitigation:**
```bash
# Restart dependency trước
docker compose restart postgres
sleep 30
docker compose restart <affected-service>
```

### Scenario D: Resource Exhaustion (File descriptors, connections)

**Symptoms:**
- "too many open files"
- "connection pool exhausted"

**Diagnosis:**
```bash
# Check file descriptors
docker exec uit-go-<service>-1 cat /proc/1/limits | grep "open files"

# Check connections
netstat -an | grep <service-port> | wc -l
```

**Mitigation:**
```yaml
# Tăng ulimits trong docker-compose.yml
services:
  api-gateway:
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
```

### Scenario E: Deadlock

**Symptoms:**
- Service running nhưng không respond
- Health check timeout
- CPU usage thấp

**Diagnosis:**
```bash
# Thread dump (nếu có endpoint)
curl -s http://localhost:8080/debug/pprof/goroutine?debug=2

# Check stuck goroutines
docker logs uit-go-<service>-1 2>&1 | grep -i deadlock
```

**Mitigation:**
```bash
# Force restart
docker compose kill <service>
docker compose up -d <service>
```

---

## Recovery Procedures

### Full Stack Restart (last resort)

```bash
# Stop all
docker compose down

# Clear potential corrupt state
docker volume prune -f

# Start infrastructure first
docker compose up -d postgres mongo redis rabbitmq
sleep 30

# Then services
docker compose up -d

# Verify
docker compose ps
```

### Partial Recovery (recommended)

```bash
# Identify failing services
docker compose ps | grep -v "running\|Up"

# Restart only failed services
docker compose restart <service1> <service2>
```

### Data Recovery (if needed)

```bash
# PostgreSQL backup
docker exec uit-go-postgres-1 pg_dump -U postgres trips > backup.sql

# MongoDB backup
docker exec uit-go-mongo-1 mongodump --out /dump
docker cp uit-go-mongo-1:/dump ./mongo-backup
```

---

## Preventive Measures

### Health Check Configuration

```yaml
# docker-compose.yml
services:
  api-gateway:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/ready"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Restart Policy

```yaml
services:
  api-gateway:
    restart: unless-stopped
    # hoặc
    deploy:
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
```

### Resource Reservation

```yaml
services:
  api-gateway:
    deploy:
      resources:
        reservations:
          cpus: '0.5'
          memory: 256M
        limits:
          cpus: '2'
          memory: 1G
```

---

## Escalation

| Thời gian | Action |
|-----------|--------|
| 0-5 phút | Restart service |
| 5-15 phút | Check logs, identify root cause |
| 15-30 phút | Escalate to L2 |
| 30+ phút | Consider rollback/full restart |

---

## Post-Incident

### Checklist
- [ ] Service đã recovered
- [ ] Health check passing
- [ ] No error logs
- [ ] Alert resolved
- [ ] Incident documented

### Questions to Answer
1. Tại sao service failed?
2. Health check có phát hiện đủ nhanh không?
3. Recovery procedure có hiệu quả không?
4. Cần thêm monitoring gì?

---

## Tài liệu liên quan

- [High Error Rate Runbook](./high-error-rate.md)
- [High Latency Runbook](./high-latency.md)
- [Docker Compose Reference](../../docker-compose.yml)
