# SPEC-QUALITY-001: Acceptance Criteria

---
spec_id: SPEC-QUALITY-001
title: TRUST 5 Quality Gates - Acceptance Criteria
tags:
  - quality
  - trust5
  - validation
  - acceptance
---

## Quality Gate Criteria

### Definition of Done

- [ ] Gate 인터페이스가 정의되고 TrustGate가 이를 구현한다
- [ ] 5개 TRUST 원칙 Validator가 모두 구현된다
- [ ] 페이즈별 품질 게이트(plan/run/sync)가 동작한다
- [ ] 회귀 감지 로직이 구현된다
- [ ] 단위 테스트 커버리지 >= 85%
- [ ] 모든 테스트가 `go test -race`로 통과한다
- [ ] golangci-lint 경고 0개
- [ ] godoc 주석이 모든 exported 타입/함수에 작성된다
- [ ] Report JSON 직렬화가 정상 동작한다
- [ ] DDD/TDD/Hybrid 3가지 모드별 품질 게이트가 동작한다
- [ ] Hybrid 모드에서 git diff 기반 변경 분류(new/legacy)가 정확하다
- [ ] 모드 전환 시 경고 로그가 기록되고 기준선이 재계산된다
- [ ] 유효하지 않은 development_mode가 명확한 에러로 거부된다
- [ ] Report에 development_mode 필드가 항상 포함된다

### Verification Methods

| Method | Tool | Target |
|--------|------|--------|
| Unit Test | `go test -race -cover ./internal/core/quality/...` | 85%+ coverage |
| Lint | `golangci-lint run ./internal/core/quality/...` | 0 warnings |
| Benchmark | `go test -bench=. ./internal/core/quality/...` | Full validation < 5s |
| Race Detection | `go test -race ./internal/core/quality/...` | 0 data races |

---

## Test Scenarios

### TS-001: Full TRUST 5 Validation - All Pass

**Given** 모든 LSP 진단이 깨끗하고 (errors=0, type_errors=0, lint_errors=0, security_warnings=0)
**And** 테스트 커버리지가 90%이고
**And** 커밋 메시지가 Conventional Commits 형식을 따르고
**And** 구조화된 로그가 사용되고 있을 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** `Report.Passed`가 `true`이고
**And** `Report.Score`가 0.7 이상이고
**And** 모든 `PrincipleResult.Passed`가 `true`이고
**And** 모든 `PrincipleResult.Issues`가 빈 슬라이스이다

---

### TS-002: Full TRUST 5 Validation - Tested Fails

**Given** LSP type errors가 3개 존재하고
**And** 테스트 커버리지가 60% (임계값 85% 미만)일 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** `Report.Passed`가 `false`이고
**And** `Report.Principles["tested"].Passed`가 `false`이고
**And** `Report.Principles["tested"].Issues`에 type error 관련 Issue 3개와 커버리지 미달 Issue 1개가 포함되고
**And** 나머지 원칙(readable, understandable, secured, trackable)은 정상적으로 검증된다

---

### TS-003: Full TRUST 5 Validation - Multiple Principles Fail

**Given** LSP type errors가 2개, lint errors가 5개, security warnings가 1개 존재할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** `Report.Passed`가 `false`이고
**And** `Report.Principles["tested"].Passed`가 `false`이고
**And** `Report.Principles["readable"].Passed`가 `false`이고
**And** `Report.Principles["secured"].Passed`가 `false`이고
**And** `Report.Score`가 0.7 미만이다

---

### TS-004: Single Principle Validation - Valid Name

**Given** TRUST 게이트가 초기화되어 있을 때
**When** `Gate.ValidatePrinciple(ctx, "tested")`를 호출하면
**Then** `PrincipleResult`가 반환되고
**And** `PrincipleResult.Name`이 `"tested"`이고
**And** 다른 원칙은 실행되지 않는다

---

### TS-005: Single Principle Validation - Invalid Name

**Given** TRUST 게이트가 초기화되어 있을 때
**When** `Gate.ValidatePrinciple(ctx, "nonexistent")`를 호출하면
**Then** error가 반환되고
**And** error 메시지에 유효한 원칙 이름 목록이 포함된다

---

### TS-006: Context Cancellation During Validation

**Given** TRUST 게이트가 초기화되어 있고
**And** LSPClient의 진단 수집이 5초 이상 소요될 때
**When** 1초 타임아웃의 context로 `Gate.Validate(ctx)`를 호출하면
**Then** `context.DeadlineExceeded` error가 반환되고
**And** 부분 결과가 포함된 Report가 함께 반환된다

---

### TS-007: Context Cancellation - Manual Cancel

**Given** TRUST 게이트가 초기화되어 있고
**And** context가 검증 도중 취소될 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** `context.Canceled` error가 반환되고
**And** 이미 완료된 원칙의 결과가 Report에 포함된다

---

## Phase-Specific Validation Scenarios

### TS-PHASE-001: Plan Phase - Baseline Capture

**Given** 현재 워크플로우 페이즈가 `plan`이고
**And** LSP 진단이 errors=5, warnings=12, type_errors=2를 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** baseline이 캡처되어 저장되고 (errors=5, warnings=12, type_errors=2)
**And** Report가 현재 진단 상태를 포함하여 반환된다

---

### TS-PHASE-002: Run Phase - Zero Tolerance Pass

**Given** 현재 워크플로우 페이즈가 `run`이고
**And** LSP 진단이 errors=0, type_errors=0, lint_errors=0을 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 통과하고
**And** `Report.Passed`가 `true`이다

---

### TS-PHASE-003: Run Phase - Zero Tolerance Fail (Errors)

**Given** 현재 워크플로우 페이즈가 `run`이고
**And** LSP 진단이 errors=1을 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 실패하고
**And** `Report.Passed`가 `false`이고
**And** Issue에 "run phase requires zero errors" 메시지가 포함된다

---

### TS-PHASE-004: Run Phase - Zero Tolerance Fail (Type Errors)

**Given** 현재 워크플로우 페이즈가 `run`이고
**And** LSP 진단이 type_errors=2를 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 실패하고
**And** Issue에 "run phase requires zero type errors" 메시지가 포함된다

---

### TS-PHASE-005: Run Phase - Zero Tolerance Fail (Lint Errors)

**Given** 현재 워크플로우 페이즈가 `run`이고
**And** LSP 진단이 lint_errors=3을 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 실패하고
**And** Issue에 "run phase requires zero lint errors" 메시지가 포함된다

---

### TS-PHASE-006: Sync Phase - Clean State Pass

**Given** 현재 워크플로우 페이즈가 `sync`이고
**And** LSP 진단이 errors=0, warnings=8을 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 통과하고 (warnings 8 <= threshold 10)
**And** `Report.Passed`가 `true`이다

---

### TS-PHASE-007: Sync Phase - Clean State Fail (Warnings Exceeded)

**Given** 현재 워크플로우 페이즈가 `sync`이고
**And** LSP 진단이 errors=0, warnings=15를 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 실패하고 (warnings 15 > threshold 10)
**And** Issue에 "sync phase allows maximum 10 warnings" 메시지가 포함된다

---

### TS-PHASE-008: Sync Phase - Clean State Fail (Errors)

**Given** 현재 워크플로우 페이즈가 `sync`이고
**And** LSP 진단이 errors=1을 보고할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 페이즈별 검증이 실패하고
**And** Issue에 "sync phase requires zero errors" 메시지가 포함된다

---

## Regression Detection Scenarios

### TS-REG-001: No Regression - Improvement

**Given** baseline이 errors=5, warnings=20, type_errors=3이고
**And** 현재 결과가 errors=3, warnings=15, type_errors=1일 때
**When** 회귀 감지를 실행하면
**Then** 회귀가 감지되지 않고
**And** Report에 회귀 관련 Issue가 없다

---

### TS-REG-002: Error Regression Detected

**Given** baseline이 errors=2이고
**And** error_increase_threshold가 0이고
**And** 현재 결과가 errors=3일 때
**When** 회귀 감지를 실행하면
**Then** 에러 회귀가 감지되고
**And** Issue에 "error count increased from 2 to 3 (threshold: 0)" 메시지가 포함되고
**And** Issue의 Severity가 "error"이다

---

### TS-REG-003: Warning Regression - Within Threshold

**Given** baseline이 warnings=10이고
**And** warning_increase_threshold가 10이고
**And** 현재 결과가 warnings=18일 때
**When** 회귀 감지를 실행하면
**Then** 경고 회귀가 감지되지 않는다 (증가량 8 <= threshold 10)

---

### TS-REG-004: Warning Regression - Exceeds Threshold

**Given** baseline이 warnings=10이고
**And** warning_increase_threshold가 10이고
**And** 현재 결과가 warnings=25일 때
**When** 회귀 감지를 실행하면
**Then** 경고 회귀가 감지되고
**And** Issue에 "warning count increased from 10 to 25 (threshold: 10)" 메시지가 포함되고
**And** Issue의 Severity가 "warning"이다

---

### TS-REG-005: Type Error Regression Detected

**Given** baseline이 type_errors=0이고
**And** type_error_increase_threshold가 0이고
**And** 현재 결과가 type_errors=1일 때
**When** 회귀 감지를 실행하면
**Then** 타입 에러 회귀가 감지되고
**And** Issue에 "type error count increased from 0 to 1 (threshold: 0)" 메시지가 포함된다

---

### TS-REG-006: No Baseline Available

**Given** baseline이 존재하지 않을 때
**When** 회귀 감지를 실행하면
**Then** 회귀 감지가 스킵되고
**And** slog에 "no baseline available, skipping regression detection" 로그가 기록되고
**And** Report에 회귀 관련 Issue가 없다

---

## Per-Principle Validation Scenarios

### TS-TESTED-001: Tested - All Checks Pass

**Given** LSP type errors == 0이고
**And** LSP errors == 0이고
**And** 테스트 커버리지가 90%일 때
**When** TestedValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `true`이고
**And** `PrincipleResult.Score`가 1.0이고
**And** `PrincipleResult.Issues`가 빈 슬라이스이다

---

### TS-TESTED-002: Tested - Coverage Below Threshold

**Given** LSP type errors == 0이고
**And** LSP errors == 0이고
**And** 테스트 커버리지가 70% (임계값 85% 미만)일 때
**When** TestedValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue에 "test coverage 70% is below target 85%" 메시지가 포함된다

---

### TS-TESTED-003: Tested - Type Errors Present

**Given** LSP type errors가 5개 존재할 때
**When** TestedValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue 5개가 각각 파일, 줄 번호, 메시지를 포함한다

---

### TS-READABLE-001: Readable - Clean

**Given** LSP lint errors == 0이고
**And** 네이밍 컨벤션 위반이 없을 때
**When** ReadableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `true`이고
**And** `PrincipleResult.Score`가 1.0이다

---

### TS-READABLE-002: Readable - Lint Errors

**Given** LSP lint errors가 4개 존재할 때
**When** ReadableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue 4개가 각각 Rule 필드를 포함한다

---

### TS-UNDERSTANDABLE-001: Understandable - Acceptable

**Given** LSP warnings가 5개이고 (threshold 10 이내)
**And** 코드 복잡도가 허용 범위 내이고
**And** 문서화가 완성되어 있을 때
**When** UnderstandableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `true`이다

---

### TS-UNDERSTANDABLE-002: Understandable - Warnings Exceeded

**Given** LSP warnings가 15개이고 (threshold 10 초과)일 때
**When** UnderstandableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue에 경고 수 초과 메시지가 포함된다

---

### TS-SECURED-001: Secured - Clean

**Given** LSP security warnings == 0이고
**And** 보안 스캔이 통과할 때
**When** SecuredValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `true`이다

---

### TS-SECURED-002: Secured - Vulnerabilities Found

**Given** LSP security warnings가 2개 존재할 때
**When** SecuredValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue 2개가 각각 Severity "error"와 보안 관련 Rule을 포함한다

---

### TS-TRACKABLE-001: Trackable - Clean

**Given** 마지막 커밋 메시지가 "feat(quality): add TRUST 5 validation" 형식이고
**And** 구조화된 로그가 사용되고
**And** LSP 진단 이력이 추적되고 있을 때
**When** TrackableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `true`이다

---

### TS-TRACKABLE-002: Trackable - Invalid Commit Message

**Given** 마지막 커밋 메시지가 "fixed stuff"이고 (Conventional Commits 미준수)일 때
**When** TrackableValidator를 실행하면
**Then** `PrincipleResult.Passed`가 `false`이고
**And** Issue에 "commit message does not follow Conventional Commits format" 메시지가 포함된다

---

## Negative Test Scenarios

### TS-NEG-001: No Panic on LSP Failure

**Given** LSPClient가 error를 반환할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** panic이 발생하지 않고
**And** error가 정상적으로 반환되고
**And** 부분 결과가 Report에 포함된다

---

### TS-NEG-002: No Blocking Without Context

**Given** LSPClient가 응답하지 않고
**And** context 타임아웃이 3초로 설정되어 있을 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 3초 이내에 error와 함께 반환된다

---

### TS-NEG-003: Configuration Immutability

**Given** quality.yaml에서 `test_coverage_target: 85`가 로드되었을 때
**When** `Gate.Validate(ctx)`를 실행한 후
**Then** `test_coverage_target` 값이 여전히 85이다 (변경되지 않음)

---

## Report Structure Scenarios

### TS-REPORT-001: JSON Serialization

**Given** 검증이 완료된 Report가 존재할 때
**When** `json.Marshal(report)`를 호출하면
**Then** 유효한 JSON이 생성되고
**And** `json.Valid(data)`가 `true`를 반환하고
**And** 역직렬화(`json.Unmarshal`)한 결과가 원본과 동일하다

---

### TS-REPORT-002: Report Timestamp

**Given** 검증이 실행될 때
**When** Report가 생성되면
**Then** `Report.Timestamp`가 현재 시각(UTC)으로 설정되고
**And** 이전 검증의 Timestamp보다 이후이다

---

### TS-REPORT-003: Score Calculation

**Given** 5개 원칙의 점수가 다음과 같을 때:
  - Tested: 1.0 (weight 0.30)
  - Readable: 0.8 (weight 0.15)
  - Understandable: 0.6 (weight 0.15)
  - Secured: 1.0 (weight 0.25)
  - Trackable: 0.5 (weight 0.15)
**When** 총점이 계산되면
**Then** `Report.Score`가 `1.0*0.30 + 0.8*0.15 + 0.6*0.15 + 1.0*0.25 + 0.5*0.15 = 0.835`이고
**And** `Report.Passed`가 `true`이다 (0.835 >= 0.7)

---

### TS-REPORT-004: Score Below Threshold

**Given** 5개 원칙의 점수가 모두 0.5일 때
**When** 총점이 계산되면
**Then** `Report.Score`가 `0.5`이고
**And** `Report.Passed`가 `false`이다 (0.5 < 0.7)

---

## Concurrency Scenarios

### TS-CONC-001: Parallel Principle Validation

**Given** 5개 Validator가 각각 500ms 소요될 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** 전체 소요 시간이 1500ms 이내이다 (병렬 실행 효과)
**And** 모든 5개 원칙 결과가 Report에 포함된다

---

### TS-CONC-002: Race Condition Safety

**Given** 5개 Validator가 동시에 결과를 기록할 때
**When** `go test -race ./internal/core/quality/...`를 실행하면
**Then** data race가 감지되지 않는다

---

### TS-CONC-003: Partial Result on Goroutine Error

**Given** 5개 Validator 중 1개가 500ms 후 error를 반환하고
**And** 나머지 4개가 100ms 내에 성공할 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** error가 반환되고
**And** 성공한 4개 원칙의 결과가 Report에 포함된다

---

## Methodology-Specific Quality Gate Scenarios

### TS-METH-001: DDD Mode Validates Characterization Tests

**Given** `development_mode`가 `"ddd"`이고
**And** 기존 소스 파일이 수정되었을 때
**When** 품질 검증이 실행되면
**Then** 수정된 파일에 대한 characterization 테스트가 존재해야 하고
**And** behavior snapshot이 회귀를 보이지 않아야 한다

---

### TS-METH-002: DDD Mode Allows Gradual Coverage Improvement

**Given** `development_mode`가 `"ddd"`이고
**And** 테스트 커버리지가 75% (전체 목표 85% 미만)이지만 이전보다 향상되었을 때
**When** 품질 검증이 실행되면
**Then** 커버리지 미달로 인한 즉시 실패가 발생하지 않고
**And** 커버리지 향상이 Report에 기록된다

---

### TS-METH-003: DDD Mode Requires PRESERVE Before IMPROVE

**Given** `development_mode`가 `"ddd"`이고
**And** PRESERVE 단계가 완료되지 않았을 때
**When** IMPROVE 단계의 품질 검증이 실행되면
**Then** 검증이 실패하고
**And** Issue에 "PRESERVE step must complete before IMPROVE" 메시지가 포함된다

---

### TS-METH-004: TDD Mode Requires Test-First Development

**Given** `development_mode`가 `"tdd"`이고
**And** 새로운 구현 코드가 추가되었을 때
**When** 품질 검증이 실행되면
**Then** 대응하는 테스트 파일이 구현 코드보다 이른 타임스탬프를 가져야 하고
**And** 커밋당 커버리지가 `min_coverage_per_commit` 임계값을 충족해야 한다

---

### TS-METH-005: TDD Mode Rejects Coverage Exemptions

**Given** `development_mode`가 `"tdd"`이고
**And** 커버리지 면제가 요청되었을 때
**When** 품질 검증이 실행되면
**Then** 커버리지 면제가 거부되고
**And** Issue에 "coverage exemptions not allowed in TDD mode" 메시지가 포함된다

---

### TS-METH-006: TDD Mode Minimum Coverage Per Commit

**Given** `development_mode`가 `"tdd"`이고
**And** `min_coverage_per_commit`가 80%이고
**And** 커밋의 커버리지가 65%일 때
**When** 품질 검증이 실행되면
**Then** 검증이 실패하고
**And** Issue에 "commit coverage 65% is below TDD minimum 80%" 메시지가 포함된다

---

### TS-METH-007: Hybrid Mode Classifies Changes Correctly

**Given** `development_mode`가 `"hybrid"`이고
**And** 새로운 파일과 수정된 기존 파일이 모두 존재할 때
**When** 품질 검증이 실행되면
**Then** 새로운 파일은 TDD 규칙으로 검증되고
**And** 수정된 파일은 DDD 규칙으로 검증되고
**And** 각 카테고리에 대해 별도의 커버리지 목표가 적용된다

---

### TS-METH-008: Hybrid Mode New Functions in Existing Files

**Given** `development_mode`가 `"hybrid"`이고
**And** 기존 파일에 새로운 함수가 추가되었을 때
**When** 품질 검증이 실행되면
**Then** 새로운 함수는 TDD 규칙 (min_coverage_new=90%)으로 검증되고
**And** 기존 함수 수정은 DDD 규칙 (min_coverage_legacy=85%)으로 검증된다

---

### TS-METH-009: Hybrid Mode Separate Coverage Targets

**Given** `development_mode`가 `"hybrid"`이고
**And** `min_coverage_new`가 90%이고
**And** `min_coverage_legacy`가 85%이고
**And** 새로운 코드 커버리지가 88% (미달)이고
**And** 기존 코드 커버리지가 87% (충족)일 때
**When** 품질 검증이 실행되면
**Then** 새로운 코드에 대한 검증이 실패하고
**And** 기존 코드에 대한 검증이 통과하고
**And** Report에 각 카테고리의 커버리지가 별도로 표시된다

---

### TS-METH-010: Methodology Transition Warning

**Given** `development_mode`가 `"ddd"`에서 `"hybrid"`로 변경되었을 때
**When** 설정 변경 후 품질 검증이 실행되면
**Then** slog에 "development mode changed from ddd to hybrid" 경고 로그가 기록되고
**And** 새로운 모드에 맞는 품질 기준선이 재계산되고
**And** 이전에 통과한 코드가 소급하여 실패 처리되지 않는다

---

### TS-METH-011: Invalid Development Mode Rejected

**Given** quality.yaml의 `development_mode`가 `"invalid"`로 설정되어 있을 때
**When** 설정 검증이 실행되면
**Then** 검증 에러가 반환되고
**And** 에러 메시지에 유효한 옵션(ddd, tdd, hybrid)이 나열된다

---

### TS-METH-012: Report Includes Development Mode

**Given** `development_mode`가 `"tdd"`로 설정되어 있을 때
**When** `Gate.Validate(ctx)`를 호출하면
**Then** `Report.DevelopmentMode`가 `"tdd"`이고
**And** JSON 직렬화 시 `"development_mode": "tdd"` 필드가 포함된다
