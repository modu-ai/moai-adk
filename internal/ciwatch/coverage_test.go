package ciwatch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

const testYAML = `version: 1
branches:
  main:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
      - CodeQL
  release/*:
    contexts:
      - Lint
auxiliary:
  - claude-code-review
`

func setupClassifier(t *testing.T) (*ciwatch.Classifier, string) {
	t.Helper()
	dir := t.TempDir()
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "required-checks.yml"), []byte(testYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := ciwatch.NewClassifier(dir)
	if err != nil {
		t.Fatal(err)
	}
	return c, dir
}

func TestClassifier_RequiredNames(t *testing.T) {
	c, _ := setupClassifier(t)

	names := c.RequiredNames("main")
	if len(names) == 0 {
		t.Error("expected non-empty RequiredNames for main")
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["Lint"] {
		t.Error("expected Lint in RequiredNames for main")
	}
	if !found["CodeQL"] {
		t.Error("expected CodeQL in RequiredNames for main")
	}

	// claude-code-review is auxiliary — must not appear.
	if found["claude-code-review"] {
		t.Error("auxiliary check must not appear in RequiredNames")
	}
}

func TestClassifier_RequiredNames_NoMatch(t *testing.T) {
	c, _ := setupClassifier(t)
	names := c.RequiredNames("develop") // not in SSoT
	if len(names) != 0 {
		t.Errorf("expected empty RequiredNames for unknown branch, got %v", names)
	}
}

func TestClassifier_AuxiliaryNames(t *testing.T) {
	c, _ := setupClassifier(t)
	names := c.AuxiliaryNames()
	if len(names) != 1 {
		t.Fatalf("expected 1 auxiliary, got %d", len(names))
	}
	if names[0] != "claude-code-review" {
		t.Errorf("expected claude-code-review, got %q", names[0])
	}
}

func TestClassifier_IsAuxiliary(t *testing.T) {
	c, _ := setupClassifier(t)
	if !c.IsAuxiliary("claude-code-review") {
		t.Error("expected claude-code-review to be auxiliary")
	}
	if c.IsAuxiliary("Lint") {
		t.Error("expected Lint NOT to be auxiliary")
	}
}

func TestClassifier_IsKnown(t *testing.T) {
	c, _ := setupClassifier(t)
	if !c.IsKnown("Lint") {
		t.Error("Lint should be known")
	}
	if !c.IsKnown("claude-code-review") {
		t.Error("claude-code-review (auxiliary) should be known")
	}
	if c.IsKnown("some-random-check") {
		t.Error("unknown check should not be known")
	}
}

func TestClassifier_GlobPattern_Release(t *testing.T) {
	c, _ := setupClassifier(t)
	// release/* pattern in SSoT matches release/v1.0.0
	if !c.IsRequired("Lint", "release/v1.0.0") {
		t.Error("expected Lint required for release/v1.0.0")
	}
	// CodeQL is only in main, not release/*
	if c.IsRequired("CodeQL", "release/v2.0.0") {
		t.Error("CodeQL should not be required for release/* branch")
	}
}

func TestStateFile_DeleteState(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	ws := ciwatch.WatchState{PRNumber: 999}
	if err := ciwatch.WriteState(statePath, ws); err != nil {
		t.Fatal(err)
	}

	if err := ciwatch.DeleteState(statePath); err != nil {
		t.Fatalf("DeleteState: %v", err)
	}

	// File should be gone.
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("state file should be deleted")
	}

	// Double-delete should not error (graceful).
	if err := ciwatch.DeleteState(statePath); err != nil {
		t.Errorf("DeleteState on missing file should not error: %v", err)
	}
}

func TestStateFile_WriteState_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	// Nested dir that doesn't exist yet.
	statePath := filepath.Join(dir, "subdir", "nested", "ci-watch-active.flag")
	ws := ciwatch.WatchState{PRNumber: 42}
	if err := ciwatch.WriteState(statePath, ws); err != nil {
		t.Fatalf("WriteState should create dirs: %v", err)
	}
	if _, err := os.Stat(statePath); err != nil {
		t.Errorf("state file not created: %v", err)
	}
}

func TestHandoff_ItaoNegative(t *testing.T) {
	// Test FormatStatusUpdate with zero values.
	state := ciwatch.CIState{PRNumber: 0, Branch: ""}
	out := ciwatch.FormatStatusUpdate(state)
	if out == "" {
		t.Error("expected non-empty output for zero state")
	}
}

func TestFormatStatusUpdate_LongOutput(t *testing.T) {
	// Edge case: many checks — ensure 200 char limit is respected.
	var failed []ciwatch.CheckResult
	for i := 0; i < 20; i++ {
		failed = append(failed, ciwatch.CheckResult{Name: "Check-With-A-Very-Long-Name", Conclusion: "failure"})
	}
	state := ciwatch.CIState{
		PRNumber:       9999,
		RequiredFailed: failed,
		RequiredPassed: 50,
	}
	out := ciwatch.FormatStatusUpdate(state)
	if len(out) > 200 {
		t.Errorf("output length %d exceeds 200 chars", len(out))
	}
}
