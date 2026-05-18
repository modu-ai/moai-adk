package config

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// @MX:NOTE: [AUTO] Thread-safe configuration loader that reads YAML section files and applies defaults
//
// Loader reads configuration from YAML section files.
// It is thread-safe via sync.RWMutex.
type Loader struct {
	mu             sync.RWMutex
	loadedSections map[string]bool
}

// NewLoader creates a new Loader instance.
func NewLoader() *Loader {
	return &Loader{}
}

// Load reads all configuration section files from the given .moai directory
// and returns a merged Config with defaults applied for missing fields.
// Missing files use default values. Invalid YAML files are skipped with a warning.
func (l *Loader) Load(configDir string) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.loadedSections = make(map[string]bool)
	cfg := NewDefaultConfig()

	sectionsDir := filepath.Join(filepath.Clean(configDir), "config", "sections")

	// If sections directory does not exist, return defaults
	if _, err := os.Stat(sectionsDir); os.IsNotExist(err) {
		slog.Warn("config sections directory not found, using defaults", "path", sectionsDir)
		return cfg, nil
	}

	// Load user section
	l.loadUserSection(sectionsDir, cfg)

	// Load language section
	l.loadLanguageSection(sectionsDir, cfg)

	// Load quality section
	l.loadQualitySection(sectionsDir, cfg)

	// Load git convention section
	l.loadGitConventionSection(sectionsDir, cfg)

	// Load LLM section
	l.loadLLMSection(sectionsDir, cfg)

	// Load ralph section (RalphConfig + Session.StaleSeconds)
	l.loadRalphSection(sectionsDir, cfg)

	// Load state section
	l.loadStateSection(sectionsDir, cfg)

	// Load statusline section
	l.loadStatuslineSection(sectionsDir, cfg)

	// Load research section
	l.loadResearchSection(sectionsDir, cfg)

	// Load constitution section (REQ-MIG003-001/002)
	l.loadConstitutionSection(sectionsDir, cfg)

	// Load context_search section (REQ-MIG003-010)
	l.loadContextSection(sectionsDir, cfg)

	// Load interview section (REQ-MIG003-011)
	l.loadInterviewSection(sectionsDir, cfg)

	// Load design section (REQ-MIG003-014)
	l.loadDesignSection(sectionsDir, cfg)

	// Emit once-per-session DORMANT notice for sunset.yaml (REQ-MIG003-018)
	emitSunsetDormantNotice(sectionsDir)

	return cfg, nil
}

// LoadedSections returns a copy of the map indicating which sections
// were successfully loaded from YAML files.
func (l *Loader) LoadedSections() map[string]bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]bool, len(l.loadedSections))
	maps.Copy(result, l.loadedSections)
	return result
}

// loadUserSection loads the user configuration section from user.yaml.
func (l *Loader) loadUserSection(dir string, cfg *Config) {
	wrapper := &userFileWrapper{User: cfg.User}
	loaded, err := loadYAMLFile(dir, "user.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load user config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.User = wrapper.User
		l.loadedSections["user"] = true
	}
}

// loadLanguageSection loads the language configuration section from language.yaml.
func (l *Loader) loadLanguageSection(dir string, cfg *Config) {
	wrapper := &languageFileWrapper{Language: cfg.Language}
	loaded, err := loadYAMLFile(dir, "language.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load language config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Language = wrapper.Language
		l.loadedSections["language"] = true
	}
}

// loadQualitySection loads the quality configuration section from quality.yaml.
// The quality.yaml file uses "constitution:" as the top-level key for
// backward compatibility with Python MoAI-ADK.
func (l *Loader) loadQualitySection(dir string, cfg *Config) {
	wrapper := &qualityFileWrapper{Constitution: cfg.Quality}
	loaded, err := loadYAMLFile(dir, "quality.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load quality config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Quality = wrapper.Constitution
		l.loadedSections["quality"] = true
	}
}

// loadGitConventionSection loads the git convention configuration from git-convention.yaml.
func (l *Loader) loadGitConventionSection(dir string, cfg *Config) {
	wrapper := &gitConventionFileWrapper{GitConvention: cfg.GitConvention}
	loaded, err := loadYAMLFile(dir, "git-convention.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load git convention config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.GitConvention = wrapper.GitConvention
		l.loadedSections["git_convention"] = true
	}
}

// loadLLMSection loads the LLM configuration section from llm.yaml.
func (l *Loader) loadLLMSection(dir string, cfg *Config) {
	wrapper := &llmFileWrapper{LLM: cfg.LLM}
	loaded, err := loadYAMLFile(dir, "llm.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load LLM config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.LLM = wrapper.LLM
		l.loadedSections["llm"] = true
	}
}

// loadStateSection loads the state configuration section from state.yaml.
func (l *Loader) loadStateSection(dir string, cfg *Config) {
	wrapper := &stateFileWrapper{State: cfg.State}
	loaded, err := loadYAMLFile(dir, "state.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load state config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.State = wrapper.State
		l.loadedSections["state"] = true
	}
}

// loadStatuslineSection loads the statusline configuration section from statusline.yaml.
func (l *Loader) loadStatuslineSection(dir string, cfg *Config) {
	wrapper := &statuslineFileWrapper{Statusline: cfg.Statusline}
	loaded, err := loadYAMLFile(dir, "statusline.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load statusline config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Statusline = wrapper.Statusline
		l.loadedSections["statusline"] = true
	}
}

// loadRalphSection loads the ralph configuration section from ralph.yaml.
// ralph.yaml의 ralph.stale_seconds 키를 Config.Session.StaleSeconds에 주입합니다.
// SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS 기본값 3600, ralph.yaml에서 오버라이드 가능.
func (l *Loader) loadRalphSection(dir string, cfg *Config) {
	wrapper := &ralphFileWrapper{}
	// ralph.yaml 기본값 초기화 (inline 필드)
	wrapper.Ralph.RalphConfig = cfg.Ralph
	loaded, err := loadYAMLFile(dir, "ralph.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load ralph config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Ralph = wrapper.Ralph.RalphConfig
		// stale_seconds가 0이 아닌 경우에만 오버라이드 (0은 명시적 설정 없음으로 간주)
		if wrapper.Ralph.StaleSeconds > 0 {
			cfg.Session.StaleSeconds = wrapper.Ralph.StaleSeconds
		}
		l.loadedSections["ralph"] = true
	}
}

// loadResearchSection loads the research configuration section from research.yaml.
func (l *Loader) loadResearchSection(dir string, cfg *Config) {
	wrapper := &researchFileWrapper{Research: cfg.Research}
	loaded, err := loadYAMLFile(dir, "research.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load research config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Research = wrapper.Research
		l.loadedSections["research"] = true
	}
}

// validHarnessLevels는 FROZEN level enum입니다.
// REQ-HRN-001-017: {minimal, standard, thorough} 외의 레벨은 ErrUnknownLevel을 반환합니다.
var validHarnessLevels = map[string]bool{
	"minimal":  true,
	"standard": true,
	"thorough": true,
}

// knownHarnessTopLevelKeys는 harnessFileWrapper.Harness 하위의 알려진 최상위 키 집합입니다.
// REQ-HRN-001-019: 알 수 없는 키 감지를 위한 기준 집합.
var knownHarnessTopLevelKeys = map[string]bool{
	"default_profile":      true,
	"mode_defaults":        true,
	"auto_detection":       true,
	"escalation":           true,
	"effort_mapping":       true,
	"levels":               true,
	"model_upgrade_review": true,
	"plan_audit_global":    true,
	"evaluator":            true,
	"learning":             true, // 레거시 하위 시스템 (out-of-scope, 경고 억제)
}

// LoadHarnessConfig는 주어진 경로의 harness.yaml 파일을 읽어 HarnessConfig를 반환합니다.
//
// HRN-001 run-phase: evaluator.memory_scope FROZEN 검증 + 전체 스키마 파싱.
// HRN-001 확장 검증:
//   - level enum {minimal, standard, thorough} FROZEN (REQ-HRN-001-017)
//   - MOAI_CONFIG_STRICT=1 시 알 수 없는 키 → ErrSchemaDrift 오류 (REQ-HRN-001-019)
//   - MOAI_CONFIG_STRICT 미설정 시 알 수 없는 키 → slog.Warn (REQ-HRN-001-019)
//
// @MX:ANCHOR: [AUTO] harness.yaml 로더 함수 — 다수의 호출자 (router, CLI, ConfigManager.Reload)
// @MX:REASON: fan_in >= 3: HarnessRouter.Route, CLI validate, ConfigManager.Reload에서 호출
func LoadHarnessConfig(path string) (*HarnessConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("LoadHarnessConfig: %w", ErrConfigNotFound)
		}
		return nil, fmt.Errorf("LoadHarnessConfig read %s: %w", path, err)
	}

	// 1단계: 구조체로 언마샬링
	var wrapper harnessFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadHarnessConfig parse %s: %w", path, ErrInvalidYAML)
	}

	cfg := &wrapper.Harness

	// 2단계: schema drift 감지 (알 수 없는 최상위 키 탐지)
	// REQ-HRN-001-019: MOAI_CONFIG_STRICT=1 시 오류, 미설정 시 경고.
	if driftErr := detectSchemaDrift(data, path); driftErr != nil {
		return nil, driftErr
	}

	// 3단계: evaluator.memory_scope FROZEN 검증 (HRN-002 substrate 유지)
	// design-constitution §11.4.1 (SPEC-V3R2-HRN-002) 참조
	if cfg.Evaluator.MemoryScope == "" {
		return nil, &ValidationError{
			Field:   "evaluator.memory_scope",
			Message: "required field missing; must be 'per_iteration'",
			Wrapped: ErrEvalMemoryFrozen,
		}
	}
	if cfg.Evaluator.MemoryScope != "per_iteration" {
		return nil, &ValidationError{
			Field:   "evaluator.memory_scope",
			Message: "value is FROZEN at 'per_iteration' per design-constitution §11.4.1",
			Value:   cfg.Evaluator.MemoryScope,
			Wrapped: ErrEvalMemoryFrozen,
		}
	}

	// 4단계: level enum 검증 (FROZEN: {minimal, standard, thorough})
	// REQ-HRN-001-017.
	for levelName := range cfg.Levels {
		if !validHarnessLevels[levelName] {
			return nil, &ValidationError{
				Field:   fmt.Sprintf("levels.%s", levelName),
				Message: fmt.Sprintf("unknown level %q; valid values are {minimal, standard, thorough} (FROZEN per SPEC-V3R2-HRN-001 REQ-017)", levelName),
				Value:   levelName,
				Wrapped: ErrUnknownLevel,
			}
		}
	}

	return cfg, nil
}

// detectSchemaDrift는 harness.yaml의 harness 블록에서 알 수 없는 최상위 키를 탐지합니다.
// REQ-HRN-001-019: MOAI_CONFIG_STRICT=1 시 ErrSchemaDrift 반환, 미설정 시 slog.Warn.
//
// @MX:WARN: [AUTO] FROZEN-zone 검증 — MOAI_CONFIG_STRICT=1 시 알 수 없는 키는 오류로 처리
// @MX:REASON: REQ-HRN-001-019, design-constitution §5 스키마 드리프트 감지
func detectSchemaDrift(data []byte, path string) error {
	// 알 수 없는 키 수집을 위해 map으로 파싱
	var rawWrapper map[string]map[string]any
	if err := yaml.Unmarshal(data, &rawWrapper); err != nil {
		// 파싱 오류 시 drift 감지 생략 (기존 오류로 처리)
		return nil
	}

	harnessBlock, ok := rawWrapper["harness"]
	if !ok {
		return nil
	}

	var unknownKeys []string
	for key := range harnessBlock {
		if !knownHarnessTopLevelKeys[key] {
			unknownKeys = append(unknownKeys, key)
		}
	}

	if len(unknownKeys) == 0 {
		return nil
	}

	// MOAI_CONFIG_STRICT=1 시 오류 반환
	strict := os.Getenv("MOAI_CONFIG_STRICT")
	msg := fmt.Sprintf("HRN_SCHEMA_DRIFT: unknown keys in harness.yaml: %v (path: %s)", unknownKeys, path)

	if strict == "1" {
		return &ValidationError{
			Field:   "harness",
			Message: msg,
			Value:   unknownKeys,
			Wrapped: ErrSchemaDrift,
		}
	}

	// 비-strict 모드: 경고만 출력
	slog.Warn(msg, "unknown_keys", unknownKeys, "path", path)
	return nil
}

// loadYAMLFile reads a YAML file from the given directory and unmarshals it
// into the target struct. Returns (true, nil) if the file was found and parsed,
// (false, nil) if the file does not exist, or (false, error) on failure.
func loadYAMLFile(dir, filename string, target any) (bool, error) {
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("parse %s: %w", filename, ErrInvalidYAML)
	}

	return true, nil
}
