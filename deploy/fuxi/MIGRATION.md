# newAPI 数据迁移

迁移工具位于镜像内 `/app/migrate-newapi`，本地源码入口为 `backend/cmd/migrate-newapi`。

## 参数

```bash
/app/migrate-newapi \
  --source-dsn "$NEWAPI_SOURCE_DSN" \
  --target-dsn "$FUXI_TARGET_DSN" \
  --source-name staging \
  --mode dry-run \
  --report-dir /app/reports
```

- `--source-name`: `prod` 或 `staging`
- `--mode`: `dry-run` 只生成报告；`apply` 会先执行 sub2api migrations，再写入 `legacy_newapi` 和目标正式表
- `--report-dir`: 输出 JSON 与 Markdown 报告

## 执行顺序

1. 对生产库执行 `dry-run`，生成 shadow 迁移前报告。
2. 对生产库迁移到 shadow 目标库，核对用户数、余额汇总、Key 数、账号数和日志数。
3. 对预发源库迁移到新预发库，启动 `fuxi-api-staging`。
4. 预发验收通过后，重新从生产库 fresh migration 到新生产目标库。
5. 人工确认后再切生产 Caddy。

旧表会完整导入目标库 `legacy_newapi` schema；不等价或高风险数据仍以归档为准，例如旧管理日志、系统日志、错误日志、Passkey/TOTP 格式不兼容数据。
