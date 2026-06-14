# SPEC-SEC-HARDEN-005 — Progress

> 4-phase 진행 신호 보드. plan-phase는 manager-spec이 scaffold; `plan_complete_at` / `plan_status: audit-ready`는 plan-auditor PASS 후 orchestrator가 append.

## §A — Plan-phase Context

- **SPEC**: SPEC-SEC-HARDEN-005 — SEC-HARDEN §F residual containment (${IFS} shell-aware word-split + update env-trust allowlist)
- **Tier**: M (Medium) — plan-auditor PASS threshold 0.80
- **era**: V3R6
- **선행**: SEC-HARDEN-001/002/003/004 (전부 `completed`)
- **산출물 (5-file set)**: spec.md + plan.md + acceptance.md + design.md + progress.md
- **cycle_type (run-phase)**: tdd (quality.yaml development_mode: tdd)
- **branch 전략**: main 직진 (Hybrid Trunk 1-person OSS; --worktree/--branch 미사용)
- **수정 표면**:
  - §F.1 (PRIMARY): `internal/permission/stack.go` — `hasUnquotedShellSeparator`(L172) + `Matches` `:*` 브랜치(L100,127-136). NEW dep `mvdan.cc/sh/v3/syntax`.
  - §F.2 (PRIMARY): `internal/cli/deps.go` — `EnsureUpdate`(L250-309) env-read 블록. scheme+host allowlist.
  - §F.3 (OPTIONAL, 비요구): `restoreTargetContained`/`parentChainContained`/`runMXScan` godoc TOCTOU note. 코드 동작 변경 없음.

## §B — Plan-phase Self-Check

- [x] SPEC ID Pre-Write Self-Check: `decomposition: SPEC ✓ | SEC ✓ | HARDEN ✓ | 005 ✓ → PASS`
- [x] Frontmatter 12 canonical fields + era:V3R6 + tier:M, status: draft
- [x] `created:`/`updated:`/`tags:` (snake_case alias 없음)
- [x] GEARS format requirements (Ubiquitous / Event-driven; canonical 2-form per plan-auditor D4 정정)
- [x] Exclusions section §F with h3 sub-sections (MissingExclusions 회피)
- [x] design.md 포함 (신규 의존성 보안-게이트 통합 설계)
- [x] AC 13개 (재현 5 + 회귀 4 + fail-closed 1 + 의존성/범위 2 + 전역 NFR 1), 모든 grep AC `$` anchor + 명시 테스트 실행 (non-vacuous)
- [x] OPT-SEC5-001 (TOCTOU)는 OPTIONAL, AC 게이트 아님
- [x] anti-over-engineering: 새 패키지/플래그 금지(예외 mvdan.cc/sh + 검증 로직)
- [x] spec-lint clean (orchestrator 검증 — `✓ No findings`)
- [x] plan-auditor PASS-WITH-DEBT 0.86 ≥ 0.80 (Tier M); D1 BLOCKING + D2/D3/D4 전부 orchestrator-direct 정정

## §C — Milestone Tracker (run-phase, manager-develop)

| Milestone | 설명 | 상태 | commit SHA |
|-----------|------|------|------------|
| M1 | mvdan.cc/sh dep + §F.1 ${IFS} RED + legit baseline 고정 | pending | — |
| M2 | §F.1 GREEN — hasIFSWordSplit 헬퍼 + Matches 배선 | pending | — |
| M3 | §F.2 RED+GREEN — update env-trust allowlist | pending | — |
| M4 | §F.3 OPTIONAL godoc + 전체 검증 batch | pending | — |

## §D — Phase 0.5 SKIP Rationale (placeholder)

- plan-auditor verdict: PASS-WITH-DEBT 0.86 (Clarity 0.90 / Completeness 0.92 / Testability 0.74 / Traceability 1.00; MP-1..MP-4 PASS). 보고서: `.moai/reports/plan-audit/SPEC-SEC-HARDEN-005-2026-06-14.md`. 결함 D1 BLOCKING(C-HRA-008 grep idiom 불능) + D2 SHOULD-FIX(TestX$ trailing-`$` 경계 미고정) + D3/D4 MINOR 전부 orchestrator-direct 정정 (D1 canonical 필터 0 반환 + spec-lint clean 독립 검증).
- SKIP 적용 여부: **NOT skip-eligible (0.86 < 0.90)** → run-phase `/moai run` Phase 0.5 plan-auditor 재실행 필수.
- GATE-2 (plan→run HUMAN GATE): score 무관 — 사용자 명시 승인 필수(skip-eligible 0.90 autonomous bypass는 Phase 0.5 verdict 재실행에만 적용, GATE-2에는 미적용).

## §E — Audit-Ready Signals (4-phase, append-only)

> 각 phase 완료 후 해당 agent가 append. plan-phase signal은 plan-auditor PASS 후 orchestrator가 기록.

### §E.1 Plan-phase Audit-Ready Signal
- plan_complete_at: 2026-06-14
- plan_status: audit-ready
- plan_commit_sha: 328ff95e3

### §E.2 Run-phase Evidence
- _(manager-develop append — REQ-ARR-002)_

### §E.3 Run-phase Audit-Ready Signal
- _(manager-develop append)_

### §E.4 Sync-phase Audit-Ready Signal
- _(manager-docs append — REQ-ARR-003)_

### §E.5 Mx-phase Audit-Ready Signal
- _(manager-docs OR orchestrator append)_
- mx_commit_sha: _(pending)_
