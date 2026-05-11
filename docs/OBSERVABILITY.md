# Observability Stack

Prometheus + Grafana for metrics collection. Docker Compose-based deployment with Loki for logs, Jaeger for distributed tracing, and Alertmanager for alert routing.

## Quick Start

```bash
cd observability
docker compose up -d
```

Access:
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / admin or `$GF_ADMIN_PASSWORD`)
- **Alertmanager**: http://localhost:9093
- **Jaeger UI**: http://localhost:16686
- **Loki**: http://localhost:3100

## Architecture

```
┌─────────────────┐     scrape      ┌────────────────────────┐
│   hslee-app     │ ◄────────────── │                        │
│  :8080/metrics  │                 │      Prometheus        │
└─────────────────┘                 │        :9090           │
                                    │                        │
┌─────────────────┐     scrape      │  rules/slo.yml         │
│  node-exporter  │ ◄────────────── │  rules/alerting.yml    │
│    :9100        │                 └───────┬────────────────┘
└─────────────────┘                         │
                                            │ query / alerts
┌─────────────────┐     probe       ┌───────▼────────┐     ┌──────────────┐
│ blackbox-exp    │ ◄────────────── │                │     │ Alertmanager │
│    :9115        │                 │    Grafana     │     │    :9093     │
└─────────────────┘                 │    :3000       │     │  → Slack     │
                                    │                │     └──────────────┘
┌─────────────────┐     push logs   │  Dashboards:   │
│   Promtail      │ ──────────────► │  sre-signals   │
│  (no port)      │                 │  infrastructure│
└─────────────────┘                 │  logs          │
                                    └───────▲────────┘
┌─────────────────┐     OTLP        ┌───────┴────────┐
│   hslee-app     │ ──────────────► │    Jaeger      │
│  (traces)       │                 │    :16686 UI   │
└─────────────────┘                 └────────────────┘
```

## Components

| Component | Port | Purpose |
|-----------|------|---------|
| Prometheus | 9090 | Metrics collection, storage, alerting evaluation |
| Grafana | 3000 | Dashboards and visualization |
| Alertmanager | 9093 | Alert deduplication and Slack routing |
| node-exporter | 9100 | Host CPU/memory/disk/network metrics |
| blackbox-exporter | 9115 | HTTP endpoint probing |
| Loki | 3100 | Log aggregation |
| Promtail | — | Log shipping to Loki (no exposed port) |
| Jaeger | 16686 | Distributed tracing UI (OTLP on 4317/4318) |

## Dashboards

### SRE Golden Signals (`sre-signals.json`)

Covers the four golden signals for the hslee application:

| Panel | Metric |
|-------|--------|
| Request Rate | `job:http_requests_total:rate5m` |
| Error Rate | `job:http_error_ratio:rate5m` |
| Latency p50/p95/p99 | `job:http_request_duration_p{50,95,99}:rate5m` |
| Saturation (CPU) | `node_cpu_seconds_total` |

### Infrastructure Metrics (`infrastructure.json`)

Host-level metrics via node-exporter: CPU utilization, memory breakdown (used/buffers/cached), disk usage per mountpoint, disk I/O (read/write bytes/s), network throughput per interface.

### Logs (`logs.json`)

Loki-backed log explorer with service and level filters. Correlates with traces via Jaeger datasource link.

## Alert Rules

### `prometheus/rules/slo.yml` — SLO Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| `HighErrorRate` | 5xx error rate > 1% for 5m | warning |
| `EndpointDown` | `probe_success == 0` for 1m | critical |
| `HighLatencyP95` | p95 latency > 300ms for 5m | warning |
| `HighLatencyP99` | p99 latency > 500ms for 5m | critical |

### `prometheus/rules/alerting.yml` — Infrastructure Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| `HealthCheckFailed` | Health probe fails 2+ consecutive times | critical |
| `PodRestartingTooOften` | Pod restarts > 3 in 5m | warning |
| `HighDiskUsage` | Disk usage > 85% for 5m | warning |
| `HighDiskUsageCritical` | Disk usage > 95% for 5m | critical |
| `HighMemoryUsage` | Memory usage > 85% for 5m | warning |
| `HighMemoryUsageCritical` | Memory usage > 95% for 5m | critical |

Alerts route to Slack via Alertmanager. Set `SLACK_WEBHOOK_URL` env var before starting.

## Directory Structure

```
observability/
├── docker-compose.yml
├── alertmanager/
│   └── alertmanager.yml       # Slack routing and grouping config
├── loki/
│   └── loki.yml               # Loki storage and retention config
├── promtail/
│   └── promtail.yml           # Log scrape config for Docker containers
├── prometheus/
│   ├── prometheus.yml         # Scrape targets and alertmanager config
│   ├── blackbox.yml           # Blackbox exporter probe modules
│   └── rules/
│       ├── slo.yml            # Recording rules and SLO alert rules
│       └── alerting.yml       # Infrastructure and Kubernetes alert rules
└── grafana/
    ├── provisioning/
    │   ├── datasources/
    │   │   ├── prometheus.yml # Auto-provision Prometheus datasource
    │   │   ├── loki.yml       # Auto-provision Loki datasource
    │   │   └── jaeger.yml     # Auto-provision Jaeger datasource
    │   └── dashboards/
    │       └── default.yml    # Dashboard provider config
    └── dashboards/
        ├── sre-signals.json   # Golden Signals dashboard
        ├── infrastructure.json # Node Exporter dashboard
        └── logs.json          # Loki logs dashboard
```

## Application Instrumentation

Expose Prometheus metrics from the Go application:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

Expose distributed traces via OTLP to Jaeger (`localhost:4317`). See `cmd/server/main.go` for the OpenTelemetry SDK setup.

## Operations

### Reload Prometheus config (no restart)
```bash
curl -X POST http://localhost:9090/-/reload
```

### Check active alerts
```bash
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[]'
```

### Add new scrape target

Edit `observability/prometheus/prometheus.yml`, add a job under `scrape_configs`, then reload:
```bash
curl -X POST http://localhost:9090/-/reload
```

### Persistent data

Prometheus metrics persist in Docker volume `prometheus_data` (30-day retention).
Grafana config persists in `grafana_data`.

To reset:
```bash
docker compose down -v
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GF_ADMIN_PASSWORD` | `admin` | Grafana admin password |
| `SLACK_WEBHOOK_URL` | *(required for alerts)* | Alertmanager Slack webhook |

## Integrations

- **CI/CD**: GitHub Actions workflow publishes metrics via OTLP on test runs.
- **Kubernetes**: `alerting.yml` includes pod restart alerts compatible with `kube-state-metrics`.
- **Loki + Grafana**: Log lines linked to traces via `traceId` field extracted by Promtail pipeline.

## Runbooks

Incident response runbooks for every alert rule are in [`docs/runbooks/`](./runbooks/README.md).

| Alert | Severity | Runbook |
|-------|----------|---------|
| `HealthCheckFailed` | critical | [HealthCheckFailed.md](./runbooks/HealthCheckFailed.md) |
| `PodRestartingTooOften` | warning | [PodRestartingTooOften.md](./runbooks/PodRestartingTooOften.md) |
| `HighDiskUsage` | warning | [HighDiskUsage.md](./runbooks/HighDiskUsage.md) |
| `HighDiskUsageCritical` | critical | [HighDiskUsageCritical.md](./runbooks/HighDiskUsageCritical.md) |
| `HighMemoryUsage` | warning | [HighMemoryUsage.md](./runbooks/HighMemoryUsage.md) |
| `HighMemoryUsageCritical` | critical | [HighMemoryUsageCritical.md](./runbooks/HighMemoryUsageCritical.md) |
| `HighErrorRate` | warning | [HighErrorRate.md](./runbooks/HighErrorRate.md) |
| `EndpointDown` | critical | [EndpointDown.md](./runbooks/EndpointDown.md) |
| `HighLatencyP95` | warning | [HighLatencyP95.md](./runbooks/HighLatencyP95.md) |
| `HighLatencyP99` | critical | [HighLatencyP99.md](./runbooks/HighLatencyP99.md) |
