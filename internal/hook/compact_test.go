package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
		name            string
		input           *HookInput
		setupDir        bool
		setProjectDir   bool
		wantDecision    string
	}{
		{
			name: "normal context preservation",
			input: &HookInput{
				SessionID:     "sess-compact-1",
				HookEventName: "PreCompact",
			},
			setupDir:     true,
			setProjectDir: true,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact without memory dir auto-creates it",
			input: &HookInput{
				SessionID:     "sess-compact-2",
				HookEventName: "PreCompact",
			},
			setupDir:     false,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact with no project dir",
			input: &HookInput{
				SessionID:     "sess-compact-3",
				HookEventName: "PreCompact",
			},
			setupDir:     false,
			wantDecision: DecisionAllow,
		},
		{
			name: "compact preserves session id in data",
			input: &HookInput{
				SessionID:     "sess-compact-preserve",
				HookEventName: "PreCompact",
			},
			setupDir:      false,
			setProjectDir: true,
			wantDecision:  DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use an isolated temp dir for all tests to avoid polluting
			// /tmp/.moai which causes findProjectRoot failures in other packages
			// on Linux (where t.TempDir() is under /tmp/).
			tmpDir := t.TempDir()
			if tt.setupDir {
				stateDir := filepath.Join(tmpDir, ".moai", "state")
				if err := os.MkdirAll(stateDir, 0o755); err != nil {
					t.Fatalf("setup state dir: %v", err)
				}
			}
			tt.input.CWD = tmpDir
			if tt.setProjectDir {
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
			} else if got.HookSpecificOutput != nil {
				// PreCompact does NOT use hookSpecificOutput per Claude Code protocol
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

	tmpDir := t.TempDir()
	input := &HookInput{
		SessionID:     "sess-data-check",
		CWD:           tmpDir,
		ProjectDir:    tmpDir,
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

func TestCompactHandler_Handle_WritesMemoFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	input := &HookInput{
		SessionID:     "sess-memo-write",
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "PreCompact",
	}

	h := NewCompactHandler()
	_, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	memoPath := filepath.Join(projectDir, ".moai", "state", "session-memo.md")
	data, err := os.ReadFile(memoPath)
	if err != nil {
		t.Fatalf("session memo not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "sess-memo-write") {
		t.Errorf("session memo should contain session_id, got:\n%s", content)
	}
}

func TestCompactHandler_Handle_ReadsPersistentMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	modeFile := filepath.Join(stateDir, "persistent-mode.json")
	if err := os.WriteFile(modeFile, []byte(`{"mode":"cg","active":true}`), 0o644); err != nil {
		t.Fatalf("write mode file: %v", err)
	}

	input := &HookInput{
		SessionID:     "sess-mode-test",
		CWD:           projectDir,
		HookEventName: "PreCompact",
	}

	h := NewCompactHandler()
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	memoPath := filepath.Join(projectDir, ".moai", "state", "session-memo.md")
	data, _ := os.ReadFile(memoPath)
	if !strings.Contains(string(data), "mode") {
		t.Errorf("memo should include persistent mode info, got:\n%s", string(data))
	}
}
