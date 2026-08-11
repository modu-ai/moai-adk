# Goal Directive (`/moai goal`) — Autonomous Continuation

`/moai goal` is MoAI's session-scoped completion condition: a condition-declared loop that keeps the session working across turns until the condition holds. It is the single goal-arming surface for every orchestrator emission path.

> **Full detail** (Comparing Approaches table, condition-authoring guide, T1-T4 condition templates, MoAI Integration Notes, Native `/goal` Prohibition rationale) lives in `goal-directive-detail.md`. Load it when actively arming a goal or choosing between `/moai goal` and `/moai loop`.

## What It Is

`/moai goal "<condition>"` registers a completion condition and arms it for the active session. The condition text is parsed into a `conditions[]` array mixing **mechanical** conditions (a shell command whose exit code decides) and **model** conditions (a claim the transcript must demonstrate). The `moai hook stop-goal` Stop-hook evaluator loads the session's goal state at each turn-end and emits a block decision until the goal converges or a bound fires — so the session keeps working without a prompt at each step.

State lives at `.moai/state/goal/<session-id>.json`, one file per session. A turn ceiling (default 30) bounds the loop; at the ceiling the evaluator emits a 5-section verdict (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) and stops blocking. The runtime's consecutive-block cap (default 8, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`) overrides the block first on an unattended run — the effective bound is `min(ceiling, cap)` and a missing verdict must not be read as convergence. A stagnation guard halts the loop after N consecutive no-progress iterations with an E1/E3 escalation note.

`/moai goal` is **arm-only** (see § Goal-Presentation Timing). The delivered verbs are `/moai goal "<condition>"` (register + arm), `/moai goal status [--all]`, and `/moai goal clear`. Full verb surface, the progression-mode axis, and the safety invariants live in `.claude/skills/moai/workflows/goal.md`.

Availability: `/moai goal` needs hooks enabled (its evaluator IS a Stop hook); it is unavailable when `disableAllHooks` or `allowManagedHooksOnly` is set. It carries no runtime-version floor of its own.

## Goal-Presentation Timing

**`/moai goal` is arm-only.** Arming records the condition in goal state and causes the `stop-goal` evaluator to block turn-end until the condition holds; it starts no work of its own. The consequence is concrete: a goal armed while nothing is running spins idle turns until the ceiling, because each turn-end finds the condition unmet and no work advancing it. Arming is therefore always paired with a work-starting action, and it never substitutes for one. This is why a paste-ready resume keeps a work-starting command (`/moai run SPEC-X`) as Block 5's single primary action rather than a bare goal-arming directive — see `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (Block 5).

**The goal is presented at the Implementation Kickoff Approval gate.** When the orchestrator runs Implementation Kickoff Approval (the `AskUserQuestion` round at the plan→run boundary), the goal is offered there as the **autonomous vs semi-autonomous progression-mode axis** — a DISTINCT axis from the approve/decline decision. The orchestrator arms the goal only after the gate passes. The axis and its two modes are specified in `.claude/skills/moai/workflows/goal.md` § Progression Mode.

**Arming a goal does not authorize autonomous run-phase entry.** The Implementation Kickoff Approval human gate remains required in both progression modes: the progression-mode axis selects only what happens AFTER the gate passes, and is never a gate bypass or a relaxation. An armed goal likewise never authorizes creating a PR or performing a destructive operation — the evaluator decides only whether the turn continues.

## Hard Preconditions for Every Recommendation

- **Implementation Kickoff Approval comes first**: any run-phase goal-arming is downstream of the Implementation Kickoff Approval human gate (`AskUserQuestion`, plan→run) and never substitutes for or bypasses it. `run.md` § Run-phase Autonomy #1 owns the preferences-drained rationale.
- **Arming is programmatic**: `/moai goal` is orchestrator-invocable, so the orchestrator arms the condition itself once the gate has passed. Where the nearest native equivalent is HUMAN-ONLY, this MoAI subcommand is the only pipeline path.
- **Safety boundary unchanged**: an armed goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary.
- **`run.md` "set" shorthand**: `run.md` § Run-phase Autonomy states the orchestrator MAY set the `ac_converge` goal — the orchestrator arms it via `/moai goal` after the gate passes.

## Proactive Recommendation Triggers

When the orchestrator recognizes a situation where a condition-declared loop is the right continuation primitive, it arms one rather than driving the work turn by turn. T1-T4 one-liners (full condition templates in `goal-directive-detail.md`):

- **T1 — Long run-phase / multi-milestone (Tier M/L)**: the run-phase autonomy wiring in `run.md` § Run-phase Autonomy (`ac_converge`) owns this case — surface that block verbatim.
- **T2 — Migration / refactoring across many call sites**: arm once the call-site inventory is enumerated and transcript-visible.
- **T3 — TDD cycle / SPEC AC convergence**: for SPEC-scoped work, T1 and `run.md` `ac_converge` is the SSOT; for non-SPEC TDD work arm a test-suite-shaped condition.
- **T4 — `/moai loop` alternative**: when work is better expressed as a verifiable end-state than as "fix what the tooling flags", surface the two options (decision axis: what should start the next turn).

## Cross-references

- `.claude/skills/moai/workflows/goal.md` — verb surface, progression-mode axis, semi-autonomous checkpoint flow, safety invariants
- `goal-directive-detail.md` — Comparing Approaches table, condition templates, MoAI Integration Notes, Native `/goal` Prohibition rationale
- `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (Block 5) — the resume-context goal line and the arm-only consequence on the handoff surface
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy — the `ac_converge` condition wiring

---

Version: 2.1.0 (stub reduced; detail moved to `goal-directive-detail.md` lazy companion)
Classification: Evolvable orchestration guidance — applies to autonomous multi-turn continuation
