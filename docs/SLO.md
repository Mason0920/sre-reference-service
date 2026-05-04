# Service Level Objectives (SLOs)

This document defines SLIs and SLOs for the top user-facing services in the hslee platform. All windows are 30-day rolling unless stated otherwise.

> **Last updated:** 2026-05-04
> **Owner:** DevOps ([@Devops](/Hslee/agents/devops))
> **Related issues:** [HSLA-20](/Hslee/issues/HSLA-20), [HSLA-18](/Hslee/issues/HSLA-18), [HSLA-22](/Hslee/issues/HSLA-22)

---

## 1. Application API — `hslee` HTTP Server

The primary user-facing service running on port 8080.

### SLIs

| SLI | Metric | Query (Prometheus) |
|-----|--------|--------------------|
| Availability | Ratio of HTTP 2xx responses to total requests | `sum(rate(http_requests_total{job="hslee",status=~"2.."}[5m])) / sum(rate(http_requests_total{job="hslee"}[5m]))` |
| Latency p99 | 99th percentile request duration | `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="hslee"}[5m]))` |
| Latency p50 | Median request duration | `histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job="hslee"}[5m]))` |
| Error rate | Ratio of HTTP 5xx responses to total requests | `sum(rate(http_requests_total{job="hslee",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="hslee"}[5m]))` |

### SLOs

| SLO | Target | Error Budget (30 days) |
|-----|--------|------------------------|
| Availability ≥ 99.9% | 99.9% | 43.2 min/month |
| p99 latency < 500ms | ≤ 500ms | N/A (burn-rate alert at > 1s for > 5% of requests) |
| Error rate < 0.1% | < 0.1% | 43.2 min/month |

### Alert Thresholds

| Alert | Condition | Severity |
|-------|-----------|----------|
| Fast burn (1h) | Error budget burn rate > 14.4x for 5min | critical |
| Fast burn (6h) | Error budget burn rate > 6x for 30min | critical |
| Slow burn (3d) | Error budget burn rate > 3x for 6h | warning |
| p99 latency breach | p99 > 500ms for > 5min | warning |

---

## 2. Health Check Endpoint — `GET /health`

Used by Kubernetes liveness probes and external uptime monitors.

### SLIs

| SLI | Metric | Query (Prometheus) |
|-----|--------|--------------------|
| Availability | HTTP 200 responses to `/health` / total requests | `sum(rate(http_requests_total{job="hslee",handler="/health",status="200"}[5m])) / sum(rate(http_requests_total{job="hslee",handler="/health"}[5m]))` |
| Response time | `/health` p99 latency | `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="hslee",handler="/health"}[5m]))` |

### SLOs

| SLO | Target | Error Budget (30 days) |
|-----|--------|------------------------|
| Availability ≥ 99.95% | 99.95% | 21.6 min/month |
| Response time p99 < 100ms | ≤ 100ms | N/A |

---

## 3. CI Pipeline — GitHub Actions

Measures build reliability for the `main` branch pipeline (lint → test → build → smoke).

### SLIs

| SLI | Metric | Source |
|-----|--------|--------|
| Pipeline success rate | Ratio of successful runs to total runs on `main` | GitHub Actions API / Prometheus exporter |
| Build duration p95 | 95th percentile end-to-end pipeline duration | GitHub Actions API |
| Lint pass rate | Ratio of lint-passing runs to total runs | GitHub Actions API |

### SLOs

| SLO | Target | Error Budget (30 days) |
|-----|--------|------------------------|
| Pipeline success rate ≥ 95% | 95% | ~36 failed runs/month (at 24 runs/day) |
| Build duration p95 < 10 min | ≤ 10 min | N/A |
| Lint pass rate ≥ 99% | 99% | ~7 failed lint runs/month |

### Alert Thresholds

| Alert | Condition | Severity |
|-------|-----------|----------|
| High failure rate | CI failure rate > 10% over 24h rolling window | warning |
| Persistent failures | 3 consecutive failures on `main` | critical |
| Build duration spike | p95 > 15 min | warning |

---

## 4. CD Pipeline — Deployment

Measures deployment reliability from merge to production availability.

### SLIs

| SLI | Metric | Source |
|-----|--------|--------|
| Deployment success rate | Ratio of successful deployments to attempted deployments | GitHub Actions API / deployment events |
| Deployment duration p95 | 95th percentile time from merge to traffic switch | GitHub Actions API |
| Rollback rate | Ratio of deployments requiring rollback to total deployments | Deployment events |

### SLOs

| SLO | Target | Error Budget (30 days) |
|-----|--------|------------------------|
| Deployment success rate ≥ 99% | 99% | ~1 failed deployment/month (at 5/day) |
| Deployment duration p95 < 15 min | ≤ 15 min | N/A |
| Rollback rate < 2% | < 2% | N/A |

### Alert Thresholds

| Alert | Condition | Severity |
|-------|-----------|----------|
| Deployment failure | Any failed deployment to staging or production | critical |
| Deployment stalled | Deployment running > 20 min | warning |
| Rollback triggered | Any rollback event | critical |

---

## 5. Smoke Test Suite

Post-deployment validation checks that gates production promotion.

### SLIs

| SLI | Metric | Source |
|-----|--------|--------|
| Smoke test pass rate | Ratio of passing smoke test runs to total runs | GitHub Actions / smoketest binary |
| Smoke test duration p95 | 95th percentile execution time | GitHub Actions |

### SLOs

| SLO | Target | Error Budget (30 days) |
|-----|--------|------------------------|
| Smoke test pass rate ≥ 99.5% | 99.5% | N/A |
| Smoke test duration p95 < 2 min | ≤ 2 min | N/A |

### Alert Thresholds

| Alert | Condition | Severity |
|-------|-----------|----------|
| Smoke test failure | Any failed smoke test on staging or production | critical |
| Smoke test timeout | Run exceeds 3 min | warning |

---

## Error Budget Policy

| Remaining Budget | Action |
|------------------|--------|
| > 50% | Normal development pace, experiments allowed |
| 25–50% | Reduce risky deployments, review recent changes |
| 10–25% | Freeze non-critical changes, focus on reliability |
| < 10% | Freeze all changes, incident declared, postmortem required |
| Exhausted (0%) | Immediate incident response, board notification |

---

## Grafana Dashboards

Error budget dashboards are located at:

- **Application API Error Budget**: `grafana/dashboards/slo-api-error-budget.json`
- **CI/CD Pipeline SLOs**: `grafana/dashboards/slo-cicd-pipeline.json`

Import via: `Grafana UI → Dashboards → Import → Upload JSON file`.

Or provision automatically by placing JSON files under `grafana/provisioning/dashboards/`.

---

## Review Cadence

| Review | Frequency | Owner |
|--------|-----------|-------|
| SLO burn rate | Weekly | DevOps |
| Error budget status | Monthly | DevOps + Engineering |
| SLO target revision | Quarterly | DevOps + Engineering Lead |
