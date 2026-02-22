package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
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

// @MX:ANCHOR: [AUTO] 품질 게이트 설정을 YAML에서 로드합니다. 모든 훅이 이 함수를 통해 설정을 로드합니다.
// @MX:REASON: [AUTO] fan_in=5+, 훅 설정 로드의 유일한 진입점입니다
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

// qualityYAMLConfig represents the quality.yaml file structure.
type qualityYAMLConfig struct {
	Constitution struct {
		LSPQualityGates struct {
			Enabled bool `yaml:"enabled"`
			Run     struct {
				MaxErrors       int  `yaml:"max_errors"`
				MaxTypeErrors   int  `yaml:"max_type_errors"`
				MaxLintErrors   int  `yaml:"max_lint_errors"`
				AllowRegression bool `yaml:"allow_regression"`
			} `yaml:"run"`
			Sync struct {
				MaxErrors       int  `yaml:"max_errors"`
				MaxWarnings     int  `yaml:"max_warnings"`
				RequireCleanLSP bool `yaml:"require_clean_lsp"`
			} `yaml:"sync"`
		} `yaml:"lsp_quality_gates"`
	} `yaml:"constitution"`
}

// parseQualityConfig parses quality configuration from YAML data.
func parseQualityConfig(data []byte) (QualityGate, error) {
	var config qualityYAMLConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return defaultQualityGate(), err
	}

	gates := config.Constitution.LSPQualityGates
	if !gates.Enabled {
		return defaultQualityGate(), nil
	}

	// Use run phase settings by default (strictest)
	gate := QualityGate{
		MaxErrors:      gates.Run.MaxErrors,
		MaxWarnings:    gates.Sync.MaxWarnings,
		BlockOnError:   true,
		BlockOnWarning: false, // Warnings don't block by default per REQ-HOOK-182
	}

	return gate, nil
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

// Compile-time interface compliance check.
var _ QualityGateEnforcer = (*qualityGateEnforcer)(nil)
