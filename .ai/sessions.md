# Recent Sessions

## 2026-05-18 - Legacy AI Architecture Compatibility Migration

- Rechecked old `F:\newAPI` AI architecture: root `AGENTS.md`/`CLAUDE.md`, `.ai` router/systems/generated/session layers, `.agents` release skills, and `tools/check_ai_docs.py`.
- Confirmed the current project had the high-level `.ai` skeleton but was not fully compatible with the old governance model.
- Migrated the missing reusable governance pieces into the current sub2api project:
  - `tools/check_ai_docs.py` and `tools/MANIFEST.md`
  - `.ai/assistant_guide.md`
  - `.ai/router.md` `last_verified` schema
  - `.ai/generated/runtime-summary.md`
  - `.ai/generated/ai-doc-health-summary.md`
- Updated release/environment docs so `更新发布版并归档` means version closure: version bump, verification, release-prep commit/push when ready, `.ai` archive updates, and environment deployment only when explicitly confirmed.
- Kept old new-api assumptions legacy-only: GORM, React/Rsbuild/Bun, old `/data/new-api/**` scripts, and old `new-api` containers do not override current Fuxi API rules.
- Release-prep commit pushed to `origin/main`: `41d81ff1 Prepare 0.1.127 account table release`.

See `.ai/archive/sessions/2026-05.md` for the detailed archive.

## 2026-05-18 - Account Table Column Resize Release Prep

- Added draggable header resize handles to the shared frontend `DataTable` component.
- Added account-table default/min/max column widths and persisted admin custom widths under `account-column-widths`.
- Kept the select checkbox column fixed while allowing the remaining visible columns to be resized.
- Verified frontend `typecheck`, frontend `lint:check`, and `git diff --check`.
- Bumped the embedded backend release version from `0.1.126` to `0.1.127` for the next release candidate.
- Release-prep commit pushed to `origin/main`: `41d81ff1 Prepare 0.1.127 account table release`.
- Confirmed the local preview can proxy to staging at `http://localhost:3000/`, with `/api/status` returning 200.
- Confirmed 429 / `7d 100%` exhausted accounts are rate-limited and temporarily excluded from scheduling until reset, not directly disabled.

See `.ai/archive/sessions/2026-05.md` for the detailed archive.

## 2026-05-18 - Post-cutover Cleanup and Release Sync

- Reconfirmed local `main` matches `origin/main` before cleanup.
- Classified old `new-api`, `new-api-staging`, `/data/new-api/**`, and `F:\newAPI` as protected rollback/backup resources under project rules, not safe deletion targets.
- Verified production Caddy remains on the new Fuxi API target `127.0.0.1:3300`.
- Verified production feature flags: `available_channels_enabled=true`, `channel_monitor_enabled=true`, `channel_monitor_default_interval_seconds=60`.
- Re-ran local backend tests and frontend typecheck/lint before release sync.
- Prepared release sync around the current GHCR production image and a follow-up documentation commit.
- Updated production to GHCR image revision `33c5e36c31a1b4f8686526ff15fea934565f9982` / image `sha256:d2dbb784f80a563d13747bf7c9813e014e7e07d447d31e4b92ed3f175f46d567`.
- Verified `fuxi-api-prod` healthy with 0 restarts, Redis persistence OK, local/public `/api/status` 200, homepage 200, `/available-channels` 200, `/monitor` 200, authenticated `/v1/models` 200, and a real `gpt-5.4-mini` chat completion 200.
- Noted GitHub CI for `33c5e36c` had a backend `test` job failure with only `exit code 2` exposed publicly; local `go test ./...`, `make test-unit`, frontend typecheck, and frontend lint passed, while GHCR Image and Security Scan succeeded.
- Normalized production API keys to OpenAI-style `sk-` prefix: updated 16 existing keys (`active=13`, `disabled=3`), saved a root-only server backup at `/data/fuxi-api/prod/reports/api-key-prefix-backup-20260518-184747.csv`, cleared Redis auth cache, restarted `fuxi-api-prod`, and verified prefixed `key_id=10` works while the old unprefixed form returns 401.

## 2026-05-18 - Account and Channel Visibility Follow-up

- Added an old-repo-style account statistics strip to the admin account management page.
- Confirmed production settings were not disabled: available channels and channel monitor feature flags are enabled.
- Found the display gap was data-shape related: production `channels` has 0 rows while migrated runtime resources live in `accounts` and `account_groups`.
- Added available-channel account-pool fallback and channel-status "unmonitored" fallback without fabricating model support, latency, or availability data.
- Verified local backend unit tests plus frontend typecheck and lint.
- Confirmed GitHub Actions and GHCR image build passed, then redeployed `fuxi-api-prod`.
- Verified production `/api/status`, `/health`, homepage, container health, and old new-api rollback resources.

See `.ai/archive/sessions/2026-05.md` for the detailed archive.

## 2026-05-18 - sub2api Production Cutover

- Migrated from the old new-api runtime to the new Fuxi API sub2api architecture.
- Confirmed repository remote: `https://github.com/jiutubaba/fx-api.git`.
- Confirmed old local and server resources are preserved: `F:\newAPI`, `/data/new-api/**`, `new-api`, `new-api-staging`.
- Fixed migration of legacy account-pool and auto group bindings.
- Fixed Redis volume ownership for Fuxi deploy scripts.
- Added runtime proxy and CORS defaults for Caddy-backed production/staging.
- Fresh-migrated production data into `fuxi_api_prod`.
- Switched production Caddy to `127.0.0.1:3300` after explicit user confirmation.
- Verified production status, frontend, authenticated models, and `free`/`auto` chat completions.

See `.ai/archive/sessions/2026-05.md` for the detailed archive.
