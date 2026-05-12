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

// LoadHarnessConfig는 주어진 경로의 harness.yaml 파일을 읽어 HarnessConfig를 반환합니다.
// evaluator.memory_scope가 per_iteration이 아니거나 비어 있는 경우
// ErrEvalMemoryFrozen 또는 ErrInvalidConfig 오류를 반환합니다.
// HRN-002 run-phase minimal substrate — HRN-001 run-phase에서 routing/profile 확장 예정.
func LoadHarnessConfig(path string) (*HarnessConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("LoadHarnessConfig: %w", ErrConfigNotFound)
		}
		return nil, fmt.Errorf("LoadHarnessConfig read %s: %w", path, err)
	}

	var wrapper harnessFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadHarnessConfig parse %s: %w", path, ErrInvalidYAML)
	}

	cfg := &wrapper.Harness

	// evaluator.memory_scope는 per_iteration으로 FROZEN됩니다
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

	return cfg, nil
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
