/**
 * Stress / peak load test: 2x expected peak (acceptance criteria)
 *
 * Usage:
 *   k6 run tests/load/stress.js
 *   BASE_URL=http://myhost:8080 k6 run tests/load/stress.js
 *
 * Stages: ramp 0→200 RPS (60s), hold 200 RPS (120s), ramp down (30s)
 * Acceptance: p95 < 50ms at 2x peak (200 RPS)
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
    stress: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { target: 100, duration: '30s' },  // ramp to peak (100 RPS)
        { target: 100, duration: '60s' },  // hold peak
        { target: 200, duration: '30s' },  // ramp to 2x peak (200 RPS)
        { target: 200, duration: '120s' }, // hold 2x peak (acceptance window)
        { target: 0,   duration: '30s' },  // ramp down
      ],
    },
  },
  thresholds: {
    http_req_duration: [
      `p(95)<${SLO.p95}`,
      `p(99)<${SLO.p99}`,
    ],
    error_rate: [`rate<${SLO.errorRate}`],
  },
};

export default function () {
  const healthRes = http.get(`${BASE_URL}/health`, {
    tags: { endpoint: 'health' },
  });
  const healthOk = check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health body ok': (r) => {
      try {
        return JSON.parse(r.body).status === 'ok';
      } catch {
        return false;
      }
    },
  });
  errorRate.add(!healthOk);
  healthLatency.add(healthRes.timings.duration);

  const versionRes = http.get(`${BASE_URL}/version`, {
    tags: { endpoint: 'version' },
  });
  const versionOk = check(versionRes, {
    'version status 200': (r) => r.status === 200,
  });
  errorRate.add(!versionOk);
  versionLatency.add(versionRes.timings.duration);

  sleep(0.05);
}

export function handleSummary(data) {
  return {
    'tests/load/results/stress-summary.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data),
  };
}

function textSummary(data) {
  const m = data.metrics;
  const d = m.http_req_duration;
  const fmt = (v) => (v == null ? 'N/A' : Number(v).toFixed(2));
  const p95 = d?.values['p(95)'] ?? Infinity;
  const p99 = d?.values['p(99)'] ?? Infinity;
  const errRate = m.error_rate?.values.rate ?? 0;
  return `
=== Stress Load Test Results (2x Peak) ===
Endpoint: ${BASE_URL}
Requests: ${m.http_reqs?.values.count} total, peak=${fmt(m.http_reqs?.values.rate)}/s

Latency:
  p50:  ${fmt(d?.values['p(50)'])}ms
  p95:  ${fmt(p95)}ms  [SLO: <${SLO.p95}ms]  ${p95 < SLO.p95 ? 'PASS' : 'FAIL'}
  p99:  ${fmt(p99)}ms  [SLO: <${SLO.p99}ms]  ${p99 < SLO.p99 ? 'PASS' : 'FAIL'}
  max:  ${fmt(d?.values.max)}ms

Error rate: ${fmt(errRate * 100)}% [SLO: <${SLO.errorRate * 100}%]  ${errRate < SLO.errorRate ? 'PASS' : 'FAIL'}
`;
}
