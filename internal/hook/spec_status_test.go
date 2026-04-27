package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// TestSpecStatusHandler_PostToolUse tests the hook handler triggers on git commit
func TestSpecStatusHandler_PostToolUse(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test SPECs
	for i := 1; i <= 2; i++ {
		specDir := filepath.Join(tmpDir, ".moai", "specs", fmt.Sprintf("SPEC-HOOK-%03d", i))
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Fatalf("failed to create spec dir: %v", err)
		}

		specPath := filepath.Join(specDir, "spec.md")
		content := fmt.Sprintf("---\nstatus: draft\n---\n# Test SPEC %d\n", i)
		if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec file: %v", err)
		}
	}

	handler := NewSpecStatusHandler()

	// Simulate PostToolUse event for git commit
	commitMsg := "feat(SPEC-HOOK-001): Implement feature"
	input := &HookInput{
		SessionID:     "test-session",
		ProjectDir:   tmpDir,
		HookEventName: "PostToolUse",
		Data:         buildHookData("Bash", "git commit -m '"+commitMsg+"'"),
	}

	ctx := context.Background()
	output, err := handler.Handle(ctx, input)

	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	// Hook should always succeed (non-blocking)
	if output == nil {
		t.Fatal("expected non-nil output")
	}

	// Verify SPEC-HOOK-001 status was updated
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-HOOK-001")
	status, err := spec.ParseStatus(specDir)
	if err != nil {
		t.Fatalf("failed to parse status: %v", err)
	}

	if status != "implemented" {
		t.Errorf("expected status 'implemented', got %q", status)
	}

	// SPEC-HOOK-002 should remain unchanged (not in commit message)
	specDir2 := filepath.Join(tmpDir, ".moai", "specs", "SPEC-HOOK-002")
	status2, _ := spec.ParseStatus(specDir2)
	if status2 == "implemented" {
		t.Error("SPEC-HOOK-002 should not be updated")
	}
}

// TestSpecStatusHandler_MultipleSPECs tests extracting multiple SPEC-IDs from commit
func TestSpecStatusHandler_MultipleSPECs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test SPECs
	specIDs := []string{"SPEC-MULTI-001", "SPEC-MULTI-002", "SPEC-MULTI-003"}
	for _, specID := range specIDs {
		specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Fatalf("failed to create spec dir: %v", err)
		}

		specPath := filepath.Join(specDir, "spec.md")
		content := "---\nstatus: draft\n---\n"
		if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec file: %v", err)
		}
	}

	handler := NewSpecStatusHandler()

	// Commit message with multiple SPEC-IDs
	commitMsg := "feat(SPEC-MULTI-001, SPEC-MULTI-002): Implement features"
	input := &HookInput{
		SessionID:     "test-session",
		ProjectDir:   tmpDir,
		HookEventName: "PostToolUse",
		Data:         buildHookData("Bash", "git commit -m '"+commitMsg+"'"),
	}

	ctx := context.Background()
	_, err := handler.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	// Verify both SPECs were updated
	for i := 1; i <= 2; i++ {
		specDir := filepath.Join(tmpDir, ".moai", "specs", fmt.Sprintf("SPEC-MULTI-%03d", i))
		status, err := spec.ParseStatus(specDir)
		if err != nil {
			t.Fatalf("failed to parse status: %v", err)
		}

		if status != "implemented" {
			t.Errorf("SPEC-MULTI-%03d: expected status 'implemented', got %q", i, status)
		}
	}

	// SPEC-MULTI-003 should remain draft
	specDir3 := filepath.Join(tmpDir, ".moai", "specs", "SPEC-MULTI-003")
	status3, _ := spec.ParseStatus(specDir3)
	if status3 != "draft" {
		t.Errorf("SPEC-MULTI-003: expected status 'draft', got %q", status3)
	}
}

// TestSpecStatusHandler_NonGitCommand tests that non-git commands are ignored
func TestSpecStatusHandler_NonGitCommand(t *testing.T) {
	tmpDir := t.TempDir()

	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-NGIT-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := "---\nstatus: draft\n---\n"
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	handler := NewSpecStatusHandler()

	input := &HookInput{
		SessionID:     "test-session",
		ProjectDir:   tmpDir,
		HookEventName: "PostToolUse",
		Data:         buildHookData("Bash", "go test ./..."),
	}

	ctx := context.Background()
	_, err := handler.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	// Status should remain unchanged
	status, _ := spec.ParseStatus(specDir)
	if status != "draft" {
		t.Errorf("expected status 'draft', got %q", status)
	}
}

// TestSpecStatusHandler_NoSpecsDirectory tests graceful handling when .moai/specs doesn't exist
func TestSpecStatusHandler_NoSpecsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create .moai/specs

	handler := NewSpecStatusHandler()

	input := &HookInput{
		SessionID:     "test-session",
		ProjectDir:   tmpDir,
		HookEventName: "PostToolUse",
		Data:         buildHookData("Bash", "git commit -m 'feat(SPEC-NOEXIST-001): test'"),
	}

	ctx := context.Background()
	output, err := handler.Handle(ctx, input)

	// Should not error (non-blocking)
	if err != nil {
		t.Fatalf("Handle() should not error: %v", err)
	}

	if output == nil {
		t.Fatal("expected non-nil output")
	}
}

// TestExtractSPECIDs tests the SPEC-ID extraction pattern
func TestExtractSPECIDs(t *testing.T) {
	tests := []struct {
		name     string
		commitMsg string
		expected []string
	}{
		{
			name:     "single SPEC",
			commitMsg: "feat(SPEC-TEST-001): Implement feature",
			expected: []string{"SPEC-TEST-001"},
		},
		{
			name:     "multiple SPECs",
			commitMsg: "feat(SPEC-TEST-001, SPEC-AUTH-002): Implement features",
			expected: []string{"SPEC-TEST-001", "SPEC-AUTH-002"},
		},
		{
			name:     "SPEC with dash in name",
			commitMsg: "feat(SPEC-V3R2-CON-001): Add constitution",
			expected: []string{"SPEC-V3R2-CON-001"},
		},
		{
			name:     "no SPEC patterns",
			commitMsg: "feat: add some feature",
			expected: nil,
		},
		{
			name:     "conventional commit without SPEC",
			commitMsg: "fix(auth): handle edge case",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := extractSPECIDs(tt.commitMsg)

			if len(found) != len(tt.expected) {
				t.Errorf("expected %d SPEC-IDs, got %d", len(tt.expected), len(found))
			}

			for i, expected := range tt.expected {
				if i >= len(found) {
					t.Errorf("missing expected SPEC-ID: %s", expected)
					continue
				}
				if found[i] != expected {
					t.Errorf("expected %q at index %d, got %q", expected, i, found[i])
				}
			}
		})
	}
}

// buildHookData creates JSON data for HookInput.Data
func buildHookData(toolName, command string) json.RawMessage {
	data := map[string]any{
		"tool_name": toolName,
		"command":   command,
	}
	raw, _ := json.Marshal(data)
	return raw
}
