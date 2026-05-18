# Project Guide

Fuxi API now runs on the sub2api architecture.

## Repository Facts

- Local workspace: `F:\伏羲API`
- Remote: `https://github.com/jiutubaba/fx-api.git`
- Upstream reference: `https://github.com/Wei-Shaw/sub2api`
- Go module: `github.com/jiutubaba/fx-api`
- Docker image: `ghcr.io/jiutubaba/fx-api`

## Runtime Facts

- Server: `38.12.6.32`
- SSH: `ssh -i C:\Users\Administrator\.ssh\id_ed25519_fx_api root@38.12.6.32`
- Production URL: `https://fuxiapi.top`
- Staging URL: `https://staging.fuxiapi.top`
- New production app: `fuxi-api-prod`, bound to `127.0.0.1:3300`
- New staging app: `fuxi-api-staging`, bound to `127.0.0.1:3200`
- PostgreSQL container: `new-api-postgres`
- PostgreSQL network: `new-api-net`

## Preserved Legacy Runtime

- Old production app: `new-api`, bound to `127.0.0.1:3000`
- Old staging app: `new-api-staging`, bound to `127.0.0.1:3100`
- Old runtime directory: `/data/new-api/**`
- Old local source: `F:\newAPI`

These are backup and rollback sources. Do not delete them during normal sub2api work.

