package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompactHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewCompactHandler()

	if got := h.EventType(); got != EventPreCompact {
		t.Errorf("EventType() = %q, want %q", got, EventPreCompact)
	}
}

func TestCompactHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		setupDir     bool
		wantDecision string
	}{
		{
			name: "normal context preservation",
			input: &HookInput{
				SessionID:     "sess-compact-1",
				CWD:           "", // will be set in test
				HookEventName: "PreCompact",
			},
			setupDir:     true,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact without memory dir auto-creates it",
			input: &HookInput{
				SessionID:     "sess-compact-2",
				CWD:           "", // will be set in test
				HookEventName: "PreCompact",
			},
			setupDir:     false,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact with no project dir",
			input: &HookInput{
				SessionID:     "sess-compact-3",
				CWD:           "/tmp",
				HookEventName: "PreCompact",
			},
			setupDir:     false,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact preserves session id in data",
			input: &HookInput{
				SessionID:     "sess-compact-preserve",
				CWD:           "/tmp",
				ProjectDir:    "/tmp",
				HookEventName: "PreCompact",
			},
			setupDir:     false,
			wantDecision: DecisionAllow,
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

			h := NewCompactHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// PreCompact does NOT use hookSpecificOutput per Claude Code protocol
			if got.HookSpecificOutput != nil {
				t.Errorf("HookSpecificOutput should be nil for PreCompact, got %+v", got.HookSpecificOutput)
			}
			if got.Data != nil && !json.Valid(got.Data) {
				t.Errorf("Data is not valid JSON: %s", got.Data)
			}
		})
	}
}

func TestCompactHandler_Handle_DataContainsSessionID(t *testing.T) {
	t.Parallel()

	h := NewCompactHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-data-check",
		CWD:           "/tmp",
		ProjectDir:    "/tmp",
		HookEventName: "PreCompact",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got.Data == nil {
		t.Fatal("Data should not be nil")
	}

	var data map[string]any
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("unmarshal Data: %v", err)
	}

	if data["session_id"] != "sess-data-check" {
		t.Errorf("session_id = %v, want sess-data-check", data["session_id"])
	}
	if data["status"] != "preserved" {
		t.Errorf("status = %v, want preserved", data["status"])
	}
	if data["snapshot_created"] != true {
		t.Errorf("snapshot_created = %v, want true", data["snapshot_created"])
	}
}
