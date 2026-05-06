# Runbook Index

Incident response runbooks for all Prometheus alert rules.

## SLO Alerts (`observability/prometheus/rules/slo.yml`)

| Alert | Severity | Condition | Runbook |
|-------|----------|-----------|---------|
| `HighErrorRate` | warning | 5xx error rate > 1% for 5m | [HighErrorRate.md](./HighErrorRate.md) |
| `EndpointDown` | critical | `probe_success == 0` for 1m | [EndpointDown.md](./EndpointDown.md) |
| `HighLatencyP95` | warning | p95 latency > 300ms for 5m | [HighLatencyP95.md](./HighLatencyP95.md) |
| `HighLatencyP99` | critical | p99 latency > 500ms for 5m | [HighLatencyP99.md](./HighLatencyP99.md) |

## Infrastructure Alerts (`observability/prometheus/rules/alerting.yml`)

| Alert | Severity | Condition | Runbook |
|-------|----------|-----------|---------|
| `HealthCheckFailed` | critical | Health probe fails for 1m | [HealthCheckFailed.md](./HealthCheckFailed.md) |
| `PodRestartingTooOften` | warning | Pod restarts > 3 in 5m | [PodRestartingTooOften.md](./PodRestartingTooOften.md) |
| `HighDiskUsage` | warning | Disk > 85% for 5m | [HighDiskUsage.md](./HighDiskUsage.md) |
| `HighDiskUsageCritical` | critical | Disk > 95% for 5m | [HighDiskUsageCritical.md](./HighDiskUsageCritical.md) |
| `HighMemoryUsage` | warning | Memory > 85% for 5m | [HighMemoryUsage.md](./HighMemoryUsage.md) |
| `HighMemoryUsageCritical` | critical | Memory > 95% for 5m | [HighMemoryUsageCritical.md](./HighMemoryUsageCritical.md) |

## Severity Quick Reference

| Severity | Alertmanager repeat | Action |
|----------|---------------------|--------|
| `critical` | every 1h | Page on-call immediately |
| `warning` | every 4h | Investigate within 15 min |

## Common Commands

```bash
# Check all active alerts
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | {name: .labels.alertname, severity: .labels.severity, state: .state}'

# Reload Prometheus config (no restart)
curl -X POST http://localhost:9090/-/reload

# Check blackbox probe
curl -s "http://localhost:9115/probe?target=<url>&module=http_2xx" | grep probe_success
```
