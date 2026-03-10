package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/loop"
)

// --- Command registration validation tests ---

func TestLoopCmd_Exists(t *testing.T) {
	if loopCmd == nil {
		t.Fatal("loopCmd should not be nil")
	}
}

func TestLoopCmd_Use(t *testing.T) {
	if loopCmd.Use != "loop" {
		t.Errorf("loopCmd.Use = %q, want %q", loopCmd.Use, "loop")
	}
}

func TestLoopCmd_GroupID(t *testing.T) {
	if loopCmd.GroupID != "tools" {
		t.Errorf("loopCmd.GroupID = %q, want %q", loopCmd.GroupID, "tools")
	}
}

func TestLoopCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "loop" {
			found = true
			break
		}
	}
	if !found {
		t.Error("loop should be registered as a subcommand of root")
	}
}

func TestLoopCmd_HasExpectedSubcommands(t *testing.T) {
	expected := []string{"start", "status", "pause", "resume", "cancel"}
	for _, name := range expected {
		found := false
		for _, cmd := range loopCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("loop should have %q subcommand", name)
		}
	}
}

// --- noopFeedbackGenerator tests ---

func TestNoopFeedbackGenerator_Collect(t *testing.T) {
	gen := &noopFeedbackGenerator{}
	fb, err := gen.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil feedback")
	}
}

// --- Command handler tests (no dependencies) ---

func TestRunLoopStart_NilDeps(t *testing.T) {
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	err := runLoopStart(loopStartCmd, []string{"SPEC-001"})
	if err == nil {
		t.Error("expected error when deps is nil")
	}
}

func TestRunLoopStatus_NilDeps(t *testing.T) {
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	err := runLoopStatus(loopStatusCmd, []string{})
	if err == nil {
		t.Error("expected error when deps is nil")
	}
}

func TestRunLoopPause_NilDeps(t *testing.T) {
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	err := runLoopPause(loopPauseCmd, []string{})
	if err == nil {
		t.Error("expected error when deps is nil")
	}
}

func TestRunLoopResume_NilDeps(t *testing.T) {
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	err := runLoopResume(loopResumeCmd, []string{"SPEC-001"})
	if err == nil {
		t.Error("expected error when deps is nil")
	}
}

func TestRunLoopCancel_NilDeps(t *testing.T) {
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	err := runLoopCancel(loopCancelCmd, []string{})
	if err == nil {
		t.Error("expected error when deps is nil")
	}
}

// --- Integration tests with real LoopController ---

func newTestLoopDeps(t *testing.T) *Dependencies {
	t.Helper()
	storage := loop.NewFileStorage(t.TempDir())
	engine := &stubDecisionEngine{}
	gen := &noopFeedbackGenerator{}
	ctrl := loop.NewLoopController(storage, engine, gen, 3)
	return &Dependencies{
		LoopController: ctrl,
	}
}

// stubDecisionEngine is a test engine that always returns ActionAbort.
type stubDecisionEngine struct{}

func (e *stubDecisionEngine) Decide(_ context.Context, _ *loop.LoopState, _ *loop.Feedback) (*loop.Decision, error) {
	return &loop.Decision{Action: loop.ActionAbort}, nil
}

func TestRunLoopStatus_NoActiveLoop(t *testing.T) {
	origDeps := deps
	deps = newTestLoopDeps(t)
	defer func() { deps = origDeps }()

	buf := new(bytes.Buffer)
	loopStatusCmd.SetOut(buf)
	defer loopStatusCmd.SetOut(nil)

	if err := runLoopStatus(loopStatusCmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No active loop") {
		t.Errorf("expected 'No active loop', got %q", buf.String())
	}
}

func TestRunLoopPause_NotRunning(t *testing.T) {
	origDeps := deps
	deps = newTestLoopDeps(t)
	defer func() { deps = origDeps }()

	err := runLoopPause(loopPauseCmd, []string{})
	if err == nil {
		t.Error("expected error when pausing a non-running loop")
	}
	if !strings.Contains(err.Error(), "loop pause") {
		t.Errorf("error should mention 'loop pause', got %q", err.Error())
	}
}

func TestRunLoopCancel_NotRunning(t *testing.T) {
	origDeps := deps
	deps = newTestLoopDeps(t)
	defer func() { deps = origDeps }()

	err := runLoopCancel(loopCancelCmd, []string{})
	if err == nil {
		t.Error("expected error when cancelling a non-running loop")
	}
	if !strings.Contains(err.Error(), "loop cancel") {
		t.Errorf("error should mention 'loop cancel', got %q", err.Error())
	}
}

func TestRunLoopResume_NotPaused(t *testing.T) {
	origDeps := deps
	deps = newTestLoopDeps(t)
	defer func() { deps = origDeps }()

	err := runLoopResume(loopResumeCmd, []string{"SPEC-001"})
	if err == nil {
		t.Error("expected error when resuming a non-paused loop")
	}
	if !strings.Contains(err.Error(), "loop resume") {
		t.Errorf("error should mention 'loop resume', got %q", err.Error())
	}
}

// TestLoopSubcmds_ArgsValidation is a table-driven test for required argument validation per subcommand.
func TestLoopSubcmds_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		runFunc func() error
		wantErr bool
	}{
		{
			name:    "start_nil_deps",
			runFunc: func() error { return runLoopStart(loopStartCmd, []string{"SPEC-X"}) },
			wantErr: true,
		},
		{
			name:    "status_nil_deps",
			runFunc: func() error { return runLoopStatus(loopStatusCmd, []string{}) },
			wantErr: true,
		},
		{
			name:    "pause_nil_deps",
			runFunc: func() error { return runLoopPause(loopPauseCmd, []string{}) },
			wantErr: true,
		},
		{
			name:    "resume_nil_deps",
			runFunc: func() error { return runLoopResume(loopResumeCmd, []string{"SPEC-X"}) },
			wantErr: true,
		},
		{
			name:    "cancel_nil_deps",
			runFunc: func() error { return runLoopCancel(loopCancelCmd, []string{}) },
			wantErr: true,
		},
	}

	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.runFunc()
			if tt.wantErr && err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
			}
		})
	}
}
