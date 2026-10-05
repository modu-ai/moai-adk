package cli

// REQ-TRI-007 non-regression guard (AC-TRI-008): the init/update/profile
// surfaces absorbed the huh v2 wizard (card t756 decision D1 — the v2
// absorption is kept), so huh v1 (github.com/charmbracelet/huh) must never
// re-enter the module requirement or the surface packages' imports. The sweep
// is data-driven (scanHuhV1 takes the go.mod text and the source map), which
// is what makes the mutant observations below run through the same code path
// the guard uses in production.

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	// huhV1Path is the huh v1 module/import path. huh v2 lives at
	// charm.land/huh/v2, so a plain substring match cannot false-positive on
	// the v2 dependency.
	huhV1Path = "github.com/charmbracelet/huh"
	// huhV1MinFiles is the floor for a non-vacuous sweep: the two surface
	// directories carry far more than this many Go files today; a sweep below
	// the floor means the directories moved and the guard is scanning nothing.
	huhV1MinFiles = 20
)

// scanHuhV1 requires that huh v1 appears neither in the go.mod text nor in any
// swept source file. Test files are skipped: a v1 test-only import could not
// even compile without the module in go.mod, and the go.mod check is the
// module-level backstop (the guard's own mutant fixtures live in a test file
// and must not trip the production sweep).
func scanHuhV1(goMod string, sources map[string]string) error {
	if strings.Contains(goMod, huhV1Path) {
		return errors.New("go.mod requires huh v1 (" + huhV1Path + ") — the v2 absorption (D1) forbids its return (REQ-TRI-007)")
	}
	if len(sources) < huhV1MinFiles {
		return fmt.Errorf("swept only %d source files, want >= %d — the surface directories moved and the guard would pass vacuously", len(sources), huhV1MinFiles)
	}
	for _, name := range slices.Sorted(maps.Keys(sources)) {
		if strings.Contains(sources[name], "\""+huhV1Path+"\"") {
			return fmt.Errorf("%s imports huh v1 (%s) — the v2 absorption (D1) forbids its return (REQ-TRI-007)", name, huhV1Path)
		}
	}
	return nil
}

// surfaceSources sweeps this package's and the wizard package's .go files
// (the init/update/profile surfaces, SPEC-CLI-TUX-RENDER-I18N-001 §B).
func surfaceSources(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, dir := range []string{".", filepath.Join(".", "wizard")} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			files[filepath.Join(dir, name)] = string(b)
		}
	}
	return files
}

// goModText reads the module's go.mod (two levels above this package).
func goModText(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	return string(b)
}

// TestHuhV1NonRegressionGuard is AC-TRI-008: the production tree carries no
// huh v1 module requirement and no huh v1 import in the surface packages, and
// the guard FAILS on both mutant shapes (go.mod require added; import added).
func TestHuhV1NonRegressionGuard(t *testing.T) {
	goMod := goModText(t)
	sources := surfaceSources(t)

	if err := scanHuhV1(goMod, sources); err != nil {
		t.Fatalf("production tree failed the huh v1 guard: %v", err)
	}
	t.Logf("swept %d production source files + go.mod: no huh v1 (REQ-TRI-007)", len(sources))

	// Mutant (a): huh v1 added to go.mod.
	if err := scanHuhV1(goMod+"\n"+huhV1Path+" v1.14.0 // mutant\n", sources); err == nil {
		t.Fatal("mutant go.mod carrying huh v1 PASSED the guard — the guard is blind to the module requirement")
	} else {
		t.Logf("go.mod mutant correctly rejected: %v", err)
	}

	// Mutant (b): huh v1 import added to a swept source.
	mutant := maps.Clone(sources)
	some := slices.Sorted(maps.Keys(mutant))[0]
	mutant[some] = mutant[some] + "\nvar _ = \"" + huhV1Path + "\" // mutant\n"
	if err := scanHuhV1(goMod, mutant); err == nil {
		t.Fatal("mutant source importing huh v1 PASSED the guard — the guard is blind to imports")
	} else {
		t.Logf("import mutant correctly rejected: %v", err)
	}
}
