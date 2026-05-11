// Load test configuration and SLO targets
export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// SLO targets (ms)
export const SLO = {
  p50: 20,   // p50 < 20ms
  p95: 50,   // p95 < 50ms
  p99: 100,  // p99 < 100ms
  errorRate: 0.001, // error rate < 0.1%
};

// Test scenarios
export const SCENARIOS = {
  baseline: {
    rps: 50,
    duration: '60s',
    description: 'Baseline: 50 RPS for 1m',
  },
  peak: {
    rps: 100,
    duration: '60s',
    description: 'Expected peak: 100 RPS for 1m',
  },
  peak2x: {
    rps: 200,
    duration: '60s',
    description: '2x peak: 200 RPS for 1m (acceptance criteria)',
  },
};
