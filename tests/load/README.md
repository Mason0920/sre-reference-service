# Load Tests

k6-based load test suite for the hslee service.

## Prerequisites

```bash
brew install k6
```

## Test Scripts

| Script                   | Purpose                                        |
|--------------------------|------------------------------------------------|
| `baseline.js`            | Baseline: ramp to 50 RPS, hold 60s            |
| `stress.js`              | Stress: ramp to 200 RPS (2x peak), hold 120s  |
| `alerting-validation.js` | Validate HealthCheckFailed alert firing        |

## Running

```bash
# Start the server (from repo root)
go run ./cmd/hslee &

# Baseline (50 RPS)
BASE_URL=http://localhost:8080 k6 run tests/load/baseline.js

# Stress / 2x peak (200 RPS)
BASE_URL=http://localhost:8080 k6 run tests/load/stress.js

# Alerting validation (healthy server — alert should NOT fire)
BASE_URL=http://localhost:8080 k6 run tests/load/alerting-validation.js

# Alerting validation (bad URL — HealthCheckFailed SHOULD fire within 1m)
BASE_URL=http://localhost:9999 k6 run tests/load/alerting-validation.js
```

## SLOs

| Metric     | Target  |
|------------|---------|
| p50        | < 20ms  |
| p95        | < 50ms  |
| p99        | < 100ms |
| Error rate | < 0.1%  |

## Baseline Results (2026-05-06)

See [`results/baseline-results.md`](results/baseline-results.md) for full results.

| Test         | RPS  | p95    | p99    | Errors | Result |
|--------------|------|--------|--------|--------|--------|
| Baseline     | 75   | 1.43ms | 4.73ms | 0%     | PASS   |
| Stress (2x)  | 290  | 0.91ms | 2.26ms | 0%     | PASS   |

## Output

Results are written to `tests/load/results/` as JSON summary files.
