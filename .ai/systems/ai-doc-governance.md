# AI Doc Governance

Last verified: 2026-05-18

- `.ai/` is a collaboration layer, not source code truth.
- Stable project facts belong in `.ai/systems/*.md`.
- Generated snapshots belong in `.ai/generated/*.md`. They are useful for context recall but are never the only source of truth.
- Current operational state belongs in `.ai/MEMORY.md`.
- Recent high-value session summaries belong in `.ai/sessions.md`.
- Long historical records belong in `.ai/archive/`.
- Legacy old-project facts must be marked as legacy and must not override active sub2api rules.
- `tools/check_ai_docs.py` is the health checker for this layer.

Before editing `.ai/`, check:

1. Does the fact come from current code, deploy scripts, server output, or a migration report?
2. Is the fact stable enough for a system doc, or only hot enough for memory?
3. Could the note leak secrets or credentials?

## Read Order

1. `.ai/router.md`
2. Relevant `.ai/systems/*.md`
3. `.ai/MEMORY.md`

Use `.ai/assistant_guide.md` when the task involves release closure, archiving, or cross-session handoff.

## Router Schema

`.ai/router.md` must keep these columns:

```text
scope/path_glob | status | must_read | related | keywords | last_verified
```

Valid status values are `active`, `stale`, `archived`, and `legacy`.

## Three-Pass Review

1. Scope review: only allowed and relevant paths changed.
2. Fact review: each stable statement traces to current code, scripts, server output, or existing archive.
3. Collaboration review: no sensitive data, no accidental deletion of another agent's notes, no legacy rule promoted to active behavior.

## Health Check

Run after AI documentation changes:

```powershell
python tools/check_ai_docs.py --summary --details
```

Use context recall sampling for high-risk paths:

```powershell
python tools/check_ai_docs.py --context backend/cmd/server/VERSION --details
python tools/check_ai_docs.py --context deploy/fuxi/deploy-staging.sh --details
python tools/check_ai_docs.py --context frontend/src/views/admin/AccountsView.vue --details
```
