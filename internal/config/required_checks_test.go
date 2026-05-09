package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadRequiredChecks_Main verifies that the main branch contexts include
// required CI checks after loading the SSoT YAML.
func TestLoadRequiredChecks_Main(t *testing.T) {
	dir := t.TempDir()
	writeRequiredChecksYAML(t, dir)

	rc, err := LoadRequiredChecks(dir)
	if err != nil {
		t.Fatalf("LoadRequiredChecks() unexpected error: %v", err)
	}

	main, ok := rc.Branches["main"]
	if !ok {
		t.Fatalf("expected branch 'main' in RequiredChecks, branches: %v", rc.Branches)
	}

	wantContexts := []string{"Lint", "Test (ubuntu-latest)"}
	for _, want := range wantContexts {
		found := false
		for _, c := range main.Contexts {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("main contexts missing %q; got: %v", want, main.Contexts)
		}
	}
}

// TestLoadRequiredChecks_ReleaseGlob verifies that the release/* branch contexts
// contain exactly 3 required checks.
func TestLoadRequiredChecks_ReleaseGlob(t *testing.T) {
	dir := t.TempDir()
	writeRequiredChecksYAML(t, dir)

	rc, err := LoadRequiredChecks(dir)
	if err != nil {
		t.Fatalf("LoadRequiredChecks() unexpected error: %v", err)
	}

	rel, ok := rc.Branches["release/*"]
	if !ok {
		t.Fatalf("expected branch 'release/*' in RequiredChecks, branches: %v", rc.Branches)
	}

	if len(rel.Contexts) != 3 {
		t.Errorf("release/* contexts: want 3 items, got %d: %v", len(rel.Contexts), rel.Contexts)
	}
}

// TestLoadRequiredChecks_Missing verifies that missing YAML returns an error
// with a helpful message indicating where to place the file.
func TestLoadRequiredChecks_Missing(t *testing.T) {
	dir := t.TempDir()
	// Intentionally do NOT create the YAML file.

	_, err := LoadRequiredChecks(dir)
	if err == nil {
		t.Fatal("LoadRequiredChecks() expected error for missing file, got nil")
	}

	if !strings.Contains(err.Error(), ".github/required-checks.yml") {
		t.Errorf("error should mention file path, got: %v", err)
	}
}

// TestNoHardcodedContexts verifies that branch_protection.go and github_init.go
// do NOT contain the literal string "Test (ubuntu-latest)".
// Only the loader file (required_checks.go) may reference context names.
func TestNoHardcodedContexts(t *testing.T) {
	// Find project root: walk up from this test's directory until go.mod is found.
	root := findProjectRoot(t)

	filesToCheck := []string{
		filepath.Join(root, "internal", "cli", "branch_protection.go"),
		filepath.Join(root, "internal", "cli", "github_init.go"),
	}

	forbidden := "Test (ubuntu-latest)"

	for _, f := range filesToCheck {
		data, err := os.ReadFile(f)
		if err != nil {
			// File may not exist yet (branch_protection.go created in W1-T06).
			t.Logf("skipping %s (not found): %v", f, err)
			continue
		}
		if strings.Contains(string(data), forbidden) {
			t.Errorf("file %s must not contain hardcoded context %q", f, forbidden)
		}
	}

	// Also check scripts/ci-mirror directory.
	scriptsDir := filepath.Join(root, "scripts")
	if _, err := os.Stat(scriptsDir); err == nil {
		err = filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			data, err2 := os.ReadFile(path)
			if err2 != nil {
				return nil
			}
			if strings.Contains(string(data), forbidden) {
				t.Errorf("scripts file %s must not contain hardcoded context %q", path, forbidden)
			}
			return nil
		})
		if err != nil {
			t.Logf("walk error (non-fatal): %v", err)
		}
	}
}

// writeRequiredChecksYAML creates a .github/required-checks.yml in the given dir.
func writeRequiredChecksYAML(t *testing.T, dir string) {
	t.Helper()
	ghDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(ghDir, 0o755); err != nil {
		t.Fatalf("mkdir .github: %v", err)
	}
	content := `version: 1
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
`
	p := filepath.Join(ghDir, "required-checks.yml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write required-checks.yml: %v", err)
	}
}

// findProjectRoot walks up the directory tree from the test binary location
// until it finds a go.mod file.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}
