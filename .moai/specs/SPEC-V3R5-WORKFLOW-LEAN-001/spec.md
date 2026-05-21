---
id: SPEC-V3R5-WORKFLOW-LEAN-001
title: "Workflow LEAN — Tier-based SPEC + Section A-E Optional + plan-auditor Escalation"
version: "0.2.0"
status: implemented
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.0.0 — Round 5"
module: workflow-rules
lifecycle: spec-anchored
tags: "workflow, lean, optimization, v3r5"
tier: S
---

# SPEC-V3R5-WORKFLOW-LEAN-001 — Workflow LEAN

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-20 | GOOS Kim | Initial draft. Targets root cause of SPEC-V3R5-WORKFLOW-OPT-001 over-formalization observed in LANG-COMPLIANCE-001 plan-phase abandonment (2026-05-20). Introduces SPEC complexity Tier S/M/L, Section A-E optional clause, plan-auditor STOP escalation + tier threshold + max 3 iterations. **LEAN dogfooding**: this SPEC itself is Tier S (2 artifacts, ≤800 LOC, AC inline). |
| 0.2.0 | 2026-05-20 | manager-develop (run-phase) | Run-phase complete. 4 milestones merged into main (M1 spec-workflow.md + spec-frontmatter-schema.md, M2 manager-develop-prompt-template.md, M3 plan-auditor.md, M4 spec-assembly.md) + 1 alignment fix. 5 files modified, +101/-1 LOC. All 9 ACs (AC-WL-001 through AC-WL-011 except AC-WL-005 run-phase deferred + AC-WL-006/010/011 dogfooding) PASS. Status: draft → implemented. |

## 1. Goal

Fix the workflow inflation root cause introduced by SPEC-V3R5-WORKFLOW-OPT-001 (PR #1025/#1026 merged 2026-05-20). The over-formalization made every SPEC pay meta-overhead designed for W3-scale complexity (4-Tier + 5-Layer + 18 sentinels), causing simple SPECs like LANG-COMPLIANCE-001 to abandon plan-phase after 30+ min, 6 Agent() spawns, 80+ tool calls, and a score-regression pattern (0.78 → 0.81 → 0.77).

Introduce SPEC complexity tiering, optional template application, and bounded plan-auditor iteration.

## 2. Context

Triggering incident: 2026-05-20 LANG-COMPLIANCE-001 plan-phase abandonment. Reference: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_workflow_inflation_root_cause.md`.

Five inflation issues identified:

- Issue #1 [P0]: Section A-E 5-section template uniformly mandatory (~4500-token baseline per delegation, irrelevant B1-B8 known issues for simple SPECs)
- Issue #2 [P0]: plan-auditor 4-iteration score-regression pattern (no escalation, no iteration cap, uniform 0.85 threshold)
- Issue #3 [P0]: manifest enumeration scope creep (self-referential dependency chain)
- Issue #4 [P1]: 5-artifact (spec/plan/acceptance/design/research) uniformly mandatory regardless of SPEC complexity
- Issue #5 [P1]: subagent fresh-context grep duplication (orchestrator baseline + N subagent re-greps)

Issues #1, #2, #4 are in scope. Issues #3, #5 are out of scope (see §6).

W3 dogfooding -73% wall-time reduction was a function of W3's intrinsic complexity, not of the formalization itself. Applying the same formalization to simple SPECs inverts the cost-benefit ratio.

## 3. EARS Requirements + Acceptance Criteria (inline)

### 3.1 Tier Classification

- REQ-WL-001 (Ubiquitous): The SPEC plan-phase shall classify each SPEC into Tier S, M, or L before artifact creation begins.
- REQ-WL-002 (State-driven): While Tier=S, the artifact set shall be 2 files (spec.md + plan.md), with AC inline in spec.md §3.
- REQ-WL-003 (State-driven): While Tier=M, the artifact set shall be 3 files (spec.md + plan.md + acceptance.md).
- REQ-WL-004 (State-driven): While Tier=L, the artifact set shall be 5 files (spec.md + plan.md + acceptance.md + design.md + research.md), preserving current default behavior.
- REQ-WL-009 (Ubiquitous): The Tier classification shall be recorded in `spec.md` frontmatter as a `tier` field (enum: S | M | L).

AC for §3.1:

- AC-WL-001: `grep -n "Tier S/M/L\|complexity tier" .claude/rules/moai/workflow/spec-workflow.md` returns at least one line referencing tier definition.
- AC-WL-009: `grep -n "tier:" .claude/rules/moai/development/spec-frontmatter-schema.md` returns at least one line documenting the optional `tier` field.

### 3.2 Section A-E Template Optional

- REQ-WL-005 (Optional): Where the SPEC tier is S, the Section A-E delegation template MAY be skipped; orchestrator delegation prompts MAY use a minimal form (task + constraints + AC only).
- REQ-WL-011 (Optional): Where the SPEC tier is M or L, the Section A-E template SHOULD be applied; Section B (known issues) MAY filter B1-B8 categories by domain relevance.

AC for §3.2:

- AC-WL-002: `grep -n "Tier S\|optional\|Tier S/M/L" .claude/rules/moai/development/manager-develop-prompt-template.md` returns at least one line near the Section A-E heading documenting the optional clause.

### 3.3 plan-auditor Escalation + Bounds

- REQ-WL-006 (Event-driven): When `plan-auditor` iter(N+1) aggregate score is lower than iter(N) aggregate score, the orchestrator shall STOP further iterations and emit a scope-reduction proposal to the user.
- REQ-WL-007 (State-driven): While Tier=S, the plan-auditor PASS threshold shall be 0.75; while Tier=M, 0.80; while Tier=L, 0.85.
- REQ-WL-008 (Unwanted): The plan-auditor shall not exceed 3 iterations per SPEC plan-phase. After iteration 3, the orchestrator shall escalate (PASS-with-debt OR scope-reduction OR explicit user override).

AC for §3.3:

- AC-WL-003: `grep -nE "STOP|score regression|escalation|iter[23]\b" .claude/agents/moai/plan-auditor.md` returns at least one line documenting the STOP-on-regression behavior.
- AC-WL-007: `grep -nE "0\.75|0\.80|tier threshold|tier-differentiated" .claude/agents/moai/plan-auditor.md` returns at least one line documenting tier-specific thresholds.
- AC-WL-008: `grep -nE "max 3|iteration cap|3 iterations|three iterations" .claude/agents/moai/plan-auditor.md` returns at least one line documenting the iteration cap.

### 3.4 Tier Judgment Surface

- REQ-WL-010 (Event-driven): When the orchestrator enters plan-phase via `/moai plan`, the `spec-assembly` skill shall present a Tier judgment question to the user (Socratic AskUserQuestion), with concrete LOC thresholds (<300 / 300-1000 / 1000+) as decision guidance.
- REQ-WL-012 (Optional): Where the user explicitly provides Tier in the request (e.g., "Tier S"), the Tier judgment question MAY be skipped.

AC for §3.4:

- AC-WL-004: `grep -nE "Tier|complexity tier" .claude/skills/moai/workflows/plan/spec-assembly.md` returns at least one line documenting the Tier judgment step.

### 3.5 End-to-End Validation

- REQ-WL-013 (Ubiquitous): The LEAN workflow shall demonstrate measurable reduction in plan-phase cost on a simple SPEC (Tier S) compared to the SPEC-V3R5-WORKFLOW-OPT-001 baseline.

AC for §3.5:

- AC-WL-005: Apply the LEAN workflow to the next simple SPEC (Tier S) in run-phase. Verify (manual measurement, documented in run-phase progress.md):
  - Total tool calls ≤ 30 (LANG-COMPLIANCE-001 baseline: 80+)
  - Wall-time ≤ 15 minutes (LANG-COMPLIANCE-001 baseline: 30+ min, abandoned)
  - plan-auditor PASS on iter1 (LANG-COMPLIANCE-001 baseline: 3 iter score regression)
  - Artifact count = 2 (LANG-COMPLIANCE-001 baseline: 5 + manifest)

### 3.6 Self-validation (Dogfooding This SPEC)

- REQ-WL-014 (Ubiquitous): This SPEC itself shall conform to Tier S constraints to demonstrate the rule self-applies.

AC for §3.6:

- AC-WL-006: `ls .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/` returns exactly 2 files (spec.md, plan.md). No acceptance.md, design.md, research.md, or manifest files.
- AC-WL-010: `wc -l .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/spec.md .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/plan.md` shows total line count ≤ 800.
- AC-WL-011: `go run ./cmd/moai spec lint --strict .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/spec.md` exits 0 with no findings.

## 4. Risks

- R-WL-001 (High): Tier judgment is subjective; implementers may miscategorize complex SPECs as Tier S to avoid overhead. Mitigation: REQ-WL-010 Socratic question presents explicit LOC thresholds; plan-auditor first-pass score regression triggers tier-up suggestion.

- R-WL-002 (Medium): Backward compatibility — existing SPECs (created pre-LEAN) implicitly assumed Tier L (5 artifacts). Mitigation: REQ-WL-009 `tier` field is optional in frontmatter (absence = Tier L for backward compat); existing SPECs require no migration.

- R-WL-003 (Medium): plan-auditor STOP-on-regression may trigger on legitimate refinement cycles (e.g., iter1 adds defects then iter2 resolves them with a score dip before climbing). Mitigation: REQ-WL-006 STOP emits a proposal, not a hard abort; user MAY override and continue.

- R-WL-004 (Low): Threshold differentiation (0.75/0.80/0.85) may produce inconsistent quality across tiers. Mitigation: Tier S scope is intrinsically narrower → lower-threshold passes are still high-confidence in absolute terms. Tier L retains current strict 0.85.

- R-WL-005 (Low): The 3-iteration cap (REQ-WL-008) may cut off SPECs that genuinely need 4+ iterations. Mitigation: After iter3, the escalation path includes "explicit user override" — user can extend cap with conscious choice, not silent drift.

## 5. Constraints

- C-WL-001: No changes to `internal/spec/lint.go` `FrontmatterSchemaRule`. The `tier` field is optional and unenforced by lint; documentation-only addition to `spec-frontmatter-schema.md`.
- C-WL-002: No changes to `internal/spec-auditor/` runtime (plan-auditor agent). All threshold/escalation logic lives in `.claude/agents/moai/plan-auditor.md` prompt body.
- C-WL-003: All affected files are markdown rules/skills/agents; no Go code modification required.
- C-WL-004: Late-branch workflow (SPEC-V3R5-LATE-BRANCH-001 v0.3.0) discipline applies — commits accumulate on main, PR branch created late.

## 6. Out of Scope

### 6.1 Out of Scope

- **Issue #3 (manifest enumeration scope creep)**: Manifest file pattern itself is not banned by this SPEC. Future SPECs MAY introduce manifests when justified (e.g., automated tooling input). Tier S guidance discourages but does not prohibit.
- **Issue #5 (subagent fresh-context grep duplication)**: Subagent baseline caching is deferred. Effective caching requires fresh-context policy revision, which is a separate concern from workflow tiering.
- **plan-auditor agent Go-level changes**: All behavior changes are prompt-only (`.claude/agents/moai/plan-auditor.md`).
- **`tier` field lint enforcement**: The `tier` field is optional; `FrontmatterSchemaRule` is not modified.
- **Retroactive Tier assignment to existing SPECs**: Existing SPECs created pre-LEAN retain implicit Tier L behavior; no migration mandated.
- **manager-develop runtime changes**: Section A-E template is markdown rule; manager-develop agent code unchanged.
- **CI/lint validation of LEAN compliance**: AC-WL-005 (end-to-end metrics) is measured manually in run-phase, not enforced by CI in this SPEC.
- **SPEC-V3R5-WORKFLOW-OPT-001 rollback**: WORKFLOW-OPT-001 remains merged; LEAN refines it, does not revert it.

## 7. Affected Files (planned for run-phase)

- `.claude/rules/moai/workflow/spec-workflow.md` — Tier S/M/L definition (M1)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E optional clause for Tier S (M2)
- `.claude/agents/moai/plan-auditor.md` — STOP-on-regression + tier thresholds + max-3-iter cap (M3)
- `.claude/skills/moai/workflows/plan/spec-assembly.md` — Tier judgment Socratic question (M4)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Optional `tier` field documentation (M1 supplement)

## 8. REQ ↔ AC Traceability

| REQ | AC | Verification |
|-----|----|----|
| REQ-WL-001 | AC-WL-001 | grep in spec-workflow.md |
| REQ-WL-002, 003, 004 | AC-WL-001 | grep in spec-workflow.md (tier→artifact rule co-located) |
| REQ-WL-005, 011 | AC-WL-002 | grep in manager-develop-prompt-template.md |
| REQ-WL-006 | AC-WL-003 | grep STOP/escalation in plan-auditor.md |
| REQ-WL-007 | AC-WL-007 | grep tier thresholds in plan-auditor.md |
| REQ-WL-008 | AC-WL-008 | grep iteration cap in plan-auditor.md |
| REQ-WL-009 | AC-WL-009 | grep tier field in spec-frontmatter-schema.md |
| REQ-WL-010, 012 | AC-WL-004 | grep tier judgment in spec-assembly.md |
| REQ-WL-013 | AC-WL-005 | run-phase manual metrics |
| REQ-WL-014 | AC-WL-006, 010, 011 | this SPEC's own 2-file, ≤800-LOC, spec-lint compliance |
