#!/usr/bin/env bash
# Fires a synthetic alert to Alertmanager to verify Slack routing end-to-end.
# Usage: ./scripts/test-alert.sh [alertmanager-url]
set -euo pipefail

ALERTMANAGER_URL="${1:-http://localhost:9093}"

echo "Sending synthetic test alert to ${ALERTMANAGER_URL}..."

STARTS_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
ENDS_AT="$(date -u -v+10M +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '+10 minutes' +%Y-%m-%dT%H:%M:%SZ)"

PAYLOAD=$(cat <<EOF
[
  {
    "labels": {
      "alertname": "SyntheticTestAlert",
      "severity": "info",
      "environment": "production",
      "synthetic": "true",
      "instance": "test-host"
    },
    "annotations": {
      "summary": "Synthetic test alert — Alertmanager → Slack routing verified",
      "description": "This alert was fired manually via scripts/test-alert.sh to verify alerting pipeline integration."
    },
    "startsAt": "${STARTS_AT}",
    "endsAt": "${ENDS_AT}"
  }
]
EOF
)

HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "${ALERTMANAGER_URL}/api/v2/alerts" \
  -H "Content-Type: application/json" \
  -d "${PAYLOAD}")

if [ "${HTTP_STATUS}" = "200" ]; then
  echo "✓ Alert fired (HTTP 200). Check #alerts Slack channel."
  echo ""
  echo "To resolve the alert, run:"
  echo "  curl -s -X DELETE '${ALERTMANAGER_URL}/api/v2/alerts' ..."
  echo "  (Alertmanager auto-resolves when endsAt passes or alert stops firing)"
else
  echo "✗ Unexpected HTTP status: ${HTTP_STATUS}"
  exit 1
fi
