# Release And Environments

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

