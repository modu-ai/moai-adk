---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Progress Tracking"
version: "0.1.1"
status: draft
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, progress"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Progress Tracking

## Status

**Phase**: plan-phase (draft v0.1.1) — M1 already executed + committed; plan-audit iter-1 remediation applied
**Created**: 2026-05-23
**Last update**: 2026-05-30 (plan-audit iter-1 remediation: rescope to NEUTRALITY-unique C1/C2/C4/C5/C6/C8 + C3/C7 deferred to ISOLATION-001 + AC awk fix + baseline refresh + M1 marked complete)

## Plan-Phase Summary

Tier L SPEC. 13 GEARS REQs (11 active + REQ-TNA-003/007 deferred to ISOLATION-001) + 11 active binary ACs (+ AC-TNA-003/007 deferred markers) + 6 milestones (M1–M6). M1 already executed (commit `367a84715`); run-phase enters at M2.

### Rescope (v0.1.1, plan-audit iter-1 — FAIL 0.71 remediation)

- **KEPT in NEUTRALITY** (unique, not covered by ISOLATION leak test): C1 macOS paths, C2 `V3R[0-9]` refs (the bulk, 73 files), C4 `feedback_`/`memory.md` refs, C5 `CLAUDE.local.md` refs, C6 `PR #N` refs, C8 `GOOS=` PRESERVE.
- **DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001** (`internal_content_leak_test.go` strict-tier owns these): C3 dates (`S1-internal-date`), C7 commit hash (`S2-short-sha-sentence-final`).
- **C4 kept despite SCOPE defer instruction**: the leak test does NOT enforce the `feedback_`/`memory.md` substring class (default OR strict) — its C5 enforces only memory *paths*, a disjoint pattern. Deferring C4 would silently drop enforcement, so C4 stays NEUTRALITY-owned (verified 2026-05-30).
- **D2 AC awk fix**: AC-TNA-002/004/005 rewritten with non-self-terminating awk `awk '/^### CN /{f=1;next} /^### C[0-9] /{f=0} f'` (the original range `awk '/^### CN /,/^### C[0-9]+/'` self-terminated → allow-list count = 0 → `actual <= 0` impossible).

### Deliverables (this phase)

- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md` (§1–§5)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/plan.md` (§1 Milestones + §2 Risks + §3 Dependencies + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md` (§1 ACs + §2 Test plan + Out of Scope)
- `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/progress.md` (this file)

### Baseline measurements (re-measured 2026-05-30 at HEAD `ecda4ef04`, absolute-path grep — point-in-time; run-phase M-stages re-measure before fixing)

| Category | Pattern | Files (2026-05-30) | Files (2026-05-23) | Partition |
|---|---|--:|--:|---|
| C1 macOS-bias | `/Users/` | 4 | 4 | KEEP (binary) |
| C2 V3R refs | `V3R[0-9]` | **73** | 70 | KEEP (advisory, allow-list 15) |
| C3 Dates | `2026-0[5-9]` | 39 | 32 | **DEFER → ISOLATION** (informational) |
| C4 feedback/memory | `feedback_\|memory\.md` | 9 | 9 | KEEP (advisory, allow-list 7) |
| C5 CLAUDE.local | `CLAUDE\.local\.md` | **3** | 10 | KEEP (binary, allow-list empty) |
| C6 PR #N | `PR #[0-9]+` | 3 | 3 | KEEP (binary) |
| C7 Commit hash | hex 7-8 (leak-test S2) | ~2 | ~2 | **DEFER → ISOLATION** (informational) |
| C8 False positive | `GOOS=(linux\|windows\|darwin\|...)` | 3 (4 hits) | 3 (4 hits) | KEEP (PRESERVE) |

KEPT-class scope (C1/C2/C4/C5/C6/C8). Drift since plan-phase: C2 +3, C3 +7 (deferred), C5 −7 (partial prior cleanup). Run-phase M-stages re-measure at run HEAD before applying fixes.

## Milestone Status

| Milestone | Status | Owner | Trigger |
|---|---|---|---|
| M1 — SPEC scope finalize + allow-list draft | **complete** (commit `367a84715`; migration-matrix.md 232L, 8/8 sections, AC-TNA-010 PASS) | orchestrator-direct | (done) |
| M2 — C1 macOS-bias fix (4 files) | pending | manager-develop ddd Tier L | After plan-audit PASS + GATE-2 |
| M3 — C2 V3R refs classification + fix (73 files) | pending | manager-develop ddd Tier L | After M2 |
| M4 — memory + CLAUDE.local refs (C4+C5, ~12 files; C3 deferred) | pending | manager-develop ddd Tier L | After M3 |
| M5 — PR refs + audit script + CI guard (C6; C7 deferred) | pending | manager-develop ddd Tier L | After M4 |
| M6 — Migration matrix finalize + guideline + chore implemented | pending | orchestrator-direct chore | After M5 |

## Next Steps

After this plan-audit iter-1 remediation (v0.1.1), expected sequence:

1. **plan-auditor iter-2** verification of the 5 artifacts (spec.md / plan.md / acceptance.md / migration-matrix.md / progress.md) with Tier L threshold ≥ 0.85. Expected to clear ≥ 0.85: the rescope deletes D1 (ISOLATION overlap), the awk fix resolves D2, the C7 defer shrinks D3, M1-complete resolves D4, baselines refresh resolves D6.
2. If plan-auditor PASS → GATE-2 (plan-to-implement HUMAN GATE) → run-phase enters at **M2** (M1 already complete, commit `367a84715`).
3. M2 (C1) → M3 (C2) → M4 (C4+C5; C3 deferred) → M5 (C6 + audit-script + CI-guard; C7 deferred) — sequential manager-develop delegation (Tier L Section A-E MANDATORY).
4. M6 chore + status update to `implemented v0.2.0`.

## Evidence Tracker

M1 evidence populated (migration-matrix.md shipped `367a84715`). M2–M6 populated during run-phase.

| AC | M-source | Evidence | Verified |
|---|---|---|---|
| AC-TNA-001 | M2 | (pending) | ☐ |
| AC-TNA-002 | M3 | corrected awk allow-list=15 computable (D2 fix verified 2026-05-30) | ☐ (run-phase) |
| AC-TNA-003 | — | **DEFERRED → ISOLATION-001** (leak-test `S1-internal-date`) | n/a |
| AC-TNA-004 | M4 | corrected awk allow-list=7 computable (D2 fix verified 2026-05-30) | ☐ (run-phase) |
| AC-TNA-005 | M4 | binary; allow-list empty (corrected awk=0) | ☐ (run-phase) |
| AC-TNA-006 | M5 | (pending) | ☐ |
| AC-TNA-007 | — | **DEFERRED → ISOLATION-001** (leak-test `S2-short-sha-sentence-final`) | n/a |
| AC-TNA-008 | M5 | (pending) | ☐ |
| AC-TNA-009 | M5 | (pending) | ☐ |
| AC-TNA-010 | M1 | **PASS** — migration-matrix.md 8/8 sections, `grep -cE '^### C[1-8] '` = 8 (commit `367a84715`, re-verified 2026-05-30) | ☑ |
| AC-TNA-011 | M5 | C8 baseline = 3 files preserved (verified 2026-05-30) | ☐ (run-phase) |
| AC-TNA-012 | M6 | (pending) | ☐ |
| AC-TNA-013 | M6 | (pending) | ☐ |

## Notes

- **fc47f31a7 prior context**: 본 SPEC은 commit `fc47f31a7`의 Critical 4 violations fix 이후 audit에서 추가 발견된 138 file scope를 다룬다. fc47f31a7의 3 변경 (SKILL.md / spec-frontmatter-schema.md / sprint-round-naming.md 신규)은 본 SPEC scope **외** — 이미 머지 완료, 재발 방지는 본 SPEC의 audit script (REQ-TNA-009) + CI guard (REQ-TNA-010)가 담당.
- **Sprint 1 cross-Sprint caveat**: SKILL-GEARS-ALIGN-001 (plan-complete `81d47a445`, 머지 대기) + CODE-COMMENTS-EN-001 Wave 4+ (single-SPEC within-phase, 진행 중)와 disjoint scope. PRESERVE list로 충돌 차단. Sprint = multi-SPEC group; Wave/Round = single-SPEC internal phase (per `.claude/rules/moai/development/sprint-round-naming.md`).
- **GEARS notation self-dogfooding**: §4 13 REQs 모두 Where/When/While/If-Then GEARS keywords 사용. 본 SPEC은 SKILL-GEARS-ALIGN-001 doctrine를 verbatim 적용한다.
- **Frontmatter canonical 12-field SSOT**: 4개 파일 모두 `.claude/rules/moai/development/spec-frontmatter-schema.md`의 12 필드 + optional `tier` 필드 사용. snake_case alias (`created_at`, `updated_at`, `labels`) 미사용.
