---
spec_id: SPEC-DRIFT-001
plan_version: "0.1.0"
plan_date: 2026-05-14
plan_author: manager-spec (auto)
plan_status: draft
---

## 1. Plan Overview

SPEC-DRIFT-001 introduces persistent tasks.md for task decomposition and real-time drift guard to detect scope creep during implementation. Codebase analysis reveals this is **already implemented** in run.md:

- **tasks.md generation**: Present at run.md Phase 1.5 (lines 518-533) with structured format (Task ID, Description, REQ mapping, Dependencies, Status, Planned files)
- **Drift guard calculation**: Present in both Phase 2A (DDD, lines 693-702) and Phase 2B (TDD, lines 736-745) with percentage-based thresholds
- **Drift thresholds**: <= 20% informational, 20-30% warning, > 30% cumulative triggers re-planning

The remaining work is verification that all REQ are satisfied and edge cases are handled correctly.

## 2. Gap Analysis

### Already Implemented

| REQ | Status | Evidence |
|-----|--------|----------|
| REQ-001 | DONE | tasks.md generation at run.md lines 518-533 |
| REQ-002 | DONE | Drift guard in run.md lines 693-702 (DDD) and 736-745 (TDD) |
| REQ-003 | DONE | Warning to progress.md + > 30% cumulative triggers Phase 2.7 |

### Remaining Gaps

1. **tasks.md schema validation**: No automated check that generated tasks.md follows the required field structure.
2. **Exclusion patterns**: Drift guard should exclude .gitignore patterns, test fixtures, and generated files. Verify this is documented.
3. **Deterministic output**: REQ-001 requires "no timestamps" for git-trackability. Verify tasks.md generation avoids timestamps.

## 3. Milestone Breakdown

### M1 -- Verify tasks.md Generation -- Priority P0

Verify tasks.md generation satisfies REQ-001:
- Confirm all required fields present (Task ID, Description, REQ mapping, Dependencies, Status, Planned files)
- Confirm output is deterministic (no timestamps, sorted consistently)
- Confirm git-trackable format
- Confirm each task maps to at least one REQ-XXX requirement

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` lines 518-533 (verify only)

### M2 -- Verify Drift Guard Logic -- Priority P0

Verify drift guard calculation satisfies REQ-002:
- Confirm drift formula: (unplanned_new_files / total_planned_files) * 100
- Confirm threshold tiers: <= 20% info, 20-30% warning, > 30% re-plan trigger
- Confirm drift check runs after each DDD/TDD cycle completion
- Confirm drift measurement logged to progress.md

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` lines 693-702, 736-745 (verify only)

### M3 -- Verify Scope Alarm Integration -- Priority P1

Verify scope alarm satisfies REQ-003:
- Confirm warnings appended to progress.md with drift details
- Confirm cumulative > 30% triggers Phase 2.7 re-planning gate
- Confirm re-planning is an extension of existing Phase 2.7, not a replacement

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` lines 700-702, 744-745 (verify only)

### M4 -- Add Exclusion Pattern Documentation -- Priority P2

Document drift exclusion patterns explicitly:
- .gitignore-matched files excluded from drift calculation
- Test fixtures excluded
- Generated files (e.g., embedded.go) excluded
- .moai/state/, .moai/reports/, .moai/cache/ excluded

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (add exclusion pattern note)

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 518-533 | Verify | tasks.md generation |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 693-702 | Verify | DDD drift guard |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 736-745 | Verify | TDD drift guard |
| `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | Drift Guard section | Verify | Drift guard reference |

## 5. Quality Gates

- tasks.md generated with all required fields for any SPEC entering Run phase
- Drift guard correctly calculates percentage and applies threshold tiers
- Cumulative drift > 30% triggers Phase 2.7
- No regressions in existing Run phase flow

## 6. Dependencies

- SPEC-EVAL-001: evaluator-active evaluates against acceptance criteria (runs after drift guard)
- run.md: Phase 2.7 re-planning gate must exist before scope alarm integration
