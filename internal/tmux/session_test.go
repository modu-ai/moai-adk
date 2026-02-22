package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"testing"
)

func TestSessionManager_Create_SinglePane(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "test-session",
		Panes: []PaneConfig{
			{SpecID: "SPEC-ISSUE-1", Command: "moai worktree go SPEC-ISSUE-1"},
		},
		MaxVisible: 3,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SessionName != "test-session" {
		t.Errorf("SessionName = %q, want %q", result.SessionName, "test-session")
	}
	if result.PaneCount != 1 {
		t.Errorf("PaneCount = %d, want 1", result.PaneCount)
	}

	// Verify: new-session with command for first pane, then select-pane.
	assertCallContains(t, calls, "new-session")
}

func TestSessionManager_Create_ThreePanes(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "three-panes",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
			{SpecID: "SPEC-3", Command: "cmd3"},
		},
		MaxVisible: 3,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PaneCount != 3 {
		t.Errorf("PaneCount = %d, want 3", result.PaneCount)
	}

	// Verify vertical splits for panes 2 and 3.
	verticalSplits := countSplits(calls, "-v")
	if verticalSplits != 2 {
		t.Errorf("vertical splits = %d, want 2", verticalSplits)
	}
}

func TestSessionManager_Create_FourPanes_HorizontalOverflow(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "four-panes",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
			{SpecID: "SPEC-3", Command: "cmd3"},
			{SpecID: "SPEC-4", Command: "cmd4"},
		},
		MaxVisible: 3,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PaneCount != 4 {
		t.Errorf("PaneCount = %d, want 4", result.PaneCount)
	}

	// First 2 additional panes use vertical, 4th uses horizontal.
	verticalSplits := countSplits(calls, "-v")
	horizontalSplits := countSplits(calls, "-h")

	if verticalSplits != 2 {
		t.Errorf("vertical splits = %d, want 2", verticalSplits)
	}
	if horizontalSplits != 1 {
		t.Errorf("horizontal splits = %d, want 1", horizontalSplits)
	}
}

func TestSessionManager_Create_NoPanes(t *testing.T) {
	t.Parallel()

	runner := func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name:  "empty",
		Panes: nil,
	}

	_, err := mgr.Create(context.Background(), cfg)
	if !errors.Is(err, ErrNoPanes) {
		t.Errorf("error = %v, want ErrNoPanes", err)
	}
}

func TestSessionManager_Create_EmptyPaneSlice(t *testing.T) {
	t.Parallel()

	runner := func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name:  "empty-slice",
		Panes: []PaneConfig{},
	}

	_, err := mgr.Create(context.Background(), cfg)
	if !errors.Is(err, ErrNoPanes) {
		t.Errorf("error = %v, want ErrNoPanes", err)
	}
}

func TestSessionManager_Create_SessionCreateFails(t *testing.T) {
	t.Parallel()

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if slices.Contains(args, "new-session") {
			return "", fmt.Errorf("session create failed")
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "fail-session",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
		},
	}

	_, err := mgr.Create(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "create session") {
		t.Errorf("error = %v, should mention 'create session'", err)
	}
}

func TestSessionManager_Create_SendKeysForEachPane(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "sendkeys-test",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
		},
		MaxVisible: 3,
	}

	_, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sendKeysCount := countCallsContaining(calls, "send-keys")
	// Each pane gets a send-keys command.
	if sendKeysCount != 2 {
		t.Errorf("send-keys count = %d, want 2", sendKeysCount)
	}
}

func TestSessionManager_Create_DefaultMaxVisible(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	// MaxVisible = 0 should use default of 3.
	cfg := &SessionConfig{
		Name: "default-max",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
			{SpecID: "SPEC-3", Command: "cmd3"},
			{SpecID: "SPEC-4", Command: "cmd4"},
		},
		MaxVisible: 0,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaneCount != 4 {
		t.Errorf("PaneCount = %d, want 4", result.PaneCount)
	}

	// With default MaxVisible=3: 2 vertical + 1 horizontal.
	horizontalSplits := countSplits(calls, "-h")
	if horizontalSplits != 1 {
		t.Errorf("horizontal splits = %d, want 1 with default MaxVisible", horizontalSplits)
	}
}

func TestSessionManager_Create_SplitWindowFails_Continues(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		if slices.Contains(args, "split-window") {
			return "", fmt.Errorf("split failed")
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "split-fail",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
		},
		MaxVisible: 3,
	}

	// Should succeed; pane creation failure is non-fatal.
	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaneCount != 2 {
		t.Errorf("PaneCount = %d, want 2 (reports configured count)", result.PaneCount)
	}
}

func TestSessionManager_Create_SendKeysFails_Continues(t *testing.T) {
	t.Parallel()

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if slices.Contains(args, "send-keys") {
			return "", fmt.Errorf("send-keys failed")
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "sendkeys-fail",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
		},
		MaxVisible: 3,
	}

	// Should succeed; send-keys failure is non-fatal.
	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SessionName != "sendkeys-fail" {
		t.Errorf("SessionName = %q, want %q", result.SessionName, "sendkeys-fail")
	}
}

func TestSessionManager_Create_SendKeysFailsForAdditionalPanes(t *testing.T) {
	t.Parallel()

	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		// send-keys fails only for panes beyond the first.
		for i, a := range args {
			if a == "send-keys" {
				// Fail if target contains ".1" or ".2" (additional panes).
				for _, arg := range args[i:] {
					if strings.Contains(arg, ".1") || strings.Contains(arg, ".2") {
						return "", fmt.Errorf("send-keys failed for additional pane")
					}
				}
			}
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	cfg := &SessionConfig{
		Name: "sendkeys-additional-fail",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
			{SpecID: "SPEC-2", Command: "cmd2"},
			{SpecID: "SPEC-3", Command: "cmd3"},
		},
		MaxVisible: 3,
	}

	// Should succeed; send-keys failure for additional panes is non-fatal.
	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PaneCount != 3 {
		t.Errorf("PaneCount = %d, want 3", result.PaneCount)
	}
}

func TestSessionManager_WithSessionLogger(t *testing.T) {
	t.Parallel()

	runner := func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	}

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	mgr := NewSessionManager(
		WithSessionRunFunc(runner),
		WithSessionLogger(logger),
	)

	cfg := &SessionConfig{
		Name: "logger-test",
		Panes: []PaneConfig{
			{SpecID: "SPEC-1", Command: "cmd1"},
		},
		MaxVisible: 3,
	}

	result, err := mgr.Create(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SessionName != "logger-test" {
		t.Errorf("SessionName = %q, want %q", result.SessionName, "logger-test")
	}
}

// --- Test Helpers ---

func assertCallContains(t *testing.T, calls [][]string, target string) {
	t.Helper()
	for _, call := range calls {
		if slices.Contains(call, target) {
			return
		}
	}
	t.Errorf("no call contains %q in %d calls", target, len(calls))
}

func countSplits(calls [][]string, direction string) int {
	count := 0
	for _, call := range calls {
		hasSplitWindow := false
		hasDirection := false
		for _, arg := range call {
			if arg == "split-window" {
				hasSplitWindow = true
			}
			if arg == direction {
				hasDirection = true
			}
		}
		if hasSplitWindow && hasDirection {
			count++
		}
	}
	return count
}

func countCallsContaining(calls [][]string, target string) int {
	count := 0
	for _, call := range calls {
		if slices.Contains(call, target) {
			count++
		}
	}
	return count
}
