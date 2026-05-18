# Assistant Execution Guide

## Communication

- Use simplified Chinese.
- Lead with the conclusion, then the evidence, then verification and risks.
- State precise status: `已修改，已验证`, `已修改，未测试`, `已归档，未发布`, or `已发布并验证`.
- Do not weaken verification gaps. If a command was not run, say so.

## Execution Boundaries

- Work inside this repository unless the task explicitly requires checking preserved runtime state or the legacy `F:\newAPI` reference.
- Treat `F:\newAPI`, `/data/new-api/**`, `new-api`, and `new-api-staging` as protected rollback/legacy resources.
- Do not copy old new-api stack assumptions into active code paths. GORM, React/Rsbuild/Bun, old `/data/new-api/**` scripts, and old `new-api` containers are legacy-only here.
- Archive inside `.ai/`; do not create desktop notes or out-of-project handoff files.

## Release Requests

When the user says `更新发布版并归档`, interpret it as release-version closure, not a simple version-file edit.

Default meaning:

1. Read `.ai/systems/release-and-environments.md`.
2. Bump the patch version unless the user explicitly asks for minor or major.
3. Run the relevant local verification for the touched areas.
4. Commit and push the release-prep change when the work is ready for release.
5. If the user has explicitly confirmed staging or production update, run the corresponding skill and verify the deployed environment.
6. Record results in `.ai/MEMORY.md`, `.ai/sessions.md`, and `.ai/archive/sessions/`.

Production remains confirmation-gated. If the user only says `更新发布版并归档`, prepare and archive the release candidate, but do not deploy production unless the surrounding context explicitly confirms production update.
