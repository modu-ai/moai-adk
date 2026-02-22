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
