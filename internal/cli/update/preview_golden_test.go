package update

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// preview_golden_test.go is MECHANISM 1 of the split presentation-regression
// coverage (SPEC-CLI-TUI-MODERNIZE-001 D-5): structure goldens captured under a
// forced monochrome axis.
//
// These goldens pin LAYOUT — borders, column alignment, row order, card shape —
// and are deliberately PALETTE-INSENSITIVE: the monochrome axis carries no
// colour tokens, so editing a value in internal/tui cannot break an
// internal/cli/update test. That insensitivity is the point, and it is also the
// limit: removing the theme wiring entirely would leave these goldens unchanged
// and GREEN. The regression-detection burden for that defect class sits on
// mechanism 2 (preview_presentation_test.go), not here. See AC-TUIM-030a/030c.
//
// Regenerate with:
//
//	go test ./internal/cli/update/ -run TestPreviewStructureGolden -update-golden

var updateGolden = flag.Bool("update-golden", false, "rewrite the preview structure goldens")

// assertGolden compares got against testdata/<name>, or rewrites it under
// -update-golden.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("golden %s rewritten (%d bytes)", path, len(got))
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (regenerate with -update-golden)", path, err)
	}
	if got != string(want) {
		t.Errorf("structure golden %s mismatch (AC-TUIM-030a)\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

// TestPreviewStructureGoldenTable pins the table sub-view's structure.
func TestPreviewStructureGoldenTable(t *testing.T) {
	withPreviewTheme(t, tui.MonochromeTheme())
	assertGolden(t, "preview_table_mono.golden", newFixtureModel().tableView())
}

// TestPreviewStructureGoldenDiff pins the diff sub-view's structure.
func TestPreviewStructureGoldenDiff(t *testing.T) {
	withPreviewTheme(t, tui.MonochromeTheme())
	assertGolden(t, "preview_diff_mono.golden", newFixtureModel().selectRow().diffView())
}

// TestPreviewStructureGoldenFallback pins the plain-text fallback's layout
// (AC-TUIM-013). The fallback is structurally colour-free, so this golden needs
// no axis forcing.
func TestPreviewStructureGoldenFallback(t *testing.T) {
	classes := classifyAll(allFourClassesInputs(), allFourClassesPredicate())
	assertGolden(t, "preview_fallback.golden", renderFallback(classes, true))
}

// TestPreviewStructureGoldensAreANSIFree guards the goldens themselves: a
// structure golden carrying an escape sequence would silently re-couple these
// tests to the palette, defeating the split.
func TestPreviewStructureGoldensAreANSIFree(t *testing.T) {
	for _, name := range []string{"preview_table_mono.golden", "preview_diff_mono.golden", "preview_fallback.golden"} {
		b, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if n := strings.Count(string(b), "\x1b"); n != 0 {
			t.Errorf("golden %s carries %d escape bytes; structure goldens must be palette-insensitive (AC-TUIM-030a)", name, n)
		}
	}
}
