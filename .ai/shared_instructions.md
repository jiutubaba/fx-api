# Shared Instructions

## Read Order

1. `router`: `.ai/router.md`
2. `systems`: relevant `.ai/systems/*.md`
3. `memory`: `.ai/MEMORY.md`
4. `assistant_guide`: `.ai/assistant_guide.md` for release closure or archival work.

## Write Boundaries

- Do not store secrets, DSNs, OAuth secrets, JWT secrets, API keys, passwords, password hashes, or private account data in `.ai/`.
- `MEMORY.md` records current operational state and risks only.
- `sessions.md` records recent high-value session summaries.
- `archive/` stores historical records and legacy summaries.
- `generated/` stores rebuildable snapshots only.
- Generated or reconstructed facts must identify their source when possible.

## Multi-Agent Rules

- Assume other agents or the user may have changed the workspace.
- Do not revert unrelated changes.
- Before production changes, verify Caddy, containers, database target, and rollback path.
- For old `F:\newAPI` content, archive facts without making old project rules authoritative in this repository.

## Release Closure

- `更新发布版并归档` means version bump, relevant verification, release-prep commit/push when ready, and `.ai` archive updates.
- It does not override production confirmation gates.
- Record whether the result is `已归档，未发布`, `已准备发布候选`, `已发布预发`, or `已发布生产`.
