package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRequiredChecks_Auxiliary verifies that the SSoT YAML's `auxiliary:`
// list (REQ-CIAUT-009 / AC-CIAUT-021) is parsed correctly and IsAuxiliary
// returns true exactly for the listed contexts.
func TestLoadRequiredChecks_Auxiliary(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".github"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	yaml := `version: 1
branches:
  main:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
auxiliary:
  - claude-code-review
  - llm-panel
  - docs-i18n-check
`
	if err := os.WriteFile(filepath.Join(tmp, RequiredChecksFile), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	rc, err := LoadRequiredChecks(tmp)
	if err != nil {
		t.Fatalf("LoadRequiredChecks: %v", err)
	}

	if got, want := len(rc.Auxiliary), 3; got != want {
		t.Fatalf("Auxiliary length = %d, want %d (REQ-CIAUT-009 mandates 3 auxiliary checks)", got, want)
	}

	tests := []struct {
		ctx        string
		isAuxWant  bool
	}{
		{"claude-code-review", true},
		{"llm-panel", true},
		{"docs-i18n-check", true},
		{"Lint", false},
		{"Test (ubuntu-latest)", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.ctx, func(t *testing.T) {
			if got := rc.IsAuxiliary(tt.ctx); got != tt.isAuxWant {
				t.Errorf("IsAuxiliary(%q) = %v, want %v", tt.ctx, got, tt.isAuxWant)
			}
		})
	}
}
