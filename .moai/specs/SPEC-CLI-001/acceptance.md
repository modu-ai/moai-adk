# SPEC-CLI-001: Acceptance Criteria

---
spec_id: SPEC-CLI-001
title: CLI Command Composition & Integration - Acceptance Criteria
phase: "Phase 5 - CLI (Final, composition root)"
status: Planned
priority: Medium
tags: [cli, composition-root, di-wiring, integration, migration, cobra, pkg]
---

## 1. Command Routing (커맨드 라우팅)

### AC-010: Root 커맨드 기본 동작

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 인자 없이 "moai"를 실행하면
Then 시스템은 도움말 텍스트를 출력하고
And 사용 가능한 서브커맨드 목록을 표시하고
And exit code 0을 반환한다
```

### AC-011: Version 커맨드 (완료)

```gherkin
Given moai 바이너리가 ldflags로 버전 정보가 주입되어 빌드되어 있고
When 사용자가 "moai version"을 실행하면
Then 시스템은 "moai-adk {version} (commit: {hash}, built: {date})" 형식으로 출력하고
And exit code 0을 반환한다
```

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 "moai --version"을 실행하면
Then 시스템은 버전 정보를 출력하고
And exit code 0을 반환한다
```

### AC-012: Init 커맨드 라우팅

```gherkin
Given 프로젝트 디렉토리가 존재하고
And SPEC-INIT-001 도메인 모듈이 DI로 주입되어 있고
When 사용자가 "moai init"을 실행하면
Then 시스템은 프로젝트 초기화 로직을 SPEC-INIT-001 모듈에 위임하고
And 초기화 결과를 출력한다
```

```gherkin
Given 프로젝트 디렉토리가 존재하고
When 사용자가 "moai init --non-interactive --locale ko"를 실행하면
Then 시스템은 non-interactive 모드로 한국어 설정으로 초기화를 수행한다
```

```gherkin
Given 프로젝트 디렉토리가 존재하고
When 사용자가 "moai init --help"를 실행하면
Then 시스템은 다음 flag를 포함한 도움말을 출력한다:
  | Flag | 설명 |
  | --non-interactive, -y | Non-interactive mode |
  | --mode | Project mode (personal/team) |
  | --locale | Language setting |
  | --language | Programming language |
  | --force | Force re-initialization |
```

### AC-013: Doctor 커맨드 라우팅

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
When 사용자가 "moai doctor"를 실행하면
Then 시스템은 시스템 진단을 수행하고
And Git, Go, Claude Code, config 무결성 등의 검사 결과를 출력하고
And 각 항목에 대해 성공/실패/경고 상태를 표시한다
```

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
When 사용자가 "moai doctor --verbose"를 실행하면
Then 시스템은 상세 진단 정보(도구 버전, 경로 등)를 추가로 출력한다
```

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
When 사용자가 "moai doctor --export /tmp/diag.json"을 실행하면
Then 시스템은 진단 결과를 JSON 형식으로 지정된 경로에 저장한다
```

### AC-014: Status 커맨드

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
And .moai/config/sections/ 에 유효한 설정 파일이 존재하고
When 사용자가 "moai status"를 실행하면
Then 시스템은 프로젝트 이름, 타입, MoAI 버전, config 상태를 출력하고
And SPEC 진행 상황 요약을 표시한다
```

### AC-015: Update 커맨드 라우팅

```gherkin
Given moai 바이너리가 실행 중이고
And SPEC-UPDATE-001 도메인 모듈이 DI로 주입되어 있고
When 사용자가 "moai update"를 실행하면
Then 시스템은 self-update 워크플로우를 SPEC-UPDATE-001 모듈에 위임한다
```

```gherkin
Given moai 바이너리가 실행 중이고
When 사용자가 "moai update --check"를 실행하면
Then 시스템은 최신 버전 확인만 수행하고 업데이트는 실행하지 않는다
```

### AC-016: Hook 커맨드 라우팅

```gherkin
Given moai 바이너리가 실행 중이고
And SPEC-HOOK-001 도메인 모듈이 DI로 주입되어 있고
When Claude Code가 "moai hook session-start"를 stdin JSON과 함께 실행하면
Then 시스템은 session-start 핸들러를 SPEC-HOOK-001 모듈에 위임하고
And JSON 결과를 stdout으로 출력하고
And exit code 0(허용) 또는 2(차단)를 반환한다
```

```gherkin
Given moai 바이너리가 실행 중이고
When 사용자가 "moai hook list"를 실행하면
Then 시스템은 등록된 모든 hook 이벤트와 핸들러 목록을 출력한다
```

```gherkin
Given moai 바이너리가 실행 중이고
When 사용자가 "moai hook --help"를 실행하면
Then 시스템은 사용 가능한 hook 서브커맨드 목록을 출력한다:
  | 서브커맨드 | 설명 |
  | session-start | 세션 시작 핸들러 |
  | pre-tool | 도구 실행 전 핸들러 |
  | post-tool | 도구 실행 후 핸들러 |
  | session-end | 세션 종료 핸들러 |
  | stop | 중지 핸들러 |
  | compact | 컨텍스트 압축 핸들러 |
  | list | 등록된 hook 목록 |
```

### AC-017: CC 커맨드 (top-level)

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
And .moai/config/sections/ 에 LLM 설정이 존재하고
When 사용자가 "moai cc"를 실행하면
Then 시스템은 Claude 백엔드로 설정을 전환하고
And 전환 성공 메시지를 출력한다
```

```gherkin
Given moai 바이너리가 빌드되어 있고
When "moai --help" 출력을 확인하면
Then "cc" 커맨드가 top-level 커맨드 목록에 표시되고
And "switch" 커맨드는 존재하지 않는다
```

### AC-018: GLM 커맨드 (top-level)

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
When 사용자가 "moai glm sk-abc123"을 실행하면
Then 시스템은 GLM 백엔드로 전환하고
And API 키 "sk-abc123"을 안전하게 저장한다
```

```gherkin
Given MoAI 프로젝트가 초기화되어 있고
When 사용자가 "moai glm"을 인자 없이 실행하면
Then 시스템은 기존 저장된 API 키로 GLM 백엔드로 전환한다
```

```gherkin
Given moai 바이너리가 빌드되어 있고
When "moai --help" 출력을 확인하면
Then "glm" 커맨드가 top-level 커맨드 목록에 표시되고
And "switch" 커맨드는 존재하지 않는다
```

### AC-019: Rank 커맨드

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 "moai rank --help"를 실행하면
Then 시스템은 7개 서브커맨드를 표시한다:
  | 서브커맨드 | 설명 |
  | login | MoAI Cloud 인증 |
  | status | 랭킹 상태 표시 |
  | logout | 로그아웃 |
  | sync | 메트릭 동기화 |
  | exclude | 제외 패턴 추가 |
  | include | 포함 패턴 추가 |
  | register | 조직 등록 |
```

```gherkin
Given 사용자가 MoAI Cloud에 인증되어 있고
When 사용자가 "moai rank status"를 실행하면
Then 시스템은 SPEC-TOOL-001 랭킹 모듈을 통해 현재 랭킹 상태를 출력한다
```

### AC-020: Worktree 커맨드

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 "moai worktree --help"를 실행하면
Then 시스템은 6개 서브커맨드를 표시한다:
  | 서브커맨드 | 설명 |
  | new | 새 worktree 생성 |
  | list | worktree 목록 |
  | switch | worktree 전환 |
  | sync | worktree 동기화 |
  | remove | worktree 제거 |
  | clean | worktree 정리 |
```

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 "moai wt list"를 실행하면
Then 시스템은 "moai worktree list"와 동일한 결과를 출력한다
And worktree alias가 정상 동작한다
```

```gherkin
Given Git 리포지토리가 초기화되어 있고
And SPEC-GIT-001 도메인 모듈이 DI로 주입되어 있고
When 사용자가 "moai worktree new SPEC-AUTH-001"을 실행하면
Then 시스템은 SPEC-GIT-001 모듈을 통해 새 worktree를 생성한다
```

---

## 2. DI Wiring (의존성 주입 구성)

### AC-001: Composition Root 의존성 구성

```gherkin
Given moai 바이너리가 빌드되어 있고
And 모든 선행 SPEC 도메인 모듈이 구현되어 있고
When moai 프로세스가 시작되면
Then Composition Root에서 다음 의존성이 순서대로 초기화된다:
  | 순서 | 모듈 | 설명 |
  | 1 | Logger | slog 기반 구조화 로거 |
  | 2 | Config Manager | Viper + typed struct |
  | 3 | Hook Registry | Hook 핸들러 등록 |
  | 4 | Template Deployer | go:embed 기반 |
  | 5 | Manifest Manager | 파일 provenance 추적 |
  | 6 | Git Repository | go-git + system Git |
  | 7 | Version Migrator | config 마이그레이션 |
  | 8 | Integration Engine | 크로스 패키지 검증 |
```

### AC-002: 인터페이스 기반 의존성 주입

```gherkin
Given CLI 커맨드 파일(init.go, doctor.go 등)의 소스코드를 분석하면
When 도메인 서비스 참조를 검사하면
Then Composition Root를 제외한 모든 CLI 파일에서
And 구체 타입(struct) 직접 참조가 0건이고
And 모든 의존성이 인터페이스를 통해 접근된다
```

### AC-003: 의존성 주입 실패 처리

```gherkin
Given moai 바이너리가 실행 중이고
When 의존성 초기화 과정에서 Config 로드가 실패하면
Then 시스템은 의미 있는 에러 메시지를 출력하고
And 필요한 조치 방법(moai init 실행 등)을 안내하고
And exit code 1을 반환한다
```

---

## 3. Version Migration (버전 마이그레이션)

### AC-040: 정상 마이그레이션

```gherkin
Given .moai/config/sections/ 에 v1.0.0 형식의 설정 파일이 존재하고
And 현재 ADK 버전이 v2.0.0이고
When moai 프로세스가 시작되면
Then 시스템은 v1.0.0 -> v2.0.0 마이그레이션이 필요함을 감지하고
And 마이그레이션을 자동으로 실행하고
And 설정 파일을 v2.0.0 형식으로 업데이트한다
```

### AC-041: 백업 생성

```gherkin
Given 마이그레이션이 필요한 설정 파일이 존재하고
When 마이그레이션이 시작되면
Then 시스템은 .moai/backup/{timestamp}/ 디렉토리를 생성하고
And .moai/config/sections/ 의 모든 YAML 파일을 백업하고
And 백업 완료 후 마이그레이션을 진행한다
```

### AC-042: 마이그레이션 실패 시 롤백

```gherkin
Given 마이그레이션이 진행 중이고
And 백업이 .moai/backup/{timestamp}/ 에 생성되어 있고
When 마이그레이션 도중 에러가 발생하면
Then 시스템은 백업에서 원본 설정을 자동으로 복원하고
And 에러 메시지와 함께 마이그레이션 실패를 보고하고
And 원본 설정이 손상되지 않았음을 확인한다
```

### AC-043: 마이그레이션 불필요 시 스킵

```gherkin
Given .moai/config/sections/ 에 현재 ADK 버전과 동일한 형식의 설정이 존재하고
When moai 프로세스가 시작되면
Then 시스템은 마이그레이션이 불필요함을 감지하고
And 마이그레이션을 건너뛰고
And 정상적으로 CLI를 시작한다
```

### AC-044: Python ADK 하위 호환성

```gherkin
Given Python MoAI-ADK에서 생성된 .moai/config/sections/ YAML 파일이 존재하고
When Go MoAI-ADK가 최초 실행되면
Then 시스템은 Python 형식의 설정을 Go 형식으로 마이그레이션하고
And 사용자 설정값(이름, 언어, 품질 설정 등)이 보존된다
```

### AC-045: 백업 정리

```gherkin
Given .moai/backup/ 에 10개 이상의 백업이 존재하고
When 새 마이그레이션이 완료되면
Then 시스템은 최근 5개의 백업만 유지하고
And 오래된 백업을 자동으로 삭제한다
```

---

## 4. Integration Test Engine (통합 테스트 엔진)

### AC-030: 통합 테스트 엔진 실행

```gherkin
Given 통합 테스트 엔진에 테스트 스위트가 등록되어 있고
When "go test -tags=integration ./internal/core/integration/..."을 실행하면
Then 시스템은 등록된 모든 테스트 스위트를 순차적으로 실행하고
And IntegrationReport를 생성한다
```

### AC-031: 통합 보고서 모델

```gherkin
Given 통합 테스트가 완료되면
When IntegrationReport를 검사하면
Then 보고서는 다음 필드를 포함한다:
  | 필드 | 타입 | 설명 |
  | Suites | []SuiteResult | 각 스위트별 결과 |
  | TotalPass | int | 통과한 테스트 수 |
  | TotalFail | int | 실패한 테스트 수 |
  | TotalSkip | int | 건너뛴 테스트 수 |
  | Duration | time.Duration | 총 실행 시간 |
```

### AC-032: Config -> CLI 크로스 패키지 검증

```gherkin
Given 유효한 .moai/config/sections/ 설정이 존재하고
When Config -> CLI 통합 테스트가 실행되면
Then Config Manager가 설정을 로드하고
And CLI 커맨드가 로드된 설정을 사용하여 정상 실행되고
And 테스트 결과가 Pass로 기록된다
```

### AC-033: Hook -> Dispatch 크로스 패키지 검증

```gherkin
Given Hook Registry에 session-start 핸들러가 등록되어 있고
When Hook -> Dispatch 통합 테스트가 실행되면
Then Hook Protocol이 stdin JSON을 파싱하고
And Registry가 적절한 핸들러를 디스패치하고
And 핸들러가 JSON 결과를 반환하고
And 테스트 결과가 Pass로 기록된다
```

### AC-034: Template -> Manifest 크로스 패키지 검증

```gherkin
Given go:embed 템플릿이 바이너리에 번들되어 있고
When Template -> Manifest 통합 테스트가 실행되면
Then Template Deployer가 파일을 배포하고
And Manifest Manager가 배포된 파일을 추적하고
And provenance가 "template_managed"로 기록되고
And 테스트 결과가 Pass로 기록된다
```

### AC-035: Migration -> Config 크로스 패키지 검증

```gherkin
Given 구 버전 형식의 설정 파일이 존재하고
When Migration -> Config 통합 테스트가 실행되면
Then Migrator가 구 형식을 신 형식으로 변환하고
And Config Manager가 변환된 설정을 정상 로드하고
And 테스트 결과가 Pass로 기록된다
```

---

## 5. CLI 전체 Smoke Test (모든 커맨드)

### AC-060: 전체 커맨드 Help 출력

```gherkin
Given moai 바이너리가 빌드되어 있고
When 다음 각 커맨드에 대해 "--help" 를 실행하면
Then 각 커맨드는 사용법, 설명, flag 목록을 포함한 도움말을 출력하고
And exit code 0을 반환한다:
  | 커맨드 |
  | moai |
  | moai version |
  | moai init |
  | moai doctor |
  | moai status |
  | moai update |
  | moai hook |
  | moai cc |
  | moai glm |
  | moai rank |
  | moai worktree |
```

### AC-061: 존재하지 않는 커맨드 에러 처리

```gherkin
Given moai 바이너리가 빌드되어 있고
When 사용자가 "moai nonexistent"를 실행하면
Then 시스템은 "unknown command" 에러 메시지를 출력하고
And 유사한 커맨드를 제안하고
And exit code 1을 반환한다
```

### AC-062: 드롭된 커맨드 부재 확인

```gherkin
Given moai 바이너리가 빌드되어 있고
When "moai --help" 출력을 확인하면
Then "language" 커맨드는 존재하지 않고
And "analyze" 커맨드는 존재하지 않고
And "switch" 커맨드는 존재하지 않는다
```

### AC-063: Top-Level 커맨드 확인

```gherkin
Given moai 바이너리가 빌드되어 있고
When "moai --help" 출력에서 Available Commands 섹션을 확인하면
Then 다음 커맨드가 top-level에 나열된다:
  | 커맨드 | 유형 |
  | version | Leaf |
  | init | Leaf |
  | doctor | Leaf |
  | status | Leaf |
  | update | Leaf |
  | hook | Group |
  | cc | Leaf (top-level) |
  | glm | Leaf (top-level) |
  | rank | Group |
  | worktree | Group |
And "cc"와 "glm"은 별도 top-level 커맨드로 표시되고
And "switch" 커맨드는 존재하지 않는다
```

---

## 6. pkg/ 공개 패키지

### AC-050: version 패키지 API

```gherkin
Given pkg/version/ 패키지가 빌드 타임에 ldflags로 주입되어 있고
When 외부 패키지에서 version.GetVersion()을 호출하면
Then 빌드 시 주입된 버전 문자열을 반환한다
```

```gherkin
Given pkg/version/ 패키지가 ldflags 없이 빌드되었고
When version.GetVersion()을 호출하면
Then 기본값 "dev"를 반환한다
```

### AC-051: models 패키지 직렬화

```gherkin
Given pkg/models/ 패키지의 ProjectConfig 구조체가 있고
When json.Marshal()로 직렬화하면
Then JSON 필드 이름이 snake_case로 출력되고
And yaml.Marshal()로 직렬화하면 YAML 필드 이름이 snake_case로 출력된다
```

```gherkin
Given 유효한 YAML 설정 데이터가 있고
When pkg/models/UserConfig로 yaml.Unmarshal()하면
Then Name 필드가 정확히 매핑된다
```

### AC-052: utils 패키지 기능

```gherkin
Given .moai/ 디렉토리가 존재하는 프로젝트 루트가 있고
And 현재 작업 디렉토리가 프로젝트의 하위 디렉토리이고
When utils.FindProjectRoot()를 호출하면
Then .moai/ 디렉토리를 포함하는 프로젝트 루트 경로를 반환한다
```

```gherkin
Given .moai/ 디렉토리가 존재하지 않는 디렉토리에서
When utils.FindProjectRoot()를 호출하면
Then os.ErrNotExist 에러를 반환한다
```

```gherkin
Given MOAI_LOG_LEVEL 환경변수가 "debug"로 설정되어 있고
When utils.InitLogger()를 호출하면
Then slog.LevelDebug로 설정된 로거를 반환한다
```

---

## 7. Non-Functional Requirements (비기능 요구사항)

### AC-070: 순환 의존성 없음

```gherkin
Given 전체 Go 프로젝트 소스가 있고
When "go vet ./..."를 실행하면
Then 순환 의존성 관련 에러가 0건이다
```

### AC-071: CLI 시작 성능

```gherkin
Given moai 바이너리가 release 모드로 빌드되어 있고
When "moai version"을 cold start로 실행하면
Then 전체 실행 시간이 50ms 미만이다
```

### AC-072: pkg에서 internal 참조 없음

```gherkin
Given pkg/ 디렉토리의 모든 Go 파일을 검사하면
When import 구문을 분석하면
Then "internal/" 패키지를 참조하는 import가 0건이다
And "go build ./pkg/..." 가 정상 컴파일된다
```

### AC-073: 테스트 커버리지

```gherkin
Given 전체 테스트가 실행되면
When 커버리지를 측정하면
Then internal/cli/ 레이어 커버리지가 70% 이상이고
And pkg/ 패키지 커버리지가 95% 이상이고
And internal/core/migration/ 커버리지가 85% 이상이다
```

### AC-074: Race Condition 없음

```gherkin
Given 전체 테스트 스위트가 있고
When "go test -race ./..."로 실행하면
Then race condition 탐지 결과가 0건이다
```

---

## 8. Quality Gates (품질 게이트)

### AC-080: TRUST 5 - Tested

```gherkin
Given SPEC-CLI-001 구현이 완료되면
When 전체 테스트를 실행하면
Then 모든 CLI 커맨드에 대한 스모크 테스트가 존재하고
And 통합 테스트 스위트 5개 이상이 통과하고
And 마이그레이션 백업/롤백 테스트가 통과한다
```

### AC-081: TRUST 5 - Readable

```gherkin
Given SPEC-CLI-001 구현 소스코드를 검사하면
When 코드 품질을 분석하면
Then 모든 공개 함수/타입에 godoc 주석이 존재하고
And 함수명이 Go 네이밍 컨벤션을 준수하고
And golangci-lint 경고가 0건이다
```

### AC-082: TRUST 5 - Unified

```gherkin
Given SPEC-CLI-001 구현 소스코드를 검사하면
When 코드 스타일을 분석하면
Then gofumpt 포매팅이 적용되어 있고
And goimports로 import 정렬이 완료되어 있고
And 일관된 에러 처리 패턴이 사용된다
```

### AC-083: TRUST 5 - Secured

```gherkin
Given SPEC-CLI-001 구현 소스코드를 검사하면
When gosec 보안 스캔을 실행하면
Then 보안 취약점이 0건이고
And CLI 인자에 대한 입력 검증이 존재하고
And API 키 등 민감 정보가 평문으로 저장되지 않는다
```

### AC-084: TRUST 5 - Trackable

```gherkin
Given SPEC-CLI-001 관련 커밋이 작성되면
When 커밋 메시지를 분석하면
Then Conventional Commits 형식을 준수하고
And SPEC-CLI-001 참조가 포함된다
```

---

## 9. Verification Methods (검증 방법)

| 검증 항목 | 방법 | 도구 |
|----------|------|------|
| 커맨드 라우팅 | Cobra 커맨드 실행 + stdout 캡처 | `go test`, `bytes.Buffer` |
| DI Wiring | 구체 타입 import 검색 | `grep`, 코드 리뷰 |
| 순환 의존성 | 컴파일러 분석 | `go vet ./...` |
| Race Condition | 동시성 테스트 | `go test -race ./...` |
| 커버리지 | 커버리지 리포트 | `go test -cover ./...` |
| 성능 | 벤치마크 | `go test -bench ./...`, `time moai version` |
| 보안 | 정적 분석 | `gosec`, `golangci-lint` |
| 통합 검증 | 통합 테스트 엔진 | `go test -tags=integration` |
| 마이그레이션 | 테스트 fixture | `testdata/` + table-driven tests |
| Help 출력 | 스모크 테스트 | Cobra stdout 캡처 |
