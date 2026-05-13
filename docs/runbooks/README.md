# Incident Response Runbooks

Actionable runbooks for on-call engineers. Each runbook targets a 5-minute time-to-mitigate goal.

| Runbook | Alert | Severity |
|---------|-------|----------|
| [High Error Rate (5xx Spike)](./high-error-rate.md) | Fast Burn 1h / 6h | Critical |
| [High Latency / p99 SLO Breach](./high-latency-p99.md) | p99 Latency Breach | Warning |
| [Deployment Rollback](./deployment-rollback.md) | Rollback Triggered | Critical |
| [Service Down / Health Check Failure](./service-down.md) | Health Check Failure | Critical |

## SLO Targets (quick reference)

| Service | SLO | Error Budget |
|---------|-----|-------------|
| API availability | ≥ 99.9% | 43.2 min/month |
| API error rate | < 0.1% | 43.2 min/month |
| API p99 latency | < 500ms | — |
| Health check availability | ≥ 99.95% | 21.6 min/month |
| Health check p99 | < 100ms | — |

Full SLO definitions: [`docs/SLO.md`](../SLO.md)
Rollback reference: [`docs/ROLLBACK.md`](../ROLLBACK.md)
Grafana dashboards: `grafana/dashboards/`

## Escalation Contacts

- On-call: PagerDuty schedule
- Engineering lead: Slack `#incidents`
- Status page: update for outages > 5 min
