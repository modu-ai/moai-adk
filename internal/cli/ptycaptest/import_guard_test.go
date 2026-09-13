package ptycaptest

// Import guard: production code (any non-_test.go file in the module) must not
// import this package. `go list` reads imports without compiling anything, so
// the guard can scan every package of the module, internal/cli included.

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

// importListing is the go list template: two lines per package, "P <path>
// <non-test imports...>" and "T <path> <_test.go imports...>".
const importListing = `P {{.ImportPath}}{{range .Imports}} {{.}}{{end}}{{"\n"}}T {{.ImportPath}}{{range .TestImports}} {{.}}{{end}}{{range .XTestImports}} {{.}}{{end}}`

func goList(t *testing.T, dir string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("go list %v: %v\n%s", args, err, stderr)
	}
	return strings.TrimSpace(string(out))
}

// importersOf returns the packages whose kind line ("P" or "T") in listing
// lists target among its imports.
func importersOf(listing, kind, target string) []string {
	var hits []string
	for _, line := range strings.Split(listing, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] != kind {
			continue
		}
		if slices.Contains(f[2:], target) {
			hits = append(hits, f[1])
		}
	}
	return hits
}

// listedPackages returns the package paths that have a line of kind in listing.
func listedPackages(listing, kind string) []string {
	var pkgs []string
	for _, line := range strings.Split(listing, "\n") {
		if f := strings.Fields(line); len(f) >= 2 && f[0] == kind {
			pkgs = append(pkgs, f[1])
		}
	}
	return pkgs
}

// checkNoProductionImport is the guard's verdict over a listing: every package
// whose non-test imports include target (empty means the guard passes).
func checkNoProductionImport(listing, target string) []string {
	return importersOf(listing, "P", target)
}

func TestNoProductionImport(t *testing.T) {
	self := goList(t, ".", "-f", "{{.ImportPath}}", ".")
	root := goList(t, ".", "-m", "-f", "{{.Dir}}")
	listing := goList(t, root, "-f", importListing, "./...")

	// Positive existence first: the scan covered the module, this package
	// among it, and the wizard package whose tests use the harness.
	pkgs := listedPackages(listing, "P")
	wizard := strings.TrimSuffix(self, "/ptycaptest") + "/wizard"
	t.Logf("go list scanned %d packages from %s; self %s", len(pkgs), root, self)
	if len(pkgs) < 50 || !slices.Contains(pkgs, self) || !slices.Contains(pkgs, wizard) {
		t.Fatalf("scan does not cover the module: %d packages, self listed %v, wizard listed %v",
			len(pkgs), slices.Contains(pkgs, self), slices.Contains(pkgs, wizard))
	}

	// Positive control: the same scan sees the import where it exists — in
	// the wizard package's _test.go files — so an empty verdict below is not
	// the matcher failing to see anything.
	if testers := importersOf(listing, "T", self); !slices.Contains(testers, wizard) {
		t.Fatalf("positive control: %s test imports of %s not seen; test importers %v", wizard, self, testers)
	} else {
		t.Logf("test-only importers (allowed): %v", testers)
	}

	if hits := checkNoProductionImport(listing, self); len(hits) > 0 {
		t.Fatalf("production (non-_test.go) code imports the test-only package %s: %v", self, hits)
	}
}

func TestCheckNoProductionImport_Synthetic(t *testing.T) {
	const target = "example.com/m/internal/cli/ptycaptest"
	clean := "P example.com/m/a example.com/m/b\nT example.com/m/a " + target + "\nP " + target + " testing"
	if hits := checkNoProductionImport(clean, target); len(hits) != 0 {
		t.Errorf("test-only import reported as production: %v", hits)
	}
	dirty := clean + "\nP example.com/m/prod fmt " + target
	if hits := checkNoProductionImport(dirty, target); !slices.Equal(hits, []string{"example.com/m/prod"}) {
		t.Errorf("production import not reported: %v", hits)
	}
	// A path that merely starts with the target is a different package.
	near := "P example.com/m/x " + target + "/sub"
	if hits := checkNoProductionImport(near, target); len(hits) != 0 {
		t.Errorf("prefix match reported as an import: %v", hits)
	}
}
