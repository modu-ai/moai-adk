# Goal Directive (`/moai goal`) — Autonomous Continuation

> Moved to the detail companion: `goal-directive-detail.md` ("Goal Directive (`/moai goal`) — Autonomous Continuation").

## What It Is

> Moved to the detail companion: `goal-directive-detail.md` ("What It Is").

Prefix the condition with `model:` or `cmd:` to declare which tier is meant — without a prefix, a non-English claim is read as a shell command that never exits 0 and blocks every turn until the bound fires.

## Goal-Presentation Timing

**`/moai goal` is arm-only.** Arming records the condition in goal state and causes the `stop-goal` evaluator to block turn-end until the condition holds; it starts no work of its own. The consequence is concrete: a goal armed while nothing is running spins idle turns until the ceiling, because each turn-end finds the condition unmet and no work advancing it. Arming is therefore always paired with a work-starting action, and it never substitutes for one. This is why a paste-ready resume keeps a work-starting command (`/moai run SPEC-X`) as Block 5's single primary action rather than a bare goal-arming directive — see `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (Block 5).

**The goal is presented at the plan→run Kickoff gate** (autonomous by default per the transition in `.claude/rules/moai/workflow/auto-semantics.md` §9.1; the operator form survives for keep-set cases) — offered there as the **autonomous vs semi-autonomous progression-mode axis**, DISTINCT from the approve/decline decision, and armed only after the gate is met. The axis and its two modes: `.claude/skills/moai/workflows/goal.md` § Progression Mode.

**Arming a goal does not authorize autonomous run-phase entry.** The Kickoff gate's evidence standard holds in both progression modes — the gate is met in its default autonomous form (audit-cross evidence + decision record) or the operator form keep-set cases keep — and is never a bypass or a relaxation. An armed goal likewise never authorizes creating a PR or a destructive operation — the evaluator decides only whether the turn continues.

<!-- moai:contract-mode-start id="contract-signing-goal" -->
Where `workflow.autonomy.mode: contract` — the Kickoff gate named here is the contract signature checked by `moai contract kickoff-check`; signing never arms a goal by itself. See `.claude/rules/moai/workflow/contract-autonomy.md` § The signing gate.

<!-- moai:contract-mode-end -->
## Hard Preconditions for Every Recommendation

- **The Kickoff gate comes first**: any run-phase goal-arming is downstream of the plan→run Kickoff gate — autonomous by default (audit-cross evidence + decision record, `.claude/rules/moai/workflow/auto-semantics.md` §9.1; the operator form for keep-set cases) — and never substitutes for or bypasses it. `run.md` § Run-phase Autonomy #1 owns the preferences-drained rationale.
- **Arming is programmatic**: `/moai goal` is orchestrator-invocable, so the orchestrator arms the condition itself once the gate has passed. The native equivalent's evaluator makes no tool calls, so machine-verifiable condition judgment is impossible there — this MoAI subcommand is the only pipeline path.
- **Safety boundary unchanged**: an armed goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary.
- **`run.md` "set" shorthand**: `run.md` § Run-phase Autonomy states the orchestrator MAY set the `ac_converge` goal — the orchestrator arms it via `/moai goal` after the gate passes.

## Cross-references

> Moved to the detail companion: `goal-directive-detail.md` ("Cross-references").

---
