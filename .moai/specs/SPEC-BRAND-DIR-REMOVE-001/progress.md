---
id: SPEC-BRAND-DIR-REMOVE-001
title: "Clean Removal of .moai/project/brand Directory — Progress"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P2
phase: "v3.0.0"
module: "internal/harness/safety,internal/hook,internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cleanup, brand, design-retirement, removal, deprecation"
---

# SPEC-BRAND-DIR-REMOVE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-08
- Tier: M (standard)
- 산출물: spec.md + plan.md + acceptance.md + progress.md (4)
- 실측 재검증: 완료 (grep 4종 + DeprecatedPaths 소비 경로 추적 + defaults.go 위치)
- 핵심 결정: dirs.go DeprecatedPaths 항목 **보존**(제거 드라이버) → AC 정련
- scope 결정: (1) migrate Phase 2 제거, (2) design.yaml/DesignConfig OUT(tighter scope), (3) `SentinelHarnessFrozenConfig` 상수 KEEP(8-sentinel 택소노미, D4)
- plan-audit iter1: **FAIL 0.74** → 7 defect(D1~D7) 정정 반영 후 재제출
  - D1: AC-BDR-001a를 in-scope 3사이트 "0 매치"로 정련 + AC-BDR-001c(full-tree 11 survivor enumeration) 추가
  - D2: Scenario 1에서 `config/defaults.go`를 "0 매치" 대상에서 제외(OUT survivor)
  - D3: `safety_preservation_test`(package-harness) IN→무변경 재분류
  - D4: `sentinel_catalog_test` 인벤토리 추가 + `SentinelHarnessFrozenConfig` 상수 보존(REQ-BDR-011)
  - D5: frontmatter `tier: M` 추가(4개 아티팩트) → Tier M 0.80 게이트 적용
  - D6: REQ-BDR 번호 순서 정렬(§B.4 009→010, §B.5 011)
  - D7: migrate summary `"Next step: /moai design"` 제거 M2 편입(REQ-BDR-005b)

## §E.2 Run-phase Evidence

- cycle_type: tdd (RED→GREEN→REFACTOR); Tier M single-SPEC.
- run commit: `ec841d6b5` (isolated worktree) → cherry-picked to main as run-phase commit.
- AC PASS/FAIL matrix (independently re-verified by orchestrator in main checkout):
  - AC-BDR-001a in-scope grep (`internal/harness/ internal/hook/ internal/cli/` non-test) → 0 matches. PASS.
  - AC-BDR-001c full-tree survivors (`internal/ cmd/ pkg/` non-test) → exactly 11 (dirs.go:292 RETAINED + 10 declared-OUT). PASS.
  - AC-BDR-006 dirs.go DeprecatedPaths RETAIN → 1 match (removal driver preserved). PASS.
  - AC-BDR-008 `grep "moai design" migrate_agency.go` → 0 (both L297 + L301 removed). PASS.
  - AC-BDR-004 `moai init` scaffold check → `.moai/project/brand/` NOT created. PASS.
  - md 6개 (template + local) deleted; template dir gone. PASS.
- build: `go build ./...` exit 0; cross-platform `GOOS=windows` build exit 0 (per manager-develop §E2).
- scope tests: `internal/harness/safety` ok, `internal/hook` ok, `internal/cli` migrate_agency ok.
- RETAIN confirmed untouched: dirs.go:292, SentinelHarnessFrozenConfig const, defaults.go, template design.yaml/constitution.md/zone-registry.md.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- All blocking ACs PASS (independently verified). Residual: 6 pre-existing statusline golden failures (TestDoctor_*/TestStatus_*) — verified unrelated to brand removal (identical on pristine 5f33dfaa9 via git stash); out of scope.
- Next: sync-phase (manager-docs) → sync-auditor → 3-phase close (`implemented → completed` on sync commit).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
