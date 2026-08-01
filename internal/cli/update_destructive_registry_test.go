package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Drift guard for the destructive-target registry
// (SPEC-UPDATE-DATA-SURVIVAL-001 M2, AC-UDS-005; REQ-UDS-006/007/010).
//
// The guard enumerates destructive call sites by PARSING THE GO SOURCE, never
// by reading destructiveTargetRegistry. Deriving both sides of the comparison
// from the registry would yield count == count and pass forever (REQ-UDS-007);
// the source scan is what makes this test able to fail when a new destructive
// site is added without a registry row.
//
// Falsification: acceptance.md §C.4 injects an unregistered os.RemoveAll site
// and requires an observed --- FAIL. Because the scan reads the on-disk tree
// via go/parser, `go test -overlay` is invisible to it; the injection must be
// applied in a scratch `git worktree` driven with `go -C` (§C.4 caveat).

// destructiveScanDir is scanned recursively; destructiveScanGlob is scanned
// flat. Together they reproduce the acceptance-criteria scan scope
// `internal/cli/update/ internal/cli/update*.go`, excluding _test.go files.
const (
	destructiveScanDir  = "internal/cli/update"
	destructiveScanGlob = "internal/cli/update*.go"
)

// destructiveFuncs are the irreversible os operations the registry tracks.
var destructiveFuncs = map[string]bool{"RemoveAll": true, "Rename": true}

func TestDestructiveTargetRegistry_CoversAllSites(t *testing.T) {
	root := moduleRootForScan(t)
	scanned := scanDestructiveSites(t, root)

	// The registry side, keyed identically to the scan.
	registered := make(map[string]destructiveSite, len(destructiveTargetRegistry))
	for _, row := range destructiveTargetRegistry {
		if prev, dup := registered[row.key()]; dup {
			t.Errorf("duplicate registry row for %q (sites %d and %d)", row.key(), prev.Sites, row.Sites)
			continue
		}
		registered[row.key()] = row
	}

	// Every scanned site must be registered with a matching occurrence count.
	for _, key := range sortedKeys(scanned) {
		row, ok := registered[key]
		if !ok {
			t.Errorf("unregistered destructive site: %s has %d os.RemoveAll/os.Rename call site(s) but no registry row", key, scanned[key])
			continue
		}
		if row.Sites != scanned[key] {
			t.Errorf("site count drift for %s: registry records %d, source has %d", key, row.Sites, scanned[key])
		}
	}

	// Every registry row must still correspond to a real site.
	for _, key := range sortedRegistryKeys(registered) {
		if _, ok := scanned[key]; !ok {
			t.Errorf("stale registry row: %s is registered but the source has no destructive call site there", key)
		}
	}

	// Every row carries exactly one of a protection assignment or an exemption
	// reason (REQ-UDS-007). A row with neither is an unassigned destructive
	// site; a row with both hides which one actually applies.
	for _, row := range destructiveTargetRegistry {
		hasProtection := strings.TrimSpace(row.Protection) != ""
		hasExemption := strings.TrimSpace(row.Exemption) != ""
		switch {
		case !hasProtection && !hasExemption:
			t.Errorf("registry row %s carries neither a protection assignment nor an exemption reason", row.key())
		case hasProtection && hasExemption:
			t.Errorf("registry row %s carries both a protection assignment and an exemption reason", row.key())
		}
	}

	if t.Failed() {
		t.Logf("scanned %d (file, function) pair(s) across %s and %s; registry has %d row(s)",
			len(scanned), destructiveScanDir, destructiveScanGlob, len(destructiveTargetRegistry))
	}
}

// scanDestructiveSites parses every non-test Go file in scope and returns the
// number of os.RemoveAll / os.Rename call sites per "file function" key.
func scanDestructiveSites(t *testing.T, root string) map[string]int {
	t.Helper()

	files := destructiveScanFiles(t, root)
	if len(files) == 0 {
		t.Fatalf("scan found no source files under %s — the scan scope is broken, so a PASS would be vacuous", root)
	}

	counts := make(map[string]int)
	fset := token.NewFileSet()

	for _, abs := range files {
		file, err := parser.ParseFile(fset, abs, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", abs, err)
		}

		rel, err := filepath.Rel(root, abs)
		if err != nil {
			t.Fatalf("relativize %s: %v", abs, err)
		}
		rel = filepath.ToSlash(rel)

		inFuncs := 0
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if n := countDestructiveCalls(fn.Body); n > 0 {
				counts[rel+" "+fn.Name.Name] += n
				inFuncs += n
			}
		}

		// A destructive call outside any function declaration (a package-level
		// var initializer, say) would be silently invisible to the per-function
		// tally above, so the registry could never cover it. Fail loudly.
		if total := countDestructiveCalls(file); total != inFuncs {
			t.Fatalf("%s: %d destructive call site(s) found, but only %d are inside a function declaration", rel, total, inFuncs)
		}
	}

	return counts
}

// countDestructiveCalls reports how many os.RemoveAll / os.Rename calls appear
// anywhere within n, including inside nested function literals.
func countDestructiveCalls(n ast.Node) int {
	count := 0
	ast.Inspect(n, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "os" {
			return true
		}
		if destructiveFuncs[sel.Sel.Name] {
			count++
		}
		return true
	})
	return count
}

// destructiveScanFiles returns the absolute paths of every non-test Go file in
// the scan scope.
func destructiveScanFiles(t *testing.T, root string) []string {
	t.Helper()

	seen := make(map[string]bool)
	var files []string
	add := func(path string) {
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") || seen[path] {
			return
		}
		seen[path] = true
		files = append(files, path)
	}

	dir := filepath.Join(root, filepath.FromSlash(destructiveScanDir))
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			add(path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}

	matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(destructiveScanGlob)))
	if err != nil {
		t.Fatalf("glob %s: %v", destructiveScanGlob, err)
	}
	for _, m := range matches {
		add(m)
	}

	sort.Strings(files)
	return files
}

// moduleRootForScan walks up from the test's working directory to the module
// root, so the scan reads the tree the test is running against — including a
// scratch worktree used for falsification.
func moduleRootForScan(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod found above the working directory")
		}
		dir = parent
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedRegistryKeys(m map[string]destructiveSite) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
