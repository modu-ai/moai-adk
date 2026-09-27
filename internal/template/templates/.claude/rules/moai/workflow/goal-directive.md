# Goal Directive (`/moai goal`) — Autonomous Continuation

`/moai goal` is MoAI's session-scoped completion condition: a condition-declared loop that keeps the session working across turns until the condition holds. It is the single goal-arming surface for every orchestrator emission path.

> **Full detail** (Comparing Approaches table, condition-authoring guide, T1-T4 condition templates, MoAI Integration Notes, Native `/goal` Prohibition rationale) lives in `goal-directive-detail.md`. Load it when actively arming a goal or choosing between `/moai goal` and `/moai loop`.

## What It Is

`/moai goal "<condition>"` registers a completion condition and arms it for the active session. The condition text is parsed into a `conditions[]` array mixing **mechanical** conditions (a shell command whose exit code decides) and **model** conditions (a claim the transcript must demonstrate). The `moai hook stop-goal` Stop-hook evaluator loads the session's goal state at each turn-end and emits a block decision until the goal converges or a bound fires — so the session keeps working without a prompt at each step.

Prefix the condition with `model:` or `cmd:` to declare which tier is meant. Without a prefix the tier is inferred from an English substring test, which reads a claim written in any other language as a shell command — and a shell command that cannot run never exits 0, so the goal blocks every turn until its bound fires.

State lives at `.moai/state/goal/<session-id>.json`, one file per session. A turn ceiling (default 30) bounds the loop; at the ceiling the evaluator emits a 5-section verdict and stops blocking. The runtime's consecutive-block cap (default 8, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`) overrides the block first on an unattended run — the effective bound is `min(ceiling, cap)`, and a missing verdict must not be read as convergence. A stagnation guard halts the loop after N consecutive no-progress iterations.

`/moai goal` is **arm-only** (see § Goal-Presentation Timing). The delivered verbs are `/moai goal "<condition>"` (register + arm), `/moai goal status [--all]`, and `/moai goal clear`. Full verb surface, the progression-mode axis, and the safety invariants live in `.claude/skills/moai/workflows/goal.md`.

Availability: `/moai goal` needs hooks enabled (its evaluator IS a Stop hook); it is unavailable when `disableAllHooks` or `allowManagedHooksOnly` is set. It carries no runtime-version floor of its own.

## Goal-Presentation Timing

**`/moai goal` is arm-only.** Arming records the condition in goal state and causes the `stop-goal` evaluator to block turn-end until the condition holds; it starts no work of its own. The consequence is concrete: a goal armed while nothing is running spins idle turns until the ceiling, because each turn-end finds the condition unmet and no work advancing it. Arming is therefore always paired with a work-starting action, and it never substitutes for one. This is why a paste-ready resume keeps a work-starting command (`/moai run SPEC-X`) as Block 5's single primary action rather than a bare goal-arming directive — see `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (Block 5).

**The goal is presented at the Implementation Kickoff Approval gate** — offered there as the **autonomous vs semi-autonomous progression-mode axis**, DISTINCT from the approve/decline decision, and armed only after the gate passes. The axis and its two modes: `.claude/skills/moai/workflows/goal.md` § Progression Mode.

**Arming a goal does not authorize autonomous run-phase entry.** The Implementation Kickoff Approval human gate remains required in both progression modes: the axis selects only what happens AFTER the gate passes, and is never a gate bypass or a relaxation. An armed goal likewise never authorizes creating a PR or a destructive operation — the evaluator decides only whether the turn continues.

<!-- moai:contract-mode-start id="contract-signing-goal" -->
Where `workflow.autonomy.mode: contract` — the Kickoff gate named here is the contract signature checked by `moai contract kickoff-check`; signing never arms a goal by itself. See `.claude/rules/moai/workflow/contract-autonomy.md` § The signing gate.

<!-- moai:contract-mode-end -->
## Hard Preconditions for Every Recommendation

- **Implementation Kickoff Approval comes first**: any run-phase goal-arming is downstream of the Implementation Kickoff Approval human gate (`AskUserQuestion`, plan→run) and never substitutes for or bypasses it. `run.md` § Run-phase Autonomy #1 owns the preferences-drained rationale.
- **Arming is programmatic**: `/moai goal` is orchestrator-invocable, so the orchestrator arms the condition itself once the gate has passed. The native equivalent's evaluator makes no tool calls, so machine-verifiable condition judgment is impossible there — this MoAI subcommand is the only pipeline path.
- **Safety boundary unchanged**: an armed goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary.
- **`run.md` "set" shorthand**: `run.md` § Run-phase Autonomy states the orchestrator MAY set the `ac_converge` goal — the orchestrator arms it via `/moai goal` after the gate passes.

## Cross-references

- `.claude/skills/moai/workflows/goal.md` — verb surface, progression-mode axis, semi-autonomous checkpoint flow, safety invariants
- `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy — the `ac_converge` condition wiring
- `goal-directive-detail.md` — the lazy companion. Load it for § Comparing Autonomous-Continuation Approaches · § Writing an Effective Condition · § Arming Under Multi-Session Concurrency · § Trigger condition templates · § Proactive Recommendation Triggers · § MoAI Integration Notes · § Native `/goal` Prohibition

---

Version: 2.1.0 (stub reduced; detail moved to `goal-directive-detail.md` lazy companion)
Classification: Evolvable orchestration guidance — applies to autonomous multi-turn continuation
