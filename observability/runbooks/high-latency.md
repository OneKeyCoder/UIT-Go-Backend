# Runbook: High Latency Alert

## Alert: SLO_HighLatency_P95

### Mô tả
P95 latency của API Gateway vượt ngưỡng 500ms, gây ảnh hưởng đến user experience.

### Mức độ nghiêm trọng
- **Warning**: Cần điều tra trong vòng 30 phút
- **Critical** (nếu > 2s): Cần xử lý ngay

### Impact
- User experience chậm chạp
- Timeout errors có thể tăng
- Downstream services có thể bị cascade effect

---

## Quy trình xử lý

### Bước 1: Xác nhận vấn đề (2 phút)

1. **Kiểm tra Grafana Dashboard**
   - Mở http://localhost:3000
   - Vào Dashboard "UIT-Go SLO/SLI"
   - Xác nhận P95 Latency panel

2. **Query trực tiếp**
   ```promql
   # P95 latency hiện tại
   histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))
   
   # So sánh với baseline (24h trước)
   histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m] offset 24h)) by (le, service))
   ```

### Bước 2: Xác định bottleneck (5 phút)

1. **Latency breakdown theo endpoint**
   ```promql
   histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
   ```

2. **Tìm slow traces trong Jaeger**
   - Mở http://localhost:16686
   - Service: api-gateway
   - Min Duration: 500ms
   - Limit: 20

3. **Phân tích trace**
   - Xem span nào chiếm thời gian nhiều nhất
   - Database query? External API? Processing?

### Bước 3: Kiểm tra resource usage (5 phút)

```bash
# CPU & Memory
docker stats --no-stream

# Goroutines (nếu có metric)
curl -s http://localhost:9090/api/v1/query?query=go_goroutines

# Database connections
docker exec uit-go-postgres-1 psql -U postgres -c "
SELECT 
  application_name,
  state,
  count(*) 
FROM pg_stat_activity 
WHERE datname IS NOT NULL 
GROUP BY application_name, state;"
```

### Bước 4: Kiểm tra dependency latency

```bash
# PostgreSQL query time
docker exec uit-go-postgres-1 psql -U postgres -c "
SELECT 
  query,
  calls,
  mean_exec_time,
  total_exec_time
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 10;"

# MongoDB stats
docker exec uit-go-mongo-1 mongosh --eval "db.serverStatus().opcounters"

# Redis latency
docker exec uit-go-redis-1 redis-cli --latency-history -i 1
```

---

## Root Cause Scenarios

### Scenario A: Slow Database Queries

**Symptoms:**
- Trace shows database spans > 200ms
- Connection pool exhausted

**Diagnosis:**
```promql
# Database operation latency (nếu có metric)
histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le, operation))
```

**Mitigation:**
```bash
# Identify slow queries
docker exec uit-go-postgres-1 psql -U postgres -c "
SELECT query, mean_exec_time 
FROM pg_stat_statements 
WHERE mean_exec_time > 100 
ORDER BY mean_exec_time DESC;"

# Có thể cần:
# 1. Add index
# 2. Optimize query
# 3. Tăng connection pool
# 4. Add caching
```

### Scenario B: N+1 Query Problem

**Symptoms:**
- Nhiều database spans trong một trace
- Mỗi span nhỏ nhưng tổng lớn

**Mitigation:**
```bash
# Review code for N+1 patterns
# Implement batch loading
# Use JOIN instead of multiple queries
```

### Scenario C: External API Slowdown

**Symptoms:**
- External API spans chiếm phần lớn latency
- Không phải do code của mình

**Mitigation:**
```bash
# Implement timeout
# Add caching layer
# Circuit breaker pattern
# Fallback response
```

### Scenario D: GC Pressure

**Symptoms:**
- Spike latency không đều
- Memory usage cao

**Diagnosis:**
```promql
# GC duration (nếu có metric)
rate(go_gc_duration_seconds_sum[5m]) / rate(go_gc_duration_seconds_count[5m])
```

**Mitigation:**
```bash
# Restart service (short term)
docker compose restart api-gateway

# Long term: Profile memory allocation
# Reduce allocation in hot paths
```

### Scenario E: Load Spike

**Symptoms:**
- Request rate tăng đột biến
- Latency tăng tương ứng

**Diagnosis:**
```promql
rate(http_requests_total[5m])
```

**Mitigation:**
```bash
# Rate limiting
# Auto-scaling (nếu có)
# Caching aggressive hơn
```

---

## Quick Fixes

### Restart service
```bash
docker compose restart api-gateway
```

### Clear caches (nếu có issue)
```bash
docker exec uit-go-redis-1 redis-cli FLUSHALL
```

### Tăng resources
```yaml
# docker-compose.yml
services:
  api-gateway:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
```

---

## Xác nhận fix

1. **Monitor P95 latency**
   ```promql
   histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[1m])) by (le))
   ```

2. **Test manual**
   ```bash
   # Time một request
   time curl -s http://localhost:8080/health
   
   # Multiple requests để check consistency
   for i in {1..10}; do
     time curl -s http://localhost:8080/health > /dev/null
   done
   ```

---

## Prevention

### Code Review Checklist
- [ ] Có timeout cho external calls
- [ ] Sử dụng batch operations thay vì N+1
- [ ] Index cho các query phổ biến
- [ ] Caching cho data ít thay đổi

### Monitoring Additions
```yaml
# Thêm alert cho database latency
- alert: DatabaseSlowQueries
  expr: histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le)) > 0.1
  for: 5m
  labels:
    severity: warning
```

---

## Tài liệu liên quan

- [High Error Rate Runbook](./high-error-rate.md)
- [Database Optimization Guide](../docs/database.md)
- [Caching Strategy](../docs/caching.md)
