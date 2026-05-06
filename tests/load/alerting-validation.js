/**
 * Alerting threshold validation test
 *
 * Validates that HealthCheckFailed alert fires when service is degraded.
 * Runs against a deliberately wrong URL to trigger probe_success == 0,
 * or uses the real server to confirm health probes pass under load.
 *
 * Usage:
 *   # Validate healthy service (alert should NOT fire)
 *   k6 run tests/load/alerting-validation.js
 *
 *   # Simulate down service (alert SHOULD fire within 1m)
 *   BASE_URL=http://localhost:9999 k6 run tests/load/alerting-validation.js
 *
 *   # Check Prometheus alertmanager for fired alerts
 *   curl http://localhost:9093/api/v2/alerts
 */

import http from 'k6/http';
import { check } from 'k6';
import { Rate } from 'k6/metrics';
import { BASE_URL } from './config.js';

const errorRate = new Rate('error_rate');

// Blackbox exporter interval is 30s, alert fires after 1m (2 failed probes)
// Run for 90s to observe alert behaviour
export const options = {
  scenarios: {
    alerting_probe: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1s',
      duration: '90s',
      preAllocatedVUs: 5,
      maxVUs: 20,
    },
  },
  thresholds: {
    // If targeting a healthy server, expect 0% errors
    error_rate: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/health`, {
    timeout: '5s',
    tags: { scenario: 'alerting-validation' },
  });

  const ok = check(res, {
    'health returns 200': (r) => r.status === 200,
    'health body has status': (r) => {
      try {
        return typeof JSON.parse(r.body).status === 'string';
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!ok);
}

export function handleSummary(data) {
  const m = data.metrics;
  const rate = m.error_rate?.values.rate ?? 0;

  const verdict = rate > 0
    ? `HIGH ERROR RATE (${(rate * 100).toFixed(1)}%) — HealthCheckFailed alert should fire within 1m`
    : `CLEAN (0% errors) — HealthCheckFailed alert should NOT fire`;

  console.log(`\n=== Alerting Validation: ${verdict} ===`);
  console.log(`Requests: ${m.http_reqs.values.count}, rate: ${m.http_reqs.values.rate.toFixed(2)}/s`);
  console.log(`Check AlertManager: curl http://localhost:9093/api/v2/alerts | jq '.[] | .labels.alertname'`);

  return {
    'tests/load/results/alerting-validation-summary.json': JSON.stringify(data, null, 2),
  };
}
