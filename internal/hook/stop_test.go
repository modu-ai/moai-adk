package hook

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/lifecycle"
)

// writeTestFile is a test helper that writes data to a file path.
func writeTestFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func TestStopHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewStopHandler()

	if got := h.EventType(); got != EventStop {
		t.Errorf("EventType() = %q, want %q", got, EventStop)
	}
}

func TestStopHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "normal stop",
			input: &HookInput{
				SessionID:     "sess-stop-1",
				CWD:           t.TempDir(),
				HookEventName: "Stop",
				ProjectDir:    t.TempDir(),
			},
		},
		{
			name: "stop without project dir",
			input: &HookInput{
				SessionID:     "sess-stop-2",
				CWD:           "/tmp",
				HookEventName: "Stop",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewStopHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			} else if got.HookSpecificOutput != nil {
				// Stop hooks return empty JSON {} per Claude Code protocol
				// They should NOT have hookSpecificOutput set
				t.Error("Stop hook should not set hookSpecificOutput")
			}
		})
	}
}

func TestStopHandler_Handle_StopHookActive(t *testing.T) {
	t.Parallel()

	h := NewStopHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-stop-active",
		CWD:            "/tmp",
		HookEventName:  "Stop",
		StopHookActive: true,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		// When StopHookActive is true, handler should return empty to break loop
		t.Errorf("Decision should be empty when StopHookActive=true, got %q", got.Decision)
	}
}

func TestStopHandler_Handle_StopHookNotActive(t *testing.T) {
	t.Parallel()

	h := NewStopHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-stop-normal",
		CWD:            "/tmp",
		HookEventName:  "Stop",
		StopHookActive: false,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.HookSpecificOutput != nil {
		// Default behavior: allow stop
		t.Error("Stop hook should not set hookSpecificOutput")
	}
}

// TestStopHandler_PersistentMode verifies that when persistent-mode.json is active,
// the handler blocks the stop.
func TestStopHandler_PersistentMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := lifecycle.ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-persist-block",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: false,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "block" {
		t.Errorf("Decision = %q, want %q", got.Decision, "block")
	} else if got.Reason == "" {
		t.Error("Reason should not be empty when blocking")
	}
}

// TestStopHandler_PersistentMode_CompletionMarker verifies that a completion marker
// in ToolOutput deactivates persistent mode and allows the stop.
func TestStopHandler_PersistentMode_CompletionMarker(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := lifecycle.ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-persist-done",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: false,
		ToolOutput:     json.RawMessage(`"task complete <moai>DONE</moai>"`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		// Completion marker should allow the stop (no block)
		t.Errorf("Decision = %q, want empty (stop allowed after completion marker)", got.Decision)
	}

	// Verify persistent mode was deactivated
	mode, err := lifecycle.CheckPersistentMode(projectDir)
	if err != nil {
		t.Fatalf("CheckPersistentMode() error = %v", err)
	}
	if mode == nil {
		t.Fatal("persistent-mode.json should still exist")
	} else if mode.Active {
		t.Error("persistent mode should be deactivated after completion marker")
	}
}

// TestStopHandler_PersistentMode_Expired verifies that an expired persistent mode
// allows the stop.
func TestStopHandler_PersistentMode_Expired(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// First activate so the directory exists
	if err := lifecycle.ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}
	// Overwrite with an expired timestamp (started 2 hours ago, max 60 minutes)
	expiredMode := lifecycle.PersistentMode{
		Active:             true,
		Workflow:           "run",
		SpecID:             "SPEC-001",
		StartedAt:          time.Now().Add(-2 * time.Hour),
		MaxDurationMinutes: 60,
	}
	expiredData, err := json.MarshalIndent(expiredMode, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent error = %v", err)
	}
	stateFile := projectDir + "/.moai/state/persistent-mode.json"
	if err := writeTestFile(stateFile, expiredData); err != nil {
		t.Fatalf("failed to write expired mode: %v", err)
	}

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-persist-expired",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: false,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		t.Errorf("Decision = %q, want empty (stop allowed when expired)", got.Decision)
	}
}

// TestStopHandler_PersistentMode_Inactive verifies that inactive persistent mode
// does not block the stop.
func TestStopHandler_PersistentMode_Inactive(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// Activate then immediately deactivate
	if err := lifecycle.ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}
	if err := lifecycle.DeactivatePersistentMode(projectDir); err != nil {
		t.Fatalf("DeactivatePersistentMode() error = %v", err)
	}

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-persist-inactive",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: false,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		t.Errorf("Decision = %q, want empty (inactive mode should not block)", got.Decision)
	}
}

// TestStopHandler_PersistentMode_NoFile verifies that missing persistent-mode.json
// does not block the stop.
func TestStopHandler_PersistentMode_NoFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-no-persist-file",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: false,
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		t.Errorf("Decision = %q, want empty (no file should allow stop)", got.Decision)
	}
}

// TestStopHandler_StopHookActive_OverridesPersistentMode verifies that
// StopHookActive=true always allows stop, even when persistent mode is active.
func TestStopHandler_StopHookActive_OverridesPersistentMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := lifecycle.ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-stop-hook-active-wins",
		CWD:            projectDir,
		HookEventName:  "Stop",
		StopHookActive: true, // This must override persistent mode
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	} else if got.Decision != "" {
		// stop_hook_active always wins — no block
		t.Errorf("Decision = %q, want empty (stop_hook_active overrides persistent mode)", got.Decision)
	}
}
