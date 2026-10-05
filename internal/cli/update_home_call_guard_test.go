// Package cli — update_home_call_guard_test.go
//
// Card t891 (follow-up to t813), widened by card t956: a mechanical guard
// against a NEW direct HOME call entering the update flow.
//
// What the existing nets do NOT cover. The package-wide TestMain sandbox
// (main_test.go, card t661) redirects HOME by replacing userHomeDirFn, and
// main_test.go says so in its own words: production sites that call
// paths.Home() or os.UserHomeDir() DIRECTLY are outside that net. t813 found
// exactly one such site — runAgencyMigrationAdapter resolved HOME through
// paths.Home(), so a test driving `moai update` against a fixture still
// derived <realHome>/.moai/.migrate-tx-<id>.json — and repaired it. Its
// sibling guard, TestUpdateSubsystem_HomeSeamReach, then pinned the three
// known call sites.
//
// But a seam-reach guard is per-site: it proves the sites it names route
// through the seam, and says nothing about a site that does not exist yet.
// The next direct call lands unnoticed, and the failure is silent in the worst
// way — the test still passes, while the run writes into the developer's real
// ~/.claude or ~/.moai.
//
// This guard is the complementary shape: not "do the known sites route
// correctly" but "does any site the update flow reaches resolve HOME outside
// the seam at all". It is a whole-set assertion, so a site added tomorrow is
// covered without anyone remembering to extend a list.
//
// Why a source scan, and why the AST rather than a grep:
//
//   - A source scan is this repository's established idiom for forbidding a
//     call in a file set (internal/template/agent_askuser_audit_test.go and the
//     *_boundary_test.go family). It rides `go test`, so it runs on the merge
//     gate with no external binary to install — and nothing to skip when that
//     binary is missing.
//   - A grep would be wrong here, not merely coarse: update.go carries a
//     comment that NAMES paths.Home() precisely to warn the next reader off it.
//     A textual guard fires on that comment, so the honest documentation of the
//     rule would be what breaks the rule's own check. The AST sees expressions
//     and never comments.
//
// # How the scanned set is defined (card t956)
//
// t891 defined the set by file NAME — `update*.go` — and wrote the cost of
// that choice into its own Gaps: a file the update flow reaches under another
// name is not scanned. t956 replaces the name with REACHABILITY, and keeps the
// name glob as a floor so no coverage is lost:
//
//	scanned = { functions reachable from the update command's entry points }
//	        ∪ { everything declared in update*.go }
//
// The union is not belt-and-braces. The left term is what this card adds; the
// right term is t891's set preserved verbatim, so the guard cannot regress
// below what it already covered even if reachability analysis loses an edge.
// Measured at the time of writing: reachability alone is already a strict
// superset (69 files vs 24, with no update*.go file falling outside it), so the
// union costs nothing today — it costs something only on the day reachability
// would otherwise have narrowed, which is exactly when it should.
//
// Reachability is computed at FUNCTION granularity, not file granularity, and
// that is load-bearing rather than fussy. migrate_agency.go is in the reachable
// FILE set (some function in it is reached) while runMigrateAgency — the
// function holding its paths.Home() call — is NOT reached: it is a separate
// CLI entry point. A file-level guard fires there, and the red is false. The
// finding is a measurement, not a worry: it is the one place where the two
// granularities disagree today.
//
// The reachability edges deliberately OVER-approximate:
//
//   - a method call x.Foo() adds an edge to any package-level Foo, because
//     resolving the receiver's type needs full type-checking this guard does
//     not pay for;
//   - a bare identifier naming a package-level function adds an edge even
//     where it is not called, because a function installed into a struct field
//     (`RunMigrateAgency: runAgencyMigrationAdapter`) is invoked through that
//     field later. t813's incident site has exactly that shape, so a
//     call-only graph is the wrong instrument for this specific hazard.
//
// Over-approximation is the correct DIRECTION here, and it is the opposite of
// the choice made for shadow detection below. The two answer different
// questions. Shadow detection asks "is this expression a home call?" — a wrong
// yes is a false red on innocent code, so it under-approximates. Reachability
// asks "does the update flow get here?" — a wrong no is a SILENT miss, the
// failure this guard exists to prevent, so it over-approximates and pays for
// it with a red a reviewer can resolve by documenting an exception. Measured
// today: of six direct-home-call sites in the package, the generous graph
// admits exactly one, the seam's own implementation, which is exempt below.
//
// The test reads source files and writes nothing.

package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// forbiddenHomeCalls maps a package-qualified call to the reason it is
// forbidden inside the update flow. The qualifier is the import PATH, not the
// local name, so an aliased import cannot walk past the check.
var forbiddenHomeCalls = map[string]map[string]string{
	"github.com/modu-ai/moai-adk/internal/paths": {
		"Home": "resolves the real HOME directly; the update flow must go through the userHomeDirFn seam so an injected home reaches template rendering and checkpoint paths",
	},
	"os": {
		"UserHomeDir": "reads the OS home directly and ignores HOME on Windows; the update flow must go through the userHomeDirFn seam",
	},
}

// homeCallGuardExceptions lists sites permitted to call a forbidden function
// anyway, keyed either by file base name (exempting the whole file) or by
// "file.go:funcName" (exempting one function), with the reason.
//
// Under t891's name glob the map was empty, and that was a measurement: the
// update flow had zero direct calls once t813 repaired the last one. Widening
// the set to reachability (t956) pulled in exactly one new site that must stay,
// and it is the seam's own implementation — userHomeDir delegating to
// paths.Home is the single permitted call by construction, and the guard is
// meaningless without it.
//
// An entry added later must carry a reason a reviewer can weigh — "it was
// already there" is not one.
var homeCallGuardExceptions = map[string]string{
	"homedir.go:userHomeDir": "the seam's own implementation: userHomeDir() is what every other site is required to route through, so its single delegation to paths.Home() is permitted by construction rather than by exception",
}

// updateCommandName is the cobra `Use:` value whose command literal the entry
// points are read from.
const updateCommandName = "update"

// updateEntryFields are the cobra.Command fields that name an entry function.
var updateEntryFields = map[string]bool{
	"RunE":              true,
	"Run":               true,
	"PreRunE":           true,
	"PreRun":            true,
	"PostRunE":          true,
	"PostRun":           true,
	"PersistentPreRunE": true,
	"PersistentPreRun":  true,
}

// packageFile is one parsed non-test source file of this package.
type packageFile struct {
	base    string
	file    *ast.File
	imports map[string]string // local name -> import path
}

// funcNode is one function declaration, indexed by its bare name.
type funcNode struct {
	base  string
	decl  *ast.FuncDecl
	edges map[string]bool
}

// parsePackageSources parses every non-test .go file in the package directory.
func parsePackageSources(t *testing.T, fset *token.FileSet) []packageFile {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(".", "*.go"))
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}

	var out []packageFile
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		src, err := os.ReadFile(m)
		if err != nil {
			t.Fatalf("read %s: %v", m, err)
		}
		f, err := parser.ParseFile(fset, m, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", m, err)
		}
		out = append(out, packageFile{
			base:    filepath.Base(m),
			file:    f,
			imports: importPathsByLocalName(f),
		})
	}
	return out
}

// importPathsByLocalName maps each import's local name to its import path, so
// a selector `x.Home` can be resolved to the package it actually names.
func importPathsByLocalName(file *ast.File) map[string]string {
	byName := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := path
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			name = path[idx+1:]
		}
		if spec.Name != nil {
			name = spec.Name.Name
		}
		byName[name] = path
	}
	return byName
}

// indexFunctions returns every function declaration by bare name, with its
// outgoing reachability edges. See the file header for why the edge set
// deliberately over-approximates.
func indexFunctions(files []packageFile) map[string][]*funcNode {
	declared := map[string]bool{}
	for _, pf := range files {
		for _, d := range pf.file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				declared[fd.Name.Name] = true
			}
		}
	}

	index := map[string][]*funcNode{}
	for _, pf := range files {
		for _, d := range pf.file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			node := &funcNode{base: pf.base, decl: fd, edges: map[string]bool{}}
			ast.Inspect(fd, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					switch callee := x.Fun.(type) {
					case *ast.Ident:
						node.edges[callee.Name] = true
					case *ast.SelectorExpr:
						node.edges[callee.Sel.Name] = true
					}
				case *ast.Ident:
					// Function-as-value: `RunMigrateAgency: runAgencyMigrationAdapter`.
					if declared[x.Name] {
						node.edges[x.Name] = true
					}
				}
				return true
			})
			index[fd.Name.Name] = append(index[fd.Name.Name], node)
		}
	}
	return index
}

// updateEntryPoints reads the entry function names off the cobra command
// literal whose Use: is "update".
//
// Reading them from the declaration rather than hardcoding a name is what
// keeps the premise honest: a rewired or renamed command yields zero entries
// and the caller fails loudly, instead of the guard quietly walking a graph
// rooted at a function that no longer starts anything.
func updateEntryPoints(files []packageFile) []string {
	found := map[string]bool{}
	for _, pf := range files {
		ast.Inspect(pf.file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Command" {
				return true
			}

			isUpdate := false
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "Use" {
					continue
				}
				basic, ok := kv.Value.(*ast.BasicLit)
				if !ok || basic.Kind != token.STRING {
					continue
				}
				if v, err := strconv.Unquote(basic.Value); err == nil && v == updateCommandName {
					isUpdate = true
				}
			}
			if !isUpdate {
				return true
			}

			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || !updateEntryFields[key.Name] {
					continue
				}
				if fn, ok := kv.Value.(*ast.Ident); ok {
					found[fn.Name] = true
				}
			}
			return true
		})
	}

	entries := make([]string, 0, len(found))
	for name := range found {
		entries = append(entries, name)
	}
	sort.Strings(entries)
	return entries
}

// reachableFunctions returns the transitive closure of function names reached
// from the given entry points.
func reachableFunctions(index map[string][]*funcNode, entries []string) map[string]bool {
	seen := map[string]bool{}
	queue := append([]string{}, entries...)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		for _, node := range index[cur] {
			for edge := range node.edges {
				if !seen[edge] {
					queue = append(queue, edge)
				}
			}
		}
	}
	return seen
}

// isUpdateGlobFile reports whether the file is in t891's original name-glob
// set, which this guard keeps as a floor under the reachable set.
func isUpdateGlobFile(base string) bool {
	return strings.HasPrefix(base, "update") && strings.HasSuffix(base, ".go")
}

// reportForbiddenCalls walks one node and reports every direct HOME call.
func reportForbiddenCalls(t *testing.T, fset *token.FileSet, pf packageFile, root ast.Node, scopeNote string) {
	t.Helper()

	ast.Inspect(root, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		// A local variable shadowing the package name resolves to an object; a
		// package qualifier does not. Skipping the former keeps the guard from
		// firing on `paths.Home()` where `paths` is a struct.
		if pkgIdent.Obj != nil {
			return true
		}
		importPath, known := pf.imports[pkgIdent.Name]
		if !known {
			return true
		}
		forbidden, listed := forbiddenHomeCalls[importPath]
		if !listed {
			return true
		}
		reason, bad := forbidden[sel.Sel.Name]
		if !bad {
			return true
		}

		pos := fset.Position(call.Pos())
		t.Errorf("%s:%d calls %s.%s() directly — %s.\n  in scope because: %s\n  Route it through userHomeDir() (homedir.go) or the userHomeDirFn seam. If this site is genuinely exempt, add it to homeCallGuardExceptions with the reason.",
			pos.Filename, pos.Line, pkgIdent.Name, sel.Sel.Name, reason, scopeNote)
		return true
	})
}

// TestUpdateFlow_NoDirectHomeCalls is the whole-set assertion: nothing the
// update flow reaches resolves HOME outside the userHomeDirFn seam.
func TestUpdateFlow_NoDirectHomeCalls(t *testing.T) {
	fset := token.NewFileSet()
	files := parsePackageSources(t, fset)

	// Premise assertion 1 — the package was actually read. A guard that walks
	// an empty set passes for the wrong reason.
	if len(files) < 50 {
		t.Fatalf("parsed %d non-test files in this package; want >= 50 — the source scan no longer sees the package, so this guard would pass vacuously", len(files))
	}

	// Premise assertion 2 — the entry points were discovered, not assumed. A
	// renamed or rewired update command yields none, and the reachable set
	// would then be empty with no other symptom.
	entries := updateEntryPoints(files)
	if len(entries) == 0 {
		t.Fatalf("no entry function found on the cobra command with Use: %q — the command was renamed or rewired, so reachability would be rooted at nothing and this guard would pass vacuously", updateCommandName)
	}

	index := indexFunctions(files)
	reachable := reachableFunctions(index, entries)

	// Premise assertion 3 — the graph actually expanded. A single unexpanded
	// entry point means the edges stopped resolving.
	if len(reachable) < len(entries)*2 {
		t.Fatalf("reachability from %v expanded to only %d functions — the call graph is not resolving, so this guard would scan almost nothing", entries, len(reachable))
	}

	scannedFuncs := 0
	globFiles := 0

	for _, pf := range files {
		if reason, exempt := homeCallGuardExceptions[pf.base]; exempt {
			t.Logf("%s is exempt from the direct-HOME-call guard: %s", pf.base, reason)
			continue
		}

		// Scope term 1 — t891's name glob, preserved verbatim: every node in
		// an update*.go file, package-level declarations included.
		if isUpdateGlobFile(pf.base) {
			globFiles++
			reportForbiddenCalls(t, fset, pf, pf.file, "the file is in the update*.go name set (card t891)")
			continue
		}

		// Scope term 2 — card t956: functions this file declares that the
		// update flow reaches. A file-level test here would be wrong; see the
		// file header on migrate_agency.go.
		for _, d := range pf.file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if !reachable[fd.Name.Name] {
				continue
			}
			key := pf.base + ":" + fd.Name.Name
			if reason, exempt := homeCallGuardExceptions[key]; exempt {
				t.Logf("%s is exempt from the direct-HOME-call guard: %s", key, reason)
				continue
			}
			scannedFuncs++
			reportForbiddenCalls(t, fset, pf, fd, "the update flow reaches "+fd.Name.Name+"() (card t956)")
		}
	}

	// Premise assertion 4 — both scope terms contributed. Either one silently
	// collapsing to nothing is the vacuous-green shape this guard must not have.
	if globFiles < 10 {
		t.Fatalf("update*.go name set has %d files; want >= 10 — t891's floor no longer finds the update flow", globFiles)
	}
	if scannedFuncs < 20 {
		t.Fatalf("reachability contributed only %d scanned functions outside update*.go; want >= 20 — the widening this guard exists for is not happening", scannedFuncs)
	}

	t.Logf("scanned: %d update*.go files (t891 floor) + %d reachable functions elsewhere (t956), from entry points %v", globFiles, scannedFuncs, entries)
}
