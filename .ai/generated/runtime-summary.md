# Runtime Summary

Generated: 2026-05-18

This file is a rebuildable snapshot. Source code, deploy scripts, and server output remain authoritative.

## Repository

- Root: `F:\伏羲API`
- Remote: `https://github.com/jiutubaba/fx-api.git`
- Go module: `github.com/jiutubaba/fx-api`
- Go version declaration: `go 1.26.3`
- Embedded version source: `backend/cmd/server/VERSION`
- Current local embedded version: `0.1.127`

## Stack

- Backend: Go, Gin, Ent, PostgreSQL, Redis.
- Frontend: Vue 3, TypeScript, Vite, Pinia.
- Frontend package manager: pnpm.
- Runtime image: `ghcr.io/jiutubaba/fx-api`.

## Release

- Release workflow: `.github/workflows/release.yml`
- Release trigger: `v*` tag push or manual `workflow_dispatch`.
- Release workflow syncs `backend/cmd/server/VERSION` from the tag version.
- Production deploy interface: `.agents/skills/update-server/`
- Staging deploy interface: `.agents/skills/update-staging-server/`

## Environments

- Production URL: `https://fuxiapi.top`
- Staging URL: `https://staging.fuxiapi.top`
- Server: `38.12.6.32`
- Production runtime root: `/data/fuxi-api/prod`
- Staging runtime root: `/data/fuxi-api/staging`
- Production container: `fuxi-api-prod`
- Staging container: `fuxi-api-staging`
- PostgreSQL container: `new-api-postgres`

## Legacy Boundary

- `F:\newAPI`, `/data/new-api/**`, `new-api`, and `new-api-staging` are preserved rollback/legacy resources.
- Old GORM/React/Bun/new-api deployment rules are not active for this repository.
