# Recent Sessions

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
