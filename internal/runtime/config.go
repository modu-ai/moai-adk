// Package runtime provides the Token Circuit Breaker for MoAI-ADK.
//
// SPEC-V3R3-ARCH-007: per-agent token budget tracking + stall detection.
// Warning-first policy (BC-V3R3-006); will switch to hard-fail in P5.
// /clear is never triggered automatically (HARD constraint per MEMORY.md).
package runtime

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Default threshold and budget constants.
// These are used when runtime.yaml is missing (REQ-ARCH007-011).
const (
	// DefaultPreClearThreshold is the 75% soft warning threshold.
	DefaultPreClearThreshold = 0.75
	// DefaultHardClearThreshold is the 90% hard recommendation threshold.
	DefaultHardClearThreshold = 0.90
	// DefaultBudget is the token budget for agents not listed in per_agent_budget.
	DefaultBudget = 30000
	// DefaultStallDetectionSeconds is the stall window in seconds.
	DefaultStallDetectionSeconds = 60
	// DefaultRetryMax is the max stall retry count before fallback recommendation.
	DefaultRetryMax = 3
	// DefaultFallback is the recommended action when retry_max is exhausted.
	DefaultFallback = "split_into_waves"
)

// RuntimeConfig holds the parsed runtime.yaml configuration.
// All fields correspond to the schema defined in SPEC-V3R3-ARCH-007 REQ-ARCH007-001.
type RuntimeConfig struct {
	// PreClearThreshold is the ratio at which a warning is emitted and progress.md is saved.
	PreClearThreshold float64
	// HardClearThreshold is the ratio at which a strong /clear recommendation is emitted.
	HardClearThreshold float64
	// PerAgentBudget maps agent name to total token budget (input + output).
	// The "default" key is used for unlisted agents.
	PerAgentBudget map[string]int
	// StallDetectionSeconds is the window without a RecordCall before stall is declared.
	StallDetectionSeconds int
	// RetryMax is the number of stall events before fallback recommendation.
	RetryMax int
	// Fallback is the recommended action when retry_max is exhausted.
	Fallback string
	// AutoSaveAtThreshold enables automatic progress.md save at PreClearThreshold.
	AutoSaveAtThreshold bool
	// SavePathTemplate is the template for the progress.md path.
	// {SPEC_ID} is replaced with the actual SPEC ID at runtime.
	SavePathTemplate string
	// ResumeMessageFormat is the paste-ready resume message template.
	ResumeMessageFormat string
}

// runtimeYAML is the intermediate struct for YAML unmarshalling.
// It mirrors the runtime.yaml schema.
type runtimeYAML struct {
	Runtime struct {
		ContextWindow struct {
			PreClearThreshold  float64 `yaml:"pre_clear_threshold"`
			HardClearThreshold float64 `yaml:"hard_clear_threshold"`
		} `yaml:"context_window"`
		PerAgentBudget     map[string]int `yaml:"per_agent_budget"`
		CircuitBreaker     struct {
			StallDetectionSeconds int    `yaml:"stall_detection_seconds"`
			RetryMax              int    `yaml:"retry_max"`
			Fallback              string `yaml:"fallback"`
		} `yaml:"circuit_breaker"`
		ProgressPersistence struct {
			AutoSaveAtThreshold bool   `yaml:"auto_save_at_threshold"`
			SavePathTemplate    string `yaml:"save_path_template"`
			ResumeMessageFormat string `yaml:"resume_message_format"`
		} `yaml:"progress_persistence"`
	} `yaml:"runtime"`
}

// LoadRuntime reads and parses a runtime.yaml file at the given path.
// Returns an error if the file does not exist or cannot be parsed.
// Use DefaultRuntimeConfig() as a fallback per REQ-ARCH007-011.
func LoadRuntime(path string) (*RuntimeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw runtimeYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	cfg := DefaultRuntimeConfig()

	r := raw.Runtime
	if r.ContextWindow.PreClearThreshold > 0 {
		cfg.PreClearThreshold = r.ContextWindow.PreClearThreshold
	}
	if r.ContextWindow.HardClearThreshold > 0 {
		cfg.HardClearThreshold = r.ContextWindow.HardClearThreshold
	}
	if len(r.PerAgentBudget) > 0 {
		cfg.PerAgentBudget = r.PerAgentBudget
	}
	if r.CircuitBreaker.StallDetectionSeconds > 0 {
		cfg.StallDetectionSeconds = r.CircuitBreaker.StallDetectionSeconds
	}
	if r.CircuitBreaker.RetryMax > 0 {
		cfg.RetryMax = r.CircuitBreaker.RetryMax
	}
	if r.CircuitBreaker.Fallback != "" {
		cfg.Fallback = r.CircuitBreaker.Fallback
	}
	cfg.AutoSaveAtThreshold = r.ProgressPersistence.AutoSaveAtThreshold
	if r.ProgressPersistence.SavePathTemplate != "" {
		cfg.SavePathTemplate = r.ProgressPersistence.SavePathTemplate
	}
	if r.ProgressPersistence.ResumeMessageFormat != "" {
		cfg.ResumeMessageFormat = r.ProgressPersistence.ResumeMessageFormat
	}

	return cfg, nil
}

// DefaultRuntimeConfig returns a RuntimeConfig with built-in default values.
// Used as fallback when runtime.yaml is missing (REQ-ARCH007-011).
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		PreClearThreshold:     DefaultPreClearThreshold,
		HardClearThreshold:    DefaultHardClearThreshold,
		PerAgentBudget:        defaultPerAgentBudget(),
		StallDetectionSeconds: DefaultStallDetectionSeconds,
		RetryMax:              DefaultRetryMax,
		Fallback:              DefaultFallback,
		AutoSaveAtThreshold:   true,
		SavePathTemplate:      ".moai/specs/{SPEC_ID}/progress.md",
		ResumeMessageFormat:   "ultrathink. {wave_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",
	}
}

// defaultPerAgentBudget returns the built-in per-agent token budget map.
func defaultPerAgentBudget() map[string]int {
	return map[string]int{
		"default":            DefaultBudget,
		"manager-strategy":   60000,
		"manager-spec":       40000,
		"expert-backend":     40000,
		"expert-frontend":    40000,
		"expert-security":    40000,
		"expert-testing":     40000,
		"expert-debug":       40000,
		"expert-performance": 40000,
		"expert-refactoring": 40000,
		"expert-devops":      40000,
		"expert-mobile":      40000,
		"evaluator-active":   20000,
		"plan-auditor":       20000,
	}
}
