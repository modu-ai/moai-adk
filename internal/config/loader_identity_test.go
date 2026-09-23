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

// TestLoadIdentity_RejectsValuesTheRenderCannotCarry pins the card t1139
// follow-up (sync-audit F1/F3): the update renders project.yaml / user.yaml
// with these names inside a double-quoted YAML scalar and then runs the
// renderer's unexpanded-token check over the output. A name that trips that
// check halts the whole update at "Validate Templates"; a name that breaks the
// double-quoted scalar makes the 3-way merge fail on every update. Such a value
// must read as "" — the pre-t1139 render — so the merge keeps the user's value
// as a customization instead.
func TestLoadIdentity_RejectsValuesTheRenderCannotCarry(t *testing.T) {
	t.Parallel()
	rejected := map[string]string{
		"dollar env token":    "$TEAM",
		"dollar embedded":     "team-$TEAM-x",
		"dollar brace token":  "${TEAM}",
		"lone dollar":         "cost$",
		"template action":     "{{.Version}}",
		"template open only":  "a{{b",
		"template close only": "a}}b",
		"double quote":        `Kim "Goos"`,
		"backslash":           `C:\Users\x`,
		"newline":             "a\nb",
		"carriage return":     "a\rb",
		"tab":                 "a\tb",
		"nul":                 "a\x00b",
		"escape":              "a\x1bb",
		"del":                 "a\x7fb",
		"next line":           "a\u0085b",
		"line separator":      "a\u2028b",
		"byte order mark":     "a\ufeffb",
	}
	for label, value := range rejected {
		root := t.TempDir()
		quoted := yamlQuoted(value)
		writeIdentitySection(t, root, "project.yaml", "project:\n  name: "+quoted+"\n")
		writeIdentitySection(t, root, "user.yaml", "user:\n  name: "+quoted+"\n")
		if got := LoadProjectName(root); got != "" {
			t.Errorf("%s: LoadProjectName = %q, want empty (value cannot be rendered verbatim)", label, got)
		}
		if got := LoadUserName(root); got != "" {
			t.Errorf("%s: LoadUserName = %q, want empty (value cannot be rendered verbatim)", label, got)
		}
	}
}

// TestLoadIdentity_KeepsOrdinaryValues is the positive control for the
// rejection above: ordinary punctuation and non-ASCII letters pass through.
func TestLoadIdentity_KeepsOrdinaryValues(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"p-gpt", "구스", "Goos Kim", "o'neil", "a: b # c", "{single}", "50%", "a{b}c", "ümlaut_ß"} {
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
