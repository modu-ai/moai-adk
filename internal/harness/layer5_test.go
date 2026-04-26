package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldHarnessDir_BaselineSevenFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "harness")
	opts := ScaffoldOpts{Domain: "ios-mobile", SpecID: "SPEC-PROJ-INIT-001"}
	if err := ScaffoldHarnessDir(dir, opts); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	expected := []string{
		"main.md",
		"plan-extension.md",
		"run-extension.md",
		"sync-extension.md",
		"chaining-rules.yaml",
		"interview-results.md",
		"README.md",
	}
	for _, name := range expected {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("missing file %s: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s is empty", name)
		}
	}
	// design-extension.md must NOT exist when IncludeDesignExtension=false.
	if _, err := os.Stat(filepath.Join(dir, "design-extension.md")); !os.IsNotExist(err) {
		t.Errorf("design-extension.md should not exist when IncludeDesignExtension=false")
	}
}

func TestScaffoldHarnessDir_AdvancedAddsDesignExtension(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "harness")
	opts := ScaffoldOpts{Domain: "ios-mobile", SpecID: "S-1", IncludeDesignExtension: true}
	if err := ScaffoldHarnessDir(dir, opts); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "design-extension.md")); err != nil {
		t.Errorf("design-extension.md missing: %v", err)
	}
	// Verify count: 7 baseline + 1 = 8 .md/.yaml files
	entries, _ := os.ReadDir(dir)
	if len(entries) != 8 {
		t.Errorf("expected 8 files, got %d", len(entries))
	}
}

func TestScaffoldHarnessDir_DomainEmbeddedInMain(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "harness")
	opts := ScaffoldOpts{Domain: "android-mobile", SpecID: "S-X"}
	if err := ScaffoldHarnessDir(dir, opts); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "main.md"))
	if !strings.Contains(string(data), "android-mobile") {
		t.Errorf("main.md missing domain")
	}
	if !strings.Contains(string(data), "S-X") {
		t.Errorf("main.md missing SPEC ID")
	}
}

func TestScaffoldHarnessDir_FilePurposeFirstLine(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "harness")
	if err := ScaffoldHarnessDir(dir, ScaffoldOpts{Domain: "d", SpecID: "S"}); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"main.md":              "# Harness Main",
		"plan-extension.md":    "# Plan Phase Harness Extension",
		"run-extension.md":     "# Run Phase Harness Extension",
		"sync-extension.md":    "# Sync Phase Harness Extension",
		"README.md":            "# .moai/harness/",
		"chaining-rules.yaml":  "# .moai/harness/chaining-rules.yaml",
		"interview-results.md": "---",
	}
	for file, prefix := range cases {
		data, _ := os.ReadFile(filepath.Join(dir, file))
		first := firstLine(string(data))
		if !strings.HasPrefix(first, prefix) {
			t.Errorf("%s first line %q lacks expected prefix %q", file, first, prefix)
		}
	}
}

func TestScaffoldHarnessDir_EmptyDir(t *testing.T) {
	if err := ScaffoldHarnessDir("", ScaffoldOpts{}); err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestScaffoldHarnessDir_CreatesParentDir(t *testing.T) {
	// Nested non-existent path — must create.
	dir := filepath.Join(t.TempDir(), "deep", "nested", "harness")
	if err := ScaffoldHarnessDir(dir, ScaffoldOpts{Domain: "d", SpecID: "S"}); err != nil {
		t.Fatalf("expected mkdir-p behavior, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.md")); err != nil {
		t.Errorf("main.md missing in nested path")
	}
}

func TestScaffoldHarnessDir_NoEnsureAllowedCall(t *testing.T) {
	// Smoke test: t.TempDir() absolute path far outside any allowed prefix
	// MUST work because layer5 does not invoke EnsureAllowed.
	dir := filepath.Join(t.TempDir(), "harness")
	if err := ScaffoldHarnessDir(dir, ScaffoldOpts{Domain: "d", SpecID: "S"}); err != nil {
		t.Fatalf("layer5 must not call EnsureAllowed; got: %v", err)
	}
}

func firstLine(s string) string {
	idx := strings.Index(s, "\n")
	if idx < 0 {
		return s
	}
	return s[:idx]
}
