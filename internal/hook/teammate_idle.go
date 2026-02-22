package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"gopkg.in/yaml.v3"
)

// teammateIdleHandler processes TeammateIdle events.
// In team mode, it enforces quality gates before accepting idle state.
type teammateIdleHandler struct{}

// NewTeammateIdleHandler creates a new TeammateIdle event handler.
func NewTeammateIdleHandler() Handler {
	return &teammateIdleHandler{}
}

// EventType returns EventTeammateIdle.
func (h *teammateIdleHandler) EventType() EventType {
	return EventTeammateIdle
}

// Handle processes a TeammateIdle event.
// Returns empty output (exit code 0) to accept idle.
// Returns NewTeammateKeepWorkingOutput() (exit code 2) to keep working.
//
// Quality gate enforcement only applies in team mode (TeamName non-empty).
// Graceful degradation: if config or baseline cannot be loaded, idle is accepted.
func (h *teammateIdleHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("teammate idle",
		"session_id", input.SessionID,
		"teammate", input.TeammateName,
		"team", input.TeamName,
	)

	// Only enforce quality gates in team mode.
	if input.TeamName == "" {
		return &HookOutput{}, nil
	}

	// Resolve project directory.
	projectDir := input.ProjectDir
	if projectDir == "" {
		projectDir = input.CWD
	}
	if projectDir == "" {
		slog.Info("teammate_idle: no project dir, allowing idle")
		return &HookOutput{}, nil
	}

	// Load quality gate configuration.
	enforcer := lsphook.NewQualityGateEnforcer(projectDir)
	gate, err := enforcer.LoadConfig()
	if err != nil {
		// Graceful degradation: no config means no enforcement.
		slog.Info("teammate_idle: failed to load quality config, allowing idle", "error", err)
		return &HookOutput{}, nil
	}

	// Skip enforcement if gate is not configured to block on errors.
	if !gate.BlockOnError {
		return &HookOutput{}, nil
	}

	// Load the diagnostics baseline to get total error counts.
	counts, err := loadBaselineCounts(projectDir)
	if err != nil {
		// No baseline yet: allow idle (first run, nothing to check).
		slog.Info("teammate_idle: no diagnostics baseline, allowing idle", "error", err)
		return &HookOutput{}, nil
	}

	// Check quality gate.
	if enforcer.ShouldBlock(counts, gate) {
		msg := fmt.Sprintf(
			"Quality gate failed for teammate %q: %s\nFix errors before going idle.",
			input.TeammateName,
			lsphook.FormatGateResult(counts, gate),
		)
		fmt.Fprint(os.Stderr, msg)
		slog.Warn("teammate_idle: blocking idle - quality gate failed",
			"teammate", input.TeammateName,
			"errors", counts.Errors,
			"max_errors", gate.MaxErrors,
		)
		return NewTeammateKeepWorkingOutput(), nil
	}

	// Check test coverage gate.
	if coveragePct, ok := loadCoverageData(projectDir); ok {
		threshold := loadCoverageThreshold(projectDir)
		if coveragePct < threshold {
			msg := fmt.Sprintf(
				"Coverage gate failed for teammate %q: %.1f%% < %.1f%% required. Run tests to improve coverage.",
				input.TeammateName, coveragePct, threshold,
			)
			fmt.Fprint(os.Stderr, msg)
			slog.Warn("teammate_idle: blocking idle - coverage gate failed",
				"teammate", input.TeammateName,
				"coverage", coveragePct,
				"threshold", threshold,
			)
			return NewTeammateKeepWorkingOutput(), nil
		}
	}

	return &HookOutput{}, nil
}

// defaultCoverageThreshold is used when quality.yaml does not specify test_coverage_target.
const defaultCoverageThreshold = 85.0

// coverageData represents the JSON structure of .moai/memory/coverage.json.
type coverageData struct {
	CoveragePercent float64 `json:"coverage_percent"`
	UpdatedAt       string  `json:"updated_at"`
}

// coverageThresholdConfig represents the subset of quality.yaml needed for coverage threshold.
type coverageThresholdConfig struct {
	Constitution struct {
		TestCoverageTarget float64 `yaml:"test_coverage_target"`
	} `yaml:"constitution"`
}

// loadCoverageData reads coverage data from .moai/memory/coverage.json.
// Returns (percent, true) on success, or (0, false) if the file does not exist or cannot be parsed.
func loadCoverageData(projectDir string) (float64, bool) {
	coverageFile := filepath.Join(projectDir, defs.MoAIDir, defs.MemorySubdir, "coverage.json")
	data, err := os.ReadFile(coverageFile)
	if err != nil {
		slog.Info("teammate_idle: no coverage data, skipping coverage check", "error", err)
		return 0, false
	}

	var cov coverageData
	if err := json.Unmarshal(data, &cov); err != nil {
		slog.Info("teammate_idle: failed to parse coverage data, skipping coverage check", "error", err)
		return 0, false
	}
	return cov.CoveragePercent, true
}

// loadCoverageThreshold reads the test_coverage_target from quality.yaml.
// Returns defaultCoverageThreshold (85.0) if the file cannot be read or parsed.
func loadCoverageThreshold(projectDir string) float64 {
	configPath := filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir, defs.QualityYAML)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultCoverageThreshold
	}

	var cfg coverageThresholdConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultCoverageThreshold
	}

	if cfg.Constitution.TestCoverageTarget <= 0 {
		return defaultCoverageThreshold
	}
	return cfg.Constitution.TestCoverageTarget
}

// loadBaselineCounts reads the diagnostics baseline file and sums error counts
// across all tracked files.
func loadBaselineCounts(projectDir string) (lsphook.SeverityCounts, error) {
	baselineFile := filepath.Join(projectDir, ".moai", "memory", lsphook.BaselineFileName)
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		return lsphook.SeverityCounts{}, err
	}

	var baseline struct {
		Files map[string]struct {
			Diagnostics []struct {
				Severity string `json:"severity"`
			} `json:"diagnostics"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		return lsphook.SeverityCounts{}, fmt.Errorf("parse baseline: %w", err)
	}

	var total lsphook.SeverityCounts
	for _, fb := range baseline.Files {
		for _, d := range fb.Diagnostics {
			switch lsphook.DiagnosticSeverity(d.Severity) {
			case lsphook.SeverityError:
				total.Errors++
			case lsphook.SeverityWarning:
				total.Warnings++
			case lsphook.SeverityInformation:
				total.Information++
			case lsphook.SeverityHint:
				total.Hints++
			}
		}
	}
	return total, nil
}
