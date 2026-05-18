# MEMORY

Last updated: 2026-05-18

## Hot State

- Fuxi API production has been cut over from old `new-api` to new sub2api-based `fuxi-api-prod`.
- Production Caddy currently routes `fuxiapi.top`, `www.fuxiapi.top`, and `api.fuxiapi.top` to `127.0.0.1:3300`.
- Staging Caddy routes `staging.fuxiapi.top` to `127.0.0.1:3200`.
- Old `new-api` and `new-api-staging` containers remain running and preserved for rollback.
- Old `/data/new-api/**` and local `F:\newAPI` are preserved backup and legacy reference sources.
- Latest repository commit at cutover: `a8dbd12b Configure fuxi proxy runtime defaults`.
- Latest production follow-up commit: `321f1f31 Record production API key prefix normalization`.
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
- Production release sync completed for image revision `a0177bcb59d932d825e7a82237a432aed9aad886`; `fuxi-api-prod` is healthy with 0 restarts.
- Authenticated `https://fuxiapi.top/v1/models`: 200 after release sync.
- `free` group `gpt-5.4-mini` chat completion: 200 after release sync.
- Production release update completed at 2026-05-18 18:33 +08:00 for image revision `33c5e36c31a1b4f8686526ff15fea934565f9982`; `fuxi-api-prod` runs image `sha256:d2dbb784f80a563d13747bf7c9813e014e7e07d447d31e4b92ed3f175f46d567`, healthy with 0 restarts.
- Verified after the update: local/public `/api/status` 200, homepage 200, `/available-channels` 200, `/monitor` 200, Redis persistence OK, authenticated `/v1/models` 200 (`key_id=10`, 4 models), and `gpt-5.4-mini` `/v1/chat/completions` 200.
- Production API keys were normalized to the OpenAI-style `sk-` prefix at 2026-05-18 18:47 +08:00: 16 existing `api_keys.key` values were updated (`active=13`, `disabled=3`), with no collisions. Server-side secret backup: `/data/fuxi-api/prod/reports/api-key-prefix-backup-20260518-184747.csv`.
- After API key prefix normalization, Redis auth cache was cleared and `fuxi-api-prod` restarted healthy. Verified `key_id=10` with prefixed key: `/v1/models` 200 and `gpt-5.4-mini` chat 200; old unprefixed form returns 401 `INVALID_API_KEY`.
- GET `https://fuxiapi.top/api/status`: 200. Use GET for this endpoint; HEAD is not registered by Gin for the route.
- GET `https://fuxiapi.top/health`: 200.
- Old `/data/new-api`, `new-api`, and `new-api-staging` were reconfirmed present after the follow-up deploy and remain intentionally preserved.
- Local release candidate `0.1.127` prepared at 2026-05-18 19:52 +08:00 with admin account-table resizable columns. This is repository-side release prep only; production has not yet been updated to this version in the current session.
- Admin account table now supports drag-resizing headers, fixed persisted column widths via `localStorage` key `account-column-widths`, and protected selection column width.
- Legacy `F:\newAPI` AI architecture was rechecked on 2026-05-18. Current project now has the migrated governance pieces that were missing: `tools/check_ai_docs.py`, `tools/MANIFEST.md`, `.ai/assistant_guide.md`, router `last_verified` schema, generated runtime/AI-doc snapshots, and explicit release-closure semantics.
- Active meaning of `更新发布版并归档`: bump semantic version patch by default, run relevant local verification, commit/push release-prep when ready, update `.ai/MEMORY.md` + `.ai/sessions.md` + `.ai/archive/sessions/`, and deploy staging/production only when that environment update is explicitly confirmed.
- Root agent entries `AGENTS.md` and `CLAUDE.md` now point to `.ai/assistant_guide.md`, `tools/check_ai_docs.py`, and the release-closure meaning of `更新发布版并归档`.

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
- `a0177bcb Document post-cutover release sync`
- `33c5e36c Record production release sync verification`
- `0da8c375 Record production update to latest release`
- `321f1f31 Record production API key prefix normalization`

## Current Risks

- `plus` has an active key but zero user balance, so it returns 403 until balance or group policy changes.
- Some old consumption logs are archived only because old token/channel rows did not map to target rows.
- Legacy custom OAuth bindings were archived only; do not assume they are active in the new schema.
- `security.url_allowlist.enabled=false` remains a runtime warning inherited from current config; evaluate separately before tightening upstream URL policy.
- Production `channels` is empty after migration; available-channel UI currently relies on the account-pool fallback from `accounts` and `groups`.
- Production `channel_monitors` is empty; channel status UI displays an "unmonitored" account-pool fallback until real monitor tasks are configured.
- Existing clients must update API keys to include the new `sk-` prefix; unprefixed API keys now intentionally fail authentication.
- GitHub CI for `33c5e36c` recorded a failed backend `test` job with only `exit code 2` visible via public annotations; local `go test ./...`, `make test-unit`, frontend `typecheck`, and frontend `lint:check` passed before publishing, and GHCR Image/Security Scan were successful.
- Deleting or overwriting old rollback resources still requires a future project rule change; current rules prohibit it.
- `0.1.127` is prepared locally but not yet tagged, pushed, or deployed to production from this session.
- AI docs health check passed after migration: `python tools/check_ai_docs.py --summary --details` reported 0 failures and 0 warnings; context recall passed for `backend/cmd/server/VERSION`, `frontend/src/views/admin/AccountsView.vue`, and `deploy/fuxi/deploy-staging.sh`.
