package quality

// astgrep_gate_walk_ancestor_test.go covers card t1689 repair round 2 (sync-audit
// finding F1/F2): the walk's directory-name exclusions (vendor/.git/node_modules/
// __pycache__/testdata) must apply only to path segments RELATIVE to the walked
// project root. The pre-fix predicate split the ABSOLUTE path, so a project
// checked out under an ancestor literally named one of the five excluded names
// (e.g. /tmp/testdata/proj) was swept as zero files and the suppression-pairing
// check passed vacuously — an unpaired ast-grep-ignore bypassed
// SUPPRESSION_WITHOUT_REASON enforcement.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeUnpairedIgnoreFixture writes a production .go file carrying an unpaired
// ast-grep-ignore marker (no adjacent @MX:REASON), the exact input the
// suppression-pairing check must flag.
func writeUnpairedIgnoreFixture(t *testing.T, dir string) string {
	t.Helper()
	src := `package prod

func Suppressed() {
	// ast-grep-ignore
	var payload interface{}
	_ = payload
}
`
	p := filepath.Join(dir, "prod.go")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestWalkSourceFiles_ExcludedNameAncestorStillSweeps verifies that an
// excluded-name directory ABOVE the project root does not suppress the sweep:
// the project itself is walked and its unpaired marker is reported.
func TestWalkSourceFiles_ExcludedNameAncestorStillSweeps(t *testing.T) {
	for _, ancestor := range []string{"testdata", "vendor", "node_modules", "__pycache__"} {
		t.Run(ancestor, func(t *testing.T) {
			projectDir := filepath.Join(t.TempDir(), ancestor, "proj")
			if err := os.MkdirAll(projectDir, 0o755); err != nil {
				t.Fatal(err)
			}
			prodFile := writeUnpairedIgnoreFixture(t, projectDir)

			files := walkSourceFiles(projectDir)
			var swept bool
			for _, f := range files {
				if f == prodFile {
					swept = true
				}
			}
			if !swept {
				t.Fatalf("walkSourceFiles skipped the project under a %q ANCESTOR; got %v (pre-fix bug: absolute-path segments exclude every file, the pairing check passes vacuously)", ancestor, files)
			}

			violations := checkSuppressionPairing(prodFile)
			if len(violations) == 0 {
				t.Fatal("checkSuppressionPairing reported no violation for an unpaired ast-grep-ignore marker in a swept production file")
			}
			for _, v := range violations {
				if v.Type != "SUPPRESSION_WITHOUT_REASON" {
					t.Errorf("unexpected violation type %q; want SUPPRESSION_WITHOUT_REASON", v.Type)
				}
			}
		})
	}
}

// TestRunAstGrepGateV2_TestdataAncestorStillBlocks exercises the gate-entry
// surface: a project under a testdata-named ancestor carrying an unpaired
// marker must FAIL at step 1 (suppression pairing, RulesDir empty so no scan
// runs) — not pass vacuously because the walk returned zero files.
func TestRunAstGrepGateV2_TestdataAncestorStillBlocks(t *testing.T) {
	t.Setenv("PATH", "")

	projectDir := filepath.Join(t.TempDir(), "testdata", "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeUnpairedIgnoreFixture(t, projectDir)

	cfg := &AstGrepGateConfig{Enabled: true, RulesDir: ""}
	ok, output := RunAstGrepGateV2(context.Background(), projectDir, cfg)
	if ok {
		t.Errorf("RunAstGrepGateV2 passed vacuously for a project under a testdata-named ancestor with an unpaired marker; the pairing step must block. output:\n%s", output)
	}
	if !strings.Contains(output, "SUPPRESSION_WITHOUT_REASON") {
		t.Errorf("expected a SUPPRESSION_WITHOUT_REASON violation in the gate output; got:\n%s", output)
	}
}

// TestWalkSourceFiles_PositiveControl_NoExcludedAncestor is the control arm:
// the identical fixture NOT under an excluded-name ancestor sweeps and is
// reported both before and after the fix — proving the ancestor cases above
// measure the scoping change, not a broken fixture.
func TestWalkSourceFiles_PositiveControl_NoExcludedAncestor(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "plain", "proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	prodFile := writeUnpairedIgnoreFixture(t, projectDir)

	files := walkSourceFiles(projectDir)
	if len(files) == 0 {
		t.Fatal("positive control failed: walkSourceFiles returned no files for a plain project; the swept set must contain the production file")
	}
	var swept bool
	for _, f := range files {
		if f == prodFile {
			swept = true
		}
	}
	if !swept {
		t.Fatalf("positive control failed: production file not swept; got %v", files)
	}
	if v := checkSuppressionPairing(prodFile); len(v) == 0 {
		t.Fatal("positive control failed: unpaired marker not reported in a plain project")
	}
}
