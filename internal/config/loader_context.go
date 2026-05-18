package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadContextConfig reads context.yaml from the given file path and
// returns a typed ContextConfig.
//
// Return semantics:
//   - (cfg, nil) on successful parse
//   - (defaultCfg, nil) on file-not-found (REQ-MIG003-004 graceful default)
//   - (nil, ErrInvalidYAML) on malformed YAML
//
// @MX:ANCHOR: [AUTO] LoadContextConfig — public loader for context_search section, consumed by Loader.Load() and CLAUDE.md §16 runtime
// @MX:REASON: fan_in >= 3: loadContextSection (Loader.Load chain), standalone CLI consumer, CLAUDE.md §16 Context Search consumer
// REQ-MIG003-002/003/004/010.
func LoadContextConfig(path string) (*ContextConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultContextConfig()
			return &defaults, nil
		}
		return nil, fmt.Errorf("LoadContextConfig read %s: %w", path, err)
	}

	var wrapper contextFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadContextConfig parse %s: %w", path, ErrInvalidYAML)
	}

	return &wrapper.ContextSearch, nil
}

// loadContextSection loads the context_search section via the loadYAMLFile helper.
// Note: YAML file is "context.yaml" but top-level key is "context_search:" (research.md §3.2).
func (l *Loader) loadContextSection(dir string, cfg *Config) {
	wrapper := &contextFileWrapper{ContextSearch: cfg.ContextSearch}
	loaded, err := loadYAMLFile(dir, "context.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load context config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.ContextSearch = wrapper.ContextSearch
		l.loadedSections["context_search"] = true
	}
}
