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
| `moai goal arm "<condition>"` | Register + arm a goal on the active session (`moai goal "<condition>"` is also an arm alias) |
| `moai goal status` | Print the goal status of the active session |
| `moai goal clear` | Clear the goal of the active session |

## Common flags

| Flag | Description |
|--------|------|
| `--session <id>` | Override the session id (default: resolved via `moai session current`) |
| `--json` | Machine-readable JSON output |
| `--all` | (`status` only) List goals across all sessions, not just the active one |

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
