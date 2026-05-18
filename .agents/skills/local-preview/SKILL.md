---
name: local-preview
description: Start or inspect local Fuxi API preview processes for backend and frontend development.
---

# Local Preview

## Purpose

Use local preview for uncommitted UI or API changes before staging.

## Commands

Backend tests:

```powershell
go test ./...
```

Frontend checks:

```powershell
npx pnpm --dir frontend run typecheck
npx pnpm --dir frontend run lint:check
```

Frontend dev server:

```powershell
npx pnpm --dir frontend run dev
```

Backend local runtime depends on local configuration and database availability. Do not invent local secrets; use existing local env/config only.

