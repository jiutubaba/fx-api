# `.ai/` - AI Documentation Center

This directory stores project facts, routing, session summaries, operational memory, and historical archive for AI agents working on Fuxi API.

Source code, deploy scripts, server state, and migration reports remain the final source of truth. `.ai/` is a navigation and collaboration layer.

## New Session Reading Order

1. Read `.ai/router.md` to find the narrow context for the task.
2. Read the relevant `.ai/systems/*.md` files.
3. Read `.ai/MEMORY.md` for the current hot state.

## Structure

```text
.ai/
├── README.md
├── router.md
├── shared_instructions.md
├── project_guide.md
├── rules.md
├── code_style.md
├── systems/
├── MEMORY.md
├── sessions.md
├── backlog.md
└── archive/
```

## Legacy Note

The old `F:\newAPI` AI architecture inspired this structure. Its old paths, release scripts, GORM/React/Bun assumptions, and `/data/new-api/**` runtime model are archived as legacy context only.

