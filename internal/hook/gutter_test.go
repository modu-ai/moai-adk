package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestGutterTracker_NewTracker verifies that a new tracker is created with clean state.
func TestGutterTracker_NewTracker(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-1"

	tracker := NewGutterTracker(projectRoot, sessionID)

	if tracker == nil {
		t.Fatal("NewGutterTracker returned nil")
	}
	if tracker.projectRoot != projectRoot {
		t.Errorf("projectRoot = %q, want %q", tracker.projectRoot, projectRoot)
	}
	if tracker.sessionID != sessionID {
		t.Errorf("sessionID = %q, want %q", tracker.sessionID, sessionID)
	}
	if tracker.state == nil {
		t.Fatal("state should not be nil")
	}
	if tracker.state.SessionID != sessionID {
		t.Errorf("state.SessionID = %q, want %q", tracker.state.SessionID, sessionID)
	}
	if tracker.state.GutterDetected {
		t.Error("GutterDetected should be false for new tracker")
	}
	if tracker.state.TotalFailures != 0 {
		t.Errorf("TotalFailures = %d, want 0", tracker.state.TotalFailures)
	}
	if len(tracker.state.Patterns) != 0 {
		t.Errorf("Patterns length = %d, want 0", len(tracker.state.Patterns))
	}
}

// TestGutterTracker_TrackFailure_SingleError verifies tracking a single error.
func TestGutterTracker_TrackFailure_SingleError(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-single"
	tracker := NewGutterTracker(projectRoot, sessionID)

	gutterDetected, err := tracker.TrackFailure("Edit", "file not found")
	if err != nil {
		t.Fatalf("TrackFailure error: %v", err)
	}
	if gutterDetected {
		t.Error("Gutter should not be detected after single error")
	}

	// Verify state
	if tracker.state.TotalFailures != 1 {
		t.Errorf("TotalFailures = %d, want 1", tracker.state.TotalFailures)
	}

	signature := tracker.patternSignature("Edit", "file not found")
	pattern, exists := tracker.state.Patterns[signature]
	if !exists {
		t.Fatal("Pattern not found in state")
	}
	if pattern.Count != 1 {
		t.Errorf("Pattern count = %d, want 1", pattern.Count)
	}
}

// TestGutterTracker_TrackFailure_DifferentErrors verifies tracking multiple different errors.
func TestGutterTracker_TrackFailure_DifferentErrors(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-different"
	tracker := NewGutterTracker(projectRoot, sessionID)

	// Track different errors
	tools := []string{"Edit", "Read", "Bash"}
	errors := []string{"file not found", "permission denied", "command failed"}

	for i := range tools {
		gutterDetected, err := tracker.TrackFailure(tools[i], errors[i])
		if err != nil {
			t.Fatalf("TrackFailure %d error: %v", i, err)
		}
		if gutterDetected {
			t.Errorf("Gutter should not be detected after error %d", i)
		}
	}

	// Verify state
	if tracker.state.TotalFailures != 3 {
		t.Errorf("TotalFailures = %d, want 3", tracker.state.TotalFailures)
	}
	if len(tracker.state.Patterns) != 3 {
		t.Errorf("Patterns length = %d, want 3", len(tracker.state.Patterns))
	}
	if tracker.state.GutterDetected {
		t.Error("GutterDetected should be false for different errors")
	}
}

// TestGutterTracker_TrackFailure_SameErrorThreeTimes verifies gutter detection threshold.
func TestGutterTracker_TrackFailure_SameErrorThreeTimes(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-threshold"
	tracker := NewGutterTracker(projectRoot, sessionID)

	// Track same error 3 times
	errorMsg := "file not found: /path/to/file"
	for i := 0; i < 3; i++ {
		gutterDetected, err := tracker.TrackFailure("Edit", errorMsg)
		if err != nil {
			t.Fatalf("TrackFailure iteration %d error: %v", i, err)
		}
		// Gutter should be detected on the 3rd attempt
		expectedGutter := (i >= 2)
		if gutterDetected != expectedGutter {
			t.Errorf("Iteration %d: gutterDetected = %v, want %v", i, gutterDetected, expectedGutter)
		}
	}

	// Verify state
	signature := tracker.patternSignature("Edit", errorMsg)
	pattern, exists := tracker.state.Patterns[signature]
	if !exists {
		t.Fatal("Pattern not found in state")
	}
	if pattern.Count != 3 {
		t.Errorf("Pattern count = %d, want 3", pattern.Count)
	}
	if !tracker.state.GutterDetected {
		t.Error("GutterDetected should be true after 3 same errors")
	}
}

// TestGutterTracker_TrackFailure_SameErrorExceedsThreshold verifies gutter stays detected.
func TestGutterTracker_TrackFailure_SameErrorExceedsThreshold(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-exceed"
	tracker := NewGutterTracker(projectRoot, sessionID)

	errorMsg := "connection timeout"
	// Track same error 5 times
	for i := 0; i < 5; i++ {
		gutterDetected, err := tracker.TrackFailure("Bash", errorMsg)
		if err != nil {
			t.Fatalf("TrackFailure iteration %d error: %v", i, err)
		}
		if !gutterDetected && i >= 2 {
			t.Errorf("Iteration %d: gutterDetected should be true", i)
		}
	}

	// Verify state
	signature := tracker.patternSignature("Bash", errorMsg)
	pattern, exists := tracker.state.Patterns[signature]
	if !exists {
		t.Fatal("Pattern not found in state")
	}
	if pattern.Count != 5 {
		t.Errorf("Pattern count = %d, want 5", pattern.Count)
	}
	if !tracker.state.GutterDetected {
		t.Error("GutterDetected should remain true")
	}
}

// TestGutterTracker_LoadSave verifies state persistence roundtrip.
func TestGutterTracker_LoadSave(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-persist"

	// Create tracker and track some errors
	tracker1 := NewGutterTracker(projectRoot, sessionID)
	_, _ = tracker1.TrackFailure("Edit", "error 1")
	_, _ = tracker1.TrackFailure("Read", "error 2")
	_, _ = tracker1.TrackFailure("Edit", "error 1") // Same error twice

	// Create new tracker with same session ID - should load existing state
	tracker2 := NewGutterTracker(projectRoot, sessionID)
	if err := tracker2.loadState(); err != nil {
		t.Fatalf("loadState error: %v", err)
	}

	// Verify loaded state
	if tracker2.state.TotalFailures != 3 {
		t.Errorf("Loaded TotalFailures = %d, want 3", tracker2.state.TotalFailures)
	}
	if len(tracker2.state.Patterns) != 2 {
		t.Errorf("Loaded Patterns length = %d, want 2", len(tracker2.state.Patterns))
	}

	// Verify specific pattern
	sig := tracker2.patternSignature("Edit", "error 1")
	pattern, exists := tracker2.state.Patterns[sig]
	if !exists {
		t.Fatal("Edit:error 1 pattern not found in loaded state")
	}
	if pattern.Count != 2 {
		t.Errorf("Loaded pattern count = %d, want 2", pattern.Count)
	}
}

// TestGutterTracker_ErrorLogAppend verifies error log JSONL format.
func TestGutterTracker_ErrorLogAppend(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-log"
	tracker := NewGutterTracker(projectRoot, sessionID)

	// Track some errors
	_, _ = tracker.TrackFailure("Edit", "error 1")
	_, _ = tracker.TrackFailure("Bash", "error 2")

	// Read log file
	logPath := tracker.errorLogPath()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Read log file error: %v", err)
	}

	logContent := string(data)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")

	if len(lines) != 2 {
		t.Fatalf("Expected 2 log lines, got %d", len(lines))
	}

	// Verify JSONL format
	for i, line := range lines {
		if line == "" {
			continue
		}
		var entry errorLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d: invalid JSON: %v", i, err)
		}
		if entry.SessionID != sessionID {
			t.Errorf("Line %d: session_id = %q, want %q", i, entry.SessionID, sessionID)
		}
		if entry.ToolName == "" {
			t.Errorf("Line %d: tool_name should not be empty", i)
		}
		if entry.Error == "" {
			t.Errorf("Line %d: error should not be empty", i)
		}
	}
}

// TestGutterTracker_PatternSignature verifies signature generation.
func TestGutterTracker_PatternSignature(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	tracker := NewGutterTracker(projectRoot, "test-session")

	tests := []struct {
		name      string
		toolName  string
		errorMsg  string
		wantParts []string // Should contain these parts
	}{
		{
			name:      "simple case",
			toolName:  "Edit",
			errorMsg:  "file not found",
			wantParts: []string{"Edit:", "file not found"},
		},
		{
			name:      "different tool",
			toolName:  "Read",
			errorMsg:  "permission denied",
			wantParts: []string{"Read:", "permission denied"},
		},
		{
			name:      "long error truncated",
			toolName:  "Bash",
			errorMsg:  strings.Repeat("x", 150), // 150 chars
			wantParts: []string{"Bash:", strings.Repeat("x", 100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sig := tracker.patternSignature(tt.toolName, tt.errorMsg)

			for _, part := range tt.wantParts {
				if !strings.Contains(sig, part) {
					t.Errorf("Signature %q should contain %q", sig, part)
				}
			}
		})
	}
}

// TestGutterTracker_PatternSignature_Consistency verifies same error produces same signature.
func TestGutterTracker_PatternSignature_Consistency(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	tracker := NewGutterTracker(projectRoot, "test-session")

	errorMsg := "file not found: /path/to/file"
	sig1 := tracker.patternSignature("Edit", errorMsg)
	sig2 := tracker.patternSignature("Edit", errorMsg)

	if sig1 != sig2 {
		t.Errorf("Same error should produce same signature: %q != %q", sig1, sig2)
	}
}

// TestGutterTracker_Reset verifies state can be reset.
func TestGutterTracker_Reset(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	sessionID := "test-session-reset"
	tracker := NewGutterTracker(projectRoot, sessionID)

	// Trigger gutter
	errorMsg := "error"
	for i := 0; i < 3; i++ {
		_, _ = tracker.TrackFailure("Edit", errorMsg)
	}

	if !tracker.state.GutterDetected {
		t.Error("Gutter should be detected")
	}

	// Create new tracker (simulating reset)
	tracker2 := NewGutterTracker(projectRoot, sessionID+"-new")
	if tracker2.state.GutterDetected {
		t.Error("New tracker should not have gutter detected")
	}
	if tracker2.state.TotalFailures != 0 {
		t.Errorf("New tracker should have 0 failures, got %d", tracker2.state.TotalFailures)
	}
}

// TestIsGutterRelevantError verifies error filtering logic.
func TestIsGutterRelevantError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *HookInput
		expected bool
	}{
		{
			name: "normal error",
			input: &HookInput{
				Error:       "something went wrong",
				IsInterrupt: false,
			},
			expected: true,
		},
		{
			name: "interrupt error",
			input: &HookInput{
				Error:       "interrupted",
				IsInterrupt: true,
			},
			expected: false,
		},
		{
			name: "empty error",
			input: &HookInput{
				Error:       "",
				IsInterrupt: false,
			},
			expected: false,
		},
		{
			name: "whitespace only error",
			input: &HookInput{
				Error:       "   ",
				IsInterrupt: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsGutterRelevantError(tt.input)
			if result != tt.expected {
				t.Errorf("IsGutterRelevantError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetSystemMessageForGutter verifies system message format.
func TestGetSystemMessageForGutter(t *testing.T) {
	t.Parallel()

	msg := GetSystemMessageForGutter("Edit", 3)

	if !strings.Contains(msg, "Edit") {
		t.Errorf("SystemMessage should contain tool name 'Edit', got: %s", msg)
	}
	if !strings.Contains(msg, "3") {
		t.Errorf("SystemMessage should contain count '3', got: %s", msg)
	}
	if !strings.Contains(msg, "MoAI Gutter Detection") {
		t.Errorf("SystemMessage should contain 'MoAI Gutter Detection', got: %s", msg)
	}
	if !strings.Contains(msg, "Auto-compact") {
		t.Errorf("SystemMessage should mention 'Auto-compact', got: %s", msg)
	}
}

// TestResolveGutterTracker verifies tracker resolution logic.
func TestResolveGutterTracker(t *testing.T) {
	t.Parallel()

	// Test with no .moai directory
	input := &HookInput{
		SessionID: "test-session",
		CWD:       "/tmp/nonexistent",
	}
	tracker := ResolveGutterTracker(input)
	if tracker != nil {
		t.Error("ResolveGutterTracker should return nil when no .moai directory exists")
	}

	// Test with valid .moai directory
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, defs.MoAIDir)
	if err := os.Mkdir(moaiDir, 0o755); err != nil {
		t.Fatalf("Failed to create .moai directory: %v", err)
	}

	input2 := &HookInput{
		SessionID: "test-session-2",
		CWD:       tmpDir,
	}
	tracker2 := ResolveGutterTracker(input2)
	if tracker2 == nil {
		t.Fatal("ResolveGutterTracker should return tracker when .moai directory exists")
	}
	if tracker2.sessionID != "test-session-2" {
		t.Errorf("sessionID = %q, want 'test-session-2'", tracker2.sessionID)
	}
}

// TestGutterTracker_statePath verifies state file path generation.
func TestGutterTracker_statePath(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	tracker := NewGutterTracker(projectRoot, "test-session")

	expectedPath := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir, defs.ErrorTrackerJSON)
	actualPath := tracker.statePath()

	if actualPath != expectedPath {
		t.Errorf("statePath() = %q, want %q", actualPath, expectedPath)
	}
}

// TestGutterTracker_errorLogPath verifies error log path generation.
func TestGutterTracker_errorLogPath(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	tracker := NewGutterTracker(projectRoot, "test-session")

	expectedPath := filepath.Join(projectRoot, defs.MoAIDir, defs.LogsSubdir, defs.ErrorsLog)
	actualPath := tracker.errorLogPath()

	if actualPath != expectedPath {
		t.Errorf("errorLogPath() = %q, want %q", actualPath, expectedPath)
	}
}
