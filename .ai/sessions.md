# Recent Sessions

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

