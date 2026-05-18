# Skills Index

Generated: 2026-05-18

Migrated and adapted from the old `F:\newAPI` AI architecture. This is a rebuildable snapshot.

| skill | status | purpose |
|---|---|---|
| `.agents/skills/update-server` | active | Production update for the new sub2api runtime; use when the user says `更新发布版`, `更新发布版并归档`, or explicitly asks for production deploy. |
| `.agents/skills/update-staging-server` | active | Staging update for `https://staging.fuxiapi.top`. |
| `.agents/skills/local-preview` | active | Local validation and preview guidance for backend/frontend work. |

## Release Semantics

- `update-staging-server`: may execute when the user clearly asks for staging.
- `update-server`: production app update for confirmed release/deploy requests.
- `更新发布版` / `更新发布版并归档`: commit/push, GitHub tag/Release, production deploy to `https://fuxiapi.top/`, verification, and archive.
