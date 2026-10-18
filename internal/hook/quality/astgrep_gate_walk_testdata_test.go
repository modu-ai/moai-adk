package quality

// astgrep_gate_walk_testdata_test.go covers card t1689: walkSourceFiles must
// skip directories named testdata, mirroring the findings-side exclusion
// (astGrepExcludedPathPatterns "/testdata/"). Shipped fixtures place
// @MX:REASON on the line BEFORE the // ast-grep-ignore marker (ast-grep
// attaches the marker to the line immediately preceding the target code), so
// the pairing checker reads them as unpaired when the walk sweeps testdata.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWalkSourceFiles_SkipsTestdata verifies walkSourceFiles does not return
// files under any directory named testdata, while still returning a sibling
// production file outside testdata (positive control: the swept set is not
// empty, so a skip-everything regression cannot pass vacuously).
func TestWalkSourceFiles_SkipsTestdata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fixtureDir := filepath.Join(root, "testdata", "fixtures", "go")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixtureSrc := `package fixtures

// @MX:REASON test fixture; ordering intentional for ast-grep suppression
// ast-grep-ignore
var payload interface{}
`
	if err := os.WriteFile(filepath.Join(fixtureDir, "x.go"), []byte(fixtureSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	prodSrc := `package prod

func F() {}
`
	if err := os.WriteFile(filepath.Join(root, "prod.go"), []byte(prodSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	files := walkSourceFiles(root)
	if len(files) == 0 {
		t.Fatal("positive control failed: walkSourceFiles returned no files at all; the swept set must contain the production file outside testdata")
	}
	var sawProd bool
	for _, f := range files {
		if strings.Contains(filepath.ToSlash(f), "/testdata/") {
			t.Errorf("walkSourceFiles returned a path under testdata: %q", f)
		}
		if strings.HasSuffix(f, "prod.go") {
			sawProd = true
		}
	}
	if !sawProd {
		t.Errorf("walkSourceFiles did not return the production file outside testdata; got %v", files)
	}
}

// TestRunAstGrepGateV2_TestdataMarkerLayoutDoesNotBlock exercises the
// gate-entry surface over a temp project whose testdata directory carries the
// shipped fixture's exact marker layout (@MX:REASON line, then
// // ast-grep-ignore, then the target code). The gate must pass: testdata
// fixtures are not production code. RulesDir is empty so the sg scan is
// skipped with a reason; the suppression-pairing step (step 1) is the surface
// under test.
func TestRunAstGrepGateV2_TestdataMarkerLayoutDoesNotBlock(t *testing.T) {
	t.Setenv("PATH", "")

	projectDir := t.TempDir()
	fixtureDir := filepath.Join(projectDir, "testdata", "fixtures", "go")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixtureSrc := `package fixtures

func Suppressed() {
	// @MX:REASON test fixture for suppression policy; interface{} intentional
	// ast-grep-ignore
	var payload interface{}
	_ = payload
}
`
	if err := os.WriteFile(filepath.Join(fixtureDir, "suppressed.go"), []byte(fixtureSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &AstGrepGateConfig{Enabled: true, RulesDir: ""}
	ok, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)
	if !ok {
		t.Errorf("RunAstGrepGateV2 should pass when the only suppression-marker file lives under testdata; got ok=false, output:\n%s", output)
	}
}

// TestWalkSourceFiles_SkipsShippedFixture walks the REAL repository tree from
// the package directory and asserts no path under astgrep/testdata/ is swept.
// This is the shipped-fixture characterization: the fixture
// internal/astgrep/testdata/fixtures/go/suppressed.go must never reach
// checkSuppressionPairing via the walk, because reordering its marker lines
// would break ast-grep suppression attachment.
func TestWalkSourceFiles_SkipsShippedFixture(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	files := walkSourceFiles(repoRoot)
	if len(files) == 0 {
		t.Fatal("positive control failed: walkSourceFiles returned no files for the repository root; the swept set must be non-empty")
	}
	for _, f := range files {
		if strings.Contains(filepath.ToSlash(f), "astgrep/testdata/") {
			t.Errorf("walkSourceFiles swept the shipped fixture tree: %q", f)
		}
	}
}
