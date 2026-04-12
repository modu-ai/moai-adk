package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/pkg/models"
	"gopkg.in/yaml.v3"
)

const (
	// ExitCodeSuccess indicates no quality gate violations.
	ExitCodeSuccess = 0

	// ExitCodeQualityGateFailed indicates quality gate violations per REQ-HOOK-181.
	ExitCodeQualityGateFailed = 2
)

// qualityGateEnforcer implements QualityGateEnforcer interface.
// It enforces quality gate rules per REQ-HOOK-180 through REQ-HOOK-182.
type qualityGateEnforcer struct {
	projectDir string
	configPath string
}

// NewQualityGateEnforcer creates a new quality gate enforcer.
// projectDir is the project root directory containing .moai/config/.
func NewQualityGateEnforcer(projectDir string) *qualityGateEnforcer {
	return &qualityGateEnforcer{
		projectDir: projectDir,
		configPath: filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir, defs.QualityYAML),
	}
}

// ShouldBlock determines if the counts exceed gate thresholds per REQ-HOOK-180.
// Returns true if execution should be blocked.
func (e *qualityGateEnforcer) ShouldBlock(counts SeverityCounts, gate QualityGate) bool {
	// Check error threshold per REQ-HOOK-181
	if gate.BlockOnError && counts.Errors > gate.MaxErrors {
		return true
	}

	// Check warning threshold per REQ-HOOK-182
	// Warnings only block if BlockOnWarning is true
	if gate.BlockOnWarning && counts.Warnings > gate.MaxWarnings {
		return true
	}

	return false
}

// WarningsExceedThreshold checks if warnings exceed the threshold.
// This is separate from ShouldBlock because warnings may exceed threshold
// without blocking (per REQ-HOOK-182: log warning but continue).
func (e *qualityGateEnforcer) WarningsExceedThreshold(counts SeverityCounts, gate QualityGate) bool {
	return counts.Warnings > gate.MaxWarnings
}

// GetExitCode returns the appropriate exit code based on counts and gate.
// Per REQ-HOOK-181: exit code 2 when errors exceed threshold.
func (e *qualityGateEnforcer) GetExitCode(counts SeverityCounts, gate QualityGate) int {
	if e.ShouldBlock(counts, gate) {
		return ExitCodeQualityGateFailed
	}
	return ExitCodeSuccess
}

// @MX:ANCHOR: [AUTO] Loads quality gate settings from YAML. All hooks load configuration through this function.
// @MX:REASON: [AUTO] fan_in=5+, sole entry point for hook configuration loading
// LoadConfig loads quality gate configuration from YAML per REQ-HOOK-180.
func (e *qualityGateEnforcer) LoadConfig() (QualityGate, error) {
	data, err := os.ReadFile(e.configPath)
	if err != nil {
		// Return sensible defaults if config not found
		return defaultQualityGate(), nil
	}

	return parseQualityConfig(data)
}

// CheckWithConfig loads config and checks if should block.
func (e *qualityGateEnforcer) CheckWithConfig(counts SeverityCounts) (shouldBlock bool, gate QualityGate, err error) {
	gate, err = e.LoadConfig()
	if err != nil {
		return false, gate, err
	}

	shouldBlock = e.ShouldBlock(counts, gate)
	return shouldBlock, gate, nil
}

// qualityFileWrapper matches the quality.yaml top-level structure.
// Used internally when reading the file without the full ConfigManager stack.
type qualityFileWrapper struct {
	Constitution models.QualityConfig `yaml:"constitution"`
}

// parseQualityConfig parses quality configuration from YAML data.
// REQ-QG-006: uses pkg/models.QualityConfig instead of the previously-duplicate struct.
func parseQualityConfig(data []byte) (QualityGate, error) {
	var wrapper qualityFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return defaultQualityGate(), err
	}

	gates := wrapper.Constitution.LSPQualityGates
	if !gates.Enabled {
		return defaultQualityGate(), nil
	}

	// Use run phase settings by default (strictest), with sync max_warnings for backward compat.
	gate := QualityGate{
		MaxErrors:      gates.Run.MaxErrors,
		MaxWarnings:    gates.Sync.MaxWarnings,
		BlockOnError:   true,
		BlockOnWarning: false, // Warnings don't block by default per REQ-HOOK-182
	}

	return gate, nil
}

// parseFullQualityConfig parses the full quality config from YAML data using pkg/models (REQ-QG-006).
func parseFullQualityConfig(data []byte) (models.QualityConfig, error) {
	var wrapper qualityFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return models.QualityConfig{}, fmt.Errorf("parse quality config: %w", err)
	}
	return wrapper.Constitution, nil
}

// defaultQualityGate returns sensible default quality gate settings.
func defaultQualityGate() QualityGate {
	return QualityGate{
		MaxErrors:      0,  // Zero tolerance for errors
		MaxWarnings:    10, // Allow some warnings
		BlockOnError:   true,
		BlockOnWarning: false,
	}
}

// FormatGateResult formats the quality gate check result as a human-readable string.
func FormatGateResult(counts SeverityCounts, gate QualityGate) string {
	var sb strings.Builder

	sb.WriteString("Quality Gate Check:\n")
	fmt.Fprintf(&sb, "  Errors: %d (max: %d)", counts.Errors, gate.MaxErrors)

	if counts.Errors > gate.MaxErrors {
		sb.WriteString(" [EXCEEDED]")
	}
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "  Warnings: %d (max: %d)", counts.Warnings, gate.MaxWarnings)
	if counts.Warnings > gate.MaxWarnings {
		sb.WriteString(" [EXCEEDED]")
	}
	sb.WriteString("\n")

	if counts.Errors > gate.MaxErrors && gate.BlockOnError {
		sb.WriteString("  Status: BLOCKED - error threshold exceeded\n")
	} else if counts.Warnings > gate.MaxWarnings && gate.BlockOnWarning {
		sb.WriteString("  Status: BLOCKED - warning threshold exceeded\n")
	} else if counts.Warnings > gate.MaxWarnings {
		sb.WriteString("  Status: WARNING - warning threshold exceeded (continuing)\n")
	} else {
		sb.WriteString("  Status: PASSED\n")
	}

	return sb.String()
}

// EnforcePhase evaluates the quality gate for the given phase (REQ-QG-001 through REQ-QG-004).
// baseline is the prior diagnostic snapshot used for regression detection (REQ-QG-007).
// Pass nil baseline to skip regression comparison.
//
// @MX:ANCHOR: [AUTO] EnforcePhase is the phase-aware gate enforcement entry point
// @MX:REASON: fan_in >= 3 — run workflow, sync workflow, and plan workflow all call this to validate LSP state per phase
// @MX:NOTE: [AUTO] Phase dispatch: plan captures baseline, run enforces error/type/lint thresholds, sync enforces max_warnings and require_clean_lsp
func (e *qualityGateEnforcer) EnforcePhase(phase PhaseType, counts SeverityCounts, baseline *SeverityCounts) (PhaseEnforceResult, error) {
	cfg, err := e.loadModelsConfig()
	if err != nil {
		return PhaseEnforceResult{Phase: phase}, fmt.Errorf("enforce phase: %w", err)
	}

	gates := cfg.LSPQualityGates
	if !gates.Enabled {
		return PhaseEnforceResult{Phase: phase}, nil
	}

	// Resolve effective phase: auto falls back to run.
	effective := phase
	if effective == PhaseAuto {
		effective = PhaseRun
	}

	result := PhaseEnforceResult{Phase: effective}

	switch effective {
	case PhasePlan:
		// REQ-QG-002: capture baseline if require_baseline is true.
		result.BaselineCaptured = gates.Plan.RequireBaseline

	case PhaseRun:
		// REQ-QG-003: enforce max_errors, max_type_errors, max_lint_errors, allow_regression.
		if counts.Errors > gates.Run.MaxErrors {
			result.ShouldBlock = true
			result.ViolatedThreshold = "max_errors"
		}
		if counts.TypeErrors > gates.Run.MaxTypeErrors {
			result.ShouldBlock = true
			result.ViolatedThreshold = "max_type_errors"
		}
		if counts.LintErrors > gates.Run.MaxLintErrors {
			result.ShouldBlock = true
			result.ViolatedThreshold = "max_lint_errors"
		}

		// REQ-QG-007: regression detection.
		if baseline != nil {
			rd := cfg.LSPIntegration.RegressionDetection
			if DetectRegression(counts, *baseline, rd.ErrorIncreaseThreshold, rd.WarningIncreaseThreshold) {
				result.HasRegression = true
				if !gates.Run.AllowRegression {
					result.ShouldBlock = true
					if result.ViolatedThreshold == "" {
						result.ViolatedThreshold = "regression"
					}
				}
			}
		}

	case PhaseSync:
		// REQ-QG-004: enforce max_errors, max_warnings, require_clean_lsp.
		if counts.Errors > gates.Sync.MaxErrors {
			result.ShouldBlock = true
			result.ViolatedThreshold = "max_errors"
		}
		if counts.Warnings > gates.Sync.MaxWarnings {
			result.ShouldBlock = true
			result.ViolatedThreshold = "max_warnings"
		}
		if gates.Sync.RequireCleanLSP && counts.Total() > 0 {
			result.ShouldBlock = true
			if result.ViolatedThreshold == "" {
				result.ViolatedThreshold = "require_clean_lsp"
			}
		}
	}

	return result, nil
}

// LoadPhaseAwareConfig loads all phase-specific config fields using pkg/models (REQ-QG-005, REQ-QG-006).
func (e *qualityGateEnforcer) LoadPhaseAwareConfig() (LSPQualityGatesConfig, error) {
	cfg, err := e.loadModelsConfig()
	if err != nil {
		return LSPQualityGatesConfig{}, err
	}

	g := cfg.LSPQualityGates
	return LSPQualityGatesConfig{
		Enabled: g.Enabled,
		Plan: LSPPlanGateConfig{
			RequireBaseline: g.Plan.RequireBaseline,
		},
		Run: LSPRunGateConfig{
			MaxErrors:       g.Run.MaxErrors,
			MaxTypeErrors:   g.Run.MaxTypeErrors,
			MaxLintErrors:   g.Run.MaxLintErrors,
			AllowRegression: g.Run.AllowRegression,
		},
		Sync: LSPSyncGateConfig{
			MaxErrors:       g.Sync.MaxErrors,
			MaxWarnings:     g.Sync.MaxWarnings,
			RequireCleanLSP: g.Sync.RequireCleanLSP,
		},
		CacheTTLSeconds: g.CacheTTLSeconds,
		TimeoutSeconds:  g.TimeoutSeconds,
	}, nil
}

// LoadTRUST5Config loads the trust5_integration dimension mapping (REQ-QG-009).
//
// @MX:NOTE: [AUTO] trust5_integration maps TRUST 5 dimensions to LSP diagnostic checks; influences scoring of tested/readable/understandable/secured/trackable dimensions
func (e *qualityGateEnforcer) LoadTRUST5Config() (TRUST5Config, error) {
	cfg, err := e.loadModelsConfig()
	if err != nil {
		return TRUST5Config{}, err
	}

	t5 := cfg.LSPIntegration.TRUST5Integration
	return TRUST5Config{
		Tested:         t5.Tested,
		Readable:       t5.Readable,
		Understandable: t5.Understandable,
		Secured:        t5.Secured,
		Trackable:      t5.Trackable,
	}, nil
}

// loadModelsConfig reads and parses quality.yaml returning models.QualityConfig (REQ-QG-005, REQ-QG-006).
// This replaces direct yaml reads via the removed qualityYAMLConfig duplicate struct.
func (e *qualityGateEnforcer) loadModelsConfig() (models.QualityConfig, error) {
	data, err := os.ReadFile(e.configPath)
	if err != nil {
		// Return sensible defaults if config not found.
		return defaultModelsConfig(), nil
	}
	return parseFullQualityConfig(data)
}

// defaultModelsConfig returns a sensible default models.QualityConfig.
func defaultModelsConfig() models.QualityConfig {
	return models.QualityConfig{
		LSPQualityGates: models.LSPQualityGates{
			Enabled: false,
		},
	}
}

// DetectRegression returns true when counts represent a regression against baseline
// per the provided increase thresholds (REQ-QG-007).
//
// errorThreshold: maximum allowed increase in error count before it is a regression.
// warnThreshold: maximum allowed increase in warning count before it is a regression.
func DetectRegression(current, baseline SeverityCounts, errorThreshold, warnThreshold int) bool {
	errorIncrease := current.Errors - baseline.Errors
	if errorIncrease > errorThreshold {
		return true
	}
	warnIncrease := current.Warnings - baseline.Warnings
	if warnIncrease > warnThreshold {
		return true
	}
	return false
}

// Compile-time interface compliance check.
var _ QualityGateEnforcer = (*qualityGateEnforcer)(nil)
