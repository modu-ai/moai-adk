---
id: SPEC-AUTONOMY-RUN-GOAL-001
title: "Run-phase autonomy — Mode 6 (workflow) + /goal ac_converge wrapping with GATE-2 preservation"
version: "0.1.0"
status: in-progress
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "v3.x"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "autonomy, workflow, goal, run, gate2"
tier: M
---

# SPEC-AUTONOMY-RUN-GOAL-001 — Run-phase Autonomy: Mode 6 (workflow) + `/goal` `ac_converge` with GATE-2 Preservation

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring. Priority-High entry of the autonomous-workflow integration line. Run group design authored FRESH (original was rate-limited, not content-rejected per `.moai/docs/autonomous-workflow-strategy.md` §4.2-B) and verified against the verdict's 6 safety conditions. |

## §A — Context and Motivation (WHY)

Two Claude Code orchestration primitives — **Dynamic Workflows** (a script that fans out dozens-to-hundreds of agents; v2.1.154+) and **`/goal`** (a session-scoped completion condition that continues turns until a fast model confirms; v2.1.139+) — are documented at the doctrine layer (`.claude/rules/moai/workflow/dynamic-workflows.md`, `.claude/rules/moai/workflow/goal-directive.md`) but are **not wired into any `/moai` subcommand**. The approved strategy report (`.moai/docs/autonomous-workflow-strategy.md`) identifies this as the core gap and decomposes the closure into a roadmap. This SPEC is the **Priority-High `run` entry** of that roadmap (`SPEC-AUTONOMY-RUN-GOAL`).

The motivation is to grant **phase-internal autonomy** to `/moai run` while **preserving the phase-boundary human gate (GATE-2)** as an absolute invariant. GATE-2 (plan→run) is a mandatory `AskUserQuestion` human approval that is never auto-bypassed regardless of plan-auditor score (CLAUDE.local.md §19.1 / REQ-ATR-015). Granting `/goal` autonomy (which removes per-turn STOP prompts) and adding a Mode 6 (workflow) fan-out option MUST NOT erode this gate. This SPEC delivers run-phase autonomy in a way that demonstrably preserves GATE-2.

### §A.1 The autonomous-strategy verdict's 6 safety conditions (HARD)

The strategy report flags that the run group's original design was rate-limited (a missing-input verdict, **not** a content rejection). This SPEC authors the run design FRESH and MUST satisfy all 6 conditions verbatim:

1. **C1 — GATE-2 mandatory, score-independent**: GATE-2 is a mandatory `AskUserQuestion` human gate; plan-auditor PASS or score ≥ 0.90 never auto-bypasses it. The skip-eligible ≥ 0.90 autonomous bypass applies ONLY to Phase 0.5 plan-auditor verdict re-execution, NOT to GATE-2.
2. **C2 — Preferences collected before launch**: all user preferences are collected by the orchestrator via `AskUserQuestion` BEFORE any Workflow or subagent launch. Workflow agents and `/goal`-turn agents cannot prompt the user (asymmetric boundary).
3. **C3 — No subagent-spawns-subagent**: only the orchestrator (main session) spawns. Mode 6 Workflow is a scaling mechanism, NOT nested spawning (Finding A1).
4. **C4 — Background agents read-only**: `Agent(run_in_background: true)` MUST NOT Write/Edit; implementation writes are foreground. (Workflow agents are a separate primitive that run in `acceptEdits`; the background-write prohibition applies to background `Agent()` only.)
5. **C5 — Every `/goal` condition transcript-measurable AND bounded**: each condition is measured against orchestrator-surfaced transcript lines (the evaluator does NOT run tools or read files) and carries an explicit `max N turns` bound.
6. **C6 — Workflow only for genuinely-parallel high-volume mechanical work**: deterministic diagnostics stay in `/moai loop`; coding-heavy work stays Mode 5 sequential sub-agent (Finding A4).

Plus the **named-script-API prohibition**: this SPEC MUST NOT assert a typed named-script Workflow API (e.g., `agent()`/`parallel()`/`pipeline()`/`phase()` functions). The official Claude Code docs do not document one. The design describes only the conceptual *coordinate-agents → results in script variables → final synthesis* model.

## §B — Scope (WHAT this SPEC delivers)

This SPEC delivers EXACTLY three deliverables. It is doctrine/rules-focused with low Go-code volume.

- **D1 — Mode 6 (workflow) addition** to the Phase 0.95 mode catalog in `.claude/rules/moai/workflow/orchestration-mode-selection.md` (currently a 5-mode catalog: trivial / background / agent-team / parallel / sub-agent). Strict entry conditions; selectable only AFTER GATE-2 has already passed.
- **D2 — Run-phase `/goal` wrapping** of the `ac_converge` condition (strategy §5.2). This SPEC **adds a NEW self-contained `## Run-phase Autonomy (/goal ac_converge)` section to the router file `.claude/skills/moai/workflows/run.md`** that physically co-locates BOTH (a) the GATE-2 `AskUserQuestion` ordering reference AND (b) the `/goal ac_converge` set, so the orchestrator sets the goal AFTER GATE-2 approval and both ordering markers live in ONE file. The condition predicate is transcript-measurable (never a file-path predicate).
  - **Empty-router consequence (read before implementing)**: today `.claude/skills/moai/workflows/run.md` is a thin Phase-routing router with ZERO GATE-2 / `/goal` content (a tree-wide grep for `GATE-2` over the run skill tree returns 0 matches; the actual Phase-0.5 gate logic lives in the `run/phase-execution.md` sub-skill, which EX-5 excludes from modification). Therefore the GATE-2 ordering reference AND the `/goal` set markers that D3's regression guard searches for **do not exist yet** — D2 is the deliverable that introduces both into `run.md`. The run-phase author MUST add this NEW section rather than expecting pre-existing markers.
- **D3 — GATE-2 preservation regression test(s)** proving run-phase entry still requires explicit `AskUserQuestion` human approval regardless of plan-auditor score (even ≥ 0.90 skip-eligible), and that `/goal`/Mode-6 launch cannot cross the plan→run boundary. The guard searches the NEW `run.md` autonomy section introduced by D2 (both ordering markers are co-located there per the D2 self-contained-section design).

## §C — GEARS Requirements

> Notation: GEARS (current). `<subject>` is generalized. Each REQ is traceable to an AC in `acceptance.md`.

### §C.1 D1 — Mode 6 (workflow) catalog addition

- **REQ-ARG-001** (Ubiquitous): The `orchestration-mode-selection.md` rule shall define exactly one additional execution mode, `workflow` (Mode 6), appended to the existing 5-mode catalog without removing or renumbering modes 1–5.
- **REQ-ARG-002** (Capability gate): **Where** the run-phase scope is ≥ ~30 files AND the transformation is mechanical (call-site rename, import-path bulk change, or equivalent) AND the work is genuinely parallel (no inter-file dependency), the orchestrator shall treat Mode 6 (workflow) as a selection candidate.
- **REQ-ARG-003** (State-driven): **While** the run-phase work is coding-heavy or multi-domain, the orchestrator shall prefer Mode 5 (sub-agent sequential) over Mode 6, citing the Finding A4 caveat ("most coding tasks involve fewer truly parallelizable tasks than research").
- **REQ-ARG-004** (Event-driven): **When** the orchestrator selects Mode 6 in Phase 0.95, the orchestrator shall record the selection AND a confirmation that GATE-2 already passed AND that all preferences were collected, in `.moai/specs/SPEC-{ID}/progress.md` under the `Mode Selection` section, before launching the Workflow.
- **REQ-ARG-005** (Unwanted behavior, event-detected): **When** a Mode 6 launch is attempted before GATE-2 has passed, the orchestrator shall not launch the Workflow and shall return control to the GATE-2 `AskUserQuestion` gate.

### §C.2 D2 — Run-phase `/goal` `ac_converge` wrapping

- **REQ-ARG-006** (Ubiquitous): The router file `.claude/skills/moai/workflows/run.md` shall define, in a NEW self-contained `## Run-phase Autonomy (/goal ac_converge)` section, the `ac_converge` `/goal` condition wiring point such that the goal is set ONLY after GATE-2 approval is obtained. The GATE-2 `AskUserQuestion` ordering reference and the `/goal ac_converge` set MUST be co-located in this same `run.md` section.
- **REQ-ARG-007** (Ubiquitous): The `ac_converge` condition shall be transcript-measurable — every predicate references a line the orchestrator surfaces in the conversation (test output, build exit code, explicit `AC-id: PASS` line, `git status` output), never a file-path predicate the `/goal` evaluator would have to read.
- **REQ-ARG-008** (Ubiquitous): The `ac_converge` condition shall carry an explicit bound of `max 20 turns`.
- **REQ-ARG-009** (Event-driven): **When** a semantic failure (data race, deadlock, panic, or test assertion failure) is surfaced during the run-phase autonomous loop, the orchestrator shall clear the `/goal` and escalate via `AskUserQuestion` rather than auto-fixing the semantic failure (CONST-V3R5-010 alignment).
- **REQ-ARG-010** (State-driven): **While** the `ac_converge` goal is active, the orchestrator shall not treat the goal as authorization to bypass GATE-2, create a PR, or perform any destructive operation — these remain explicit gates surfaced separately.

### §C.3 D3 — GATE-2 preservation regression

- **REQ-ARG-011** (Ubiquitous): The acceptance criteria shall include a verifiable regression test proving that run-phase entry emits a GATE-2 `AskUserQuestion` human-approval gate before any `/goal` set or Mode-6 launch, independent of the plan-auditor score.
- **REQ-ARG-012** (Capability gate): **Where** the plan-auditor verdict is PASS with score ≥ 0.90 (skip-eligible), the `.claude/skills/moai/workflows/run.md` autonomy section shall still emit the GATE-2 `AskUserQuestion` gate — skip-eligibility applies only to Phase 0.5 verdict re-execution, not to GATE-2.
- **REQ-ARG-013** (Ubiquitous): The regression verification shall assert (via grep/rule cross-reference) that the `.claude/skills/moai/workflows/run.md` autonomy section emits the GATE-2 `AskUserQuestion` ordering before any `/goal` set, and that a doctrine cross-reference to CLAUDE.local.md §19.1 / REQ-ATR-015 is present.

### §C.4 The 6 HARD safety conditions as cross-cutting constraints

- **REQ-ARG-014** (Ubiquitous): The delivered rule and skill-body content shall preserve all 6 safety conditions of §A.1 (C1–C6) AND the named-script-API prohibition; no delivered text shall assert a typed named-script Workflow API.
- **REQ-ARG-015** (Event-driven): **When** a Mode-6 Workflow or `/goal`-turn agent lacks a required input, that agent shall return a structured blocker report; the orchestrator shall run an `AskUserQuestion` round and re-delegate (agents never prompt the user — asymmetric boundary, C2/C3).
- **REQ-ARG-016** (Ubiquitous): The template-first mirror obligation (CLAUDE.local.md §2) shall be satisfied — every modified `.claude/` file shall have its corresponding `internal/template/templates/` mirror updated and `make build` run, verified at run-phase.

## §D — Exclusions (What NOT to Build)

[HARD] This SPEC explicitly does NOT deliver the following. Each is owned by a sibling roadmap SPEC or is deliberately out of scope.

### §D.1 Out of Scope (Exclusions — What NOT to Build)

- **EX-1 — `workflow.yaml` autonomy-profile schema (Go nested struct + accessors)**: owned by `SPEC-AUTONOMY-CONFIG` (strategy §8 / §9 Priority-Med). This SPEC references the `run` profile's `goal_condition_template: "ac_converge"` and `max_turns: 20` as design intent only; it does NOT implement the `internal/config` struct.
- **EX-2 — Workflow script pattern catalog (5 patterns)**: owned by `SPEC-AUTONOMY-WORKFLOW-PATTERNS` (strategy §6 / §9 Priority-Med). This SPEC describes only the single run-phase mechanical-migration fan-out shape conceptually; it does NOT register the full pattern catalog in `dynamic-workflows.md`.
- **EX-3 — Full `/goal` condition template registry (all subcommands)**: owned by `SPEC-AUTONOMY-GOAL-CONDITIONS` (strategy §5 / §9 Priority-High sibling). This SPEC wires only the `run` condition (`ac_converge`); it does NOT register plan/sync/loop/coverage/review/mx/clean conditions.
- **EX-4 — Other subcommands' autonomy wiring**: plan, sync, fix, loop, review, coverage, e2e, mx, codemaps, clean, brain, design, db autonomy wiring is out of scope (owned by their respective roadmap SPECs).
- **EX-5 — Rewriting the legacy run sub-skill agent chain / Phase 0.5 gate in `phase-execution.md`**: the run sub-skill `.claude/skills/moai/workflows/run/phase-execution.md` contains pre-existing drift (archived-agent references `manager-strategy`/`manager-quality`, a stale 5-mode table predating `orchestration-mode-selection.md`, and the legacy Phase 0.5 / Decision-Point gate logic). Rewriting that legacy agent chain or the Phase 0.5 gate in `phase-execution.md` is NOT in scope. **What IS in scope (not excluded by EX-5)**: adding a NEW self-contained `## Run-phase Autonomy (/goal ac_converge)` section to the router file `.claude/skills/moai/workflows/run.md` (D2) — this is new additive content in a DIFFERENT file (`run.md`, the router), not a modification of the legacy `phase-execution.md` gate. The Mode 6 catalog addition (D1) lands in the canonical rule `orchestration-mode-selection.md`. The empty-router consequence (no pre-existing GATE-2/`/goal` markers in `run.md`) is acknowledged in §B D2 so the run-phase author adds — rather than edits — the markers.
- **EX-6 — A typed/named Workflow script API**: deliberately excluded per the named-script-API prohibition. No `agent()`/`parallel()`/`pipeline()`/`phase()` function signatures are asserted.
- **EX-7 — Enabling autonomy by default**: `autonomy.enabled` remains off by default (research preview). This SPEC does NOT flip any default-on switch; Mode 6 / `/goal` are opt-in and gated behind capability + version preflight (owned by EX-1's config SPEC).
- **EX-8 — Background-agent or Agent-Teams changes**: no modification to the existing Mode 2 (background) or Mode 3 (agent-team) behavior.

## §E — Dependencies and Follow-ups

- **Follow-up (not a blocker)**: `SPEC-AUTONOMY-CONFIG` (workflow.yaml schema), `SPEC-AUTONOMY-WORKFLOW-PATTERNS` (pattern catalog), `SPEC-AUTONOMY-GOAL-CONDITIONS` (full condition registry). This SPEC is authored to stand alone: it hard-codes the single `run` `ac_converge` condition inline so it does NOT block on the config schema or the condition registry.
- **Doctrine anchors (read, do not restate)**: `.claude/rules/moai/workflow/goal-directive.md`, `.claude/rules/moai/workflow/dynamic-workflows.md`, `.claude/rules/moai/workflow/orchestration-mode-selection.md`, CLAUDE.local.md §19.1 (REQ-ATR-015), `.claude/rules/moai/core/agent-common-protocol.md` (User Interaction Boundary, Pre-Spawn Sync Check), `.claude/rules/moai/core/askuser-protocol.md`.

## §F — Constraints Summary (cross-reference)

| Constraint | Source | Enforced by |
|------------|--------|-------------|
| GATE-2 score-independent (C1) | CLAUDE.local.md §19.1 / REQ-ATR-015 | REQ-ARG-005, REQ-ARG-011, REQ-ARG-012, REQ-ARG-013 |
| Preferences-before-launch (C2) | dynamic-workflows.md §MoAI Integration | REQ-ARG-004, REQ-ARG-015 |
| No nested spawn (C3) | Finding A1 / agent-common-protocol.md | REQ-ARG-014, REQ-ARG-015 |
| Background read-only (C4) | CONST-V3R2-020 / -044 | REQ-ARG-014 (design narrative) |
| Transcript-measurable + bounded goal (C5) | goal-directive.md §Writing an Effective Condition | REQ-ARG-007, REQ-ARG-008 |
| Workflow-parallel-only / loop-diagnostic (C6) | Finding A4 / dynamic-workflows.md | REQ-ARG-002, REQ-ARG-003 |
| No named-script API | dynamic-workflows.md (no documented API) | REQ-ARG-014, EX-6 |
| Template-first mirror | CLAUDE.local.md §2 | REQ-ARG-016 |

### §F.1 Two-axis confusion warning (D1 wiring author)

[HARD] "Mode 6 (workflow)" added by D1 is appended to the **Phase 0.95 execution-mode catalog** in `orchestration-mode-selection.md` — a 5-item list (`trivial` / `background` / `agent-team` / `parallel` / `sub-agent`). This is a DIFFERENT axis from the `run.md` `--mode` **dispatch axis** — also a (separate) ~5-item list (`autopilot` / `loop` / `team` / `pipeline` / `background`) documented in `spec-workflow.md` § Subcommand Classification and the `run.md` Mode Dispatch table. The two lists are independent: the execution-mode catalog (Phase 0.95) governs HOW the orchestrator spawns (concurrency/spawn-surface); the `--mode` dispatch axis governs WHICH workflow variant runs (`autopilot` vs `loop` vs `team`). The wiring author MUST NOT conflate them — Mode 6 (`workflow`) is added ONLY to the execution-mode catalog; it is NOT a new `--mode` dispatch value, and no `--mode workflow` flag is introduced. `orchestration-mode-selection.md`'s own header cross-reference already notes the two axes "interact with — but are separate from" each other.
