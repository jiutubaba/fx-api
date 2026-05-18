# AI Context Router

> Path-level source for AI context recall. Keep this table compatible with `tools/check_ai_docs.py`.

| scope/path_glob | status | must_read | related | keywords | last_verified |
|---|---|---|---|---|---|
| `AGENTS.md` | active | `.ai/README.md` | `.ai/shared_instructions.md` `.ai/assistant_guide.md` | `入口` `项目规则` | 2026-05-18 |
| `CLAUDE.md` | active | `.ai/README.md` | `.ai/shared_instructions.md` `.ai/assistant_guide.md` | `入口` | 2026-05-18 |
| `.ai/**` | active | `.ai/README.md` | `.ai/systems/ai-doc-governance.md` `.ai/generated/ai-doc-health-summary.md` | `AI文档` `归档` | 2026-05-18 |
| `tools/check_ai_docs.py` | active | `.ai/systems/ai-doc-governance.md` | `tools/MANIFEST.md` `.ai/generated/ai-doc-health-summary.md` | `AI文档检查` | 2026-05-18 |
| `.agents/skills/**` | active | `.ai/systems/release-and-environments.md` | `.ai/generated/skills-index.md` `.ai/MEMORY.md` | `agent技能` `发布` | 2026-05-18 |
| `backend/cmd/server/VERSION` | active | `.ai/systems/release-and-environments.md` | `.ai/generated/runtime-summary.md` `.ai/MEMORY.md` | `版本` `发布版` | 2026-05-18 |
| `.github/workflows/**` | active | `.ai/systems/release-and-environments.md` | `.ai/generated/runtime-summary.md` `.ai/MEMORY.md` | `CI` `GHCR` `release` | 2026-05-18 |
| `backend/cmd/migrate-newapi/**` | active | `.ai/systems/migration-newapi.md` | `.ai/systems/release-and-environments.md` `.ai/archive/legacy-newapi/README.md` | `迁移` `legacy_newapi` | 2026-05-18 |
| `backend/ent/schema/**` | active | `.ai/project_guide.md` | `.ai/rules.md` | `Ent` `schema` | 2026-05-18 |
| `backend/migrations/**` | active | `.ai/project_guide.md` | `.ai/rules.md` | `SQL迁移` | 2026-05-18 |
| `backend/internal/handler/**` | active | `.ai/systems/relay-and-upstream.md` | `.ai/project_guide.md` | `handler` `API` | 2026-05-18 |
| `backend/internal/service/**` | active | `.ai/systems/relay-and-upstream.md` | `.ai/project_guide.md` | `service` `调度` `上游` | 2026-05-18 |
| `backend/internal/repository/**` | active | `.ai/project_guide.md` | `.ai/rules.md` | `repository` `数据库` | 2026-05-18 |
| `backend/internal/server/**` | active | `.ai/project_guide.md` | `.ai/systems/release-and-environments.md` | `server` `routes` | 2026-05-18 |
| `frontend/**` | active | `.ai/systems/frontend-vue.md` | `.ai/code_style.md` `.ai/generated/runtime-summary.md` | `Vue` `Vite` `pnpm` | 2026-05-18 |
| `deploy/fuxi/**` | active | `.ai/systems/release-and-environments.md` | `.ai/systems/migration-newapi.md` `.ai/MEMORY.md` | `部署` `Caddy` `生产` `预发` | 2026-05-18 |
| `F:\newAPI` | legacy | `.ai/archive/legacy-newapi/README.md` | `.ai/systems/migration-newapi.md` | `旧项目` `只读` `回滚来源` | 2026-05-18 |
