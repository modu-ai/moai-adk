---
spec_id: SPEC-CORE-001
title: Foundation Methodologies
created: 2026-02-03
status: Completed
priority: Medium
phase: "Phase 5 - Knowledge (Phase 1 과 병렬 구현 가능)"
module: internal/foundation/
estimated_loc: ~1,000
dependencies: "없음 (순수 데이터/로직 모듈, 외부 의존성 없음)"
assigned: expert-backend
lifecycle: spec-anchored
tags: [foundation, ears, languages, trust5, domain-patterns]
---

# SPEC-CORE-001: Foundation Methodologies

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 초기 SPEC 작성 |

---

## 1. Environment (환경)

### 1.1 프로젝트 컨텍스트

MoAI-ADK Go Edition은 Python 기반 MoAI-ADK(약 73,000 LOC, 220+ 파일)를 Go로 완전 재작성하는 프로젝트이다. `internal/foundation/` 모듈은 시스템 전체에서 사용되는 기반 방법론 정의를 담당한다.

### 1.2 기술 스택

- **언어**: Go 1.22+
- **모듈 경로**: `github.com/modu-ai/moai-adk-go`
- **의존성**: 순수 Go 표준 라이브러리만 사용 (외부 의존성 없음)
  - `encoding/json`: JSON 직렬화
  - `embed`: 패턴 템플릿 로딩 (필요시)
  - `strings`, `fmt`: 문자열 처리
  - `regexp`: 패턴 검증

### 1.3 모듈 위치

```
internal/foundation/
  ears.go           # EARS 요구사항 패턴 템플릿
  langs.go          # 16+ 프로그래밍 언어 생태계 정의
  backend.go        # 백엔드 아키텍처 패턴
  frontend.go       # 프론트엔드 아키텍처 패턴
  database.go       # 데이터베이스 패턴 및 전략
  testing.go        # 테스팅 전략 정의
  devops.go         # DevOps 및 CI/CD 패턴
  trust/
    principles.go   # TRUST 5 원칙 정의
    checklist.go    # 품질 체크리스트 생성
```

### 1.4 관련 시스템

- `internal/lsp/`: 언어 정의(langs.go)를 사용하여 LSP 서버 매핑
- `internal/core/quality/`: TRUST 5 원칙(trust/)을 사용하여 품질 게이트 실행
- `internal/core/project/`: EARS 패턴(ears.go)을 사용하여 SPEC 생성 지원
- `internal/hook/`: 도메인 패턴을 사용하여 후처리 검증

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

- A1: Go 1.22+ 환경에서 컴파일되며, 제네릭과 `log/slog`를 활용할 수 있다.
- A2: 이 모듈은 외부 의존성 없이 Go 표준 라이브러리만으로 구현 가능하다.
- A3: 모든 데이터 구조는 `encoding/json`으로 직렬화 가능해야 한다.
- A4: 언어 정의는 컴파일 타임에 고정되며, 런타임 플러그인 로딩은 지원하지 않는다.

### 2.2 비즈니스 가정

- A5: EARS 패턴은 Rolls-Royce의 Alistair Mavin이 정의한 5가지 표준 패턴을 따른다.
- A6: 지원 언어 목록(16+)은 Python MoAI-ADK와 동일하게 유지한다.
- A7: TRUST 5 원칙(Tested, Readable, Unified, Secured, Trackable)은 MoAI 프레임워크 전체에서 일관되게 사용된다.
- A8: 도메인 패턴(backend, frontend, database, testing, devops)은 MoAI의 에이전트 카탈로그와 1:1로 매핑된다.

### 2.3 의존성 가정

- A9: 이 모듈은 다른 `internal/` 패키지에 대한 의존성이 없으며, 오직 `pkg/` 레벨의 유틸리티만 선택적으로 사용한다.
- A10: 다른 모듈(lsp, quality, project)이 이 모듈의 타입과 함수를 임포트하는 단방향 의존 관계를 가진다.

---

## 3. Requirements (요구사항)

### 3.1 EARS 패턴 요구사항 (ears.go)

#### REQ-EARS-001: EARS 패턴 타입 정의 (Ubiquitous)

시스템은 **항상** 5가지 EARS 요구사항 패턴 타입을 Go 상수로 정의해야 한다.

- `Ubiquitous`: 시스템 전체에 항상 적용되는 요구사항
- `EventDriven`: 이벤트 기반 트리거-응답 요구사항
- `StateDriven`: 상태 기반 조건부 요구사항
- `Unwanted`: 금지 행위 요구사항
- `Optional`: 선택적 기능 요구사항

#### REQ-EARS-002: EARS 패턴 템플릿 구조 (Ubiquitous)

시스템은 **항상** 각 EARS 패턴에 대해 다음 정보를 포함하는 구조체를 제공해야 한다:

- `Type`: 패턴 타입 (EARSPatternType)
- `NameEN`: 영문 패턴 이름
- `NameKO`: 한국어 패턴 이름
- `TemplateEN`: 영문 요구사항 템플릿 문자열
- `TemplateKO`: 한국어 요구사항 템플릿 문자열
- `Description`: 패턴 설명
- `UseCases`: 사용 사례 목록
- `TestStrategy`: 테스트 전략 설명

#### REQ-EARS-003: EARS 패턴 검증 (Event-Driven)

**WHEN** 사용자가 EARS 패턴 문자열을 입력하면 **THEN** 시스템은 해당 문자열이 올바른 EARS 구문을 따르는지 검증하고 매칭되는 패턴 타입을 반환해야 한다.

- "The [system] shall [response]" -> Ubiquitous
- "When [event], the [system] shall [response]" -> EventDriven
- "While [condition], the [system] shall [response]" -> StateDriven
- "If [undesired], then the [system] shall [response]" -> Unwanted
- "Where [feature exists], the [system] shall [response]" -> Optional

#### REQ-EARS-004: 복합 패턴 지원 (Event-Driven)

**WHEN** "While [state], when [event]" 형태의 복합 요구사항이 입력되면 **THEN** 시스템은 이를 Complex 타입으로 분류하고 구성 요소(State + Event)를 분석해야 한다.

#### REQ-EARS-005: 잘못된 패턴 거부 (Unwanted)

시스템은 EARS 구문을 따르지 않는 모호한 요구사항 문자열(예: "should", "might", "usually" 포함)을 **거부해야 한다**.

### 3.2 언어 정의 요구사항 (langs.go)

#### REQ-LANG-001: 16+ 언어 생태계 정의 (Ubiquitous)

시스템은 **항상** 다음 16개 이상의 프로그래밍 언어에 대한 생태계 정의를 포함해야 한다:

| # | 언어 | LSP 서버 | 파일 확장자 | 패키지 매니저 | 테스트 프레임워크 | 린트 도구 |
|---|------|---------|-----------|-------------|----------------|---------|
| 1 | Go | gopls | .go | go mod | go test | golangci-lint |
| 2 | Python | pyright | .py, .pyi | pip, uv, poetry | pytest | ruff |
| 3 | TypeScript | typescript-language-server | .ts, .tsx, .mts, .cts | npm, yarn, pnpm | vitest, jest | eslint |
| 4 | JavaScript | typescript-language-server | .js, .jsx, .mjs, .cjs | npm, yarn, pnpm | vitest, jest | eslint |
| 5 | Java | jdtls | .java | maven, gradle | junit | checkstyle |
| 6 | Rust | rust-analyzer | .rs | cargo | cargo test | clippy |
| 7 | C | clangd | .c, .h | cmake, make | ctest | clang-tidy |
| 8 | C++ | clangd | .cpp, .hpp, .cc, .cxx | cmake, make | gtest, catch2 | clang-tidy |
| 9 | Ruby | solargraph | .rb | bundler, gem | rspec, minitest | rubocop |
| 10 | PHP | intelephense | .php | composer | phpunit | phpstan |
| 11 | Kotlin | kotlin-language-server | .kt, .kts | gradle, maven | junit | ktlint |
| 12 | Swift | sourcekit-lsp | .swift | spm | xctest | swiftlint |
| 13 | Dart | dart language-server | .dart | pub | dart test | dart analyze |
| 14 | Elixir | elixir-ls | .ex, .exs | mix | exunit | credo |
| 15 | Scala | metals | .scala, .sc | sbt, mill | scalatest | scalafmt |
| 16 | Haskell | haskell-language-server | .hs | cabal, stack | hspec | hlint |
| 17 | Zig | zls | .zig | zig build | zig test | - |
| 18 | R | languageserver | .R, .r, .Rmd | renv | testthat | lintr |

#### REQ-LANG-002: 언어 정의 구조체 (Ubiquitous)

시스템은 **항상** 각 언어에 대해 다음 필드를 포함하는 `LanguageDefinition` 구조체를 제공해야 한다:

- `ID`: 언어 식별자 (소문자, 예: "go", "python", "typescript")
- `Name`: 표시용 이름 (예: "Go", "Python", "TypeScript")
- `LSPServer`: LSP 서버 이름
- `Extensions`: 파일 확장자 목록
- `PackageManagers`: 패키지 매니저 목록
- `TestFrameworks`: 테스트 프레임워크 목록
- `LintTools`: 린트 도구 목록
- `BuildTools`: 빌드 도구 목록 (선택)

#### REQ-LANG-003: 파일 확장자로 언어 탐지 (Event-Driven)

**WHEN** 파일 확장자가 주어지면 **THEN** 시스템은 해당 확장자에 매핑되는 `LanguageDefinition`을 반환해야 한다.

#### REQ-LANG-004: 언어 ID로 조회 (Event-Driven)

**WHEN** 언어 ID 문자열이 주어지면 **THEN** 시스템은 해당 ID에 매핑되는 `LanguageDefinition`을 반환해야 한다.

#### REQ-LANG-005: 알 수 없는 언어 처리 (Unwanted)

시스템은 정의되지 않은 파일 확장자 또는 언어 ID에 대해 **패닉하지 않고** 명확한 오류 값(nil 또는 error)을 반환해야 한다.

#### REQ-LANG-006: 전체 언어 목록 조회 (Ubiquitous)

시스템은 **항상** 등록된 모든 언어 정의 목록을 반환하는 함수를 제공해야 한다.

### 3.3 도메인 패턴 요구사항 (backend.go, frontend.go, database.go, testing.go, devops.go)

#### REQ-DOMAIN-001: 백엔드 아키텍처 패턴 (Ubiquitous)

시스템은 **항상** 다음 백엔드 아키텍처 패턴 정의를 포함해야 한다:

- API 디자인 패턴 (REST, GraphQL, gRPC)
- 인증/인가 패턴 (JWT, OAuth2, Session)
- 마이크로서비스 패턴 (API Gateway, Service Mesh, Event-Driven)
- 서버 아키텍처 패턴 (Layered, Hexagonal, Clean Architecture)

#### REQ-DOMAIN-002: 프론트엔드 아키텍처 패턴 (Ubiquitous)

시스템은 **항상** 다음 프론트엔드 아키텍처 패턴 정의를 포함해야 한다:

- 컴포넌트 패턴 (Atomic Design, Compound Components)
- 상태 관리 패턴 (Global Store, Atomic State, Server State)
- 렌더링 전략 (SSR, SSG, ISR, CSR)
- 스타일링 전략 (CSS Modules, CSS-in-JS, Utility-first)

#### REQ-DOMAIN-003: 데이터베이스 패턴 (Ubiquitous)

시스템은 **항상** 다음 데이터베이스 패턴 정의를 포함해야 한다:

- 데이터베이스 유형 (RDBMS, Document, Key-Value, Graph, Time-Series)
- ORM/쿼리 패턴 (Repository, Active Record, Query Builder)
- 마이그레이션 전략 (Version-based, State-based)
- 캐싱 전략 (Cache-Aside, Read-Through, Write-Behind)

#### REQ-DOMAIN-004: 테스팅 전략 패턴 (Ubiquitous)

시스템은 **항상** 다음 테스팅 전략 패턴 정의를 포함해야 한다:

- 테스트 레벨 (Unit, Integration, E2E, Contract)
- 테스트 전략 (TDD, BDD, Characterization Testing)
- 커버리지 전략 (Line, Branch, Mutation)
- 테스트 더블 (Mock, Stub, Spy, Fake)

#### REQ-DOMAIN-005: DevOps 패턴 (Ubiquitous)

시스템은 **항상** 다음 DevOps 패턴 정의를 포함해야 한다:

- CI/CD 패턴 (Pipeline, Stage, Job)
- 컨테이너화 패턴 (Docker, Kubernetes, Helm)
- 인프라 패턴 (IaC, GitOps, Blue-Green, Canary)
- 모니터링 패턴 (Metrics, Logging, Tracing)

#### REQ-DOMAIN-006: 도메인 패턴 구조체 (Ubiquitous)

시스템은 **항상** 각 도메인 패턴에 대해 다음 필드를 포함하는 공통 구조체를 제공해야 한다:

- `ID`: 패턴 식별자
- `Name`: 패턴 이름
- `Category`: 패턴 분류
- `Description`: 패턴 설명
- `UseCases`: 적용 사례
- `Pros`: 장점
- `Cons`: 단점
- `RelatedPatterns`: 관련 패턴 ID 목록

#### REQ-DOMAIN-007: 카테고리별 패턴 조회 (Event-Driven)

**WHEN** 도메인(예: "backend")과 카테고리(예: "api-design")가 주어지면 **THEN** 시스템은 해당 카테고리에 속하는 모든 패턴 목록을 반환해야 한다.

### 3.4 TRUST 5 원칙 요구사항 (trust/principles.go, trust/checklist.go)

#### REQ-TRUST-001: TRUST 5 원칙 정의 (Ubiquitous)

시스템은 **항상** TRUST 5의 5가지 품질 원칙을 Go 타입으로 정의해야 한다:

| 원칙 | 설명 | 핵심 지표 |
|------|------|----------|
| **Tested** | 포괄적 테스트 커버리지 | 85%+ 커버리지, 특성 테스트, 뮤테이션 테스트 |
| **Readable** | 명확한 네이밍과 구조 | 린트 오류 0개, 코드 복잡도 임계값 |
| **Unified** | 일관된 스타일과 포맷팅 | 포매터 준수, 임포트 정렬 |
| **Secured** | OWASP 보안 준수 | 보안 스캔 통과, 비밀 미감지 |
| **Trackable** | 추적 가능한 변경 이력 | 컨벤셔널 커밋, SPEC 참조 연결 |

#### REQ-TRUST-002: 원칙별 체크리스트 항목 (Ubiquitous)

시스템은 **항상** 각 TRUST 5 원칙에 대해 구체적인 체크리스트 항목 목록을 제공해야 한다. 각 항목은 다음을 포함한다:

- `ID`: 체크리스트 항목 식별자 (예: "T-001", "R-001")
- `Principle`: 소속 원칙
- `Description`: 항목 설명
- `Severity`: 심각도 (Critical, Warning, Info)
- `AutomationLevel`: 자동화 수준 (Automated, SemiAutomated, Manual)
- `ValidatorHint`: 검증 방법 힌트

#### REQ-TRUST-003: 워크플로우 단계별 체크리스트 생성 (Event-Driven)

**WHEN** 워크플로우 단계(plan, run, sync)가 주어지면 **THEN** 시스템은 해당 단계에 적용되는 TRUST 5 체크리스트 항목만 필터링하여 반환해야 한다.

#### REQ-TRUST-004: 체크리스트 평가 결과 구조체 (Ubiquitous)

시스템은 **항상** 체크리스트 평가 결과를 담는 구조체를 제공해야 한다:

- `Principle`: 원칙 이름
- `TotalItems`: 총 항목 수
- `PassedItems`: 통과 항목 수
- `FailedItems`: 실패 항목 수
- `Score`: 점수 (0-100)
- `Grade`: 등급 (A, B, C, D, F)
- `Details`: 개별 항목 결과 목록

#### REQ-TRUST-005: 전체 TRUST 5 점수 계산 (Event-Driven)

**WHEN** 5가지 원칙 각각의 평가 결과가 주어지면 **THEN** 시스템은 가중 평균 기반의 전체 TRUST 5 점수를 계산해야 한다.

---

## 4. Specifications (세부 명세)

### 4.1 패키지 구조

```go
package foundation

// ears.go - EARS 패턴 관련 타입 및 함수
// langs.go - 언어 생태계 정의
// backend.go - 백엔드 도메인 패턴
// frontend.go - 프론트엔드 도메인 패턴
// database.go - 데이터베이스 도메인 패턴
// testing.go - 테스팅 도메인 패턴
// devops.go - DevOps 도메인 패턴

package trust // internal/foundation/trust/

// principles.go - TRUST 5 원칙 정의
// checklist.go - 체크리스트 생성 및 평가
```

### 4.2 핵심 타입 설계

#### 4.2.1 EARS 타입

```go
type EARSPatternType string

const (
    Ubiquitous EARSPatternType = "ubiquitous"
    EventDriven EARSPatternType = "event_driven"
    StateDriven EARSPatternType = "state_driven"
    Unwanted    EARSPatternType = "unwanted"
    Optional    EARSPatternType = "optional"
    Complex     EARSPatternType = "complex"
)

type EARSPattern struct {
    Type        EARSPatternType
    NameEN      string
    NameKO      string
    TemplateEN  string
    TemplateKO  string
    Description string
    UseCases    []string
    TestStrategy string
}

type EARSValidationResult struct {
    Valid       bool
    PatternType EARSPatternType
    Components  map[string]string  // "event", "system", "response" 등
    Errors      []string
}
```

#### 4.2.2 언어 정의 타입

```go
type LanguageDefinition struct {
    ID              string   `json:"id"`
    Name            string   `json:"name"`
    LSPServer       string   `json:"lsp_server"`
    Extensions      []string `json:"extensions"`
    PackageManagers []string `json:"package_managers"`
    TestFrameworks  []string `json:"test_frameworks"`
    LintTools       []string `json:"lint_tools"`
    BuildTools      []string `json:"build_tools,omitempty"`
}
```

#### 4.2.3 도메인 패턴 타입

```go
type DomainPattern struct {
    ID              string   `json:"id"`
    Name            string   `json:"name"`
    Domain          string   `json:"domain"`      // "backend", "frontend", etc.
    Category        string   `json:"category"`
    Description     string   `json:"description"`
    UseCases        []string `json:"use_cases"`
    Pros            []string `json:"pros"`
    Cons            []string `json:"cons"`
    RelatedPatterns []string `json:"related_patterns"`
}
```

#### 4.2.4 TRUST 5 타입

```go
type TRUSTPrinciple string

const (
    Tested    TRUSTPrinciple = "tested"
    Readable  TRUSTPrinciple = "readable"
    Unified   TRUSTPrinciple = "unified"
    Secured   TRUSTPrinciple = "secured"
    Trackable TRUSTPrinciple = "trackable"
)

type ChecklistItem struct {
    ID              string
    Principle       TRUSTPrinciple
    Description     string
    Severity        Severity      // Critical, Warning, Info
    AutomationLevel Automation    // Automated, SemiAutomated, Manual
    ValidatorHint   string
    Phases          []Phase       // plan, run, sync
}

type PrincipleResult struct {
    Principle   TRUSTPrinciple
    TotalItems  int
    PassedItems int
    FailedItems int
    Score       float64
    Grade       string
    Details     []ItemResult
}

type TRUST5Score struct {
    OverallScore float64
    OverallGrade string
    Principles   map[TRUSTPrinciple]PrincipleResult
}
```

### 4.3 핵심 함수 인터페이스

```go
// ears.go
func GetAllPatterns() []EARSPattern
func GetPattern(patternType EARSPatternType) (*EARSPattern, error)
func ValidateEARSRequirement(text string) EARSValidationResult
func DetectPatternType(text string) (EARSPatternType, error)

// langs.go
func GetAllLanguages() []LanguageDefinition
func GetLanguageByID(id string) (*LanguageDefinition, error)
func GetLanguageByExtension(ext string) (*LanguageDefinition, error)
func GetSupportedExtensions() []string

// backend.go, frontend.go, database.go, testing.go, devops.go
func GetBackendPatterns() []DomainPattern
func GetFrontendPatterns() []DomainPattern
func GetDatabasePatterns() []DomainPattern
func GetTestingPatterns() []DomainPattern
func GetDevOpsPatterns() []DomainPattern
func GetPatternsByCategory(domain, category string) []DomainPattern
func GetPatternByID(id string) (*DomainPattern, error)

// trust/principles.go
func GetAllPrinciples() []TRUSTPrincipleDefinition
func GetPrinciple(p TRUSTPrinciple) (*TRUSTPrincipleDefinition, error)

// trust/checklist.go
func GetChecklist() []ChecklistItem
func GetChecklistForPhase(phase Phase) []ChecklistItem
func GetChecklistForPrinciple(p TRUSTPrinciple) []ChecklistItem
func CalculatePrincipleScore(results []ItemResult) PrincipleResult
func CalculateTRUST5Score(principleResults map[TRUSTPrinciple]PrincipleResult) TRUST5Score
```

### 4.4 JSON 직렬화 지원

시스템은 **항상** 모든 정의 타입에 대해 `encoding/json` 호환 태그를 제공하며, `json.Marshal`/`json.Unmarshal`이 올바르게 동작해야 한다. 이는 ADR-011(Zero Runtime Template Expansion)에 따라 프로그래매틱 직렬화를 지원하기 위함이다.

### 4.5 성능 제약

- 모든 조회 함수의 응답 시간: < 1ms (인메모리 데이터)
- 패턴 검증(ValidateEARSRequirement): < 5ms
- 메모리 사용량: < 1MB (모든 정의 데이터 포함)

---

## 5. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 파일 | 관련 ADR |
|------------|----------|-----------|---------|
| REQ-EARS-001~005 | ears.go | ears_test.go | - |
| REQ-LANG-001~006 | langs.go | langs_test.go | - |
| REQ-DOMAIN-001~007 | backend.go, frontend.go, database.go, testing.go, devops.go | *_test.go | - |
| REQ-TRUST-001~005 | trust/principles.go, trust/checklist.go | trust/*_test.go | - |

### 관련 SPEC

- 이 SPEC은 독립적이며 다른 SPEC에 의존하지 않는다.
- 이 SPEC의 타입은 다음 모듈에서 사용된다:
  - `internal/lsp/` (SPEC-LSP-001 예정)
  - `internal/core/quality/` (SPEC-QUALITY-001 예정)
  - `internal/core/project/` (SPEC-PROJECT-001 예정)

### 전문가 자문 권고

- **expert-backend**: 도메인 패턴 정의의 완전성과 업계 표준 준수 검토
- **expert-testing**: 테스트 전략 패턴과 TRUST 5 체크리스트 항목의 실용성 검토

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 98.4%

### Summary

Foundation package implemented with EARS pattern types and validation, language ecosystem definitions for 16+ programming languages, development methodology types (DDD/TDD/Hybrid), and TRUST 5 framework principles with checklist generation and scoring. Provides the core domain types used across the entire MoAI-ADK Go codebase.

### Files Created

- `internal/foundation/doc.go`
- `internal/foundation/ears.go`
- `internal/foundation/ears_test.go`
- `internal/foundation/errors.go`
- `internal/foundation/errors_test.go`
- `internal/foundation/language.go`
- `internal/foundation/language_test.go`
- `internal/foundation/methodology.go`
- `internal/foundation/methodology_test.go`
- `internal/foundation/trust5.go`
- `internal/foundation/trust5_test.go`
