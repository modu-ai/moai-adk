package ciwatch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

// TestIsRequired_TableDriven validates required-checks SSoT classification
// for 6 cases per AC-CIAUT-005.
func TestIsRequired_TableDriven(t *testing.T) {
	// Minimal valid required-checks.yml fixture.
	const yamlContent = `version: 1
branches:
  main:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
      - "Test (macos-latest)"
      - "Test (windows-latest)"
      - "Build (linux/amd64)"
      - CodeQL
  release/*:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
      - "Test (macos-latest)"
auxiliary:
  - claude-code-review
  - llm-panel
  - docs-i18n-check
`

	dir := t.TempDir()
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ghDir, "required-checks.yml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := ciwatch.NewClassifier(dir)
	if err != nil {
		t.Fatalf("NewClassifier: %v", err)
	}

	tests := []struct {
		name       string
		check      string
		branch     string
		wantReq    bool
	}{
		// Case 1: main branch, required check
		{
			name:    "main_required_lint",
			check:   "Lint",
			branch:  "main",
			wantReq: true,
		},
		// Case 2: main branch, required check with spaces
		{
			name:    "main_required_test_ubuntu",
			check:   "Test (ubuntu-latest)",
			branch:  "main",
			wantReq: true,
		},
		// Case 3: release/* branch, required check
		{
			name:    "release_required_lint",
			check:   "Lint",
			branch:  "release/v1.0.0",
			wantReq: true,
		},
		// Case 4: auxiliary check — NOT required regardless of branch
		{
			name:    "auxiliary_not_required",
			check:   "claude-code-review",
			branch:  "main",
			wantReq: false,
		},
		// Case 5: unknown check name — not in any list
		{
			name:    "unknown_check_not_required",
			check:   "some-unknown-check",
			branch:  "main",
			wantReq: false,
		},
		// Case 6: release/* branch, check only in main (CodeQL not in release/*)
		{
			name:    "release_branch_main_only_check",
			check:   "CodeQL",
			branch:  "release/v2.0.0",
			wantReq: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.IsRequired(tt.check, tt.branch)
			if got != tt.wantReq {
				t.Errorf("IsRequired(%q, %q) = %v, want %v", tt.check, tt.branch, got, tt.wantReq)
			}
		})
	}
}

// TestNewClassifier_MissingFile verifies error when SSoT file is absent.
func TestNewClassifier_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ciwatch.NewClassifier(dir)
	if err == nil {
		t.Fatal("expected error for missing required-checks.yml, got nil")
	}
}

// TestNewClassifier_EmptySSoT verifies graceful handling of empty YAML.
func TestNewClassifier_EmptySSoT(t *testing.T) {
	dir := t.TempDir()
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid YAML but empty structures.
	const emptyContent = "version: 1\n"
	if err := os.WriteFile(filepath.Join(ghDir, "required-checks.yml"), []byte(emptyContent), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := ciwatch.NewClassifier(dir)
	if err != nil {
		t.Fatalf("NewClassifier with empty SSoT: %v", err)
	}
	// No checks registered — everything is not required.
	if c.IsRequired("Lint", "main") {
		t.Error("expected IsRequired to return false for empty SSoT")
	}
}
