# MEMORY

Last updated: 2026-05-18

## Hot State

- Fuxi API production has been cut over from old `new-api` to new sub2api-based `fuxi-api-prod`.
- Production Caddy currently routes `fuxiapi.top`, `www.fuxiapi.top`, and `api.fuxiapi.top` to `127.0.0.1:3300`.
- Staging Caddy routes `staging.fuxiapi.top` to `127.0.0.1:3200`.
- Old `new-api` and `new-api-staging` containers remain running and preserved for rollback.
- Old `/data/new-api/**` and local `F:\newAPI` are preserved backup and legacy reference sources.
- Latest repository commit at cutover: `a8dbd12b Configure fuxi proxy runtime defaults`.
- Latest production follow-up commit: `870e51b4 Archive account visibility follow-up`.
- Post-cutover cleanup boundary: old `new-api`, `new-api-staging`, `/data/new-api/**`, and local `F:\newAPI` are protected rollback/backup resources, not disposable residue.
- Production candidate image verified after cutover: `sha256:59879147153e0714e6eaf5df61f4c51168d6a9d65b24d3ab8388a479d66b714e`.
- Caddy backup from cutover: `/etc/caddy/Caddyfile.bak.20260518-062510`.
- Fresh production migration report: `/data/fuxi-api/prod/reports/prod-20260518-135648.md`.

## Verified After Cutover

- `https://fuxiapi.top/api/status`: 200
- `https://fuxiapi.top/health`: 200
- `https://fuxiapi.top/`: 200
- Authenticated `https://fuxiapi.top/v1/models`: 200
- `free` group `gpt-5.4-mini` chat completion: 200
- `auto` group `gpt-5.4-mini` chat completion: 200
- `plus` group chat completion: 403 `INSUFFICIENT_BALANCE`, expected because the selected user balance is zero.
- Production Redis persistence and write checks passed after fixing volume ownership.
- Follow-up deploy for account/channel visibility completed; `fuxi-api-prod` is healthy on `ghcr.io/jiutubaba/fx-api:latest`.
- Post-cutover local verification passed: `go test ./...`, frontend `typecheck`, and frontend `lint:check`.
- Production feature flags are enabled: `available_channels_enabled=true`, `channel_monitor_enabled=true`, `channel_monitor_default_interval_seconds=60`.
- GET `https://fuxiapi.top/api/status`: 200. Use GET for this endpoint; HEAD is not registered by Gin for the route.
- GET `https://fuxiapi.top/health`: 200.
- Old `/data/new-api`, `new-api`, and `new-api-staging` were reconfirmed present after the follow-up deploy and remain intentionally preserved.

## Active Rollback

To roll production Caddy back to the old container:

```bash
CONFIRM_SWITCH=fuxiapi.top NEW_TARGET=127.0.0.1:3000 /data/fuxi-api/deploy/switch-prod-caddy.sh
```

## Recent Fixes

- `0d007d76 Migrate legacy auto group bindings`
- `feb84b83 Fix fuxi redis volume ownership`
- `a8dbd12b Configure fuxi proxy runtime defaults`
- `409cb4ee Improve account and channel visibility`
- `870e51b4 Archive account visibility follow-up`

## Current Risks

- `plus` has an active key but zero user balance, so it returns 403 until balance or group policy changes.
- Some old consumption logs are archived only because old token/channel rows did not map to target rows.
- Legacy custom OAuth bindings were archived only; do not assume they are active in the new schema.
- `security.url_allowlist.enabled=false` remains a runtime warning inherited from current config; evaluate separately before tightening upstream URL policy.
- Production `channels` is empty after migration; available-channel UI currently relies on the account-pool fallback from `accounts` and `groups`.
- Production `channel_monitors` is empty; channel status UI displays an "unmonitored" account-pool fallback until real monitor tasks are configured.
- Deleting or overwriting old rollback resources still requires a future project rule change; current rules prohibit it.
