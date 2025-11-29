# Hướng dẫn Truy vấn Log với Loki

## Tổng quan

Loki sử dụng LogQL - ngôn ngữ query tương tự PromQL nhưng cho logs. Document này cung cấp các query patterns phổ biến cho UIT-Go.

---

## LogQL Basics

### Cấu trúc Query

```logql
{label_selector} |= "string" | json | line_format "{{.field}}"
│                │           │     │
│                │           │     └── Format output
│                │           └── Parse JSON
│                └── Filter string
└── Select streams
```

### Label Selectors

```logql
# Exact match
{container_name="uit-go-api-gateway-1"}

# Regex match
{container_name=~"uit-go-.*"}

# Exclude
{container_name!~".*postgres.*"}

# Multiple conditions
{container_name=~"uit-go-.*", service="api-gateway"}
```

---

## Query Patterns cho UIT-Go

### 1. Xem logs theo service

```logql
# API Gateway logs
{container_name=~".*api-gateway.*"}

# Authentication service
{container_name=~".*authentication.*"}

# Trip service
{container_name=~".*trip-service.*"}

# Location service
{container_name=~".*location-service.*"}

# Logger service
{container_name=~".*logger-service.*"}
```

### 2. Filter theo log level

```logql
# Chỉ errors
{container_name=~"uit-go-.*"} |= "error"

# Errors và warnings
{container_name=~"uit-go-.*"} |~ "error|warn"

# Exclude debug
{container_name=~"uit-go-.*"} != "debug"

# JSON log level (nếu log format JSON)
{container_name=~"uit-go-.*"} | json | level="error"
```

### 3. Parse JSON logs

```logql
# Parse và filter
{container_name=~"uit-go-.*"} 
  | json 
  | level="error" 
  | line_format "{{.time}} [{{.level}}] {{.msg}}"

# Extract specific fields
{container_name=~"uit-go-.*"} 
  | json 
  | trace_id != ""
  | line_format "trace_id={{.trace_id}} msg={{.msg}}"
```

### 4. Tìm logs theo trace_id

```logql
# Tất cả logs của một trace
{container_name=~"uit-go-.*"} |= "trace_id=abc123"

# Hoặc với JSON parsing
{container_name=~"uit-go-.*"} 
  | json 
  | trace_id="abc123def456"
```

### 5. Tìm request logs

```logql
# HTTP requests
{container_name=~".*api-gateway.*"} |= "request"

# gRPC calls
{container_name=~"uit-go-.*"} |= "grpc"

# Specific endpoint
{container_name=~".*api-gateway.*"} |= "/grpc/auth"
```

### 6. Error investigation

```logql
# Database errors
{container_name=~"uit-go-.*"} |~ "database|connection|postgres|mongo"

# Timeout errors
{container_name=~"uit-go-.*"} |= "timeout"

# Panic/crash
{container_name=~"uit-go-.*"} |~ "panic|fatal|crash"
```

---

## Metric Queries (LogQL v2)

### Rate of log lines

```logql
# Log lines per second
rate({container_name=~"uit-go-.*"}[5m])

# Error logs per second
rate({container_name=~"uit-go-.*"} |= "error" [5m])
```

### Count over time

```logql
# Total errors in 1 hour
sum(count_over_time({container_name=~"uit-go-.*"} |= "error" [1h]))

# Errors by service
sum by (container_name) (count_over_time({container_name=~"uit-go-.*"} |= "error" [1h]))
```

### Top errors

```logql
# Highest error rate services
topk(5, sum by (container_name) (rate({container_name=~"uit-go-.*"} |= "error" [5m])))
```

---

## Grafana Explore Examples

### Panel 1: Recent Errors

```logql
{container_name=~"uit-go-.*"} |= "error" | json | line_format "{{.time}} [{{.container_name}}] {{.msg}}"
```

### Panel 2: Request Logs with Latency

```logql
{container_name=~".*api-gateway.*"} 
  | json 
  | latency_ms > 500
  | line_format "{{.method}} {{.path}} - {{.latency_ms}}ms"
```

### Panel 3: Error Count Graph

```logql
sum by (container_name) (rate({container_name=~"uit-go-.*"} |= "error" [5m]))
```

---

## Correlation với Jaeger

### Bước 1: Tìm trace_id từ logs

```logql
{container_name=~"uit-go-.*"} |= "error" | json | line_format "trace_id={{.trace_id}} {{.msg}}"
```

### Bước 2: Click trace_id link trong Grafana

Grafana tự động tạo link đến Jaeger nếu cấu hình datasource đúng.

### Bước 3: Xem tất cả logs của trace

```logql
{container_name=~"uit-go-.*"} | json | trace_id="<paste-trace-id>"
```

---

## Performance Tips

### Do:
- Sử dụng label selectors cụ thể
- Filter sớm bằng `|=` trước khi parse JSON
- Giới hạn time range
- Sử dụng `line_format` để giảm output

### Don't:
- Query tất cả logs không filter
- Parse JSON khi không cần
- Query quá nhiều containers cùng lúc
- Time range quá dài (>7 days)

### Ví dụ tối ưu:

```logql
# BAD - parse JSON cho tất cả logs
{container_name=~".*"} | json | level="error"

# GOOD - filter text trước, chỉ parse khi cần
{container_name=~"uit-go-api-gateway.*"} |= "error" | json | msg=~"database.*"
```

---

## Alerting với LogQL

### Trong Grafana:

1. Create Alert Rule
2. Query type: Loki
3. Expression: 
   ```logql
   sum(rate({container_name=~"uit-go-.*"} |= "error" [5m])) > 0.1
   ```
4. Condition: When value is above 0.1
5. Evaluate every: 1m
6. For: 5m

### Ví dụ Alert Rules:

```yaml
# High error rate
sum(rate({container_name=~"uit-go-.*"} |= "error" [5m])) > 0.5

# Panic detected
count_over_time({container_name=~"uit-go-.*"} |= "panic" [1m]) > 0

# Database connection issues
count_over_time({container_name=~"uit-go-.*"} |= "connection refused" [5m]) > 5
```

---

## Quick Reference

| Task | Query |
|------|-------|
| All logs | `{container_name=~"uit-go-.*"}` |
| Errors only | `\|= "error"` |
| Parse JSON | `\| json` |
| Filter field | `\| field="value"` |
| Regex filter | `\|~ "pattern"` |
| Exclude | `!= "string"` |
| Format output | `\| line_format "{{.field}}"` |
| Count | `count_over_time(...[5m])` |
| Rate | `rate(...[5m])` |

---

## Tài liệu tham khảo

- [LogQL Documentation](https://grafana.com/docs/loki/latest/logql/)
- [Grafana Loki Best Practices](https://grafana.com/docs/loki/latest/best-practices/)
- [Label Best Practices](https://grafana.com/docs/loki/latest/best-practices/#labels)
