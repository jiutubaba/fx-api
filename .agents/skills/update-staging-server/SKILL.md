---
name: update-staging-server
description: Update the Fuxi API staging environment at https://staging.fuxiapi.top.
---

# Update Staging Server

## Trigger Rules

- If the user clearly asks to update staging, execute directly.
- Do not use this skill for production.

## Staging Facts

- Server: `38.12.6.32`
- SSH: `ssh -i C:\Users\Administrator\.ssh\id_ed25519_fx_api root@38.12.6.32`
- Runtime root: `/data/fuxi-api/staging`
- Deploy scripts: `/data/fuxi-api/deploy`
- Container: `fuxi-api-staging`
- Redis: `fuxi-api-staging-redis`
- Public URL: `https://staging.fuxiapi.top`
- Public Caddy target: `127.0.0.1:3200`

## Workflow

```bash
/data/fuxi-api/deploy/deploy-staging.sh
```

Verify:

```bash
curl -fsS http://127.0.0.1:3200/api/status
curl -fsS https://staging.fuxiapi.top/api/status
docker exec fuxi-api-staging-redis redis-cli INFO persistence
```

Staging updates must not remove or restart old production `new-api`.

