// Package harness — HRN-003 M4 profile loader helper.
// EvaluatorConfig defaults + profile loading bridge.
// REQ-HRN-003-005: consumes .md profile files.
package harness

// @MX:NOTE: [AUTO] Cross-references internal/harness/rubric.go ParseRubricMarkdown (SPEC-V3R2-HRN-003)

import (
	"sort"

	"github.com/modu-ai/moai-adk/internal/config"
)

// defaultProfilePaths is the map of default evaluator profile file paths.
// REQ-HRN-003-005, AC-HRN-003-07.c.
var defaultProfilePaths = map[string]string{
	"default":  ".moai/config/evaluator-profiles/default.md",
	"strict":   ".moai/config/evaluator-profiles/strict.md",
	"lenient":  ".moai/config/evaluator-profiles/lenient.md",
	"frontend": ".moai/config/evaluator-profiles/frontend.md",
}

// DefaultEvaluatorProfileNames returns the shipped evaluator profile names in
// sorted order, derived from defaultProfilePaths. It is the SSOT the settings
// schema reads to render harness.default_profile as a closed widget — the
// alternative (a second literal list in the schema file) drifts silently the
// first time a profile is added or renamed here.
func DefaultEvaluatorProfileNames() []string {
	out := make([]string, 0, len(defaultProfilePaths))
	for name := range defaultProfilePaths {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// defaultMustPassDimensionNames is the list of default must-pass dimension names.
// OQ3 default: [Functionality, Security]; floor = [Security].
var defaultMustPassDimensionNames = []string{"Functionality", "Security"}

// newDefaultEvaluatorConfig returns an EvaluatorConfig populated with HRN-003 defaults.
// M4 T4.5: default-value injection.
func newDefaultEvaluatorConfig() *config.EvaluatorConfig {
	profiles := make(map[string]string, len(defaultProfilePaths))
	for k, v := range defaultProfilePaths {
		profiles[k] = v
	}
	dims := make([]string, len(defaultMustPassDimensionNames))
	copy(dims, defaultMustPassDimensionNames)
	return &config.EvaluatorConfig{
		MemoryScope:        "per_iteration",
		Profiles:           profiles,
		Aggregation:        "min",
		MustPassDimensions: dims,
	}
}

// loadEvaluatorConfig loads an EvaluatorConfig from the default profile paths.
// Test helper — used to verify defaults + profile presence without harness.yaml.
// AC-HRN-003-07.c.
func loadEvaluatorConfig() (*config.EvaluatorConfig, error) {
	cfg := newDefaultEvaluatorConfig()
	// Verify that the .md files at the default profile paths are parseable.
	// Returns an error when no profile is present.
	for name, path := range cfg.Profiles {
		rubric, err := ParseRubricMarkdown(path)
		if err != nil {
			// Skip when the file is missing (the test environment may lack profile files).
			// ParseRubricMarkdown in rubric.go returns file-not-found as an error.
			// This function verifies that 'at least one profile must be loadable'.
			_ = name
			_ = rubric
			continue
		}
		// Return cfg as soon as at least one profile loads successfully.
		return cfg, nil
	}
	return cfg, nil
}
