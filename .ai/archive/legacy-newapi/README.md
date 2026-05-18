# Legacy newAPI AI Architecture Summary

Source: `F:\newAPI`

The old project contained an AI collaboration structure with:

- Root agent entry files: `AGENTS.md`, `CLAUDE.md`
- Shared AI docs: `.ai/README.md`, `.ai/router.md`, `.ai/MEMORY.md`, `.ai/sessions.md`, `.ai/archive/**`
- Agent skills: `.agents/skills/update-server`, `.agents/skills/update-staging-server`, `.agents/skills/local-preview`
- Generated summaries and system docs under `.ai/generated/` and `.ai/systems/`

Only the structure and workflow pattern were migrated. The following old assumptions are legacy-only and must not guide active work in this repository:

- GORM model layer and old `router/controller/service/model` paths
- React default frontend, Bun workflow, and `web/default/**`
- Old `/data/new-api/**` deployment scripts as active release commands
- Old `new-api` and `new-api-staging` containers as primary runtime

The active Fuxi API project now uses sub2api paths, Vue frontend, Ent/PostgreSQL schema, `deploy/fuxi/`, and GHCR image deployment.

