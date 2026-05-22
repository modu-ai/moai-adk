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
// Injects the ralph.stale_seconds key in ralph.yaml into Config.Session.StaleSeconds.
// SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS defaults to 3600 and is overridable from ralph.yaml.
func (l *Loader) loadRalphSection(dir string, cfg *Config) {
	wrapper := &ralphFileWrapper{}
	// Initialize defaults for ralph.yaml (inline field)
	wrapper.Ralph.RalphConfig = cfg.Ralph
	loaded, err := loadYAMLFile(dir, "ralph.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load ralph config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Ralph = wrapper.Ralph.RalphConfig
		// Only override when stale_seconds is non-zero (0 is treated as no explicit setting)
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

// validHarnessLevels is the FROZEN level enum.
// REQ-HRN-001-017: levels outside {minimal, standard, thorough} return ErrUnknownLevel.
var validHarnessLevels = map[string]bool{
	"minimal":  true,
	"standard": true,
	"thorough": true,
}

// knownHarnessTopLevelKeys is the set of recognized top-level keys under harnessFileWrapper.Harness.
// REQ-HRN-001-019: reference set for detecting unknown keys.
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
	"learning":             true, // Legacy sub-system (out-of-scope, warning suppressed)
}

// LoadHarnessConfig reads the harness.yaml file at the given path and returns a HarnessConfig.
//
// HRN-001 run-phase: evaluator.memory_scope FROZEN validation + full schema parsing.
// HRN-001 extended validation:
//   - level enum {minimal, standard, thorough} FROZEN (REQ-HRN-001-017)
//   - With MOAI_CONFIG_STRICT=1, unknown keys → ErrSchemaDrift error (REQ-HRN-001-019)
//   - Without MOAI_CONFIG_STRICT, unknown keys → slog.Warn (REQ-HRN-001-019)
//
// @MX:ANCHOR: [AUTO] harness.yaml loader function — many callers (router, CLI, ConfigManager.Reload)
// @MX:REASON: fan_in >= 3: called by HarnessRouter.Route, CLI validate, and ConfigManager.Reload
func LoadHarnessConfig(path string) (*HarnessConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("LoadHarnessConfig: %w", ErrConfigNotFound)
		}
		return nil, fmt.Errorf("LoadHarnessConfig read %s: %w", path, err)
	}

	// Step 1: unmarshal into the struct
	var wrapper harnessFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadHarnessConfig parse %s: %w", path, ErrInvalidYAML)
	}

	cfg := &wrapper.Harness

	// Step 2: detect schema drift (unknown top-level keys)
	// REQ-HRN-001-019: error under MOAI_CONFIG_STRICT=1, warning otherwise.
	if driftErr := detectSchemaDrift(data, path); driftErr != nil {
		return nil, driftErr
	}

	// Step 3: evaluator.memory_scope FROZEN validation (preserve HRN-002 substrate)
	// See design-constitution §11.4.1 (SPEC-V3R2-HRN-002)
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

	// Step 4: validate level enum (FROZEN: {minimal, standard, thorough})
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

// detectSchemaDrift detects unknown top-level keys inside the harness block of harness.yaml.
// REQ-HRN-001-019: returns ErrSchemaDrift under MOAI_CONFIG_STRICT=1, slog.Warn otherwise.
//
// @MX:WARN: [AUTO] FROZEN-zone validation — under MOAI_CONFIG_STRICT=1, unknown keys are treated as errors
// @MX:REASON: REQ-HRN-001-019, design-constitution §5 schema drift detection
func detectSchemaDrift(data []byte, path string) error {
	// Parse into a map to collect unknown keys
	var rawWrapper map[string]map[string]any
	if err := yaml.Unmarshal(data, &rawWrapper); err != nil {
		// On parse error, skip drift detection (treat as a pre-existing error)
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

	// Return an error under MOAI_CONFIG_STRICT=1
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

	// Non-strict mode: emit a warning only
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
