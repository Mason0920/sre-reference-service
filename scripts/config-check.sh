#!/usr/bin/env bash
# config-check.sh — Validate required environment variables for the hslee application.
# Exits 0 on success, 1 on any missing or invalid value.
# Usage: APP_ENV=development scripts/config-check.sh
set -euo pipefail

ERRORS=()

check_required() {
  local key="$1"
  local value="${!key:-}"
  if [[ -z "$value" ]]; then
    ERRORS+=("MISSING  $key  (required, no default)")
  fi
}

check_enum() {
  local key="$1"
  local value="${!key:-}"
  shift
  local valid=("$@")
  if [[ -z "$value" ]]; then
    return
  fi
  local found=false
  for v in "${valid[@]}"; do
    [[ "$value" == "$v" ]] && found=true && break
  done
  if [[ "$found" == "false" ]]; then
    ERRORS+=("INVALID  $key=${value}  (must be one of: ${valid[*]})")
  fi
}

check_float_range() {
  local key="$1"
  local value="${!key:-}"
  if [[ -z "$value" ]]; then
    return
  fi
  if ! echo "$value" | grep -qE '^[0-9]*\.?[0-9]+$'; then
    ERRORS+=("INVALID  $key=${value}  (must be a float between 0.0 and 1.0)")
    return
  fi
  local out_of_range
  out_of_range=$(awk -v v="$value" 'BEGIN { print (v < 0 || v > 1) ? "yes" : "no" }')
  if [[ "$out_of_range" == "yes" ]]; then
    ERRORS+=("INVALID  $key=${value}  (must be between 0.0 and 1.0)")
  fi
}

# ── Required ──────────────────────────────────────────────────────────────────
check_required APP_ENV

# ── Enum validation ───────────────────────────────────────────────────────────
check_enum APP_ENV development staging production
check_enum LOG_LEVEL debug info warn error

# ── Optional float range ──────────────────────────────────────────────────────
check_float_range OTEL_TRACES_SAMPLER_ARG

# ── Report ────────────────────────────────────────────────────────────────────
if [[ ${#ERRORS[@]} -gt 0 ]]; then
  echo "config-check FAILED:" >&2
  for e in "${ERRORS[@]}"; do
    echo "  ✗ $e" >&2
  done
  exit 1
fi

echo "config-check PASSED"
printf "  %-35s %s\n" "APP_ENV"              "${APP_ENV}"
printf "  %-35s %s\n" "LOG_LEVEL"            "${LOG_LEVEL:-info (default)}"
printf "  %-35s %s\n" "LISTEN_ADDR"          "${LISTEN_ADDR:-:8080 (default)}"
printf "  %-35s %s\n" "OTEL_SERVICE_NAME"    "${OTEL_SERVICE_NAME:-hslee (default)}"
printf "  %-35s %s\n" "OTEL_EXPORTER_OTLP_ENDPOINT" "${OTEL_EXPORTER_OTLP_ENDPOINT:-(disabled)}"
printf "  %-35s %s\n" "OTEL_TRACES_SAMPLER_ARG" "${OTEL_TRACES_SAMPLER_ARG:-(always sample)}"
