# Release And Environments

Last verified: 2026-05-18

## Active Runtime

- Server: `38.12.6.32`
- SSH: `ssh -i C:\Users\Administrator\.ssh\id_ed25519_fx_api root@38.12.6.32`
- Production root: `/data/fuxi-api/prod`
- Staging root: `/data/fuxi-api/staging`
- Deploy scripts: `/data/fuxi-api/deploy`
- Production container: `fuxi-api-prod`
- Staging container: `fuxi-api-staging`
- Production Redis: `fuxi-api-prod-redis`
- Staging Redis: `fuxi-api-staging-redis`
- PostgreSQL container: `new-api-postgres`

## Ports

- Production public Caddy target: `127.0.0.1:3300`
- Staging public Caddy target: `127.0.0.1:3200`
- Old production retained: `127.0.0.1:3000`
- Old staging retained: `127.0.0.1:3100`

## Scripts

- Staging deploy: `deploy/fuxi/deploy-staging.sh`
- Production candidate prepare: `deploy/fuxi/prepare-prod.sh`
- Migration: `deploy/fuxi/migrate.sh`
- Production Caddy switch: `deploy/fuxi/switch-prod-caddy.sh`

## Version Source

- Embedded application version source: `backend/cmd/server/VERSION`.
- Current release workflow also syncs `backend/cmd/server/VERSION` from the tag version when GitHub Release runs.
- Default version bump policy: increment patch unless the user explicitly asks for minor or major.
- Active release tags use `vX.Y.Z`; the embedded `VERSION` file stores `X.Y.Z` without the `v` prefix.
- If `backend/cmd/server/VERSION`, `.github/workflows/release.yml`, Dockerfile release settings, or deploy scripts change, verify the frontend build artifact and Go build version path together.

## Meaning of "更新发布版" and "更新发布版并归档"

These phrases are full production release requests. They are not only version-file edits.

Default scope:

1. Inspect current worktree and confirm the target changes are included.
2. Bump the patch version in `backend/cmd/server/VERSION` unless the user asks for another semantic version level or the requested release already exists.
3. Run local verification appropriate to the touched areas:
   - Backend: `go test ./...`
   - Frontend: `npx pnpm --dir frontend run typecheck`
   - Frontend lint: `npx pnpm --dir frontend run lint:check`
   - Diff hygiene: `git diff --check`
   - AI docs if `.ai/` changed: `python tools/check_ai_docs.py --summary --details`
4. Commit and push release changes to Git.
5. Publish or update the GitHub tag/Release.
6. Deploy production so `https://fuxiapi.top/` is actually updated.
7. Verify production status, Redis persistence/write state, authenticated `/v1/models`, and at least one real `/v1/chat/completions` request.
8. Update `.ai/MEMORY.md`, `.ai/sessions.md`, and `.ai/archive/sessions/` with the version, scope, verification, deployment status, and known risks.

`更新发布版` and `更新发布版并归档` are explicit confirmation for a normal production app update. Production Caddy cutover/rollback target changes and deletion of preserved legacy resources still require separate explicit confirmation.

## Staging and Production Rules

- Staging may be updated when the user clearly asks for staging.
- Production app update may proceed when the user says `更新发布版`, `更新发布版并归档`, or otherwise explicitly asks to update production.
- Before production update, confirm local branch is pushed, CI/GHCR target commit is successful or record the risk, and rollback remains available.
- Never delete old `new-api`, `new-api-staging`, `/data/new-api/**`, or local `F:\newAPI` while these are protected rollback/legacy resources.

Production Caddy switch requires:

```bash
CONFIRM_SWITCH=fuxiapi.top
```

## Verification

- `https://fuxiapi.top/api/status`
- `https://staging.fuxiapi.top/api/status`
- Authenticated `/v1/models`
- `/v1/chat/completions` for active production groups
- Redis persistence and write status
- Caddy target lines
