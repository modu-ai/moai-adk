package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeIdentitySection(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestLoadProjectAndUserName pins that the readers return the stored value
// verbatim (no trimming), so re-rendering it reproduces the stored text.
func TestLoadProjectAndUserName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeIdentitySection(t, root, "project.yaml", "project:\n  name: \"p-gpt \"\n  description: \"\"\n")
	writeIdentitySection(t, root, "user.yaml", "user:\n  name: \"goos\"\n")

	if got := LoadProjectName(root); got != "p-gpt " {
		t.Errorf("LoadProjectName = %q, want %q", got, "p-gpt ")
	}
	if got := LoadUserName(root); got != "goos" {
		t.Errorf("LoadUserName = %q, want %q", got, "goos")
	}
}

func TestLoadProjectAndUserName_FallbackEmpty(t *testing.T) {
	t.Parallel()
	missing := t.TempDir()
	if got := LoadProjectName(missing); got != "" {
		t.Errorf("missing project.yaml: LoadProjectName = %q, want empty", got)
	}
	if got := LoadUserName(missing); got != "" {
		t.Errorf("missing user.yaml: LoadUserName = %q, want empty", got)
	}

	broken := t.TempDir()
	writeIdentitySection(t, broken, "project.yaml", "project: [unclosed\n")
	writeIdentitySection(t, broken, "user.yaml", "user:\n  other: x\n")
	if got := LoadProjectName(broken); got != "" {
		t.Errorf("unparseable project.yaml: LoadProjectName = %q, want empty", got)
	}
	if got := LoadUserName(broken); got != "" {
		t.Errorf("absent user.name key: LoadUserName = %q, want empty", got)
	}
}
