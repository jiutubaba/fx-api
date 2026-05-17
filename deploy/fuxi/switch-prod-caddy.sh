#!/usr/bin/env bash
set -euo pipefail

if [ "${CONFIRM_SWITCH:-}" != "fuxiapi.top" ]; then
  echo "Refusing to switch production. Rerun with CONFIRM_SWITCH=fuxiapi.top after explicit approval."
  exit 1
fi

CADDYFILE="${CADDYFILE:-/etc/caddy/Caddyfile}"
OLD_TARGET="${OLD_TARGET:-127.0.0.1:3000}"
NEW_TARGET="${NEW_TARGET:-127.0.0.1:3300}"
BACKUP="${CADDYFILE}.bak.$(date +%Y%m%d-%H%M%S)"

cp "${CADDYFILE}" "${BACKUP}"
python3 - "$CADDYFILE" "$OLD_TARGET" "$NEW_TARGET" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
old = sys.argv[2]
new = sys.argv[3]
text = path.read_text()
if old not in text:
    raise SystemExit(f"old target {old!r} not found in {path}")
path.write_text(text.replace(old, new, 1))
PY

caddy validate --config "${CADDYFILE}"
systemctl reload caddy
echo "Production Caddy switched to ${NEW_TARGET}. Backup: ${BACKUP}"
