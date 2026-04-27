package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSpecStatusCommand tests the basic command structure
func TestSpecStatusCommand(t *testing.T) {
	cmd := newSpecCmd()

	if cmd.Use != "spec" {
		t.Errorf("expected command Use 'spec', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected command Short to be set")
	}
}

// TestSpecStatusUpdate tests updating a SPEC status via CLI
func TestSpecStatusUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-CLI-TEST-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-CLI-TEST-001
status: draft
---
# Test SPEC
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Override project root for test
	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return tmpDir, nil
	}

	// Execute command
	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"SPEC-CLI-TEST-001", "completed"})
	output := &strings.Builder{}
	cmd.SetOut(output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "status updated") {
		t.Errorf("expected output to contain 'status updated', got: %s", outputStr)
	}

	// Verify the file was updated
	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec: %v", err)
	}

	if !strings.Contains(string(updatedContent), "status: completed") {
		t.Error("status was not updated in file")
	}
}

// TestSpecStatusDryRun tests the --dry-run flag
func TestSpecStatusDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-DRYRUN-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
status: draft
---
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return tmpDir, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"SPEC-DRYRUN-001", "completed", "--dry-run"})
	output := &strings.Builder{}
	cmd.SetOut(output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	// File should NOT be modified
	originalContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}

	if strings.Contains(string(originalContent), "status: completed") {
		t.Error("file was modified during dry-run")
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Would update") {
		t.Errorf("expected dry-run output, got: %s", outputStr)
	}
}

// TestSpecStatusList tests the --list flag
func TestSpecStatusList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple SPECs
	for i := 1; i <= 3; i++ {
		specDir := filepath.Join(tmpDir, ".moai", "specs", fmt.Sprintf("SPEC-LIST-%03d", i))
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Fatalf("failed to create spec dir: %v", err)
		}

		status := "draft"
		if i == 2 {
			status = "completed"
		}

		specPath := filepath.Join(specDir, "spec.md")
		content := fmt.Sprintf("---\nstatus: %s\n---\n", status)
		if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec file: %v", err)
		}
	}

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return tmpDir, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"--list"})
	output := &strings.Builder{}
	cmd.SetOut(output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "SPEC-LIST-001") {
		t.Error("output should contain SPEC-LIST-001")
	}
	if !strings.Contains(outputStr, "SPEC-LIST-002") {
		t.Error("output should contain SPEC-LIST-002")
	}
	if !strings.Contains(outputStr, "draft") {
		t.Error("output should contain draft status")
	}
	if !strings.Contains(outputStr, "completed") {
		t.Error("output should contain completed status")
	}
}

// TestSpecStatusNotFound tests error handling for non-existent SPEC
func TestSpecStatusNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return tmpDir, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"SPEC-NONEXISTENT", "completed"})
	output := &strings.Builder{}
	cmd.SetOut(output)
	cmd.SetErr(output)

	if err := cmd.Execute(); err == nil {
		t.Error("command should return error for non-existent SPEC")
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "not found") && !strings.Contains(outputStr, "Error") {
		t.Errorf("expected error message, got: %s", outputStr)
	}
}
