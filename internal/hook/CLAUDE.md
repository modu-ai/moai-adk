# internal/hook — Module Conventions

## Purpose

`internal/hook` implements the Claude Code hook event handlers — `SessionStart`, `SessionEnd`, `PreToolUse`, `PostToolUse`, `Stop`, `SubagentStop`, `Notification`, `UserPromptSubmit`, `PreCompact`, `TeammateIdle`, `TaskCompleted`. Each event reads JSON from stdin (per Claude Code spec) and writes structured response to stdout. Hooks are invoked by Claude Code via the wrapper scripts in `.claude/hooks/moai/handle-*.sh`, which `exec moai hook <event>` and pipe stdin through.

This package is the runtime observation + control plane. Hook handlers run with ≤5s timeout per default — exceeding it stalls the user's Claude Code session.

## Conventions

- **Shell-script wrappers (CLAUDE.local.md §7 [HARD])**: moai-adk uses shell scripts in `.claude/hooks/moai/handle-*.sh`, NOT Python. The wrapper does `INPUT=$(cat); moai hook <event> <<< "$INPUT"`. Faster startup than uv/python. Settings.json command MUST quote `$CLAUDE_PROJECT_DIR` (CLAUDE.local.md §7).
- **`$CLAUDE_PROJECT_DIR` resolution priority** (B7 known issue): Hook code that needs the project root MUST resolve in this order: (1) `os.Getenv("CLAUDE_PROJECT_DIR")` if set, (2) `os.Getwd()` fallback. Never assume cwd equals project root — worktree hooks run with cwd inside `~/.moai/worktrees/{project}/{spec}/`. The `observer.go` path-resolution helper centralizes this — use it, never inline `os.Getenv`.
- **Subagent boundary (C-HRA-008)**: Hook handlers MUST NOT call `AskUserQuestion` or `mcp__askuser__*`. Hooks run in subagent context with no user interaction channel. Use stderr for diagnostics, structured stdout for hook response. Static guard: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` MUST return 0 matches.
- **Background agent write restriction (CLAUDE.md §14 [HARD])**: Hook handlers that mutate files (`PostToolUse` cache update, `SessionEnd` state persistence) MUST NOT be invoked with `run_in_background: true` from upstream. Background subagents auto-deny Write/Edit. Read-only hooks (`Notification`, `UserPromptSubmit` observers) are safe to background.
- **OpenTelemetry env var safety in tests (CLAUDE.local.md §2 [WARN])**: NEVER `t.Setenv("OTEL_EXPORTER_*", ...)` in parallel tests. The OTEL SDK initializes global state from env on first use — concurrent tests cause data races. Use a fake exporter (`sdktrace.NewTracerProvider(sdktrace.WithSyncer(testExporter))`) instead.
- **Timeout discipline**: Default 5s per hook event. If a hook needs >5s (large repo scan, network call), the hook config in `settings.json` MUST specify `"timeout": <ms>`. Hardcoded sleeps inside a hook are forbidden — use `context.WithTimeout` and return early on cancellation.

## Key Patterns

- **`handle-session-start.sh` env injection**: SessionStart hook reads `~/.moai/.env.glm` (if present) and injects GLM credentials into `.claude/settings.local.json`. This is the runtime-managed half of CLAUDE.local.md §2 separation contract. The injection MUST be idempotent — running twice yields the same `settings.local.json`.
- **`harness-classify` subcommand pattern** (SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001): New hook subcommands register under existing `hookCmd` namespace via `hookCmd.AddCommand(&cobra.Command{Use: "<name>"})`. Backward-compat preserved by NOT renaming `hookCmd` — it remains `moai hook *`.
- **JSON I/O contract**: Read stdin via `json.NewDecoder(os.Stdin).Decode(&payload)`, write response via `json.NewEncoder(os.Stdout).Encode(&response)`. Claude Code spec mandates exit code 0 = success, exit code 2 = block tool/keep working (per CLAUDE.md §15 TeammateIdle/TaskCompleted), other = error.
- **Observer registration**: New observers register via `observer.Register(eventName, ObserverFunc)`. Observers run in order of registration — keep them lightweight (<50ms each). Heavy work (file scans, network) goes in async goroutines with context cancellation.

## Cross-References

- Root CLAUDE.md §14 (Background Agent Write Restriction), §15 (Agent Teams Hook Events — TeammateIdle/TaskCompleted exit code semantics)
- CLAUDE.local.md §2 (settings.local.json separation + OTEL env race), §7 (Hook Development Guidelines)
- `internal/hook/observer.go` — `$CLAUDE_PROJECT_DIR` resolution helper (B7 canonical)
- `internal/hook/router.go` — event-name → handler dispatch
- `.claude/hooks/moai/handle-*.sh` — shell-script wrappers (template-managed)
- `internal/cli/harness/route.go` — `moai hook harness-classify` subcommand wiring
