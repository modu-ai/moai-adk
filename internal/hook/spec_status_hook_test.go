package hook

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

func TestIsGhPrMergeCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "gh pr merge with --squash",
			command: "gh pr merge 123 --squash",
			want:    true,
		},
		{
			name:    "gh pr merge with --merge",
			command: "gh pr merge 123 --merge",
			want:    true,
		},
		{
			name:    "gh pr merge with --delete-branch",
			command: "gh pr merge 123 --delete-branch",
			want:    true,
		},
		{
			name:    "gh pr merge with multiple flags",
			command: "gh pr merge 123 --squash --delete-branch --comment 'auto-merge'",
			want:    true,
		},
		{
			name:    "gh pr view - not a merge",
			command: "gh pr view 123",
			want:    false,
		},
		{
			name:    "git commit - not gh pr merge",
			command: "git commit -m 'message'",
			want:    false,
		},
		{
			name:    "empty command",
			command: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &specStatusHandler{}
			if got := handler.isGhPrMergeCommand(tt.command); got != tt.want {
				t.Errorf("isGhPrMergeCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpecStatusHandler_Idempotency(t *testing.T) {
	// Create a temporary project directory structure
	tmpDir := t.TempDir()
	specDir := tmpDir + "/.moai/specs/SPEC-V3R4-STATUS-LIFECYCLE-001"

	// Create spec directory
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec directory: %v", err)
	}

	// Create a spec.md file with YAML frontmatter and status=implemented
	specContent := `---
status: implemented
---

# SPEC-V3R4-STATUS-LIFECYCLE-001

## Requirements
- REQ-001: Test requirement
`
	specFile := specDir + "/spec.md"
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	// First update: should change to implemented (no-op if already there)
	handler := &specStatusHandler{}
	input := &HookInput{
		CWD: tmpDir, // Project root, not spec dir
	}
	// Simulate gh pr merge
	dataJSON, _ := json.Marshal(map[string]interface{}{
		"tool_name": "PostToolUse",
		"command":   "gh pr merge 123 --squash",
		"title":     "feat(SPEC-V3R4-STATUS-LIFECYCLE-001): implement REQ-001",
	})
	input.ToolInput = dataJSON

	ctx := context.Background()
	_, err := handler.Handle(ctx, input)
	if err != nil {
		t.Fatalf("First Handle() failed: %v", err)
	}

	// Check status is implemented
	status, err := spec.ParseStatus(specDir)
	if err != nil {
		t.Fatalf("Failed to parse status after first update: %v", err)
	}
	if status != "implemented" {
		t.Errorf("After first update: status = %q, want %q", status, "implemented")
	}

	// Second update with same target: should be no-op (idempotent)
	_, err = handler.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Second Handle() failed: %v", err)
	}

	// Status should still be implemented
	status, err = spec.ParseStatus(specDir)
	if err != nil {
		t.Fatalf("Failed to parse status after second update: %v", err)
	}
	if status != "implemented" {
		t.Errorf("After second update (idempotent): status = %q, want %q", status, "implemented")
	}
}

func TestSpecStatusHandler_Integration(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		title         string
		initialStatus string
		wantStatus    string
		wantErr       bool
	}{
		{
			name:          "git commit with SPEC ID -> implemented",
			command:       "git commit -m 'feat(SPEC-V3R4-STATUS-LIFECYCLE-001): implement'",
			title:         "feat(SPEC-V3R4-STATUS-LIFECYCLE-001): implement",
			initialStatus: "planned",
			wantStatus:    "implemented",
			wantErr:       false,
		},
		{
			name:          "gh pr merge plan -> planned",
			command:       "gh pr merge 123 --squash",
			title:         "plan(spec): SPEC-V3R4-STATUS-LIFECYCLE-001 — initial draft",
			initialStatus: "draft",
			wantStatus:    "planned",
			wantErr:       false,
		},
		{
			name:          "gh pr merge feat -> implemented",
			command:       "gh pr merge 456 --merge",
			title:         "feat(SPEC-V3R4-STATUS-LIFECYCLE-001): implement REQ-1",
			initialStatus: "in-progress",
			wantStatus:    "implemented",
			wantErr:       false,
		},
		{
			name:          "gh pr merge sync -> completed",
			command:       "gh pr merge 789 --squash",
			title:         "docs(sync): SPEC-V3R4-STATUS-LIFECYCLE-001 status=completed",
			initialStatus: "implemented",
			wantStatus:    "completed",
			wantErr:       false,
		},
		{
			name:          "skip meta - no status change",
			command:       "gh pr merge 999 --squash",
			title:         "chore(spec): auto-sync status for #999",
			initialStatus: "implemented",
			wantStatus:    "implemented", // unchanged
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specDir := tmpDir + "/.moai/specs/SPEC-V3R4-STATUS-LIFECYCLE-001"

			// Create spec directory
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			// Create spec.md with YAML frontmatter and initial status
			specContent := `---
status: ` + tt.initialStatus + `
---

# SPEC-V3R4-STATUS-LIFECYCLE-001

## Requirements
- REQ-001: Test requirement
`
			specFile := specDir + "/spec.md"
			if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
				t.Fatalf("Failed to create spec file: %v", err)
			}

			// Create hook input
			data := map[string]interface{}{
				"tool_name": "PostToolUse",
				"command":   tt.command,
				"title":     tt.title,
			}
			dataJSON, _ := json.Marshal(data)

			input := &HookInput{
				CWD:       tmpDir, // Project root, not spec dir
				ToolInput: dataJSON,
			}

			ctx := context.Background()
			handler := &specStatusHandler{}

			_, err := handler.Handle(ctx, input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify final status
			finalStatus, err := spec.ParseStatus(specDir)
			if err != nil {
				t.Fatalf("Failed to parse final status: %v", err)
			}

			if finalStatus != tt.wantStatus {
				t.Errorf("Final status = %q, want %q", finalStatus, tt.wantStatus)
			}
		})
	}
}

