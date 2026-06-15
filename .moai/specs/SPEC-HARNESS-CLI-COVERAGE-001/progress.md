---
id: SPEC-HARNESS-CLI-COVERAGE-001
title: "internal/cli/harness 테스트 커버리지 ≥90% 상향 — 진행 추적"
version: "0.1.0"
status: draft
created: 2026-06-15
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/harness"
lifecycle: spec-anchored
tags: "harness, coverage, test, tdd, self-harness"
era: V3R6
tier: S
---

# SPEC-HARNESS-CLI-COVERAGE-001 — 진행 추적 (progress.md)

## §A — Plan-phase

- plan-phase 산출물: spec.md + plan.md + acceptance.md + progress.md (이 커밋)
- cycle_type: tdd / Tier S (minimal, test-only)
- ground-truth baseline (spec.md §B SSOT): 패키지 total **77.9%**
  - `NewInstallCmd` 33.3% (최대 gap) / `NewExecuteCmd` 46.2% / `runExecuteCommand` 62.5% /
    `runPropose` 78.1% / `RunInstall` 80.0% / `RunExecute` 87.5% / `runExecuteWith` 90.0%
- 목표: 패키지 total ≥ **90.0%** (test-only, 프로덕션 non-test `.go` 무수정 원칙)
- os.Exit 분기 결정 (plan.md §D): **(a) documented acceptable residual** — 서브프로세스 re-exec
  미도입. 로직은 `runExecuteCommand`에 추출되어 별도 커버됨.
- [ ] plan-auditor gate (Tier S thresh 0.75; skip-eligible 아님 — 사용자 구현 착수 승인 필요)

## §B — Run-phase (manager-develop, cycle_type=tdd)

- [ ] M1 — `NewInstallCmd` RunE 클로저 (install.go:125-155): success / default-cwd(non-parallel) / RunInstall-error
- [ ] M2 — `RunInstall` error 분기 (install.go:61-63, 74-76) → 100%
- [ ] M3 — `runExecuteCommand` success stdout + InheritedFlags success (execute.go:358-360, 366-369)
- [ ] M4 — `runPropose` non-dry-run WriteProposals + default-flag fallback (propose.go:80-95, 125-137)
- [ ] M5 — `RunExecute`/`runExecuteWith` 잔존 error 분기 (도달 가능분만; getwd/Abs 실패는 residual 후보)
- [ ] M6 — 통합 검증 + residual 문서화 + traceability

### Run-phase 검증 게이트 (각 milestone)
- `go test -coverprofile` 재측정 → 목표 함수 커버리지 증가
- `go test ./...` 전체 green (회귀 catch)
- `git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'` → 출력 없음 (프로덕션 무수정)

## §E.2 Sync-phase Audit-Ready Signal

> (manager-docs가 sync-phase에서 채움 — sync_commit_sha + CHANGELOG 진입점 + status 전이)

- **sync_commit_sha**: (pending — manager-docs 기록)
- **sync_deliverables**: (pending)
  - CHANGELOG.md [Unreleased] 적절 섹션 진입점 추가 (커버리지 77.9%→≥90%, test-only, N AC)
  - spec.md frontmatter status: draft → implemented
  - progress.md §E.2 sync audit-ready signal 기록
- **Trust-but-verify**: (pending)

## §E.3 Lifecycle Status

- **status**: draft (현재) → in-progress (M1 commit) → implemented (sync) → completed (Mx)
- **era**: V3R6 (H-5 tie-breaker: created 2026-06-15 ≥ 2026-04-01; H-4는 §E.2/§E.5 SHA 충족 후)

## §E.5 Mx-phase Audit-Ready Signal

> (manager-docs 또는 orchestrator가 Mx-phase 4-phase close에서 채움)

- **mx_commit_sha**: (pending — close commit)
- **mx_deliverables**: (pending)
  - @MX 코드 주석 검토 (test-only — 신규 @MX 태그 가능성 낮음, 기존 godoc 충분 예상)
  - sync-auditor 검증 (Functionality/Security MUST-PASS)
  - spec.md status: implemented → completed (4-phase close)
- **documented residual** (AC-HCC-009 — sync/Mx 시 최종 확정):
  - `execute.go:329-334` os.Exit 분기 (표준 테스트 도달 불가)
  - `execute.go:344-345` MarkFlagRequired panic (방어적)
  - `install.go:168-169` MarkFlagRequired panic (방어적)
  - (조건부) `execute.go:116/124/159` getwd/Abs 실패 분기 (결정론적 유발 불가 시)
- **4-phase 완결**: plan (이 커밋) + run (M1~M6) + sync (§E.2) + Mx (§E.5)
