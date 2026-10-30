package web

// browser_select_readback_test.go — AC-AF-020 (f), the read-back half: a
// selection that is in the project's workflow section reads back as the same
// value, for each of the three choices. The saving is the settings layer's
// (lane 17); this test writes the file itself and reads it through the same
// function the page uses.
//
// Defect answer: these tests catch a read that normalizes, trims or defaults the
// operator's choice, and a missing file or key that is read as a selection.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserSelect_SavedValueReadsBack(t *testing.T) {
	for _, sel := range []string{"egolite", "aside", "chrome"} {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		doc := "workflow:\n  agentic_factory:\n    browser:\n      selected: " + sel + "\n"
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(doc), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := readBrowserSelected(root)
		if err != nil || got != sel {
			t.Errorf("saved %q read back as %q, %v", sel, got, err)
		}
	}
	// Nothing saved is no selection (the resolver then uses chrome without saving it).
	if got, err := readBrowserSelected(t.TempDir()); err != nil || got != "" {
		t.Errorf("no file read as %q, %v", got, err)
	}
	other := t.TempDir()
	dir := filepath.Join(other, ".moai", "config", "sections")
	_ = os.MkdirAll(dir, 0o700)
	_ = os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte("workflow:\n  other: 1\n"), 0o600)
	if got, err := readBrowserSelected(other); err != nil || got != "" {
		t.Errorf("a file without the key read as %q, %v", got, err)
	}
}
