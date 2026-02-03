# SPEC-QUALITY-001: TRUST 5 Quality Gates

---
id: SPEC-QUALITY-001
title: TRUST 5 Quality Gates
status: Completed
priority: High
phase: "Phase 2 - Core Domains"
module: internal/core/quality/
files:
  - trust.go
  - validators.go
estimated_loc: ~800
dependencies:
  - SPEC-LSP-001
  - SPEC-GIT-001
created: 2026-02-03
assigned: expert-backend
tags:
  - quality
  - trust5
  - validation
  - lsp-integration
  - regression-detection
---

## HISTORY

| Version | Date | Description |
|---------|------|-------------|
| 1.0.0 | 2026-02-03 | Initial SPEC creation |
| 1.1.0 | 2026-02-03 | Add methodology-specific quality gates (DDD, TDD, Hybrid) |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK (Go Edition)은 Claude Code용 고성능 개발 도구이다. TRUST 5 Quality Gates는 코드 변경에 대한 5가지 품질 원칙(Tested, Readable, Unified, Secured, Trackable)을 자동 검증하는 프레임워크를 제공한다.

### 1.2 Technical Context

- **Language**: Go 1.22+
- **Module**: `github.com/modu-ai/moai-adk-go`
- **Package**: `internal/core/quality/`
- **Dependencies**:
  - `internal/lsp/` - LSP 진단 수집 (type errors, lint errors, security warnings)
  - `internal/core/git/` - 커밋 메시지 형식 검증, 진단 이력 추적
  - `internal/astgrep/` - 구조적 코드 패턴 분석
  - `context` (stdlib) - 취소 및 타임아웃 제어
  - `log/slog` (stdlib) - 구조화된 로깅

### 1.3 Configuration Reference

품질 설정은 `.moai/config/sections/quality.yaml`에서 관리된다:

- `constitution.enforce_quality: true` - 품질 게이트 강제 적용
- `constitution.test_coverage_target: 85` - 최소 테스트 커버리지
- `constitution.lsp_quality_gates` - 페이즈별 LSP 임계값
- `constitution.lsp_integration.trust5_integration` - TRUST 5 원칙별 LSP 매핑
- `constitution.lsp_integration.regression_detection` - 회귀 감지 임계값

### 1.4 Interfaces

```go
// Gate defines the TRUST 5 quality gate interface.
type Gate interface {
    // Validate runs all 5 TRUST principles and returns an aggregated report.
    Validate(ctx context.Context) (*Report, error)

    // ValidatePrinciple runs a single TRUST principle by name.
    ValidatePrinciple(ctx context.Context, principle string) (*PrincipleResult, error)
}

// Report is the aggregated quality validation report.
type Report struct {
    Passed          bool                       `json:"passed"`
    Score           float64                    `json:"score"`
    DevelopmentMode string                     `json:"development_mode"`
    Principles      map[string]PrincipleResult `json:"principles"`
    Timestamp       time.Time                  `json:"timestamp"`
}

// PrincipleResult is the validation result for a single TRUST principle.
type PrincipleResult struct {
    Name   string  `json:"name"`
    Passed bool    `json:"passed"`
    Score  float64 `json:"score"`
    Issues []Issue `json:"issues"`
}

// Issue represents a single quality violation.
type Issue struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Rule     string `json:"rule"`
}
```

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID | Assumption | Confidence | Risk if Wrong |
|----|------------|------------|---------------|
| A-1 | LSP 클라이언트(`internal/lsp/`)가 진단 데이터를 제공할 수 있다 | High | LSP 미구현 시 stub/mock으로 대체 필요 |
| A-2 | Git 매니저(`internal/core/git/`)가 커밋 메시지 접근을 지원한다 | High | Git 미구현 시 Trackable 원칙 검증 불가 |
| A-3 | `context.Context`를 통한 타임아웃이 모든 검증기에서 존중된다 | High | 무한 대기 가능성 |
| A-4 | quality.yaml 설정이 `internal/config/`를 통해 로드된다 | High | 설정 누락 시 기본값 사용 |
| A-5 | 각 원칙 검증기는 독립적으로 실행 가능하다 (goroutine-safe) | High | 경쟁 조건 발생 가능 |

### 2.2 Design Assumptions

- Gate 인터페이스는 DDD 경계를 준수하여 `internal/core/quality/` 패키지 내에 정의된다
- 외부 의존성(LSP, Git, AST-Grep)은 Go 인터페이스로 추상화되어 테스트 시 mock 교체 가능
- Report 구조체는 JSON 직렬화를 지원하여 CI/CD 파이프라인 통합이 가능하다
- 점수(Score)는 0.0~1.0 범위의 float64이며, 각 원칙의 가중 평균으로 계산된다

---

## 3. Requirements (R)

### 3.1 TRUST 5 Principles - Ubiquitous Requirements

시스템은 **항상** 다음 5가지 품질 원칙을 검증할 수 있어야 한다.

#### REQ-T-001: Tested (테스트됨)

시스템은 **항상** 다음 조건을 검증해야 한다:

- 단위 테스트 통과 여부 확인
- LSP type errors == 0 검증
- LSP errors == 0 검증
- 테스트 커버리지가 설정된 임계값(기본 85%) 이상인지 확인

#### REQ-R-001: Readable (읽기 쉬움)

시스템은 **항상** 다음 조건을 검증해야 한다:

- 네이밍 컨벤션 준수 여부 확인
- LSP lint errors == 0 검증

#### REQ-U-001: Understandable (이해 가능함)

시스템은 **항상** 다음 조건을 검증해야 한다:

- 문서화 완성도 확인
- 코드 복잡도가 허용 범위 내인지 확인
- LSP warnings < threshold 검증

#### REQ-S-001: Secured (보안됨)

시스템은 **항상** 다음 조건을 검증해야 한다:

- 보안 스캔 통과 여부 확인
- LSP security warnings == 0 검증

#### REQ-K-001: Trackable (추적 가능함)

시스템은 **항상** 다음 조건을 검증해야 한다:

- 구조화된 로그 사용 여부 확인
- LSP 진단 이력 추적 여부 확인

### 3.2 Gate Interface - Event-Driven Requirements

#### REQ-GATE-001: Full Validation

**WHEN** `Validate(ctx)` 메서드가 호출되면 **THEN** 시스템은 5개 TRUST 원칙을 모두 실행하고, 각 원칙 결과를 집계하여 `Report`를 반환해야 한다.

#### REQ-GATE-002: Single Principle Validation

**WHEN** `ValidatePrinciple(ctx, "tested")` 메서드가 특정 원칙 이름으로 호출되면 **THEN** 시스템은 해당 원칙만 실행하고 `PrincipleResult`를 반환해야 한다.

#### REQ-GATE-003: Invalid Principle Name

**WHEN** `ValidatePrinciple(ctx, "unknown")` 메서드가 유효하지 않은 원칙 이름으로 호출되면 **THEN** 시스템은 명확한 오류 메시지와 함께 error를 반환해야 한다.

#### REQ-GATE-004: Context Cancellation

**WHEN** context가 취소되면 **THEN** 시스템은 진행 중인 검증을 즉시 중단하고, 부분 결과와 함께 context error를 반환해야 한다.

### 3.3 Phase-Specific Quality Gates - State-Driven Requirements

#### REQ-PHASE-001: Plan Phase Baseline

**IF** 현재 워크플로우 페이즈가 `plan`이면 **THEN** 시스템은 LSP 진단 기준선(baseline)을 캡처하고 저장해야 한다.

#### REQ-PHASE-002: Run Phase Zero Tolerance

**IF** 현재 워크플로우 페이즈가 `run`이면 **THEN** 시스템은 다음 조건을 강제해야 한다:

- max_errors: 0
- max_type_errors: 0
- max_lint_errors: 0
- allow_regression: false

#### REQ-PHASE-003: Sync Phase Clean State

**IF** 현재 워크플로우 페이즈가 `sync`이면 **THEN** 시스템은 다음 조건을 강제해야 한다:

- max_errors: 0
- max_warnings: 10
- require_clean_lsp: true

### 3.4 Regression Detection - Combined Requirements

#### REQ-REG-001: Error Regression

**IF** 이전 기준선이 존재하고 **WHEN** 새로운 검증이 실행되면 **THEN** 시스템은 에러 수 증가량이 `error_increase_threshold`(기본 0)를 초과하지 않는지 검증해야 한다.

#### REQ-REG-002: Warning Regression

**IF** 이전 기준선이 존재하고 **WHEN** 새로운 검증이 실행되면 **THEN** 시스템은 경고 수 증가량이 `warning_increase_threshold`(기본 10)를 초과하지 않는지 검증해야 한다.

#### REQ-REG-003: Type Error Regression

**IF** 이전 기준선이 존재하고 **WHEN** 새로운 검증이 실행되면 **THEN** 시스템은 타입 에러 수 증가량이 `type_error_increase_threshold`(기본 0)를 초과하지 않는지 검증해야 한다.

### 3.5 Unwanted Requirements

#### REQ-NEG-001: No Panic

시스템은 검증 실패 시 **panic을 발생시키지 않아야 한다**. 모든 오류는 error 값으로 반환되어야 한다.

#### REQ-NEG-002: No Blocking

시스템은 context 타임아웃 없이 **무한 대기하지 않아야 한다**. 모든 외부 호출(LSP, Git, AST-Grep)은 context 기반 타임아웃을 적용해야 한다.

#### REQ-NEG-003: No Configuration Mutation

시스템은 검증 과정에서 quality.yaml 설정을 **변경하지 않아야 한다**. 설정은 읽기 전용으로 접근되어야 한다.

### 3.6 Optional Requirements

#### REQ-OPT-001: Parallel Validation

**가능하면** 5개 TRUST 원칙 검증을 goroutine을 사용하여 병렬 실행 제공. `errgroup.WithContext`를 활용하여 하나의 원칙이 실패해도 나머지는 계속 실행한다.

#### REQ-OPT-002: Caching

**가능하면** LSP 진단 결과에 TTL 기반 캐싱(기본 5초)을 제공하여 반복 검증 시 성능을 최적화한다.

#### REQ-OPT-003: Report Export

**가능하면** Report를 JSON 파일로 내보내는 기능을 제공하여 CI/CD 파이프라인과 통합한다.

### 3.7 Methodology-Specific Quality Gates

#### REQ-METH-001: DDD Mode Quality Gate (State-Driven)

**IF** `quality.development_mode`가 `ddd`이면 **THEN** 시스템은 다음 규칙을 강제해야 한다:

- Characterization 테스트가 코드 수정 전에 존재해야 한다
- PRESERVE 단계가 성공적으로 완료된 후에만 IMPROVE 단계를 진행해야 한다
- Behavior snapshot이 회귀되지 않아야 한다
- 커버리지는 점진적으로 향상될 수 있다 (커밋당 최소 커버리지 강제 없음)
- 커버리지 면제가 정당화와 함께 허용된다 (`max_exempt_percentage`까지)

#### REQ-METH-002: TDD Mode Quality Gate (State-Driven)

**IF** `quality.development_mode`가 `tdd`이면 **THEN** 시스템은 다음 규칙을 강제해야 한다:

- 테스트가 구현 코드보다 먼저 작성되어야 한다
- RED-GREEN-REFACTOR 사이클이 필수이다:
  - RED: 실패하는 테스트가 존재해야 한다
  - GREEN: 테스트를 통과시키는 최소한의 코드만 작성한다
  - REFACTOR: 테스트가 여전히 통과하는 상태에서 정리한다
- 커밋당 최소 커버리지: `tdd_settings.min_coverage_per_commit` (기본 80%)
- 커버리지 면제가 허용되지 않는다
- Mutation testing (선택): 활성화 시 mutation score가 임계값을 초과해야 한다

#### REQ-METH-003: Hybrid Mode Quality Gate (State-Driven)

**IF** `quality.development_mode`가 `hybrid`이면 **THEN** 시스템은 다음 규칙을 강제해야 한다:

- **새로운(NEW) 코드 경로** (git diff 분석으로 판별): TDD 규칙 적용
  - 테스트 우선, 최소 커버리지 `hybrid_settings.min_coverage_new` (기본 90%)
- **기존(EXISTING) 코드 수정**: DDD 규칙 적용
  - Characterization 테스트, 동작 보존
  - 커버리지 목표: `hybrid_settings.min_coverage_legacy` (기본 85%)
- 시스템은 각 변경을 다음 기준으로 "new" 또는 "legacy"로 분류해야 한다:
  - 새로운 파일 → TDD 규칙
  - 기존 파일 수정 → DDD 규칙
  - 기존 파일 내 새로운 함수 → 해당 함수에 TDD 규칙

#### REQ-METH-004: Methodology Transition Warning (Event-Driven)

**WHEN** `development_mode`가 모드 간 변경되면 **THEN** 시스템은 다음을 수행해야 한다:

- 모드 전환을 경고 로그로 기록한다 (예: ddd → hybrid)
- 새로운 모드에 맞게 품질 기준선을 재계산한다
- 이전에 통과한 코드를 소급하여 실패 처리하지 않는다

#### REQ-METH-005: Methodology-Aware Scoring (Ubiquitous)

시스템은 **항상** 활성 `development_mode`를 품질 Report에 포함해야 한다:

```go
type Report struct {
    Passed          bool                       `json:"passed"`
    Score           float64                    `json:"score"`
    DevelopmentMode string                     `json:"development_mode"`
    Principles      map[string]PrincipleResult `json:"principles"`
    Timestamp       time.Time                  `json:"timestamp"`
}
```

---

## 4. Specifications (S)

### 4.1 Package Structure

```
internal/core/quality/
  trust.go           # Gate 인터페이스 구현, Report 집계, 페이즈별 검증
  validators.go      # 5개 TRUST 원칙별 Validator 구현
  trust_test.go      # Gate 단위 테스트
  validators_test.go # Validator 단위 테스트
```

### 4.2 Type Definitions

```go
package quality

// Principle constants
const (
    PrincipleTested       = "tested"
    PrincipleReadable     = "readable"
    PrincipleUnderstandable = "understandable"
    PrincipleSecured      = "secured"
    PrincipleTrackable    = "trackable"
)

// Severity levels for Issue
const (
    SeverityError   = "error"
    SeverityWarning = "warning"
    SeverityHint    = "hint"
    SeverityInfo    = "info"
)

// ValidPrinciples lists all supported principle names
var ValidPrinciples = []string{
    PrincipleTested,
    PrincipleReadable,
    PrincipleUnderstandable,
    PrincipleSecured,
    PrincipleTrackable,
}

// DevelopmentMode represents the configured development methodology.
type DevelopmentMode string

const (
    ModeDDD    DevelopmentMode = "ddd"
    ModeTDD    DevelopmentMode = "tdd"
    ModeHybrid DevelopmentMode = "hybrid"
)

// ValidDevelopmentModes lists all supported development modes.
var ValidDevelopmentModes = []DevelopmentMode{
    ModeDDD,
    ModeTDD,
    ModeHybrid,
}

// MethodologyValidator determines which quality rules to apply based on development_mode.
type MethodologyValidator struct {
    Mode           DevelopmentMode
    DDDSettings    DDDSettings
    TDDSettings    TDDSettings
    HybridSettings HybridSettings
}

// DDDSettings holds DDD-specific quality gate configuration.
type DDDSettings struct {
    RequireExistingTests    bool `yaml:"require_existing_tests"`
    CharacterizationTests   bool `yaml:"characterization_tests"`
    BehaviorSnapshots       bool `yaml:"behavior_snapshots"`
    MaxTransformationSize   string `yaml:"max_transformation_size"`
}

// TDDSettings holds TDD-specific quality gate configuration.
type TDDSettings struct {
    MinCoveragePerCommit   int  `yaml:"min_coverage_per_commit"`
    RequireTestFirst       bool `yaml:"require_test_first"`
    MutationTestingEnabled bool `yaml:"mutation_testing_enabled"`
    MutationScoreThreshold int  `yaml:"mutation_score_threshold"`
}

// HybridSettings holds hybrid mode quality gate configuration.
type HybridSettings struct {
    MinCoverageNew    int `yaml:"min_coverage_new"`
    MinCoverageLegacy int `yaml:"min_coverage_legacy"`
}
```

### 4.3 Validator Interface

```go
// Validator is the internal interface for individual principle validators.
type Validator interface {
    Name() string
    Validate(ctx context.Context) (*PrincipleResult, error)
}
```

### 4.4 Dependencies (Interface Abstractions)

```go
// LSPClient abstracts LSP diagnostic collection for testability.
type LSPClient interface {
    CollectDiagnostics(ctx context.Context) ([]Diagnostic, error)
}

// GitManager abstracts Git operations for trackable validation.
type GitManager interface {
    LastCommitMessage(ctx context.Context) (string, error)
    DiagnosticHistory(ctx context.Context) ([]DiagnosticSnapshot, error)
}

// ASTAnalyzer abstracts AST pattern matching for code analysis.
type ASTAnalyzer interface {
    Analyze(ctx context.Context, patterns []string) ([]ASTMatch, error)
}
```

### 4.5 Configuration Struct

```go
// QualityConfig holds quality gate configuration loaded from quality.yaml.
type QualityConfig struct {
    DevelopmentMode     DevelopmentMode   `yaml:"development_mode"`
    EnforceQuality      bool              `yaml:"enforce_quality"`
    TestCoverageTarget  int               `yaml:"test_coverage_target"`
    LSPGates            PhaseGates        `yaml:"lsp_quality_gates"`
    RegressionDetection RegressionConfig  `yaml:"regression_detection"`
    DDDSettings         DDDSettings       `yaml:"ddd_settings"`
    TDDSettings         TDDSettings       `yaml:"tdd_settings"`
    HybridSettings      HybridSettings    `yaml:"hybrid_settings"`
    CacheTTL            time.Duration     `yaml:"cache_ttl_seconds"`
    Timeout             time.Duration     `yaml:"timeout_seconds"`
}

// PhaseGates contains phase-specific quality thresholds.
type PhaseGates struct {
    Plan PlanGate `yaml:"plan"`
    Run  RunGate  `yaml:"run"`
    Sync SyncGate `yaml:"sync"`
}

// RegressionConfig contains regression detection thresholds.
type RegressionConfig struct {
    ErrorIncreaseThreshold     int `yaml:"error_increase_threshold"`
    WarningIncreaseThreshold   int `yaml:"warning_increase_threshold"`
    TypeErrorIncreaseThreshold int `yaml:"type_error_increase_threshold"`
}
```

### 4.6 Score Calculation

총점은 5개 원칙의 가중 평균으로 계산된다:

| Principle | Weight | Rationale |
|-----------|--------|-----------|
| Tested | 0.30 | 코드 신뢰성의 핵심 |
| Readable | 0.15 | 유지보수성 영향 |
| Understandable | 0.15 | 팀 협업 효율 |
| Secured | 0.25 | 보안 취약점의 심각성 |
| Trackable | 0.15 | 이력 추적 및 감사 |

Pass 기준: 총점 >= 0.7 이고 모든 Severity "error" 이슈가 0개

### 4.7 Performance Targets

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Full TRUST 5 validation | < 5s | Integration benchmark |
| Single principle validation | < 2s | Unit benchmark |
| Regression comparison | < 100ms | Unit benchmark |
| Report JSON serialization | < 10ms | Unit benchmark |

### 4.8 Traceability

| Requirement | Implementation | Test |
|-------------|---------------|------|
| REQ-T-001 | `validators.go:TestedValidator` | `validators_test.go:TestTestedValidator` |
| REQ-R-001 | `validators.go:ReadableValidator` | `validators_test.go:TestReadableValidator` |
| REQ-U-001 | `validators.go:UnderstandableValidator` | `validators_test.go:TestUnderstandableValidator` |
| REQ-S-001 | `validators.go:SecuredValidator` | `validators_test.go:TestSecuredValidator` |
| REQ-K-001 | `validators.go:TrackableValidator` | `validators_test.go:TestTrackableValidator` |
| REQ-GATE-001 | `trust.go:TrustGate.Validate` | `trust_test.go:TestFullValidation` |
| REQ-GATE-002 | `trust.go:TrustGate.ValidatePrinciple` | `trust_test.go:TestSinglePrinciple` |
| REQ-GATE-003 | `trust.go:TrustGate.ValidatePrinciple` | `trust_test.go:TestInvalidPrinciple` |
| REQ-GATE-004 | `trust.go:TrustGate.Validate` | `trust_test.go:TestContextCancellation` |
| REQ-PHASE-001 | `trust.go:TrustGate.captureBaseline` | `trust_test.go:TestPlanPhaseBaseline` |
| REQ-PHASE-002 | `trust.go:TrustGate.validateRunPhase` | `trust_test.go:TestRunPhaseZeroTolerance` |
| REQ-PHASE-003 | `trust.go:TrustGate.validateSyncPhase` | `trust_test.go:TestSyncPhaseClean` |
| REQ-REG-001 | `trust.go:TrustGate.detectRegression` | `trust_test.go:TestErrorRegression` |
| REQ-REG-002 | `trust.go:TrustGate.detectRegression` | `trust_test.go:TestWarningRegression` |
| REQ-REG-003 | `trust.go:TrustGate.detectRegression` | `trust_test.go:TestTypeErrorRegression` |
| REQ-NEG-001 | 전체 패키지 | `trust_test.go:TestNoPanic` |
| REQ-NEG-002 | 전체 패키지 | `trust_test.go:TestTimeoutEnforcement` |
| REQ-NEG-003 | 전체 패키지 | `trust_test.go:TestConfigImmutability` |
| REQ-METH-001 | `trust.go:TrustGate.validateMethodology` | `trust_test.go:TestDDDModeQualityGate` |
| REQ-METH-002 | `trust.go:TrustGate.validateMethodology` | `trust_test.go:TestTDDModeQualityGate` |
| REQ-METH-003 | `trust.go:TrustGate.validateMethodology` | `trust_test.go:TestHybridModeQualityGate` |
| REQ-METH-004 | `trust.go:TrustGate.handleModeTransition` | `trust_test.go:TestMethodologyTransitionWarning` |
| REQ-METH-005 | `trust.go:TrustGate.buildReport` | `trust_test.go:TestReportIncludesDevelopmentMode` |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 96.8%

### Summary

Quality gates package implemented with TRUST 5 framework validation (Tested, Readable, Understandable, Secured, Trackable). Includes phase-specific quality thresholds for plan/run/sync workflow phases, LSP diagnostic regression detection, development methodology validation (DDD/TDD/Hybrid), and comprehensive report generation with weighted scoring. Supports methodology-aware quality gate enforcement with configurable thresholds.

### Files Created

- `internal/core/quality/trust.go`
- `internal/core/quality/trust_test.go`
- `internal/core/quality/validators.go`
- `internal/core/quality/validators_test.go`
