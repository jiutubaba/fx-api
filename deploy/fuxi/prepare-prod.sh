#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="/data/fuxi-api/prod"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_UID="${APP_UID:-1000}"
APP_GID="${APP_GID:-1000}"
REDIS_UID="${REDIS_UID:-999}"
REDIS_GID="${REDIS_GID:-1000}"

mkdir -p "${APP_ROOT}/data" "${APP_ROOT}/reports" "${APP_ROOT}/legal" "${APP_ROOT}/redis"
chown -R "${APP_UID}:${APP_GID}" "${APP_ROOT}/data" "${APP_ROOT}/reports"
chown -R "${REDIS_UID}:${REDIS_GID}" "${APP_ROOT}/redis"
if [ ! -f "${APP_ROOT}/.env" ]; then
  cp "${SCRIPT_DIR}/env.prod.example" "${APP_ROOT}/.env"
  echo "Created ${APP_ROOT}/.env. Fill secrets and rerun."
  exit 1
fi

cp "${SCRIPT_DIR}/docker-compose.app.yml" "${APP_ROOT}/docker-compose.yml"
cp -R "${SCRIPT_DIR}/legal/." "${APP_ROOT}/legal/"
cd "${APP_ROOT}"
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d
docker compose --env-file .env -f docker-compose.yml ps
echo "Production candidate prepared at http://127.0.0.1:3300. Caddy was not changed."
