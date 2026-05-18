# Skills Index

Generated: 2026-05-18

Migrated and adapted from the old `F:\newAPI` AI architecture. This is a rebuildable snapshot.

| skill | status | purpose |
|---|---|---|
| `.agents/skills/update-server` | active | Confirmation-gated production update for new sub2api runtime. |
| `.agents/skills/update-staging-server` | active | Staging update for `https://staging.fuxiapi.top`. |
| `.agents/skills/local-preview` | active | Local validation and preview guidance for backend/frontend work. |

## Release Semantics

- `update-staging-server`: may execute when the user clearly asks for staging.
- `update-server`: production confirmation-gated.
- `更新发布版并归档`: version closure and archive workflow; production deploy only if explicitly confirmed.
