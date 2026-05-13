/**
 * Version endpoint load test: /version
 *
 * Usage:
 *   k6 run tests/load/version.js
 *   BASE_URL=http://staging k6 run tests/load/version.js
 *
 * Stages: ramp 0→20 VUs (10s), hold (30s), ramp down (10s)
 * Thresholds: p95 < 500ms, error rate < 1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { BASE_URL } from './config.js';

const errorRate = new Rate('error_rate');

export const options = {
  stages: [
    { duration: '10s', target: 20 },
    { duration: '30s', target: 20 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    error_rate: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/version`);
  const ok = check(res, {
    'version status 200': (r) => r.status === 200,
    'version has version field': (r) => {
      try {
        return typeof JSON.parse(r.body).version === 'string';
      } catch {
        return false;
      }
    },
  });
  errorRate.add(!ok);
  sleep(1);
}

export function handleSummary(data) {
  const m = data.metrics;
  const d = m.http_req_duration;
  const fmt = (v) => (v == null ? 'N/A' : Number(v).toFixed(2));
  const p95 = d?.values['p(95)'] ?? Infinity;
  const errRate = (m.error_rate?.values.rate ?? 0) * 100;
  const pass = p95 < 500 && errRate < 1;

  const html = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>k6 — /version Load Test</title>
<style>body{font-family:monospace;padding:2rem}h1{color:#333}
.pass{color:green}.fail{color:red}table{border-collapse:collapse;margin-top:1rem}
td,th{border:1px solid #ccc;padding:.4rem .8rem}</style></head>
<body>
<h1>Load Test: <code>/version</code> — ${pass ? '<span class="pass">PASS</span>' : '<span class="fail">FAIL</span>'}</h1>
<p>Endpoint: ${BASE_URL}/version</p>
<table>
<tr><th>Metric</th><th>Value</th><th>Threshold</th><th>Result</th></tr>
<tr><td>p95 latency</td><td>${fmt(p95)}ms</td><td>&lt;500ms</td><td class="${p95 < 500 ? 'pass' : 'fail'}">${p95 < 500 ? 'PASS' : 'FAIL'}</td></tr>
<tr><td>error rate</td><td>${fmt(errRate)}%</td><td>&lt;1%</td><td class="${errRate < 1 ? 'pass' : 'fail'}">${errRate < 1 ? 'PASS' : 'FAIL'}</td></tr>
<tr><td>total requests</td><td>${m.http_reqs?.values.count ?? 'N/A'}</td><td>—</td><td>—</td></tr>
<tr><td>p50 latency</td><td>${fmt(d?.values['p(50)'])}ms</td><td>—</td><td>—</td></tr>
<tr><td>p99 latency</td><td>${fmt(d?.values['p(99)'])}ms</td><td>—</td><td>—</td></tr>
</table>
</body></html>`;

  return {
    'tests/load/results/version-summary.html': html,
    'tests/load/results/version-summary.json': JSON.stringify(data, null, 2),
    stdout: `\n=== /version Load Test ===\np95: ${fmt(p95)}ms  error_rate: ${fmt(errRate)}%  ${pass ? 'PASS' : 'FAIL'}\n`,
  };
}
