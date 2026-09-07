package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initSyncGitFixture builds a throwaway git project whose main-branch history
// names specID (so the git-implied status is "implemented") and whose SPEC
// frontmatter carries the given status. Everything lives under t.TempDir().
func initSyncGitFixture(t *testing.T, specID, status string) string {
	t.Helper()

	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := "---\nid: " + specID + "\nstatus: " + status + "\n---\n\n# SPEC\n"
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	run := func(args ...string) {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("-C", tmpDir, "init", "-q", "-b", "main")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "add", "-A")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "commit", "-qm", "chore: seed")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "commit", "-qm", "feat("+specID+"): M1 implementation", "--allow-empty")

	return tmpDir
}

// SPEC-STATUS-DRYRUN-001 AC-006 (REQ-005): `--sync-git --dry-run --yes`
// computes and prints the reconciliation but writes NOTHING.
func TestSyncGitSpecStatuses_DryRunWritesNothing(t *testing.T) {
	projectRoot := initSyncGitFixture(t, "SPEC-DRYGIT-001", "draft")
	specPath := filepath.Join(projectRoot, ".moai", "specs", "SPEC-DRYGIT-001", "spec.md")

	before, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return projectRoot, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"--sync-git", "--dry-run", "--yes"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec after dry-run: %v", err)
	}

	if string(before) != string(after) {
		t.Errorf("dry-run modified spec.md:\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}

	outputStr := out.String()
	if !strings.Contains(outputStr, "SPEC-DRYGIT-001") || !strings.Contains(outputStr, "draft") {
		t.Errorf("dry-run output must carry the would-change plan (SPEC + current status), got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "dry-run") {
		t.Errorf("dry-run output must state the dry-run form, got: %s", outputStr)
	}
	if strings.Contains(outputStr, "Summary: updated") {
		t.Errorf("dry-run summary must not claim a write count, got: %s", outputStr)
	}
}

// SPEC-STATUS-DRYRUN-001 AC-007 (REQ-006): the real (non-dry-run) write path
// still updates the eligible SPEC's frontmatter status.
func TestSyncGitSpecStatuses_RealRunUpdatesFrontmatter(t *testing.T) {
	projectRoot := initSyncGitFixture(t, "SPEC-REALSYNC-001", "draft")
	specPath := filepath.Join(projectRoot, ".moai", "specs", "SPEC-REALSYNC-001", "spec.md")

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return projectRoot, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"--sync-git", "--yes"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	updated, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec after sync: %v", err)
	}

	if !strings.Contains(string(updated), "status: implemented") {
		t.Errorf("frontmatter status not updated by real sync run, got:\n%s", updated)
	}
	if !strings.Contains(out.String(), "Summary: updated 1") {
		t.Errorf("expected real-run summary with one write, got: %s", out.String())
	}
}

// SPEC-STATUS-DRYRUN-001 AC-008 (REQ-007): a SPEC whose parsed status is not a
// member of the canonical enum is skipped loudly (stderr warning naming the
// SPEC and the offending value) and never written.
func TestSyncGitSpecStatuses_InvalidStatusSkippedLoudly(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-INVALIDST-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	// No frontmatter + a body history table whose header cell parses to the
	// stray token "Notes" — not a member of the 8-value enum.
	specPath := filepath.Join(specDir, "spec.md")
	content := "# SPEC\n\n| Version | Date | Status | Notes |\n|---|---|---|---|\n| 0.1.0 | 2026-01-01 | draft | initial |\n"
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	run := func(args ...string) {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("-C", tmpDir, "init", "-q", "-b", "main")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "add", "-A")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "commit", "-qm", "chore: seed")
	run("-C", tmpDir, "-c", "user.name=t", "-c", "user.email=t@t.local", "commit", "-qm", "feat(SPEC-INVALIDST-001): M1 implementation", "--allow-empty")

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) {
		return tmpDir, nil
	}

	cmd := newSpecStatusCmd()
	cmd.SetArgs([]string{"--sync-git", "--yes"})
	out := &strings.Builder{}
	errOut := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	warnings := errOut.String()
	if !strings.Contains(warnings, "SPEC-INVALIDST-001") || !strings.Contains(warnings, "Notes") {
		t.Errorf("expected stderr warning naming the SPEC and offending value, got stderr: %q stdout: %s", warnings, out.String())
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec after sync: %v", err)
	}
	if string(after) != content {
		t.Errorf("skipped SPEC must not be written, got:\n%s", after)
	}
	if strings.Contains(out.String(), "Summary: updated 1") {
		t.Errorf("invalid-status SPEC must not count as updated, got: %s", out.String())
	}
}

// SPEC-STATUS-DRYRUN-001 AC-010 (REQ-008): help text stays accurate for the
// --sync-git + --dry-run combination.
func TestSpecStatusHelpText_DryRunAccurate(t *testing.T) {
	cmd := newSpecStatusCmd()

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("--dry-run flag not registered")
	}
	if !strings.Contains(dryRunFlag.Usage, "Preview change without writing") {
		t.Errorf("--dry-run usage must keep the documented description, got: %q", dryRunFlag.Usage)
	}
	if !strings.Contains(cmd.Long, "--sync-git --dry-run") {
		t.Errorf("help must document the --sync-git + --dry-run combination, got:\n%s", cmd.Long)
	}
}

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
