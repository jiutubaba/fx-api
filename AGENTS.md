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
- Execution and release-closure guide: `.ai/assistant_guide.md`.
- Current hot state: `.ai/MEMORY.md`.
- Recent session archive: `.ai/sessions.md` and `.ai/archive/sessions/`.
- AI docs health check: `python tools/check_ai_docs.py --summary --details`.

The old `F:\newAPI` AI documentation was migrated as a legacy reference model only. It is not authoritative for this repository's architecture.

When the user says "更新发布版并归档", treat it as a release-version closure workflow:

1. Bump the patch version by default in `backend/cmd/server/VERSION`.
2. Run relevant verification for touched areas.
3. Commit and push release-prep changes when ready.
4. Update `.ai/MEMORY.md`, `.ai/sessions.md`, and `.ai/archive/sessions/`.
5. Deploy staging or production only when that environment update is explicitly confirmed.

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

AI docs:

```powershell
python tools/check_ai_docs.py --summary --details
```

Deployment verification must include `/api/status`, authenticated `/v1/models`, and at least one real `/v1/chat/completions` path for active groups.
