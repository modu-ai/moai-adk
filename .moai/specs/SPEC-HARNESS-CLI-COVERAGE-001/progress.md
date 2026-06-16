---
id: SPEC-HARNESS-CLI-COVERAGE-001
title: "internal/cli/harness 테스트 커버리지 ≥90% 상향 — 진행 추적"
version: "0.1.0"
status: in-progress
created: 2026-06-15
updated: 2026-06-16
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
- [x] plan-auditor gate: **PASS-WITH-DEBT 0.86** (Tier S thresh 0.75 통과, iter-1) — baseline 77.9% 독립 일치, 90% test-only 도달 가능 입증(상한 95.3%). debt: D1(AC-008a grep)/D2(§B.3 :74→:77) → manager-spec 정정 완료(`845191b8d`), D3(NewExecuteCmd Execute() 경로) → run-phase M3 반영. 구현 착수 승인 **GRANTED** (사용자 informed consent, verdict 제시 후).

## §B — Run-phase (manager-develop, cycle_type=tdd)

- [x] M1 — `NewInstallCmd` RunE 클로저 (install.go:125-155): success / default-cwd(non-parallel) / RunInstall-error — 신규 `install_cmd_test.go` 3 테스트. `NewInstallCmd` 33.3% → 87.5%
- [x] M2 — `RunInstall` error 분기 (install.go:61-63, 74-76) → 100% — `install_test.go`에 EmptyProjectRoot + ScaffoldFails 2 테스트 추가. `RunInstall` 80.0% → 100.0%
- [x] M3 — `runExecuteCommand` success stdout + InheritedFlags success (execute.go:358-360, 366-369) + NewExecuteCmd RunE success (D3) — 신규 `execute_cmd_test.go` 3 테스트. `runExecuteCommand` 62.5% → 100.0%, `NewExecuteCmd` 46.2% → 69.2%
- [x] M4 — `runPropose` non-dry-run WriteProposals + default-flag fallback (propose.go:80-95, 125-137) + read/write error 분기 — 신규 `propose_run_test.go` 4 테스트. `runPropose` 78.1% → 93.8%
- [x] M5 — `RunExecute`/`runExecuteWith` 잔존 error 분기 (execute.go:116/124/159 getwd/Abs 실패): test-only 결정론적 유발 불가 → §E.5 documented residual로 명시(강제 미도달, AP-2 회피)
- [x] M6 — 통합 검증 + residual 문서화 + traceability — package total 77.9% → **93.0%** (AC-HCC-001 PASS)

### Run-phase 측정 결과 (관측값 — `go test -coverprofile=/tmp/hcc-final.out ./internal/cli/harness/...`)
- package total: **93.0%** (baseline 77.9% → +15.1pp)
- `NewInstallCmd` 87.5% / `RunInstall` 100.0% / `runExecuteCommand` 100.0% / `runPropose` 93.8% / `NewExecuteCmd` 69.2% / `RunExecute` 87.5% / `runExecuteWith` 90.0%
- `go test ./...` 전체 green (exit 0), `go vet ./...` clean, `golangci-lint run` 0 issues(baseline 유지), `go test -race` clean
- cross-build: `go build ./...` + `GOOS=windows`/`GOOS=linux` 모두 exit 0
- `git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'` → 출력 없음 (프로덕션 non-test `.go` 무수정 — AC-HCC-010 PASS)
- 회귀 가드 green: `TestPropose_NoAskUserQuestion`(AC-HCC-008a) + `execute_e2e_test.go`(AC-HCC-011)

### §B.1 — AC 결과 매트릭스 (관측값)

| AC | 목표 | 관측 결과 | 상태 |
|----|------|-----------|------|
| AC-HCC-001 | package ≥90% | 93.0% | PASS |
| AC-HCC-002 | NewInstallCmd ≥87.5% (Option A 재조정) | 87.5% | **PASS** (`845191b8d`) |
| AC-HCC-003 | RunInstall 100% | 100.0% | PASS |
| AC-HCC-004 | runExecuteCommand ≥90% | 100.0% | PASS |
| AC-HCC-005 | runPropose ≥90% | 93.8% | PASS |
| AC-HCC-006 | NewExecuteCmd ≥69.2% (Option A 재조정) | 69.2% | **PASS** (`845191b8d`) |
| AC-HCC-007 | 격리 (t.TempDir + os.Chdir non-parallel) | race clean | PASS |
| AC-HCC-008a | TestPropose_NoAskUserQuestion green | PASS | PASS |
| AC-HCC-009 | documented residual 열거 | §E.5 enumerated | PASS |
| AC-HCC-010 | 프로덕션 non-test `.go` 무수정 | diff empty | PASS |
| AC-HCC-011 | execute_e2e_test.go green (fail-close 무회귀) | PASS | PASS |
| AC-HCC-012 | test/vet/lint/cross-build green | 모두 exit 0 | PASS |

### §B.2 — AC-HCC-002 / AC-HCC-006 미달 원인 (spec-internal contradiction — **RESOLVED via Option A**)

> **RESOLUTION (사용자 승인 Option A, 2026-06-16)**: 두 per-function AC를 reachable ceiling로 재조정(AC-002 ≥87.5%, AC-006 ≥69.2%) — manager-spec commit `845191b8d`. test-only 유지·프로덕션 무수정. 패키지 ≥90% 주 게이트는 93.0%로 충족. AC 매트릭스 12/12 green(현실 정렬). 아래는 blocker 분석 원본(이력 보존).

두 per-function AC는 **test-only로 수학적 도달 불가**다. 미커버 잔존이 전부 §E.5 documented
residual(os.Exit / getwd / Abs / panic)이고, 그 residual이 함수 statement의 >10%를 차지하기 때문이다:

- **AC-HCC-002 (NewInstallCmd ≥90%)**: 함수 24 statement 중 3개 미커버 = 87.5% ceiling.
  미커버 3개: `install.go:129-131`(os.Getwd 실패), `137-139`(filepath.Abs 실패), `168-169`(MarkFlagRequired panic).
  3개 모두 표준 Go 테스트 도달 불가(getwd는 macOS에서 cwd 삭제 후에도 실패 안 함을 실측 확인;
  Abs는 상대경로+getwd 실패 시에만 에러; panic은 방어적). → ≥90%는 1개 미커버만 허용하나 3개가 unreachable.
- **AC-HCC-006 (NewExecuteCmd ≥90%)**: 함수 13 statement 중 4개 미커버 = 69.2% ceiling.
  미커버 4개: `execute.go:329-334`(os.Exit body 3 stmt — `SilenceUsage`+`SilenceErrors`+`os.Exit`), `344-345`(panic 1 stmt).
  AC-HCC-006 본문이 "미커버 잔존은 os.Exit(329-334) + panic(344-345)에 한정"이라고 명시하나,
  그 4 statement = 함수의 31%이므로 "≥90% AND 잔존=os.Exit+panic"은 자기모순(둘 다 동시 성립 불가).
  os.Exit body가 1 statement라는 D3 전제가 under-count(실제 3 statement). RunE success 경로(327 entry + 335 return nil)는 도달 완료.

해소 경로(둘 다 manager-develop 단독 권한 밖):
- (A) acceptance.md AC-HCC-002/006 임계값 재조정(NewInstallCmd 85% / NewExecuteCmd 65% 등 reachable ceiling으로) — manager-spec 도메인(SPEC body 수정).
- (B) REQ-HCC-022 seam(injectable exitFunc 프로덕션 도입) — AC-HCC-008b 발동 → plan-auditor 재검토 + 사용자 승인.
→ blocker report 반환. 나머지 10 AC는 전부 PASS, package ≥90%(주 게이트) 달성.

### Run-phase 검증 게이트 (각 milestone)
- `go test -coverprofile` 재측정 → 목표 함수 커버리지 증가
- `go test ./...` 전체 green (회귀 catch)
- `git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'` → 출력 없음 (프로덕션 무수정)

## §C — Phase 0.95 Mode Selection

- **Input parameters**: tier=S, scope=1 package (`internal/cli/harness`, 소스 3파일 ~694 LOC), domain count=1 (Go 테스트), file language=100% Go, concurrency benefit=LOW (coding-heavy/new-code).
- **Mode evaluation**: Mode 1 trivial / 2 background / 3 agent-team / 4 parallel = not selected. Mode 6 workflow = **not selected (PROHIBITED for coding/new-code per `orchestration-mode-selection.md` §E; <30 files, mechanical-uniform 아님)**. Mode 5 sub-agent = **selected** (default fallback).
- **Decision**: sub-agent
- **Justification**: Test-only new-code, 단일 소형 패키지. Anthropic coding-task parallelism caveat → sequential sub-agent(manager-develop)가 기본값. `ultracode` 키워드에도 Mode 6 workflow는 new-code에 HARD-prohibited — ultracode 철저함은 plan-auditor + sync-auditor + Trust-but-verify 다중 검증으로 충족. 구현 착수 승인 사용자 승인 완료(informed consent).

## §E.2 Sync-phase Audit-Ready Signal

> (manager-docs가 sync-phase에서 채움 — sync_commit_sha + CHANGELOG 진입점 + status 전이)

- **sync_commit_sha**: (sync commit SHA — 이 커밋)
- **sync_deliverables**: 
  - CHANGELOG.md [Unreleased] ### Added 섹션: SPEC-HARNESS-CLI-COVERAGE-001 진입점 추가 (coverage 77.9%→93.0%, test-only, AC 12/12 PASS)
  - spec.md frontmatter: `status: in-progress` → `status: implemented` + `updated: 2026-06-16`
  - progress.md §E.2/§E.3 sync audit-ready signal 기록 (본 섹션)
- **Trust-but-verify**: 
  - `go test -coverprofile=/tmp/hcc-sync.out ./internal/cli/harness/... && go tool cover -func=/tmp/hcc-sync.out | tail -1` → **93.0%** ✓
  - `go test ./... 2>&1 | tail -5` → **all green, 0 failures** ✓
  - `git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'` → **empty (프로덕션 무수정)** ✓

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
- **documented residual** (AC-HCC-009 — run-phase 관측으로 최종 확정):
  - `execute.go:329-334` `NewExecuteCmd` os.Exit body (3 stmt: `SilenceUsage`/`SilenceErrors`/`os.Exit`) — 표준 Go 테스트 도달 불가(os.Exit가 테스트 프로세스 종료)
  - `execute.go:344-345` `NewExecuteCmd` MarkFlagRequired panic (1 stmt, 방어적 — flag 이름 부재 시에만 발동)
  - `install.go:168-169` `NewInstallCmd` MarkFlagRequired panic (1 stmt, 방어적)
  - `install.go:129-131` `NewInstallCmd` os.Getwd 실패 분기 (1 stmt, 환경 의존 — macOS에서 cwd 삭제 후에도 getwd 실패 안 함 실측)
  - `install.go:137-139` `NewInstallCmd` filepath.Abs 실패 분기 (1 stmt, 상대경로+getwd 실패 시에만 — 결정론적 유발 불가)
  - `execute.go:116-118` `RunExecute` os.Getwd 실패 분기 (1 stmt, 환경 의존)
  - `execute.go:124-126` `RunExecute` filepath.Abs 실패 분기 (1 stmt, 환경 의존)
  - `execute.go:159-161` `runExecuteWith` os.Getwd 실패 분기 (1 stmt, 환경 의존)
  - `propose.go:130-132` `runPropose` `result.Proposals == nil` 분기 (1 stmt, 구조적 dead — MapPromotions가 non-nil empty slice를 항상 반환하므로 도달 불가)
  - `propose.go:135-137` `runPropose` json.Marshal 실패 분기 (1 stmt, 도달 불가 — GeneratorResult는 항상 marshal 가능)
- **note**: AC-HCC-005 도달을 위해 `propose.go:93-95`(ReadPromotions error) + `125-127`(WriteProposals error)는 결정론적 IO 에러(디렉터리-as-input / regular-file-as-output-parent)로 **도달 완료** — residual 아님.
- **4-phase 완결**: plan + run (M1~M6, AC 10/12 PASS, AC-HCC-002/006 blocker §B.2) + sync (§E.2) + Mx (§E.5)
