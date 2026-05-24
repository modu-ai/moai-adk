# internal/cli — Module Conventions

## Purpose

`internal/cli` hosts every `moai` CLI subcommand (`init`, `update`, `hook`, `spec`, `worktree`, `harness`, `cg`, `cc`, `glm`, `doctor`, ...). Each subcommand registers under the root cobra command in `root.go` and lives in a dedicated subdirectory (`harness/`, `worktree/`, `spec/`, ...). All user-facing CLI surface flows through this package — adding a subcommand here means a new entry on `moai --help`.

This package is the boundary between the user's terminal and every backing subsystem. Anything that breaks here breaks discovery, scripting, CI integration, and Claude Code hook invocation.

## Conventions

- **Subagent boundary (C-HRA-008 / REQ-PGN-012)**: CLI code MUST NOT call `AskUserQuestion` or `mcp__askuser__*`. The CLI runs in subagent context — orchestrator owns user interaction. Replace any interactive prompt with positional arguments + `--flag` defaults + structured stderr errors. Static guard pattern: `TestNew_NoAskUserQuestion` (see `worktree/new_test.go`) — every interactive-shaped subcommand needs the equivalent grep-based test.
- **Cobra subcommand registration**: Add new subcommands via `xxxCmd := &cobra.Command{...}` then `rootCmd.AddCommand(xxxCmd)` in the package's `init()` or in `root.go`. Never declare two subcommands with the same `Use:` prefix — cobra panics at runtime. The pattern catches `harness-classify`-style multi-word names colliding with existing `harness` namespace.
- **Exit code discipline**: `os.Exit(0)` success, `os.Exit(1)` user error, `os.Exit(2)` system error. Never `panic()` — wrap upstream errors and report to stderr.
- **Output streams**: stdout = structured machine-readable output (JSON when `--json`, plain text otherwise). stderr = human progress messages, warnings, errors. Never mix.
- **Absolute paths**: Resolve user-supplied paths via `filepath.Abs(userPath)`. Never `filepath.Join(cwd, userPath)` — on macOS `t.TempDir()` returns `/var/folders/...` and `filepath.Join` does NOT strip leading `/` from the second arg (see CLAUDE.local.md §6 Test Isolation).
- **Cross-platform**: All path handling MUST work on linux/darwin/windows. Use `filepath.Join` for in-repo paths, `os.PathSeparator` for printing. Verify with `GOOS=windows GOARCH=amd64 go build ./...` before commit.
- **Env var access**: Use constants from `internal/config/envkeys.go` — never inline `os.Getenv("ANTHROPIC_*")` strings. Hardcoded env names are a §14 violation per CLAUDE.local.md.

## Key Patterns

- **`syscall.Exec` replacement** (`worktree new --team` Pattern P3): Use `syscall.Exec(argv0, argv, env)` to replace the current process rather than spawn-and-wait. The current shell becomes the new process. Windows lacks this primitive — gate behind `team_launch_posix.go` build tag and provide a stub in `team_launch_windows.go`.
- **tmux pane spawning** (`worktree new --team` Pattern P1/P2): Use `tmux new-window` (within a session) not `tmux new-session` (creates a detached session). Read pane ID from stdout for swarm registry.
- **Settings.json mutation**: All mutations go through `internal/cli/settings.go` helpers — never `json.Marshal` settings directly. The helpers preserve user-only keys like `defaultMode`, `teammateMode`, `env.PATH` per CLAUDE.local.md §22.
- **Sandbox-safe destructive ops**: `git reset --hard`, `rm -rf` are blocked by claude-code sandbox. Use `git reset --keep` + `git stash push --include-untracked` patterns per CLAUDE.local.md §23.5/§23.6 instead.
- **Background subagent restriction**: Subcommands that write files MUST NOT be spawned with `run_in_background: true` — background subagents auto-deny Write/Edit per CLAUDE.md §14. Read-only commands (`status`, `list`, `version`) are safe to background.

## Cross-References

- Root CLAUDE.md §4 (Agent Catalog), §14 (Parallel Execution Safeguards § Background Agent Write Restriction)
- CLAUDE.local.md §13 (GLM integration test isolation), §14 (Hardcoding prevention), §22 (settings.local.json separation)
- `internal/cli/worktree/new_test.go` — `TestNew_NoAskUserQuestion` canonical static guard
- `internal/cli/harness/route_test.go` — `TestPropose_NoAskUserQuestion` second canonical instance
- `internal/cli/worktree/team_launch.go` + `team_launch_posix.go` + `team_launch_windows.go` — cross-platform exec/spawn pattern
- `internal/config/envkeys.go` — canonical env var constants (use these, never inline strings)
