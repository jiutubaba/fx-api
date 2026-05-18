---
name: update-server
description: Update the live Fuxi API sub2api production server from https://github.com/jiutubaba/fx-api.git. Use when the user explicitly confirms production update, production deploy, or server update.
---

# Update Server

## Trigger Rules

- Production is confirmation-gated.
- Do not run this workflow for discussion, planning, or staging.
- Do not delete old `new-api` resources.

## Production Facts

- Server: `38.12.6.32`
- SSH: `ssh -i C:\Users\Administrator\.ssh\id_ed25519_fx_api root@38.12.6.32`
- Runtime root: `/data/fuxi-api/prod`
- Deploy scripts: `/data/fuxi-api/deploy`
- Container: `fuxi-api-prod`
- Redis: `fuxi-api-prod-redis`
- Image: `ghcr.io/jiutubaba/fx-api:latest`
- Public URL: `https://fuxiapi.top`
- Public Caddy target: `127.0.0.1:3300`
- Old rollback target: `127.0.0.1:3000`

## Workflow

1. Confirm local `main` is pushed to `https://github.com/jiutubaba/fx-api.git`.
2. Confirm GitHub CI, Security Scan, and GHCR Image are successful for the target commit.
3. On server:

```bash
cd /data/fuxi-api/prod
docker compose --env-file .env -f docker-compose.yml pull fuxi-api
docker compose --env-file .env -f docker-compose.yml up -d fuxi-api
```

4. Verify:

```bash
curl -fsS http://127.0.0.1:3300/api/status
curl -fsS https://fuxiapi.top/api/status
docker exec fuxi-api-prod-redis redis-cli INFO persistence
```

5. Verify authenticated `/v1/models` and at least one real `/v1/chat/completions` request.

## Rollback

If Caddy target must roll back to old new-api:

```bash
CONFIRM_SWITCH=fuxiapi.top NEW_TARGET=127.0.0.1:3000 /data/fuxi-api/deploy/switch-prod-caddy.sh
```

