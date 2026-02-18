package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskCompletedHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewTaskCompletedHandler()

	if got := h.EventType(); got != EventTaskCompleted {
		t.Errorf("EventType() = %q, want %q", got, EventTaskCompleted)
	}
}

func TestTaskCompletedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		setupDir     func(t *testing.T) string // returns projectDir; nil means no setup
		wantExitCode int
	}{
		{
			name: "no team mode - always allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-1",
				TaskSubject:  "Implement SPEC-FEAT-001 backend",
				TeammateName: "worker-1",
				// TeamName empty: not in team mode
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with task subject without SPEC ID - allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-2",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				TaskSubject:  "Fix linting errors in service layer",
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with SPEC ID and spec.md exists - allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-3",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				TaskSubject:  "Implement SPEC-TEAM-001 quality hooks",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				specDir := filepath.Join(dir, ".moai", "specs", "SPEC-TEAM-001")
				if err := os.MkdirAll(specDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# SPEC"), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with SPEC ID but spec.md missing - reject completion",
			input: &HookInput{
				SessionID:    "sess-tc-4",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				TaskSubject:  "Implement SPEC-TEAM-001 quality hooks",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				// Project dir exists but .moai/specs/SPEC-TEAM-001/spec.md does not.
				return t.TempDir()
			},
			wantExitCode: 2,
		},
		{
			name: "team mode with SPEC ID and no project dir - allow completion (graceful)",
			input: &HookInput{
				SessionID:    "sess-tc-5",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				TaskSubject:  "Implement SPEC-TEAM-001 quality hooks",
				// No CWD or ProjectDir: cannot verify, allow completion.
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with multiple SPEC IDs in subject - uses first match",
			input: &HookInput{
				SessionID:    "sess-tc-6",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				TaskSubject:  "Implement SPEC-TEAM-001 and SPEC-FEAT-002",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// SPEC-TEAM-001 exists; SPEC-FEAT-002 does not (but won't be checked).
				specDir := filepath.Join(dir, ".moai", "specs", "SPEC-TEAM-001")
				if err := os.MkdirAll(specDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# SPEC"), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input
			if tt.setupDir != nil {
				projectDir := tt.setupDir(t)
				clone := *input
				clone.CWD = projectDir
				input = &clone
			}

			h := NewTaskCompletedHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, tt.wantExitCode)
			}
		})
	}
}
