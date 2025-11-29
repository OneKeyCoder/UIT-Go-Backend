# Runbook: High Error Rate Alert

## Alert: SLO_HighErrorRate_FastBurn / SLO_HighErrorRate_SlowBurn

### Mô tả
Error rate của API Gateway vượt quá ngưỡng SLO cho phép, đang tiêu thụ Error Budget nhanh hơn bình thường.

### Mức độ nghiêm trọng
- **FastBurn**: Critical - cần xử lý ngay lập tức
- **SlowBurn**: Warning - cần điều tra trong vòng 1 giờ

### Impact
- User experience bị ảnh hưởng
- Có thể vi phạm SLA với stakeholders
- Error budget đang bị tiêu hao

---

## Quy trình xử lý

### Bước 1: Xác nhận vấn đề (2 phút)

1. **Kiểm tra Grafana Dashboard**
   - Mở http://localhost:3000
   - Vào Dashboard "UIT-Go SLO/SLI"
   - Xác nhận Error Rate panel

2. **Kiểm tra scope ảnh hưởng**
   ```promql
   # Error rate theo service
   sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
   / sum(rate(http_requests_total[5m])) by (service)
   
   # Error rate theo endpoint
   sum(rate(http_requests_total{status=~"5.."}[5m])) by (path, method)
   / sum(rate(http_requests_total[5m])) by (path, method)
   ```

### Bước 2: Thu thập context (5 phút)

1. **Xem recent logs**
   ```bash
   # Loki query trong Grafana Explore
   {service="api-gateway"} |= "error" | json
   
   # Hoặc qua Docker
   docker logs --since 15m uit-go-api-gateway-1 2>&1 | grep -i error
   ```

2. **Tìm trace của request lỗi**
   - Mở Jaeger UI: http://localhost:16686
   - Service: api-gateway
   - Tags: `error=true` hoặc `http.status_code=500`
   - Limit: 20 traces gần nhất

3. **Kiểm tra downstream services**
   ```bash
   # Health check tất cả services
   docker compose ps
   
   # Kiểm tra metrics của từng service
   curl -s http://localhost:9090/api/v1/query?query=up
   ```

### Bước 3: Xác định root cause (10 phút)

#### Scenario A: Database connection issues
```bash
# Kiểm tra PostgreSQL
docker exec uit-go-postgres-1 pg_isready

# Kiểm tra connections
docker exec uit-go-postgres-1 psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"
```

**Mitigation**: Restart service hoặc tăng connection pool

#### Scenario B: Downstream service failure
```bash
# Kiểm tra auth service
curl -s http://localhost:8080/health/ready

# Kiểm tra trip service logs
docker logs --tail 50 uit-go-trip-service-1
```

**Mitigation**: Restart failing service, check resource usage

#### Scenario C: Resource exhaustion
```bash
# Kiểm tra memory/CPU
docker stats --no-stream

# Kiểm tra in-flight requests
curl -s http://localhost:9090/api/v1/query?query=http_requests_in_flight
```

**Mitigation**: Scale up hoặc restart để clear memory

#### Scenario D: Bad deployment
```bash
# Kiểm tra recent deployments
git log --oneline -5

# So sánh với thời điểm bắt đầu lỗi
```

**Mitigation**: Rollback deployment

### Bước 4: Thực hiện mitigation (tùy scenario)

#### Quick fixes:
```bash
# Restart single service
docker compose restart api-gateway

# Restart all services (last resort)
docker compose restart

# Scale down traffic (nếu có load balancer)
# Thêm circuit breaker config
```

### Bước 5: Xác nhận fix (5 phút)

1. **Monitor error rate**
   ```promql
   sum(rate(http_requests_total{service="api-gateway", status=~"5.."}[1m]))
   / sum(rate(http_requests_total{service="api-gateway"}[1m]))
   ```

2. **Chờ alert tự resolve** (thường 5 phút sau khi error rate giảm)

3. **Test manual**
   ```bash
   curl -X POST http://localhost:8080/grpc/auth \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"password123"}'
   ```

---

## Post-incident

### Checklist sau khi fix:
- [ ] Error rate đã về mức bình thường
- [ ] Không còn alert active
- [ ] Test các critical flows thành công
- [ ] Ghi chép incident vào log

### Post-mortem questions:
1. Root cause là gì?
2. Tại sao monitoring không phát hiện sớm hơn?
3. Có thể tự động fix được không?
4. Cần thêm alert nào?
5. Cần cải thiện gì trong code/infra?

---

## Liên hệ escalation

| Level | Contact | Khi nào |
|-------|---------|---------|
| L1 | On-call engineer | Đầu tiên |
| L2 | Team lead | Nếu không fix được trong 30 phút |
| L3 | Architect | Nếu cần thay đổi kiến trúc |

---

## Tài liệu liên quan

- [SLO Definitions](../docs/slo.md)
- [Architecture Overview](../../ARCHITECTURE.md)
- [Deployment Guide](../../README.md)
