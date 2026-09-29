// product_jev_boundary_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 AC-TCD-013:
// the scripts/jev product-path absence check, WITH its positive control.
//
// The scan asserts zero references to the local-only Jev tooling
// (scripts/jev) across the product path set internal/, pkg/, cmd/, and
// internal/template/templates/. A zero-hit scan without a control is
// unmeasured (a zero result needs a positive control — the repo's own
// verification rule), so the same scanner is run against a planted fixture
// tree that carries a reference and MUST hit. The control is planted at test
// time under t.TempDir(), so the check is clone-independent: scripts/jev
// itself is local-only and absent from a clean checkout, but the boundary
// contract it enforces is not.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// jevBoundaryScan walks root and returns (hit files, swept file count) for
// the literal "scripts/jev" reference. Only regular non-test-binary text
// files are read; directories named vendor and .git are skipped.
func jevBoundaryScan(t *testing.T, root string) (hits []string, swept int) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", ".git", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		// _test.go files are not product surfaces (the C-HRA-008/B3 boundary
		// grep precedent): prose in a test comment naming the local tool is
		// not an import, an exec, or a path constant in shipped code.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".ico") ||
			strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".gz") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		swept++
		if strings.Contains(string(data), "scripts/jev") {
			hits = append(hits, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan %s: %v", root, err)
	}
	return hits, swept
}

func TestProductPathsCarryNoJevReference(t *testing.T) {
	// The test's own package dir anchors the repo-relative product roots.
	repoRoot := filepath.Join("..", "..")
	for _, rel := range []string{"internal", "pkg", "cmd"} {
		root := filepath.Join(repoRoot, rel)
		hits, swept := jevBoundaryScan(t, root)
		if swept == 0 {
			t.Fatalf("scanned %s and swept zero files — an empty sweep asserts nothing", root)
		}
		if len(hits) > 0 {
			t.Errorf("product path %s carries %d scripts/jev reference(s): %v (REQ-TCD-012: no product path references local-only tooling)",
				root, len(hits), hits)
		}
	}
	// The template mirror is scanned separately and reported with its own
	// swept count so a partial sweep cannot read as a pass.
	hits, swept := jevBoundaryScan(t, filepath.Join(repoRoot, "internal", "template", "templates"))
	if swept == 0 {
		t.Fatalf("scanned the template tree and swept zero files")
	}
	if len(hits) > 0 {
		t.Errorf("template tree carries %d scripts/jev reference(s): %v", len(hits), hits)
	}
}

// TestJevBoundaryScanPositiveControl proves the scanner can fire: the same
// function run against a planted tree carrying a scripts/jev reference must
// hit it. Without this arm the zero above is unmeasured.
func TestJevBoundaryScanPositiveControl(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "local")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	planted := filepath.Join(dir, "jev-operations.md")
	body := "# local operations\n\nRun scripts/jev/triage.sh before dispatch.\n"
	if err := os.WriteFile(planted, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	// A clean file proves the scan is selective, not a constant-true probe.
	clean := filepath.Join(dir, "clean.md")
	if err := os.WriteFile(clean, []byte("# no tooling reference here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hits, swept := jevBoundaryScan(t, root)
	if swept < 2 {
		t.Fatalf("control sweep touched %d files, want both fixtures", swept)
	}
	if len(hits) != 1 {
		t.Fatalf("positive control: hits = %v, want exactly the planted file", hits)
	}
	if !strings.HasSuffix(hits[0], filepath.Join("docs", "local", "jev-operations.md")) {
		t.Errorf("positive control hit %s, want the planted jev-operations.md", hits[0])
	}
}
