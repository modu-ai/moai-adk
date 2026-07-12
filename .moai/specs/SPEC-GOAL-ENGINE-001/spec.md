---
id: SPEC-GOAL-ENGINE-001
title: "Goal Engine — /moai goal condition-declared universal agentic loop (MoAI-owned /goal reimplementation)"
version: "0.1.0"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/goal, internal/cli, .claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "agentic-core, goal-engine, stop-hook, autonomous-loop, per-session-state, axis-b"
era: V3R6
tier: L
depends_on: [SPEC-ANALYZE-FIRST-ROUTING-001]
---

# SPEC-GOAL-ENGINE-001 — Goal Engine (`/moai goal`)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial plan-phase authoring. Epic AGENTIC-CORE, SPEC 2 of 3. depends_on SPEC-ANALYZE-FIRST-ROUTING-001. Shared findings in that SPEC's research.md (§C.4 Axis B, §C.5 Stop hooks). |

> **Epic**: AGENTIC-CORE (schema has no `epic:` field; recorded in body). SPEC 2 of 3.
> **Artifact-set note (Tier L, LEAN)**: per the Epic leader's explicit scope
> decision, this SPEC ships 3 core artifacts (spec.md + plan.md + acceptance.md)
> plus progress.md, and cross-references the SHARED `research.md` in
> `SPEC-ANALYZE-FIRST-ROUTING-001/`. The Tier L design content is folded into
> `plan.md` § Technical Design rather than a separate design.md (LEAN artifact
> consolidation for this Epic). `tier: L` is retained for the PASS threshold (0.85)
> and Section A-E delegation requirement.

## §A — Context and Motivation (WHY)

Native `/goal` is a **HUMAN-ONLY** Claude Code command: it sets a session-scoped
completion condition and keeps Claude working across turns until a fast model
confirms the condition holds, but the model CANNOT set it on the user's behalf
(`native-invocation-model.md` `:29`/`:44`; `goal-directive.md` `:49`). The
**Axis B** doctrine (`native-invocation-model.md` `:62`, worked illustration
`:71-73`) records that where the nearest native equivalent is HUMAN-ONLY, a MoAI
subcommand automating that capability inside the pipeline is NOT redundant
reinvention — it is the ONLY pipeline path. The illustration explicitly reserves
this justification for "a future subcommand facing a genuinely HUMAN-ONLY native
counterpart." **`/moai goal` IS that subcommand** (shared research.md §C.4).

This SPEC delivers `/moai goal` — a MoAI-owned, **PROGRAMMATIC** reimplementation
of `/goal` semantics that an orchestrator (or user) can register and arm without
a human typing the native `/goal` line. It is a universal agentic loop: a goal
declares completion conditions; the session iterates any work until the conditions
hold or a ceiling is reached. `SPEC-LOOP-SWEEP-001` builds `/moai loop` as a
preset ON this engine.

### §A.1 Relationship to the native `ac_converge` wiring (HARD, no conflict)

`SPEC-AUTONOMY-RUN-GOAL-001` (completed) wraps the **native** `/goal` at the
run-phase boundary via `.claude/skills/moai/workflows/run.md § Run-phase Autonomy
(/goal ac_converge)`. This SPEC does NOT modify that section. `/moai goal`
(programmatic) and the `run.md ac_converge` native-`/goal` wrapping are DISTINCT
sibling surfaces (research.md §D.1). REQ-GLE-006 references the `run.md` wiring
point read-only.

### §A.2 Safety invariants inherited (HARD)

`/moai goal` MUST preserve the same safety envelope as native `/goal` and the
AUTONOMY-RUN-GOAL 6 conditions: (C1) Implementation Kickoff Approval is
mandatory, score-independent, and never bypassed by an armed goal; (C2) all user
preferences are collected before the goal is armed (subagents/goal-turns cannot
prompt); (C5) every condition is transcript/mechanically measurable AND bounded
by a turn ceiling.

## §B — Scope (WHAT this SPEC delivers)

- **D1 — `/moai goal` workflow + router registration**: a new
  `.claude/skills/moai/workflows/goal.md` workflow file AND registration in the
  `/moai` SKILL.md Intent Router P1 subcommand list AND the Workflow Quick
  Reference entry. Verbs: `"<condition>"` (register + arm), `status [--all]`,
  `clear`, `resume`. Cross-file reachability: SKILL.md P1 registration and the
  Quick Reference entry are SEPARATE pinned ACs (research.md lesson: a feature in
  file A unregistered in router file B is inert).
- **D2 — Per-session goal state**: `.moai/state/goal/<session-id>.json` (NOT a
  single shared file — multi-session write race). Schema: goal text; `conditions[]`
  (each `{type:mechanical, cmd, expect_exit}` or `{type:model, claim}`); `ceiling
  {max_turns default 30}`; `progress` (append-only); `session_id`; `created_at`;
  `status`. Atomic write (temp + rename). Orphan pruning at session-start (absent
  from `active-sessions.json` OR TTL-expired → `.moai/state/goal/consumed/`).
  `writer_pid` fallback when session id is unavailable (reuse the `moai session
  current` fallback contract).
- **D3 — Hybrid 2-tier Stop-hook evaluator**: a new Go verb `moai hook stop-goal`
  plus a new `handle-stop-goal.sh` wrapper (settled — new wrapper, clean
  composition with the HARNESS-EVOLVE observer). Tier 1: run mechanical
  conditions; any FAIL → exit 0 stdout
  `{"decision":"block","reason":"<failed condition + output tail>"}`. Tier 2:
  only when ALL mechanical conditions PASS AND model conditions exist → model
  judgment. Ceiling: a turns counter in state; at the ceiling emit a 5-section
  verdict (`verification-claim-integrity.md` §3) and stop blocking. Goal-eval may
  exceed the MoAI 5s hook policy — the plan addresses a per-hook timeout override.
- **D4 — Safety wiring**: a goal never bypasses Implementation Kickoff Approval;
  an active native `/goal` → `stop-goal` yields (pass-through; detection settled —
  degrade-to-always-evaluate when the runtime does not expose the native-goal
  signal, recorded as accepted DEBT in plan.md); a stagnation guard (N no-progress
  iterations → stop + E1/E3 escalation note); destructive-action confirmation
  unchanged.
- **D5 — Handoff + doctrine integration**: `session-handoff.md` Block 5 MAY carry
  `Run: /moai goal "<condition>"`; the post-paste native-`/goal` follow-up is
  demoted to an optional variant. `goal-directive.md` gains a `/moai goal` row
  (PROGRAMMATIC MoAI counterpart + Axis B citation). `native-invocation-model.md`
  Axis B illustration updates ("MoAI does not currently reimplement any of the
  three" → now `/moai goal` does).
- **D6 — Analyze-First integration**: the `SPEC-ANALYZE-FIRST-ROUTING-001` §2
  pipeline stage ③ MAY register a derived completion condition as a goal; stage ⑤
  termination judge = the goal evaluator. Document the boundary with the
  `moai.md` Agentic Completion Loop (phase-granular pipeline loop vs task-granular
  goal engine). `agentic_loop_distinctness_test.go` stays green.
- **D7 — Go packages**: minimal layout — `internal/goal/` (state + schema +
  evaluator), `internal/cli` hook verb, `settings.json.tmpl` Stop-hook
  registration. Coverage 85%+ (critical-package policy).

## §C — GEARS Requirements

> Notation: GEARS (current). `<subject>` generalized. Each REQ maps to an AC.

### §C.1 D1 — `/moai goal` workflow + reachable registration

- **REQ-GLE-001** (Ubiquitous): The repository shall contain a new workflow file
  `.claude/skills/moai/workflows/goal.md` defining the `/moai goal` subcommand and
  its four verbs (`"<condition>"`, `status [--all]`, `clear`, `resume`).
- **REQ-GLE-002** (Ubiquitous): The `/moai` SKILL.md Intent Router P1 subcommand
  list shall register `goal` as a routed subcommand (router-registration surface —
  distinct from REQ-GLE-001's workflow-file surface).
- **REQ-GLE-003** (Ubiquitous): The `/moai` SKILL.md Workflow Quick Reference
  shall carry a `goal` entry (Quick-Reference surface — distinct from REQ-GLE-002's
  P1-list surface; both must be present for the feature to be reachable).

### §C.2 D2 — Per-session goal state

- **REQ-GLE-004** (Ubiquitous): The goal engine shall persist state to
  `.moai/state/goal/<session-id>.json` — one file per session — and shall not use
  a single shared state file (multi-session write-race avoidance).
- **REQ-GLE-005** (Ubiquitous): The goal state schema shall carry goal text,
  `conditions[]` (each either `{type:mechanical, cmd, expect_exit}` or
  `{type:model, claim}`), `ceiling {max_turns}` defaulting to 30, an append-only
  `progress` log, `session_id`, `created_at`, and `status`.
- **REQ-GLE-006** (Event-driven): **When** the goal engine writes state, it shall
  write atomically (temp file + rename), never a partial in-place write.
- **REQ-GLE-007** (Event-driven): **When** a session starts, the goal engine shall
  prune orphan state files (session absent from `active-sessions.json` OR TTL
  expired) by moving them to `.moai/state/goal/consumed/`.
- **REQ-GLE-008** (Capability gate): **Where** the runtime does not expose a
  session id, the goal engine shall fall back to a `writer_pid` discriminator
  (reusing the `moai session current` fallback contract).

### §C.3 D3 — Hybrid 2-tier Stop-hook evaluator

- **REQ-GLE-009** (Ubiquitous): The repository shall provide a Go Stop-hook verb
  `moai hook stop-goal` that evaluates the active session's goal on turn-end.
- **REQ-GLE-010** (Event-driven): **When** any mechanical condition fails
  (Tier 1), `stop-goal` shall exit 0 with stdout JSON
  `{"decision":"block","reason":"<failed condition + output tail>"}` (continuing
  the turn per Claude Code hook semantics).
- **REQ-GLE-011** (State-driven): **While** all mechanical conditions pass AND at
  least one model condition exists, `stop-goal` shall gate Tier-2 evaluation so it
  is reached only after all mechanical conditions pass, surfacing the model claim
  in the block `reason` for **orchestrator** evaluation against
  conversation-surfaced evidence (settled Option B — orchestrator self-eval;
  provider-agnostic incl. GLM; `stop-goal` itself does not run a model call).
- **REQ-GLE-012** (Event-driven): **When** all conditions pass (mechanical and,
  where present, model), `stop-goal` shall NOT emit a block decision (the goal is
  satisfied; the turn ends and the goal clears).
- **REQ-GLE-013** (Event-driven): **When** the turns counter reaches the ceiling,
  `stop-goal` shall stop blocking and emit a 5-section verdict
  (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
- **REQ-GLE-014** (Ubiquitous): The `stop-goal` hook shall not invoke
  `AskUserQuestion` or any user-prompting tool (subagent/hook boundary — it emits
  structured JSON only; grep-verifiable).

### §C.4 D4 — Safety wiring

- **REQ-GLE-015** (State-driven): **While** a goal is armed, the goal shall not be
  treated as authorization to bypass Implementation Kickoff Approval, create a PR,
  or perform any destructive operation.
- **REQ-GLE-016** (Event-driven): **When** an active native `/goal` is detected,
  `stop-goal` shall yield (pass-through — not double-block the turn).
- **REQ-GLE-017** (Event-driven): **When** N consecutive no-progress iterations
  are observed (stagnation), `stop-goal` shall stop blocking and record an
  E1/E3-escalation note in the 5-section verdict.

### §C.5 D5 — Handoff + doctrine integration

- **REQ-GLE-018** (Ubiquitous): `session-handoff.md` shall document that, where the
  next SPEC declares a machine-verifiable end-state, the Block 5 `Run:` line may
  carry `/moai goal "<condition>"`, and shall demote the post-paste native-`/goal`
  follow-up to a documented optional variant.
- **REQ-GLE-019** (Ubiquitous): `goal-directive.md` shall carry a `/moai goal`
  entry describing it as the PROGRAMMATIC MoAI counterpart of native `/goal`, with
  an Axis B citation.
- **REQ-GLE-020** (Ubiquitous): The `native-invocation-model.md` Axis B worked
  illustration shall be updated to reflect that `/moai goal` now reimplements the
  HUMAN-ONLY `/goal` capability programmatically.

### §C.6 D6 — Analyze-First integration

- **REQ-GLE-021** (Ubiquitous): The `SPEC-ANALYZE-FIRST-ROUTING-001` §2 pipeline
  stage ⑤ termination judge shall reference the goal evaluator (clause a), AND the
  goal engine shall document, in `.claude/skills/moai/workflows/moai.md` § Agentic
  Completion Loop, the boundary between the phase-granular Agentic Completion Loop
  and the task-granular goal engine (clause b). (The Agentic Completion Loop lives
  in `.claude/skills/moai/workflows/moai.md`, NOT `.claude/output-styles/moai/moai.md`.)
- **REQ-GLE-022** (Unwanted behavior): The goal engine and its config shall not
  collapse the `workflow.agentic_loop.max_iterations` vs
  `loop_prevention.max_iterations` distinctness guarded by
  `internal/config/agentic_loop_distinctness_test.go` (the test stays green).

### §C.7 D7 — Go packages + quality

- **REQ-GLE-023** (Ubiquitous): The goal-engine Go code shall live in a minimal
  layout: `internal/goal/` (state + schema + evaluator) and the `internal/cli`
  hook verb, with the Stop hook registered in `settings.json.tmpl`.
- **REQ-GLE-024** (Ubiquitous): The `internal/goal/` package shall reach ≥ 85%
  statement coverage (critical-package policy).
- **REQ-GLE-025** (Event-driven): **When** the `.claude/` files or
  `settings.json.tmpl` change, the corresponding `internal/template/templates/`
  mirrors shall be updated and `make build` shall succeed; template bodies shall
  carry no internal SPEC ID (§25 neutrality).

## §D — Exclusions (What NOT to Build)

[HARD] This SPEC explicitly does NOT deliver the following.

### §D.1 Out of Scope — modifying the native `ac_converge` wiring

- The `.claude/skills/moai/workflows/run.md § Run-phase Autonomy (/goal
  ac_converge)` section (owned by `SPEC-AUTONOMY-RUN-GOAL-001`) is NOT modified.
  `/moai goal` is a sibling surface, not a rewrite of run-phase native-`/goal`.

### §D.2 Out of Scope — the `/moai loop` redefinition

- Redefining `/moai loop` as a goal preset is owned by `SPEC-LOOP-SWEEP-001`. This
  SPEC ships only the reusable goal engine; the loop preset is downstream.

### §D.3 Out of Scope — Go `moai loop` / `internal/ralph` unification

- Unifying the goal engine with the Go `moai loop` SPEC-lifecycle controller
  (`internal/cli/loop.go`, `internal/loop`, `internal/ralph`) is out of scope
  (research.md §C.1); those surfaces are untouched here.

### §D.4 Out of Scope — enabling autonomy by default

- `/moai goal` is opt-in. This SPEC does NOT arm any goal automatically, does NOT
  flip a default-on autonomy switch, and does NOT relax the Implementation Kickoff
  Approval gate.

### §D.5 Out of Scope — docs-site 4-locale translation

- Translating the `/moai goal` documentation into docs-site locales is a DEFERRED
  follow-up (plan.md § Deferred), NOT run-phase scope.

## §E — Dependencies and Follow-ups

- **depends_on**: `SPEC-ANALYZE-FIRST-ROUTING-001` (frontmatter) — the §2 pipeline
  stage ⑤ goal-evaluator reference (REQ-GLE-021) requires that SPEC's §2 rewrite.
  Per the Depends_on pre-flight, run-phase entry requires ANALYZE-FIRST to be
  `completed` (or an explicit `--ignore-deps` override).
- **Blocks**: `SPEC-LOOP-SWEEP-001` (the loop preset is built on this engine).
- **Coordination (no hard block)**: `SPEC-HARNESS-EVOLVE-001` also adds a Stop-hook
  surface; both register as SEPARATE Stop-hook entries in `settings.json.tmpl`
  (Stop hooks compose — see plan.md § Dependencies).
- **Doctrine anchors**: shared `research.md` §C.4/§C.5/§D; `goal-directive.md`;
  `native-invocation-model.md`; `session-handoff.md`;
  `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary /
  § Hook Invocation Surface.
