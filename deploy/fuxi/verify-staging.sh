#!/usr/bin/env bash
set -euo pipefail

LOCAL_URL="${LOCAL_URL:-http://127.0.0.1:3200}"
PUBLIC_URL="${PUBLIC_URL:-https://staging.fuxiapi.top}"

check_url() {
  local url="$1"
  curl -fsS --max-time 10 "${url}" >/dev/null
  echo "OK ${url}"
}

check_url "${LOCAL_URL}/health"
check_url "${LOCAL_URL}/api/status"
check_url "${LOCAL_URL}/setup/status"

if [ "${CHECK_PUBLIC:-false}" = "true" ]; then
  check_url "${PUBLIC_URL}/health"
  check_url "${PUBLIC_URL}/api/status"
fi

if [ -n "${API_KEY:-}" ]; then
  curl -fsS --max-time 20 \
    -H "Authorization: Bearer ${API_KEY}" \
    "${PUBLIC_URL}/v1/models" >/dev/null
  echo "OK ${PUBLIC_URL}/v1/models"
else
  echo "Skipped /v1/models because API_KEY is not set."
fi
