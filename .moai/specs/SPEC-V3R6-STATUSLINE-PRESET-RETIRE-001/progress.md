---
id: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
title: "Progress — Retire statusline preset system + remove web-console statusline panel"
version: "0.2.0"
status: draft
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "progress"
lifecycle: spec-anchored
tags: "progress, statusline, preset, mode, retire"
tier: M
---

# Progress — SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001

## §A. Current Phase

**plan-phase** — artifacts created (spec.md + plan.md + acceptance.md +
progress.md). Awaiting plan-auditor independent audit, then Implementation
Kickoff Approval (the plan-to-implement human gate), then run-phase entry.

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/spec.md` | draft v0.2.0 (24 REQs, 28 ACs, 8 exclusions) |
| plan.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/plan.md` | draft v0.2.0 (6 milestones M1-M6, D1-D7 remediation) |
| acceptance.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/acceptance.md` | draft v0.2.0 (25 MUST + 3 SHOULD ACs) |
| progress.md | this file | §E skeleton emitted |

## §C. Milestone Tracker

| Milestone | Scope | Status | Evidence |
|-----------|-------|--------|----------|
| M1 | Research / baseline capture | pending | _<pending run-phase>_ |
| M2 | Go code removal (models/statusline/cli/profile) | pending | _<pending run-phase>_ |
| M3 | Web console removal (fieldsets/handlers/validate + 3 tests) | pending | _<pending run-phase>_ |
| M4 | Wizard cleanup (profile_setup.go) | pending | _<pending run-phase>_ |
| M5 | Template + docs-site (4-locale) | pending | _<pending run-phase>_ |
| M6 | Final verification + sync prep | pending | _<pending run-phase>_ |

## §D. Blockers / Open Items

None at plan-phase. The 3 open questions in `spec.md §F.2` are run-phase
decisions, not user-blocking.

---

## §E.1 Plan-phase Audit-Ready Signal

The plan-phase artifact set is complete and internally consistent (v0.2.0,
iter-1 FAIL remediation applied):

- **spec.md**: 12 canonical frontmatter fields present; SPEC ID matches the
  canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition: SPEC ✓ |
  V3R6 ✓ | STATUSLINE ✓ | PRESET ✓ | RETIRE ✓ | 001 ✓ → PASS); 24 GEARS-format
  requirements (REQ-SPR-001 through REQ-SPR-024, with REQ-SPR-022/023/024
  added in v0.2.0 for D2 fieldset-caller, D3 wizard mode Select, D6 i18n
  cleanup); 8 explicit exclusions; 3 confirmed user decisions captured in §A.1
  (preset retire + web panel removal + mode preferences axis retire).
- **plan.md**: 6 milestones (M1-M6) with ordered edits, dependency-aware
  sequencing (struct fields first, then consumers); D1-D7 iter-1 defects
  addressed (lowercase presetToSegments wrapper D1, fieldset caller D2, mode
  axis D3, 7 test files D4, statuslineData D5, mode i18n D6, docs mode: D7);
  8 anti-patterns documented; pre-flight checklist present.
- **acceptance.md**: 28 ACs (25 MUST + 3 SHOULD), all observable (grep/build/
  test/byte-diff verifiable); traceability matrix covers 24/24 requirements;
  6 edge cases enumerated. AC-SPR-025/026/027/028 added in v0.2.0 for D3
  wizard mode Select, D6 i18n, D4 test files, D3 Builder API preservation.
- **progress.md**: this file — §E.2 through §E.5 emitted as placeholder
  headings only (no populated evidence; that belongs to run/sync/Mx phases).

**Era classification (per `.claude/rules/moai/workflow/lifecycle-sync-gate.md`)**:
H-2 fallback would fire on this minimal progress.md (no `§E.2`-`§E.5` markers
populated yet) — but the SPEC is freshly created at plan-phase, so the H-5
tie-breaker (created ≥ 2026-04-01) AND the `era: V3R6` override (to be set if
needed) keep it in the V3R6 bucket. No grandfather-clause risk: this is a new
SPEC, subject to modern-era drift detection once run-phase populates §E.2.

**Ready for**: plan-auditor independent audit → Implementation Kickoff
Approval (human gate, §19.1 CLAUDE.local.md) → `/moai run` delegation to
manager-develop with `cycle_type=ddd`.

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates this section with verbatim
command output per `manager-develop-prompt-template.md §E` (E1-E7).
Evidence rows: baseline test/coverage capture (M1), per-milestone grep/build
output (M2-M5), final verification matrix (M6).>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — emitted by manager-develop upon M6 completion.>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this section. Records the
sync commit SHA, CHANGELOG entry, README update (if any), and the
frontmatter status transition (in-progress → implemented).>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — orchestrator-direct or manager-docs populates this
section with the Mx commit SHA, 4-phase close confirmation, and final
audit-ready YAML block.>_

---

## §F. Cross-References

- `spec.md` (canonical SSOT — §A overview, §B requirements, §C constraints,
  §E exclusions, §F risks)
- `plan.md` (M1-M6 milestone scope, anti-patterns, pre-flight checks)
- `acceptance.md` (30 ACs, traceability matrix, quality gates)
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (era classification,
  §E.1-§E.5 marker semantics)
- `.claude/rules/moai/development/manager-develop-prompt-template.md §E`
  (E1-E7 self-verification deliverables — the run-phase evidence shape)

---

## Out of Scope (progress-phase)

The following are explicitly out of scope for this progress tracker:

### 1. Out of Scope — populated §E.2-§E.5 evidence

- The §E.2 Run-phase Evidence, §E.3 Run-phase Audit-Ready Signal, §E.4
  Sync-phase Audit-Ready Signal, and §E.5 Mx-phase Audit-Ready Signal sections
  are placeholder skeletons at plan-phase. Populating them belongs to
  manager-develop (§E.2/§E.3), manager-docs (§E.4), and orchestrator-direct
  Mx (§E.5) per the SPEC artifact ownership matrix. This progress.md MUST NOT
  carry populated evidence at plan-phase.

### 2. Out of Scope — non-SPEC runtime state

- This progress tracker does NOT record runtime-managed state
  (`.moai/state/*`, `.moai/logs/*`, `.moai/cache/*`). Those are owned by the
  runtime and hooks, not by the SPEC lifecycle.
