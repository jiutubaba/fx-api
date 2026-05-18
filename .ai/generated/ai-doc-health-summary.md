# AI Doc Health Summary

Generated: 2026-05-18

## Structure

- Entry: `.ai/README.md`
- Router: `.ai/router.md`
- Shared rules: `.ai/shared_instructions.md`
- Assistant execution guide: `.ai/assistant_guide.md`
- System docs: `.ai/systems/*.md`
- Generated snapshots: `.ai/generated/*.md`
- Hot memory: `.ai/MEMORY.md`
- Recent sessions: `.ai/sessions.md`
- Session archive: `.ai/archive/sessions/2026-05.md`
- Legacy new-api summary: `.ai/archive/legacy-newapi/README.md`

## Migrated From Legacy AI Architecture

- Thin entry docs.
- Path router with `last_verified`.
- System docs for stable facts.
- Generated snapshots for rebuildable summaries.
- Hot memory plus recent sessions plus archive split.
- Health checker at `tools/check_ai_docs.py`.

## Active Rules

- `.ai/` is a collaboration layer, not a source of truth.
- `generated/` snapshots can be rebuilt and must not contain unique facts or secrets.
- `MEMORY.md` should stay hot-state focused.
- `sessions.md` keeps recent high-value sessions; detailed history goes to `.ai/archive/sessions/`.
- Legacy `F:\newAPI` facts must stay marked as legacy and must not override active sub2api paths.

## Check Command

```powershell
python tools/check_ai_docs.py --summary --details
```
