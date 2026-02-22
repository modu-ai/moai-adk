package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/pkg/models"
	"gopkg.in/yaml.v3"
)

// managerState represents the lifecycle state of the ConfigManager.
type managerState int

const (
	stateUninitialized managerState = iota
	stateInitialized
	stateWatching
)

// @MX:ANCHOR: [AUTO] ConfigManager는 스레드 안전한 설정 관리를 제공하는 핵심 구조체입니다. Load() 호출 후 사용해야 합니다.
// @MX:REASON: fan_in=12+, 모든 설정 접근의 진입점이며 시스템 전체에서 사용됩니다
// ConfigManager provides thread-safe configuration management.
// It must be initialized via Load() before use.
type ConfigManager struct {
	mu             sync.RWMutex
	config         *Config
	root           string
	state          managerState
	loader         *Loader
	callbacks      []func(Config)
	loadedSections map[string]bool
}

// NewConfigManager creates a new ConfigManager instance in uninitialized state.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		loader: NewLoader(),
		state:  stateUninitialized,
	}
}

// @MX:NOTE: [AUTO] 파일 값, 컴파일된 기본값, 환경 변수 우선순위를 병합합니다. MOAI_CONFIG_DIR 환경 변수로 설정 디렉토리를 재정의할 수 있습니다.
// Load reads configuration from the project root's .moai/ directory.
// It merges file values with compiled defaults and applies environment
// variable overrides. The configuration is validated before being stored.
func (m *ConfigManager) Load(projectRoot string) (*Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	configDir := filepath.Join(filepath.Clean(projectRoot), defs.MoAIDir)

	// Support MOAI_CONFIG_DIR environment variable override
	if envDir := os.Getenv("MOAI_CONFIG_DIR"); envDir != "" {
		configDir = filepath.Clean(envDir)
	}

	cfg, err := m.loader.Load(configDir)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Track which sections were loaded from files
	m.loadedSections = m.loader.LoadedSections()

	// Apply environment variable overrides (higher priority than files)
	applyEnvOverrides(cfg)

	// Validate the merged configuration
	if err := Validate(cfg, m.loadedSections); err != nil {
		return nil, err
	}

	m.config = cfg
	m.root = projectRoot
	m.state = stateInitialized

	return cfg, nil
}

// Get returns the current in-memory configuration.
// Returns nil if the manager has not been initialized via Load().
func (m *ConfigManager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetSection returns a named configuration section.
// Returns ErrNotInitialized if Load() has not been called.
// Returns ErrSectionNotFound if the section name is invalid.
func (m *ConfigManager) GetSection(name string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.state == stateUninitialized {
		return nil, ErrNotInitialized
	}

	return m.getSectionLocked(name)
}

// SetSection updates a named configuration section in memory.
// Returns ErrNotInitialized if Load() has not been called.
// Returns ErrSectionNotFound if the section name is invalid.
// Returns ErrSectionTypeMismatch if the value type does not match.
func (m *ConfigManager) SetSection(name string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == stateUninitialized {
		return ErrNotInitialized
	}

	return m.setSectionLocked(name, value)
}

// Save persists the current configuration to disk atomically.
// Each section is saved to its corresponding YAML file using
// temp file + os.Rename for atomic writes.
// Returns ErrNotInitialized if Load() has not been called.
func (m *ConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == stateUninitialized {
		return ErrNotInitialized
	}

	sectionsDir := filepath.Join(filepath.Clean(m.root), defs.MoAIDir, defs.SectionsSubdir)

	// Ensure directory exists
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Save user section
	if err := saveSection(sectionsDir, "user.yaml", userFileWrapper{User: m.config.User}); err != nil {
		return fmt.Errorf("save user config: %w", err)
	}

	// Save language section
	if err := saveSection(sectionsDir, "language.yaml", languageFileWrapper{Language: m.config.Language}); err != nil {
		return fmt.Errorf("save language config: %w", err)
	}

	// Save quality section
	if err := saveSection(sectionsDir, "quality.yaml", qualityFileWrapper{Constitution: m.config.Quality}); err != nil {
		return fmt.Errorf("save quality config: %w", err)
	}

	// Save git convention section
	if err := saveSection(sectionsDir, "git-convention.yaml", gitConventionFileWrapper{GitConvention: m.config.GitConvention}); err != nil {
		return fmt.Errorf("save git convention config: %w", err)
	}

	// Save LLM section
	if err := saveSection(sectionsDir, "llm.yaml", llmFileWrapper{LLM: m.config.LLM}); err != nil {
		return fmt.Errorf("save LLM config: %w", err)
	}

	return nil
}

// Reload forces a re-read from disk, replacing the in-memory configuration.
// Returns ErrNotInitialized if Load() has not been called.
func (m *ConfigManager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == stateUninitialized {
		return ErrNotInitialized
	}

	configDir := filepath.Join(filepath.Clean(m.root), defs.MoAIDir)
	if envDir := os.Getenv("MOAI_CONFIG_DIR"); envDir != "" {
		configDir = filepath.Clean(envDir)
	}

	cfg, err := m.loader.Load(configDir)
	if err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	m.loadedSections = m.loader.LoadedSections()
	applyEnvOverrides(cfg)

	if err := Validate(cfg, m.loadedSections); err != nil {
		return err
	}

	m.config = cfg

	// Notify registered callbacks
	for _, cb := range m.callbacks {
		cb(*m.config)
	}

	return nil
}

// Watch registers a callback to be invoked when configuration is reloaded.
// Returns ErrNotInitialized if Load() has not been called.
func (m *ConfigManager) Watch(callback func(Config)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == stateUninitialized {
		return ErrNotInitialized
	}

	m.callbacks = append(m.callbacks, callback)
	m.state = stateWatching
	return nil
}

// getSectionLocked returns a section by name. Caller must hold at least RLock.
func (m *ConfigManager) getSectionLocked(name string) (any, error) {
	switch name {
	case "user":
		return m.config.User, nil
	case "language":
		return m.config.Language, nil
	case "quality":
		return m.config.Quality, nil
	case "project":
		return m.config.Project, nil
	case "git_strategy":
		return m.config.GitStrategy, nil
	case "git_convention":
		return m.config.GitConvention, nil
	case "system":
		return m.config.System, nil
	case "llm":
		return m.config.LLM, nil
	case "pricing":
		return m.config.Pricing, nil
	case "ralph":
		return m.config.Ralph, nil
	case "workflow":
		return m.config.Workflow, nil
	default:
		return nil, ErrSectionNotFound
	}
}

// setSectionLocked updates a section by name. Caller must hold Lock.
func (m *ConfigManager) setSectionLocked(name string, value any) error {
	switch name {
	case "user":
		v, ok := value.(models.UserConfig)
		if !ok {
			return fmt.Errorf("%w: expected UserConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.User = v
	case "language":
		v, ok := value.(models.LanguageConfig)
		if !ok {
			return fmt.Errorf("%w: expected LanguageConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Language = v
	case "quality":
		v, ok := value.(models.QualityConfig)
		if !ok {
			return fmt.Errorf("%w: expected QualityConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Quality = v
	case "project":
		v, ok := value.(models.ProjectConfig)
		if !ok {
			return fmt.Errorf("%w: expected ProjectConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Project = v
	case "git_strategy":
		v, ok := value.(GitStrategyConfig)
		if !ok {
			return fmt.Errorf("%w: expected GitStrategyConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.GitStrategy = v
	case "git_convention":
		v, ok := value.(models.GitConventionConfig)
		if !ok {
			return fmt.Errorf("%w: expected GitConventionConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.GitConvention = v
	case "system":
		v, ok := value.(SystemConfig)
		if !ok {
			return fmt.Errorf("%w: expected SystemConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.System = v
	case "llm":
		v, ok := value.(LLMConfig)
		if !ok {
			return fmt.Errorf("%w: expected LLMConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.LLM = v
	case "pricing":
		v, ok := value.(PricingConfig)
		if !ok {
			return fmt.Errorf("%w: expected PricingConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Pricing = v
	case "ralph":
		v, ok := value.(RalphConfig)
		if !ok {
			return fmt.Errorf("%w: expected RalphConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Ralph = v
	case "workflow":
		v, ok := value.(WorkflowConfig)
		if !ok {
			return fmt.Errorf("%w: expected WorkflowConfig for section %q", ErrSectionTypeMismatch, name)
		}
		m.config.Workflow = v
	default:
		return ErrSectionNotFound
	}
	return nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables have higher priority than file-based values.
func applyEnvOverrides(cfg *Config) {
	if mode := os.Getenv("MOAI_DEVELOPMENT_MODE"); mode != "" {
		cfg.Quality.DevelopmentMode = models.DevelopmentMode(mode)
	}
	if level := os.Getenv("MOAI_LOG_LEVEL"); level != "" {
		cfg.System.LogLevel = level
	}
	if format := os.Getenv("MOAI_LOG_FORMAT"); format != "" {
		cfg.System.LogFormat = format
	}
	if noColor := os.Getenv("MOAI_NO_COLOR"); noColor == "true" || noColor == "1" {
		cfg.System.NoColor = true
	}
}

// saveSection marshals data to YAML and writes it atomically.
func saveSection(dir, filename string, data any) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filename, err)
	}

	path := filepath.Join(dir, filename)
	return atomicWrite(path, yamlData)
}

// atomicWrite writes data to a file atomically using temp file + os.Rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".moai-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // cleanup on error path

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	return os.Rename(tmpName, path)
}
