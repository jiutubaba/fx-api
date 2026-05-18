#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="/data/fuxi-api/staging"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "${APP_ROOT}/data" "${APP_ROOT}/reports" "${APP_ROOT}/legal"
if [ ! -f "${APP_ROOT}/.env" ]; then
  cp "${SCRIPT_DIR}/env.staging.example" "${APP_ROOT}/.env"
  echo "Created ${APP_ROOT}/.env. Fill secrets and rerun."
  exit 1
fi

cp "${SCRIPT_DIR}/docker-compose.app.yml" "${APP_ROOT}/docker-compose.yml"
cp -R "${SCRIPT_DIR}/legal/." "${APP_ROOT}/legal/"
cd "${APP_ROOT}"
docker compose --env-file .env -f docker-compose.yml pull
docker compose --env-file .env -f docker-compose.yml up -d
docker compose --env-file .env -f docker-compose.yml ps
server_port="$(awk -F= '$1 == "SERVER_PORT" {print $2; exit}' .env)"
server_port="${server_port:-3200}"
for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${server_port}/health" >/dev/null; then
    echo "Staging health check passed at http://127.0.0.1:${server_port}/health"
    exit 0
  fi
  sleep 2
done
echo "Staging container started, but health check did not pass within 60s. Check docker logs for ${APP_CONTAINER_NAME:-fuxi-api-staging}."
exit 1
