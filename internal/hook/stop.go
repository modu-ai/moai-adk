// Resolution: KEEP — Ralph state update, telemetry pruning, reflective learning.
package hook

import (
	"context"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// stopHandler processes Stop events.
// It performs graceful shutdown, saves in-progress work state, and preserves
// loop controller (Ralph) state (REQ-HOOK-035). Always returns "allow".
type stopHandler struct{}

// NewStopHandler creates a new Stop event handler.
func NewStopHandler() Handler {
	return &stopHandler{}
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

	projectDir := input.ProjectDir
	if projectDir == "" {
		projectDir = input.CWD
	}

	// Prune telemetry files older than 90 days (SPEC-TELEMETRY-001 R4).
	// Best-effort: errors are logged and never propagated.
	if projectDir != "" {
		if pruneErr := telemetry.PruneOldFiles(projectDir, 90); pruneErr != nil {
			slog.Warn("stop: telemetry pruning failed", "error", pruneErr)
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
		_ = telemetry.PruneOldFiles(projectDir, 30)
	}

	// Evidence gate (SPEC-STOP-EVIDENCE-GATE-001, advisory-only): reads the
	// already-loadable session ledger and surfaces an unbacked success claim to
	// stderr. NEVER blocks stop (fail-open per REQ-SEG-005). Purely additive —
	// inserted after all pre-existing steps, before the final return (REQ-SEG-009).
	runEvidenceGate(projectDir, input.SessionID)

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
