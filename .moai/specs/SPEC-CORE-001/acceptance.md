---
spec_id: SPEC-CORE-001
title: Foundation Methodologies - Acceptance Criteria
created: 2026-02-03
status: Planned
tags: [foundation, ears, languages, trust5, domain-patterns]
---

# SPEC-CORE-001: Foundation Methodologies - 수용 기준

## 1. EARS 패턴 수용 기준

### AC-EARS-001: 5가지 EARS 패턴 타입 정의 확인

```gherkin
Given EARS 패턴 모듈이 로드되었을 때
When GetAllPatterns() 함수를 호출하면
Then 정확히 6개의 패턴이 반환된다 (Ubiquitous, EventDriven, StateDriven, Unwanted, Optional, Complex)
And 각 패턴은 NameEN, NameKO, TemplateEN, TemplateKO, Description, UseCases, TestStrategy 필드를 가진다
And 모든 필드는 빈 문자열이 아니다
```

### AC-EARS-002: 개별 패턴 조회

```gherkin
Given EARS 패턴 모듈이 로드되었을 때
When GetPattern(EventDriven) 함수를 호출하면
Then EventDriven 패턴 정보가 반환된다
And TemplateEN은 "When" 키워드를 포함한다
And TemplateKO는 "WHEN" 또는 "이벤트" 키워드를 포함한다

Given EARS 패턴 모듈이 로드되었을 때
When GetPattern("invalid_type") 함수를 호출하면
Then nil과 오류가 반환된다
And 오류 메시지는 "unknown EARS pattern type" 문구를 포함한다
```

### AC-EARS-003: Ubiquitous 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("The system shall log all API requests") 함수를 호출하면
Then Valid가 true이다
And PatternType이 Ubiquitous이다
And Components에 "system"과 "response" 키가 존재한다
```

### AC-EARS-004: EventDriven 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("When the user clicks submit, the system shall validate the form") 함수를 호출하면
Then Valid가 true이다
And PatternType이 EventDriven이다
And Components["event"]는 "the user clicks submit"를 포함한다
And Components["response"]는 "validate the form"을 포함한다
```

### AC-EARS-005: StateDriven 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("While the user is authenticated, the system shall display the dashboard") 함수를 호출하면
Then Valid가 true이다
And PatternType이 StateDriven이다
And Components["condition"]는 "the user is authenticated"를 포함한다
```

### AC-EARS-006: Unwanted 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("If the session expires, then the system shall redirect to login page") 함수를 호출하면
Then Valid가 true이다
And PatternType이 Unwanted이다
```

### AC-EARS-007: Optional 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("Where OAuth is available, the system shall provide social login") 함수를 호출하면
Then Valid가 true이다
And PatternType이 Optional이다
```

### AC-EARS-008: Complex 패턴 검증

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("While the server is under high load, when a new request arrives, the system shall queue the request") 함수를 호출하면
Then Valid가 true이다
And PatternType이 Complex이다
And Components에 "condition"과 "event" 키가 모두 존재한다
```

### AC-EARS-009: 모호한 요구사항 거부

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("The system should handle errors gracefully") 함수를 호출하면
Then Valid가 false이다
And Errors에 "ambiguous" 또는 "should" 관련 오류 메시지가 포함된다

Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("Users might want to export data") 함수를 호출하면
Then Valid가 false이다
And Errors에 "ambiguous" 또는 "might" 관련 오류 메시지가 포함된다

Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("The system usually responds within 2 seconds") 함수를 호출하면
Then Valid가 false이다
And Errors에 "ambiguous" 또는 "usually" 관련 오류 메시지가 포함된다
```

### AC-EARS-010: 빈 문자열 처리

```gherkin
Given EARS 검증 엔진이 준비되었을 때
When ValidateEARSRequirement("") 함수를 호출하면
Then Valid가 false이다
And Errors에 "empty" 관련 오류 메시지가 포함된다
```

### AC-EARS-011: DetectPatternType 함수

```gherkin
Given EARS 패턴 탐지가 준비되었을 때
When DetectPatternType("When X, the system shall Y") 함수를 호출하면
Then EventDriven이 반환된다

When DetectPatternType("random text without EARS keywords") 함수를 호출하면
Then 오류가 반환된다
```

---

## 2. 언어 정의 수용 기준

### AC-LANG-001: 전체 언어 목록 확인

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetAllLanguages() 함수를 호출하면
Then 최소 16개의 언어 정의가 반환된다
And 각 정의는 ID, Name, LSPServer, Extensions, PackageManagers, TestFrameworks, LintTools 필드를 가진다
And 모든 필수 필드는 빈 값이 아니다
```

### AC-LANG-002: 18개 언어 개별 조회 테스트

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetLanguageByID("go") 함수를 호출하면
Then Name이 "Go"이다
And LSPServer가 "gopls"이다
And Extensions에 ".go"가 포함된다
And PackageManagers에 "go mod"가 포함된다
And TestFrameworks에 "go test"가 포함된다
And LintTools에 "golangci-lint"가 포함된다

When GetLanguageByID("python") 함수를 호출하면
Then Name이 "Python"이다
And LSPServer가 "pyright"이다
And Extensions에 ".py"와 ".pyi"가 포함된다

When GetLanguageByID("typescript") 함수를 호출하면
Then Name이 "TypeScript"이다
And Extensions에 ".ts", ".tsx", ".mts", ".cts"가 포함된다

When GetLanguageByID("javascript") 함수를 호출하면
Then Name이 "JavaScript"이다
And Extensions에 ".js", ".jsx", ".mjs", ".cjs"가 포함된다

When GetLanguageByID("java") 함수를 호출하면
Then Name이 "Java"이다
And LSPServer가 "jdtls"이다

When GetLanguageByID("rust") 함수를 호출하면
Then Name이 "Rust"이다
And LSPServer가 "rust-analyzer"이다
And LintTools에 "clippy"가 포함된다

When GetLanguageByID("c") 함수를 호출하면
Then Name이 "C"이다
And LSPServer가 "clangd"이다

When GetLanguageByID("cpp") 함수를 호출하면
Then Name이 "C++"이다
And Extensions에 ".cpp", ".hpp"가 포함된다

When GetLanguageByID("ruby") 함수를 호출하면
Then Name이 "Ruby"이다
And LSPServer가 "solargraph"이다

When GetLanguageByID("php") 함수를 호출하면
Then Name이 "PHP"이다
And LSPServer가 "intelephense"이다

When GetLanguageByID("kotlin") 함수를 호출하면
Then Name이 "Kotlin"이다
And Extensions에 ".kt"와 ".kts"가 포함된다

When GetLanguageByID("swift") 함수를 호출하면
Then Name이 "Swift"이다
And LSPServer가 "sourcekit-lsp"이다

When GetLanguageByID("dart") 함수를 호출하면
Then Name이 "Dart"이다
And Extensions에 ".dart"가 포함된다

When GetLanguageByID("elixir") 함수를 호출하면
Then Name이 "Elixir"이다
And Extensions에 ".ex"와 ".exs"가 포함된다

When GetLanguageByID("scala") 함수를 호출하면
Then Name이 "Scala"이다
And LSPServer가 "metals"이다

When GetLanguageByID("haskell") 함수를 호출하면
Then Name이 "Haskell"이다
And LSPServer가 "haskell-language-server"이다

When GetLanguageByID("zig") 함수를 호출하면
Then Name이 "Zig"이다
And LSPServer가 "zls"이다

When GetLanguageByID("r") 함수를 호출하면
Then Name이 "R"이다
And Extensions에 ".R"과 ".r"이 포함된다
```

### AC-LANG-003: 파일 확장자 기반 언어 탐지

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetLanguageByExtension(".go") 함수를 호출하면
Then 반환된 언어의 ID가 "go"이다

When GetLanguageByExtension(".py") 함수를 호출하면
Then 반환된 언어의 ID가 "python"이다

When GetLanguageByExtension(".ts") 함수를 호출하면
Then 반환된 언어의 ID가 "typescript"이다

When GetLanguageByExtension(".tsx") 함수를 호출하면
Then 반환된 언어의 ID가 "typescript"이다

When GetLanguageByExtension(".rs") 함수를 호출하면
Then 반환된 언어의 ID가 "rust"이다

When GetLanguageByExtension(".kt") 함수를 호출하면
Then 반환된 언어의 ID가 "kotlin"이다

When GetLanguageByExtension(".swift") 함수를 호출하면
Then 반환된 언어의 ID가 "swift"이다

When GetLanguageByExtension(".zig") 함수를 호출하면
Then 반환된 언어의 ID가 "zig"이다
```

### AC-LANG-004: 알 수 없는 언어 처리

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetLanguageByID("brainfuck") 함수를 호출하면
Then nil과 오류가 반환된다
And 프로그램이 패닉하지 않는다

When GetLanguageByExtension(".xyz") 함수를 호출하면
Then nil과 오류가 반환된다
And 프로그램이 패닉하지 않는다

When GetLanguageByExtension("") 함수를 호출하면
Then nil과 오류가 반환된다
And 오류 메시지는 "empty extension" 문구를 포함한다
```

### AC-LANG-005: 대소문자 무시 확장자 매칭

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetLanguageByExtension(".R") 함수를 호출하면
Then 반환된 언어의 ID가 "r"이다

When GetLanguageByExtension(".r") 함수를 호출하면
Then 반환된 언어의 ID가 "r"이다
```

### AC-LANG-006: 전체 확장자 목록 조회

```gherkin
Given 언어 정의 모듈이 로드되었을 때
When GetSupportedExtensions() 함수를 호출하면
Then 반환된 목록의 길이가 40 이상이다 (18개 언어의 확장자 합계)
And ".go", ".py", ".ts", ".rs", ".java"가 포함된다
And 중복 확장자가 없다
```

### AC-LANG-007: JSON 직렬화/역직렬화

```gherkin
Given Go 언어 정의를 GetLanguageByID("go")로 조회했을 때
When json.Marshal()로 직렬화하면
Then 유효한 JSON 문자열이 생성된다
And json.Unmarshal()로 역직렬화하면 원본과 동일한 구조체가 복원된다
```

---

## 3. 도메인 패턴 수용 기준

### AC-DOMAIN-001: 백엔드 패턴 조회

```gherkin
Given 백엔드 패턴 모듈이 로드되었을 때
When GetBackendPatterns() 함수를 호출하면
Then 최소 10개의 백엔드 패턴이 반환된다
And "api-design", "authentication", "microservice", "architecture" 카테고리가 존재한다
And 각 패턴은 ID, Name, Category, Description 필드를 가진다
```

### AC-DOMAIN-002: 프론트엔드 패턴 조회

```gherkin
Given 프론트엔드 패턴 모듈이 로드되었을 때
When GetFrontendPatterns() 함수를 호출하면
Then 최소 8개의 프론트엔드 패턴이 반환된다
And "component", "state-management", "rendering", "styling" 카테고리가 존재한다
```

### AC-DOMAIN-003: 데이터베이스 패턴 조회

```gherkin
Given 데이터베이스 패턴 모듈이 로드되었을 때
When GetDatabasePatterns() 함수를 호출하면
Then 최소 8개의 데이터베이스 패턴이 반환된다
And "database-type", "orm", "migration", "caching" 카테고리가 존재한다
```

### AC-DOMAIN-004: 테스팅 패턴 조회

```gherkin
Given 테스팅 패턴 모듈이 로드되었을 때
When GetTestingPatterns() 함수를 호출하면
Then 최소 8개의 테스팅 패턴이 반환된다
And "test-level", "test-strategy", "coverage", "test-double" 카테고리가 존재한다
```

### AC-DOMAIN-005: DevOps 패턴 조회

```gherkin
Given DevOps 패턴 모듈이 로드되었을 때
When GetDevOpsPatterns() 함수를 호출하면
Then 최소 8개의 DevOps 패턴이 반환된다
And "cicd", "container", "infrastructure", "monitoring" 카테고리가 존재한다
```

### AC-DOMAIN-006: 카테고리별 패턴 필터링

```gherkin
Given 도메인 패턴 모듈이 로드되었을 때
When GetPatternsByCategory("backend", "api-design") 함수를 호출하면
Then 반환된 패턴의 Domain이 모두 "backend"이다
And 반환된 패턴의 Category가 모두 "api-design"이다
And 최소 2개의 패턴이 반환된다 (예: REST, GraphQL, gRPC)

When GetPatternsByCategory("unknown", "unknown") 함수를 호출하면
Then 빈 슬라이스가 반환된다 (nil이 아님)
And 프로그램이 패닉하지 않는다
```

### AC-DOMAIN-007: 패턴 ID 조회

```gherkin
Given 도메인 패턴 모듈이 로드되었을 때
When GetPatternByID("backend-rest-api") 함수를 호출하면
Then 해당 패턴의 Name이 빈 문자열이 아니다
And Domain이 "backend"이다

When GetPatternByID("nonexistent-pattern") 함수를 호출하면
Then nil과 오류가 반환된다
```

### AC-DOMAIN-008: 관련 패턴 참조 무결성

```gherkin
Given 모든 도메인 패턴이 로드되었을 때
When 각 패턴의 RelatedPatterns 필드를 순회하면
Then 참조된 모든 패턴 ID가 실제로 존재한다 (고아 참조 없음)
```

---

## 4. TRUST 5 수용 기준

### AC-TRUST-001: 5가지 원칙 정의 확인

```gherkin
Given TRUST 5 모듈이 로드되었을 때
When GetAllPrinciples() 함수를 호출하면
Then 정확히 5개의 원칙이 반환된다
And Tested, Readable, Unified, Secured, Trackable 원칙이 모두 포함된다
And 각 원칙은 Name, Description, KeyMetrics, Weight 필드를 가진다
And 모든 Weight의 합이 1.0 (100%)이다
```

### AC-TRUST-002: 체크리스트 항목 확인

```gherkin
Given TRUST 5 모듈이 로드되었을 때
When GetChecklist() 함수를 호출하면
Then 최소 25개의 체크리스트 항목이 반환된다 (원칙당 최소 5개)
And 각 항목은 ID, Principle, Description, Severity, AutomationLevel, Phases 필드를 가진다
And 모든 항목의 Principle이 유효한 TRUST 5 원칙에 속한다
```

### AC-TRUST-003: 원칙별 체크리스트 필터링

```gherkin
Given TRUST 5 모듈이 로드되었을 때
When GetChecklistForPrinciple(Tested) 함수를 호출하면
Then 반환된 모든 항목의 Principle이 Tested이다
And 최소 5개의 항목이 반환된다

When GetChecklistForPrinciple(Secured) 함수를 호출하면
Then 반환된 모든 항목의 Principle이 Secured이다
And OWASP 관련 항목이 포함된다
```

### AC-TRUST-004: 워크플로우 단계별 체크리스트 필터링

```gherkin
Given TRUST 5 모듈이 로드되었을 때
When GetChecklistForPhase(PhasePlan) 함수를 호출하면
Then 반환된 모든 항목의 Phases에 PhasePlan이 포함된다
And plan 단계 비해당 항목은 포함되지 않는다

When GetChecklistForPhase(PhaseRun) 함수를 호출하면
Then 반환된 항목 수가 PhasePlan보다 많다 (run 단계가 가장 많은 검증 항목)

When GetChecklistForPhase(PhaseSync) 함수를 호출하면
Then 문서화 및 추적 관련 항목이 포함된다
```

### AC-TRUST-005: 원칙 점수 계산

```gherkin
Given 다음 ItemResult 목록이 주어졌을 때:
  - T-001: Passed
  - T-002: Passed
  - T-003: Failed
  - T-004: Passed
  - T-005: Passed
When CalculatePrincipleScore(results) 함수를 호출하면
Then TotalItems이 5이다
And PassedItems이 4이다
And FailedItems이 1이다
And Score가 80.0이다
And Grade가 "B"이다
```

### AC-TRUST-006: 전체 TRUST 5 점수 계산

```gherkin
Given 5가지 원칙의 PrincipleResult가 다음과 같을 때:
  - Tested: Score 90 (Weight 0.2)
  - Readable: Score 85 (Weight 0.2)
  - Unified: Score 100 (Weight 0.2)
  - Secured: Score 75 (Weight 0.2)
  - Trackable: Score 80 (Weight 0.2)
When CalculateTRUST5Score(principleResults) 함수를 호출하면
Then OverallScore가 86.0이다 (가중 평균)
And OverallGrade가 "B"이다
And Principles 맵에 5가지 원칙 결과가 모두 포함된다
```

### AC-TRUST-007: 등급 경계값 테스트

```gherkin
Given 점수별 등급 매핑이 다음과 같을 때:
When 점수가 90이면 Then 등급은 "A"이다
When 점수가 89이면 Then 등급은 "B"이다
When 점수가 80이면 Then 등급은 "B"이다
When 점수가 79이면 Then 등급은 "C"이다
When 점수가 70이면 Then 등급은 "C"이다
When 점수가 69이면 Then 등급은 "D"이다
When 점수가 60이면 Then 등급은 "D"이다
When 점수가 59이면 Then 등급은 "F"이다
When 점수가 0이면 Then 등급은 "F"이다
When 점수가 100이면 Then 등급은 "A"이다
```

### AC-TRUST-008: 빈 결과 처리

```gherkin
Given 빈 ItemResult 목록이 주어졌을 때
When CalculatePrincipleScore([]ItemResult{}) 함수를 호출하면
Then TotalItems이 0이다
And Score가 0.0이다
And Grade가 "F"이다
And 프로그램이 패닉하지 않는다 (0 나누기 방지)
```

---

## 5. 공통 수용 기준

### AC-COMMON-001: JSON 직렬화 호환성

```gherkin
Given 모든 exported 구조체에 대해
When json.Marshal()을 호출하면
Then 유효한 JSON이 생성된다
And json.Unmarshal()로 원본 구조체를 완전히 복원할 수 있다
And JSON 키는 snake_case를 따른다
```

### AC-COMMON-002: 성능 요구사항

```gherkin
Given 벤치마크 테스트 환경에서
When GetAllLanguages()를 1000번 호출하면
Then 평균 실행 시간이 1ms 미만이다

When GetLanguageByExtension(".go")를 1000번 호출하면
Then 평균 실행 시간이 100ns 미만이다

When ValidateEARSRequirement("When X, the system shall Y")를 1000번 호출하면
Then 평균 실행 시간이 5ms 미만이다
```

### AC-COMMON-003: 메모리 사용량

```gherkin
Given 모든 foundation 패키지가 초기화된 상태에서
When runtime.MemStats를 측정하면
Then 총 할당 메모리가 1MB 미만이다
```

### AC-COMMON-004: 패키지 의존성 제약

```gherkin
Given internal/foundation/ 패키지에 대해
When go list -m all로 외부 의존성을 확인하면
Then 외부 모듈 의존성이 0개이다 (Go 표준 라이브러리만 사용)

When go list -deps ./internal/foundation/...으로 내부 의존성을 확인하면
Then internal/ 하위의 다른 패키지에 대한 의존성이 없다
And 순환 의존성이 없다
```

### AC-COMMON-005: godoc 문서화

```gherkin
Given internal/foundation/ 패키지의 모든 exported 함수에 대해
When godoc을 생성하면
Then 모든 exported 함수에 godoc 주석이 존재한다
And 모든 exported 타입에 godoc 주석이 존재한다
And 패키지 레벨 doc.go 파일이 존재한다
```

### AC-COMMON-006: 린트 및 포맷팅

```gherkin
Given internal/foundation/ 패키지에 대해
When golangci-lint run을 실행하면
Then 린트 오류가 0개이다

When gofumpt -l로 포맷팅을 검사하면
Then 포맷팅 위반 파일이 0개이다
```

---

## 6. 에지 케이스 및 오류 처리

### AC-EDGE-001: nil 입력 안전성

```gherkin
Given 모든 조회 함수에 대해
When nil 또는 빈 문자열 입력이 주어지면
Then 프로그램이 패닉하지 않는다
And 명확한 오류 메시지가 반환된다
```

### AC-EDGE-002: 동시 접근 안전성

```gherkin
Given 여러 고루틴에서 동시에
When GetAllLanguages(), GetAllPatterns(), GetChecklist() 함수를 호출하면
Then 데이터 경합(data race)이 발생하지 않는다
And -race 플래그로 테스트 시 경고가 없다
```

### AC-EDGE-003: 확장자 경계값

```gherkin
Given 언어 탐지 함수에 대해
When 점(.)이 없는 확장자 "go"가 주어지면
Then nil과 오류가 반환된다

When 점만 있는 확장자 "."가 주어지면
Then nil과 오류가 반환된다

When 다중 점 확장자 ".test.go"가 주어지면
Then nil과 오류가 반환된다 (또는 ".go"로 정규화하여 Go로 매핑)
```

---

## 7. Definition of Done

이 SPEC은 다음 조건이 모두 충족될 때 완료로 간주한다:

- [ ] 모든 AC(수용 기준) 테스트가 통과
- [ ] 단위 테스트 커버리지 95% 이상
- [ ] 벤치마크 테스트가 성능 목표를 충족
- [ ] `go test -race ./internal/foundation/...` 통과
- [ ] `golangci-lint run ./internal/foundation/...` 오류 0개
- [ ] `gofumpt` 포맷팅 준수
- [ ] 모든 exported 심볼에 godoc 주석 작성
- [ ] 패키지 doc.go 파일 작성
- [ ] 외부 의존성 0개 확인
- [ ] 18개 언어 전수 테스트 통과
- [ ] EARS 5+1 패턴 전수 테스트 통과
- [ ] TRUST 5 점수 계산 정확성 검증
- [ ] JSON 직렬화/역직렬화 왕복 테스트 통과
