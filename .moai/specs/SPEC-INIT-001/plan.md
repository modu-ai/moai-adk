# SPEC-INIT-001: 구현 계획

---
spec_id: SPEC-INIT-001
title: Project Initialization & Detection - Implementation Plan
status: Planned
tags: [init, project, detection, wizard, bubbletea]
---

## 1. 구현 전략 개요

### 1.1 접근 방식

DDD(Domain-Driven Development) ANALYZE-PRESERVE-IMPROVE 사이클을 적용한다. 기존 Python MoAI-ADK의 `core/project/` 도메인을 분석하고, Go 관용적 패턴으로 재구현한다. 인터페이스 기반 설계(ADR-004)를 적용하여 각 컴포넌트가 독립적으로 테스트 가능하도록 구성한다.

### 1.2 핵심 설계 원칙

- **인터페이스 우선**: design.md에 정의된 `Initializer`, `Detector`, `ProjectValidator` 인터페이스를 먼저 구현
- **의존성 역전**: 구체 구현이 아닌 인터페이스에 의존
- **명시적 오류 처리**: 모든 오류를 컨텍스트와 함께 래핑 반환
- **구조체 직렬화**: JSON/YAML 생성은 반드시 `json.Marshal()` / `yaml.Marshal()` 사용 (ADR-008, ADR-011)

---

## 2. 마일스톤

### 마일스톤 1: Detector 구현 (Primary Goal)

**목표**: 프로젝트 루트에서 언어, 프레임워크, 프로젝트 타입을 자동 감지한다.

**구현 파일**: `internal/core/project/detector.go`

**태스크**:

- [ ] `Detector` 인터페이스 정의 (design.md 기준)
- [ ] `Language`, `Framework` 타입 정의
- [ ] `DetectLanguages(root string) ([]Language, error)` 구현
  - 파일 시스템 스캔: `package.json`, `go.mod`, `pyproject.toml`, `Cargo.toml` 등 16개 이상 매핑
  - Confidence 계산: 해당 언어 파일 수 / 전체 소스 파일 수
  - 결과를 Confidence 내림차순으로 정렬
- [ ] `DetectFrameworks(root string) ([]Framework, error)` 구현
  - 감지된 언어별 설정 파일 파싱 (`package.json` dependencies, `pyproject.toml` dependencies 등)
  - 프레임워크별 버전 추출
- [ ] `DetectProjectType(root string) (ProjectType, error)` 구현
  - 디렉토리 구조 기반 분류 로직 (cmd/ -> cli, public/ -> web-app, api/ -> api, 기본 -> library)
- [ ] 테이블 기반 단위 테스트 작성 (`detector_test.go`)
  - testdata/ 디렉토리에 각 언어/프레임워크별 샘플 프로젝트 구조 구성
  - 에지 케이스: 빈 디렉토리, 혼합 언어, 감지 불가 프로젝트

**의존성**: 없음 (독립 모듈)

**검증 기준**: 단위 테스트 통과, 커버리지 90% 이상

---

### 마일스톤 2: MethodologyDetector 구현 (Primary Goal)

**목표**: 프로젝트 테스트 커버리지를 분석하고 개발 방법론(DDD/TDD/Hybrid)을 추천한다.

**구현 파일**: `internal/core/project/methodology_detector.go`

**태스크**:

- [ ] `MethodologyDetector` 인터페이스 정의 (spec.md 기준)
- [ ] `MethodologyRecommendation`, `AlternativeMethodology` 타입 정의
- [ ] Tier 1: 언어별 테스트 파일 패턴 스캐너 구현
  - Go: `*_test.go`
  - Python: `test_*.py`, `*_test.py`
  - TypeScript/JavaScript: `*.test.ts`, `*.spec.ts`, `*.test.js`, `*.spec.js`
  - Java: `*Test.java`
  - Rust: `tests/` 디렉토리, `#[cfg(test)]` 모듈
  - Ruby: `*_spec.rb`, `*_test.rb`
  - PHP: `*Test.php`
  - C#: `*Tests.cs`
  - Kotlin: `*Test.kt`
- [ ] Tier 2: 커버리지 추정 로직 구현
  - `coverage_estimate = (test_file_count / code_file_count) * 100 * 0.2`
  - 보정 계수 0.2 (테스트 파일 1개 = 소스 코드 20% 커버리지 보수적 추정)
- [ ] Tier 3: 실제 커버리지 도구 실행 (선택 사항) 구현
  - `go test -cover ./...`, `pytest --cov`, `npm test -- --coverage` 등
  - 타임아웃 제어: `context.WithTimeout(30s)`
- [ ] 결정 트리 로직 구현
  - greenfield (code_file_count == 0): hybrid 추천
  - coverage >= 50%: tdd 추천
  - coverage >= 10%: hybrid 추천
  - coverage < 10%: ddd 추천 (TDD 경고 포함)
- [ ] `AlternativeMethodology` 생성 로직
  - 추천 외 대안 모드 제시
  - TDD 선택 시 예상 테스트 수 경고 메시지 생성
- [ ] 테이블 기반 단위 테스트 작성 (`methodology_detector_test.go`)
  - 케이스: greenfield, brownfield (테스트 없음), 부분 테스트, 높은 커버리지
  - testdata/ 디렉토리에 각 시나리오별 샘플 프로젝트 구조
- [ ] 벤치마크 테스트: Tier 1+2 < 500ms 검증

**의존성**: 마일스톤 1 (DetectLanguages 결과를 입력으로 사용)

**검증 기준**: 단위 테스트 통과, 커버리지 90% 이상, Tier 1+2 성능 < 500ms

---

### 마일스톤 3: Validator 구현 (Primary Goal)

**목표**: 프로젝트 구조의 무결성을 검증하고, 기존 MoAI 프로젝트 감지를 수행한다.

**구현 파일**: `internal/core/project/validator.go`

**태스크**:

- [ ] `ProjectValidator` 인터페이스 정의 (design.md 기준)
- [ ] `ValidationResult` 타입 정의
- [ ] `Validate(root string) (*ValidationResult, error)` 구현
  - `.moai/` 디렉토리 존재 여부 확인
  - `.claude/` 디렉토리 존재 여부 확인
  - `CLAUDE.md` 존재 여부 확인
  - 필수 하위 디렉토리 확인 (config/sections/, specs/, memory/)
- [ ] `ValidateMoAI(root string) (*ValidationResult, error)` 구현
  - 설정 파일 유효성 검증 (YAML 파싱 가능 여부)
  - 매니페스트 무결성 검증 (manifest.json 존재 및 유효성)
  - 에이전트/스킬 파일 존재 확인
- [ ] `--force` 재초기화 시 백업 로직
  - 기존 `.moai/` 을 `.moai.backup.{timestamp}/` 으로 이동
- [ ] 테이블 기반 단위 테스트 작성 (`validator_test.go`)
  - 케이스: 새 프로젝트, 기존 프로젝트, 손상된 프로젝트, 부분 초기화 프로젝트

**의존성**: 없음 (독립 모듈)

**검증 기준**: 단위 테스트 통과, 커버리지 90% 이상

---

### 마일스톤 4: Initializer 구현 (Secondary Goal)

**목표**: 프로젝트 스캐폴딩 오케스트레이션을 수행한다.

**구현 파일**: `internal/core/project/initializer.go`

**태스크**:

- [ ] `Initializer` 인터페이스 정의 (design.md 기준)
- [ ] `InitOptions` 타입 정의 (Force 필드 추가)
- [ ] `Init(ctx context.Context, opts InitOptions) error` 구현
  - 단계 1: `.moai/` 디렉토리 구조 생성
    - `config/sections/`, `specs/`, `reports/`, `memory/`, `logs/`
  - 단계 2: 설정 파일 생성 (SPEC-CONFIG-001 `config.Manager` 인터페이스 호출)
    - `user.yaml`, `language.yaml`, `quality.yaml`, `workflow.yaml`
  - 단계 3: `.claude/` 템플릿 배포 (SPEC-TEMPLATE-001 `template.Deployer` 인터페이스 호출)
    - `settings.json`, agents, skills, commands, rules, output-styles
  - 단계 4: `CLAUDE.md` 렌더링 (SPEC-TEMPLATE-001 `template.Renderer` 인터페이스 호출)
  - 단계 5: 매니페스트 초기화 (SPEC-TEMPLATE-001 `manifest.Manager` 인터페이스 호출)
  - 단계 6: Git 상태 확인 (SPEC-GIT-001 인터페이스 호출, 선택적)
- [ ] `Force` 모드 처리
  - Validator를 통한 기존 프로젝트 확인
  - 백업 수행 후 재초기화
- [ ] 오류 시 부분 롤백 로직
  - 초기화 중 실패 시 생성된 디렉토리/파일 정리
- [ ] 결과 출력 헬퍼
  - 생성된 파일 목록 포매팅
  - 다음 단계 안내 메시지
- [ ] 통합 테스트 작성 (`initializer_test.go`)
  - 의존 모듈은 mockery로 생성한 mock 사용
  - 케이스: 정상 초기화, 비인터랙티브 초기화, Force 재초기화, 의존 모듈 실패 시 롤백

**의존성**: SPEC-CONFIG-001, SPEC-TEMPLATE-001, SPEC-GIT-001 (인터페이스만 필요, mock으로 테스트 가능)

**검증 기준**: 단위/통합 테스트 통과, 커버리지 85% 이상

---

### 마일스톤 5: Phase Orchestrator 구현 (Secondary Goal)

**목표**: CLI 명령과 도메인 로직 사이의 초기화 페이즈를 조율한다.

**구현 파일**: `internal/core/project/phase.go`

**태스크**:

- [ ] 초기화 페이즈 정의
  - `PhaseDetect`: 언어/프레임워크/타입 감지
  - `PhaseMethodology`: 개발 방법론 자동 감지 (NEW)
  - `PhaseWizard`: 사용자 입력 수집 (또는 비인터랙티브 기본값, 방법론 선택 포함)
  - `PhaseValidate`: 프로젝트 구조 검증
  - `PhaseInit`: 실제 초기화 수행
  - `PhaseComplete`: 결과 출력
- [ ] `PhaseExecutor` 구현
  - 각 페이즈를 순차적으로 실행
  - `context.Context`를 통한 취소 지원
  - 페이즈별 진행 상황 콜백 (UI Progress 연동)
- [ ] TTY 감지 로직
  - `os.Stdin` 이 터미널인지 확인 (`term.IsTerminal()` 또는 동등 함수)
  - TTY 없으면 자동으로 비인터랙티브 모드 전환
- [ ] Wizard 인터페이스 연동
  - `internal/ui/` 의 `Wizard` 인터페이스를 주입받아 호출
  - 비인터랙티브 모드에서는 `NonInteractive` 인터페이스 사용
- [ ] MethodologyDetector 연동
  - `PhaseMethodology`에서 `MethodologyDetector.DetectMethodology()` 호출
  - `--development-mode` 플래그 제공 시 자동 감지 건너뛰기 (REQ-S-006)
  - 비인터랙티브 모드에서 자동 감지 결과 적용 (REQ-S-005)
  - 위저드에서 추천 방법론 표시 및 사용자 선택 지원 (REQ-E-021)
  - 추천과 다른 선택 시 경고 표시 (REQ-E-023)
- [ ] CLI 명령 연동 (`internal/cli/init.go` 수정)
  - Cobra 플래그 정의: `--non-interactive`, `-y`, `--force`, `--name`, `--language`, `--framework`, `--development-mode`
  - PhaseExecutor 인스턴스 생성 및 실행
- [ ] E2E 테스트 작성 (`phase_test.go`)
  - 전체 초기화 흐름 검증 (Detector -> Wizard -> Validator -> Initializer)

**의존성**: 마일스톤 1-4, `internal/ui/` (Wizard 인터페이스)

**검증 기준**: E2E 테스트 통과, 전체 모듈 커버리지 85% 이상

---

### 마일스톤 6: CLI 통합 및 마무리 (Final Goal)

**목표**: `internal/cli/init.go`와 도메인 로직을 연결하고, 엣지 케이스를 처리한다.

**태스크**:

- [ ] `internal/cli/init.go` Cobra 명령 완성
  - 플래그 파싱 및 InitOptions 매핑
  - PhaseExecutor 의존성 주입 (config.Manager, template.Deployer, git.Repository 등)
- [ ] 크로스 플랫폼 테스트
  - Windows 경로 처리 (`filepath.Clean()` 정규화 확인)
  - `os.Getenv("USER")` vs `os.Getenv("USERNAME")` 플랫폼 분기
- [ ] 벤치마크 테스트 작성
  - 언어 감지 < 100ms 검증
  - 전체 초기화 < 2s 검증
- [ ] golangci-lint 통과 확인
- [ ] godoc 주석 완성

**의존성**: 마일스톤 1-5, 의존 SPEC 모듈의 실제 구현 (통합 테스트용)

**검증 기준**: TRUST 5 품질 게이트 통과, 전체 커버리지 85% 이상

---

## 3. 기술 설계 방향

### 3.1 파일별 책임

| 파일 | 책임 | 예상 LOC |
|------|------|----------|
| `detector.go` | 언어/프레임워크/타입 감지 순수 로직 | ~300 |
| `methodology_detector.go` | 개발 방법론 자동 감지, 커버리지 추정, 추천 로직 | ~250 |
| `validator.go` | 프로젝트 구조 검증, 기존 프로젝트 감지 | ~200 |
| `initializer.go` | 초기화 오케스트레이션, 디렉토리 생성, 의존 모듈 호출 | ~350 |
| `phase.go` | 페이즈 실행 조율, TTY 감지, Wizard 연동, 방법론 선택 | ~200 |
| **합계** | | **~1,300** |

### 3.2 의존성 주입 구조

```go
// phase.go 에서 의존성을 주입받는 구조
type PhaseExecutor struct {
    detector            Detector
    methodologyDetector MethodologyDetector  // NEW: 개발 방법론 감지
    validator           ProjectValidator
    initializer         Initializer
    wizard              ui.Wizard            // internal/ui/ 패키지
    nonInteractive      ui.NonInteractive
    progress            ui.Progress
    logger              *slog.Logger
}

// initializer.go 에서 의존성을 주입받는 구조
type projectInitializer struct {
    configMgr    config.Manager       // SPEC-CONFIG-001
    deployer     template.Deployer    // SPEC-TEMPLATE-001
    renderer     template.Renderer    // SPEC-TEMPLATE-001
    manifestMgr  manifest.Manager     // SPEC-TEMPLATE-001
    gitRepo      git.Repository       // SPEC-GIT-001 (optional)
    logger       *slog.Logger
}
```

### 3.3 오류 처리 전략

```go
// 도메인별 오류 타입 정의
var (
    ErrProjectExists         = errors.New("project already initialized")
    ErrNoLanguageFound       = errors.New("no programming language detected")
    ErrInvalidRoot           = errors.New("invalid project root path")
    ErrInitFailed            = errors.New("initialization failed")
    ErrInvalidDevelopmentMode = errors.New("invalid development mode: must be ddd, tdd, or hybrid")
    ErrMethodologyDetection  = errors.New("methodology detection failed")
)

// 컨텍스트를 포함한 오류 래핑
fmt.Errorf("detect languages in %s: %w", root, err)
fmt.Errorf("create config: %w", ErrInitFailed)
```

### 3.4 테스트 전략

| 테스트 유형 | 위치 | 목적 | 도구 |
|------------|------|------|------|
| 단위 테스트 | `*_test.go` | 개별 함수/메서드 검증 | `testing`, `testify` |
| 테이블 기반 테스트 | `*_test.go` | 다양한 입력 시나리오 | `t.Run()`, `testify` |
| Mock 테스트 | `*_test.go` | 의존 인터페이스 격리 | `mockery` |
| 벤치마크 | `*_bench_test.go` | 성능 목표 검증 | `testing.B` |
| 테스트 데이터 | `testdata/` | 샘플 프로젝트 구조 | 파일 시스템 |

### 3.5 testdata 구조 (예시)

```
internal/core/project/testdata/
├── go-project/
│   ├── go.mod
│   ├── cmd/main.go
│   └── internal/
├── node-react/
│   ├── package.json        (react, react-dom 의존성)
│   ├── src/
│   └── public/
├── python-fastapi/
│   ├── pyproject.toml       (fastapi 의존성)
│   └── app/
├── rust-actix/
│   ├── Cargo.toml           (actix-web 의존성)
│   └── src/
├── empty-project/           (감지 불가 케이스, greenfield)
├── multi-language/          (Go + TypeScript 혼합)
│   ├── go.mod
│   ├── package.json
│   └── ...
├── existing-moai/           (기존 프로젝트 케이스)
│   ├── .moai/
│   │   └── config/sections/
│   └── .claude/
├── brownfield-no-tests/     (방법론 감지: DDD 추천 케이스)
│   ├── go.mod
│   ├── cmd/main.go
│   ├── internal/handler/user.go
│   ├── internal/handler/auth.go
│   └── internal/service/      (테스트 파일 없음, 소스 50개)
├── brownfield-high-coverage/  (방법론 감지: TDD 추천 케이스)
│   ├── go.mod
│   ├── internal/handler/user.go
│   ├── internal/handler/user_test.go
│   ├── internal/service/auth.go
│   ├── internal/service/auth_test.go
│   └── ...                   (소스 30개, 테스트 25개)
├── brownfield-partial-tests/  (방법론 감지: Hybrid 추천 케이스)
│   ├── go.mod
│   ├── internal/handler/user.go
│   ├── internal/handler/user_test.go
│   └── ...                   (소스 100개, 테스트 15개)
└── python-no-tests/           (방법론 감지: Python DDD 케이스)
    ├── pyproject.toml
    ├── app/main.py
    ├── app/routes/user.py
    └── app/models/user.py     (테스트 파일 없음)
```

---

## 4. 리스크 및 대응

| 리스크 | 영향 | 대응 방안 |
|--------|------|-----------|
| 의존 SPEC (CONFIG, TEMPLATE, GIT) 구현 지연 | Initializer 통합 테스트 불가 | 인터페이스 기반 mock으로 단위 테스트 선행, 통합 테스트는 의존 SPEC 완료 후 |
| bubbletea 위저드 미구현 | Phase Orchestrator UI 연동 불가 | UI 인터페이스를 mock으로 대체하여 도메인 로직 먼저 완성 |
| 16개 이상 언어 감지의 정확도 | 잘못된 감지로 사용자 혼란 | Confidence 기반 순위 시스템, "기타" 옵션 제공, 사용자 수정 허용 |
| Windows 경로 호환성 | 경로 구분자 문제 | `filepath.Clean()` 일괄 적용, CI에서 Windows 테스트 |
| 대규모 프로젝트(10,000+ 파일)에서 감지 성능 | 감지 시간 초과 | `context.WithTimeout(100ms)` 적용, 최대 스캔 깊이 제한 |
| 방법론 커버리지 추정 정확도 | Tier 2 보정 계수(0.2)가 실제와 크게 다를 수 있음 | Tier 3 실제 실행 옵션 제공, 추정치임을 명시, 사용자 수정 허용 |
| Tier 3 커버리지 도구 미설치 | 실제 커버리지 측정 불가 | Tier 3는 선택 사항으로 유지, 도구 미설치 시 graceful fallback to Tier 2 결과 |
| 모노레포/멀티 언어 프로젝트의 방법론 감지 | 언어별 커버리지가 크게 다를 수 있음 | 주 언어(최고 Confidence) 기준으로 추천, 사용자에게 언어별 분석 결과 표시 |

---

## 5. 아키텍처 의사결정 참조

| ADR | 관련성 | 적용 방식 |
|-----|--------|-----------|
| ADR-003 (go:embed) | 높음 | 템플릿 배포 시 embed.FS 활용 |
| ADR-004 (인터페이스 경계) | 높음 | Detector, Validator, Initializer 인터페이스 정의 |
| ADR-005 (log/slog) | 높음 | 모든 로깅에 slog 사용 |
| ADR-006 (Charmbracelet) | 높음 | Wizard, Selector, Progress 연동 |
| ADR-008 (프로그래밍 방식 JSON 생성) | 높음 | config/settings 생성 시 구조체 직렬화 |
| ADR-010 (bubbletea TUI) | 높음 | 인터랙티브 초기화 위저드 |
| ADR-011 (런타임 템플릿 확장 금지) | 높음 | 생성 파일에 동적 토큰 금지 |
