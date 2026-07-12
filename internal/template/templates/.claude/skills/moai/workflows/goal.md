# /moai goal — Condition-Declared Universal Agentic Loop

> `/moai goal` is the MoAI-owned, PROGRAMMATIC reimplementation of native `/goal`
> semantics. Native `/goal` is HUMAN-ONLY (the model cannot set it on the user's
> behalf); `/moai goal` is the only pipeline path to arm a goal-condition loop
> programmatically. See `.claude/rules/moai/workflow/goal-directive.md` and
> `.claude/rules/moai/workflow/native-invocation-model.md` § Axis B.

## What It Is

A goal declares completion conditions (mechanical shell commands + model-evaluated
claims). Once armed, the session iterates any work until the conditions hold or a
turn ceiling is reached. The `moai hook stop-goal` Stop-hook evaluator loads the
session's goal state each turn-end and emits a block decision until the goal
converges or the ceiling fires.

State lives at `.moai/state/goal/<session-id>.json` (per-session — never a single
shared file). A turn ceiling (default 30) bounds the loop. At the ceiling, the
evaluator emits a 5-section verdict (Claim / Evidence / Baseline-attribution /
Gaps / Residual-risk) and stops blocking.

## Verbs

### `/moai goal "<condition>"` — register + arm

Register the condition text and arm the goal for the active session. The
condition is parsed into a `conditions[]` array: a bare shell-command string is a
mechanical condition (`go test ./... exits 0`); a claim referencing the
transcript is a model condition (`all AC rows show PASS in the transcript`).
The orchestrator MAY pass a structured condition set when arming programmatically.

Arming writes `.moai/state/goal/<session-id>.json` (atomic temp+rename). The
Stop hook `handle-stop-goal.sh` picks it up on the next turn-end.

### `/moai goal status [--all]`

Print the active session's goal (or all sessions' goals with `--all`): the
condition text, the conditions array, turns used vs ceiling, the progress log,
and the lifecycle status (`armed` / `satisfied` / `ceiling-exit` / `cleared`).

### `/moai goal clear`

Clear the active session's goal (delete its state file). The Stop hook then sees
no armed goal and stops blocking. This is how the orchestrator ends the loop once
it has evaluated the model claim as met.

### `/moai goal resume` — deferred (follow-up), NOT delivered

**Out of scope — deferred to a follow-up.** The `resume` verb (best-effort re-arm
of a previously cleared goal by restoring from the `consumed/` archive) is NOT
delivered by the current arm CLI; `moai goal --help` lists only `arm` / `status`
/ `clear`. The reason it is deferred: `clear` DELETES the state file (it does not
tombstone into `consumed/`), and `consumed/` is the orphan-prune archive, not a
`clear` destination — so a goal cleared via `clear` never lands in `consumed/` and
cannot be resumed from it. Delivering a working `resume` would require changing
`clear` from a delete to a tombstone-move, a semantic change to the existing
`clear` contract that is out of scope here.

## Progression Mode (Autonomous / Semi-autonomous) — chosen at Implementation Kickoff Approval

When the orchestrator runs Implementation Kickoff Approval (`AskUserQuestion` at
the plan→run boundary), it offers an **autonomous vs semi-autonomous**
progression-mode choice as a DISTINCT axis from the approve/decline decision.

- **Approval remains required in both modes.** The progression-mode axis selects
  ONLY what happens AFTER the gate passes — it is never a gate bypass, never a
  relaxation of Implementation Kickoff Approval. An armed goal never authorizes
  run-phase entry, never creates a PR, and never performs a destructive operation
  regardless of the selected mode.
- **Autonomous mode** (default): the evaluator blocks each turn until the
  conditions hold or the ceiling is reached, with NO per-turn user prompt. This
  is the existing Stop-hook behavior — no NEW behavioral surface beyond the
  `progression_mode` state field.
- **Semi-autonomous mode**: the `stop-goal` hook emits a **checkpoint-signal**
  block JSON at each turn boundary for orchestrator-side `AskUserQuestion`
  confirmation. The hook itself NEVER calls `AskUserQuestion` (subagent/hook
  boundary — it emits structured JSON only); the orchestrator reads the
  checkpoint and runs the confirm round.

The selected mode is persisted in goal state as `progression_mode` (default
`autonomous` when the user declines to choose).

### Semi-autonomous checkpoint flow (the orchestrator bridge)

When `progression_mode == "semi-autonomous"` and the goal is not yet satisfied
and the ceiling is not yet reached, `stop-goal` emits exit-0 stdout JSON shaped:

```json
{
  "decision": "block",
  "reason": "semi-autonomous checkpoint: orchestrator to confirm continuation (turn 3 of 30)",
  "mode": "semi-autonomous",
  "turn": 3,
  "ceiling": 30,
  "last_progress": "M2 evaluator Tier-1 gate implemented",
  "failed_conditions": [
    {"cmd": "go test ./internal/goal/...", "exit": 1, "tail": "FAIL: TestX ... (output tail)"}
  ]
}
```

The orchestrator (which CAN call `AskUserQuestion` — it is the main session) reads
this checkpoint JSON and runs a confirm round offering at least: continue to next
step, clear the goal, or switch to autonomous. The `failed_conditions` array
carries the failed-condition + output-tail detail so the confirm round surfaces
WHY the goal isn't satisfied — the generic `reason` label alone is insufficient
for an informed continue/clear/switch decision. When no mechanical condition is
failing (e.g., the checkpoint fires because a model claim is not yet satisfied),
`failed_conditions` is empty.

This reuses the existing orchestrator-translation-responsibility pattern already
codified for `team-ac-verify.sh` and `sync-phase-quality-gate.sh` (hooks emit
structured JSON; the orchestrator translates to `AskUserQuestion`). No NEW
boundary-crossing mechanism is invented.

## Safety Invariants

1. **Implementation Kickoff Approval is mandatory in both modes.** The
   progression-mode axis is a post-approval progression CHOICE, not a relaxation
   of the gate. The gate stays mandatory and score-independent in both autonomous
   and semi-autonomous modes.
2. **An armed goal does not bypass Kickoff**, does not auto-create a PR, does not
   perform destructive operations. The evaluator only decides whether the turn
   continues; it never pre-approves irreversible actions.
3. **The `stop-goal` hook never calls `AskUserQuestion`** or any user-prompting
   tool (subagent/hook boundary). It emits structured JSON only.
4. **Native `/goal` yield**: when the runtime signals an active native `/goal`,
   `stop-goal` yields (no double-block). Where the runtime does not expose the
   signal, the hook degrades to "always evaluate the MoAI goal" (accepted DEBT —
   no correctness hazard; possible double evaluation only).
5. **Stagnation guard**: N consecutive no-progress iterations halt the loop and
   emit a 5-section verdict carrying an E1/E3 escalation note.

## Goal `cmd`s SHOULD be fast

The goal evaluator runs at turn-end. Prefer `go test -run <pattern>` over the
full suite, and prefer deterministic commands over long-running ones — the
Stop-hook timeout for `stop-goal` is 120s, but a faster cmd keeps the turn loop
tight.

## Cross-references

- `.claude/rules/moai/workflow/goal-directive.md` — native `/goal` semantics + the `/moai goal` PROGRAMMATIC counterpart row + Axis B.
- `.claude/rules/moai/workflow/native-invocation-model.md` § Axis B — the HUMAN-ONLY automation justification.
- `.claude/rules/moai/workflow/session-handoff.md` § Post-Paste /goal Follow-up Block — the post-paste native `/goal` is now an optional variant; Block 5 MAY carry `/moai goal "<condition>"`.
- The goal-engine Go package and the `moai hook stop-goal` verb (implementation-owned; see the local project source tree).
