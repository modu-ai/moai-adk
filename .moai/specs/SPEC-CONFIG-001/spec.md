---
id: SPEC-CONFIG-001
version: "1.1.0"
status: Completed
created: "2026-02-03"
updated: "2026-02-03"
author: GOOS
priority: P0-Critical
---

# SPEC-CONFIG-001: Configuration Management System

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-02-03 | GOOS | 최초 작성 |
| 1.1.0 | 2026-02-03 | GOOS | 개발 방법론 모드 지원 추가 (TDD, DDD, Hybrid) |

---

## 1. 개요

### 1.1 목적

MoAI-ADK Go Edition의 Configuration Management System을 구현한다. Python 기반 `UnifiedConfigManager` singleton(~73,000 LOC 코드베이스의 핵심 모듈)을 Viper-backed YAML 설정 + typed Go structs + `sync.RWMutex` 기반 thread-safe 동시 접근 시스템으로 재작성한다.

### 1.2 범위

- 모듈 경로: `internal/config/`
- 대상 파일: `manager.go`, `types.go`, `defaults.go`, `migration.go`, `validation.go`
- 예상 LOC: ~1,200
- Phase: Phase 1 - Foundation (P0 Critical)
- 의존성: 없음 (Foundation 모듈)

### 1.3 배경

Python MoAI-ADK에서 configuration 관련 25개 GitHub Issues가 보고되었다:
- #315: Config loading 실패 시 복구 불가
- #283: Concurrent access에서 race condition 발생
- #206: Type mismatch로 인한 runtime error
- #245: Default 값 누락으로 partial config 로드
- #243: YAML section 파일 간 불일치
- #304: Template variable 미치환으로 인한 config 오류

### 1.4 ADR 참조

- **ADR-001** (Modular Monolithic): Single binary 내 명확한 domain boundary로 config module 분리
- **ADR-003** (Interface-Based DDD): `ConfigManager` interface 기반 compile-time contract
- **ADR-011** (Zero Runtime Template Expansion): Config 파일은 Go struct serialization(`yaml.Marshal`)으로 생성

---

## 2. Environment (환경)

### 2.1 기술 스택

| 구성 요소 | 기술 | 버전 |
|-----------|------|------|
| Configuration Library | `github.com/spf13/viper` | v1.18+ |
| YAML Marshaling | `gopkg.in/yaml.v3` | v3.0+ |
| Concurrency | `sync.RWMutex` (stdlib) | Go 1.22+ |
| Atomic File Save | `os.Rename` (stdlib) | Go 1.22+ |
| Logging | `log/slog` (stdlib) | Go 1.22+ |
| Testing | `github.com/stretchr/testify` | v1.9+ |

### 2.2 설정 파일 구조

```
.moai/config/
  sections/
    user.yaml          # UserConfig
    language.yaml      # LanguageConfig
    quality.yaml       # QualityConfig (TRUST 5 포함)
    git_strategy.yaml  # GitStrategyConfig
    system.yaml        # SystemConfig
    llm.yaml           # LLMConfig
    pricing.yaml       # PricingConfig
    ralph.yaml         # RalphConfig
    workflow.yaml      # WorkflowConfig
```

### 2.3 환경 변수 오버라이드

| 환경 변수 | 대상 설정 | 기본값 |
|-----------|-----------|--------|
| `MOAI_CONFIG_DIR` | `.moai/` 디렉토리 위치 | 프로젝트 루트의 `.moai/` |
| `MOAI_LOG_LEVEL` | `system.log_level` | `info` |
| `MOAI_LOG_FORMAT` | `system.log_format` | `text` |
| `MOAI_NO_COLOR` | `system.no_color` | `false` |
| `MOAI_DEVELOPMENT_MODE` | `quality.development_mode` | `ddd` |

### 2.4 호환성 제약

- Python MoAI-ADK의 `.moai/config/sections/` YAML 형식과 하위 호환성 유지
- 기존 프로젝트의 설정 파일을 변경 없이 로드 가능해야 함
- 설정 파일에 unexpanded dynamic token(`$VAR`, `{{VAR}}`, `${SHELL}`)이 존재하면 안 됨 (ADR-011)

---

## 3. Assumptions (가정)

### 3.1 기술적 가정

- A1: 프로젝트 루트 디렉토리에 `.moai/config/sections/` 디렉토리가 존재한다고 가정한다. 존재하지 않을 경우 default 값으로 초기화한다.
- A2: Viper v1.18+는 YAML section 파일 로딩, environment variable binding, file watching을 안정적으로 지원한다.
- A3: `sync.RWMutex`는 goroutine 간 concurrent read/exclusive write 패턴을 보장한다.
- A4: `os.Rename`은 동일 파일시스템 내에서 atomic operation으로 동작한다.

### 3.2 비즈니스 가정

- A5: 설정 section은 10개 이내로 유지된다 (현재 9개 section).
- A6: 설정 파일의 총 크기는 section당 10KB를 초과하지 않는다.
- A7: 설정 변경은 초당 1회 미만으로 발생한다 (빈번한 hot reload 불필요).

### 3.3 위험 가정

- A8: Legacy JSON 형식 설정 파일이 존재할 수 있으며, YAML sections 형식으로 migration이 필요할 수 있다.
- A9: CI/CD 환경에서는 환경 변수를 통한 설정 오버라이드가 주요 사용 패턴이다.
- A10: `development_mode` 값은 `ddd`, `tdd`, `hybrid` 중 하나로 제한된다.

---

## 4. Requirements (요구사항)

### REQ-01: Configuration Loading (Ubiquitous)

시스템은 **항상** 프로젝트 루트의 `.moai/config/sections/` 디렉토리에서 YAML section 파일들을 읽어 typed Go struct로 deserialize해야 한다.

**세부 요구사항:**

- REQ-01.1: The system **shall** load all YAML section files from `.moai/config/sections/` directory and merge them into a single `Config` struct.
- REQ-01.2: The system **shall** apply compiled default values for any missing fields using the `defaults.go` definitions.
- REQ-01.3: The system **shall** bind environment variables with `MOAI_` prefix to corresponding configuration fields via Viper.
- REQ-01.4: The system **shall** resolve environment variable overrides with higher priority than file-based values.

### REQ-02: Thread-Safe Access (Ubiquitous)

시스템은 **항상** `sync.RWMutex`를 사용하여 concurrent goroutine 환경에서 안전한 configuration 읽기/쓰기를 보장해야 한다.

**세부 요구사항:**

- REQ-02.1: The system **shall** acquire a read lock (`RLock`) for all `Get()` and `GetSection()` operations.
- REQ-02.2: The system **shall** acquire a write lock (`Lock`) for all `SetSection()`, `Save()`, and `Reload()` operations.
- REQ-02.3: The system **shall** prevent data races detectable by Go's `-race` flag under concurrent access patterns.

### REQ-03: Event-Driven Operations (Event-Driven)

#### REQ-03.1: File Change Detection

**WHEN** 설정 파일이 디스크에서 변경되면 **THEN** 시스템은 등록된 callback 함수를 호출하여 변경 사항을 통지해야 한다.

- **When** a YAML section file is modified on disk, the system **shall** detect the change via Viper's file watcher and invoke all registered callbacks with the updated `Config`.

#### REQ-03.2: Section Update

**WHEN** `SetSection()` 호출로 특정 section이 업데이트되면 **THEN** 시스템은 in-memory config를 갱신하고 해당 section YAML 파일을 atomic하게 저장해야 한다.

- **When** `SetSection(name, value)` is called, the system **shall** validate the value against the section's type definition, update the in-memory config, and persist to the corresponding YAML file atomically using `os.Rename`.

#### REQ-03.3: Forced Reload

**WHEN** `Reload()` 가 호출되면 **THEN** 시스템은 write lock을 획득하고 디스크에서 모든 설정 파일을 다시 읽어 in-memory config를 완전히 갱신해야 한다.

- **When** `Reload()` is invoked, the system **shall** acquire a write lock, re-read all section files from disk, and replace the in-memory config atomically.

### REQ-04: Unwanted Behavior Prevention (Unwanted)

#### REQ-04.1: Invalid YAML Handling

**If** 설정 파일에 유효하지 않은 YAML이 포함되어 있으면, **then** 시스템은 해당 파일을 skip하고 default 값을 적용하며 `slog.Warn`으로 경고를 출력해야 한다. 전체 로딩을 실패시키지 않아야 한다.

- **If** a YAML section file contains invalid syntax, **then** the system **shall** log a warning with the file path and parsing error, skip the invalid file, and use compiled defaults for that section.

#### REQ-04.2: Partial Config Prevention

**If** 필수 설정 필드(`user.name` 등)가 누락되어 있으면, **then** 시스템은 validation error를 반환하되, actionable error message를 포함해야 한다.

- **If** a required field marked with `validate:"required"` is empty after loading, **then** the system **shall** return an error containing the field name, expected type, and example value.

#### REQ-04.3: Dynamic Token Prevention

**If** 설정 값에 unexpanded dynamic token(`$VAR`, `{{VAR}}`, `${SHELL}`)이 포함되어 있으면, **then** 시스템은 해당 값을 거부하고 validation error를 반환해야 한다.

- **If** any configuration value contains patterns matching `$\{.*\}`, `\{\{.*\}\}`, or `$[A-Z_]+`, **then** the system **shall** reject the value with an error identifying the field and the unexpanded token.

#### REQ-04.4: Concurrent Write Corruption

시스템은 동시 `Save()` 호출 시 설정 파일이 손상되는 것을 방지**하지 않아야 한다**.

- **If** multiple goroutines call `Save()` simultaneously, **then** the system **shall** serialize writes via the write lock, ensuring only one save operation proceeds at a time.

### REQ-05: State-Driven Operations (State-Driven)

#### REQ-05.1: Initialized State

**IF** ConfigManager가 정상적으로 `Load()`를 완료한 상태이면 **THEN** `Get()`, `GetSection()`, `Save()`, `Watch()`, `Reload()` 호출이 정상 동작해야 한다.

- **While** the ConfigManager is in the initialized state (after successful `Load()`), the system **shall** serve all read and write operations without requiring re-initialization.

#### REQ-05.2: Uninitialized State

**IF** ConfigManager가 아직 `Load()`를 호출하지 않은 상태이면 **THEN** `Get()` 호출은 nil을 반환하고, 다른 write 연산은 `ErrNotInitialized` error를 반환해야 한다.

- **While** the ConfigManager has not yet called `Load()`, the system **shall** return `nil` from `Get()` and `ErrNotInitialized` from `SetSection()`, `Save()`, `Watch()`, and `Reload()`.

#### REQ-05.3: Watching State

**IF** ConfigManager가 `Watch()`를 통해 file watcher가 활성화된 상태이면 **THEN** 설정 파일 변경 시 callback이 호출되어야 한다.

- **While** the ConfigManager has an active file watcher (after successful `Watch()` call), the system **shall** invoke registered callbacks within 1 second of detecting a file modification.

### REQ-06: Optional Features (Optional)

#### REQ-06.1: Legacy Migration

**Where** 프로젝트에 legacy JSON 형식 설정 파일(`config.json`, `unified_config.json`)이 존재하는 경우, 시스템은 YAML sections 형식으로 자동 migration을 제공해야 한다.

- **Where** legacy JSON configuration files exist, the system **shall** offer automatic migration to YAML sections format, preserving all existing values and creating a backup of the original files.

#### REQ-06.2: Configuration Diff

**Where** 가능하면, 시스템은 현재 설정과 default 값 간의 diff를 표시하는 기능을 제공해야 한다.

- **Where** the feature is available, the system **shall** provide a `Diff()` method returning a list of fields that differ from their compiled defaults.

### REQ-07: Methodology Selection (개발 방법론 선택)

#### REQ-07.1: Development Mode Validation (Ubiquitous)

시스템은 **항상** `development_mode` 값이 `ddd`, `tdd`, `hybrid` 중 하나인지 검증해야 한다. 유효하지 않은 값이 제공되면 유효한 옵션을 포함한 validation error를 반환해야 한다.

- **REQ-METHOD-001**: The system **shall** always validate that `development_mode` is one of: `ddd`, `tdd`, `hybrid`. **If** an invalid value is provided, **then** the system **shall** return a validation error with the list of valid options.

#### REQ-07.2: DDD Mode Activation (State-Driven)

**IF** `development_mode`가 `ddd`이면 **THEN** 시스템은 `ddd_settings`를 품질 관리에 적용해야 한다: 변경 전 characterization test, 동작 보존, 점진적 변환.

- **REQ-METHOD-002**: **While** `development_mode` is `ddd`, the system **shall** use `ddd_settings` for quality enforcement: characterization tests before changes, behavior preservation, and incremental transformations.

#### REQ-07.3: TDD Mode Activation (State-Driven)

**IF** `development_mode`가 `tdd`이면 **THEN** 시스템은 `tdd_settings`를 품질 관리에 적용해야 한다: test-first 개발 (RED-GREEN-REFACTOR), 커밋당 최소 커버리지, 선택적 mutation testing.

- **REQ-METHOD-003**: **While** `development_mode` is `tdd`, the system **shall** use `tdd_settings` for quality enforcement: test-first development (RED-GREEN-REFACTOR), minimum coverage per commit, and optional mutation testing.

#### REQ-07.4: Hybrid Mode Activation (State-Driven)

**IF** `development_mode`가 `hybrid`이면 **THEN** 시스템은 `hybrid_settings`를 적용해야 한다: 신규 기능/모듈에는 TDD workflow, 레거시 코드 리팩토링에는 DDD workflow, 별도의 커버리지 목표.

- **REQ-METHOD-004**: **While** `development_mode` is `hybrid`, the system **shall** use `hybrid_settings`: TDD workflow for new features/modules, DDD workflow for legacy code refactoring, and separate coverage targets.

#### REQ-07.5: Mode-Specific Settings Loading (Event-Driven)

**WHEN** 시스템이 quality.yaml을 로드하면 **THEN** 설정된 `development_mode`에 해당하는 설정만 활성화하고 다른 모드의 설정은 무시해야 한다.

- **REQ-METHOD-005**: **When** the system loads quality.yaml, the system **shall** activate only the settings corresponding to the configured `development_mode` and ignore settings for other modes.

#### REQ-07.6: Environment Variable Override (Ubiquitous)

시스템은 **항상** 환경 변수 오버라이드를 지원해야 한다: `MOAI_DEVELOPMENT_MODE`는 quality.yaml의 `development_mode` 값을 오버라이드한다.

- **REQ-METHOD-006**: The system **shall** always support environment variable override: `MOAI_DEVELOPMENT_MODE` overrides `development_mode` in quality.yaml.

---

## 5. Specifications (상세 설계)

### 5.1 Interface Definitions

```go
// ConfigManager provides thread-safe configuration management.
// Location: internal/config/manager.go
type ConfigManager interface {
    // Load reads configuration from the project root's .moai/config/sections/ directory.
    Load(projectRoot string) (*Config, error)

    // Get returns the current in-memory configuration. Thread-safe via RWMutex.
    Get() *Config

    // GetSection returns a named configuration section (user, language, quality, etc.).
    GetSection(name string) (interface{}, error)

    // SetSection updates a named configuration section in memory and persists to disk.
    SetSection(name string, value interface{}) error

    // Save persists the current configuration to disk atomically (temp + rename).
    Save() error

    // Watch registers a callback invoked when configuration files change on disk.
    Watch(callback func(Config)) error

    // Reload forces a re-read from disk, acquiring write lock.
    Reload() error
}
```

### 5.2 Config Struct Hierarchy

```go
// Config is the root configuration aggregate containing all sections.
// Location: internal/config/types.go
type Config struct {
    User        UserConfig        `yaml:"user"`
    Language    LanguageConfig    `yaml:"language"`
    Project     ProjectConfig     `yaml:"project"`
    Quality     QualityConfig     `yaml:"quality"`
    GitStrategy GitStrategyConfig `yaml:"git_strategy"`
    System      SystemConfig      `yaml:"system"`
    LLM         LLMConfig         `yaml:"llm"`
    Pricing     PricingConfig     `yaml:"pricing"`
    Ralph       RalphConfig       `yaml:"ralph"`
    Workflow    WorkflowConfig    `yaml:"workflow"`
}

type UserConfig struct {
    Name string `yaml:"name" validate:"required"`
}

type LanguageConfig struct {
    ConversationLanguage     string `yaml:"conversation_language" default:"en"`
    ConversationLanguageName string `yaml:"conversation_language_name" default:"English"`
    AgentPromptLanguage      string `yaml:"agent_prompt_language" default:"en"`
    GitCommitMessages        string `yaml:"git_commit_messages" default:"en"`
    CodeComments             string `yaml:"code_comments" default:"en"`
    Documentation            string `yaml:"documentation" default:"en"`
    ErrorMessages            string `yaml:"error_messages" default:"en"`
}

// DevelopmentMode defines the development methodology.
type DevelopmentMode string

const (
    ModeDDD    DevelopmentMode = "ddd"     // Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE)
    ModeTDD    DevelopmentMode = "tdd"     // Test-Driven Development (RED-GREEN-REFACTOR)
    ModeHybrid DevelopmentMode = "hybrid"  // Hybrid (TDD for new code, DDD for legacy)
)

type QualityConfig struct {
    DevelopmentMode    DevelopmentMode    `yaml:"development_mode" default:"hybrid"`
    EnforceQuality     bool               `yaml:"enforce_quality" default:"true"`
    TestCoverageTarget int                `yaml:"test_coverage_target" default:"85"`
    DDDSettings        DDDSettings        `yaml:"ddd_settings"`
    TDDSettings        TDDSettings        `yaml:"tdd_settings"`
    HybridSettings     HybridSettings     `yaml:"hybrid_settings"`
    CoverageExemptions CoverageExemptions `yaml:"coverage_exemptions"`
    LSPQualityGates    LSPQualityGates    `yaml:"lsp_quality_gates"`
}

type DDDSettings struct {
    RequireExistingTests  bool   `yaml:"require_existing_tests" default:"true"`
    CharacterizationTests bool   `yaml:"characterization_tests" default:"true"`
    BehaviorSnapshots     bool   `yaml:"behavior_snapshots" default:"true"`
    MaxTransformationSize string `yaml:"max_transformation_size" default:"small"`
}

// TDDSettings configures Test-Driven Development mode.
// Best for: Isolated new modules with no existing code dependencies (rare).
type TDDSettings struct {
    RedGreenRefactor       bool `yaml:"red_green_refactor" default:"true"`
    TestFirstRequired      bool `yaml:"test_first_required" default:"true"`
    MinCoveragePerCommit   int  `yaml:"min_coverage_per_commit" default:"80"`
    MutationTestingEnabled bool `yaml:"mutation_testing_enabled" default:"false"`
}

// HybridSettings configures Hybrid mode (TDD for new, DDD for legacy).
// Best for: All development work (new projects, new features, ongoing development).
type HybridSettings struct {
    NewFeatures         string `yaml:"new_features" default:"tdd"`       // tdd for new code
    LegacyRefactoring   string `yaml:"legacy_refactoring" default:"ddd"` // ddd for existing code
    MinCoverageNew      int    `yaml:"min_coverage_new" default:"90"`
    MinCoverageLegacy   int    `yaml:"min_coverage_legacy" default:"85"`
    PreserveRefactoring bool   `yaml:"preserve_refactoring" default:"true"`
}

// CoverageExemptions allows gradual coverage improvement for legacy code.
type CoverageExemptions struct {
    Enabled              bool `yaml:"enabled" default:"false"`
    RequireJustification bool `yaml:"require_justification" default:"true"`
    MaxExemptPercentage  int  `yaml:"max_exempt_percentage" default:"5"`
}

type LSPQualityGates struct {
    Enabled         bool     `yaml:"enabled" default:"true"`
    Plan            PlanGate `yaml:"plan"`
    Run             RunGate  `yaml:"run"`
    Sync            SyncGate `yaml:"sync"`
    CacheTTLSeconds int      `yaml:"cache_ttl_seconds" default:"5"`
    TimeoutSeconds  int      `yaml:"timeout_seconds" default:"3"`
}

type PlanGate struct {
    RequireBaseline bool `yaml:"require_baseline" default:"true"`
}

type RunGate struct {
    MaxErrors       int  `yaml:"max_errors" default:"0"`
    MaxTypeErrors   int  `yaml:"max_type_errors" default:"0"`
    MaxLintErrors   int  `yaml:"max_lint_errors" default:"0"`
    AllowRegression bool `yaml:"allow_regression" default:"false"`
}

type SyncGate struct {
    MaxErrors       int  `yaml:"max_errors" default:"0"`
    MaxWarnings     int  `yaml:"max_warnings" default:"10"`
    RequireCleanLSP bool `yaml:"require_clean_lsp" default:"true"`
}

type GitStrategyConfig struct {
    AutoBranch   bool   `yaml:"auto_branch" default:"false"`
    BranchPrefix string `yaml:"branch_prefix" default:"moai/"`
    CommitStyle  string `yaml:"commit_style" default:"conventional"`
    WorktreeRoot string `yaml:"worktree_root"`
}

type SystemConfig struct {
    Version        string `yaml:"version"`
    LogLevel       string `yaml:"log_level" default:"info"`
    LogFormat      string `yaml:"log_format" default:"text"`
    NoColor        bool   `yaml:"no_color" default:"false"`
    NonInteractive bool   `yaml:"non_interactive" default:"false"`
}

type LLMConfig struct {
    DefaultModel string `yaml:"default_model" default:"sonnet"`
    QualityModel string `yaml:"quality_model" default:"opus"`
    SpeedModel   string `yaml:"speed_model" default:"haiku"`
}

type PricingConfig struct {
    TokenBudget  int  `yaml:"token_budget" default:"250000"`
    CostTracking bool `yaml:"cost_tracking" default:"false"`
}

type RalphConfig struct {
    MaxIterations int  `yaml:"max_iterations" default:"5"`
    AutoConverge  bool `yaml:"auto_converge" default:"true"`
    HumanReview   bool `yaml:"human_review" default:"true"`
}

type WorkflowConfig struct {
    AutoClear  bool `yaml:"auto_clear" default:"true"`
    PlanTokens int  `yaml:"plan_tokens" default:"30000"`
    RunTokens  int  `yaml:"run_tokens" default:"180000"`
    SyncTokens int  `yaml:"sync_tokens" default:"40000"`
}
```

### 5.3 Error Types

```go
// Sentinel errors for ConfigManager operations.
var (
    ErrNotInitialized     = errors.New("config: manager not initialized, call Load() first")
    ErrSectionNotFound    = errors.New("config: section not found")
    ErrSectionTypeMismatch = errors.New("config: section type mismatch")
    ErrValidationFailed   = errors.New("config: validation failed")
    ErrDynamicToken            = errors.New("config: unexpanded dynamic token detected")
    ErrInvalidYAML             = errors.New("config: invalid YAML syntax")
    ErrInvalidDevelopmentMode  = errors.New("config: invalid development_mode, must be one of: ddd, tdd, hybrid")
)
```

### 5.4 파일별 책임

| 파일 | 책임 | Python 대응 |
|------|------|-------------|
| `manager.go` | `ConfigManager` interface 구현, `sync.RWMutex`, Viper 통합 | `core/config/unified.py` |
| `types.go` | 모든 Config struct 정의 (yaml tags, default tags, validate tags) | -- (NEW) |
| `defaults.go` | Compiled default values, `NewDefaultConfig()` 함수 | -- (NEW) |
| `migration.go` | Legacy JSON -> YAML sections migration, backup 생성 | `core/config/migration.py` |
| `validation.go` | Struct tag 기반 validation, dynamic token detection | -- (NEW) |

### 5.5 Internal Module Dependency

```
manager.go ──→ types.go (Config struct 정의)
           ──→ defaults.go (기본값 적용)
           ──→ validation.go (로드 후 검증)
           ──→ migration.go (legacy 형식 감지 시)
```

외부 의존성 방향:
```
internal/cli/ ──→ internal/config/ (설정 로드)
internal/hook/ ──→ internal/config/ (설정 참조)
internal/core/* ──→ internal/config/ (설정 참조)
```

### 5.6 Methodology-Specific Behavior (방법론별 동작)

#### DDD Mode (default)

- **대상**: 테스트 커버리지가 10% 미만인 기존 프로젝트
- **Run Phase 동작**: ANALYZE-PRESERVE-IMPROVE 사이클 사용
  - ANALYZE: 기존 코드 분석, 의존성 파악, 도메인 경계 매핑
  - PRESERVE: characterization test 작성, 현재 동작 캡처 (`behavior_snapshots`)
  - IMPROVE: 점진적 변경, 테스트 후 변경 (`max_transformation_size` 제한)
- **커버리지 전략**: 점진적 커버리지 향상, `coverage_exemptions` 활용 가능
- **활성 설정**: `ddd_settings` (characterization_tests, behavior_snapshots, max_transformation_size)

#### TDD Mode

- **대상**: 신규 프로젝트(greenfield) 또는 테스트 커버리지 50% 이상인 프로젝트
- **Run Phase 동작**: RED-GREEN-REFACTOR 사이클 사용
  - RED: 실패하는 테스트 먼저 작성 (`test_first_required`)
  - GREEN: 테스트를 통과시키는 최소한의 코드 작성
  - REFACTOR: 코드 개선, 테스트 재실행
- **커버리지 전략**: 커밋당 최소 커버리지 강제 (`min_coverage_per_commit: 80`)
- **활성 설정**: `tdd_settings` (red_green_refactor, test_first_required, min_coverage_per_commit, mutation_testing_enabled)

#### Hybrid Mode

- **대상**: 테스트 커버리지 10-49%인 프로젝트
- **Run Phase 동작**: 코드 유형에 따라 방법론 분기
  - 신규 기능/모듈: TDD (RED-GREEN-REFACTOR) 적용 (`new_features: tdd`)
  - 레거시 코드 리팩토링: DDD (ANALYZE-PRESERVE-IMPROVE) 적용 (`legacy_refactoring: ddd`)
- **커버리지 전략**: 신규/레거시 별도 커버리지 목표
  - 신규 코드: `min_coverage_new: 90`
  - 레거시 코드: `min_coverage_legacy: 85`
- **활성 설정**: `hybrid_settings` (new_features, legacy_refactoring, min_coverage_new, min_coverage_legacy, preserve_refactoring)

#### 방법론 선택 가이드

| 프로젝트 상태 | 권장 모드 | 이유 |
|---------------|-----------|------|
| 신규 프로젝트 (greenfield) | `tdd` | 처음부터 높은 커버리지 확보 가능 |
| 기존 프로젝트 (커버리지 < 10%) | `ddd` | 기존 동작 보존하며 점진적 개선 |
| 기존 프로젝트 (커버리지 10-49%) | `hybrid` | 신규는 TDD, 레거시는 DDD로 이원화 |
| 기존 프로젝트 (커버리지 >= 50%) | `tdd` | 충분한 테스트 기반으로 TDD 전환 가능 |

---

## 6. Traceability (추적성)

### 6.1 요구사항-파일 매핑

| 요구사항 | 구현 파일 | 테스트 파일 |
|----------|-----------|-------------|
| REQ-01 (Loading) | `manager.go`, `defaults.go` | `manager_test.go` |
| REQ-02 (Thread-Safe) | `manager.go` | `manager_test.go` (race test) |
| REQ-03 (Event-Driven) | `manager.go` | `manager_test.go` (watcher test) |
| REQ-04 (Unwanted) | `validation.go`, `manager.go` | `validation_test.go` |
| REQ-05 (State-Driven) | `manager.go` | `manager_test.go` (state test) |
| REQ-06 (Optional) | `migration.go` | `migration_test.go` |
| REQ-07 (Methodology) | `types.go`, `validation.go`, `manager.go` | `validation_test.go`, `manager_test.go` |

### 6.2 Issue 해결 매핑

| GitHub Issue | 근본 원인 | 해결 요구사항 |
|--------------|-----------|---------------|
| #315 | Config loading 실패 시 복구 불가 | REQ-04.1 (graceful degradation) |
| #283 | Concurrent access race condition | REQ-02 (RWMutex) |
| #206 | Type mismatch runtime error | REQ-01.2 (typed struct), REQ-04.2 (validation) |
| #245 | Default 값 누락 | REQ-01.2 (compiled defaults) |
| #243 | YAML section 파일 간 불일치 | REQ-01.1 (unified loading) |
| #304 | Template variable 미치환 | REQ-04.3 (dynamic token prevention) |

---

## 7. 제약사항

### 7.1 성능 제약

| 메트릭 | 목표 | 측정 방법 |
|--------|------|-----------|
| Config Load (cold) | < 10ms | Benchmark test |
| Config Load (cached) | < 1ms | Benchmark test |
| `Get()` latency | < 1us | Benchmark test (RLock 오버헤드만) |
| `Save()` latency | < 5ms | Benchmark test (atomic write) |
| Memory footprint | < 1MB | Runtime profiling |

### 7.2 보안 제약

- 설정 파일에 secret(API key, token 등)을 저장하지 않음
- 모든 파일 경로는 `filepath.Clean()` + directory containment check 적용
- `yaml.Unmarshal` 시 Go struct 기반 deserialization만 허용 (arbitrary map 금지)

### 7.3 호환성 제약

- Python MoAI-ADK의 `.moai/config/sections/` YAML 형식과 100% 하위 호환
- `yaml:"tag"` 명칭은 Python 측 key name과 동일하게 유지
- 설정 파일 인코딩은 UTF-8만 지원

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 94.1%

### Summary

Configuration management system implemented with YAML-based section loading, atomic file persistence, environment variable overlay, and runtime validation. Supports hierarchical configuration with user, language, quality, project, git strategy, and system sections. Includes file watching for live reload and thread-safe concurrent access via sync.RWMutex.

### Files Created

- `internal/config/defaults.go`
- `internal/config/defaults_test.go`
- `internal/config/errors.go`
- `internal/config/loader.go`
- `internal/config/loader_test.go`
- `internal/config/manager.go`
- `internal/config/manager_test.go`
- `internal/config/types.go`
- `internal/config/types_test.go`
- `internal/config/validation.go`
- `internal/config/validation_test.go`
