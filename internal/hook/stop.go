package hook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook/lifecycle"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// defaultCompletionMarkers is the list of default completion markers.
// These are markers that Claude includes in output when a task is complete.
var defaultCompletionMarkers = []string{
	"<moai>DONE</moai>",
	"<moai>COMPLETE</moai>",
}

// stopHandler processes Stop events.
// It performs graceful shutdown, saves in-progress work state, and preserves
// loop controller (Ralph) state (REQ-HOOK-035). Always returns "allow".
type stopHandler struct {
	// completionMarkers is the list of completion markers to detect in ToolOutput.
	completionMarkers []string
}

// NewStopHandler creates a new Stop event handler.
func NewStopHandler() Handler {
	return &stopHandler{completionMarkers: defaultCompletionMarkers}
}

// NewStopHandlerWithMarkers creates a Stop event handler with custom completion markers.
// If markers is nil or an empty slice, completion marker detection is skipped.
func NewStopHandlerWithMarkers(markers []string) Handler {
	return &stopHandler{completionMarkers: markers}
}

// EventType returns EventStop.
func (h *stopHandler) EventType() EventType {
	return EventStop
}

// Handle processes a Stop event. It logs the stop request, preserves
// any active state, and returns an appropriate response.
//
// Per Claude Code protocol:
// - Return empty JSON {} to allow Claude to stop
// - Return {"decision": "block", "reason": "..."} to keep Claude working
// - Check stop_hook_active to prevent infinite loops
//
// Errors are non-blocking: the handler logs warnings and returns empty output.
func (h *stopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("stop requested",
		"session_id", input.SessionID,
		"stop_hook_active", input.StopHookActive,
	)

	// IMPORTANT: Prevent infinite loop per Claude Code protocol
	// If stop_hook_active is true, Claude is already continuing due to a previous
	// stop hook decision. Allow Claude to stop to prevent infinite loops.
	if input.StopHookActive {
		slog.Debug("stop_hook_active is true, allowing Claude to stop")
		return &HookOutput{}, nil
	}

	// Check persistent mode — non-blocking on errors
	projectDir := input.ProjectDir
	if projectDir == "" {
		projectDir = input.CWD
	}
	if projectDir != "" {
		mode, err := lifecycle.CheckPersistentMode(projectDir)
		if err != nil {
			slog.Warn("persistent mode check failed", "error", err)
		} else if mode != nil && mode.Active {
			if h.hasCompletionMarker(input) {
				slog.Info("completion marker detected during persistent mode, deactivating")
				if err := lifecycle.DeactivatePersistentMode(projectDir); err != nil {
					slog.Warn("failed to deactivate persistent mode", "error", err)
				}
			} else if mode.IsExpired() {
				slog.Info("persistent mode expired", "max_minutes", mode.MaxDurationMinutes)
				if err := lifecycle.DeactivatePersistentMode(projectDir); err != nil {
					slog.Warn("failed to deactivate persistent mode", "error", err)
				}
			} else {
				slog.Info("persistent mode active, blocking stop",
					"workflow", mode.Workflow, "spec_id", mode.SpecID)
				return &HookOutput{
					Decision: "block",
					Reason: fmt.Sprintf("Persistent mode active: %s workflow on %s. Continuing work...",
						mode.Workflow, mode.SpecID),
				}, nil
			}
		}
	}

	// Prune telemetry files older than 90 days (SPEC-TELEMETRY-001 R4).
	// Best-effort: errors are logged and never propagated.
	// Placed before completion detection to ensure pruning runs every session end.
	if projectDir != "" {
		if pruneErr := telemetry.PruneOldFiles(projectDir, 90); pruneErr != nil {
			slog.Warn("stop: telemetry pruning failed", "error", pruneErr)
		}
	}

	// Detect completion markers in ToolOutput (observation-only, never blocks)
	if len(input.ToolOutput) > 0 && len(h.completionMarkers) > 0 {
		output := string(input.ToolOutput)
		for _, marker := range h.completionMarkers {
			if strings.Contains(output, marker) {
				slog.Info("completion marker detected",
					"marker", marker,
					"session_id", input.SessionID,
				)
				break
			}
		}
	}

	// Reflective learning: analyze session telemetry and generate proposals.
	// Non-blocking — analysis errors are logged and never affect the stop decision.
	if projectDir != "" {
		numRecords := countSessionRecords(projectDir, input.SessionID)
		if numRecords >= minToolInvocationsForReflection {
			AnalyzeSessionAndLog(projectDir, input.SessionID)
		}
	}

	// Prune old telemetry files (keep 30 days).
	if projectDir != "" {
		telemetry.PruneOldFiles(projectDir, 30)
	}

	// Stop hooks use top-level decision/reason fields per Claude Code protocol
	// Return empty JSON {} to allow Claude to stop (default behavior)
	// To keep Claude working, return: {"decision": "block", "reason": "..."}
	return &HookOutput{}, nil
}

// minToolInvocationsForReflection is the minimum number of telemetry records
// required before reflective analysis is triggered.  Mirrors
// evolution.MinToolInvocationsForAnalysis but avoids a circular import.
const minToolInvocationsForReflection = 3

// countSessionRecords returns the number of telemetry records for the session.
// Returns 0 on any error (conservative).
func countSessionRecords(projectRoot, sessionID string) int {
	records, err := telemetry.LoadBySession(projectRoot, sessionID)
	if err != nil {
		return 0
	}
	return len(records)
}

// hasCompletionMarker reports whether the input's ToolOutput contains any of the
// configured completion markers.
func (h *stopHandler) hasCompletionMarker(input *HookInput) bool {
	if len(input.ToolOutput) == 0 || len(h.completionMarkers) == 0 {
		return false
	}
	output := string(input.ToolOutput)
	for _, marker := range h.completionMarkers {
		if strings.Contains(output, marker) {
			return true
		}
	}
	return false
}
