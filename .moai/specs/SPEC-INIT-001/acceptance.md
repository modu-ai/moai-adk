# SPEC-INIT-001: 인수 기준

---
spec_id: SPEC-INIT-001
title: Project Initialization & Detection - Acceptance Criteria
status: Planned
tags: [init, project, detection, wizard, bubbletea, acceptance]
---

## 1. 인수 테스트 시나리오

### ACC-001: 새 프로젝트 인터랙티브 초기화

**관련 요구사항**: REQ-E-001, REQ-E-014, REQ-E-015, REQ-E-020

```gherkin
Feature: 새 프로젝트 인터랙티브 초기화

  Scenario: TTY 환경에서 moai init 실행 시 위저드를 통해 프로젝트를 초기화한다
    Given 빈 디렉토리 "/tmp/test-project" 가 존재한다
    And TTY가 연결되어 있다
    And .moai/ 디렉토리가 존재하지 않는다
    When 사용자가 "moai init" 을 실행한다
    Then bubbletea 인터랙티브 위저드가 시작된다
    And 위저드는 다음 항목을 순서대로 질문한다:
      | 항목 | 타입 |
      | 프로젝트 이름 | 텍스트 입력 |
      | 프로그래밍 언어 | 단일 선택 |
      | 프레임워크 | 단일 선택 |
      | 기능 선택 | 다중 선택 |
      | 사용자 이름 | 텍스트 입력 |
      | 대화 언어 | 단일 선택 |
    When 사용자가 모든 항목을 선택한다
    Then 다음 디렉토리가 생성된다:
      | 경로 |
      | .moai/config/sections/ |
      | .moai/specs/ |
      | .moai/reports/ |
      | .moai/memory/ |
      | .moai/logs/ |
      | .claude/agents/moai/ |
      | .claude/skills/ |
      | .claude/commands/moai/ |
      | .claude/rules/moai/ |
    And 다음 파일이 생성된다:
      | 파일 |
      | .moai/config/sections/user.yaml |
      | .moai/config/sections/language.yaml |
      | .moai/config/sections/quality.yaml |
      | .moai/config/sections/workflow.yaml |
      | .moai/manifest.json |
      | .claude/settings.json |
      | CLAUDE.md |
    And settings.json은 json.Valid()로 검증 가능하다
    And 생성된 모든 파일 경로는 filepath.Clean()으로 정규화되어 있다
    And 터미널에 생성된 파일 목록이 출력된다
    And "moai doctor" 또는 "moai status" 실행을 안내하는 메시지가 출력된다
```

---

### ACC-002: 기존 프로젝트 감지 및 재초기화 방지

**관련 요구사항**: REQ-E-016, REQ-N-001, REQ-S-002

```gherkin
Feature: 기존 프로젝트 감지 시 안전 처리

  Scenario: .moai/ 디렉토리가 이미 존재할 때 초기화를 거부한다
    Given 디렉토리 "/tmp/existing-project" 에 .moai/ 가 이미 존재한다
    And .moai/config/sections/user.yaml 파일이 존재한다
    When 사용자가 "moai init" 을 실행한다
    Then 오류 메시지 "project already initialized" 가 반환된다
    And "--force 플래그를 사용하여 재초기화하세요" 안내가 출력된다
    And 기존 .moai/ 디렉토리는 변경되지 않는다
```

---

### ACC-003: 언어 자동 감지

**관련 요구사항**: REQ-E-004, REQ-E-005, REQ-E-006, REQ-E-007, REQ-E-019

```gherkin
Feature: 프로젝트 루트에서 프로그래밍 언어를 자동 감지한다

  Scenario Outline: 매핑 파일 기반 언어 감지
    Given 프로젝트 루트에 <감지_파일> 이 존재한다
    When Detector.DetectLanguages(root) 를 호출한다
    Then 반환된 Language 목록에 <언어> 가 포함된다
    And 해당 Language의 Confidence 가 0.0 보다 크다
    And 해당 Language의 FileCount 가 0 보다 크다

    Examples:
      | 감지_파일 | 언어 |
      | package.json | JavaScript |
      | go.mod | Go |
      | pyproject.toml | Python |
      | requirements.txt | Python |
      | Cargo.toml | Rust |
      | pom.xml | Java |
      | build.gradle | Java |
      | Gemfile | Ruby |
      | composer.json | PHP |
      | pubspec.yaml | Dart |
      | mix.exs | Elixir |
      | build.sbt | Scala |
      | Package.swift | Swift |

  Scenario: 다중 언어 프로젝트에서 신뢰도 순 정렬
    Given 프로젝트 루트에 go.mod 와 package.json 이 모두 존재한다
    And Go 파일이 20개, TypeScript 파일이 5개 존재한다
    When Detector.DetectLanguages(root) 를 호출한다
    Then Go의 Confidence 가 TypeScript의 Confidence 보다 높다
    And 결과 목록은 Confidence 내림차순으로 정렬되어 있다

  Scenario: 빈 프로젝트에서 언어 미감지
    Given 프로젝트 루트에 인식 가능한 언어 파일이 없다
    When Detector.DetectLanguages(root) 를 호출한다
    Then 빈 Language 슬라이스가 반환된다
    And 오류가 반환되지 않는다
```

---

### ACC-004: 비인터랙티브 모드 초기화

**관련 요구사항**: REQ-E-002, REQ-E-003, REQ-S-001, REQ-S-003, REQ-N-004

```gherkin
Feature: CI/CD 환경을 위한 비인터랙티브 초기화

  Scenario: --non-interactive 플래그로 자동 초기화
    Given 빈 디렉토리 "/tmp/ci-project" 가 존재한다
    And 프로젝트 루트에 go.mod 가 존재한다
    When 사용자가 "moai init --non-interactive" 를 실행한다
    Then 사용자 입력 대기 없이 즉시 초기화가 완료된다
    And ProjectName은 현재 디렉토리 이름 "ci-project" 으로 설정된다
    And Language는 감지된 "Go" 로 설정된다
    And .moai/ 디렉토리 구조가 정상 생성된다

  Scenario: -y 축약 플래그 동작
    Given 빈 디렉토리가 존재한다
    When 사용자가 "moai init -y" 를 실행한다
    Then --non-interactive 와 동일하게 동작한다
    And 사용자 입력 대기가 발생하지 않는다

  Scenario: TTY 미연결 시 자동 전환
    Given stdin이 터미널에 연결되지 않았다 (파이프 또는 CI 환경)
    When 사용자가 "moai init" 을 실행한다
    Then 자동으로 비인터랙티브 모드로 전환된다
    And 기본값으로 초기화가 진행된다

  Scenario: 비인터랙티브 모드에서 플래그로 값 지정
    Given 빈 디렉토리가 존재한다
    When 사용자가 "moai init -y --name my-app --language typescript" 를 실행한다
    Then ProjectName이 "my-app" 으로 설정된다
    And Language가 "typescript" 로 설정된다
    And 나머지 필드는 기본값으로 설정된다
```

---

### ACC-005: 프레임워크 감지

**관련 요구사항**: REQ-E-008 ~ REQ-E-013

```gherkin
Feature: 설정 파일에서 프레임워크를 자동 감지한다

  Scenario Outline: package.json 의존성 기반 프레임워크 감지
    Given package.json에 <의존성> 이 dependencies 또는 devDependencies에 포함되어 있다
    When Detector.DetectFrameworks(root) 를 호출한다
    Then 반환된 Framework 목록에 Name이 <프레임워크> 인 항목이 포함된다
    And 해당 항목의 Version이 빈 문자열이 아니다
    And 해당 항목의 ConfigFile이 "package.json" 이다

    Examples:
      | 의존성 | 프레임워크 |
      | react | React |
      | next | Next.js |
      | vue | Vue |
      | @angular/core | Angular |
      | svelte | Svelte |

  Scenario Outline: Python 프레임워크 감지
    Given pyproject.toml에 <의존성> 이 project.dependencies에 포함되어 있다
    When Detector.DetectFrameworks(root) 를 호출한다
    Then 반환된 Framework 목록에 Name이 <프레임워크> 인 항목이 포함된다

    Examples:
      | 의존성 | 프레임워크 |
      | fastapi | FastAPI |
      | django | Django |
      | flask | Flask |

  Scenario: Go 프레임워크 감지 (go.mod 또는 import 기반)
    Given go.mod에 "github.com/gin-gonic/gin" 의존성이 존재한다
    When Detector.DetectFrameworks(root) 를 호출한다
    Then 반환된 Framework 목록에 Name이 "Gin" 인 항목이 포함된다

  Scenario: 프레임워크 미감지 프로젝트
    Given 프로젝트에 인식 가능한 프레임워크 의존성이 없다
    When Detector.DetectFrameworks(root) 를 호출한다
    Then 빈 Framework 슬라이스가 반환된다
    And 오류가 반환되지 않는다
```

---

### ACC-006: Force 재초기화

**관련 요구사항**: REQ-E-017

```gherkin
Feature: --force 플래그를 사용한 재초기화

  Scenario: 기존 프로젝트를 --force 로 재초기화
    Given 디렉토리에 .moai/ 가 이미 존재한다
    And .moai/config/sections/user.yaml 에 사용자 설정이 저장되어 있다
    When 사용자가 "moai init --force" 를 실행한다
    Then 기존 .moai/ 디렉토리가 .moai.backup.{timestamp}/ 으로 백업된다
    And 새로운 .moai/ 디렉토리 구조가 생성된다
    And 초기화가 정상 완료된다
    And 백업 경로가 출력 메시지에 포함된다

  Scenario: --force 와 --non-interactive 결합
    Given 디렉토리에 .moai/ 가 이미 존재한다
    When 사용자가 "moai init --force -y" 를 실행한다
    Then 사용자 확인 없이 백업 후 재초기화가 수행된다
    And 사용자 입력 대기가 발생하지 않는다
```

---

### ACC-007: Git 미초기화 프로젝트

**관련 요구사항**: REQ-E-018

```gherkin
Feature: Git 저장소가 없는 프로젝트 처리

  Scenario: Git 미초기화 디렉토리에서 초기화
    Given 디렉토리에 .git/ 이 존재하지 않는다
    When 사용자가 "moai init -y" 를 실행한다
    Then 경고 메시지 "Git repository not detected" 가 출력된다
    And Git 관련 기능을 건너뛰고 초기화가 정상 완료된다
    And .moai/ 디렉토리 구조가 생성된다
```

---

### ACC-008: 상태 기반 기본값 적용

**관련 요구사항**: REQ-S-003, REQ-S-004, REQ-O-001

```gherkin
Feature: 감지 결과를 기본값으로 활용

  Scenario: 감지된 언어를 위저드 기본값으로 제안
    Given 프로젝트 루트에 pyproject.toml이 존재한다
    And Python 파일이 15개 존재한다
    When 인터랙티브 위저드의 언어 선택 단계가 표시된다
    Then "Python" 이 기본 선택으로 강조된다

  Scenario: 프레임워크 미감지 시 "None" 기본 선택
    Given 프로젝트에 인식 가능한 프레임워크가 없다
    When 인터랙티브 위저드의 프레임워크 선택 단계가 표시된다
    Then "None" 이 기본 선택으로 제공된다

  Scenario: 비인터랙티브에서 프로젝트 이름 기본값
    Given 프로젝트 디렉토리 이름이 "awesome-project" 이다
    And 사용자가 --name 플래그를 지정하지 않았다
    When "moai init -y" 를 실행한다
    Then ProjectName이 "awesome-project" 으로 설정된다
```

---

### ACC-009: 금지 동작 검증

**관련 요구사항**: REQ-N-001 ~ REQ-N-005

```gherkin
Feature: 금지된 동작이 발생하지 않음을 검증

  Scenario: 확인 없이 기존 프로젝트 덮어쓰기 방지
    Given .moai/ 디렉토리가 이미 존재한다
    When "moai init" 을 실행한다 (--force 없이)
    Then 기존 .moai/ 의 어떤 파일도 수정되지 않는다
    And 새로운 파일이 .moai/ 에 추가되지 않는다

  Scenario: 생성된 JSON에 동적 토큰 미포함
    Given 초기화가 정상 완료된다
    When settings.json 파일 내용을 검사한다
    Then "$" 로 시작하는 변수 참조가 포함되지 않는다
    And "{{" 또는 "}}" 템플릿 토큰이 포함되지 않는다
    And "${" 쉘 변수 구문이 포함되지 않는다

  Scenario: 생성된 YAML에 문자열 연결 미사용 검증
    Given 초기화가 정상 완료된다
    When .moai/config/sections/ 아래 YAML 파일을 파싱한다
    Then 모든 파일이 yaml.v3 로 정상 파싱된다
    And 파싱 오류가 발생하지 않는다

  Scenario: 비인터랙티브 모드에서 stdin 대기 없음
    Given stdin이 /dev/null 로 리다이렉트되어 있다
    When "moai init -y" 를 실행한다
    Then 프로세스가 5초 이내에 종료된다
    And 행(hang) 걸림이 발생하지 않는다

  Scenario: 복구 가능한 오류에서 panic 미발생
    Given 읽기 권한이 없는 디렉토리에서 실행한다
    When "moai init -y" 를 실행한다
    Then panic이 발생하지 않는다
    And 적절한 오류 메시지가 stderr에 출력된다
    And 0이 아닌 종료 코드가 반환된다
```

---

### ACC-010: 선택적 기능 검증

**관련 요구사항**: REQ-O-001, REQ-O-002, REQ-O-003

```gherkin
Feature: 선택적 기능 동작 검증

  Scenario: 초기화 진행률 표시
    Given TTY가 연결되어 있다
    When 인터랙티브 초기화가 진행된다
    Then 각 초기화 단계마다 진행 상태가 표시된다
    And 스피너 또는 프로그레스 바가 사용된다

  Scenario: .gitignore 자동 생성
    Given 프로젝트 루트에 .gitignore 파일이 없다
    When 초기화가 완료된다
    Then .gitignore 파일이 생성된다
    And 파일 내용에 .moai/ 관련 제외 패턴이 포함된다
```

---

### ACC-011: Brownfield 프로젝트에서 DDD 방법론 자동 감지

**관련 요구사항**: REQ-E-021, REQ-E-022

```gherkin
Feature: 테스트가 없는 brownfield 프로젝트에서 DDD 방법론을 추천한다

  Scenario: 테스트 파일이 없는 Go 프로젝트에서 DDD 추천
    Given 프로젝트에 Go 소스 파일이 50개 존재한다
    And 테스트 파일(*_test.go)이 0개 존재한다
    When MethodologyDetector.DetectMethodology(root, languages) 를 호출한다
    Then Recommended가 "ddd" 이다
    And Confidence가 0.8 보다 크다
    And Rationale에 "no existing tests" 또는 "테스트 없음" 이 포함된다
    And ProjectType이 "brownfield" 이다
    And Alternatives에 TDD 모드에 대한 Warning이 포함된다
    And Warning에 예상 테스트 수가 포함된다
```

---

### ACC-012: 높은 커버리지 프로젝트에서 TDD 방법론 자동 감지

**관련 요구사항**: REQ-E-021, REQ-E-022

```gherkin
Feature: 높은 테스트 커버리지의 프로젝트에서 TDD 방법론을 추천한다

  Scenario: 테스트 파일이 풍부한 프로젝트에서 TDD 추천
    Given 프로젝트에 소스 파일이 30개 존재한다
    And 테스트 파일이 25개 존재한다
    When MethodologyDetector.DetectMethodology(root, languages) 를 호출한다
    Then Recommended가 "tdd" 이다
    And CoverageEstimate가 50 보다 크다
    And ProjectType이 "brownfield" 이다
```

---

### ACC-013: 부분적 테스트 프로젝트에서 Hybrid 방법론 자동 감지

**관련 요구사항**: REQ-E-021, REQ-E-022

```gherkin
Feature: 부분적으로 테스트된 프로젝트에서 Hybrid 방법론을 추천한다

  Scenario: 일부 테스트가 있는 프로젝트에서 Hybrid 추천
    Given 프로젝트에 소스 파일이 100개 존재한다
    And 테스트 파일이 15개 존재한다
    When MethodologyDetector.DetectMethodology(root, languages) 를 호출한다
    Then Recommended가 "hybrid" 이다
    And CoverageEstimate가 10 이상이고 50 미만이다
    And ProjectType이 "brownfield" 이다

  Scenario: 빈 프로젝트(greenfield)에서 Hybrid 추천
    Given 프로젝트에 소스 파일이 0개 존재한다
    And 테스트 파일이 0개 존재한다
    When MethodologyDetector.DetectMethodology(root, languages) 를 호출한다
    Then Recommended가 "hybrid" 이다
    And ProjectType이 "greenfield" 이다
    And Confidence가 0.7 이다
```

---

### ACC-014: 사용자가 추천과 다른 방법론을 선택할 때 경고 표시

**관련 요구사항**: REQ-E-023

```gherkin
Feature: 추천 방법론과 다른 선택 시 경고를 표시한다

  Scenario: DDD 추천 프로젝트에서 사용자가 TDD를 선택하면 경고 표시
    Given MethodologyDetector가 "ddd" 를 추천한다
    And 프로젝트에 소스 파일이 50개, 테스트 파일이 0개 존재한다
    When 사용자가 위저드에서 "tdd" 를 선택한다
    Then 예상 노력에 대한 경고 메시지가 표시된다
    And 경고에 "250개 테스트 필요" 와 같은 구체적 수치가 포함된다
    And 사용자에게 선택 확인을 요청한다
    When 사용자가 확인한다
    Then InitOptions.DevelopmentMode 가 "tdd" 로 설정된다
```

---

### ACC-015: 비인터랙티브 모드에서 자동 감지 방법론 사용

**관련 요구사항**: REQ-S-005

```gherkin
Feature: 비인터랙티브 모드에서 자동 감지된 방법론을 사용한다

  Scenario: --development-mode 플래그 없이 비인터랙티브 초기화
    Given 프로젝트에 Go 소스 파일이 30개, 테스트 파일이 2개 존재한다
    And --development-mode 플래그가 제공되지 않았다
    And 비인터랙티브 모드가 활성화되어 있다
    When 초기화가 실행된다
    Then MethodologyDetector가 자동으로 실행된다
    And 자동 감지된 추천 방법론이 사용된다
    And 사용자 프롬프트가 표시되지 않는다
    And InitOptions.DevelopmentMode 가 자동 감지 결과로 설정된다
```

---

### ACC-016: --development-mode 플래그로 자동 감지 우회

**관련 요구사항**: REQ-S-006

```gherkin
Feature: --development-mode 플래그가 자동 감지를 우회한다

  Scenario: --development-mode=hybrid 플래그로 초기화
    Given --development-mode=hybrid 플래그가 제공되었다
    When 초기화가 실행된다
    Then DevelopmentMode 가 "hybrid" 로 설정된다
    And MethodologyDetector 자동 감지가 실행되지 않는다

  Scenario: 인터랙티브 모드에서도 플래그 우선
    Given --development-mode=tdd 플래그가 제공되었다
    And TTY가 연결되어 있다
    When "moai init --development-mode=tdd" 를 실행한다
    Then 위저드에서 방법론 선택 단계가 건너뛰어진다
    And DevelopmentMode 가 "tdd" 로 설정된다
```

---

## 2. 품질 게이트 기준

### 2.1 TRUST 5 검증 체크리스트

| 원칙 | 기준 | 검증 방법 |
|------|------|-----------|
| **Tested** | 단위 테스트 커버리지 90% 이상 (core domain) | `go test -coverprofile` |
| **Tested** | 통합 테스트로 전체 초기화 흐름 검증 | E2E 테스트 |
| **Tested** | 벤치마크로 성능 목표 달성 확인 | `go test -bench` |
| **Readable** | 모든 exported 함수에 godoc 주석 | `golangci-lint` |
| **Readable** | Go 명명 규칙 준수 | `golangci-lint` |
| **Unified** | gofumpt 포매팅 통과 | `golangci-lint` |
| **Unified** | golangci-lint 경고 0개 | CI 파이프라인 |
| **Secured** | 경로 순회 방지 (filepath.Clean + 기본 디렉토리 검증) | 보안 테스트 |
| **Secured** | 입력 검증 (프로젝트 이름, 경로) | 단위 테스트 |
| **Trackable** | Conventional Commit 메시지 | Git hook |

### 2.2 성능 기준

| 메트릭 | 목표값 | 허용 범위 |
|--------|--------|-----------|
| 언어 감지 | < 100ms | 200ms 미만 |
| 프레임워크 감지 | < 200ms | 500ms 미만 |
| 방법론 감지 (Tier 1+2) | < 500ms | 1s 미만 |
| 방법론 감지 (Tier 3) | < 30s | 60s 미만 |
| 전체 초기화 | < 3s | 5s 미만 |
| 메모리 사용량 | < 50MB | 100MB 미만 |

### 2.3 호환성 기준

| 항목 | 기준 |
|------|------|
| 기존 .moai/ 구조 호환 | Python MoAI-ADK가 생성한 .moai/ 디렉토리를 감지 가능 |
| YAML 설정 호환 | 기존 config/sections/ YAML 형식과 동일한 스키마 |
| 크로스 플랫폼 | macOS, Linux, Windows에서 동일하게 동작 |

---

## 3. Definition of Done

- [ ] 모든 인수 테스트 시나리오 (ACC-001 ~ ACC-016) 통과
- [ ] 단위 테스트 커버리지 90% 이상 (internal/core/project/)
- [ ] 벤치마크 테스트로 성능 목표 달성 확인
- [ ] golangci-lint 경고 0개
- [ ] 모든 exported 타입/함수에 godoc 주석 완성
- [ ] TRUST 5 품질 게이트 전체 통과
- [ ] `moai init` CLI 명령이 end-to-end로 동작
- [ ] `moai init -y` 비인터랙티브 모드 동작
- [ ] `moai init --force` 재초기화 동작
- [ ] `moai init --development-mode` 플래그 동작
- [ ] MethodologyDetector가 Tier 1+2 감지를 500ms 이내에 완료
- [ ] 방법론 추천이 greenfield/brownfield 프로젝트를 올바르게 분류
- [ ] 추천과 다른 선택 시 경고 메시지 표시 확인
- [ ] 생성된 모든 JSON 파일이 `json.Valid()` 통과
- [ ] 생성된 모든 파일 경로가 `filepath.Clean()` 정규화 확인
- [ ] 생성된 파일에 미확장 동적 토큰 없음 확인
- [ ] quality.yaml에 선택된 development_mode가 올바르게 반영
