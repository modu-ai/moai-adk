package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadInterviewConfig reads interview.yaml from the given file path and
// returns a typed InterviewConfig.
//
// Return semantics:
//   - (cfg, nil) on successful parse
//   - (defaultCfg, nil) on file-not-found (REQ-MIG003-004 graceful default)
//   - (nil, ErrInvalidYAML) on malformed YAML
//
// @MX:ANCHOR: [AUTO] LoadInterviewConfig — public loader for interview section, consumed by Loader.Load() and SPEC-V3R2-WF-003 discovery mode
// @MX:REASON: fan_in >= 3: loadInterviewSection (Loader.Load chain), standalone CLI consumer, WF-003 discovery mode runtime
// REQ-MIG003-002/003/004/011.
func LoadInterviewConfig(path string) (*InterviewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := defaultInterviewConfig()
			return &defaults, nil
		}
		return nil, fmt.Errorf("LoadInterviewConfig read %s: %w", path, err)
	}

	var wrapper interviewFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("LoadInterviewConfig parse %s: %w", path, ErrInvalidYAML)
	}

	return &wrapper.Interview, nil
}

// loadInterviewSection loads the interview section via the loadYAMLFile helper.
func (l *Loader) loadInterviewSection(dir string, cfg *Config) {
	wrapper := &interviewFileWrapper{Interview: cfg.Interview}
	loaded, err := loadYAMLFile(dir, "interview.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load interview config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Interview = wrapper.Interview
		l.loadedSections["interview"] = true
	}
}
