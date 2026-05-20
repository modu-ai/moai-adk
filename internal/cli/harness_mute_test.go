// Package cli — harness mute/mute-list/unmute/verify CLI verb tests (M4 RED).
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHarnessMute_AddCategory verifies that `moai harness mute <category>` appends
// the category to workflow.yaml harness.proposal.mute.categories (REQ-HRA-033).
func TestHarnessMute_AddCategory(t *testing.T) {
	t.Parallel()
	dir := harnessMuteTestProject(t)

	cmd := newHarnessCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"mute", "error-handling", "--project-root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mute error-handling: %v (output: %s)", err, buf.String())
	}

	// Verify category added to workflow.yaml.
	data, err := os.ReadFile(filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatalf("workflow.yaml not found: %v", err)
	}
	if !strings.Contains(string(data), "error-handling") {
		t.Errorf("workflow.yaml missing 'error-handling': %s", data)
	}
}

// TestHarnessMuteList_ShowsCategories verifies that `moai harness mute-list` prints
// the current muted categories from workflow.yaml.
func TestHarnessMuteList_ShowsCategories(t *testing.T) {
	t.Parallel()
	dir := harnessMuteTestProject(t)

	// Pre-populate workflow.yaml with a muted category.
	workflowPath := filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml")
	content := "harness:\n  proposal:\n    mode: immediate\n    mute:\n      categories:\n        - naming\n"
	if err := os.WriteFile(workflowPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newHarnessCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"mute-list", "--project-root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mute-list: %v (output: %s)", err, buf.String())
	}
	if !strings.Contains(buf.String(), "naming") {
		t.Errorf("mute-list output missing 'naming': %s", buf.String())
	}
}

// TestHarnessUnmute_RemovesCategory verifies that `moai harness unmute <category>`
// removes the category from workflow.yaml.
func TestHarnessUnmute_RemovesCategory(t *testing.T) {
	t.Parallel()
	dir := harnessMuteTestProject(t)

	workflowPath := filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml")
	content := "harness:\n  proposal:\n    mode: immediate\n    mute:\n      categories:\n        - error-handling\n        - naming\n"
	if err := os.WriteFile(workflowPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newHarnessCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"unmute", "error-handling", "--project-root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unmute error-handling: %v (output: %s)", err, buf.String())
	}

	data, _ := os.ReadFile(workflowPath)
	if strings.Contains(string(data), "error-handling") {
		t.Errorf("workflow.yaml still contains 'error-handling' after unmute: %s", data)
	}
	if !strings.Contains(string(data), "naming") {
		t.Errorf("workflow.yaml lost 'naming' after unmute: %s", data)
	}
}

// TestHarnessMute_InvalidCategory verifies that an invalid category returns exit code 1
// with HARNESS_LEARNING_MUTE_INVALID_CATEGORY sentinel (REQ-HRA-033).
func TestHarnessMute_InvalidCategory(t *testing.T) {
	t.Parallel()
	dir := harnessMuteTestProject(t)

	cmd := newHarnessCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"mute", "invalid-unknown-xyz", "--project-root", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
	if !strings.Contains(err.Error(), "HARNESS_LEARNING_MUTE_INVALID_CATEGORY") &&
		!strings.Contains(buf.String(), "HARNESS_LEARNING_MUTE_INVALID_CATEGORY") {
		t.Errorf("expected HARNESS_LEARNING_MUTE_INVALID_CATEGORY in output, got: %s | err: %v", buf.String(), err)
	}
}

// TestHarnessVerify_DeferredMessage verifies that `moai harness verify --determinism`
// exits 0 and emits a deferred message (W4 placeholder per plan.md §6.1).
func TestHarnessVerify_DeferredMessage(t *testing.T) {
	t.Parallel()
	dir := harnessMuteTestProject(t)

	cmd := newHarnessCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"verify", "--determinism", "--project-root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify --determinism: %v (output: %s)", err, buf.String())
	}
	output := buf.String()
	if !strings.Contains(output, "deferred") && !strings.Contains(output, "W4") && !strings.Contains(output, "not yet implemented") {
		t.Errorf("verify output missing deferred/W4 message: %s", output)
	}
}

// harnessMuteTestProject creates a minimal temp project with workflow.yaml for mute tests.
func harnessMuteTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("configDir: %v", err)
	}
	// Minimal workflow.yaml.
	workflowContent := "harness:\n  proposal:\n    mode: immediate\n    mute:\n      categories: []\n"
	if err := os.WriteFile(filepath.Join(configDir, "workflow.yaml"), []byte(workflowContent), 0o644); err != nil {
		t.Fatalf("workflow.yaml: %v", err)
	}
	// Minimal harness.yaml needed by existing resolve logic.
	harnessContent := "harness:\n  learning:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "harness.yaml"), []byte(harnessContent), 0o644); err != nil {
		t.Fatalf("harness.yaml: %v", err)
	}
	return dir
}
