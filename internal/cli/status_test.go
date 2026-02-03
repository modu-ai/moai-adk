package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for status command behavior ---

func TestStatusCmd_Exists(t *testing.T) {
	if statusCmd == nil {
		t.Fatal("statusCmd should not be nil")
	}
}

func TestStatusCmd_Use(t *testing.T) {
	if statusCmd.Use != "status" {
		t.Errorf("statusCmd.Use = %q, want %q", statusCmd.Use, "status")
	}
}

func TestStatusCmd_Short(t *testing.T) {
	if statusCmd.Short == "" {
		t.Error("statusCmd.Short should not be empty")
	}
}

func TestStatusCmd_Long(t *testing.T) {
	if statusCmd.Long == "" {
		t.Error("statusCmd.Long should not be empty")
	}
}

func TestStatusCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("status should be registered as a subcommand of root")
	}
}

func TestStatusCmd_Execution(t *testing.T) {
	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("status command RunE error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("status command should produce output")
	}
}

func TestStatusCmd_HelpOutput(t *testing.T) {
	usage := statusCmd.UsageString()
	if !strings.Contains(usage, "status") {
		t.Error("status usage should contain 'status'")
	}

	if !strings.Contains(statusCmd.Long, "SPEC") || !strings.Contains(statusCmd.Long, "quality") {
		t.Error("status Long description should mention SPEC and quality")
	}
}

// --- TDD: Tests for status with different project states ---

func TestRunStatus_NoMoAIDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err = runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Not initialized") {
		t.Errorf("output should indicate not initialized, got %q", output)
	}
}

func TestRunStatus_WithMoAIDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project structure
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "specs", "SPEC-001"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "specs", "SPEC-002"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Create config files
	if err := os.WriteFile(filepath.Join(tmpDir, ".moai", "config", "sections", "user.yaml"), []byte("user:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".moai", "config", "sections", "quality.yaml"), []byte("constitution:\n  development_mode: ddd\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err = runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Initialized") {
		t.Errorf("output should contain 'Initialized', got %q", output)
	}
	if !strings.Contains(output, "SPECs") {
		t.Errorf("output should contain 'SPECs', got %q", output)
	}
	if !strings.Contains(output, "2 found") {
		t.Errorf("output should show 2 SPECs found, got %q", output)
	}
	if !strings.Contains(output, "2 section files") {
		t.Errorf("output should show 2 section files, got %q", output)
	}
}

func TestCountDirs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "dir1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "dir2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file1"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	count := countDirs(tmpDir)
	if count != 2 {
		t.Errorf("countDirs = %d, want 2", count)
	}
}

func TestCountDirs_NonExistent(t *testing.T) {
	count := countDirs("/nonexistent/path")
	if count != 0 {
		t.Errorf("countDirs for nonexistent path = %d, want 0", count)
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.yaml"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "c.json"), []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}

	count := countFiles(tmpDir, ".yaml")
	if count != 2 {
		t.Errorf("countFiles(.yaml) = %d, want 2", count)
	}

	count = countFiles(tmpDir, ".json")
	if count != 1 {
		t.Errorf("countFiles(.json) = %d, want 1", count)
	}
}

func TestCountFiles_NonExistent(t *testing.T) {
	count := countFiles("/nonexistent/path", ".yaml")
	if count != 0 {
		t.Errorf("countFiles for nonexistent path = %d, want 0", count)
	}
}

func TestRunStatus_OutputContainsProjectName(t *testing.T) {
	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err := runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Project Status") {
		t.Errorf("output should contain 'Project Status', got %q", output)
	}
	if !strings.Contains(output, "moai-adk") {
		t.Errorf("output should contain 'moai-adk', got %q", output)
	}
}
