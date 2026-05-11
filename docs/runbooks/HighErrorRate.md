# Runbook: HighErrorRate

| Field | Value |
|-------|-------|
| **Severity** | warning |
| **Source** | `observability/prometheus/rules/slo.yml` |
| **Alert Group** | `alerts.availability` |
| **Condition** | `job:http_error_ratio:rate5m > 0.01` for 5m |
| **SLO Impact** | Availability SLO breach (error budget burn) |
| **Slack Channel** | `#alerts` |

## Summary

The HTTP 5xx error rate for a service has exceeded 1% of total requests over the past 5 minutes, threatening the availability SLO.

## Symptoms

- Slack alert: `[FIRING] HighErrorRate` with `job` label
- Grafana SRE Golden Signals dashboard shows Error Rate spike
- Users reporting errors or degraded service
- Error budget is burning faster than expected

## Root Cause Analysis

1. **Check current error rate**
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=job:http_error_ratio:rate5m' | \
     jq '.data.result[] | {job: .metric.job, error_rate: .value[1]}'
   ```

2. **Check which status codes are returning errors**
   ```bash
   curl -s 'http://localhost:9090/api/v1/query?query=rate(http_requests_total{status=~"5.."}[5m])' | \
     jq '.data.result[]'
   ```

3. **Check application logs for errors**
   ```bash
   kubectl logs deployment/<service> -n <namespace> --tail=200 | grep -i "error\|panic\|fatal"
   # Or via Loki (Grafana > Explore > Loki)
   {job="<service>"} |= "error" | level = "error"
   ```

4. **Check recent deployments**
   ```bash
   kubectl rollout history deployment/<service> -n <namespace>
   ```

5. **Check upstream dependencies**
   ```bash
   kubectl logs deployment/<service> -n <namespace> | grep -i "connection refused\|timeout\|unavailable"
   ```

6. **Check resource pressure on the pod**
   ```bash
   kubectl top pod -n <namespace> | grep <service>
   kubectl describe pod -n <namespace> | grep -A5 "Conditions"
   ```

## Remediation

### Recent bad deployment — rollback
```bash
kubectl rollout undo deployment/<service> -n <namespace>
kubectl rollout status deployment/<service> -n <namespace>
```

### Upstream dependency failure
```bash
# Identify failing dependency from logs, then check its health
kubectl get pod -n <namespace> | grep <dependency>
curl -v http://<dependency-service>:<port>/health
```

### Overloaded service — scale out
```bash
kubectl scale deployment/<service> --replicas=<n> -n <namespace>
```

### Circuit breaker / retry causing cascading errors
- Review retry logic and backoff configuration in service code
- Consider enabling circuit breaker if not already present

### Verify recovery
```bash
curl -s 'http://localhost:9090/api/v1/query?query=job:http_error_ratio:rate5m' | \
  jq '.data.result[] | select(.metric.job=="<service>")'
# Should be < 0.01 (1%)
```

## Escalation Path

1. **0–10 min**: On-call investigates using steps above; rolls back if recent deploy
2. **10–20 min**: Involve service owner if root cause unclear
3. **20+ min** or **error rate > 5%**: Trigger incident, notify engineering lead

---
*Source: [observability/prometheus/rules/slo.yml](../../observability/prometheus/rules/slo.yml)*
