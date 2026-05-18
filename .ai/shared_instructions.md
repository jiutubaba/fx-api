# Shared Instructions

## Read Order

1. `router`: `.ai/router.md`
2. `systems`: relevant `.ai/systems/*.md`
3. `memory`: `.ai/MEMORY.md`

## Write Boundaries

- Do not store secrets, DSNs, OAuth secrets, JWT secrets, API keys, passwords, password hashes, or private account data in `.ai/`.
- `MEMORY.md` records current operational state and risks only.
- `sessions.md` records recent high-value session summaries.
- `archive/` stores historical records and legacy summaries.
- Generated or reconstructed facts must identify their source when possible.

## Multi-Agent Rules

- Assume other agents or the user may have changed the workspace.
- Do not revert unrelated changes.
- Before production changes, verify Caddy, containers, database target, and rollback path.
- For old `F:\newAPI` content, archive facts without making old project rules authoritative in this repository.

