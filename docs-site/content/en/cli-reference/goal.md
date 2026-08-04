---
title: moai goal Goal Loop
weight: 72
draft: false
---

`moai goal` arms/queries/clears a condition-declared agentic goal loop for the current session. The MoAI goal engine keeps the session working across turns until the declared condition is met or the turn ceiling is reached.

This is the programmatic MoAI counterpart of the native `/goal` (a user-only TUI command), letting the orchestrator register and arm a goal without a human typing the `/goal` line.

## Subcommands

| Command | Description |
|--------|------|
| `moai goal arm "<condition>"` | Register + arm a goal on the active session (`moai goal "<condition>"` is also an arm alias). Arm-only — it starts no work by itself |
| `moai goal status` | Print the goal status of the active session (use `--all` to list every session) |
| `moai goal clear` | Clear the goal of the active session |
| `moai goal render` | Render the active session's goal dashboard as a self-contained HTML file (saved next to `.moai/state/goal/`). Exits non-zero when no goal is armed |

## Common flags

| Flag | Description |
|--------|------|
| `--session <id>` | Override the session id (default: resolved via `moai session current`) |
| `--json` | Machine-readable JSON output |
| `--all` | (`status` only) List goals across all sessions, not just the active one |

## Arm flags

| Flag | Description |
|--------|------|
| `--max-turns <N>` | Turn ceiling. `0` = infinite (SPEC-INFINITE-GOAL-001); default `30` when omitted (full backward compat). **`0` (infinite) REQUIRES `--max-duration <sec>`** (arm-time fail-closed). |
| `--max-duration <sec>` | Wall-clock bound (seconds since arm time). **The actual wall-clock bound for an infinite goal (`--max-turns 0`)** — an infinite goal cannot be armed without this flag. |
| `--cost-cap <value>` | Recorded-only on the invocation ceiling — there is no enforcement logic today, so it is not an actual bound. It does not satisfy the real-bound requirement for `--max-turns 0`, so `--cost-cap` alone is rejected. |

## State and evaluation

Goal state is stored in `.moai/state/goal/<session-id>.json` (one file per session). The Stop hook `moai hook stop-goal` evaluates the goal at the end of every turn.

**Condition parsing**:

- An executable shell command (optionally suffixed with `exits <N>`) becomes a **mechanical condition**.
- A claim referencing the conversation transcript becomes a **model condition** that the orchestrator evaluates.

## Examples

```bash
# Keep working until the test suite passes
moai goal arm "go test ./... exits 0"

# Check the current goal status
moai goal status

# Clear the goal
moai goal clear
```

## Related documents

- [Autonomous Continuation Loops](/en/advanced/autonomous-loops) — `/goal` vs `/moai loop`
- [moai loop](/en/cli-reference/loop)
- [CLI Overview](/en/getting-started/cli)
