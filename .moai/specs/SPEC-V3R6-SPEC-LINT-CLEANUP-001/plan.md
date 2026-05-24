---
id: SPEC-V3R6-SPEC-LINT-CLEANUP-001
title: "spec-lint MissingExclusions baseline cleanup — Implementation Plan (Tier S minimal Section A)"
version: "0.1.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, missing-exclusions, baseline-cleanup, h3-pattern, retroactive, tier-s"
sync_commit_sha: "0d777471c21f36f827752608ea6b7bcceea09fd8"
---

# SPEC-V3R6-SPEC-LINT-CLEANUP-001 — Implementation Plan (Tier S minimal Section A)

## §A. Context

### §A.1 Tier classification + 1-pass justification

**Tier S (Simple)** — `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier 기준:

| Tier 판정 항목 | 본 SPEC 값 | Tier S 임계값 | 충족 여부 |
|---------------|----------|-------------|---------|
| LOC scope | < 300 LOC (markdown만, `.go` 파일 0개) | < 300 | ✓ |
| Files affected (run-phase) | 8 sibling `spec.md` only | < 5 (guideline) | partial — 8 files but all 동일 패턴 surgical edit |
| Risk profile | markdown H3 추가, code change 없음 | Low | ✓ |
| AC 개수 | 7 ACs (manageable) | n/a | ✓ |

**1-pass plan-phase 정당화**:
- 본 SPEC plan-phase는 어떤 code edit도 수행하지 않으며, 8개 sibling SPEC 식별 + canonical 패턴 codification + run-phase scope contract 명시에 한정. Section B-E (Known Issues / Pre-flight / Constraints / Self-Verification deliverables)는 run-phase delegation 시점에 적용되며 plan-phase 자체에는 불필요.
- 선례 정합성: ARR-001 + PROPOSAL-GEN-001 (둘 다 Tier S minimal) 모두 plan-phase 1-pass 성공. 본 SPEC도 동일 cohort 진입.
- plan-auditor PASS 임계값: Tier S 최소 **0.75** (workflow doctrine), skip-eligible 0.90 (Phase 0.5).

### §A.2 SPEC 산출물 경로 + 라인 카운트 (plan-phase 종료 시점 예상)

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main` (Hybrid Trunk per CLAUDE.local.md §23.7)
- HEAD SHA at plan-phase start: `e34e1d750d324f8e94750f40224021e0ada53b08`
- SPEC artifacts (4 files):
  - `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` — canonical SSOT (≈115 lines)
  - `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/plan.md` — this file (Tier S minimal Section A only)
  - `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/acceptance.md` — 7 ACs (REQ-SLC-001..007)
  - `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md` — lifecycle + audit-ready signal
- plan-auditor verdict: TBD at plan-phase write completion (target: PASS ≥ 0.75 Tier S 임계값).

### §A.3 Multi-session race context (2026-05-25 시점)

본 plan-phase는 SPEC-V3R6-MULTI-SESSION-COORD-001 run-phase 와 동시 진행. COORD-001은 `internal/session/*`, `internal/hook/session_start.go`, `internal/cli/session*.go`를 수정 중. 본 SPEC scope는 완전 disjoint:

- 본 SPEC plan-phase write target: `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/**` 4 파일만.
- 본 SPEC run-phase write target (deferred): `.moai/specs/{8 sibling}/spec.md` 8 파일.
- COORD-001 scope: `internal/{session,cli,hook}/**` + `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/**`.

→ 양 SPEC 간 file ownership 충돌 zero.

### §A.4 Existing infrastructure (PRESERVE vs EXTEND)

**PRESERVE (plan-phase 수정 금지) — 17 dirty/untracked entries (verified `git status --porcelain | wc -l` at plan-phase start)**:

| # | Path | Type | Source |
|---|------|------|--------|
| 1 | `.moai/config/sections/git-convention.yaml` | M | dev settings (§22) |
| 2 | `.moai/config/sections/language.yaml` | M | dev settings |
| 3 | `.moai/config/sections/quality.yaml` | M | dev settings |
| 4 | `.moai/harness/usage-log.jsonl` | M | runtime-managed |
| 5 | `internal/hook/session_start.go` | M | COORD-001 run-phase active |
| 6 | `.moai/harness/learning-history/` | ?? | runtime-managed |
| 7 | `.moai/harness/observations.yaml` | ?? | runtime-managed |
| 8 | `.moai/research/anthropic-best-practices-2026-05-24.md` | ?? | parallel session audit |
| 9 | `.moai/research/v3.0-redesign-2026-05-23.md` | ?? | parallel session research |
| 10 | `i18n-validator` | ?? | parallel session artifact |
| 11 | `internal/cli/session.go` | ?? | COORD-001 run-phase active |
| 12 | `internal/cli/session_test.go` | ?? | COORD-001 run-phase active |
| 13 | `internal/hook/session_start_multisession_test.go` | ?? | COORD-001 run-phase active |
| 14 | `internal/session/registry.go` | ?? | COORD-001 run-phase active |
| 15 | `internal/session/registry_lock_unix.go` | ?? | COORD-001 run-phase active |
| 16 | `internal/session/registry_lock_windows.go` | ?? | COORD-001 run-phase active |
| 17 | `internal/session/registry_test.go` | ?? | COORD-001 run-phase active |
| 18 | `internal/session/subagent_boundary_test.go` | ?? | COORD-001 run-phase active |

→ 본 plan-phase commit은 위 18개 모두 staged 영역에 포함 금지. `git add .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/` path-specific add만 사용.

**EXTEND (plan-phase) — 4 entries exactly**:

| # | Path | Type | Operation |
|---|------|------|-----------|
| 1 | `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` | NEW | plan-phase에서 manager-spec이 신규 작성 |
| 2 | `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/plan.md` | NEW | this file |
| 3 | `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/acceptance.md` | NEW | 7 ACs matrix |
| 4 | `.moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/progress.md` | NEW | lifecycle + audit-ready signal |

**EXTEND (future run-phase) — 8 entries deferred**:

| # | Path | Lint failure type | 분류 | 예상 edit |
|---|------|------------------|------|---------|
| 1 | `.moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/spec.md` | "section missing" | A | `## Exclusions` 아래 `### N.1 Out of Scope — <topic>` H3 추가 + 기존 list item 이동 |
| 2 | `.moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/spec.md` | "section has no items" | B | 진단 후 H3 보강 또는 list item 추가 |
| 3 | `.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md` | "section has no items" | B | `### §A.6 Out-of-scope but related` 형식 텍스트의 H3 보강 + `## §C. Exclusions`에 H3 sub-section 추가 |
| 4 | `.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-002/spec.md` | "section has no items" | B | 진단 후 H3 보강 |
| 5 | `.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-003/spec.md` | "section has no items" | B | 진단 후 H3 보강 |
| 6 | `.moai/specs/SPEC-V3R6-PROMPT-CACHE-001/spec.md` | "section has no items" | B | 진단 후 H3 보강 |
| 7 | `.moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/spec.md` | "section has no items" | B | 진단 후 H3 보강 |
| 8 | `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md` | "section missing" | A | `## §2 Non-Goals` 아래 `### §2.1 Out of Scope — <topic>` H3 추가 |

### §A.5 Run-phase delegation contract

future `/moai run SPEC-V3R6-SPEC-LINT-CLEANUP-001` 실행 시 manager-develop 위임 prompt는 다음을 포함해야 한다 (Section A-E minimal form per Tier S):

- Section A (Context): 8 sibling SPEC enumeration (위 §A.4 EXTEND table) + canonical H3 패턴 reference (spec.md §3).
- Section D (Constraints): REQ-SLC-004 (8 sibling spec.md만 수정) + REQ-SLC-005 (의미 변경 금지). 그 외 SPEC 및 plan/acceptance/progress 파일 수정 절대 금지.
- Section E (Self-Verification): REQ-SLC-006 binary verification = `moai spec lint 2>&1 | grep -c MissingExclusions` 출력값 **0**.

### §A.6 Risks

| # | Risk | Mitigation |
|---|------|-----------|
| R1 | run-phase에서 H3 텍스트만 추가하고 list item 추가를 누락 | AC-SLC-002 (acceptance.md) binary verification으로 차단 |
| R2 | run-phase가 8개 외 SPEC도 손대어 scope creep 발생 | REQ-SLC-004 [State-Driven] + Section D Constraints 명시 + post-run `git diff --name-only` audit |
| R3 | list item 본문 텍스트 의미 변경(semantic drift) | REQ-SLC-005 [Unwanted] + run-phase pre/post diff inspection |
| R4 | parallel session이 신규 SPEC을 작성하면서 `MissingExclusions` 추가 발생 | REQ-SLC-006는 baseline-cleanup 시점 기준. 신규 SPEC failure는 별도 사이클에서 처리. acceptance.md AC-SLC-006a 노트 |
| R5 | 8 sibling SPEC 중 일부가 historical archive 대상 (예: LEGACY-CLEANUP-001..003) | 본 SPEC은 archive 결정과 무관. archive 정책은 별도 SPEC. cleanup 의무는 archive 여부와 독립 |

## §B-§E (Tier S minimal — intentionally omitted)

본 SPEC은 Tier S minimal 변형이므로 plan.md Section B (Known Issues 8 카테고리) / Section C (Pre-flight checklist) / Section D (Constraints) / Section E (Self-Verification deliverables)는 plan.md 본문에서 생략한다. 이들 섹션의 내용은 run-phase delegation prompt 구성 시 §A.5에 따라 minimal form으로 inline 주입된다. 선례: ARR-001 + PROPOSAL-GEN-001 모두 동일 변형 적용.

## §F. Cross-references

- `internal/spec/lint.go:678-728` — `OutOfScopeRule.Check()` algorithm.
- spec.md §3 canonical H3 pattern definition.
- acceptance.md — REQ ↔ AC traceability matrix (7 ACs).
- progress.md — Phase 0.5 plan-auditor verdict 기록.
- ARR-001 + PROPOSAL-GEN-001 plan.md — Tier S minimal Section A 변형 선례.
