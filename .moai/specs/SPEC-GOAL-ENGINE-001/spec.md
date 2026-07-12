---
id: SPEC-GOAL-ENGINE-001
title: "Goal Engine — /moai goal condition-declared universal agentic loop (MoAI-owned /goal reimplementation)"
version: "0.3.0"
status: in-progress
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/goal, internal/cli, internal/hook, .claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "agentic-core, goal-engine, stop-hook, autonomous-loop, per-session-state, axis-b"
era: V3R6
tier: L
depends_on: [SPEC-ANALYZE-FIRST-ROUTING-001]
amendment_of: SPEC-GOAL-ENGINE-001
sync_commit_sha: 624ae8491
---

# SPEC-GOAL-ENGINE-001 — Goal Engine (`/moai goal`)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial plan-phase authoring. Epic AGENTIC-CORE, SPEC 2 of 3. depends_on SPEC-ANALYZE-FIRST-ROUTING-001. Shared findings in that SPEC's research.md (§C.4 Axis B, §C.5 Stop hooks). |
| 0.2.0 | 2026-07-12 | manager-spec | Plan-phase amendment: add D8 (Autonomous/Semi-autonomous Kickoff Progression Mode) per user mid-turn directive 2026-07-12. REQ-GLE-026..029, AC-GLE-027..034. The Implementation Kickoff Approval gate stays mandatory in both modes (§A.2 C1 clarified; §D.4 reconciled — progression-mode is an opt-in per-goal choice AT the gate, NOT a default-on autonomy switch). Semi-autonomous per-turn confirm flows via orchestrator-bridge (stop-goal hook emits checkpoint-signal JSON; orchestrator runs AskUserQuestion — REQ-GLE-014 preserved). Doc codification in CLAUDE.md / run.md / orchestration-mode-selection.md pinned as reachability ACs. Pending plan re-audit. |
| 0.2.1 | 2026-07-12 | manager-spec | Plan-phase D2 fixes from plan-auditor v0.2.0 audit (PASS 0.90). **D2-1**: enrich the §B.5 semi-autonomous checkpoint JSON with a `failed_conditions: [{cmd, exit, tail}]` array so the orchestrator's confirm AskUserQuestion can surface WHY the goal isn't satisfied; reconcile REQ-GLE-010 ↔ REQ-GLE-028 (the failed-condition+tail mandate applies in BOTH modes — autonomous via plain block `reason`, semi-autonomous via the checkpoint's `failed_conditions`; the two REQs do NOT conflict); amend AC-GLE-029 to assert `failed_conditions` is present when a mechanical condition is failing. **D2-2**: re-anchor AC-GLE-021(a) from the stale `grep -ic "goal evaluator\|goal engine" CLAUDE.md` (baseline 1 — CLAUDE.md:41 already carries "forthcoming goal engine" per ANALYZE-FIRST commit 4d7ec04e4, non-discriminating) to `awk '/^## 2\./,/^## 3\./' CLAUDE.md \| grep -ic "goal evaluator"` (verified baseline 0, discriminating, post ≥ 1). 2 D3 defects (AC-GLE-032/033 OR-regex alignment; AC-GLE-029/030 028a/028b header notation) DEFERRED to run-phase — noted in progress.md only. spec-lint clean. |
| 0.3.0 | 2026-07-12 | manager-spec | **In-place amendment** (status completed → in-progress; `amendment_of` self-ref). Root cause: two shipped deliverables are INERT (reachability gap — a token/promise exists but no code path reaches it, same class as the AC-token-presence-≠-reachability lesson). (1) **Arming path absent** — the `internal/goal` engine + `moai hook stop-goal` evaluator only LOAD + evaluate an ALREADY-armed goal; there is NO `moai goal` CLI to arm one (verified `grep goalCmd internal/cli/` → 0 command hits), so nothing writes `.moai/state/goal/<session-id>.json` despite `goal.md` promising it. The original D7 / REQ-GLE-023 specified only the `internal/cli` HOOK verb (`stop-goal`), never an arming CLI — that omission is the overclaim root. (2) **Orphan-prune unwired** — `internal/goal.PruneOrphans` (REQ-GLE-007) has ZERO call sites (verified). Adds REQ-GLE-030..034 (arm CLI + session-id-consistency + prune wiring) and AC-GLE-035..039 (reachability pins). `resume` verb deferred (§D.6). Pending plan RE-AUDIT (the amendment invalidates the cached plan-auditor PASS). |

## Amendments

> Additive record per `spec-frontmatter-schema.md` § Optional Fields (`amendment_of`).
> Original HISTORY rows above are preserved verbatim; amendment rows append below
> with monotonically increasing version.

### Amendment 0.3.0 (2026-07-12) — arm CLI + prune wiring reachability

- **prior completed version**: 0.2.1
- **prior_completed_sha**: 624ae8491
- **amendment_of**: SPEC-GOAL-ENGINE-001 (self-referential — in-place amendment, NOT a successor SPEC)
- **rationale**: SPEC-GOAL-ENGINE-001 shipped `status: completed` at 0.2.1, but two of
  its deliverables are INERT — a reachability gap (a token/promise exists but no code path
  reaches it), the same failure class as the `feedback_ac_token_presence_not_reachability`
  lesson. (1) The engine (`internal/goal/`: `NewGoal` / `SaveGoal` / `LoadGoal` /
  `ClearGoal`, plus schema / state / prune / evaluate) and the evaluator hook
  (`moai hook stop-goal`, `internal/cli/hook_stop_goal.go`) are built and tested, but the
  hook only LOADS + evaluates + re-saves an ALREADY-armed goal. `.claude/skills/moai/workflows/goal.md`
  PROMISES arming writes `.moai/state/goal/<session-id>.json`, yet there is NO `moai goal`
  CLI command (verified: `grep goalCmd internal/cli/` → 0 command hits; only a `--goal`
  flag string in `handoff.go`). So nothing arms a goal — the headline `/moai goal`
  capability is unreachable. The original D7 / REQ-GLE-023 named only the `internal/cli`
  HOOK verb (`stop-goal`), never an arming CLI; that omission is the overclaim root.
  (2) `internal/goal.PruneOrphans` (REQ-GLE-007) exists but has ZERO call sites (verified
  `grep PruneOrphans internal/ cmd/` non-test → definition only) — "prune orphans at
  session-start" never runs.
- **scope** (affected §C REQ IDs / deliverables):
  - NEW REQ-GLE-030..034 (§C.9) — the `moai goal` arm/status/clear CLI, condition
    parsing, session-id-consistency (`moai session current`, NOT the pid fallback), and
    the session-start `PruneOrphans` wiring.
  - NEW AC-GLE-035..039 (acceptance.md) — reachability pins (CLI registration; the
    make-or-break arm→eval linkage; session-id consistency; prune wired; resume-deferred).
  - NEW exclusion §D.6 — the `resume` verb is deferred (inconsistent with current
    `ClearGoal` delete semantics; see §D.6).
  - PRESERVE (NOT rewritten): the entire existing `internal/goal/` engine, the existing
    `moai hook stop-goal` verb, the existing `goal.md` skill, and REQ-GLE-001..029 +
    AC-GLE-001..034 (all intact — this amendment ADDS, it does not rewrite).

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
mandatory, score-independent, and never bypassed by an armed goal — this
invariant holds in BOTH autonomous and semi-autonomous progression modes (D8,
REQ-GLE-026); the progression-mode axis selects ONLY post-approval progression
behavior (continue autonomously vs. checkpoint-confirm each step), never whether
the gate runs; (C2) all user preferences are collected before the goal is armed
(subagents/goal-turns cannot prompt); (C5) every condition is
transcript/mechanically measurable AND bounded by a turn ceiling.

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
- **D8 — Autonomous/Semi-autonomous Kickoff Progression Mode**: the
  Implementation Kickoff Approval AskUserQuestion offers a progression-mode axis
  (autonomous vs. semi-autonomous) as a DISTINCT axis from the approve/decline
  decision — approval remains required in both modes; the mode selects only
  post-approval progression. **Autonomous mode** = continue across turns without
  per-turn user confirmation until the condition holds or the ceiling is reached
  (existing D3 Stop-hook behavior; no NEW behavioral surface — REQ-GLE-027
  codifies this as the default `progression_mode`). **Semi-autonomous mode** =
  the `stop-goal` hook emits a checkpoint-signal block JSON at each turn boundary
  for ORCHESTRATOR-side AskUserQuestion confirmation — the hook itself does NOT
  call AskUserQuestion (REQ-GLE-014 preserved; the orchestrator bridges the
  boundary per `agent-common-protocol.md` § User Interaction Boundary / Hook
  Invocation Surface orchestrator-translation-responsibility pattern). The
  selected mode is persisted in goal state (`progression_mode`, default
  `autonomous`). The progression-mode semantics + the kickoff-still-mandatory
  invariant SHALL be codified in CLAUDE.md, `run.md`, and
  `orchestration-mode-selection.md` (run-phase doc deliverables pinned as
  reachability ACs AC-GLE-031..033). Cross-file reachability: the 3 doc surfaces
  are SEPARATE pinned ACs (research.md lesson: a behavior in code file A
  undocumented in doctrine file B is unenforceable).

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

### §C.8 D8 — Autonomous/Semi-autonomous Kickoff Progression Mode

> **Safety framing (HARD)**: the progression-mode axis is a CHOICE offered AT
> the Implementation Kickoff Approval gate, NOT a bypass OF the gate. The gate
> remains mandatory in both modes (C1 preserved). The axis selects only
> post-approval progression behavior. REQ-GLE-015 (no destructive bypass) is
> unchanged.

- **REQ-GLE-026** (Event-driven): **When** the orchestrator runs Implementation
  Kickoff Approval, its AskUserQuestion SHALL offer an autonomous-vs-semi-
  autonomous progression-mode choice as a distinct axis from the approve/decline
  decision — approval remains required in both modes; the selected mode SHALL be
  persisted in goal state as `progression_mode` (default `autonomous` when the
  user declines to choose).
- **REQ-GLE-027** (State-driven): **While** a goal is armed in autonomous mode,
  the engine SHALL continue across turns without per-turn user confirmation
  until the condition holds or the ceiling is reached (existing D3 Stop-hook
  behavior — REQ-GLE-010..013 govern the block/verdict mechanics; autonomous
  mode is the default progression and introduces no NEW behavioral surface
  beyond the `progression_mode` state field).
- **REQ-GLE-028** (State-driven): **While** a goal is armed in semi-autonomous
  mode AND the goal is not yet satisfied AND the ceiling is not yet reached, the
  `stop-goal` hook SHALL emit a checkpoint-signal block
  `{"decision":"block","reason":"semi-autonomous checkpoint: orchestrator to
  confirm continuation","mode":"semi-autonomous","failed_conditions":[{"cmd","exit","tail"}],...}`
  for orchestrator-side AskUserQuestion confirmation — the hook itself does NOT
  call AskUserQuestion (REQ-GLE-014 preserved; the orchestrator bridges the
  boundary by reading the checkpoint JSON and running the confirm round). The
  `failed_conditions` array carries the failed-condition + output-tail detail
  so the orchestrator's confirm AskUserQuestion can surface WHY the goal isn't
  satisfied (the generic `reason` label alone is insufficient for an informed
  continue/clear/switch decision). When no mechanical condition is failing
  (e.g., the checkpoint fires because a model condition is not yet satisfied),
  `failed_conditions` is empty `[]` or absent. **REQ-GLE-010 ↔ REQ-GLE-028
  reconciliation**: REQ-GLE-010's failed-condition + output-tail mandate applies
  in BOTH modes — in autonomous mode via the plain block `reason`, in
  semi-autonomous mode via the checkpoint's `failed_conditions` array; the two
  REQs do NOT conflict in semi-autonomous mode (the checkpoint does NOT drop
  the diagnostic — it carries it in a structured field).
- **REQ-GLE-029** (Ubiquitous): The progression-mode semantics (both modes) AND
  the kickoff-still-mandatory invariant SHALL be codified in CLAUDE.md (the
  kickoff / approval-gates section), `.claude/skills/moai/workflows/run.md`
  (co-located with the Run-phase Autonomy section), and
  `.claude/rules/moai/workflow/orchestration-mode-selection.md` (co-located with
  the Implementation Kickoff Approval mandatory-restoration policy). (The doc
  EDITS are run-phase deliverables; each doc surface is pinned as a separate
  reachability AC with a baseline-0 discriminating grep — AC-GLE-031..033.)

### §C.9 Amendment 0.3.0 — arm CLI + prune wiring (reachability fix)

> Notation: GEARS (current). These REQs close the two inert-deliverable reachability
> gaps recorded in § Amendments 0.3.0. Each maps to a reachability AC (AC-GLE-035..039).

- **REQ-GLE-030** (Ubiquitous): The repository shall provide a top-level `moai goal`
  cobra command registered under `rootCmd` (appearing in `moai --help`), living in
  `internal/cli/goal.go`, that REUSES the existing `internal/goal` engine
  (`NewGoal` / `SaveGoal` / `LoadGoal` / `ClearGoal`) and shall NOT reimplement the
  state / schema / prune logic already in `internal/goal/`.
- **REQ-GLE-031** (Ubiquitous): The `moai goal` command shall expose the verbs
  `arm "<condition>"` (register + arm), `status`, and `clear`; the bare
  `moai goal "<condition>"` form MAY alias `arm`. The `resume` verb is EXPLICITLY
  out of scope for this amendment (§D.6) — the arm CLI shall NOT register a runnable
  `resume` subcommand.
- **REQ-GLE-032** (Event-driven): **When** `moai goal arm` parses its condition
  argument, it shall classify the condition by an EXPLICIT rule (the parse branch shall
  be specified, NOT an implicit heuristic): a bare shell-command string — a runnable
  command (e.g. `go test ./... exits 0`) — shall become a mechanical condition
  (`{type:mechanical, cmd, expect_exit}`); a natural-language claim that references the
  conversation transcript (e.g. `all AC rows show PASS in the transcript`) shall become a
  model condition (`{type:model, claim}`). The classification rule is therefore
  `shell-command string → mechanical; transcript-referencing claim → model`. The
  orchestrator MAY additionally pass an explicitly-typed structured condition set when
  arming programmatically, bypassing the string heuristic. Both forms reuse the existing
  schema `Condition` type; the armed goal shall carry the default turn ceiling
  (`ceiling.max_turns == 30`).
- **REQ-GLE-033** (Ubiquitous): The `moai goal` arming path shall resolve the active
  session id via `moai session current` (`resolveCurrentSessionID`, already implemented
  in `internal/cli/session.go`) so the arm CLI and the `moai hook stop-goal` evaluator
  key to the SAME `.moai/state/goal/<session-id>.json` file. **When** a real session id
  is resolvable, the arming path shall NOT fall back to the `WriterPidKey()`
  (`pid-<pid>`) discriminator — a pid-keyed arm file would never be found by the hook
  (which runs in a different PID), making the armed goal unreachable. (The
  `WriterPidKey()` fallback of REQ-GLE-008 stays valid ONLY when no real session id
  resolves.) The `moai goal arm` command MAY accept a `--session <id>` override — used
  for deterministic testing (AC-GLE-036 drives the registered command with `--session X`)
  and for programmatic arming; when the override is absent it resolves via
  `moai session current` as above.
- **REQ-GLE-034** (Event-driven): **When** a session starts, the session-start path
  (`internal/hook/session_start.go`) shall invoke `internal/goal.PruneOrphans` with a
  real call site, feeding the active session IDs read from the active-sessions registry
  (`.moai/state/active-sessions.json`; readers exist in `internal/cli/session.go` /
  `internal/harness/routing/pending.go`). The prune shall be fail-open — a prune error
  shall never block session start.

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
- **Reconciliation with D8 (progression-mode axis)**: the progression-mode axis
  (REQ-GLE-026) is an opt-in PER-GOAL choice offered AT the Implementation
  Kickoff Approval gate — it is NOT a default-on autonomy switch. The default
  `progression_mode` is `autonomous` ONLY when the user has already approved the
  kickoff AND explicitly selected autonomous (or declined to choose, in which
  case the existing D3 behavior applies). The kickoff gate itself is NEVER
  relaxed by the mode selection: approval is required in BOTH modes; the axis
  selects only post-approval progression. This is consistent with §A.2 C1 and
  REQ-GLE-015.

### §D.5 Out of Scope — docs-site 4-locale translation

- Translating the `/moai goal` documentation into docs-site locales is a DEFERRED
  follow-up (plan.md § Deferred), NOT run-phase scope.

### §D.6 Out of Scope — resume verb (deferred)

- The `moai goal resume` verb (best-effort re-arm of a previously cleared goal by
  restoring from `.moai/state/goal/consumed/`) is DEFERRED to a follow-up SPEC and is
  NOT delivered by this amendment. `goal.md` documents `resume` (a carry-over from the
  original REQ-GLE-001 four-verb list), but no working `resume` implementation exists.
- **Rationale**: `resume` is inconsistent with the current `ClearGoal` semantics.
  `ClearGoal` (`internal/goal/state.go`) `os.Remove`s the state file — `clear` DELETES,
  it does not tombstone. `consumed/` is `PruneOrphans`' tombstone (orphan-prune moves
  state there), NOT a `clear` destination — so a goal cleared via `clear` never lands in
  `consumed/` and cannot be resumed from it. Delivering a working `resume` would require
  changing `clear` from a delete to a tombstone-move (a semantic change to the existing,
  tested `ClearGoal` contract), which is out of scope for this reachability fix.
- The run-phase author SHALL annotate the `goal.md` `resume` section as "deferred —
  follow-up SPEC" (a `.claude/` doc edit requiring a template mirror per REQ-GLE-025);
  the arm CLI (REQ-GLE-031) delivers only `arm` / `status` / `clear`.

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
