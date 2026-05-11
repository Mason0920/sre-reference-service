# Runbook: HighLatencyP99

| Field | Value |
|-------|-------|
| **Severity** | critical |
| **Source** | `observability/prometheus/rules/slo.yml` |
| **Alert Group** | `alerts.latency` |
| **Condition** | `job:http_request_duration_p99:rate5m > 0.5` for 5m |
| **Threshold** | p99 latency > 500ms |
| **SLO Impact** | Latency SLO critical breach |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighLatencyP95](./HighLatencyP95.md) (>300ms warning) |

## Summary

The 99th percentile HTTP request duration has exceeded 500ms for 5 minutes. 1% of requests are severely slow; this typically indicates a systemic issue.

## Symptoms

- Slack alert: `[FIRING] HighLatencyP99`
- Grafana shows p99 latency >> 500ms
- Users experiencing timeouts or very slow responses
- Downstream services may be timing out waiting for responses

## Root Cause Analysis

1. **Check current p99 latency and compare with p95**
   ```bash
   # Both together
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p99:rate5m' | \
     jq '.data.result[]'
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p95:rate5m' | \
     jq '.data.result[]'
   ```

2. **Check for lock contention / long-running requests**
   ```bash
   kubectl exec -it <pod> -n <namespace> -- /bin/sh
   # Get goroutine dump (Go services)
   kill -SIGQUIT <pid>  # dumps goroutines to stderr
   ```

3. **Check database slow query log**
   ```bash
   kubectl logs deployment/<db> -n <namespace> | grep -i "slow\|lock"
   ```

4. **Check for resource saturation**
   ```bash
   kubectl top pods -n <namespace>
   kubectl top nodes
   # Check CPU throttling
   kubectl describe pod <pod> -n <namespace> | grep -i throttl
   ```

5. **Check for network latency between services**
   ```bash
   kubectl exec -it <pod> -n <namespace> -- ping <dependency-svc>
   kubectl exec -it <pod> -n <namespace> -- curl -w "@/dev/stdin" -o /dev/null -s http://<svc>/health <<< "%{time_total}"
   ```

6. **Check thread pool / connection pool exhaustion**
   ```bash
   kubectl logs deployment/<service> -n <namespace> | grep -i "pool\|queue\|waiting\|timeout"
   ```

7. **Check if specific endpoints are affected**
   ```
   histogram_quantile(0.99, sum by (job, handler, le) (rate(http_request_duration_seconds_bucket[5m])))
   ```

## Remediation

### Immediate: Scale out to reduce per-pod load
```bash
kubectl scale deployment/<service> --replicas=<n> -n <namespace>
kubectl rollout status deployment/<service> -n <namespace>
```

### Connection pool exhaustion — increase pool size
```bash
# Edit service config/env to increase DB or HTTP connection pool size
kubectl edit deployment/<service> -n <namespace>
# Update relevant env var (e.g., DB_MAX_CONNECTIONS)
```

### Recent deployment regression — rollback
```bash
kubectl rollout undo deployment/<service> -n <namespace>
```

### CPU throttling — increase CPU limit
```bash
kubectl edit deployment/<service> -n <namespace>
# Increase resources.limits.cpu
```

### Shed load (enable rate limiting or circuit breaker)
- Review service's rate limiting configuration
- Enable circuit breaker if downstream is overwhelmed

### Verify recovery
```bash
curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p99:rate5m' | \
  jq '.data.result[] | select(.metric.job=="<service>")'
# Should be < 0.5 (500ms)
```

## Escalation Path

1. **Immediate**: On-call scales out and checks for obvious resource saturation
2. **< 5 min**: If not improving, page engineering lead
3. **< 10 min**: If downstream database/dependency, involve DB or infra team
4. **If cascading to other services**: Trigger full incident

---
*Source: [observability/prometheus/rules/slo.yml](../../observability/prometheus/rules/slo.yml)*
