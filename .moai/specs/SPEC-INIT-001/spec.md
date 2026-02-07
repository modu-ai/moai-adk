# SPEC-INIT-001: 프로젝트 초기화 및 감지

---
spec_id: SPEC-INIT-001
title: Project Initialization & Detection
status: Completed
priority: High
phase: "Phase 2 - Core Domains"
module: internal/core/project/
files:
  - initializer.go
  - detector.go
  - methodology_detector.go
  - validator.go
  - phase.go
estimated_loc: ~1,300
dependencies:
  - SPEC-CONFIG-001
  - SPEC-TEMPLATE-001
  - SPEC-GIT-001
created: 2026-02-03
lifecycle: spec-anchored
tags: [init, project, detection, wizard, bubbletea, cli]
---

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|-----------|
| 2026-02-03 | 1.0.0 | 최초 SPEC 작성 |
| 2026-02-03 | 1.1.0 | 개발 방법론 자동 감지 및 선택 기능 추가 (MethodologyDetector, REQ-E-021~023, REQ-S-005~006) |

---

## 1. 환경 (Environment)

### 1.1 시스템 컨텍스트

이 모듈은 `moai init` CLI 명령의 핵심 도메인 로직을 담당한다. 사용자가 MoAI 프로젝트를 초기화할 때 실행되며, `.moai/` 디렉토리 구조 생성, `.claude/` 템플릿 배포, `CLAUDE.md` 생성을 수행한다.

### 1.2 런타임 환경

- **OS**: macOS (arm64, amd64), Linux (arm64, amd64), Windows (amd64, arm64)
- **Go 버전**: 1.22+
- **터미널**: TTY 지원 터미널 (인터랙티브 모드) 또는 CI/CD 파이프라인 (비인터랙티브 모드)
- **의존 바이너리**: Git 2.30+ (시스템 설치)

### 1.3 모듈 위치

```
internal/core/project/
├── initializer.go              # 프로젝트 스캐폴딩 및 설정 오케스트레이션
├── detector.go                 # 언어/프레임워크/프로젝트 타입 자동 감지
├── methodology_detector.go     # 개발 방법론 자동 감지 (DDD/TDD/Hybrid)
├── validator.go                # 프로젝트 구조 무결성 검증
└── phase.go                    # 초기화 페이즈 실행 조율
```

### 1.4 의존 모듈

| 의존 SPEC | 모듈 | 사용 목적 |
|-----------|------|-----------|
| SPEC-CONFIG-001 | `internal/config/` | 설정 파일 로드/저장 (Viper + 타입 구조체) |
| SPEC-TEMPLATE-001 | `internal/template/` | 템플릿 배포 (go:embed), 매니페스트 관리 |
| SPEC-GIT-001 | `internal/core/git/` | Git 초기화 감지 및 설정 |

### 1.5 UI 의존성

| 모듈 | 인터페이스 | 사용 목적 |
|------|-----------|-----------|
| `internal/ui/` | `Wizard` | bubbletea 기반 인터랙티브 위저드 |
| `internal/ui/` | `NonInteractive` | 헤드리스 모드 지원 |
| `internal/ui/` | `Selector`, `Checkbox` | 단일/다중 선택 UI |
| `internal/ui/` | `Progress` | 진행률 표시 |

---

## 2. 가정 (Assumptions)

### 2.1 기술적 가정

| ID | 가정 | 신뢰도 | 근거 | 오류 시 리스크 |
|----|------|--------|------|---------------|
| A-01 | 프로젝트 루트에서 파일시스템 읽기 권한이 있다 | High | OS 표준 권한 모델 | 감지 기능 실패 |
| A-02 | 의존 SPEC 모듈(CONFIG, TEMPLATE, GIT)이 인터페이스를 제공한다 | High | design.md 인터페이스 정의 완료 | 컴파일 에러 |
| A-03 | bubbletea 위저드가 `internal/ui/` 에서 구현된다 | High | ADR-010 (Charmbracelet 결정) | UI 없이 비인터랙티브만 지원 |
| A-04 | `go:embed`로 번들된 템플릿이 바이너리에 포함된다 | High | ADR-003 (embed 결정) | 템플릿 배포 불가 |
| A-05 | CI/CD 환경에서 TTY가 없을 수 있다 | High | 일반적 CI 환경 | 위저드 행 걸림 |

### 2.2 비즈니스 가정

| ID | 가정 | 신뢰도 | 근거 |
|----|------|--------|------|
| B-01 | 기존 Python MoAI-ADK `.moai/` 구조와 호환성을 유지해야 한다 | High | product.md 하드 제약 |
| B-02 | 지원 언어 16개 이상의 감지가 필요하다 | High | product.md LSP 통합 스펙 |
| B-03 | `--force` 플래그로 기존 프로젝트 재초기화가 가능해야 한다 | Medium | 사용자 요구 패턴 |

---

## 3. 요구사항 (Requirements)

### 3.1 항상 유효한 요구사항 (Ubiquitous)

| ID | 요구사항 |
|----|----------|
| REQ-U-001 | 시스템은 **항상** 모든 생성 파일 경로에 `filepath.Clean()`을 적용하여 경로를 정규화해야 한다 |
| REQ-U-002 | 시스템은 **항상** `context.Context`를 첫 번째 매개변수로 받아 취소 및 타임아웃을 지원해야 한다 |
| REQ-U-003 | 시스템은 **항상** 오류를 `fmt.Errorf("context: %w", err)` 형식으로 래핑하여 반환해야 한다 |
| REQ-U-004 | 시스템은 **항상** `log/slog`를 사용한 구조화된 로깅을 수행해야 한다 |
| REQ-U-005 | 시스템은 **항상** 생성된 JSON 파일에 대해 `json.Valid()`로 유효성을 검증해야 한다 |

### 3.2 이벤트 기반 요구사항 (Event-Driven)

| ID | 요구사항 |
|----|----------|
| REQ-E-001 | **WHEN** 사용자가 `moai init`을 실행하고 TTY가 감지되면 **THEN** bubbletea 인터랙티브 위저드를 시작한다 |
| REQ-E-002 | **WHEN** 사용자가 `moai init --non-interactive`를 실행하면 **THEN** 기본값 또는 플래그 값으로 자동 초기화를 수행한다 |
| REQ-E-003 | **WHEN** 사용자가 `moai init -y`를 실행하면 **THEN** `--non-interactive`와 동일하게 비인터랙티브 모드로 동작한다 |
| REQ-E-004 | **WHEN** 프로젝트 루트에 `package.json`이 존재하면 **THEN** JavaScript/TypeScript 언어를 감지하고 Node.js 프레임워크를 탐색한다 |
| REQ-E-005 | **WHEN** 프로젝트 루트에 `go.mod`가 존재하면 **THEN** Go 언어를 감지하고 모듈 경로를 추출한다 |
| REQ-E-006 | **WHEN** 프로젝트 루트에 `pyproject.toml` 또는 `requirements.txt`가 존재하면 **THEN** Python 언어를 감지한다 |
| REQ-E-007 | **WHEN** 프로젝트 루트에 `Cargo.toml`이 존재하면 **THEN** Rust 언어를 감지한다 |
| REQ-E-008 | **WHEN** `package.json`에서 `react` 의존성이 발견되면 **THEN** React 프레임워크를 감지한다 |
| REQ-E-009 | **WHEN** `package.json`에서 `next` 의존성이 발견되면 **THEN** Next.js 프레임워크를 감지한다 |
| REQ-E-010 | **WHEN** `package.json`에서 `vue` 의존성이 발견되면 **THEN** Vue 프레임워크를 감지한다 |
| REQ-E-011 | **WHEN** `pyproject.toml`에서 `fastapi` 의존성이 발견되면 **THEN** FastAPI 프레임워크를 감지한다 |
| REQ-E-012 | **WHEN** `pyproject.toml`에서 `django` 의존성이 발견되면 **THEN** Django 프레임워크를 감지한다 |
| REQ-E-013 | **WHEN** Go import에서 `gin-gonic/gin`이 발견되면 **THEN** Gin 프레임워크를 감지한다 |
| REQ-E-014 | **WHEN** 위저드에서 사용자가 프로젝트 이름, 언어, 프레임워크, 기능을 선택하면 **THEN** `WizardResult`를 생성하여 Initializer에 전달한다 |
| REQ-E-015 | **WHEN** Initializer가 `InitOptions`를 수신하면 **THEN** 다음 순서로 초기화를 실행한다: (1) `.moai/` 디렉토리 구조 생성, (2) 설정 파일 생성 (SPEC-CONFIG-001), (3) 템플릿 배포 (SPEC-TEMPLATE-001), (4) `CLAUDE.md` 생성, (5) 매니페스트 초기화 |
| REQ-E-016 | **WHEN** 이미 `.moai/` 디렉토리가 존재하고 `--force` 플래그가 없으면 **THEN** 오류를 반환하고 재초기화를 안내한다 |
| REQ-E-017 | **WHEN** `--force` 플래그와 함께 실행되면 **THEN** 기존 설정을 백업한 후 재초기화를 수행한다 |
| REQ-E-018 | **WHEN** Git 저장소가 감지되지 않으면 **THEN** 경고를 표시하고 Git 없이 초기화를 계속한다 |
| REQ-E-019 | **WHEN** 감지된 언어가 여러 개이면 **THEN** 파일 수 기준 신뢰도(Confidence)를 계산하여 주 언어를 제안한다 |
| REQ-E-020 | **WHEN** 초기화가 완료되면 **THEN** 생성된 파일 목록과 다음 단계 안내를 출력한다 |
| REQ-E-021 | **WHEN** 초기화 위저드가 인터랙티브 모드로 실행되면 **THEN** 시스템은: (1) `MethodologyDetector.DetectMethodology()`를 실행하고, (2) 프로젝트 분석 결과(파일 수, 테스트 수, 추정 커버리지)를 표시하고, (3) 추천 방법론과 근거를 제시하고, (4) 사용자에게 추천 방법론, 대안, 또는 커스텀 선택을 허용한다 |
| REQ-E-022 | **WHEN** `MethodologyDetector`가 프로젝트를 분석하면 **THEN** Tier 1 (파일 존재 확인)과 Tier 2 (커버리지 추정)를 수행한다. Tier 3 (실제 실행)는 선택 사항이며 사용자 동의 시에만 실행한다 |
| REQ-E-023 | **WHEN** 사용자가 추천과 다른 방법론을 선택하면 **THEN** 시스템은 예상 노력(테스트 수, 대략적 범위)에 대한 경고를 표시한다 |

### 3.3 상태 기반 요구사항 (State-Driven)

| ID | 요구사항 |
|----|----------|
| REQ-S-001 | **IF** TTY가 연결되지 않은 상태이면 **THEN** 자동으로 비인터랙티브 모드로 전환한다 |
| REQ-S-002 | **IF** 프로젝트가 이미 초기화된 상태이면 **THEN** `moai init`은 기존 프로젝트 감지 결과를 표시하고 `--force` 사용을 안내한다 |
| REQ-S-003 | **IF** 비인터랙티브 모드이고 필수 값(프로젝트 이름)이 누락된 상태이면 **THEN** 현재 디렉토리 이름을 기본값으로 사용한다 |
| REQ-S-004 | **IF** 감지된 프레임워크가 없는 상태이면 **THEN** 위저드에서 "None" 옵션을 기본 선택으로 제공한다 |
| REQ-S-005 | **IF** 비인터랙티브 모드이고 `--development-mode` 플래그가 제공되지 않은 상태이면 **THEN** 자동 감지된 추천 방법론을 사용한다 |
| REQ-S-006 | **IF** `--development-mode` 플래그가 제공된 상태이면 **THEN** 지정된 모드를 사용하고 자동 감지를 건너뛴다 |

### 3.4 금지 요구사항 (Unwanted)

| ID | 요구사항 |
|----|----------|
| REQ-N-001 | 시스템은 사용자 확인 없이 기존 `.moai/` 디렉토리를 **덮어쓰지 않아야 한다** |
| REQ-N-002 | 시스템은 JSON/YAML 생성 시 문자열 연결(string concatenation)을 **사용하지 않아야 한다** -- 반드시 Go 구조체 직렬화를 사용한다 |
| REQ-N-003 | 시스템은 생성된 파일에 미확장 동적 토큰(`$VAR`, `{{VAR}}`, `${SHELL}`)을 **포함하지 않아야 한다** |
| REQ-N-004 | 시스템은 비인터랙티브 모드에서 사용자 입력을 **대기하지 않아야 한다** |
| REQ-N-005 | 시스템은 `panic()`을 복구 가능한 오류에 대해 **사용하지 않아야 한다** |

### 3.5 선택 요구사항 (Optional)

| ID | 요구사항 |
|----|----------|
| REQ-O-001 | **가능하면** 감지된 언어/프레임워크 정보를 위저드 기본값으로 미리 채워 제공한다 |
| REQ-O-002 | **가능하면** 초기화 진행률을 스피너/프로그레스 바로 표시한다 |
| REQ-O-003 | **가능하면** `.gitignore` 파일이 없을 경우 프로젝트 타입에 맞는 `.gitignore`를 생성한다 |

---

## 4. 명세 (Specifications)

### 4.1 핵심 인터페이스

#### 4.1.1 Initializer

```go
// Initializer handles project scaffolding and setup.
type Initializer interface {
    // Init creates a new MoAI project with the given options.
    Init(ctx context.Context, opts InitOptions) error
}

// InitOptions configures the project initialization.
type InitOptions struct {
    ProjectRoot     string
    ProjectName     string
    Language        string
    Framework       string
    Features        []string
    UserName        string
    ConvLang        string
    DevelopmentMode string   // "ddd", "tdd", or "hybrid"
    NonInteractive  bool
    Force           bool
}
```

#### 4.1.2 Detector

```go
// Detector identifies project characteristics from the filesystem.
type Detector interface {
    // DetectLanguages scans the project root and returns detected languages.
    DetectLanguages(root string) ([]Language, error)

    // DetectFrameworks scans for known framework configuration files.
    DetectFrameworks(root string) ([]Framework, error)

    // DetectProjectType classifies the project based on structure and files.
    DetectProjectType(root string) (ProjectType, error)
}

// Language represents a detected programming language.
type Language struct {
    Name       string
    Confidence float64   // 0.0 ~ 1.0 (파일 수 비율 기반)
    FileCount  int
}

// Framework represents a detected development framework.
type Framework struct {
    Name       string
    Version    string
    ConfigFile string    // 감지 근거 파일 경로
}
```

#### 4.1.3 MethodologyDetector

```go
// MethodologyDetector analyzes project test coverage to recommend a development methodology.
type MethodologyDetector interface {
    // DetectMethodology analyzes test coverage and recommends a development mode.
    DetectMethodology(root string, languages []Language) (*MethodologyRecommendation, error)
}

// MethodologyRecommendation provides a recommended development methodology with rationale.
type MethodologyRecommendation struct {
    Recommended      string   // "ddd", "tdd", or "hybrid"
    Confidence       float64  // 0.0 ~ 1.0
    Rationale        string   // Human-readable explanation
    ProjectType      string   // "greenfield" or "brownfield"
    TestFileCount    int      // Number of test files found
    CodeFileCount    int      // Number of source code files found
    CoverageEstimate float64  // Estimated coverage percentage (0-100)
    Alternatives     []AlternativeMethodology
}

// AlternativeMethodology represents a non-recommended but available option.
type AlternativeMethodology struct {
    Mode    string
    Reason  string
    Warning string // Warning message if this mode is chosen despite recommendation
}
```

#### 4.1.4 ProjectValidator

```go
// ProjectValidator checks project structure integrity.
type ProjectValidator interface {
    // Validate checks the overall project structure.
    Validate(root string) (*ValidationResult, error)

    // ValidateMoAI checks MoAI-specific configuration and file integrity.
    ValidateMoAI(root string) (*ValidationResult, error)
}

// ValidationResult holds project validation outcomes.
type ValidationResult struct {
    Valid    bool
    Errors   []string
    Warnings []string
}
```

### 4.2 언어 감지 매핑 테이블

| 감지 파일 | 언어 | 프레임워크 탐색 대상 |
|-----------|------|---------------------|
| `package.json` | JavaScript/TypeScript | React, Vue, Next.js, Angular, Svelte, Express, Nest.js |
| `go.mod` | Go | Gin, Echo, Fiber, Chi |
| `pyproject.toml` | Python | FastAPI, Django, Flask |
| `requirements.txt` | Python | FastAPI, Django, Flask |
| `Cargo.toml` | Rust | Actix, Axum, Rocket |
| `pom.xml` | Java | Spring Boot |
| `build.gradle` / `build.gradle.kts` | Java/Kotlin | Spring Boot, Ktor |
| `Gemfile` | Ruby | Rails, Sinatra |
| `composer.json` | PHP | Laravel, Symfony |
| `Package.swift` | Swift | Vapor |
| `pubspec.yaml` | Dart | Flutter |
| `mix.exs` | Elixir | Phoenix |
| `build.sbt` | Scala | Play, Akka |
| `*.cabal` / `stack.yaml` | Haskell | -- |
| `build.zig` | Zig | -- |
| `*.csproj` / `*.sln` | C# | ASP.NET |

### 4.3 프로젝트 타입 분류 로직

```
IF cmd/ 또는 main.go 존재           → cli
ELSE IF public/ 또는 src/pages/ 존재 → web-app
ELSE IF api/ 또는 routes/ 존재      → api
ELSE                                → library
```

### 4.4 초기화 실행 시퀀스

```
moai init
  |
  v
[1] Detector.DetectLanguages(root)
  |-- 파일 시스템 스캔, 매핑 테이블 기반 언어 감지
  |-- Confidence 계산 (해당 언어 파일 수 / 전체 파일 수)
  |
  v
[2] Detector.DetectFrameworks(root)
  |-- 감지된 언어별 설정 파일 파싱
  |-- 의존성 목록에서 프레임워크 식별
  |
  v
[3] Detector.DetectProjectType(root)
  |-- 디렉토리 구조 기반 분류 (cli, web-app, api, library)
  |
  v
[4] MethodologyDetector.DetectMethodology(root, languages)    ← NEW
  |-- Tier 1: 감지된 언어별 테스트 파일 스캔 (~100ms)
  |-- Tier 2: 테스트/코드 파일 비율로 커버리지 추정 (~1-3s)
  |-- MethodologyRecommendation 생성
  |
  v
[5] Wizard.Run(ctx) 또는 NonInteractive 기본값 적용           ← UPDATED
  |-- 감지 결과를 기본값으로 위저드에 전달
  |-- 사용자 선택: 프로젝트 이름, 언어, 프레임워크, 기능, 이름, 언어 설정
  |-- 방법론 선택: 추천 결과 표시, 사용자 확인/변경 허용        ← NEW
  |
  v
[6] ProjectValidator.Validate(root)
  |-- 기존 .moai/ 존재 여부 확인
  |-- --force 플래그 확인, 필요시 백업
  |
  v
[7] Initializer.Init(ctx, opts)                               ← opts에 DevelopmentMode 포함
  |-- (a) .moai/ 디렉토리 구조 생성
  |-- (b) config.Manager를 통한 설정 파일 생성 (SPEC-CONFIG-001)
  |-- (c) template.Deployer를 통한 템플릿 배포 (SPEC-TEMPLATE-001)
  |-- (d) CLAUDE.md 렌더링 및 생성
  |-- (e) manifest.Manager를 통한 매니페스트 초기화
  |-- (f) Git 상태 확인 (SPEC-GIT-001)
  |-- (g) quality.yaml에 development_mode 반영                 ← NEW
  |
  v
[8] 결과 출력
  |-- 생성된 파일 목록
  |-- 선택된 개발 방법론 표시                                   ← NEW
  |-- 다음 단계 안내 (moai doctor, moai status)
```

### 4.5 방법론 자동 감지 로직 (Methodology Auto-Detection)

#### Tier 1: 파일 존재 확인 (~100ms)

감지된 언어별 테스트 파일 패턴을 스캔한다:

| 언어 | 테스트 파일 패턴 |
|------|-----------------|
| Go | `*_test.go` |
| Python | `test_*.py`, `*_test.py` |
| TypeScript | `*.test.ts`, `*.spec.ts` |
| JavaScript | `*.test.js`, `*.spec.js` |
| Java | `*Test.java` |
| Rust | `tests/` 디렉토리 또는 `#[cfg(test)]` 모듈 |
| Ruby | `*_spec.rb`, `*_test.rb` |
| PHP | `*Test.php` |
| C# | `*Tests.cs` |
| Kotlin | `*Test.kt` |

#### Tier 2: 커버리지 추정 (~1-3s)

```
coverage_estimate = (test_file_count / code_file_count) * 100 * 0.2
```

보정 계수 0.2는 테스트 파일 1개가 평균적으로 소스 코드의 20%를 커버한다는 보수적 추정에 기반한다.

#### Tier 3: 실제 실행 (선택 사항, ~10-30s)

사용자 동의 시에만 실행한다. 언어별 커버리지 도구를 호출한다:

| 언어 | 명령어 |
|------|--------|
| Go | `go test -cover ./...` |
| Python | `pytest --cov` |
| TypeScript/JavaScript | `npm test -- --coverage` |
| Rust | `cargo tarpaulin` (설치된 경우) |
| Java | `mvn test` (JaCoCo 출력 파싱) |

#### 결정 트리 (Decision Tree)

```
IF code_file_count == 0 (greenfield 프로젝트):
    -> Recommended: hybrid (새 기능에 TDD + DDD 구조화)
    -> Confidence: 0.7
    -> Alternative: tdd (순수 테스트 우선)

ELSE IF coverage_estimate >= 50%:
    -> Recommended: tdd
    -> Confidence: 0.85
    -> Alternative: hybrid

ELSE IF coverage_estimate >= 10%:
    -> Recommended: hybrid
    -> Confidence: 0.75
    -> Alternative: ddd

ELSE (coverage < 10%, 테스트 없는 brownfield):
    -> Recommended: ddd (강력 추천)
    -> Confidence: 0.9
    -> Warning for TDD: "{N}개 파일 x 5 테스트 = {M}개 테스트 필요. 상당한 선행 투자가 필요합니다."
    -> Alternative: hybrid
```

### 4.6 비인터랙티브 모드 기본값

| 필드 | 기본값 | 결정 로직 |
|------|--------|-----------|
| ProjectName | 현재 디렉토리 이름 | `filepath.Base(root)` |
| Language | 감지된 주 언어 또는 "go" | `Detector.DetectLanguages()` 결과의 최고 Confidence |
| Framework | 감지된 프레임워크 또는 "none" | `Detector.DetectFrameworks()` 결과 |
| Features | 빈 슬라이스 | 기본 기능만 |
| UserName | OS 사용자 이름 | `os.Getenv("USER")` 또는 `os.Getenv("USERNAME")` |
| ConvLang | "en" | 시스템 로케일 감지 실패 시 |
| DevelopmentMode | 자동 감지 추천값 또는 "ddd" | `MethodologyDetector.DetectMethodology()` 결과, 또는 `--development-mode` 플래그 값 |

### 4.7 생성 디렉토리 구조

```
{project_root}/
├── .moai/
│   ├── config/
│   │   └── sections/
│   │       ├── user.yaml
│   │       ├── language.yaml
│   │       ├── quality.yaml
│   │       └── workflow.yaml
│   ├── specs/
│   ├── reports/
│   ├── memory/
│   ├── logs/
│   └── manifest.json
├── .claude/
│   ├── settings.json
│   ├── agents/moai/
│   ├── skills/
│   ├── commands/moai/
│   ├── rules/moai/
│   └── output-styles/
├── CLAUDE.md
└── .gitignore (선택적 생성)
```

### 4.8 성능 요구사항

| 메트릭 | 목표값 | 측정 방법 |
|--------|--------|-----------|
| 언어 감지 | < 100ms | 벤치마크 테스트 |
| 프레임워크 감지 | < 200ms | 벤치마크 테스트 |
| 방법론 감지 (Tier 1+2) | < 500ms | 벤치마크 테스트 |
| 방법론 감지 (Tier 3, 선택적) | < 30s | 벤치마크 테스트 |
| 전체 초기화 (템플릿 배포 포함) | < 3s | E2E 벤치마크 |
| 메모리 사용량 | < 50MB | 런타임 프로파일링 |

---

## 5. 추적성 (Traceability)

| 요구사항 | 구현 파일 | 테스트 시나리오 |
|----------|-----------|-----------------|
| REQ-E-001 ~ REQ-E-003 | `initializer.go`, `phase.go` | ACC-001, ACC-004 |
| REQ-E-004 ~ REQ-E-013 | `detector.go` | ACC-003, ACC-005 |
| REQ-E-014 | `phase.go` (UI 모듈 연동) | ACC-001 |
| REQ-E-015 | `initializer.go` | ACC-001, ACC-004 |
| REQ-E-016 ~ REQ-E-017 | `validator.go`, `initializer.go` | ACC-002, ACC-006 |
| REQ-E-018 | `initializer.go` | ACC-007 |
| REQ-E-019 | `detector.go` | ACC-005 |
| REQ-E-020 | `initializer.go`, `phase.go` | ACC-001 |
| REQ-E-021 | `methodology_detector.go`, `phase.go` | ACC-011, ACC-012, ACC-013, ACC-014 |
| REQ-E-022 | `methodology_detector.go` | ACC-011, ACC-012, ACC-013 |
| REQ-E-023 | `phase.go` (UI 모듈 연동) | ACC-014 |
| REQ-S-001 ~ REQ-S-004 | `phase.go`, `initializer.go` | ACC-004, ACC-008 |
| REQ-S-005 | `phase.go`, `methodology_detector.go` | ACC-015 |
| REQ-S-006 | `phase.go` | ACC-016 |
| REQ-N-001 ~ REQ-N-005 | 전체 모듈 | ACC-009 |
| REQ-O-001 ~ REQ-O-003 | `phase.go`, `initializer.go` | ACC-010 |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 89.2%

### Summary

Project initialization system implemented with language and framework auto-detection, development methodology recommendation engine using decision tree logic, and multi-phase initialization wizard. Detector scans project files to identify languages (by file extension and content analysis) and frameworks (by dependency files and configuration patterns). Methodology detector recommends DDD/TDD/Hybrid based on code file count and test coverage estimates. Supports both interactive (UI wizard) and non-interactive (CI/CD) modes with sensible defaults.

### Files Created

- `internal/core/project/detector.go`
- `internal/core/project/detector_test.go`
- `internal/core/project/errors.go`
- `internal/core/project/initializer.go`
- `internal/core/project/initializer_test.go`
- `internal/core/project/methodology_detector.go`
- `internal/core/project/methodology_detector_test.go`
- `internal/core/project/phase.go`
- `internal/core/project/phase_test.go`
- `internal/core/project/validator.go`
- `internal/core/project/validator_test.go`
