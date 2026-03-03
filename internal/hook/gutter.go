package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// gutterThreshold is the number of times the same error pattern must occur
// to trigger gutter detection.
const gutterThreshold = 3

// errorPattern tracks a single error pattern occurrence.
type errorPattern struct {
	Count     int    `json:"count"`
	LastError string `json:"last_error"`
	FirstSeen string `json:"first_seen"`
	LastSeen  string `json:"last_seen"`
}

// errorTrackerState represents the persistent state stored in .moai/state/error-tracker.json.
type errorTrackerState struct {
	SessionID      string                    `json:"session_id"`
	Patterns       map[string]*errorPattern  `json:"patterns"`
	GutterDetected bool                      `json:"gutter_detected"`
	TotalFailures  int                       `json:"total_failures"`
}

// errorLogEntry is a single log entry written to .moai/logs/errors.log (JSONL format).
type errorLogEntry struct {
	Timestamp    string `json:"timestamp"`
	SessionID    string `json:"session_id"`
	ToolName     string `json:"tool_name"`
	Error        string `json:"error"`
	PatternCount int    `json:"pattern_count"`
	Gutter       bool   `json:"gutter"`
}

// GutterTracker tracks error patterns and detects gutter state.
type GutterTracker struct {
	projectRoot string
	sessionID   string
	state       *errorTrackerState
}

// NewGutterTracker creates a new GutterTracker for the given project root.
func NewGutterTracker(projectRoot, sessionID string) *GutterTracker {
	return &GutterTracker{
		projectRoot: projectRoot,
		sessionID:   sessionID,
		state: &errorTrackerState{
			SessionID:      sessionID,
			Patterns:       make(map[string]*errorPattern),
			GutterDetected: false,
			TotalFailures:  0,
		},
	}
}

// TrackFailure processes a tool failure and returns whether gutter was detected.
func (gt *GutterTracker) TrackFailure(toolName, errorMsg string) (bool, error) {
	// Load existing state
	if err := gt.loadState(); err != nil {
		slog.Warn("gutter tracker: failed to load state, starting fresh",
			"session_id", gt.sessionID,
			"error", err,
		)
		// Continue with empty state
	}

	// Update session ID if changed (session resumed)
	gt.state.SessionID = gt.sessionID

	// Generate pattern signature
	signature := gt.patternSignature(toolName, errorMsg)

	// Update pattern count
	pattern, exists := gt.state.Patterns[signature]
	now := time.Now().UTC().Format(time.RFC3339)

	if !exists {
		gt.state.Patterns[signature] = &errorPattern{
			Count:     1,
			LastError: errorMsg,
			FirstSeen: now,
			LastSeen:  now,
		}
	} else {
		pattern.Count++
		pattern.LastError = errorMsg
		pattern.LastSeen = now
	}

	gt.state.TotalFailures++

	// Check for gutter detection
	currentPattern := gt.state.Patterns[signature]
	gutterDetected := currentPattern.Count >= gutterThreshold

	if gutterDetected && !gt.state.GutterDetected {
		gt.state.GutterDetected = true
		slog.Info("gutter detected",
			"session_id", gt.sessionID,
			"tool_name", toolName,
			"pattern", signature,
			"count", currentPattern.Count,
		)
	}

	// Save updated state
	if err := gt.saveState(); err != nil {
		slog.Warn("gutter tracker: failed to save state",
			"session_id", gt.sessionID,
			"error", err,
		)
	}

	// Append to error log
	if err := gt.appendErrorLog(toolName, errorMsg, currentPattern.Count, gutterDetected); err != nil {
		slog.Warn("gutter tracker: failed to append error log",
			"session_id", gt.sessionID,
			"error", err,
		)
	}

	return gutterDetected, nil
}

// patternSignature generates a unique signature for an error pattern.
// Uses tool_name + ":" + first 100 characters of error message.
func (gt *GutterTracker) patternSignature(toolName, errorMsg string) string {
	// Truncate error message to first 100 characters
	maxLen := 100
	if len(errorMsg) > maxLen {
		errorMsg = errorMsg[:maxLen]
	}
	return toolName + ":" + errorMsg
}

// statePath returns the path to the error tracker state file.
func (gt *GutterTracker) statePath() string {
	return filepath.Join(gt.projectRoot, defs.MoAIDir, defs.StateSubdir, defs.ErrorTrackerJSON)
}

// errorLogPath returns the path to the error log file.
func (gt *GutterTracker) errorLogPath() string {
	return filepath.Join(gt.projectRoot, defs.MoAIDir, defs.LogsSubdir, defs.ErrorsLog)
}

// loadState loads the error tracker state from disk.
func (gt *GutterTracker) loadState() error {
	path := gt.statePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No existing state, use defaults
			return nil
		}
		return fmt.Errorf("read state file: %w", err)
	}

	var state errorTrackerState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parse state file: %w", err)
	}

	// Initialize patterns map if nil
	if state.Patterns == nil {
		state.Patterns = make(map[string]*errorPattern)
	}

	gt.state = &state
	return nil
}

// saveState saves the error tracker state to disk.
func (gt *GutterTracker) saveState() error {
	path := gt.statePath()
	dir := filepath.Dir(path)

	// Create state directory if needed
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(gt.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// appendErrorLog appends an error entry to the error log file.
func (gt *GutterTracker) appendErrorLog(toolName, errorMsg string, patternCount int, gutter bool) error {
	path := gt.errorLogPath()
	dir := filepath.Dir(path)

	// Create logs directory if needed
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}

	// Create log entry
	entry := errorLogEntry{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		SessionID:    gt.sessionID,
		ToolName:     toolName,
		Error:        errorMsg,
		PatternCount: patternCount,
		Gutter:       gutter,
	}

	// Marshal to JSON
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}
	line = append(line, '\n')

	// Append to log file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("write log entry: %w", err)
	}

	return nil
}

// GetSystemMessageForGutter returns the system message to display when gutter is detected.
func GetSystemMessageForGutter(toolName string, count int) string {
	return fmt.Sprintf(
		"[MoAI Gutter Detection] Same error pattern detected %d times for %s. Context may be degraded. Auto-compact will manage context recovery. Consider trying a different approach.",
		count,
		toolName,
	)
}

// ResolveGutterTracker creates a GutterTracker for the given hook input.
// Returns nil if no valid project root is found.
func ResolveGutterTracker(input *HookInput) *GutterTracker {
	projectRoot := resolveProjectRoot(input)
	if projectRoot == "" {
		slog.Debug("gutter tracker: no MoAI project root found",
			"session_id", input.SessionID,
			"cwd", input.CWD,
		)
		return nil
	}

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	return NewGutterTracker(projectRoot, sessionID)
}

// IsGutterRelevantError checks if an error should be tracked for gutter detection.
// Interrupts and user cancellations are not tracked as they don't indicate context degradation.
func IsGutterRelevantError(input *HookInput) bool {
	// Skip tracking for interrupts (user cancelled)
	if input.IsInterrupt {
		return false
	}

	// Skip tracking for empty error messages
	return strings.TrimSpace(input.Error) != ""
}
