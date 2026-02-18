package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
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

	return &HookOutput{}, nil
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
