package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConstitutionConfig reads constitution.yaml from the given file path and
// returns a typed ConstitutionConfig.
//
// Return semantics:
//   - (cfg, nil) on successful parse
//   - (defaultCfg, nil) on file-not-found (REQ-MIG003-004 graceful default)
//   - (nil, ErrInvalidYAML) on malformed YAML (REQ-MIG003-008)
//
// @MX:ANCHOR: [AUTO] LoadConstitutionConfig — public loader consumed by Loader.Load(), CLI validation, and SPEC-V3R2-EXT-004 framework hook
// @MX:REASON: fan_in >= 3: loadConstitutionSection (Loader.Load chain), standalone CLI consumer, EXT-004 enforcement hook
// REQ-MIG003-002/003/004/009.
func LoadConstitutionConfig(path string) (*ConstitutionConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultConstitutionConfig()
			return &defaults, nil
		}
		return nil, fmt.Errorf("LoadConstitutionConfig read %s: %w", path, err)
	}

	var wrapper constitutionFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadConstitutionConfig parse %s: %w", path, ErrInvalidYAML)
	}

	return &wrapper.Constitution, nil
}

// loadConstitutionSection loads the constitution section via the loadYAMLFile helper.
// On absent file, defaults from NewDefaultConfig() remain (REQ-MIG003-004).
func (l *Loader) loadConstitutionSection(dir string, cfg *Config) {
	wrapper := &constitutionFileWrapper{Constitution: cfg.Constitution}
	loaded, err := loadYAMLFile(dir, "constitution.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load constitution config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Constitution = wrapper.Constitution
		l.loadedSections["constitution"] = true
	}
}
