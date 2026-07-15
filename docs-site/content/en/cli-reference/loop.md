---
title: moai loop Feedback Loop
weight: 76
draft: false
---

`moai loop` manages the SPEC-lifecycle Ralph feedback-loop controller. It drives a state machine that iterates over the work detected by tool diagnostics for a single SPEC.

> This CLI command is distinct from the `/moai loop` skill in the Claude Code chat — the CLI manipulates the loop-controller state, while the `/moai loop` skill has the orchestrator perform the actual iterative fixes.

## Subcommands

| Command | Description |
|--------|------|
| `moai loop start <SPEC-ID>` | Start a feedback loop for a SPEC |
| `moai loop status` | Show the current loop status |
| `moai loop pause` | Pause a running loop |
| `moai loop resume <SPEC-ID>` | Resume a paused loop |
| `moai loop cancel` | Cancel a running loop |

## Examples

```bash
# Start a loop for a SPEC
moai loop start SPEC-AUTH-001

# Check the current status
moai loop status

# Pause and resume later
moai loop pause
moai loop resume SPEC-AUTH-001

# Cancel the loop
moai loop cancel
```

## Related documents

- [Autonomous Continuation Loops](/en/advanced/autonomous-loops) — the Ralph engine and goal-based loops
- [moai goal](/en/cli-reference/goal)
- [CLI Overview](/en/getting-started/cli)
