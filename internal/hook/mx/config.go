package mx

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// ValidationConfig holds configuration for the MX validation system.
// It is parsed from the `validation` section of mx.yaml.
type ValidationConfig struct {
	// Enabled controls whether MX validation is active.
	Enabled bool `yaml:"enabled"`

	// PostToolUse configures the PostToolUse hook integration.
	PostToolUse PostToolUseConfig `yaml:"post_tool_use"`

	// SessionEnd configures the SessionEnd hook integration.
	SessionEnd SessionEndConfig `yaml:"session_end"`

	// Sync configures the sync workflow integration.
	Sync SyncConfig `yaml:"sync"`

	// EnforcementLevels configures per-priority enforcement modes.
	EnforcementLevels EnforcementLevels `yaml:"enforcement_levels"`

	// TransitionMode downgrades newly-surfaced violations (method-receiver P2/P3/P4,
	// paired-REASON P1/P2) from blocking to advisory for one minor cycle.
	// Default false. Set to true to enable a grace period during v2.14 → v2.15 migration.
	TransitionMode bool `yaml:"transition_mode,omitempty" json:"transition_mode,omitempty"`
}

// PostToolUseConfig configures the PostToolUse hook MX validation.
type PostToolUseConfig struct {
	// Enabled controls whether PostToolUse MX validation runs.
	Enabled bool `yaml:"enabled"`
	// TimeoutMs is the maximum milliseconds to spend on validation (default: 500).
	TimeoutMs int `yaml:"timeout_ms"`
}

// SessionEndConfig configures the SessionEnd hook MX validation.
type SessionEndConfig struct {
	// Enabled controls whether SessionEnd MX validation runs.
	Enabled bool `yaml:"enabled"`
	// TimeoutMs is the maximum milliseconds to spend on batch validation (default: 4000).
	TimeoutMs int `yaml:"timeout_ms"`
}

// SyncConfig configures the sync workflow MX validation.
type SyncConfig struct {
	// Enforcement is either "strict" (P1/P2 block sync) or "advisory" (warn only).
	Enforcement string `yaml:"enforcement"`
	// SkipFlag is the CLI flag that bypasses MX validation (default: "--skip-mx").
	SkipFlag string `yaml:"skip_flag"`
}

// EnforcementLevels defines blocking vs advisory per priority.
// Values are "blocking" or "advisory".
type EnforcementLevels struct {
	P1Anchor string `yaml:"p1_anchor"`
	P2Warn   string `yaml:"p2_warn"`
	P3Note   string `yaml:"p3_note"`
	P4Todo   string `yaml:"p4_todo"`
}

// mxYAML is the raw YAML structure for the entire mx.yaml file.
// Only the `validation` section is of interest here.
type mxYAML struct {
	MX struct {
		Validation *ValidationConfig `yaml:"validation"`
	} `yaml:"mx"`
}

// DefaultValidationConfig returns a ValidationConfig with sensible defaults.
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Enabled: true,
		PostToolUse: PostToolUseConfig{
			Enabled:   true,
			TimeoutMs: 500,
		},
		SessionEnd: SessionEndConfig{
			Enabled:   true,
			TimeoutMs: 4000,
		},
		Sync: SyncConfig{
			Enforcement: "strict",
			SkipFlag:    "--skip-mx",
		},
		EnforcementLevels: EnforcementLevels{
			P1Anchor: "blocking",
			P2Warn:   "blocking",
			P3Note:   "advisory",
			P4Todo:   "advisory",
		},
	}
}

// ParseValidationConfig reads mx.yaml and returns the parsed ValidationConfig.
// If the file is missing or has no validation section, defaults are returned.
// Non-nil errors are only returned for malformed YAML.
func ParseValidationConfig(mxYAMLPath string) (*ValidationConfig, error) {
	defaults := DefaultValidationConfig()

	data, err := os.ReadFile(mxYAMLPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults, nil
		}
		return defaults, nil
	}

	var raw mxYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return defaults, nil
	}

	if raw.MX.Validation == nil {
		return defaults, nil
	}

	cfg := raw.MX.Validation
	applyDefaults(cfg, defaults)
	return cfg, nil
}

// applyDefaults fills in zero-value fields in cfg from defaults.
func applyDefaults(cfg, defaults *ValidationConfig) {
	// PostToolUse defaults
	if cfg.PostToolUse.TimeoutMs == 0 {
		cfg.PostToolUse.TimeoutMs = defaults.PostToolUse.TimeoutMs
	}

	// SessionEnd defaults
	if cfg.SessionEnd.TimeoutMs == 0 {
		cfg.SessionEnd.TimeoutMs = defaults.SessionEnd.TimeoutMs
	}

	// Sync defaults
	if cfg.Sync.Enforcement == "" {
		cfg.Sync.Enforcement = defaults.Sync.Enforcement
	}
	if cfg.Sync.SkipFlag == "" {
		cfg.Sync.SkipFlag = defaults.Sync.SkipFlag
	}

	// EnforcementLevels defaults
	if cfg.EnforcementLevels.P1Anchor == "" {
		cfg.EnforcementLevels.P1Anchor = defaults.EnforcementLevels.P1Anchor
	}
	if cfg.EnforcementLevels.P2Warn == "" {
		cfg.EnforcementLevels.P2Warn = defaults.EnforcementLevels.P2Warn
	}
	if cfg.EnforcementLevels.P3Note == "" {
		cfg.EnforcementLevels.P3Note = defaults.EnforcementLevels.P3Note
	}
	if cfg.EnforcementLevels.P4Todo == "" {
		cfg.EnforcementLevels.P4Todo = defaults.EnforcementLevels.P4Todo
	}
}
