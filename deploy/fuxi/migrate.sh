#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="${APP_ROOT:-/data/fuxi-api/staging}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_NAME="${SERVICE_NAME:-fuxi-api}"
SOURCE_NAME="${SOURCE_NAME:-staging}"
MODE="${MODE:-dry-run}"
REPORT_DIR="${REPORT_DIR:-/app/reports}"
LEGAL_DOC="${LEGAL_DOC:-/app/legal/user-agreement.md}"
APP_UID="${APP_UID:-1000}"
APP_GID="${APP_GID:-1000}"
REDIS_UID="${REDIS_UID:-999}"
REDIS_GID="${REDIS_GID:-1000}"

case "${SOURCE_NAME}" in
  prod|staging) ;;
  *)
    echo "SOURCE_NAME must be prod or staging."
    exit 1
    ;;
esac

case "${MODE}" in
  dry-run|apply) ;;
  *)
    echo "MODE must be dry-run or apply."
    exit 1
    ;;
esac

if [ "${MODE}" = "apply" ] && [ "${CONFIRM_APPLY:-}" != "${SOURCE_NAME}" ]; then
  echo "Refusing apply migration. Rerun with CONFIRM_APPLY=${SOURCE_NAME} after checking the target database."
  exit 1
fi

if [ ! -f "${APP_ROOT}/.env" ]; then
  echo "Missing ${APP_ROOT}/.env."
  exit 1
fi
mkdir -p "${APP_ROOT}/reports" "${APP_ROOT}/legal" "${APP_ROOT}/redis"
chown -R "${APP_UID}:${APP_GID}" "${APP_ROOT}/reports"
chown -R "${REDIS_UID}:${REDIS_GID}" "${APP_ROOT}/redis"
if [ ! -f "${APP_ROOT}/docker-compose.yml" ]; then
  cp "${SCRIPT_DIR}/docker-compose.app.yml" "${APP_ROOT}/docker-compose.yml"
fi
if [ -d "${SCRIPT_DIR}/legal" ]; then
  cp -R "${SCRIPT_DIR}/legal/." "${APP_ROOT}/legal/"
fi

cd "${APP_ROOT}"
docker compose --env-file .env -f docker-compose.yml run --rm --no-deps \
  -e MIGRATION_SOURCE_NAME="${SOURCE_NAME}" \
  -e MIGRATION_MODE="${MODE}" \
  -e MIGRATION_REPORT_DIR="${REPORT_DIR}" \
  -e MIGRATION_LEGAL_DOC="${LEGAL_DOC}" \
  "${SERVICE_NAME}" sh -ec '
    : "${NEWAPI_SOURCE_DSN:?NEWAPI_SOURCE_DSN is required in .env}"
    if [ -n "${FUXI_TARGET_DSN:-}" ]; then
      target_dsn="${FUXI_TARGET_DSN}"
    else
      target_dsn="postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT:-5432}/${DATABASE_DBNAME}?sslmode=${DATABASE_SSLMODE:-disable}"
    fi
    exec /app/migrate-newapi \
      --source-dsn "${NEWAPI_SOURCE_DSN}" \
      --target-dsn "${target_dsn}" \
      --source-name "${MIGRATION_SOURCE_NAME}" \
      --mode "${MIGRATION_MODE}" \
      --report-dir "${MIGRATION_REPORT_DIR}" \
      --legal-doc "${MIGRATION_LEGAL_DOC}"
  '
