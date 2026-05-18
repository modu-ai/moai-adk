package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadDesignConfig reads design.yaml from the given file path and
// returns a typed DesignConfig.
//
// Return semantics:
//   - (cfg, nil) on successful parse + validation
//   - (defaultCfg, nil) on file-not-found (REQ-MIG003-004 graceful default)
//   - (nil, ErrInvalidYAML) on malformed YAML
//   - (nil, ErrPassThresholdFloor) if gan_loop.pass_threshold < 0.60 (OQ2 decision)
//
// @MX:ANCHOR: [AUTO] LoadDesignConfig — public loader for design section with pass_threshold floor validation
// @MX:REASON: fan_in >= 3: loadDesignSection (Loader.Load chain), standalone CLI consumer, GAN loop runtime sprint contract
// REQ-MIG003-002/003/004/014.
func LoadDesignConfig(path string) (*DesignConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultDesignConfig()
			return &defaults, nil
		}
		return nil, fmt.Errorf("LoadDesignConfig read %s: %w", path, err)
	}

	var wrapper designFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadDesignConfig parse %s: %w", path, ErrInvalidYAML)
	}

	cfg := &wrapper.Design

	// Validate pass_threshold FROZEN floor 0.60 (OQ2: YES, reuse ErrPassThresholdFloor).
	// Only validate when the field is explicitly set (non-zero).
	// REQ-MIG003-014, design-constitution §11 GAN Loop Contract.
	if cfg.GanLoop.PassThreshold > 0 && cfg.GanLoop.PassThreshold < 0.60 {
		return nil, &ValidationError{
			Field:   "gan_loop.pass_threshold",
			Value:   cfg.GanLoop.PassThreshold,
			Message: "below FROZEN floor 0.60 (design-constitution §11 GAN Loop Contract, SPEC-V3R2-CON-001)",
			Wrapped: ErrPassThresholdFloor,
		}
	}

	return cfg, nil
}

// loadDesignSection loads the design section via the loadYAMLFile helper.
func (l *Loader) loadDesignSection(dir string, cfg *Config) {
	wrapper := &designFileWrapper{Design: cfg.Design}
	loaded, err := loadYAMLFile(dir, "design.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load design config, using defaults", "error", err)
		return
	}
	if loaded {
		// Apply pass_threshold validation even in Loader.Load() path.
		if wrapper.Design.GanLoop.PassThreshold > 0 && wrapper.Design.GanLoop.PassThreshold < 0.60 {
			slog.Warn("design.yaml gan_loop.pass_threshold below FROZEN floor 0.60; using defaults",
				"value", wrapper.Design.GanLoop.PassThreshold)
			return
		}
		cfg.Design = wrapper.Design
		l.loadedSections["design"] = true
	}
}
