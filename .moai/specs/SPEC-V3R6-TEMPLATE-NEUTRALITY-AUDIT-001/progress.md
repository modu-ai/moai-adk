---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Progress Tracking"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, progress"
tier: L
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Progress Tracking

## Status

**Phase**: plan-phase complete (draft v0.1.0)
**Created**: 2026-05-23
**Last update**: 2026-05-23 (initial plan write)

## Plan-Phase Summary

Tier L SPEC. 13 EARS GEARS REQs + 13 binary ACs + 6 milestones (M1–M6).

### Deliverables (this phase)

- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md` (§1–§5)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/plan.md` (§1 Milestones + §2 Risks + §3 Dependencies + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md` (§1 ACs + §2 Test plan + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/progress.md` (this file)

### Baseline measurements (2026-05-23, pre-fix)

| Category | Pattern | Files | Notes |
|---|---|--:|---|
| C1 macOS-bias | `/Users/` | 4 | Critical NEW — 8 lines / 4 files |
| C2 V3R refs | `V3R[0-9]` | 70 | Case-by-case (allow-list TBD M1) |
| C3 Dates | `2026-0[5-9]` | 32 | Case-by-case |
| C4 feedback/memory | `feedback_\|memory\.md` | 9 | Case-by-case |
| C5 CLAUDE.local | `CLAUDE\.local\.md` | 10 | Should be 0 |
| C6 PR #N | `PR #[0-9]+` | 3 | MUST be 0 |
| C7 Commit hash | hex 7-40 | ~2 | After dedup of false positives |
| C8 False positive | `GOOS=(linux\|windows\|darwin)` | 3 (4 hits) | PRESERVE |

Total unique affected files: ~138 (overlaps account for the dedupe).

## Milestone Status

| Milestone | Status | Owner | Trigger |
|---|---|---|---|
| M1 — SPEC scope finalize + allow-list draft | pending | orchestrator-direct | After plan-auditor PASS |
| M2 — C1 macOS-bias fix (4 files / 8 lines) | pending | manager-develop ddd Tier L | After M1 |
| M3 — C2 V3R refs classification + fix (70 files) | pending | manager-develop ddd Tier L | After M2 |
| M4 — Dates + memory + CLAUDE.local refs (59 files) | pending | manager-develop ddd Tier L | After M3 |
| M5 — PR + commit hash + audit script + CI guard | pending | manager-develop ddd Tier L | After M4 |
| M6 — Migration matrix finalize + guideline + chore implemented | pending | orchestrator-direct chore | After M5 |

## Next Steps

After this plan-phase, expected sequence:

1. **plan-auditor** verification of the 4 artifacts (spec.md / plan.md / acceptance.md / progress.md) with Tier L threshold ≥ 0.85.
2. If plan-auditor REVISE → orchestrator-direct iter 2 fix → re-verify.
3. If plan-auditor PASS → commit on main per Hybrid Trunk Tier L doctrine (Section §23.6 multi-commit safe path), then trigger M1.
4. M1 → `migration-matrix.md` 초안 작성 (orchestrator-direct or manager-spec).
5. M2~M5 sequential manager-develop delegation (Tier L Section A-E MANDATORY).
6. M6 chore + status update to `implemented v0.2.0`.

## Evidence Tracker

(Empty — populated during M1–M6 execution)

| AC | M-source | Evidence | Verified |
|---|---|---|---|
| AC-TNA-001 | M2 | (pending) | ☐ |
| AC-TNA-002 | M3 | (pending) | ☐ |
| AC-TNA-003 | M4 | (pending) | ☐ |
| AC-TNA-004 | M4 | (pending) | ☐ |
| AC-TNA-005 | M4 | (pending) | ☐ |
| AC-TNA-006 | M5 | (pending) | ☐ |
| AC-TNA-007 | M5 | (pending) | ☐ |
| AC-TNA-008 | M5 | (pending) | ☐ |
| AC-TNA-009 | M5 | (pending) | ☐ |
| AC-TNA-010 | M1+M6 | (pending) | ☐ |
| AC-TNA-011 | M5 | (pending) | ☐ |
| AC-TNA-012 | M6 | (pending) | ☐ |
| AC-TNA-013 | M6 | (pending) | ☐ |

## Notes

- **fc47f31a7 prior context**: 본 SPEC은 commit `fc47f31a7`의 Critical 4 violations fix 이후 audit에서 추가 발견된 138 file scope를 다룬다. fc47f31a7의 3 변경 (SKILL.md / spec-frontmatter-schema.md / sprint-round-naming.md 신규)은 본 SPEC scope **외** — 이미 머지 완료, 재발 방지는 본 SPEC의 audit script (REQ-TNA-009) + CI guard (REQ-TNA-010)가 담당.
- **Sprint 1 cross-Sprint caveat**: SKILL-GEARS-ALIGN-001 (plan-complete `81d47a445`, 머지 대기) + CODE-COMMENTS-EN-001 Wave 4+ (single-SPEC within-phase, 진행 중)와 disjoint scope. PRESERVE list로 충돌 차단. Sprint = multi-SPEC group; Wave/Round = single-SPEC internal phase (per `.claude/rules/moai/development/sprint-round-naming.md`).
- **GEARS notation self-dogfooding**: §4 13 REQs 모두 Where/When/While/If-Then GEARS keywords 사용. 본 SPEC은 SKILL-GEARS-ALIGN-001 doctrine를 verbatim 적용한다.
- **Frontmatter canonical 12-field SSOT**: 4개 파일 모두 `.claude/rules/moai/development/spec-frontmatter-schema.md`의 12 필드 + optional `tier` 필드 사용. snake_case alias (`created_at`, `updated_at`, `labels`) 미사용.
