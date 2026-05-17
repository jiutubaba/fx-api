#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="/data/fuxi-api/staging"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "${APP_ROOT}/data" "${APP_ROOT}/reports"
if [ ! -f "${APP_ROOT}/.env" ]; then
  cp "${SCRIPT_DIR}/env.staging.example" "${APP_ROOT}/.env"
  echo "Created ${APP_ROOT}/.env. Fill secrets and rerun."
  exit 1
fi

cp "${SCRIPT_DIR}/docker-compose.app.yml" "${APP_ROOT}/docker-compose.yml"
cd "${APP_ROOT}"
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d
docker compose --env-file .env -f docker-compose.yml ps
echo "Staging prepared at http://127.0.0.1:3200. Keep Caddy pointing staging.fuxiapi.top here only after migration checks pass."
