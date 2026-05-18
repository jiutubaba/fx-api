#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="${APP_ROOT:-/data/fuxi-api/staging}"
ENV_FILE="${ENV_FILE:-${APP_ROOT}/.env}"
CADDYFILE="${CADDYFILE:-/etc/caddy/Caddyfile}"
IMAGE="${IMAGE:-ghcr.io/jiutubaba/fx-api:latest}"
STAGING_SITE="${STAGING_SITE:-staging.fuxiapi.top}"
PROD_SITE="${PROD_SITE:-fuxiapi.top}"

failures=0

ok() {
  echo "OK   $*"
}

warn() {
  echo "WARN $*"
}

fail() {
  echo "FAIL $*"
  failures=$((failures + 1))
}

need_cmd() {
  if command -v "$1" >/dev/null 2>&1; then
    ok "command $1"
  else
    fail "missing command $1"
  fi
}

env_value() {
  local key="$1"
  awk -F= -v k="$key" '$1 == k {sub(/^[^=]*=/, ""); print; exit}' "$ENV_FILE"
}

need_env() {
  local key="$1"
  local value
  value="$(env_value "$key" 2>/dev/null || true)"
  if [ -n "$value" ]; then
    ok "env ${key} is set"
  else
    fail "env ${key} is missing or empty in ${ENV_FILE}"
  fi
}

caddy_has_site() {
  local site="$1"
  python3 - "$CADDYFILE" "$site" <<'PY'
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
site = sys.argv[2]
site_re = re.compile(rf"^\s*{re.escape(site)}(?:\s|,|\{{)")
for line in path.read_text().splitlines():
    if site_re.search(line):
        raise SystemExit(0)
raise SystemExit(1)
PY
}

need_cmd docker
need_cmd curl
need_cmd python3
need_cmd caddy

if docker compose version >/dev/null 2>&1; then
  ok "docker compose"
else
  fail "docker compose is unavailable"
fi

postgres_network="$(env_value POSTGRES_NETWORK_NAME 2>/dev/null || true)"
postgres_network="${postgres_network:-new-api-net}"
if docker network inspect "$postgres_network" >/dev/null 2>&1; then
  ok "postgres network ${postgres_network}"
else
  fail "missing postgres network ${postgres_network}"
fi

if [ -d /data/new-api ]; then
  ok "/data/new-api exists and will be left untouched"
else
  warn "/data/new-api not found; confirm old backup path before migration"
fi

if [ -f "$ENV_FILE" ]; then
  ok "env file ${ENV_FILE}"
  for key in DATABASE_HOST DATABASE_USER DATABASE_PASSWORD DATABASE_DBNAME JWT_SECRET TOTP_ENCRYPTION_KEY NEWAPI_SOURCE_DSN; do
    need_env "$key"
  done
else
  fail "missing env file ${ENV_FILE}"
fi

if docker manifest inspect "$IMAGE" >/dev/null 2>&1; then
  ok "image manifest ${IMAGE}"
else
  fail "cannot inspect image manifest ${IMAGE}"
fi

if [ -f "$CADDYFILE" ]; then
  ok "Caddyfile ${CADDYFILE}"
  if caddy validate --config "$CADDYFILE" >/dev/null 2>&1; then
    ok "caddy validate"
  else
    fail "caddy validate failed"
  fi
  caddy_has_site "$STAGING_SITE" && ok "Caddy contains ${STAGING_SITE}" || fail "Caddy missing ${STAGING_SITE}"
  caddy_has_site "$PROD_SITE" && ok "Caddy contains ${PROD_SITE}" || fail "Caddy missing ${PROD_SITE}"
else
  fail "missing Caddyfile ${CADDYFILE}"
fi

if docker ps --format '{{.Names}}' | grep -Eq '^new-api(-staging)?$'; then
  ok "legacy new-api containers are present"
else
  warn "legacy new-api containers were not found by exact name"
fi

if [ "$failures" -gt 0 ]; then
  echo "Preflight failed with ${failures} blocking issue(s)."
  exit 1
fi

echo "Preflight passed."
