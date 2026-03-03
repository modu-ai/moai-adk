package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// preCompactSnapshot represents the state saved before context compaction.
type preCompactSnapshot struct {
	SessionID      string               `json:"session_id"`
	Timestamp      string               `json:"timestamp"`
	Trigger        string               `json:"trigger"`
	ErrorTracker   *errorTrackerState   `json:"error_tracker,omitempty"`
	GutterDetected bool                 `json:"gutter_detected"`
	Summary        string               `json:"summary"`
}

// compactHandler processes PreCompact events.
// It captures context information and creates session state snapshots
// for post-compaction recovery (REQ-HOOK-036). Always returns "allow".
type compactHandler struct{}

// NewCompactHandler creates a new PreCompact event handler.
func NewCompactHandler() Handler {
	return &compactHandler{}
}

// EventType returns EventPreCompact.
func (h *compactHandler) EventType() EventType {
	return EventPreCompact
}

// Handle processes a PreCompact event. It:
// 1. Loads the current error tracker state
// 2. Saves a pre-compact snapshot with critical state
// 3. Logs the compaction event
// 4. Returns a SystemMessage with context summary for Claude Code
func (h *compactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("pre-compact context preservation",
		"session_id", input.SessionID,
		"trigger", input.Trigger,
	)

	// Resolve project root
	projectRoot := resolveProjectRoot(input)
	if projectRoot == "" {
		slog.Debug("pre-compact: no MoAI project root found",
			"session_id", input.SessionID,
			"cwd", input.CWD,
		)
		// Return SystemMessage even without project root for context awareness
		return &HookOutput{
			SystemMessage: "[MoAI Pre-Compact] State preserved. No MoAI project root found.",
		}, nil
	}

	// Load error tracker state
	errorState, err := loadErrorTrackerState(projectRoot)
	if err != nil {
		slog.Warn("pre-compact: failed to load error tracker state",
			"session_id", input.SessionID,
			"error", err,
		)
		// Continue without error state
		errorState = nil
	}

	// Create pre-compact snapshot
	snapshot := h.createSnapshot(input.SessionID, input.Trigger, errorState)

	// Save snapshot
	if err := h.saveSnapshot(projectRoot, snapshot); err != nil {
		slog.Warn("pre-compact: failed to save snapshot",
			"session_id", input.SessionID,
			"error", err,
		)
	}

	// Log compaction event
	h.logCompactionEvent(projectRoot, snapshot)

	// Return SystemMessage with context summary
	return &HookOutput{
		SystemMessage: h.formatSystemMessage(snapshot),
	}, nil
}

// createSnapshot creates a pre-compact snapshot from the current state.
func (h *compactHandler) createSnapshot(sessionID, trigger string, errorState *errorTrackerState) *preCompactSnapshot {
	now := time.Now().UTC().Format(time.RFC3339)

	gutterDetected := false
	summary := "No error patterns tracked"

	if errorState != nil {
		gutterDetected = errorState.GutterDetected
		if errorState.TotalFailures > 0 {
			if gutterDetected {
				summary = fmt.Sprintf("%d tool failures tracked, gutter state detected", errorState.TotalFailures)
			} else {
				summary = fmt.Sprintf("%d tool failures tracked, no gutter state", errorState.TotalFailures)
			}
		}
	}

	return &preCompactSnapshot{
		SessionID:      sessionID,
		Timestamp:      now,
		Trigger:        trigger,
		ErrorTracker:   errorState,
		GutterDetected: gutterDetected,
		Summary:        summary,
	}
}

// saveSnapshot saves the pre-compact snapshot to disk.
func (h *compactHandler) saveSnapshot(projectRoot string, snapshot *preCompactSnapshot) error {
	path := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir, defs.PreCompactSnapshotJSON)
	dir := filepath.Dir(path)

	// Create state directory if needed
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write snapshot file: %w", err)
	}

	slog.Info("pre-compact snapshot saved",
		"session_id", snapshot.SessionID,
		"path", path,
	)

	return nil
}

// logCompactionEvent appends a compaction event to the error log.
func (h *compactHandler) logCompactionEvent(projectRoot string, snapshot *preCompactSnapshot) {
	logPath := filepath.Join(projectRoot, defs.MoAIDir, defs.LogsSubdir, defs.ErrorsLog)
	dir := filepath.Dir(logPath)

	// Create logs directory if needed
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Warn("pre-compact: failed to create logs directory",
			"error", err,
		)
		return
	}

	// Create log entry
	entry := map[string]any{
		"timestamp":       snapshot.Timestamp,
		"session_id":      snapshot.SessionID,
		"event":           "pre_compact",
		"trigger":         snapshot.Trigger,
		"gutter_detected": snapshot.GutterDetected,
		"summary":         snapshot.Summary,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		slog.Warn("pre-compact: failed to marshal log entry",
			"error", err,
		)
		return
	}
	line = append(line, '\n')

	// Append to log file
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn("pre-compact: failed to open log file",
			"error", err,
		)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		slog.Warn("pre-compact: failed to write log entry",
			"error", err,
		)
	}
}

// formatSystemMessage creates a SystemMessage for Claude Code with compact context.
func (h *compactHandler) formatSystemMessage(snapshot *preCompactSnapshot) string {
	return fmt.Sprintf("[MoAI Pre-Compact] State preserved. %s", snapshot.Summary)
}

// loadErrorTrackerState loads the error tracker state from disk.
func loadErrorTrackerState(projectRoot string) (*errorTrackerState, error) {
	path := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir, defs.ErrorTrackerJSON)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No error tracker state exists
			return nil, nil
		}
		return nil, fmt.Errorf("read error tracker state: %w", err)
	}

	var state errorTrackerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse error tracker state: %w", err)
	}

	return &state, nil
}
