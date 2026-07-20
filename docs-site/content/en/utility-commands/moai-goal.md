---
title: /moai goal
weight: 25
draft: false
---

A **condition-declared autonomous loop** command: declare a completion condition and the session works on its own until that condition holds. When you arm a completion condition with `/moai goal "<condition>"`, the `stop-goal` Stop hook evaluates whether the condition is met at the end of every turn and automatically starts the next turn until it is satisfied.

{{< callout type="info" >}}
**One-line summary**: `/moai goal` is "a general-purpose loop that declares an end state." If `/moai loop` is a preset whose condition — "until every issue found by the diagnostic tools is gone" — is predetermined, then `/moai goal` is the general-purpose engine where you **declare the completion condition yourself**.
{{< /callout >}}

{{< callout type="info" >}}
**Programmatic command**: The native Claude Code `/goal` is a TUI command that only a user can type (HUMAN-ONLY). `/moai goal` is a MoAI-owned command that implements the same semantics **programmatically within the pipeline**, entered via `moai` skill routing and the `moai goal` CLI.
{{< /callout >}}

## Overview

Use it when you want to tell the agent "keep working on your own until this condition is met." Conditions can mix two kinds.

- **Mechanical condition**: a condition verified by a shell command. Example: `go test ./... exits 0`. It runs the command and observes the exit code.
- **Model-evaluated condition**: a condition verified by a judgment over the transcript. Example: `all AC rows recorded as PASS`. It is evaluated against what the session has surfaced so far.

This loop is the general-purpose engine of v3's second pillar, **agentic loop engineering**. Goal state is stored per session in `.moai/state/goal/<session-id>.json` (not a shared file), and a **turn ceiling (default 30)** makes the loop bounded. When the ceiling is reached, the evaluator issues a 5-section verdict (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) and stops blocking.

## Verbs

### `/moai goal "<condition>"` — register + arm

Registers the condition text and arms the goal on the active session. The condition is parsed into a `conditions[]` array — a pure shell-command string is a mechanical condition, and an assertion referencing the transcript is a model condition. On arming, `.moai/state/goal/<session-id>.json` is written atomically (temp+rename), and the `stop-goal` Stop hook picks it up and begins evaluating at the end of the next turn.

```bash
> /moai goal "go test ./... exits 0; all AC recorded as PASS, or stop after 30 turns"
```

### `/moai goal status [--all]`

Prints the active session's goal (or, with `--all`, all sessions' goals) — the condition text, the conditions array, turns used vs. the ceiling, the progress log, and the lifecycle state (`armed` / `satisfied` / `ceiling-exit` / `cleared`).

### `/moai goal clear`

Releases the active session's goal (deletes the state file). The Stop hook sees no armed goal and stops blocking. This is how the orchestrator ends the loop after judging a model condition satisfied.

{{< callout type="info" >}}
**There is no `resume` verb.** The once-discussed `resume` verb (restoring a released goal from an archive) does not exist in the current CLI — `moai goal --help` lists only `arm` / `status` / `clear`. Because `clear` **deletes** the state file (it does not tombstone to an archive), there is no original left to restore.
{{< /callout >}}

## Progression modes (autonomous / semi-autonomous)

When the orchestrator runs Implementation Kickoff Approval (the `AskUserQuestion` at the plan→run boundary), it lets you choose the **autonomous vs. semi-autonomous** progression mode as a **separate axis distinct** from the approve/reject decision. The chosen mode is stored in the goal state's `progression_mode` field (default `autonomous` if the user does not choose).

| Mode | Behavior |
|------|------|
| **autonomous (default)** | The evaluator blocks each turn until the condition is met or the ceiling is reached, without asking the user each turn. This is the existing Stop hook behavior as-is. |
| **semi-autonomous** | The `stop-goal` hook emits a **checkpoint-signal** block JSON at each turn boundary, and the orchestrator reads it to run an `AskUserQuestion` confirmation round (continue / release goal / switch to autonomous). The hook itself never calls `AskUserQuestion` (hook/subagent boundary — it emits structured JSON only). |

{{< callout type="warning" >}}
**Approval is required in both modes.** The progression-mode axis only chooses what to do **after** the gate has passed — it is not a gate bypass, nor a relaxation of Implementation Kickoff Approval. An armed goal, in any mode, does not authorize run-phase entry, create a PR, or perform destructive operations.
{{< /callout >}}

## Safety invariants

1. **Implementation Kickoff Approval is required in both modes** — the progression mode is a post-approval progression choice, not a gate relaxation, and it holds regardless of score.
2. **An armed goal does not bypass gates** — it does not auto-create a PR, and does not perform destructive operations. The evaluator only decides whether to continue turns; it does not pre-approve irreversible operations.
3. **The `stop-goal` hook does not call `AskUserQuestion`** — it emits structured JSON only (hook/subagent boundary).
4. **Stagnation guard** — when N consecutive no-progress iterations are detected, the loop stops and issues a 5-section verdict carrying an E1/E3 escalation note.

## goal conditions should be fast

The evaluator runs at the end of every turn. Prefer `go test -run <pattern>` over the full suite, and deterministic commands over long-running ones — the `stop-goal` Stop hook timeout is 120 seconds, but fast commands keep the turn loop tight.

## Relationship to /moai loop

`/moai loop` is a **preset on top of the goal engine**. If `/moai goal` is the general-purpose loop where the user declares the completion condition directly, then `/moai loop` is a preset that pre-fills the condition "until the issue queue found by the diagnostic tools is drained."

| Engine | Goal | Completion condition |
|------|------|----------|
| `/moai goal` | Condition-declared general-purpose loop | A user-defined condition expression is satisfied |
| `/moai loop` | Diagnostic fix loop (preset) | Issue queue drained + diagnostics clean (0 errors / tests pass / coverage) |

If the end state can be expressed as a condition expression, `/moai goal` is right; if it is "get rid of every problem the tools find," `/moai loop` is right.

## Related documents

- [/moai loop - iterative fix loop](/en/utility-commands/moai-loop)
- [/moai fix - one-shot auto-fix](/en/utility-commands/moai-fix)
- [/moai - fully autonomous automation](/en/utility-commands/moai)
