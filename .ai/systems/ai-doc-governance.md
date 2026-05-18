# AI Doc Governance

- `.ai/` is a collaboration layer, not source code truth.
- Stable project facts belong in `.ai/systems/*.md`.
- Current operational state belongs in `.ai/MEMORY.md`.
- Recent high-value session summaries belong in `.ai/sessions.md`.
- Long historical records belong in `.ai/archive/`.
- Legacy old-project facts must be marked as legacy and must not override active sub2api rules.

Before editing `.ai/`, check:

1. Does the fact come from current code, deploy scripts, server output, or a migration report?
2. Is the fact stable enough for a system doc, or only hot enough for memory?
3. Could the note leak secrets or credentials?

