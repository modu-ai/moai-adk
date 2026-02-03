# SPEC-QUALITY-001: Implementation Plan

---
spec_id: SPEC-QUALITY-001
title: TRUST 5 Quality Gates - Implementation Plan
tags:
  - quality
  - trust5
  - validation
---

## Implementation Strategy

### Approach

DDD의 ANALYZE-PRESERVE-IMPROVE 사이클을 따르며, 인터페이스 우선(Interface-First) 접근 방식으로 구현한다. 먼저 Gate 인터페이스와 핵심 타입을 정의하고, 개별 원칙 Validator를 구현한 후, 집계 및 페이즈별 검증 로직을 완성한다.

### Architecture Design

```
                    +-------------------+
                    |    Gate (interface)|
                    |  Validate()       |
                    |  ValidatePrinciple|
                    +---------+---------+
                              |
                    +---------+---------+
                    |    TrustGate      |
                    |  (concrete impl)  |
                    |                   |
                    |  - validators     |
                    |  - config         |
                    |  - phase          |
                    |  - baseline       |
                    |  - methodology    |
                    +---------+---------+
                              |
                    +---------+---------+
                    | MethodologyValid. |
                    |  - Mode (ddd/tdd/ |
                    |    hybrid)        |
                    |  - DDDSettings    |
                    |  - TDDSettings    |
                    |  - HybridSettings |
                    +---------+---------+
                              |
          +-------------------+-------------------+
          |         |         |         |         |
    +-----+   +----+    +----+    +----+    +----+
    |Tested|  |Read |   |Under|  |Secur|  |Track|
    |Valid.|  |Valid.|   |Valid.|  |Valid.|  |Valid.|
    +-----+   +-----+   +-----+   +-----+   +-----+
       |         |          |         |         |
    +--+--+   +--+--+   +--+--+   +--+--+   +--+--+
    | LSP |   | LSP |   | LSP |   | LSP |   | Git |
    |Client|  |Client|  |Client|  |Client|  | Mgr |
    +-----+   +-----+   +-----+   +-----+   +-----+
```

### Dependency Graph

```
quality.Gate
  |-- quality.Validator (interface, per-principle)
  |     |-- LSPClient (interface, injected)
  |     |-- GitManager (interface, injected)
  |     |-- ASTAnalyzer (interface, injected)
  |-- quality.MethodologyValidator (mode-specific rules)
  |     |-- DDDSettings (characterization tests, behavior snapshots)
  |     |-- TDDSettings (test-first, min coverage per commit)
  |     |-- HybridSettings (new/legacy classification, dual coverage)
  |-- config.QualityConfig (loaded from quality.yaml)
  |-- context.Context (stdlib, cancellation/timeout)
  |-- sync.RWMutex (stdlib, baseline state protection)
  |-- log/slog (stdlib, structured logging)
```

---

## Milestones

### Primary Goal: Gate Interface and Core Types

**Scope**: Gate 인터페이스, Report/PrincipleResult/Issue 타입, TrustGate 기본 구조체 정의

**Tasks**:

1. `trust.go`에 Gate 인터페이스 정의
2. Report, PrincipleResult, Issue 구조체 정의 (JSON tags 포함)
3. Principle/Severity 상수 정의
4. Validator 내부 인터페이스 정의
5. TrustGate 구조체 정의 (생성자 NewTrustGate 포함)
6. ValidPrinciples 유효성 검사 헬퍼 구현
7. 단위 테스트: 타입 생성, JSON 직렬화, 유효성 검사

**Output Files**:
- `internal/core/quality/trust.go` (~200 LOC)
- `internal/core/quality/trust_test.go` (~150 LOC)

**Success Criteria**:
- Gate 인터페이스 컴파일 확인
- Report JSON 직렬화/역직렬화 테스트 통과
- 100% 타입 커버리지

---

### Secondary Goal: Per-Principle Validators

**Scope**: 5개 TRUST 원칙별 Validator 구현

**Tasks**:

1. **TestedValidator** 구현
   - LSPClient에서 type errors, errors 수집
   - 테스트 커버리지 데이터 수집 (go test -cover 출력 파싱 또는 LSP)
   - 커버리지 임계값 비교 (기본 85%)
   - PrincipleResult 생성 및 점수 계산

2. **ReadableValidator** 구현
   - LSPClient에서 lint errors 수집
   - 네이밍 컨벤션 위반 감지
   - PrincipleResult 생성

3. **UnderstandableValidator** 구현
   - LSPClient에서 warnings 수집
   - 코드 복잡도 분석 (설정 가능한 임계값)
   - 문서화 완성도 검사
   - PrincipleResult 생성

4. **SecuredValidator** 구현
   - LSPClient에서 security warnings 수집
   - OWASP 관련 패턴 검사
   - PrincipleResult 생성

5. **TrackableValidator** 구현
   - GitManager에서 커밋 메시지 검증 (Conventional Commits)
   - 구조화된 로그 사용 여부 확인
   - LSP 진단 이력 추적 상태 확인
   - PrincipleResult 생성

6. 각 Validator에 대한 table-driven 단위 테스트 작성

**Output Files**:
- `internal/core/quality/validators.go` (~350 LOC)
- `internal/core/quality/validators_test.go` (~300 LOC)

**Success Criteria**:
- 5개 Validator 모두 Validator 인터페이스 충족
- Mock LSPClient/GitManager로 테스트 통과
- 각 Validator에 대해 정상/실패/경계 케이스 테스트 포함

---

### Tertiary Goal: Aggregation, Phase Gates, Regression Detection

**Scope**: TrustGate의 Validate() 메서드 완성, 페이즈별 검증, 회귀 감지

**Tasks**:

1. **Validate() 메서드 완성**
   - 5개 Validator를 errgroup으로 병렬 실행
   - 결과 집계 및 가중 평균 점수 계산
   - Pass/Fail 판정 로직 구현
   - context 취소 처리

2. **ValidatePrinciple() 메서드 완성**
   - 원칙 이름 유효성 검사
   - 단일 Validator 실행 및 결과 반환
   - 잘못된 원칙 이름에 대한 오류 처리

3. **Phase-specific validation 구현**
   - Plan phase: baseline 캡처 (LSP 진단 스냅샷 저장)
   - Run phase: zero tolerance 강제 (errors=0, type_errors=0, lint_errors=0)
   - Sync phase: clean state 강제 (errors=0, warnings<=10, clean LSP)

4. **Regression detection 구현**
   - baseline과 현재 결과 비교
   - error_increase_threshold (기본 0) 검증
   - warning_increase_threshold (기본 10) 검증
   - type_error_increase_threshold (기본 0) 검증
   - 회귀 발견 시 Issue 목록에 추가

5. **통합 테스트 작성**
   - 전체 Validate() 흐름 테스트
   - 페이즈별 시나리오 테스트
   - 회귀 감지 시나리오 테스트
   - context 취소 테스트

**Output Files**:
- `internal/core/quality/trust.go` (추가 ~250 LOC)
- `internal/core/quality/trust_test.go` (추가 ~200 LOC)

**Success Criteria**:
- 전체 Validate() 흐름이 mock 의존성으로 동작
- 3개 페이즈(plan/run/sync)별 검증 통과
- 회귀 감지가 임계값 위반을 정확히 감지
- context 취소 시 부분 결과 반환

---

### Tertiary Goal B: Methodology-Aware Quality Validation

**Scope**: DDD/TDD/Hybrid 모드별 품질 게이트 검증 로직 구현

**Tasks**:

1. **DevelopmentMode 타입 및 설정 구조체 정의**
   - `DevelopmentMode` 타입 (ddd, tdd, hybrid) 정의
   - `TDDSettings`, `HybridSettings` 구조체 정의
   - `QualityConfig`에 `DevelopmentMode`, `TDDSettings`, `HybridSettings` 필드 추가
   - 유효하지 않은 모드에 대한 검증 로직 구현

2. **MethodologyValidator 구현**
   - `MethodologyValidator` 구조체 구현
   - DDD 모드: characterization 테스트 존재 확인, PRESERVE-before-IMPROVE 검증, behavior snapshot 회귀 검사
   - TDD 모드: 테스트-우선 검증 (타임스탬프 비교), 커밋당 최소 커버리지 검증, 커버리지 면제 거부
   - Hybrid 모드: git diff 기반 변경 분류 (new/legacy), 카테고리별 별도 커버리지 목표 적용

3. **Change Classification 로직 구현 (Hybrid 모드)**
   - Git diff 분석으로 새로운 파일/함수 vs 기존 파일 수정 분류
   - 새로운 파일 → TDD 규칙 적용
   - 기존 파일 수정 → DDD 규칙 적용
   - 기존 파일 내 새로운 함수 → TDD 규칙 적용

4. **Methodology Transition 처리 구현**
   - 모드 변경 감지 및 경고 로깅
   - 새 모드에 맞는 기준선 재계산
   - 이전 통과 코드에 대한 소급 실패 방지

5. **Report에 DevelopmentMode 포함**
   - `Report.DevelopmentMode` 필드 설정
   - JSON 직렬화 시 `development_mode` 포함 확인

6. **단위 테스트 작성**
   - DDD 모드 검증 시나리오 (characterization 테스트, behavior snapshot)
   - TDD 모드 검증 시나리오 (테스트-우선, 커밋 커버리지, 면제 거부)
   - Hybrid 모드 검증 시나리오 (변경 분류, 별도 커버리지)
   - 모드 전환 시나리오
   - 유효하지 않은 모드 거부 시나리오

**Output Files**:
- `internal/core/quality/trust.go` (추가 ~150 LOC)
- `internal/core/quality/validators.go` (추가 ~200 LOC)
- `internal/core/quality/trust_test.go` (추가 ~250 LOC)
- `internal/core/quality/validators_test.go` (추가 ~200 LOC)

**Success Criteria**:
- 3가지 모드(ddd, tdd, hybrid) 모두 올바른 규칙을 적용
- Hybrid 모드에서 git diff 기반 변경 분류가 정확히 동작
- 모드 전환 시 경고 로그가 기록되고 기준선이 재계산됨
- 유효하지 않은 모드가 명확한 에러로 거부됨
- Report에 development_mode가 항상 포함됨

---

### Optional Goal: Performance Optimization and Report Export

**Scope**: LSP 캐싱, 벤치마크, JSON 리포트 내보내기

**Tasks**:

1. LSP 진단 결과 TTL 캐싱 구현 (기본 5초)
2. 벤치마크 테스트 작성 (전체 검증 < 5s 확인)
3. Report JSON 파일 내보내기 기능
4. slog 기반 구조화된 검증 로깅

**Success Criteria**:
- 캐싱으로 반복 검증 시 50% 이상 성능 향상
- 벤치마크 목표 달성 (full validation < 5s)
- JSON 리포트 파일 생성 및 유효성 확인

---

## Technical Approach

### 1. Interface-First Design

모든 외부 의존성을 Go 인터페이스로 추상화한다. 이를 통해:

- **테스트 격리**: mockery로 생성된 mock을 사용하여 각 Validator를 독립적으로 테스트
- **DIP 준수**: 구체적 구현이 아닌 인터페이스에 의존
- **향후 확장**: LSP 클라이언트 교체, Git 백엔드 변경 등이 quality 패키지에 영향 없음

### 2. Goroutine-based Parallel Validation

```go
// Validate runs all 5 TRUST principles in parallel using errgroup.
func (g *TrustGate) Validate(ctx context.Context) (*Report, error) {
    eg, ctx := errgroup.WithContext(ctx)
    results := make(map[string]*PrincipleResult)
    var mu sync.Mutex

    for _, v := range g.validators {
        v := v // capture range variable
        eg.Go(func() error {
            result, err := v.Validate(ctx)
            if err != nil {
                return err
            }
            mu.Lock()
            results[v.Name()] = result
            mu.Unlock()
            return nil
        })
    }

    if err := eg.Wait(); err != nil {
        // Return partial results with error
        return g.buildPartialReport(results), err
    }

    return g.buildReport(results), nil
}
```

### 3. Configuration-Driven Thresholds

모든 임계값은 quality.yaml에서 로드되어 하드코딩을 방지한다:

- 테스트 커버리지 목표: `test_coverage_target` (기본 85)
- 페이즈별 게이트: `lsp_quality_gates.{plan,run,sync}`
- 회귀 임계값: `regression_detection.{error,warning,type_error}_increase_threshold`
- 캐시 TTL: `cache_ttl_seconds` (기본 5)
- 타임아웃: `timeout_seconds` (기본 3)

### 4. Error Handling Strategy

Go 관용구를 따른다:

- 모든 함수는 명시적으로 error를 반환한다
- panic은 사용하지 않는다
- `fmt.Errorf("quality: %w", err)` 형태로 error wrapping
- context 취소 시 `context.Canceled` 또는 `context.DeadlineExceeded` 반환
- 부분 실패 시 가용한 결과와 error를 함께 반환

### 5. Structured Logging

```go
slog.Info("quality gate validation started",
    "phase", phase,
    "principles", len(validators),
)

slog.Warn("regression detected",
    "metric", "errors",
    "baseline", baseline.Errors,
    "current", current.Errors,
    "threshold", config.ErrorIncreaseThreshold,
)
```

---

## Risks and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| LSP 의존성 미구현 (SPEC-LSP-001 지연) | Medium | High | LSPClient 인터페이스 mock으로 개발, 실제 통합은 LSP 완성 후 |
| Git 의존성 미구현 (SPEC-GIT-001 지연) | Medium | Medium | GitManager 인터페이스 mock으로 Trackable 검증 개발 |
| 병렬 실행 경쟁 조건 | Low | High | sync.Mutex로 결과 맵 보호, -race 플래그 테스트 필수 |
| LSP 진단 수집 타임아웃 | Medium | Medium | context.WithTimeout 적용, 캐싱으로 반복 호출 최소화 |
| 설정 스키마 변경 | Low | Low | 기본값 fallback, validation 오류 시 경고 로그 |

---

## Architectural Alignment

### ADR References

- **ADR-004 (Go Interfaces for DDD Boundaries)**: Gate, Validator, LSPClient, GitManager 인터페이스 적용
- **ADR-005 (log/slog for Structured Logging)**: 모든 검증 로그에 slog 사용
- **ADR-001 (Modular Monolithic)**: `internal/core/quality/` 패키지 경계 준수

### Product Document Alignment

- **Feature 5 (Quality Gates - TRUST 5)**: 이 SPEC의 직접 대상
- **Feature 3 (LSP Integration)**: LSPClient 인터페이스를 통한 통합
- **Feature 4 (Git Operations)**: GitManager 인터페이스를 통한 통합

### Quality Configuration Alignment

- `constitution.enforce_quality: true` - Gate.Validate()의 Pass/Fail이 워크플로우를 차단
- `constitution.test_coverage_target: 85` - TestedValidator의 커버리지 임계값
- `constitution.lsp_quality_gates` - 페이즈별 검증 임계값의 직접 소스
- `constitution.lsp_integration.trust5_integration` - 각 원칙의 LSP 매핑 정의
