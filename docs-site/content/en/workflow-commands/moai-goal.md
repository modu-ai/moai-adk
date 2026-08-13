---
title: "/moai goal"
weight: 35
draft: false
new: true
added_in: "v3.1"
---

{{< new-badge v3.1 >}}

A **condition-declared autonomous loop**: declare only the completion condition, and the session keeps taking turns until that condition holds. At the end of every turn the evaluator checks the condition, and once the condition is met the loop stops on its own. You no longer have to press "continue" at every step.

{{< callout type="info" >}}
{{< icon flash primary >}} <strong>New in v3.1</strong>: infinitely-sustainable goals (<code>--max-turns 0</code>), the semi-autonomous flow, and automatic block-cap adjustment were added in this release.
{{< /callout >}}

## What this page covers

`/moai goal` registers and arms one **completion condition** on the session. The condition sentence is parsed into two kinds of predicates:

- **Mechanical condition** — the truth value is decided by a shell command's exit code. Example: "`go test ./...` exits 0".
- **Model condition** — decided by whether a particular line is present in the conversation record. Example: "every AC row is recorded as PASS".

The registered condition is read by the `moai hook stop-goal` Stop-hook evaluator at every turn end. If the condition does not hold, the evaluator blocks the turn and chains the next one; once the condition holds, the loop ends. So even when you type less, long-horizon work keeps rolling forward without stopping.

## Why you need it

In long-horizon work — multi-milestone runs, large refactors, TDD cycles — the most expensive cost is **the round trip of pressing "continue" every turn**. A single `/moai run SPEC-X` calls the implementation agent, but when that agent returns, who chains the next turn? Usually you have to type another prompt.

`/moai goal` removes this round trip. Declare the condition once, and the evaluator decides "the condition is not yet met" at every turn end and opens the next turn automatically. When the completion condition is mechanically verifiable, a single command can drive an entire cycle. That is how the "minimize user intervention" principle of harness design becomes real through this one command.

Note, however, that **it does not run unconditionally on its own**. `/moai goal` is arm-only — its only job is to "register the condition and chain turns." It must be paired with a command that actually starts the work (`/moai run SPEC-X`, etc.). Arming only the condition without starting work produces an idle loop where every turn end finds the condition unmet and the turns just spin.

## Usage

```bash
# Register + arm the condition
> /moai goal "go test ./... exits 0 && lint is clean, or stop after 20 turns"

# Check status
> /moai goal status

# Check goal status across all sessions
> /moai goal status --all

# Stop the loop
> /moai goal clear
```

Wrap the condition sentence in double quotes. Because the evaluator actually runs the mechanical-condition command at every turn end, fast and decisive commands (`go test -run <pattern>`) cost less per turn than the slow full suite (`go test ./...`).

### Three principles for writing a good condition

1. **One measurable end state** — a test result, a build exit code, a file count, a queue being empty. An abstract goal ("the code gets better") cannot be judged by either the machine or the model.
2. **State how it is measured** — "`go test ./...` exits 0", "`git status` is clean". With only "tests pass" the evaluator cannot know what to run.
3. **Include constraints** — "without touching other test files while doing so". Looking only at the end state, the middle process can change things you did not intend.

## When to use it

`/moai goal` shines in four situations:

- **T1 — Long run-phase / multi-milestone SPEC (Tier M/L)**: the run-phase autonomy wiring (the `ac_converge` block in `run.md`) owns this case. It chains the run phase until every acceptance criterion (AC) of the SPEC reads PASS.
- **T2 — Large migration / refactor touching many call sites**: the loop helps until every call site compiles and the tests pass. Arming the goal after the call-site inventory is surfaced in the conversation is the precondition.
- **T3 — TDD cycle / AC convergence**: running the RED-GREEN-REFACTOR loop, it chains until the target test suite is green and lint is clean.
- **T4 — Alternative to `/moai loop`**: a choice between "fix what the tooling flags" (`/moai loop`) and "roll until the declared end state becomes true" (`/moai goal`). When the end is clear, `/moai goal` is the better fit.

## Infinite-duration goals (v3.1)

With `--max-turns 0` you turn off the turn ceiling, so the loop does not stop until a wall-clock / stagnation guard trips. This is the right shape for long unattended runs.

```bash
# 4-hour wall-clock ceiling, unlimited turns
> /moai goal "<condition>" --max-turns 0 --max-duration 14400
```

However, the Claude Code runtime's consecutive-block cap (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, default 8) breaks the loop before the turn ceiling does. When arming an infinite goal you must raise this cap as well.

```bash
# Raise the block cap to 200 so the infinite loop truly persists
$ CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200
> /moai goal "<condition>" --max-turns 0 --max-duration 14400
```

From v3.1, arming with `moai goal arm --max-turns 0` or entering factory mode has the launcher inject this cap at 200 automatically. So users do not have to touch the environment variable themselves for a 4-hour chain not to break mid-way.

## arm-only and the safety boundary

`/moai goal` is **arm-only**. It registers the condition and chains turns; it does not start work itself. So it is always written paired with a work-starting command.

```bash
# Wrong form — arms only the goal and starts no work (idle loop)
> /moai goal "<condition>"

# Correct form — armed in the same turn as the run command
> /moai run SPEC-X     # starts the work
> /moai goal "<condition>"   # chains turns until run finishes
```

An armed goal **does not relax the safety boundary**:

- **Implementation Kickoff Approval** (the plan→run human gate) is still mandatory. Arming a goal does not pre-authorize run-phase entry.
- **The confirmation boundary for hard-to-reverse / shared-system actions** (PR creation, destructive operations) stays in place too. The evaluator decides only "whether to chain the turn"; it does not pre-authorize destructive operations.

## Evaluation cost

At every turn end the evaluator runs the mechanical-condition command. So the cost per turn is **the cost of the command inside the condition**. Fast and decisive commands tighten the turn loop. The Stop-hook timeout is 120 seconds, but faster commands make the loop more agile. Model conditions judge conversation records that already exist, so they carry no additional cost.

## Comparison — three autonomous continuation forms

| Form | What opens the next turn | Stopping condition |
|------|-------------------------|--------------------|
| **`/moai goal`** | After the previous turn ends, the evaluator decides the condition is not met | Condition holds, turn ceiling, stagnation guard, `/moai goal clear` |
| `/loop` (Claude Code built-in) | Re-runs a prompt/command after a fixed time interval passes | User cancels |
| `/moai loop` | A diagnostic scan builds a finite issue queue, and the goal engine evaluates "queue is empty and diagnostics are clean" at every turn end | Queue is empty and diagnostics are clean, or the ceiling is reached |

`/moai goal` and `/moai loop` are complementary. `/moai loop` fits "fix everything the tooling flags", and `/moai goal` fits "roll until the declared end state is proven true". They run on the same goal engine, but the actor that decides what counts as "done" differs.

## Relationship to other commands

- **`/moai loop`** — a diagnostics-driven decisive loop. The difference is that the tooling decides what to fix. [`/moai loop`](/en/utility-commands/moai-loop) lives in the utility-commands section.
- **`/moai run`** — the work-starting command. Written paired with goal. [`/moai run`](./moai-run).
- **Factory mode** — an entry switch that bundles `/moai goal`'s infinite-duration goal into a `plan → run → verify → sync` chain. Covered in [Factory mode](/en/advanced/factory-mode).

## What this command does not do (scope boundary)

- **It does not start work** — it is arm-only. Arming only the condition and not starting work produces an idle loop.
- **It does not bypass human gates** — Implementation Kickoff Approval, PR creation, and other hard-to-reverse decisions remain the user's exclusive right.
- **It does not run with hooks disabled** — when `disableAllHooks` or `allowManagedHooksOnly` is on, the Stop-hook evaluator itself does not run, so the flow degrades to the standard per-turn manual flow (graceful degradation).
- **It does not conflict with the native `/goal`** — when the runtime signals that the native `/goal` is active, the MoAI evaluator yields (preventing double-block).

## Related docs

- [Autonomous continuation loop](/en/advanced/autonomous-loops) — the goal engine's stagnation guard and ceiling semantics
- [Factory mode](/en/advanced/factory-mode) — the 4-stage chain tied by the `factory_chain` goal preset
- [`/moai loop`](/en/utility-commands/moai-loop) — the diagnostics-driven decisive loop (sibling command)
- [Harness engineering](/en/core-concepts/harness-engineering) — the path by which loops and observation flow into harness learning
