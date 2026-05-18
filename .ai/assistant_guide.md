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

When the user says `更新发布版` or `更新发布版并归档`, interpret it as a full production release request, not a simple version-file edit.

Default meaning:

1. Read `.ai/systems/release-and-environments.md`.
2. Bump the patch version unless the user explicitly asks for minor or major, or the requested release already exists.
3. Run the relevant local verification for the touched areas.
4. Commit and push the release change when the work is ready.
5. Publish or update the GitHub tag/Release.
6. Deploy production with `.agents/skills/update-server` and verify the deployed environment.
7. Record results in `.ai/MEMORY.md`, `.ai/sessions.md`, and `.ai/archive/sessions/`.

The phrases `更新发布版` and `更新发布版并归档` are explicit confirmation for a normal production app update to `https://fuxiapi.top/`. Production Caddy cutover, rollback target changes, and deletion of preserved legacy resources remain separately confirmation-gated.
