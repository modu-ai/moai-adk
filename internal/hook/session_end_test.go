package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSessionEndHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewSessionEndHandler()

	if got := h.EventType(); got != EventSessionEnd {
		t.Errorf("EventType() = %q, want %q", got, EventSessionEnd)
	}
}

func TestSessionEndHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *HookInput
		setupDir bool
	}{
		{
			name: "normal session end",
			input: &HookInput{
				SessionID:     "sess-end-1",
				CWD:           "", // will be set in test
				HookEventName: "SessionEnd",
			},
			setupDir: true,
		},
		{
			name: "session end without project dir",
			input: &HookInput{
				SessionID:     "sess-end-2",
				CWD:           "/tmp",
				HookEventName: "SessionEnd",
			},
			setupDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setupDir {
				tmpDir := t.TempDir()
				memDir := filepath.Join(tmpDir, ".moai", "memory")
				if err := os.MkdirAll(memDir, 0o755); err != nil {
					t.Fatalf("setup memory dir: %v", err)
				}
				tt.input.CWD = tmpDir
				tt.input.ProjectDir = tmpDir
			}

			h := NewSessionEndHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// SessionEnd hooks return empty JSON {} per Claude Code protocol
			// They should NOT have hookSpecificOutput set
			if got.HookSpecificOutput != nil {
				t.Error("SessionEnd hook should not set hookSpecificOutput")
			}
		})
	}
}
