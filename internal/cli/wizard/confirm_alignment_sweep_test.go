package wizard

// REQ-TRI-002 source sweep guard (AC-TRI-002): every huh.NewConfirm creation
// site in the wizard package's production sources must specify left button
// alignment. huh's default is lipgloss.Center (huh v2.0.3 field_confirm.go:53)
// — the F11-(1) defect shape — so an unguarded new site silently regresses to
// centered buttons. The sweep is data-driven (scanConfirmAlignment takes the
// source map), which is what makes the mutant observation below run through
// the same code path the guard uses in production.

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

const (
	// confirmCreateMarker opens a huh confirm creation site.
	confirmCreateMarker = "huh.NewConfirm("
	// alignmentMarker is the required left-alignment specification.
	alignmentMarker = "WithButtonAlignment(lipgloss.Left)"
	// alignmentWindow bounds the constructor chain a site may use before the
	// alignment specification must appear (one fluent call chain, generously).
	alignmentWindow = 800
)

// scanConfirmAlignment sweeps the given sources for huh.NewConfirm creation
// sites and requires each to carry WithButtonAlignment(lipgloss.Left) within
// the alignment window. It returns the number of sites swept via the sites
// pointer when non-nil. Test files are skipped: they are not user surfaces.
func scanConfirmAlignment(files map[string]string) (int, error) {
	sites := 0
	for _, name := range slices.Sorted(maps.Keys(files)) {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		rest := files[name]
		for {
			idx := strings.Index(rest, confirmCreateMarker)
			if idx < 0 {
				break
			}
			sites++
			end := min(idx+alignmentWindow, len(rest))
			if !strings.Contains(rest[idx:end], alignmentMarker) {
				return sites, fmt.Errorf("%s: huh.NewConfirm site lacks %s — huh's default Center alignment is the F11-(1) defect shape (REQ-TRI-002)", name, alignmentMarker)
			}
			rest = rest[idx+len(confirmCreateMarker):]
		}
	}
	if sites == 0 {
		return 0, errors.New("confirm alignment sweep found 0 huh.NewConfirm sites — a zero-site sweep asserts nothing; if the constructor moved, re-point " + confirmCreateMarker)
	}
	return sites, nil
}

// packageProdSources reads this package's .go sources from the test's working
// directory.
func packageProdSources(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	files := make(map[string]string, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		files[name] = string(b)
	}
	return files
}

// TestConfirmAlignmentSweep is AC-TRI-002: on the production tree every
// huh.NewConfirm site specifies left alignment, and the guard FAILS on a
// mutant whose alignment specification is removed (observed failure, not a
// vacuous green).
func TestConfirmAlignmentSweep(t *testing.T) {
	files := packageProdSources(t)

	sites, err := scanConfirmAlignment(files)
	if err != nil {
		t.Fatalf("production tree failed the confirm alignment sweep: %v", err)
	}
	t.Logf("swept %d huh.NewConfirm site(s), all left-aligned", sites)

	// Mutant: strip the alignment specification from every production source
	// and require the SAME sweep function to reject the result. The marker is
	// replaced bare — gofmt puts the fluent dot at the end of the PREVIOUS
	// line, so a dot-prefixed pattern would match nothing and the mutant would
	// be a silent no-op (observed and fixed during M2). A mutant pass would
	// mean the guard cannot detect the defect it exists for.
	mutant := maps.Clone(files)
	for name, src := range mutant {
		if !strings.HasSuffix(name, "_test.go") {
			mutant[name] = strings.ReplaceAll(src, alignmentMarker, "")
		}
	}
	if _, err := scanConfirmAlignment(mutant); err == nil {
		t.Fatal("mutant with alignment specifications removed PASSED the sweep — the guard is blind to the F11-(1) defect shape (REQ-TRI-002)")
	} else {
		t.Logf("mutant correctly rejected: %v", err)
	}
}
