package config

import (
	"os"
	"path/filepath"
	"strconv"
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

// TestLoadIdentity_ReturnsStoredValueVerbatim pins that the readers do no
// filtering of their own: values the update render cannot carry are read back
// exactly as stored, and the render-carry decision belongs to the caller
// (internal/cli loadUpdateIdentity, which can reach the renderer).
func TestLoadIdentity_ReturnsStoredValueVerbatim(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"p-gpt", "구스", "o'neil", "a: b # c", "$TEAM", "{{.Version}}", "cost$5", "a{{b", `Kim "Goos"`, `C:\Users\x`, "a\tb", "👩\u200d💻 x"} {
		root := t.TempDir()
		quoted := yamlQuoted(value)
		writeIdentitySection(t, root, "project.yaml", "project:\n  name: "+quoted+"\n")
		writeIdentitySection(t, root, "user.yaml", "user:\n  name: "+quoted+"\n")
		if got := LoadProjectName(root); got != value {
			t.Errorf("LoadProjectName(%q) = %q, want the value verbatim", value, got)
		}
		if got := LoadUserName(root); got != value {
			t.Errorf("LoadUserName(%q) = %q, want the value verbatim", value, got)
		}
	}
}

// yamlQuoted writes s as a YAML double-quoted scalar. Go's escapes for control
// characters (\n, \x00, \u2028, ...) are all valid YAML double-quoted escapes,
// so the fixture file always parses and the loader sees exactly s.
func yamlQuoted(s string) string {
	return strconv.Quote(s)
}
