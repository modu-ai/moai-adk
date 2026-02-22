package tmux

import (
	"context"
	"slices"
	"strings"
	"testing"
)

// TestIntegration_TmuxSessionWithMultiplePanes validates the full session
// lifecycle with 4 panes and MaxVisible=2, verifying correct split directions
// (vertical for first MaxVisible panes, horizontal for overflow), send-keys
// for each pane, and final tiled layout selection.
func TestIntegration_TmuxSessionWithMultiplePanes(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "integration-4panes",
		Panes: []PaneConfig{
			{SpecID: "SPEC-ISSUE-1", Command: "moai worktree go SPEC-ISSUE-1"},
			{SpecID: "SPEC-ISSUE-2", Command: "moai worktree go SPEC-ISSUE-2"},
			{SpecID: "SPEC-ISSUE-3", Command: "moai worktree go SPEC-ISSUE-3"},
			{SpecID: "SPEC-ISSUE-4", Command: "moai worktree go SPEC-ISSUE-4"},
		},
		MaxVisible: 2,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify session result.
	if result.SessionName != "integration-4panes" {
		t.Errorf("SessionName = %q, want %q", result.SessionName, "integration-4panes")
	}
	if result.PaneCount != 4 {
		t.Errorf("PaneCount = %d, want 4", result.PaneCount)
	}

	// Verify split directions: with MaxVisible=2, pane 1 gets vertical,
	// panes 2 and 3 get horizontal.
	verticalSplits := countSplits(calls, "-v")
	horizontalSplits := countSplits(calls, "-h")
	if verticalSplits != 1 {
		t.Errorf("vertical splits = %d, want 1 (pane 1 only)", verticalSplits)
	}
	if horizontalSplits != 2 {
		t.Errorf("horizontal splits = %d, want 2 (panes 2 and 3)", horizontalSplits)
	}

	// Verify send-keys called for each pane (4 commands total).
	sendKeysCount := countCallsContaining(calls, "send-keys")
	if sendKeysCount != 4 {
		t.Errorf("send-keys count = %d, want 4", sendKeysCount)
	}

	// Verify tiled layout applied.
	assertCallContains(t, calls, "tiled")

	// Verify select-pane targets pane 0.
	selectPaneFound := false
	for _, call := range calls {
		for i, arg := range call {
			if arg == "select-pane" && i+2 < len(call) && call[i+2] == "integration-4panes:0.0" {
				selectPaneFound = true
			}
		}
	}
	if !selectPaneFound {
		t.Error("select-pane to 0.0 not found in calls")
	}
}

// TestIntegration_TmuxDetectorToSessionManager validates the detector->session
// workflow: first check tmux availability, then get version, then create a session.
func TestIntegration_TmuxDetectorToSessionManager(t *testing.T) {
	t.Parallel()

	var allCalls [][]string
	mockRunner := func(_ context.Context, name string, args ...string) (string, error) {
		call := append([]string{name}, args...)
		allCalls = append(allCalls, call)

		// Respond appropriately based on the command.
		if slices.Contains(args, "-V") {
			return "tmux 3.4", nil
		}
		return "", nil
	}

	// Step 1: Detect tmux availability.
	detector := NewDetector(WithRunFunc(mockRunner))
	if !detector.IsAvailable() {
		t.Fatal("IsAvailable() = false, want true")
	}

	// Step 2: Get version.
	version, err := detector.Version()
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if version != "3.4" {
		t.Errorf("Version() = %q, want %q", version, "3.4")
	}

	// Step 3: Create a session using the same runner.
	mgr := NewSessionManager(WithSessionRunFunc(mockRunner))
	cfg := &SessionConfig{
		Name: "detector-integration",
		Panes: []PaneConfig{
			{SpecID: "SPEC-ISSUE-10", Command: "moai run SPEC-ISSUE-10"},
			{SpecID: "SPEC-ISSUE-20", Command: "moai run SPEC-ISSUE-20"},
		},
		MaxVisible: 3,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.PaneCount != 2 {
		t.Errorf("PaneCount = %d, want 2", result.PaneCount)
	}

	// Verify the full workflow produced the expected call sequence:
	// 1. tmux -V (IsAvailable)
	// 2. tmux -V (Version)
	// 3. tmux new-session ... (Create)
	// 4+ tmux send-keys, split-window, ...
	if len(allCalls) < 3 {
		t.Fatalf("expected at least 3 calls, got %d", len(allCalls))
	}

	// First two calls should be version checks.
	for i := range 2 {
		hasVersion := false
		for _, arg := range allCalls[i] {
			if arg == "-V" {
				hasVersion = true
			}
		}
		if !hasVersion {
			t.Errorf("call %d should be version check (-V), got %v", i, allCalls[i])
		}
	}

	// Third call should be new-session.
	hasNewSession := false
	for _, arg := range allCalls[2] {
		if arg == "new-session" {
			hasNewSession = true
		}
	}
	if !hasNewSession {
		t.Errorf("call 2 should be new-session, got %v", allCalls[2])
	}

	// Verify send-keys contains actual SPEC commands.
	foundSpecCommand := false
	for _, call := range allCalls {
		for _, arg := range call {
			if strings.Contains(arg, "moai run SPEC-ISSUE-10") {
				foundSpecCommand = true
			}
		}
	}
	if !foundSpecCommand {
		t.Error("send-keys should contain SPEC command 'moai run SPEC-ISSUE-10'")
	}
}
