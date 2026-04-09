package hook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook/lifecycle"
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
	projectDir := input.CWD
	if projectDir == "" {
		projectDir = input.ProjectDir
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

	// Stop hooks use top-level decision/reason fields per Claude Code protocol
	// Return empty JSON {} to allow Claude to stop (default behavior)
	// To keep Claude working, return: {"decision": "block", "reason": "..."}
	return &HookOutput{}, nil
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
