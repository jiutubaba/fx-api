# 伏羲API部署门禁

本目录用于 sub2api 架构迁移后的伏羲API部署。默认只新增 `/data/fuxi-api/**`，不删除旧 `new-api`、`new-api-staging` 容器，也不修改 `/data/new-api/**`。

## 预发

1. 在服务器创建 `/data/fuxi-api/staging/.env`，可从 `env.staging.example` 复制后填入真实 DSN、Redis、JWT、TOTP 等密钥。
2. 先迁移预发源库到新的预发目标库：

```bash
sudo APP_ROOT=/data/fuxi-api/staging SOURCE_NAME=staging MODE=dry-run ./migrate.sh
sudo APP_ROOT=/data/fuxi-api/staging SOURCE_NAME=staging MODE=apply CONFIRM_APPLY=staging ./migrate.sh
```

3. 启动新预发容器：

```bash
sudo IMAGE=ghcr.io/jiutubaba/fx-api:latest ./deploy-staging.sh
```

预发容器名为 `fuxi-api-staging`，宿主机只绑定 `127.0.0.1:3200`。

4. 本机验收通过后再切 `staging.fuxiapi.top`：

```bash
sudo ./verify-staging.sh
sudo CONFIRM_SWITCH=staging.fuxiapi.top ./switch-staging-caddy.sh
sudo CHECK_PUBLIC=true ./verify-staging.sh
```

## 生产准备

1. 在服务器创建 `/data/fuxi-api/prod/.env`，可从 `env.prod.example` 复制后填入真实配置。
2. 执行：

```bash
sudo IMAGE=ghcr.io/jiutubaba/fx-api:latest ./prepare-prod.sh
```

生产准备容器名为 `fuxi-api-prod`，默认只绑定 `127.0.0.1:3300`，不会切换公网流量。

## 生产切换

只有预发验收通过、并重新从生产库 fresh migration 到新生产库后，才允许执行：

```bash
sudo CONFIRM_SWITCH=fuxiapi.top ./switch-prod-caddy.sh
```

脚本会先备份 Caddyfile，再把 `fuxiapi.top` 反代目标改为 `127.0.0.1:3300` 并 reload。失败回滚时把 Caddyfile 恢复到备份，或直接将目标改回旧 `127.0.0.1:3000`。

## 服务器约束

- 旧 `new-api`、`new-api-staging` 容器和 `/data/new-api/**` 不删除。
- `switch-staging-caddy.sh` 只处理 `staging.fuxiapi.top`，不会修改 `fuxiapi.top`。
- `switch-prod-caddy.sh` 必须显式设置 `CONFIRM_SWITCH=fuxiapi.top`。
- `.env` 中的 DSN、OAuth secret、支付 secret 只保留在服务器，不提交到 Git。
