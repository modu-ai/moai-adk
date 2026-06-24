---
id: SPEC-HARNESS-EXECUTE-E2E-001
title: "Harness execute regression-gate 측정 root 결함 수정 + e2e 재현 테스트 — 인수 기준"
version: "0.2.0"
status: draft
created: 2026-06-15
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/harness, internal/harness"
lifecycle: spec-anchored
tags: "harness, telemetry, e2e, test, tdd, bugfix, regression-gate"
tier: M
---

# SPEC-HARNESS-EXECUTE-E2E-001 — 인수 기준 (acceptance.md)

> 본 문서가 AC의 SSOT다. 모든 AC는 anchored / testable이며 command-검증(본문-read 의존
> 최소화). `<Name>` = 신규 재현 테스트 함수명
> (예: `TestRunExecute_RegressionGateMeasuresProjectRoot_WritesKeptTelemetry`).
> `<baseline>` = run-phase M1 착수 시점에 `git rev-parse HEAD`로 캡처해 progress.md에
> 기록한 SHA. 병렬 세션이 공유 트리 main을 이동시키므로 고정 SHA(예: 과거 `bf01fed74`)가
> 아닌 **run-time HEAD**를 baseline으로 사용한다 (D6 — race-robust).

## §A — Given-When-Then 시나리오

### GWT-1 (RED) — 미수정 코드: 측정 root 결함으로 fail-close, telemetry 0건

```
Given  t.TempDir() 프로젝트 root가 real 최소 Go 모듈로 구성되어 있고
       (go.mod + project root에 통과하는 trivial 테스트 1건 + 비-FROZEN
        docs/e2e-sample.md 타겟 + .moai/harness/proposals/<id>.json proposal),
  AND  production 코드가 아직 미수정(measurementRoot = snapshot base)이다
When   재현 테스트가 RunExecute(ExecuteOptions{ID:<id>, ProjectRoot:<tempdir>})를
       직접 호출한다
Then   반환 error != nil 이고 regression-gate measurement(fail-closed) 부류 에러이며,
  AND  <tempdir>/.moai/harness/usage-log.jsonl 에 apply_outcome 라인이 0건이다
  AND  이로써 출시된 verb가 genuine proposal에서 항상 fail-close하는 결함이 재현된다
```

### GWT-2 (GREEN) — 수정 후: 올바른 project root 측정 → verdict="kept" telemetry 1건

```
Given  동일한 t.TempDir() real 최소 Go 모듈 fixture,
  AND  production fix가 적용되어 RunExecute가 .WithProjectRoot(root)를 배선했다
When   재현 테스트가 RunExecute(ExecuteOptions{ID:<id>, ProjectRoot:<tempdir>})를
       직접 호출한다
Then   production glue 전체가 구동된다:
       RunExecute → buildExecutePipelineConfig(AutoApply:true) →
       safety.NewPipeline.Evaluate(L1..L5) → DecisionApproved → Apply →
       real measurer가 project root(=fixture 모듈)에서 go test ./... 실행(통과, Δ=0) →
       Observer.RecordOutcome("kept") → usage-log append
  AND  반환 error == nil 이고,
  AND  <tempdir>/.moai/harness/usage-log.jsonl 에
       event_type == "apply_outcome" AND outcome_verdict == "kept" 라인이 정확히 1건이다
```

### GWT-3 — fix 단위 동작: set/unset 경로

```
Given  NewApplierWithRegressionGate(...)로 만든 gate-active Applier
When   .WithProjectRoot("/x")를 호출하면
Then   gate가 측정 root로 "/x"를 measurer에 전달한다 (stub measurer로 관측)
 ; And When .WithProjectRoot를 호출하지 않으면
   Then 기존 measurementRoot(snapshotDir) fallback 경로를 사용한다 (회귀 0 보존)
```

### GWT-4 — 격리: 실제 레포 harness runtime 파일 미오염

```
Given  재현 테스트가 t.TempDir() + ExecuteOptions.ProjectRoot override를 사용한다
When   GWT-1/GWT-2 테스트가 실행된다
Then   실제 레포의 .moai/harness/{usage-log.jsonl,measurements-baseline.yaml,
       regression-coverage.out} 은 테스트에 의해 write/touch되지 않는다
       (git status --porcelain .moai/harness/ 에 변경 없음)
```

### GWT-5 — production 동결: measurer 등 5 파일 불변

```
Given  이 SPEC의 run-phase가 완료되었다
When   git diff <baseline> HEAD 를 §B.2 STILL-FROZEN 5 파일에 대해 측정한다
Then   diff 라인 합계 == 0 (pipeline.go / outcome.go / observer.go /
       regression_gate.go / harness.yaml 불변; measurer 로직 무변경)
```

## §B — Edge Cases

| EC | 상황 | 기대 |
|----|------|------|
| EC-1 | fixture가 go.mod 없는 빈 디렉터리 | `go test ./...` build-fail → GREEN 불가. 테스트는 반드시 real 최소 Go 모듈(go.mod + trivial 통과 테스트)을 보장 (REQ-E2E-007) |
| EC-2 | fixture 타겟이 frontmatter 없음 | `EnrichDescription` 수정 실패 가능 → frontmatter 포함 타겟 보장 |
| EC-3 | nil sessions가 L2 Canary reject 유발 | 발생 금지 — REQ-E2E-011. gate-active production 경로에서도 nil-safe canary가 reject하지 않아 DecisionApproved 도달이 genuine |
| EC-4 | proposal이 FROZEN 경로 타겟 | 범위 밖 — 비-FROZEN 타겟만 사용하여 L1 통과 |
| EC-5 | usage-log에 apply_outcome 외 이벤트 공존 | 어서션은 `event_type == apply_outcome` 라인만 필터하여 카운트 |
| EC-6 | 기존 gate-active stub-measurer 테스트가 fix로 깨짐 | 발생 금지 — stub measurer는 `projectRoot` 인자 무시. Approach A unset fallback이 기존 동작 보존 (§D.1 blast radius) |

## §C — Quality Gate 기준

- **RED 입증**: 미수정 코드(또는 fix stash)에서 재현 테스트 실행 → fail-close 에러 +
  telemetry 0건 (run-phase 보고에 출력 캡처 포함)
- **GREEN 통과**: `go test -run <Name> ./internal/cli/harness/` exit 0
- **패키지 회귀 0**: `go test ./internal/cli/harness/ ./internal/harness/` exit 0
- **race clean**: `go test -race -run <Name> ./internal/cli/harness/` exit 0
- **cross-platform build**: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- **vet clean**: `go vet ./internal/cli/harness/ ./internal/harness/` exit 0
- **non-vacuous**: genuine RED→GREEN 전이로 입증 (scratch mis-assertion 불필요)

## §D — AC Matrix (anchored / testable)

### §D.1 핵심 AC

| AC | 요구사항 | 검증 커맨드 | PASS 조건 |
|----|----------|-------------|-----------|
| **AC-E2E-001** | 재현 테스트가 `RunExecute`를 직접 호출 (no `&Applier{}`/`applier.Apply(`/`runExecuteWith(`+stub/stub measurer) | `grep -c 'RunExecute(' internal/cli/harness/execute_e2e_test.go` (≥1) AND `grep -c '&harness.Applier{\|&Applier{\|stubMeasurer\|runExecuteWith(' internal/cli/harness/execute_e2e_test.go` (==0) | 첫 grep ≥1 AND 둘째 grep ==0 |
| **AC-E2E-002** | RED: 미수정 코드에서 fail-close + apply_outcome 0건 | RED 커밋의 `go test -run <Name>` 출력 (fix 미적용 시 FAIL/fail-close 에러) + telemetry 0건 어서션 | run-phase 보고에 RED 출력 캡처; fail-close 에러 메시지에 "measurement"/"fail-closed" 포함 |
| **AC-E2E-003** | GREEN: 수정 후 `verdict="kept"` apply_outcome 라인 정확히 1건 + nil error | `go test -run <Name> ./internal/cli/harness/` | PASS (count==1 AND verdict=="kept" AND nil-error 어서션) |
| **AC-E2E-004** | fix 단위: `WithProjectRoot` set → measurer가 그 root 수신; unset → fallback | `go test -run <UnitName> ./internal/harness/` | PASS (set/unset 두 sub-test) |
| **AC-E2E-005** | fixture가 real 최소 Go 모듈 (go.mod + trivial 통과 테스트 + 비-FROZEN 타겟) | `grep -c 'go.mod' internal/cli/harness/execute_e2e_test.go` (≥1) + GREEN 통과 | grep ≥1 AND AC-E2E-003 PASS (모듈 없으면 GREEN 불가) |
| **AC-E2E-006** | STILL-FROZEN 5 파일 diff 0 | `git diff --stat <baseline> HEAD -- internal/harness/safety/pipeline.go internal/harness/outcome.go internal/harness/observer.go internal/harness/regression_gate.go .moai/config/sections/harness.yaml` | 출력 없음 (0 diff 라인) |
| **AC-E2E-007** | 실제 레포 `.moai/harness/*` 미변경 (테스트는 `t.TempDir()` ProjectRoot만 사용) | `grep -c 't.TempDir()' internal/cli/harness/execute_e2e_test.go` (≥1) AND `git status --porcelain .moai/harness/` | grep ≥1 AND porcelain empty |
| **AC-E2E-008** | C-HRA-008: 새 테스트에 `AskUserQuestion(` 0건 | `grep -c 'AskUserQuestion(' internal/cli/harness/execute_e2e_test.go` AND `go test -run TestPropose_NoAskUserQuestion ./internal/cli/harness/` | grep count 0 AND 가드 PASS |
| **AC-E2E-009** | DoD: `internal/cli/harness` + `internal/harness` 패키지 커버리지 비회귀 | `go test -cover ./internal/cli/harness/ ./internal/harness/` (baseline 대비) | 두 패키지 커버리지 동등 이상. **D3 90% 절대 바 달성은 out of scope** |

### §D.2 Severity 분류

- **MUST-PASS (gating)**: AC-E2E-001, 002, 003, 004, 005, 006, 007, 008 — 하나라도
  FAIL 시 SPEC 미완. (RED 입증 002 + GREEN 003 + fix 단위 004는 bug-fix 핵심.)
- **SHOULD-PASS (non-gating)**: AC-E2E-009 (커버리지 비회귀; 회귀 시 보고하되 D3 90%
  미달은 본 SPEC scope 밖이므로 비회귀만 gating).

### §D.3 Traceability (AC ↔ REQ)

| AC | REQ |
|----|-----|
| AC-E2E-001 | REQ-E2E-006 |
| AC-E2E-002 | REQ-E2E-004 |
| AC-E2E-003 | REQ-E2E-005, REQ-E2E-011 |
| AC-E2E-004 | REQ-E2E-001, REQ-E2E-002, REQ-E2E-003 |
| AC-E2E-005 | REQ-E2E-007 |
| AC-E2E-006 | REQ-E2E-009 |
| AC-E2E-007 | REQ-E2E-008 |
| AC-E2E-008 | REQ-E2E-010 |
| AC-E2E-009 | (DoD — spec.md §F) |

### §D.4 Indirect Verification (FROZEN / 격리 / measurer 불변)

- **AC-E2E-006 (FROZEN diff 0)**: measurer 동결은 `git diff`로 직접 측정. 추가로
  `go test ./internal/harness/safety/`가 기존과 동일 PASS함을 확인하여 pipeline
  동작 불변 간접 확인.
- **AC-E2E-007 (격리)**: `git status --porcelain .moai/harness/` empty +
  `grep -c 't.TempDir()' execute_e2e_test.go` (≥1)로 테스트가 tempdir ProjectRoot만
  사용함을 command 검증한다 (본문 read 의존 제거, D5).
- **회귀 0 (blast radius)**: 기존 gate-active stub-measurer 테스트
  (`applier_test.go`/`outcome_test.go`/`execute_integration_test.go`/
  `regression_gate_test.go`)가 fix 후에도 전부 PASS함을 `go test ./internal/harness/`로
  확인 — stub measurer는 root를 무시하므로 Approach A가 무영향임을 실증.

## §E — Closure Gate (Definition of Done 재확인)

run-phase가 다음을 모두 충족할 때만 sync 진입:

- [ ] AC-E2E-001~008 전부 PASS (MUST-PASS)
- [ ] AC-E2E-009 비회귀 PASS (SHOULD-PASS)
- [ ] RED→GREEN 재현 전이 입증 (M1 RED 커밋 출력 + M2 GREEN 출력 보고)
- [ ] `go test ./internal/cli/harness/ ./internal/harness/` exit 0 (회귀 0)
- [ ] `go test -race -run <Name> ./internal/cli/harness/` exit 0
- [ ] `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] `go vet ./internal/cli/harness/ ./internal/harness/` exit 0
- [ ] STILL-FROZEN 5 파일 + 2 기존 테스트 파일 diff 0
- [ ] 실제 레포 `.moai/harness/*` 미변경

## §F — Forward-Looking Checks (회귀 방지)

- 재현 테스트는 향후 `RunExecute` glue / 측정 root 배선이 리팩터링되어도 telemetry
  happy-path를 보호하는 regression sentinel 역할.
- `measurementRoot`/`WithProjectRoot` 경로가 재차 snapshot base로 회귀하면 이
  테스트가 즉시 fail-close로 실패하여 알린다 (결함 재발 방지).
- `outcome.go` 스키마(`OutcomeVerdict`)가 변경되면 GREEN 어서션이 실패하여 알린다.

### §F.1 Out of Scope (acceptance-level)

- D3 패키지 커버리지 90% 절대 바 달성 AC 제외 — AC-E2E-009는 **비회귀**만 gating
  (spec.md §D.1).
- rolled-back / regression-blocked verdict AC 변형 제외 (kept happy-path + RED만).
- measurer 로직 변경 검증 제외 — measurer는 STILL FROZEN, root만 바로잡음.
- harness write surface 확장 / propose genuine 생성기화 검증 제외.
