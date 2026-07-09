---
id: SPEC-LOOP-VERDICT-CONTRACT-001
title: "Mechanical Loop Termination Predicate and Ceiling-Exit Verdict Contract for Utility Loops"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "loop, ralph, termination-predicate, verdict-contract, ceiling-exit, max-iterations-precedence, workflow-reflex"
---

# SPEC-LOOP-VERDICT-CONTRACT-001 — Mechanical Loop Termination Predicate and Ceiling-Exit Verdict Contract

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **3 of 6**.

- **Dependency notes**: SPEC 1 (SPEC-HARNESS-RATCHET-REWIRE-001) and SPEC 2 (SPEC-MODEL-ROUTING-WIRE-001) are independent of each other; this SPEC is independent of both. Downstream SPEC-ADVISOR-RUNG-001 depends on SPEC 2, NOT on this SPEC.
- **Tier**: M (standard) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| L3 | MED | `/moai loop` success-exit rides a self-emitted prose sentinel — builder both performs and declares verification; no tool-observed predicate, no independent evaluator (contrast: `/goal` uses a separate Haiku evaluator; harness `minimal` explicitly sets `evaluator: false`) |
| L4 | MED | Ceiling exits have no verdict contract — loop.md Step 9 only "Display remaining issues and options"; moai.md agentic-loop termination cause 2 (max-iterations) has no protocol while causes 3 (escalation) and 4 (context suspension) do |
| L5 | MED | `/moai fix` Phase 4 re-runs the same scanners as the builder — partial mechanical independence only (recorded as provenance; deliberately Out of Scope here — see §Out of Scope) |
| L6 | MED | max_iterations fragmentation with no precedence rule: loop.md flag default 100; workflow.yaml loop_prevention 100; ralph.yaml loop.max_iterations 10; workflow.yaml agentic_loop 10; memory-safe cap 50 — the Go-loaded typed config says 10 while the skill body says 100 |
| L8 | LOW | Remaining issues evaporate at loop exit — Level-4 manual items and ceiling leftovers land only in the transcript, never persisted; bonus doc drift: workflow.yaml agentic_loop comment "no Go-side loader field yet" is stale (AgenticLoopConfig loader implemented) |

---

## User Story

**As a** user running `/moai loop` (the Ralph diagnostic fix-loop) on a codebase with residual errors,
**I want** the loop's success declared only by a mechanical predicate over parsed scanner diagnostics — double-checked by an independent final pass — and every ceiling exit to emit a structured 5-section evidence verdict whose remaining issues persist to a state file,
**so that** the loop can never talk itself into success by emitting its own completion sentence, ceiling exits leave an auditable, resumable backlog instead of transcript-only residue, and the iteration ceiling I configure is the one that actually applies.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read. Line numbers indicative; content anchors are authoritative.

### GAP-1 — Self-emitted prose sentinel as success-exit (L3)

- **Measured source**: `.claude/skills/moai/workflows/loop.md` — Step 1 (Completion Check): *"Check whether the previous iteration's response declared loop completion in natural language. Completion sentence: 'All loop completion conditions satisfied; exiting loop.' If the completion sentence is present: Exit loop with success"*; Step 4 (Completion Condition Check): *"If all conditions met: Emit the completion sentence ... so Step 1 of the next iteration detects success-exit"*; § Completion Conditions restates sentinel detection as an exit condition.
- **Observed pattern**: The model outputs the sentence, then string-matches its own sentence next iteration. The builder both performs and declares verification — a direct tension with `verification-claim-integrity.md` §1.1 (no unobserved verification claim) and the repo's own separations: `/goal` uses a separate Haiku evaluator (goal-directive.md: *"a small fast model (Haiku by default) checks whether the condition holds"*), and harness.yaml `minimal` explicitly models evaluator presence as a switch (`evaluator: false`).
- Note: Step 4 DOES check mechanical conditions (zero errors / tests passing / coverage) before emitting the sentence — but the EXIT decision at Step 1 keys on the sentence string, not on re-verified diagnostics, so a hallucinated or stale sentence exits the loop.

### GAP-2 — No ceiling-exit verdict contract (L4)

- **Measured source**: loop.md Step 9: *"If max iterations reached: Display remaining issues and options"* (nothing more); `.claude/skills/moai/workflows/moai.md` § Agentic Completion Loop **Termination** clause — four causes: (1) condition met; (2) iteration ceiling (`workflow.agentic_loop.max_iterations`, default 10); (3) escalation; (4) context-threshold suspension. Causes 3 and 4 have protocols (no-progress escalation AskUserQuestion round; context handoff); cause 2 has none.
- **Observed pattern**: A ceiling exit produces no structured claim/evidence/gaps report and persists nothing.

### GAP-3 — max_iterations fragmentation, no precedence rule (L6)

- **Measured source**: loop.md § Supported Flags: `--max N ... (default 100)`; loop.md Step 2 area: *"If memory-safe limit reached (50 iterations): Exit with checkpoint"*; `.moai/config/sections/workflow.yaml` `loop_prevention.max_iterations: 100` (line 17) and `agentic_loop.max_iterations: 10` (line 7); `.moai/config/sections/ralph.yaml` `loop.max_iterations: 10` (line 28).
- **Observed pattern**: Five numeric ceilings across four surfaces with no stated precedence. The Go-loaded typed config (ralph.yaml → `cfg.Ralph`; workflow.yaml agentic_loop → `AgenticLoopConfig`) says 10 while the skill body's flag default says 100.

### GAP-4 — Remaining issues evaporate; stale loader comment (L8)

- **Measured source**: loop.md fix-level handling (Level 4 manual items listed in report only); workflow.yaml lines 2-5 comment: *"agentic_loop: pipeline-level completion-loop iteration ceiling, prose-read by the orchestrator (documented key; no Go-side loader field yet)"* — vs `internal/config/types.go` `AgenticLoop AgenticLoopConfig \`yaml:"agentic_loop"\`` (SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 comment block, observed near line 326) and `internal/config/defaults.go` `AgenticLoop: AgenticLoopConfig{MaxIterations: DefaultAgenticLoopMaxIterations}` (observed near lines 441-444).
- **Observed pattern**: Ceiling leftovers and Level-4 manual items exist only in the transcript; the loader comment is demonstrably stale.

### Aggregate defect claim

**The utility loop's success is self-declared, its failure exits are contract-free, its ceilings are fragmented, and its leftovers evaporate.** This SPEC replaces the sentinel with a mechanical exit predicate + independent final pass, defines the ceiling-exit verdict contract with persistence, unifies ceiling precedence, and fixes the stale loader comment.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the loop engine", "the loop workflow", "the workflow config"). No legacy `IF/THEN` modality.

### REQ-LVC-001 — Ubiquitous — mechanical success predicate

The loop engine (loop.md per-iteration cycle) SHALL determine success-exit exclusively from a mechanical predicate over the CURRENT iteration's parsed Step-3 scanner diagnostics (exit codes and error/test/coverage counts) evaluated against ralph.yaml `loop.completion` (`zero_errors`, `tests_pass`, `coverage_threshold`, `zero_warnings`). The completion sentence "All loop completion conditions satisfied; exiting loop." becomes a display-only string emitted AFTER the predicate holds.

### REQ-LVC-002 — Unwanted behavior — sentinel-based exit prohibition

The loop engine SHALL NOT exit with success on detection of the self-emitted prose completion sentence (or any natural-language completion declaration) alone. Step 1's sentence string-match exit path is removed.

### REQ-LVC-003 — Event-driven (When) — independent final pass

**When** the mechanical predicate first evaluates satisfied, the loop workflow SHALL execute an independent final verification pass that is NOT the loop executor — a fresh-context re-run of the diagnostic gate (e.g. `/moai gate` re-run or a read-only verifier `Agent()` spawn) — and SHALL declare success-exit only after the independent pass confirms the predicate. A divergence between builder-observed and independently-observed diagnostics continues the loop (or escalates per existing safety rules).

### REQ-LVC-004 — Event-driven (When) — ceiling-exit 5-section verdict

**When** the loop exits at the iteration ceiling, the loop workflow SHALL emit a structured 5-section evidence report per `verification-claim-integrity.md` §3 — Claim / Evidence / Baseline-attribution / Gaps / Residual-risk — covering: iterations consumed, per-condition final state (errors / tests / coverage vs targets), and the enumerated remaining issues. The same contract SHALL be recorded for moai.md § Agentic Completion Loop termination cause 2 (max-iterations), closing its protocol gap relative to causes 3 and 4.

### REQ-LVC-005 — Event-driven (When) — remaining-issue persistence

**When** the loop exits at the ceiling (or exits with Level-4 manual items outstanding), the loop workflow SHALL persist the remaining issues to `.moai/state/loop-verdict-<id>.json` (schema defined in plan.md §D; `<id>` = session- or timestamp-derived) or, where a task ledger is active, to the TaskList — transcript-only residue is prohibited.

### REQ-LVC-006 — Event-driven (When) — lesson-capture proposal on unsuccessful exit

**When** the loop exits unsuccessfully (ceiling reached with conditions unmet), the loop workflow SHALL require a lesson-capture proposal step (per the Lessons Protocol) before session close — the failure pattern is offered for memory capture rather than silently dropped.

### REQ-LVC-007 — Ubiquitous — single ceiling-precedence rule

A single iteration-ceiling precedence rule — **CLI `--max` flag > ralph.yaml `loop.max_iterations` > workflow.yaml `loop_prevention.max_iterations`** — SHALL be stated identically in loop.md, ralph.yaml comments, and workflow.yaml comments, and the documented defaults on all three surfaces SHALL be mutually consistent with that rule (the loop.md "--max default 100" claim reconciled against the Go-loaded ralph.yaml value 10; the memory-safe 50-iteration cap documented as an orthogonal memory-pressure checkpoint, not a fourth ceiling).

### REQ-LVC-008 — Ubiquitous — stale loader comment fix

The workflow config (workflow.yaml `agentic_loop` comment) SHALL be corrected to reference the implemented Go loader (`internal/config` `AgenticLoopConfig` + its default), removing the stale "no Go-side loader field yet" claim (live + template mirror).

### REQ-LVC-009 — Capability gate (Where) — template-first boundary

**Where** an edited surface has a template mirror under `internal/template/templates/` (verified present for: loop.md, moai.md, workflow.yaml, ralph.yaml), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees.

---

## Constraints

1. **Loop-identity preservation (HARD)** — `/moai loop` ≡ `/moai run --mode loop` alias contract, the diagnostic-loop vs pipeline-level agentic-loop distinction (loop.md header + workflow.yaml §A.4 distinctness per SPEC-V3R6-AGENTIC-LOOP-CONFIG-001), and existing safety rules (no-progress escalation, dark-flow guard, semantic-failure escalation) are all PRESERVED. This SPEC changes the EXIT decision procedure and ceiling contract, not loop identity or safety semantics.
2. **vci alignment (HARD)** — the 5-section verdict report reuses `verification-claim-integrity.md` §3 verbatim section names; the independent final pass embodies §1.1 (builder may not self-declare verification).
3. **Subagent boundary** — the independent verifier (if agent-spawned) returns results to the orchestrator; it never prompts the user. Ceiling-exit user interaction (continue / abort options) remains orchestrator-owned AskUserQuestion territory.
4. **Deliverable classification** — doctrine/skill-doc edits are primary; the verdict-file schema is doctrine-defined (orchestrator writes JSON at runtime); NO new Go loader for the verdict file in this SPEC (plan.md §D D2). The only Go-adjacent change is none-to-minimal; comment fixes are YAML-only.
5. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — SPEC-pipeline loops (plan/run/sync)

- The plan→run→sync pipeline and its auditor separation (plan-auditor, sync-auditor) already provide independent evaluation; nothing in the SPEC pipeline changes. Only utility-loop termination (`/moai loop` + the moai.md agentic completion loop's cause-2 contract) is in scope.

### Out of Scope — /goal implementation

- The `/goal` directive, its Haiku evaluator, and goal-directive.md semantics. `/goal` is cited as the existing independent-evaluator precedent, not modified.

### Out of Scope — /moai fix Phase 4 scanner independence (L5)

- Audit finding L5 (fix.md Phase 4 re-runs the same scanners as the builder — partial mechanical independence) is recorded in Traceability as provenance but deliberately NOT remediated here: `/moai fix` is a one-pass pipeline utility, not an iterative loop, and widening scope to its verification phase would blur this SPEC's termination-predicate focus. L5 remains open audit debt for a follow-up SPEC.

### Out of Scope — ralph.yaml completion-threshold values

- The values inside `loop.completion` (coverage 85, zero_warnings false, etc.) are not retuned; the predicate consumes them as-is.

### Out of Scope — Go verdict-file loader / CLI reader

- A `moai` CLI subcommand or Go loader that parses `.moai/state/loop-verdict-<id>.json`. The schema is doctrine-defined and orchestrator-written in this SPEC; mechanical consumers are follow-up territory.

### Out of Scope — loop_prevention semantics

- `workflow.yaml loop_prevention` (per-operation retry bound, failure-pattern detection) semantics are unchanged; it is only NAMED in the precedence rule as the lowest-precedence documented ceiling surface.

---

## Cross-References

- **EXTEND base (doc)**: `.claude/skills/moai/workflows/loop.md` (Step 1 / Step 4 / Step 9 / § Completion Conditions / § Supported Flags); `.claude/skills/moai/workflows/moai.md` § Agentic Completion Loop (Termination clause, cause 2); `.moai/config/sections/ralph.yaml` (loop block comments); `.moai/config/sections/workflow.yaml` (agentic_loop comment, loop_prevention comment). All four have verified template mirrors.
- **Preserved invariants**: loop.md diagnostic-loop vs pipeline-loop distinctness (SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 §A.4); loop safety rules (no-progress escalation / dark-flow guard / semantic-failure escalation) in moai.md; `/moai loop` alias contract.
- **Precedents cited**: `.claude/rules/moai/workflow/goal-directive.md` (separate Haiku evaluator); `.moai/config/sections/harness.yaml` (`levels.minimal.evaluator: false` — evaluator as explicit switch); `verification-claim-integrity.md` §1.1 + §3 (the 5-section format REQ-LVC-004 reuses).
- **Go loader referenced by REQ-LVC-008**: `internal/config/types.go` (AgenticLoop field) + `internal/config/defaults.go` (DefaultAgenticLoopMaxIterations) — referenced, not modified.
- **Epic**: Workflow-Reflex 3 of 6. Siblings: SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6), SPEC-MODEL-ROUTING-WIRE-001 (2 of 6).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 3 of 6. Mechanical exit predicate + independent final pass + ceiling verdict contract + precedence unification + stale comment fix. Tier M. |
