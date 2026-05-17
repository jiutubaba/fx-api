#!/usr/bin/env bash
set -euo pipefail

if [ "${CONFIRM_SWITCH:-}" != "staging.fuxiapi.top" ]; then
  echo "Refusing to switch staging. Rerun with CONFIRM_SWITCH=staging.fuxiapi.top after migration checks pass."
  exit 1
fi

CADDYFILE="${CADDYFILE:-/etc/caddy/Caddyfile}"
SITE="${SITE:-staging.fuxiapi.top}"
NEW_TARGET="${NEW_TARGET:-127.0.0.1:3200}"
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

start = None
brace_depth = 0
site_re = re.compile(rf"^\s*{re.escape(site)}(?:\s|,|\{{)")
for i, line in enumerate(lines):
    if start is None:
        if site_re.search(line):
            start = i
            brace_depth += line.count("{") - line.count("}")
        continue
    brace_depth += line.count("{") - line.count("}")
    if re.match(r"^\s*reverse_proxy\s+", line):
        indent = re.match(r"^(\s*)", line).group(1)
        lines[i] = f"{indent}reverse_proxy {target}\n"
        path.write_text("".join(lines))
        raise SystemExit(0)
    if brace_depth <= 0:
        break

raise SystemExit(f"reverse_proxy for site {site!r} not found in {path}")
PY

caddy validate --config "${CADDYFILE}"
systemctl reload caddy
echo "Staging Caddy switched to ${NEW_TARGET}. Backup: ${BACKUP}"
