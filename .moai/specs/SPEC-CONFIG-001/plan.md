---
id: SPEC-CONFIG-001
version: "1.0.0"
status: Planned
created: "2026-02-03"
updated: "2026-02-03"
author: GOOS
priority: P0-Critical
---

# SPEC-CONFIG-001: Implementation Plan

## 1. 구현 전략

### 1.1 접근 방식

Bottom-up 구현 전략을 채택한다. 의존성이 없는 기초 파일부터 구현하여 상위 모듈로 확장한다.

```
types.go (독립) → defaults.go (types 의존) → validation.go (types 의존)
    → manager.go (types, defaults, validation 의존) → migration.go (manager, types 의존)
```

### 1.2 DDD 원칙 적용

- **ANALYZE**: 기존 Python `UnifiedConfigManager` 동작 분석 및 현행 `.moai/config/sections/` YAML 형식 파악
- **PRESERVE**: 기존 설정 파일 형식과의 하위 호환성 보장 (YAML key 명칭, 구조 유지)
- **IMPROVE**: Go의 정적 타입, compile-time validation, `sync.RWMutex` 기반 thread safety로 개선

---

## 2. Task 분해

### Milestone 1: Type Definitions (Primary Goal)

**파일: `types.go`**

| Task | 설명 | 우선순위 |
|------|------|----------|
| T1.1 | Root `Config` struct 정의 (10개 section) | High |
| T1.2 | `UserConfig` struct 정의 (`validate:"required"` tag 포함) | High |
| T1.3 | `LanguageConfig` struct 정의 (`default` tag 포함) | High |
| T1.4 | `QualityConfig` + `DDDSettings` + `TDDSettings` + `HybridSettings` + `CoverageExemptions` + `LSPQualityGates` 중첩 struct 정의 | High |
| T1.5 | `PlanGate`, `RunGate`, `SyncGate` struct 정의 | High |
| T1.6 | `GitStrategyConfig`, `SystemConfig`, `LLMConfig` struct 정의 | High |
| T1.7 | `PricingConfig`, `RalphConfig`, `WorkflowConfig` struct 정의 | High |
| T1.8 | `ProjectConfig` struct 정의 | High |
| T1.9 | Sentinel error 변수 정의 (`ErrNotInitialized`, `ErrSectionNotFound` 등) | High |
| T1.10 | `DevelopmentMode` enum type 및 상수 정의 (`ddd`, `tdd`, `hybrid`) | High |

**예상 LOC**: ~280

### Milestone 2: Default Values (Primary Goal)

**파일: `defaults.go`**

| Task | 설명 | 우선순위 |
|------|------|----------|
| T2.1 | `NewDefaultConfig()` 함수 구현 (전체 기본값 Config 생성) | High |
| T2.2 | Section별 `NewDefault{Section}Config()` helper 함수 구현 | High |
| T2.3 | `default` struct tag 파싱 및 적용 로직 구현 | High |
| T2.4 | 기본값 상수 정의 (magic number 방지) | Medium |
| T2.5 | `NewDefaultTDDSettings()`, `NewDefaultHybridSettings()`, `NewDefaultCoverageExemptions()` 구현 | High |

**예상 LOC**: ~200

### Milestone 3: Validation (Primary Goal)

**파일: `validation.go`**

| Task | 설명 | 우선순위 |
|------|------|----------|
| T3.1 | `Validate(cfg *Config) error` 함수 구현 | High |
| T3.2 | `validate:"required"` tag 기반 필수 필드 검증 로직 | High |
| T3.3 | Dynamic token 감지 regex (`$\{.*\}`, `\{\{.*\}\}`) 구현 | High |
| T3.4 | Actionable error message 포맷 (`field: expected type, got value`) | High |
| T3.5 | Section 이름 유효성 검증 (허용된 section 목록) | Medium |
| T3.6 | 값 범위 검증 (예: `test_coverage_target` 0-100) | Medium |
| T3.7 | `development_mode` 값 검증 (`ddd`, `tdd`, `hybrid` 허용 값 확인) | High |
| T3.8 | 환경 변수 `MOAI_DEVELOPMENT_MODE` 오버라이드 검증 | Medium |

**예상 LOC**: ~250

### Milestone 4: Core Manager (Primary Goal)

**파일: `manager.go`**

| Task | 설명 | 우선순위 |
|------|------|----------|
| T4.1 | `configManager` struct 정의 (`sync.RWMutex`, Viper instance, state) | High |
| T4.2 | `New() ConfigManager` constructor 구현 | High |
| T4.3 | `Load(projectRoot string) (*Config, error)` 구현 | High |
| T4.4 | Viper 기반 YAML section 파일 로딩 로직 | High |
| T4.5 | Environment variable binding (`MOAI_` prefix) | High |
| T4.6 | `Get() *Config` 구현 (RLock) | High |
| T4.7 | `GetSection(name string) (interface{}, error)` 구현 (reflect 기반) | High |
| T4.8 | `SetSection(name string, value interface{}) error` 구현 (type 검증 포함) | High |
| T4.9 | `Save() error` 구현 (atomic write: temp file + `os.Rename`) | High |
| T4.10 | `Watch(callback func(Config)) error` 구현 (Viper WatchConfig) | Medium |
| T4.11 | `Reload() error` 구현 (write lock + 전체 re-read) | Medium |
| T4.12 | State machine (uninitialized -> initialized -> watching) 구현 | High |
| T4.13 | Default 값 병합 로직 (missing field -> default 적용) | High |

**예상 LOC**: ~450

### Milestone 5: Legacy Migration (Secondary Goal)

**파일: `migration.go`**

| Task | 설명 | 우선순위 |
|------|------|----------|
| T5.1 | Legacy JSON 감지 로직 (`config.json`, `unified_config.json` 존재 확인) | Medium |
| T5.2 | JSON -> YAML sections 변환 로직 | Medium |
| T5.3 | 원본 파일 backup 생성 (`.bak` 확장자) | Medium |
| T5.4 | Migration 결과 리포팅 (변환된 필드 수, 경고 사항) | Low |
| T5.5 | Version 기반 migration step 관리 | Low |

**예상 LOC**: ~200

---

## 3. 기술 스택 상세

### 3.1 Core Dependencies

| Package | 사용 목적 | 대안 검토 | 선택 이유 |
|---------|-----------|-----------|-----------|
| `github.com/spf13/viper` v1.18+ | YAML 로딩, env binding, file watching | `koanf`, `envconfig` | Go 생태계 표준, Cobra와 통합, file watcher 내장 |
| `gopkg.in/yaml.v3` v3.0+ | YAML marshaling/unmarshaling | `github.com/goccy/go-yaml` | 안정성, 커뮤니티 지원, Viper 내부 의존성과 일치 |
| `sync.RWMutex` (stdlib) | Thread-safe concurrent access | `sync.Mutex`, channel | Read-heavy 패턴에 최적 (다수 reader, 소수 writer) |
| `os.Rename` (stdlib) | Atomic file save | 직접 write | 동일 파일시스템 내 atomic rename 보장 |

### 3.2 Viper 설정 패턴

```go
// Section별 Viper instance 생성 패턴
func loadSection(dir, name string, target interface{}) error {
    v := viper.New()
    v.SetConfigName(name)
    v.SetConfigType("yaml")
    v.AddConfigPath(dir)
    v.AutomaticEnv()
    v.SetEnvPrefix("MOAI")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            return nil // section 파일 없으면 default 사용
        }
        return fmt.Errorf("read %s config: %w", name, err)
    }

    return v.Unmarshal(target)
}
```

### 3.3 Atomic Save 패턴

```go
// temp file 작성 후 os.Rename으로 atomic 교체
func atomicSave(path string, data []byte) error {
    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, ".moai-config-*.tmp")
    if err != nil {
        return fmt.Errorf("create temp file: %w", err)
    }
    defer os.Remove(tmp.Name()) // cleanup on error

    if _, err := tmp.Write(data); err != nil {
        tmp.Close()
        return fmt.Errorf("write temp file: %w", err)
    }
    if err := tmp.Close(); err != nil {
        return fmt.Errorf("close temp file: %w", err)
    }

    return os.Rename(tmp.Name(), path)
}
```

---

## 4. Risk 분석 및 대응

### 4.1 기술 리스크

| 리스크 | 영향 | 가능성 | 대응 전략 |
|--------|------|--------|-----------|
| Viper file watcher가 macOS에서 불안정 | 설정 변경 감지 실패 | Medium | fsnotify 직접 사용 fallback 구현, polling 기반 대안 준비 |
| `os.Rename` cross-device 실패 | NFS/Docker volume에서 atomic save 실패 | Low | 동일 디렉토리에 temp 파일 생성으로 완화, cross-device 감지 시 copy+delete fallback |
| Viper env binding과 YAML key 충돌 | 환경 변수 오버라이드 미적용 | Low | `SetEnvKeyReplacer`로 `.` -> `_` 매핑 명시, integration test 검증 |
| default tag 파싱 복잡도 | Nested struct에서 default 누락 | Medium | `NewDefaultConfig()` 함수로 프로그래밍적 기본값 설정 (tag 파싱 의존성 최소화) |

### 4.2 호환성 리스크

| 리스크 | 영향 | 가능성 | 대응 전략 |
|--------|------|--------|-----------|
| Python YAML key 명칭 불일치 | 기존 프로젝트 설정 로드 실패 | Medium | 현행 Python 설정 파일을 test fixture로 활용, round-trip 검증 |
| Legacy JSON 형식 다양성 | Migration 실패 | Low | 알려진 JSON 형식 카탈로그화, unknown field는 무시 처리 |
| 중첩 section (quality.ddd_settings) 구조 차이 | Partial unmarshal 실패 | Medium | Viper의 `Sub()` 메서드로 중첩 구조 처리, 실패 시 default 적용 |

### 4.3 리스크 완화 우선순위

1. **Python 설정 파일 round-trip test**: 현행 `.moai/config/sections/` YAML 파일을 test fixture로 사용하여 load -> save -> reload 검증
2. **Race detector CI**: 모든 테스트를 `-race` flag로 실행하여 concurrent access 문제 조기 발견
3. **Benchmark regression guard**: `Load()`, `Get()`, `Save()` 성능 목표를 benchmark test로 검증

---

## 5. 파일별 구현 계획

### 5.1 `types.go` (Milestone 1)

**목적**: 모든 Config struct 정의

```
// 구조
package config

// Root config
type Config struct { ... }

// Section configs (9개)
type UserConfig struct { ... }
type LanguageConfig struct { ... }
type ProjectConfig struct { ... }
type QualityConfig struct { ... }
type GitStrategyConfig struct { ... }
type SystemConfig struct { ... }
type LLMConfig struct { ... }
type PricingConfig struct { ... }
type RalphConfig struct { ... }
type WorkflowConfig struct { ... }

// Development methodology
type DevelopmentMode string // enum: ddd, tdd, hybrid

// Nested configs
type DDDSettings struct { ... }
type TDDSettings struct { ... }
type HybridSettings struct { ... }
type CoverageExemptions struct { ... }
type LSPQualityGates struct { ... }
type PlanGate struct { ... }
type RunGate struct { ... }
type SyncGate struct { ... }

// Sentinel errors
var ErrNotInitialized = errors.New(...)
var ErrInvalidDevelopmentMode = errors.New(...)
```

**검증 기준**: 모든 struct가 `yaml.Unmarshal`로 올바르게 deserialize되는지 unit test 확인

### 5.2 `defaults.go` (Milestone 2)

**목적**: Compiled default 값 제공

```
// 핵심 함수
func NewDefaultConfig() *Config
func NewDefaultUserConfig() UserConfig
func NewDefaultLanguageConfig() LanguageConfig
func NewDefaultQualityConfig() QualityConfig
func NewDefaultTDDSettings() TDDSettings
func NewDefaultHybridSettings() HybridSettings
func NewDefaultCoverageExemptions() CoverageExemptions
// ... (section별 default 생성 함수)

// Default 상수
const (
    DefaultConversationLanguage = "en"
    DefaultTestCoverageTarget   = 85
    DefaultLogLevel             = "info"
    DefaultTokenBudget          = 250000
    // ...
)
```

**검증 기준**: `NewDefaultConfig()`가 모든 필드에 non-zero 기본값을 설정하는지 확인

### 5.3 `validation.go` (Milestone 3)

**목적**: Config 검증 로직

```
// 핵심 함수
func Validate(cfg *Config) error
func ValidateSection(name string, value interface{}) error
func detectDynamicTokens(value string) error

// 내부 helper
func validateRequired(cfg *Config) []ValidationError
func validateRanges(cfg *Config) []ValidationError
func validateSectionName(name string) error
func validateDevelopmentMode(mode DevelopmentMode) error

// ValidationError type
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}
```

**검증 기준**: Dynamic token 감지, 필수 필드 누락 감지, 범위 초과 감지 unit test

### 5.4 `manager.go` (Milestone 4)

**목적**: ConfigManager interface 구현

```
// Core implementation
type configManager struct {
    mu        sync.RWMutex
    config    *Config
    root      string
    state     managerState
    callbacks []func(Config)
    viper     *viper.Viper
}

type managerState int
const (
    stateUninitialized managerState = iota
    stateInitialized
    stateWatching
)

// Constructor
func New() ConfigManager

// Interface methods
func (m *configManager) Load(projectRoot string) (*Config, error)
func (m *configManager) Get() *Config
func (m *configManager) GetSection(name string) (interface{}, error)
func (m *configManager) SetSection(name string, value interface{}) error
func (m *configManager) Save() error
func (m *configManager) Watch(callback func(Config)) error
func (m *configManager) Reload() error

// Internal helpers
func (m *configManager) loadSections() (*Config, error)
func (m *configManager) saveSection(name string, data []byte) error
func (m *configManager) mergeDefaults(cfg *Config) *Config
```

**검증 기준**: Thread-safety (`-race`), state machine 전환, atomic save, file watcher 동작

### 5.5 `migration.go` (Milestone 5)

**목적**: Legacy 설정 형식 migration

```
// 핵심 함수
func DetectLegacy(projectRoot string) (LegacyFormat, error)
func Migrate(projectRoot string, format LegacyFormat) (*MigrationResult, error)

// Types
type LegacyFormat int
const (
    FormatNone LegacyFormat = iota
    FormatJSON
    FormatOldYAML
)

type MigrationResult struct {
    FieldsMigrated int
    FilesCreated   []string
    BackupPath     string
    Warnings       []string
}
```

**검증 기준**: JSON -> YAML 변환 정확성, backup 생성 확인, 원본 파일 보존

---

## 6. 의존성 그래프

```
                    types.go
                   /    |    \
                  /     |     \
          defaults.go   |   validation.go
                  \     |     /
                   \    |    /
                   manager.go
                       |
                  migration.go
```

### 6.1 외부 모듈 의존 방향

Configuration module은 Foundation 모듈로서 다른 모든 `internal/` 패키지에서 참조된다:

```
internal/cli/     ──→ internal/config/  (설정 로드, CLI flag 바인딩)
internal/hook/    ──→ internal/config/  (hook 실행 시 설정 참조)
internal/core/*   ──→ internal/config/  (도메인 로직에서 설정 참조)
internal/template/──→ internal/config/  (template 렌더링 시 설정 참조)
internal/lsp/     ──→ internal/config/  (LSP gate 설정 참조)
```

Configuration module은 다른 `internal/` 패키지에 의존하지 않는다 (zero dependencies).

---

## 7. 구현 순서 요약

| 순서 | 파일 | 의존성 | 예상 LOC |
|------|------|--------|----------|
| 1 | `types.go` | 없음 | ~280 |
| 2 | `defaults.go` | `types.go` | ~200 |
| 3 | `validation.go` | `types.go` | ~250 |
| 4 | `manager.go` | `types.go`, `defaults.go`, `validation.go` | ~450 |
| 5 | `migration.go` | `manager.go`, `types.go` | ~200 |
| **합계** | | | **~1,380** |

---

## 8. Quality Gate 기준

| 항목 | 기준 |
|------|------|
| Test Coverage | >= 85% (전체), >= 90% (`manager.go`, `validation.go`) |
| Race Detection | `go test -race` 통과 |
| Lint | `golangci-lint run` zero errors |
| Benchmark | `Load()` < 10ms, `Get()` < 1us, `Save()` < 5ms |
| Python 호환성 | 현행 YAML section 파일 round-trip test 통과 |
| Error Messages | 모든 error에 context 포함 (`fmt.Errorf("context: %w", err)`) |
