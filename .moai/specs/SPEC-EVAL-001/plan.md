---
spec_id: SPEC-EVAL-001
plan_version: "0.1.0"
plan_date: 2026-05-14
plan_author: manager-spec (auto)
plan_status: draft
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-14 | manager-spec | Initial plan |

## 1. Plan Overview

SPEC-EVAL-001 defines the evaluator-active agent and Sprint Contract mechanism. Codebase analysis reveals that the core functionality is **already implemented**: evaluator-active.md agent definition exists with 4-dimension scoring, Sprint Contract negotiation (Phase 2.0) and Phase 2.8a/2.8b split are present in run.md, mode-specific deployment (sub-agent/team/CG) is documented, and intervention modes (final-pass/per-sprint) are integrated.

The remaining work is verification and gap-closure: confirming all REQ are satisfied in the current implementation, adding any missing acceptance criteria coverage, and ensuring HRN-003 hierarchical scoring integration is complete.

## 2. Gap Analysis

### Already Implemented

| REQ | Status | Evidence |
|-----|--------|----------|
| REQ-001 | DONE | `evaluator-active.md` exists at `internal/template/templates/.claude/agents/moai/evaluator-active.md` with model: sonnet, permissionMode: plan, 4 dimensions, Security hard threshold |
| REQ-002 | DONE | Phase 2.0 Sprint Contract in `run.md` line 606+, gated by harness level = thorough |
| REQ-003 | DONE | Phase 2.8a (evaluator-active) and 2.8b (manager-quality) split in `run.md` lines 818-858 |
| REQ-004 | DONE | Mode-specific deployment documented in evaluator-active.md lines 104-108 |
| REQ-005 | DONE | final-pass (standard) and per-sprint (thorough) documented in evaluator-active.md lines 99-102 |

### Gaps

1. **HRN-003 integration completeness**: evaluator-active.md includes HRN-003 hierarchical scoring protocol, but no Go-level enforcement of canonical anchors (0.25, 0.50, 0.75, 1.0) exists. This is handled at prompt level only.
2. **contract.md schema validation**: No automated check that generated contract.md follows the required schema.
3. **evaluator-active hook**: SubagentStop hook exists (`evaluator-completion`) but no validation of output format.

## 3. Milestone Breakdown

### M1 -- Verify REQ-001 Compliance -- Priority P0

Verify evaluator-active.md satisfies all REQ-001 acceptance criteria:
- Confirm permissionMode: plan (read-only)
- Confirm 4 evaluation dimensions with correct weights
- Confirm Security FAIL = Overall FAIL hard threshold
- Confirm skeptical evaluator prompt language

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (verify only, likely no changes needed)

### M2 -- Verify REQ-002/003 Phase Integration -- Priority P0

Verify run.md Phase 2.0 and Phase 2.8a/2.8b integration:
- Phase 2.0 Sprint Contract gated by thorough harness level
- contract.md persistence path correct
- Phase 2.8a evaluator-active max 3 retry cycles enforced
- Phase 2.8b manager-quality runs after 2.8a

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (verify only)

### M3 -- Add Evaluation Output Format Validation -- Priority P1

Add structured output format enforcement to evaluator-active.md:
- Validate evaluation report follows the required Markdown table format
- Add SubagentStop hook validation for output structure
- Ensure dimension scores are numeric and within [0, 100] range

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (output format section)
- `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh` (evaluator-completion validation)

### M4 -- Integration Test Scenarios -- Priority P1

Document test scenarios for acceptance criteria verification:
- Scenario 1: Standard Harness Evaluation flow
- Scenario 2: Thorough Harness Full Cycle flow
- Scenario 3: Security Hard Threshold enforcement

Files:
- `.moai/specs/SPEC-EVAL-001/acceptance.md` (verify existing scenarios match spec)

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | 1-157 | Verify | Agent definition, dimensions, thresholds |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 606-630 | Verify | Phase 2.0 Sprint Contract |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 818-858 | Verify | Phase 2.8a/2.8b split |
| `internal/template/templates/.moai/config/evaluator-profiles/default.md` | 1-EOF | Verify | Default profile dimensions/weights |

## 5. Quality Gates

- All 5 REQ acceptance criteria verifiable against existing implementation
- No regressions in existing evaluator-active behavior
- run.md phase ordering preserved (2.0 -> 2.1 -> ... -> 2.8a -> 2.8b)
- evaluator-active.md frontmatter valid per agent-authoring rules

## 6. Dependencies

- SPEC-EVALLIB-001: Evaluator profiles already created and loaded by evaluator-active
- SPEC-V3R2-HRN-003: Hierarchical scoring protocol already integrated in evaluator-active.md
- SPEC-DRIFT-001: tasks.md generation already in run.md (drift guard dependency)
