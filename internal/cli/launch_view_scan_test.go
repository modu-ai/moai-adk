package cli

// launch_view_scan_test.go — SPEC-LAUNCHER-ARGV-001 M3 (AC-LA-013, REQ-LA-001):
// the single-parse guarantee, enforced by reading the source.
//
// The six audit rounds of card t1604 each found one more spelling on which two
// hand-written readers disagreed. The structural repair is that worktree
// selection is derived in ONE file (launch_view.go); the readers on the launch
// path receive values. Not creating the parallel readers is the deletion of the
// parallel readers, and this scan is what keeps it so: a reader reintroduced in
// any file of the launch-path closure fails here.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

// launchPathClosureFiles is the launch-path call closure: the cc / glm entries,
// launch -> runLaunchClaude, and every file those reach that handles a
// worktree selection. lane_resume.go is inside it because a lane case falls
// through to the shared tail and calls laneJoinChildArgv (cc.go) — the lane
// join assembler may INSERT argv tokens, but it must not READ a worktree value.
// factory_anchor_guard.go is the guard: it judges the view and owns no reader.
var launchPathClosureFiles = []string{
	"cc.go", "glm.go", "launcher.go", "codex_hooks_seed.go", "session_worktree.go",
	"worktree_branch_flag.go", "launch_chain.go", "lane_resume.go", "factory_anchor_guard.go",
}

// resolverOwnedNames are the walk helpers that derive a worktree value from raw
// arguments. They are defined in launch_view.go and referenced nowhere else.
var resolverOwnedNames = []string{
	"worktreeFlagValue", "normalizeWorktreeFlag", "splitWorktreeBranchFlag",
	"extractWorktreeLaunchTarget", "selectBackendWorktree", "readFirstWorktree",
}

// worktreeSpellingLiterals are the string literals a hand-written worktree
// reader matches tokens against.
var worktreeSpellingLiterals = []string{"--worktree", "--worktree=", "-w", "-w="}

// parallelReaderNames are the reader family of the preserved tree of card
// t1604 that this redesign deliberately does not recreate.
var parallelReaderNames = []string{
	"worktreeScanAdvance", "worktreeNextValue", "launcherConsumedSpan", "worktreeFlagValues",
	"worktreeFlagValuesClaudeSegment", "guardArgSegments", "equalsWorktreeValue",
	"isAttachedWorktreeFlag", "claudeArgView",
}

func parseGoFile(t *testing.T, name string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0) // comments dropped: only code counts
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return file
}

// matchedLiterals returns the string literals a file compares tokens against:
// operands of == / !=, `case` expressions, and arguments of the strings prefix
// functions. A literal that merely appears as a table key (an allowlist of flag
// NAMES, as the glm usage redaction keeps) matches nothing and is not a reader.
func matchedLiterals(file *ast.File) []string {
	var out []string
	add := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if v, err := strconv.Unquote(lit.Value); err == nil {
				out = append(out, v)
			}
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if x.Op == token.EQL || x.Op == token.NEQ {
				add(x.X)
				add(x.Y)
			}
		case *ast.CaseClause:
			for _, e := range x.List {
				add(e)
			}
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "strings" {
					switch sel.Sel.Name {
					case "HasPrefix", "TrimPrefix", "CutPrefix":
						for _, a := range x.Args {
							add(a)
						}
					}
				}
			}
		}
		return true
	})
	return out
}

func TestLaunchViewSingleParse(t *testing.T) {
	// AC-LA-013: outside launch_view.go, no file of the launch-path closure
	// references a resolver-owned walk helper or matches a token against a
	// worktree spelling literal. The control half of the scan proves it can see
	// what it claims to look for: launch_view.go DEFINES every owned name and
	// carries every literal, so an empty result is not an empty sweep.
	t.Run("no_rederivation_on_launch_path", func(t *testing.T) {
		owned := map[string]bool{}
		for _, n := range resolverOwnedNames {
			owned[n] = true
		}
		literals := map[string]bool{}
		for _, l := range worktreeSpellingLiterals {
			literals[l] = true
		}

		// Control: the resolver file defines the helpers and carries the literals.
		view := parseGoFile(t, "launch_view.go")
		defined := map[string]bool{}
		sawLiteral := map[string]bool{}
		ast.Inspect(view, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				defined[fn.Name.Name] = true
			}
			return true
		})
		for _, v := range matchedLiterals(view) {
			if literals[v] {
				sawLiteral[v] = true
			}
		}
		for _, n := range resolverOwnedNames {
			if !defined[n] {
				t.Errorf("control: launch_view.go does not define %s, so the scan below would sweep for a name nothing owns", n)
			}
		}
		for _, l := range worktreeSpellingLiterals {
			if !sawLiteral[l] {
				t.Errorf("control: launch_view.go matches no %q literal, so the literal sweep below could not fire on a reader", l)
			}
		}

		scanned := 0
		var findings []string
		for _, name := range launchPathClosureFiles {
			file := parseGoFile(t, name)
			scanned++
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					if owned[x.Name] {
						findings = append(findings, name+": reference to "+x.Name)
					}
				}
				return true
			})
			for _, v := range matchedLiterals(file) {
				if literals[v] {
					findings = append(findings, name+": matches tokens against the worktree spelling literal "+strconv.Quote(v))
				}
			}
		}
		if scanned != len(launchPathClosureFiles) {
			t.Fatalf("scanned %d files, want %d: the sweep set is not the closure", scanned, len(launchPathClosureFiles))
		}
		sort.Strings(findings)
		if len(findings) != 0 {
			t.Errorf("worktree selection is derived outside launch_view.go (%d findings):\n  %v", len(findings), findings)
		}
	})

	// The parallel reader family of the preserved tree is not recreated, and the
	// guard does not borrow the resume interpreter's segment classifier to build
	// one.
	t.Run("parallel_readers_absent", func(t *testing.T) {
		files, err := filepath.Glob("*.go")
		if err != nil || len(files) == 0 {
			t.Fatalf("glob the package: %v (%d files)", err, len(files))
		}
		banned := map[string]bool{}
		for _, n := range parallelReaderNames {
			banned[n] = true
		}
		checked := 0
		for _, f := range files {
			if filepath.Ext(f) != ".go" || len(f) > len("_test.go") && f[len(f)-len("_test.go"):] == "_test.go" {
				continue
			}
			checked++
			ast.Inspect(parseGoFile(t, f), func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok && banned[fn.Name.Name] {
					t.Errorf("%s defines %s: the parallel reader family of the preserved tree must not be recreated", f, fn.Name.Name)
				}
				return true
			})
		}
		if checked < 50 {
			t.Fatalf("checked %d package files, want the whole package", checked)
		}
		ast.Inspect(parseGoFile(t, "factory_anchor_guard.go"), func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "segmentClass" {
				t.Errorf("factory_anchor_guard.go references segmentClass: the guard judges the view, it does not classify argv segments")
			}
			return true
		})
	})

	// The launcher appends --name / --settings / --effort at the END of a leader
	// argv after the guard has judged the view. The judgment is valid for the exec
	// argv only if those tokens cannot change what the readers and the backend
	// select; this is the property, checked on every §D.4 spelling.
	t.Run("appended_flags_are_worktree_neutral", func(t *testing.T) {
		pinSnapshotModel(t)
		suffixes := [][]string{
			{"--name", "leader"},
			{"--name", "leader", "--settings", "/tmp/moai-factory-1-2.json"},
			{"--name", "leader", "--settings", "/tmp/moai-factory-1-2.json", "--effort", "max"},
		}
		for _, row := range viewEqualityRows(filepath.Join(t.TempDir(), "elsewhere")) {
			base := resolveLaunchView(row.args)
			for _, suffix := range suffixes {
				appended := resolveLaunchView(append(append([]string(nil), row.args...), suffix...))
				if appended.Selection != base.Selection {
					t.Errorf("%s + %v: selection %+v, want %+v", row.name, suffix, appended.Selection, base.Selection)
				}
				if appended.StageA != base.StageA || appended.StageB != base.StageB || appended.Consumed.ChainTarget != base.Consumed.ChainTarget {
					t.Errorf("%s + %v: a stage read changed", row.name, suffix)
				}
				if got := stripLauncherAppended(appended.FinalArgv); !reflect.DeepEqual(got, base.FinalArgv) {
					t.Errorf("%s + %v: FinalArgv %q, want the unappended %q once the tail flags are stripped", row.name, suffix, got, base.FinalArgv)
				}
			}
		}
	})
}
