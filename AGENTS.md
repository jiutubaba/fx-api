# AGENTS.md - Project Conventions for Fuxi API

## Overview

This repository is the Fuxi API migration to the sub2api architecture. The active remote is:

```text
https://github.com/jiutubaba/fx-api.git
```

The Go module is `github.com/jiutubaba/fx-api`. The production public URL is `https://fuxiapi.top`; staging is `https://staging.fuxiapi.top`.

## AI Docs

Project-local AI working docs live under `.ai/`.

- Start with `.ai/README.md`.
- Path routing source: `.ai/router.md`.
- Current hot state: `.ai/MEMORY.md`.
- Recent session archive: `.ai/sessions.md` and `.ai/archive/sessions/`.

The old `F:\newAPI` AI documentation was migrated as a legacy reference model only. It is not authoritative for this repository's architecture.

## Tech Stack

- Backend: Go, Gin, Ent, PostgreSQL, Redis.
- Frontend: Vue 3, TypeScript, Vite, Pinia.
- Frontend package manager: pnpm.
- Deployment: Docker Compose under `deploy/fuxi/`, GHCR image `ghcr.io/jiutubaba/fx-api`.

## Architecture

```text
backend/
  cmd/server/              - application entrypoint
  cmd/migrate-newapi/      - old new-api to sub2api migration tool
  ent/schema/              - Ent schema definitions
  migrations/              - SQL migrations
  internal/handler/        - HTTP handlers
  internal/service/        - business logic and upstream gateway logic
  internal/repository/     - database access layer
  internal/server/         - server wiring, routes, middleware
frontend/                  - Vue admin and user console
deploy/fuxi/               - Fuxi staging/prod deployment and cutover scripts
```

## Production Safety Rules

1. Do not delete or overwrite `F:\newAPI`, `/data/new-api/**`, `new-api`, or `new-api-staging`.
2. Production Caddy cutover requires explicit user confirmation.
3. Staging may be updated directly when the user clearly asks for staging.
4. Never write secrets, DSNs, OAuth secrets, JWT secrets, API keys, or password hashes into Git.
5. Migration reports may record counts, paths, and warnings, but not credentials.
6. Keep sub2api LGPLv3 attribution and upstream source references intact.

## Verification

Backend:

```powershell
go test ./...
```

Frontend:

```powershell
npx pnpm --dir frontend run typecheck
npx pnpm --dir frontend run lint:check
```

Deployment verification must include `/api/status`, authenticated `/v1/models`, and at least one real `/v1/chat/completions` path for active groups.

