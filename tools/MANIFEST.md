# Tools Manifest

## check_ai_docs.py

Purpose: check project-local `.ai/` documentation health.

It verifies:

- `.ai/router.md` schema and referenced documents.
- Context-route recall for a requested path.
- Stale active routes and top-level `.ai/*.md` files.
- `MEMORY.md` size.
- `sessions.md` size.
- Presence of `.ai/generated/` snapshots.

Commands:

```powershell
python tools/check_ai_docs.py
python tools/check_ai_docs.py --summary
python tools/check_ai_docs.py --details
python tools/check_ai_docs.py --context frontend/src/views/admin/AccountsView.vue --details
python tools/check_ai_docs.py --stale --details
```

Notes:

- No third-party Python packages are required.
- Output is Chinese by design because project AI operations are Chinese-first.
- Non-zero exit means the AI documentation layer needs attention.
