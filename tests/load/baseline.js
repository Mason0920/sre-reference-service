/**
 * Baseline load test: /health and /version endpoints
 *
 * Usage:
 *   k6 run tests/load/baseline.js
 *   BASE_URL=http://myhost:8080 k6 run tests/load/baseline.js
 *
 * Stages: ramp 0→50 RPS (30s), hold 50 RPS (60s), ramp down (30s)
 * Acceptance: p95 < 50ms, error rate < 0.1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { BASE_URL, SLO } from './config.js';

const errorRate = new Rate('error_rate');
const healthLatency = new Trend('health_latency', true);
const versionLatency = new Trend('version_latency', true);

export const options = {
  scenarios: {
    baseline: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 20,
      maxVUs: 50,
      stages: [
        { target: 50, duration: '30s' },  // ramp up to 50 RPS
        { target: 50, duration: '60s' },  // hold 50 RPS
        { target: 0,  duration: '30s' },  // ramp down
      ],
    },
  },
  thresholds: {
    http_req_duration: [
      `p(50)<${SLO.p50}`,
      `p(95)<${SLO.p95}`,
      `p(99)<${SLO.p99}`,
    ],
    error_rate: [`rate<${SLO.errorRate}`],
    health_latency: [`p(95)<${SLO.p95}`],
    version_latency: [`p(95)<${SLO.p95}`],
  },
};

export default function () {
  // Test /health
  const healthRes = http.get(`${BASE_URL}/health`, {
    tags: { endpoint: 'health' },
  });
  const healthOk = check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health body ok': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'ok';
      } catch {
        return false;
      }
    },
  });
  errorRate.add(!healthOk);
  healthLatency.add(healthRes.timings.duration);

  // Test /version
  const versionRes = http.get(`${BASE_URL}/version`, {
    tags: { endpoint: 'version' },
  });
  const versionOk = check(versionRes, {
    'version status 200': (r) => r.status === 200,
    'version has version field': (r) => {
      try {
        const body = JSON.parse(r.body);
        return typeof body.version === 'string';
      } catch {
        return false;
      }
    },
  });
  errorRate.add(!versionOk);
  versionLatency.add(versionRes.timings.duration);

  sleep(0.1);
}

export function handleSummary(data) {
  return {
    'tests/load/results/baseline-summary.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data),
  };
}

function textSummary(data) {
  const m = data.metrics;
  const d = m.http_req_duration;
  const fmt = (v) => (v == null ? 'N/A' : Number(v).toFixed(2));
  return `
=== Baseline Load Test Results ===
Endpoint: ${BASE_URL}
Requests: ${m.http_reqs?.values.count} total, rate=${fmt(m.http_reqs?.values.rate)}/s

Latency (all endpoints):
  p50:  ${fmt(d?.values['p(50)'])}ms
  p95:  ${fmt(d?.values['p(95)'])}ms (SLO: <${SLO.p95}ms)
  p99:  ${fmt(d?.values['p(99)'])}ms (SLO: <${SLO.p99}ms)
  max:  ${fmt(d?.values.max)}ms

/health p95:  ${fmt(m.health_latency?.values['p(95)'])}ms
/version p95: ${fmt(m.version_latency?.values['p(95)'])}ms

Error rate: ${fmt((m.error_rate?.values.rate ?? 0) * 100)}% (SLO: <${SLO.errorRate * 100}%)

Checks passed: ${m.checks?.values.passes ?? 0} / ${(m.checks?.values.passes ?? 0) + (m.checks?.values.fails ?? 0)}
`;
}
