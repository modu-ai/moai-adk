# SPEC-CLIFIX-HYGIENE-001 M1 — GLM Env Inject/Clear Sites Inventory (Frozen Fixture)

> Frozen by manager-develop M1 on 2026-07-30 against worktree HEAD `359e887b9`.
> This is the baseline M3 will replace with `envkeys.go` constants + a shared
> `glmEnvVarSet()` helper per spec.md REQ-HYG-001-003 / plan.md §B row 3.

## GLM env inject/clear sites observed in glm.go / glm_tools.go / launcher.go

| File:line | Direction | Env key | Verbatim |
|---|---|---|---|
| `internal/cli/glm.go:484` | inject | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1",` |
| `internal/cli/glm.go:569` | clear | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | element of a clear-list |
| `internal/cli/glm.go:571` | clear | `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | element of a clear-list |
| `internal/cli/glm.go:616` | clear | `CLAUDE_CODE_TEAMMATE_DISPLAY` | `delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")` |
| `internal/cli/glm.go:682` | inject | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"` |
| `internal/cli/glm.go:711` | clear | `CLAUDE_CODE_TEAMMATE_DISPLAY` | `delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")` |
| `internal/cli/glm.go:926` | inject | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1",` (inside `buildGLMEnvVars` — the dead-code candidate) |
| `internal/cli/glm.go:970` | inject | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"` |
| `internal/cli/launcher.go:305` | clear | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | `delete(env, "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS")` |
| `internal/cli/launcher.go:307` | clear | `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | `delete(env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC")` |
| `internal/cli/launcher.go:309` | clear | `CLAUDE_CODE_TEAMMATE_DISPLAY` | `delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")` |
| `internal/cli/glm_tools.go:805` | (read) | `MOAI_GLM_NO_AUTO_TOOLS` | `os.Getenv("MOAI_GLM_NO_AUTO_TOOLS") == "1"` |

## Drift hazard (what M3 closes)

The three env keys `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`,
`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, and `CLAUDE_CODE_TEAMMATE_DISPLAY`
appear as inline string literals at every inject AND clear site. If one site
is renamed and the others are not, the inject/clear sets silently drift —
exactly the defect REQ-HYG-001-003 closes. M3 introduces:

1. `envkeys.go` constants for the three keys (+ `MOAI_GLM_NO_AUTO_TOOLS`).
2. One shared `glmEnvVarSet()` helper that returns the canonical inject set.
3. Every inject site calls the helper; every clear site deletes the helper's
   keys — the two sets are structurally identical by construction.

## AC-HYG-001-003 parity check (M3 will mechanize this)

After M3, `go test ./internal/cli/ -run 'GLMEnvSetParity'` asserts inject set
== clear set by reading the helper once and diffing. And `grep` for the bare
`ANTHROPIC_` / `CLAUDE_` env literals outside `envkeys.go` returns 0.
