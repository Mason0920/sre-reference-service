# Observability Stack

Prometheus + Grafana for metrics collection. Docker Compose-based deployment.

## Architecture

```
┌─────────────────┐     scrape      ┌────────────────┐
│   hslee-app     │ ◄────────────── │                │
│  :8080/metrics  │                 │   Prometheus   │
└─────────────────┘                 │    :9090       │
                                    │                │
┌─────────────────┐     scrape      │                │
│  node-exporter  │ ◄────────────── │                │
│    :9100        │                 └───────┬────────┘
└─────────────────┘                         │
                                            │ query
┌─────────────────┐     probe       ┌───────▼────────┐
│  blackbox-exp   │ ◄────────────── │                │
│    :9115        │                 │    Grafana     │
└─────────────────┘                 │    :3000       │
                                    └────────────────┘
```

## Quick Start

```bash
cd observability
docker compose up -d
```

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Components

| Component | Port | Purpose |
|-----------|------|---------|
| Prometheus | 9090 | Metrics collection, storage, alerting rules |
| Grafana | 3000 | Dashboards and visualization |
| node-exporter | 9100 | Host CPU/memory/disk/network metrics |
| blackbox-exporter | 9115 | HTTP endpoint probing |

## Dashboards

### SRE Golden Signals (`sre-signals.json`)

Monitors the four golden signals as defined by Google SRE:

| Signal | Metrics | SLO |
|--------|---------|-----|
| **Request Rate** | `http_requests_total` | — |
| **Error Rate** | `http_requests_total{status=~"5.."}` | < 0.1% |
| **Latency p50** | `http_request_duration_seconds` histogram | — |
| **Latency p95** | `http_request_duration_seconds` histogram | < 500ms |
| **Latency p99** | `http_request_duration_seconds` histogram | < 2s |
| **Saturation (CPU)** | `node_cpu_seconds_total` | < 85% |
| **Saturation (Mem)** | `node_memory_MemAvailable_bytes` | < 85% |

### Infrastructure Metrics (`infrastructure.json`)

Host-level metrics via node-exporter: CPU utilization and IOWait, memory breakdown (used/buffers/cached), disk usage per mountpoint, disk I/O (read/write bytes/s), network throughput per interface.

## Application Instrumentation

The `hslee-app` job in `prometheus.yml` scrapes `host.docker.internal:8080/metrics`. The app must expose a `/metrics` endpoint in Prometheus exposition format.

For a Go application using the standard `prometheus/client_golang` library:

```go
import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

Default Go metrics included: `go_goroutines`, `go_memstats_*`, `process_cpu_seconds_total`, `process_resident_memory_bytes`.

For HTTP request metrics, add middleware using `promhttp.InstrumentHandlerDuration` and `promhttp.InstrumentHandlerCounter`.

## Prometheus Configuration

### Scrape Targets

| Job | Target | Interval |
|-----|--------|----------|
| `prometheus` | localhost:9090 | 15s |
| `node` | node-exporter:9100 | 15s |
| `hslee-app` | host.docker.internal:8080 | 15s |
| `blackbox-http` | /health, /version endpoints | 15s |

### Alert Rules (`prometheus/rules/slo.yml`)

| Alert | Condition | Severity |
|-------|-----------|----------|
| `HighErrorRate` | 5xx error rate > 1% for 5m | warning |
| `EndpointDown` | `probe_success == 0` for 1m | critical |
| `HighLatencyP95` | p95 latency > 500ms for 5m | warning |
| `HighLatencyP99` | p99 latency > 1s for 5m | critical |

### Retention

30 days TSDB retention. Approximately 1–2 GB storage per million samples.

## Directory Structure

```
observability/
├── docker-compose.yml
├── prometheus/
│   ├── prometheus.yml       # Scrape config, alertmanager config
│   ├── blackbox.yml         # Blackbox exporter probe modules
│   └── rules/
│       └── slo.yml          # Recording rules and alert rules
└── grafana/
    ├── provisioning/
    │   ├── datasources/
    │   │   └── prometheus.yml   # Auto-provision Prometheus datasource
    │   └── dashboards/
    │       └── default.yml      # Dashboard provider config
    └── dashboards/
        ├── sre-signals.json     # Golden Signals dashboard
        └── infrastructure.json  # Node Exporter dashboard
```

## Operations

### Reload Prometheus config

```bash
curl -X POST http://localhost:9090/-/reload
```

### Check alert status

```bash
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[]'
```

### Add new scrape target

Edit `prometheus/prometheus.yml` and reload:

```yaml
- job_name: my-service
  static_configs:
    - targets:
        - my-service:9090
```

### Persistent data

Prometheus data persists in Docker volume `prometheus_data`. Grafana configuration (dashboards edited via UI, users) persists in `grafana_data`.

To reset:
```bash
docker compose down -v
```

## Integrations

- **HSLA-20**: SLO definitions and error budget dashboards build on metrics collected here.
- **HSLA-22**: Uptime monitoring complements the blackbox-exporter endpoint health checks.
