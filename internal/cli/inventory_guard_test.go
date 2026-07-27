package cli

// Inventory guard (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-006 / AC-CFS-026).
//
// The TestMain warm-up in main_test.go closes the cobra lazy-sort race only for
// command globals reachable from the package's warm-up root. A NEW parallel test
// that calls Commands() on a global registered outside that subtree would
// silently reintroduce the race, and no existing test would notice.
//
// This guard re-derives the candidate inventory from source on every run:
//
//	step 0 — build the warm-up reachable set STATICALLY from <parent>.AddCommand(<ident>)
//	         calls in non-test files, then take the transitive closure from the root.
//	         A hand-maintained literal list is deliberately NOT used: it would go
//	         stale at exactly the moment a new global is registered, which is the
//	         recurrence this guard exists to catch. Runtime traversal is not an
//	         option either — Go offers no reflection over package-level variable
//	         NAMES, so mapping a live *cobra.Command back to its identifier would
//	         require the same hand-maintained bridge.
//	step 1 — collect _test.go functions whose body contains BOTH t.Parallel() and
//	         a .Commands() call (the acceptance.md §D.5 syntactic filter).
//	step 2 — resolve each .Commands() receiver. Only a plain identifier naming a
//	         package-level var in a non-test file counts as a shared global.
//	step 3 — a shared global outside the reachable set is a violation.
//
// Known limit (acceptance.md EC-8, spec.md §H residual risk 3): a receiver that
// is a local variable, a function-call result, or a shadowed name is classified
// conservatively as local and passes. This blocks the common recurrence path,
// not every conceivable one.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// warmUpRoots maps a package directory (relative to internal/cli) to the global
// its TestMain warm-up starts from.
var warmUpRoots = map[string]string{
	".":          "rootCmd",
	"preference": "PreferenceCmd",
}

// parsedFile pairs an AST with its base filename for violation messages.
type parsedFile struct {
	name string
	ast  *ast.File
}

// parseDirFiles parses every .go file in dir, split into non-test and test sets.
func parseDirFiles(t *testing.T, dir string) (nonTest, test []parsedFile) {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		pf := parsedFile{name: name, ast: f}
		if strings.HasSuffix(name, "_test.go") {
			test = append(test, pf)
		} else {
			nonTest = append(nonTest, pf)
		}
	}
	return nonTest, test
}

// addCommandEdges collects parent -> child adjacency from <parent>.AddCommand(<ident>...)
// calls. Non-identifier parents and arguments are skipped (not statically resolvable).
func addCommandEdges(files []parsedFile) map[string][]string {
	edges := map[string][]string{}
	for _, pf := range files {
		ast.Inspect(pf.ast, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "AddCommand" {
				return true
			}
			parent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			for _, arg := range call.Args {
				if id, ok := arg.(*ast.Ident); ok {
					edges[parent.Name] = append(edges[parent.Name], id.Name)
				}
			}
			return true
		})
	}
	return edges
}

// packageLevelVars returns the names of every package-level var declared in files.
// A .Commands() receiver naming one of these is a shared global by construction:
// no type resolution is needed, because only a command exposes Commands().
func packageLevelVars(files []parsedFile) map[string]bool {
	vars := map[string]bool{}
	for _, pf := range files {
		for _, decl := range pf.ast.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					vars[name.Name] = true
				}
			}
		}
	}
	return vars
}

// reachableFrom returns the transitive closure of edges starting at root.
// skipChild, when non-empty, drops that identifier from every adjacency list —
// the falsifiability knob exercised by TestInventoryGuard_DetectsUnreachableGlobal
// and by the AC-CFS-026 round trip.
func reachableFrom(edges map[string][]string, root, skipChild string) map[string]bool {
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range edges[cur] {
			if child == skipChild || seen[child] {
				continue
			}
			seen[child] = true
			queue = append(queue, child)
		}
	}
	return seen
}

// declaresTParallel reports whether body contains a literal t.Parallel() call —
// the same syntactic filter as the acceptance.md §D.5 AWK inventory.
func declaresTParallel(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Parallel" {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "t" {
				found = true
			}
		}
		return true
	})
	return found
}

// commandsReceivers returns the plain-identifier receivers of every .Commands()
// call in body. Non-identifier receivers are omitted (conservative pass, EC-8).
func commandsReceivers(body *ast.BlockStmt) []string {
	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Commands" {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok {
			out = append(out, id.Name)
		}
		return true
	})
	return out
}

// analyzeDir runs the full step 0-3 pipeline for one package directory and
// returns one human-readable violation per offending candidate.
func analyzeDir(t *testing.T, dir, root, skipChild string) []string {
	t.Helper()
	nonTest, test := parseDirFiles(t, dir)
	reachable := reachableFrom(addCommandEdges(nonTest), root, skipChild)
	globals := packageLevelVars(nonTest)

	var violations []string
	for _, pf := range test {
		for _, decl := range pf.ast.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			if !declaresTParallel(fn.Body) {
				continue
			}
			for _, recv := range commandsReceivers(fn.Body) {
				if !globals[recv] || reachable[recv] {
					continue
				}
				violations = append(violations, fmt.Sprintf(
					"%s: %s calls %s.Commands() under t.Parallel(); %s is a package-level "+
						"global NOT reachable from the warm-up root %q, so the cobra "+
						"lazy-sort race is unguarded there",
					pf.name, fn.Name.Name, recv, recv, root))
			}
		}
	}
	sort.Strings(violations)
	return violations
}

// TestInventoryGuard_ParallelCommandsReceiversAreWarmedUp is the REQ-CFS-006 guard.
func TestInventoryGuard_ParallelCommandsReceiversAreWarmedUp(t *testing.T) {
	for dir, root := range warmUpRoots {
		if v := analyzeDir(t, dir, root, ""); len(v) > 0 {
			t.Errorf("dir %s: %d unguarded parallel Commands() receiver(s):\n  %s",
				dir, len(v), strings.Join(v, "\n  "))
		}
	}
}

// TestInventoryGuard_DetectsUnreachableGlobal is the guard's own falsifiability
// proof: with the rootCmd -> githubCmd registration edge removed from the static
// closure, githubCmd becomes an unreachable global and the parallel test that
// uses it MUST be reported. A guard that cannot fail is not a guard.
func TestInventoryGuard_DetectsUnreachableGlobal(t *testing.T) {
	v := analyzeDir(t, ".", "rootCmd", "githubCmd")
	if len(v) == 0 {
		t.Fatal("guard did not report a violation when githubCmd was excluded from the " +
			"warm-up reachable set — the detection path is dead and the guard is vacuous")
	}
	t.Logf("expected violation(s) with githubCmd excluded:\n  %s", strings.Join(v, "\n  "))
}
