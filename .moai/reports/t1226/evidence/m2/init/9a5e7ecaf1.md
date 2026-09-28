# M2 attempt — AGENTS.md | 11. moai CLI Verbs

surface = init
pre_chars = 1027
post_chars = 935
chars = 92
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

## before

````
## 11. moai CLI Verbs

| Verb | Purpose |
|------|---------|
| `moai init <project> --llm claude\|codex\|both` | Scaffold a project and select its LLM harness |
| `moai update` | Sync templates and refresh already-enabled wiring |
| `moai tool enable codex` | Add or refresh Codex wiring in an existing project |
| `moai hook <event>` | Hook dispatcher entry point (drives hooks.json / settings.json) |
| `moai doctor` | Diagnose installation and wiring health |
| `moai worktree` | Worktree lifecycle (sync / remove / clean / recover / done / snapshot / verify / restore) |
| `moai cc` / `moai glm` / `moai gpt` | Explicit Claude, GLM, or GPT session launchers |
| `moai migrate cg` | Preview legacy CG migration; role changes require explicit acceptance |
| `moai version` | Print build version and provenance |
| `moai codex` | Codex session launcher — `cli` launch, `status` readout, `app` web; `-w <worktree>` enters an existing tree and never creates one |

Run `moai --help` for the generated, current command surface.

````

## after

````
## 11. moai CLI Verbs

| Verb | Purpose |
|---|---|
| `moai init <project> --llm claude\|codex\|both` | Scaffold a project, select its LLM harness |
| `moai update` | Sync templates, refresh enabled wiring |
| `moai tool enable codex` | Add or refresh Codex wiring in an existing project |
| `moai hook <event>` | Hook dispatcher (drives hooks.json / settings.json) |
| `moai doctor` | Diagnose installation and wiring health |
| `moai worktree` | Worktree lifecycle (sync / remove / clean / recover / done / snapshot / verify / restore) |
| `moai cc` / `moai glm` / `moai gpt` | Claude, GLM, or GPT session launchers |
| `moai migrate cg` | Preview legacy CG migration; role changes need explicit acceptance |
| `moai version` | Build version and provenance |
| `moai codex` | Codex launcher — `cli`, `status`, `app` web; `-w <worktree>` enters an existing tree, never creates one |

`moai --help` gives the current command surface.

````
