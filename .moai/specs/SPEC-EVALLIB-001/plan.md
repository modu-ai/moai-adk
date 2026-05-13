---
spec_id: SPEC-EVALLIB-001
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

SPEC-EVALLIB-001 creates an Evaluator Profile library with 4 default profiles and JSONL evaluation logging. Codebase analysis reveals this is **substantially implemented**:

- **4 profile files exist**: `default.md`, `strict.md`, `lenient.md`, `frontend.md` in `internal/template/templates/.moai/config/evaluator-profiles/`
- **Profile loading logic exists**: evaluator-active.md lines 79-89 describe profile loading from SPEC frontmatter and harness.yaml fallback
- **Frontend anti-AI-slop criteria**: frontend.md profile includes AI-slop detection patterns

The remaining work is JSONL logging (REQ-003) and verification of all acceptance criteria.

## 2. Gap Analysis

### Already Implemented

| REQ | Status | Evidence |
|-----|--------|----------|
| REQ-001 | DONE | 4 profile files exist in evaluator-profiles/ directory |
| REQ-002 | DONE | frontend.md has Originality/Design Quality/Craft dimensions with AI-slop penalties |
| REQ-003 | PARTIAL | JSONL logging to .moai/metrics/evaluation-log.jsonl not yet implemented |

### Remaining Gaps

1. **JSONL evaluation logging**: No .moai/metrics/evaluation-log.jsonl append logic in run.md or evaluator-active.md
2. **Periodic analysis trigger**: No 50-entry threshold for FP/FN analysis
3. **Profile schema standardization**: Verify all 4 profiles follow consistent schema (dimensions, weights, thresholds, penalty_patterns)

## 3. Milestone Breakdown

### M1 -- Verify Profile Completeness -- Priority P0

Verify all 4 evaluator profiles are complete and correctly structured:
- default.md: Functionality 40%, Security 25%, Craft 20%, Consistency 15%
- strict.md: Security 35%, all dimensions require 80%+
- lenient.md: Functionality 60%, Security 20%, Craft 10%, Consistency 10%
- frontend.md: Originality 40%, Design Quality 30%, Craft & Functionality 30% with AI-slop patterns

Files:
- `internal/template/templates/.moai/config/evaluator-profiles/default.md` (verify)
- `internal/template/templates/.moai/config/evaluator-profiles/strict.md` (verify)
- `internal/template/templates/.moai/config/evaluator-profiles/lenient.md` (verify)
- `internal/template/templates/.moai/config/evaluator-profiles/frontend.md` (verify)

### M2 -- Verify Profile Loading Logic -- Priority P0

Verify evaluator-active.md profile loading satisfies REQ-001 acceptance:
- SPEC frontmatter `evaluator_profile` field respected
- Fallback to harness.yaml `default_profile` when no SPEC-level override
- Final fallback to built-in defaults when profile file not found

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` lines 79-89 (verify)

### M3 -- Add JSONL Evaluation Logging -- Priority P1

Implement evaluation result logging to .moai/metrics/evaluation-log.jsonl:
- Append JSONL entry after each evaluator-active completion
- Fields: timestamp, spec_id, profile_used, dimension_scores, overall_verdict, flags
- Create .moai/metrics/ directory if not exists
- Add JSONL logging instruction to run.md Phase 2.8a completion section

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (Phase 2.8a post-evaluation logging step)

### M4 -- Add Periodic Analysis Trigger -- Priority P2

Add 50-entry analysis trigger for FP/FN tuning recommendations:
- When evaluation-log.jsonl reaches 50 entries, generate summary analysis
- Output to .moai/metrics/analysis-{date}.md
- Recommendations only (not auto-applied per exclusion)
- Analysis triggered by orchestrator, not automated

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (add analysis trigger note in Phase 2.8a)

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.moai/config/evaluator-profiles/default.md` | 1-EOF | Verify | Default profile weights |
| `internal/template/templates/.moai/config/evaluator-profiles/frontend.md` | 1-EOF | Verify | AI-slop detection criteria |
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | 79-89 | Verify | Profile loading logic |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 818-858 | Edit | Add JSONL logging after Phase 2.8a |

## 5. Quality Gates

- All 4 profile files exist with correct dimensions, weights, and thresholds
- Profile loading from SPEC frontmatter works (manual verification)
- JSONL entries appended on each evaluation run
- No profile auto-modification (human review required per exclusion)

## 6. Dependencies

- SPEC-EVAL-001: evaluator-active agent definition (profile loading consumer)
- SPEC-V3R2-HRN-003: Hierarchical scoring protocol (sub-criterion anchors used in profiles)
