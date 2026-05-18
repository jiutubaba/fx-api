#!/usr/bin/env bash
set -euo pipefail

if [ "${CONFIRM_SWITCH:-}" != "fuxiapi.top" ]; then
  echo "Refusing to switch production. Rerun with CONFIRM_SWITCH=fuxiapi.top after explicit approval."
  exit 1
fi

CADDYFILE="${CADDYFILE:-/etc/caddy/Caddyfile}"
SITE="${SITE:-fuxiapi.top}"
NEW_TARGET="${NEW_TARGET:-127.0.0.1:3300}"
BACKUP="${CADDYFILE}.bak.$(date +%Y%m%d-%H%M%S)"

cp "${CADDYFILE}" "${BACKUP}"
python3 - "$CADDYFILE" "$SITE" "$NEW_TARGET" <<'PY'
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
site = sys.argv[2]
target = sys.argv[3]
lines = path.read_text().splitlines(keepends=True)

site_re = re.compile(rf"(^|[\s,])(?:https?://)?{re.escape(site)}(?=[\s,\{{]|$)")
updated = 0
i = 0
while i < len(lines):
    line = lines[i]
    if not site_re.search(line):
        i += 1
        continue

    brace_depth = line.count("{") - line.count("}")
    j = i + 1
    while j < len(lines):
        block_line = lines[j]
        if re.match(r"^\s*reverse_proxy\s+", block_line):
            indent = re.match(r"^(\s*)", block_line).group(1)
            lines[j] = f"{indent}reverse_proxy {target}\n"
            updated += 1
            break
        brace_depth += block_line.count("{") - block_line.count("}")
        if brace_depth <= 0:
            break
        j += 1
    i = j + 1

if updated == 0:
    raise SystemExit(f"reverse_proxy for site {site!r} not found in {path}")

path.write_text("".join(lines))
PY

caddy validate --config "${CADDYFILE}"
systemctl reload caddy
echo "Production Caddy switched to ${NEW_TARGET}. Backup: ${BACKUP}"
