# Runbook: HighLatencyP95

| Field | Value |
|-------|-------|
| **Severity** | warning |
| **Source** | `observability/prometheus/rules/slo.yml` |
| **Alert Group** | `alerts.latency` |
| **Condition** | `job:http_request_duration_p95:rate5m > 0.3` for 5m |
| **Threshold** | p95 latency > 300ms |
| **SLO Impact** | Latency SLO degradation |
| **Slack Channel** | `#alerts` |
| **Related Alert** | [HighLatencyP99](./HighLatencyP99.md) (>500ms critical) |

## Summary

The 95th percentile HTTP request duration has exceeded 300ms for 5 minutes. 5% of users are experiencing latency above this threshold.

## Symptoms

- Slack alert: `[FIRING] HighLatencyP95` with `job` label
- Grafana SRE dashboard shows p95 latency spike
- Users reporting slow page loads or API responses

## Root Cause Analysis

1. **Check current p95 latency**
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p95:rate5m' | \
     jq '.data.result[] | {job: .metric.job, p95_seconds: .value[1]}'
   ```

2. **Check all percentiles to understand distribution**
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p50:rate5m' | jq '.data.result[]'
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p99:rate5m' | jq '.data.result[]'
   ```

3. **Check latency trend (is it growing?)**
   ```
   # Prometheus query over time:
   job:http_request_duration_p95:rate5m[30m]
   ```

4. **Check CPU and memory pressure**
   ```bash
   kubectl top pods -n <namespace>
   kubectl top nodes
   ```

5. **Check for slow database queries or external calls**
   ```bash
   kubectl logs deployment/<service> -n <namespace> --tail=100 | grep -i "slow\|timeout\|duration"
   ```

6. **Check recent deployments**
   ```bash
   kubectl rollout history deployment/<service> -n <namespace>
   ```

7. **Check if specific endpoints are slow**
   ```
   # In Prometheus / Grafana — break down by handler/path:
   histogram_quantile(0.95, sum by (job, handler, le) (rate(http_request_duration_seconds_bucket[5m])))
   ```

## Remediation

### CPU-bound — scale out
```bash
kubectl scale deployment/<service> --replicas=<n> -n <namespace>
```

### Memory pressure causing GC pauses
```bash
kubectl edit deployment/<service> -n <namespace>
# Increase memory limit to reduce GC frequency
```

### Slow downstream dependency
```bash
# Identify from logs, check dependency health
kubectl get pod -n <namespace> | grep <dependency>
```

### Cache miss / cold start
```bash
# Warm up caches if applicable
# Check cache hit rate metrics
```

### Recent deployment regression — rollback
```bash
kubectl rollout undo deployment/<service> -n <namespace>
```

### Verify recovery
```bash
curl -s 'http://localhost:9090/api/v1/query?query=job:http_request_duration_p95:rate5m' | \
  jq '.data.result[] | select(.metric.job=="<service>")'
# Should be < 0.3 (300ms)
```

## Escalation Path

1. **0–15 min**: On-call investigates root cause (CPU, memory, downstream)
2. **15–30 min**: Involve service owner if no clear cause
3. **If p99 also breaches threshold**: See [HighLatencyP99](./HighLatencyP99.md) (critical)

---
*Source: [observability/prometheus/rules/slo.yml](../../observability/prometheus/rules/slo.yml)*
