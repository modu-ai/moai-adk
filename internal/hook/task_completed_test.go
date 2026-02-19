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

func TestParseUncheckedCriteria(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "all checked",
			content: "# SPEC\n\n## Acceptance Criteria\n\n- [x] Done 1\n- [x] Done 2\n",
			want:    nil,
		},
		{
			name:    "some unchecked",
			content: "# SPEC\n\n## Acceptance Criteria\n\n- [x] Done\n- [ ] Pending 1\n- [ ] Pending 2\n",
			want:    []string{"- [ ] Pending 1", "- [ ] Pending 2"},
		},
		{
			name:    "no acceptance criteria section",
			content: "# SPEC\n\n## Overview\n\nSome text.\n",
			want:    nil,
		},
		{
			name:    "empty file",
			content: "",
			want:    nil,
		},
		{
			name:    "acceptance criteria followed by another section",
			content: "## Acceptance Criteria\n\n- [ ] Item 1\n- [x] Item 2\n\n## Notes\n\n- [ ] This should not be collected\n",
			want:    []string{"- [ ] Item 1"},
		},
		{
			name:    "case insensitive header",
			content: "## acceptance criteria\n\n- [ ] Lower case header item\n",
			want:    []string{"- [ ] Lower case header item"},
		},
		{
			name:    "nonexistent file",
			content: "", // will use a bad path
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var specPath string
			if tt.name == "nonexistent file" {
				specPath = filepath.Join(t.TempDir(), "does-not-exist.md")
			} else {
				specPath = filepath.Join(t.TempDir(), "spec.md")
				if err := os.WriteFile(specPath, []byte(tt.content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			got := parseUncheckedCriteria(specPath)

			if len(got) != len(tt.want) {
				t.Fatalf("parseUncheckedCriteria() returned %d items, want %d\ngot:  %v\nwant: %v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("item[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
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
		{
			name: "team mode with SPEC - all criteria checked - allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-7",
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
				content := "# SPEC\n\n## Acceptance Criteria\n\n- [x] Requirement 1 is implemented\n- [x] Requirement 2 is done\n- [x] Requirement 3 is complete\n"
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with SPEC - unchecked criteria exist - reject completion",
			input: &HookInput{
				SessionID:    "sess-tc-8",
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
				content := "# SPEC\n\n## Acceptance Criteria\n\n- [x] Requirement 1 is implemented\n- [ ] Requirement 2 needs testing\n- [x] Requirement 3 is done\n- [ ] Requirement 4 is pending\n"
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantExitCode: 2,
		},
		{
			name: "team mode with SPEC - no acceptance criteria section - allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-9",
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
				content := "# SPEC\n\n## Overview\n\nSome overview text.\n\n## Requirements\n\n- Feature A\n- Feature B\n"
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with SPEC - spec.md has no criteria at all - allow completion",
			input: &HookInput{
				SessionID:    "sess-tc-10",
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
				content := "# SPEC\n\nSimple spec with no sections.\n"
				if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
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
