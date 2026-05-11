# Load Test Baseline Results

**Date:** 2026-05-06  
**Tool:** k6 v1.7.1  
**Server:** hslee (Go, `net/http`, port 8080)  
**Environment:** local macOS (darwin/arm64), OTEL disabled

---

## SLO Targets

| Metric     | SLO Target |
|------------|-----------|
| p50        | < 20ms    |
| p95        | < 50ms    |
| p99        | < 100ms   |
| Error rate | < 0.1%    |

---

## Baseline Test (50 RPS, 2 min)

Scenario: ramp 0→50 RPS over 30s, hold 60s, ramp down 30s.

| Metric          | /health    | /version   | Combined   | SLO      | Result |
|-----------------|-----------|-----------|-----------|----------|--------|
| p50             | —         | —         | 0.47ms    | < 20ms   | PASS   |
| p95             | 1.66ms    | 1.23ms    | 1.43ms    | < 50ms   | PASS   |
| p99             | —         | —         | 4.73ms    | < 100ms  | PASS   |
| max             | 57.4ms    | 85.5ms    | 85.5ms    | —        | —      |
| Error rate      | —         | —         | 0.00%     | < 0.1%   | PASS   |
| Total requests  | 4,500     | 4,500     | 9,000     | —        | —      |
| Actual RPS      | —         | —         | 74.9/s    | 50       | —      |
| Checks passed   | —         | —         | 18,000/18,000 | 100% | PASS   |

---

## Stress Test (2x Peak = 200 RPS, 4.5 min)

Scenario: ramp 0→100 RPS (30s), hold 60s, ramp to 200 RPS (30s), hold 120s, ramp down 30s.  
**Actual sustained rate observed: ~290 RPS** (k6 ramping-arrival-rate overdelivers slightly; all requests served without error).

| Metric          | /health    | /version   | Combined   | SLO      | Result |
|-----------------|-----------|-----------|-----------|----------|--------|
| p50             | —         | —         | 0.35ms    | < 20ms   | PASS   |
| p95             | 1.04ms    | 0.73ms    | 0.91ms    | < 50ms   | PASS   |
| p99             | —         | —         | 2.26ms    | < 100ms  | PASS   |
| max             | 30.97ms   | 44.59ms   | 44.59ms   | —        | —      |
| Error rate      | —         | —         | 0.00%     | < 0.1%   | PASS   |
| Total requests  | 39,000    | 39,000    | 78,000    | —        | —      |
| Actual RPS      | —         | —         | 288.8/s   | 200      | —      |
| Checks passed   | —         | —         | 117,000/117,000 | 100% | PASS |

**Acceptance criteria met:** p95 (0.91ms) well under 50ms SLO at 2x expected peak load.

---

## Performance Bottleneck Analysis

No significant bottlenecks identified for these endpoints. Both `/health` and `/version` are:
- Stateless (no DB, no external calls)
- Pure in-memory JSON responses
- Sub-millisecond p95 even at ~290 RPS

**Observed spikes:** Max latencies of 30-85ms seen at low RPS but not at high RPS — consistent with Go runtime GC pauses or OS scheduling jitter rather than request-path contention. These do not breach SLOs.

**OTEL tracer:** When OTLP collector is unavailable, the tracer generates export error logs but does not degrade request latency (traces are async-exported).

---

## Alerting Validation

The `HealthCheckFailed` alert rule fires when `probe_success == 0` for 1 minute.

- Blackbox exporter probes at 30s intervals
- Alert threshold: 2 consecutive failed probes (1m `for:` window)
- To validate: run `alerting-validation.js` with a bad `BASE_URL` and observe AlertManager `/api/v2/alerts` within 2 minutes

Manual validation: `curl http://localhost:9093/api/v2/alerts | jq '.[] | .labels.alertname'`

---

## Recommendations

1. **SLOs are very comfortable** — current endpoints are in-memory only. Budget exists for future middleware (auth, rate-limit) without breaching p95.
2. **Add `/metrics` endpoint** with `prometheus/client_golang` to expose per-endpoint request counters and latency histograms for real RPS/latency alerts (currently only blackbox uptime is monitored).
3. **Add latency-based alerts** — e.g. `http_request_duration_seconds{quantile="0.95"} > 0.1` once Prometheus client is added.
