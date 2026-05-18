# Code Style

## Backend

- Follow existing Go package structure under `backend/internal`.
- Prefer existing service/repository helpers before adding abstractions.
- Keep SQL migrations deterministic and PostgreSQL-focused unless the current code path requires otherwise.
- Do not log secrets, tokens, credentials, password hashes, or DSNs.

## Frontend

- Use Vue 3, TypeScript, Vite, Pinia, and existing component patterns.
- Use `pnpm` scripts from `frontend/package.json`.
- Keep UI copy consistent with Fuxi API branding and Chinese production usage.

## Deploy

- Use `deploy/fuxi/` scripts as the active deployment interface.
- Keep staging and production runtime directories separate.
- Keep rollback commands explicit and tested.

