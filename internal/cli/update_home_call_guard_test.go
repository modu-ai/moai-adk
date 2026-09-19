// Package cli — update_home_call_guard_test.go
//
// Card t891 (follow-up to t813): a mechanical guard against a NEW direct HOME
// call entering the update flow.
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
// correctly" but "does any site in the update flow resolve HOME outside the
// seam at all". It is a whole-set assertion, so a site added tomorrow is
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
// The test reads source files and writes nothing.

package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

// homeCallGuardExceptions lists update-flow sites permitted to call a
// forbidden function anyway, keyed by file base name, with the reason.
//
// It is empty, and that is a measurement rather than an oversight: at the time
// this guard landed the update flow had zero direct calls (t813 repaired the
// last one). An entry added later must carry a reason a reviewer can weigh —
// "it was already there" is not one.
//
// Two nearby sites are NOT exceptions because they are not in the scanned set:
// migrate_agency.go (a separate CLI entry point, outside `moai update`) and
// homedir.go (the seam's own implementation — userHomeDir delegates to
// paths.Home, which is the single permitted call by construction).
var homeCallGuardExceptions = map[string]string{}

// updateFlowFileSet returns the update-flow production files: `update*.go`
// under internal/cli, excluding tests.
//
// The set is defined by name rather than by reachability, and the trade-off is
// stated rather than hidden. A reachability set would need call-graph analysis
// and would still have to draw a line somewhere; the name glob is one a reader
// can evaluate at a glance and a contributor can predict when adding a file.
// Its cost is named in the Gaps of this card's verdict: a file reached from the
// update flow but named otherwise is not scanned.
func updateFlowFileSet(t *testing.T) []string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(".", "update*.go"))
	if err != nil {
		t.Fatalf("glob update*.go: %v", err)
	}

	var files []string
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		files = append(files, m)
	}
	return files
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

// TestUpdateFlow_NoDirectHomeCalls is the whole-set assertion: no file in the
// update flow resolves HOME outside the userHomeDirFn seam.
func TestUpdateFlow_NoDirectHomeCalls(t *testing.T) {
	files := updateFlowFileSet(t)

	// Premise assertion. A guard that walks an empty set passes for the wrong
	// reason, and a renamed or relocated update flow is exactly how that
	// happens — quietly, with the guard still green. The floor is deliberately
	// far below the current count (24 at the time of writing): it catches a
	// collapsed set without breaking on ordinary file churn.
	if len(files) < 10 {
		t.Fatalf("update-flow file set has %d files (%v); want >= 10 — the glob no longer finds the update flow, so this guard would pass vacuously", len(files), files)
	}

	fset := token.NewFileSet()
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		base := filepath.Base(path)
		if reason, exempt := homeCallGuardExceptions[base]; exempt {
			t.Logf("%s is exempt from the direct-HOME-call guard: %s", base, reason)
			continue
		}

		imports := importPathsByLocalName(file)

		ast.Inspect(file, func(n ast.Node) bool {
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
			// A local variable shadowing the package name resolves to an
			// object; a package qualifier does not. Skipping the former keeps
			// the guard from firing on `paths.Home()` where `paths` is a struct.
			if pkgIdent.Obj != nil {
				return true
			}
			importPath, known := imports[pkgIdent.Name]
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
			t.Errorf("%s:%d calls %s.%s() directly — %s.\nRoute it through userHomeDir() (homedir.go) or the userHomeDirFn seam. If this site is genuinely exempt, add it to homeCallGuardExceptions with the reason.",
				pos.Filename, pos.Line, pkgIdent.Name, sel.Sel.Name, reason)
			return true
		})
	}
}
