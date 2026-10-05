package cli

// SPEC-HOOK-STDIN-FAILCLOSED-001 AC-HSF-003(b) — structural checks over the
// non-test Go files of this package:
//
//	(b1) each entry point's ReadInput error block decides fail-closed through
//	     codexadapter.IsDecisionBearing, and uses the result in a branch condition;
//	(b2) no second list of decision-bearing events exists;
//	(b3) the parse-failure path reads no environment variable or file outside
//	     the project-root allow-list.
//
// Line-oriented grep misses a switch whose cases are split over several lines,
// which is why these are AST checks.

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

// decisionEventIdents are the selector names of the four decision-bearing
// events as they appear in source. They are the vocabulary a second list
// would be written in; the test does not use them as a classification.
var decisionEventIdents = map[string]bool{
	"EventPreToolUse":        true,
	"EventPermissionRequest": true,
	"EventStop":              true,
	"EventUserPromptSubmit":  true,
}

type pkgSource struct {
	fset  *token.FileSet
	files []*ast.File
	funcs map[string]*ast.FuncDecl // top-level functions only (no methods)
}

func loadPackageSource(t *testing.T) pkgSource {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	src := pkgSource{fset: fset, funcs: map[string]*ast.FuncDecl{}}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		src.files = append(src.files, f)
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				src.funcs[fd.Name.Name] = fd
			}
		}
	}
	if len(src.files) == 0 || src.funcs["runHookEvent"] == nil || src.funcs["runAgentHook"] == nil {
		t.Fatalf("package scan found %d files and no entry points — the scan is broken", len(src.files))
	}
	return src
}

// isSel reports whether e is pkg.name.
func isSel(e ast.Expr, pkg, name string) bool {
	s, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && id.Name == pkg && (name == "" || s.Sel.Name == name)
}

// decisionIdentIn reports the decision-event selector names under n.
func decisionIdentsIn(n ast.Node) []string {
	var out []string
	ast.Inspect(n, func(x ast.Node) bool {
		if s, ok := x.(*ast.SelectorExpr); ok && isSel(s, "hook", "") && decisionEventIdents[s.Sel.Name] {
			out = append(out, s.Sel.Name)
		}
		return true
	})
	return out
}

func distinct(names []string) int {
	set := map[string]bool{}
	for _, n := range names {
		set[n] = true
	}
	return len(set)
}

// readInputErrorBlock returns the body of `if err != nil` that immediately
// follows the `... := deps.HookProtocol.ReadInput(...)` assignment in fn.
func readInputErrorBlock(fn *ast.FuncDecl) *ast.BlockStmt {
	var found *ast.BlockStmt
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		blk, ok := n.(*ast.BlockStmt)
		if !ok || found != nil {
			return found == nil
		}
		for i, st := range blk.List {
			as, ok := st.(*ast.AssignStmt)
			if !ok || len(as.Rhs) != 1 {
				continue
			}
			call, ok := as.Rhs[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "ReadInput" || i+1 >= len(blk.List) {
				continue
			}
			if ifs, ok := blk.List[i+1].(*ast.IfStmt); ok {
				found = ifs.Body
				return false
			}
		}
		return true
	})
	return found
}

// directCallees returns the same-package top-level functions called by
// identifier inside n.
func directCallees(src pkgSource, n ast.Node) []*ast.FuncDecl {
	var out []*ast.FuncDecl
	seen := map[string]bool{}
	ast.Inspect(n, func(x ast.Node) bool {
		call, ok := x.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := call.Fun.(*ast.Ident); ok {
			if fd := src.funcs[id.Name]; fd != nil && !seen[id.Name] {
				seen[id.Name] = true
				out = append(out, fd)
			}
		}
		return true
	})
	return out
}

// isDecisionBearingCall reports whether call is codexadapter.IsDecisionBearing(x)
// with x not a hook.EventX selector literal.
func isDecisionBearingCall(call *ast.CallExpr) bool {
	if !isSel(call.Fun, "codexadapter", "IsDecisionBearing") || len(call.Args) != 1 {
		return false
	}
	return !isSel(call.Args[0], "hook", "")
}

// decisionCallUsedInBranch reports whether scope holds an IsDecisionBearing
// call (with a non-literal argument) whose result reaches an if condition or a
// switch tag/case — directly, or through a local variable.
func decisionCallUsedInBranch(scope ast.Node) bool {
	// Local variables assigned an IsDecisionBearing result.
	vars := map[string]bool{}
	ast.Inspect(scope, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, rhs := range as.Rhs {
			if call, ok := rhs.(*ast.CallExpr); ok && isDecisionBearingCall(call) && i < len(as.Lhs) {
				if id, ok := as.Lhs[i].(*ast.Ident); ok && id.Name != "_" {
					vars[id.Name] = true
				}
			}
		}
		return true
	})
	condUses := func(e ast.Node) bool {
		used := false
		ast.Inspect(e, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CallExpr:
				if isDecisionBearingCall(v) {
					used = true
				}
			case *ast.Ident:
				if vars[v.Name] {
					used = true
				}
			}
			return !used
		})
		return used
	}
	used := false
	ast.Inspect(scope, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.IfStmt:
			if condUses(v.Cond) {
				used = true
			}
		case *ast.SwitchStmt:
			if v.Tag != nil && condUses(v.Tag) {
				used = true
			}
			for _, c := range v.Body.List {
				for _, e := range c.(*ast.CaseClause).List {
					if condUses(e) {
						used = true
					}
				}
			}
		}
		return !used
	})
	return used
}

// decisionComparisons reports == / != comparisons and case clauses against a
// decision-event identifier under n.
func decisionComparisons(n ast.Node) []string {
	var out []string
	ast.Inspect(n, func(x ast.Node) bool {
		switch v := x.(type) {
		case *ast.BinaryExpr:
			if v.Op == token.EQL || v.Op == token.NEQ {
				for _, side := range []ast.Expr{v.X, v.Y} {
					if s, ok := side.(*ast.SelectorExpr); ok && isSel(s, "hook", "") && decisionEventIdents[s.Sel.Name] {
						out = append(out, v.Op.String()+" "+s.Sel.Name)
					}
				}
			}
		case *ast.CaseClause:
			for _, e := range v.List {
				for _, id := range decisionIdentsIn(e) {
					out = append(out, "case "+id)
				}
			}
		}
		return true
	})
	return out
}

// checkB1 returns the (b1) violations for one entry point.
func checkB1(src pkgSource, entry string) []string {
	fn := src.funcs[entry]
	blk := readInputErrorBlock(fn)
	if blk == nil {
		return []string{entry + ": no ReadInput error block found"}
	}
	scopes := append([]ast.Node{blk}, nodes(directCallees(src, blk))...)
	var problems []string
	used := false
	for _, s := range scopes {
		if decisionCallUsedInBranch(s) {
			used = true
		}
		for _, c := range decisionComparisons(s) {
			problems = append(problems, entry+": decision-event comparison in the parse-failure scope: "+c)
		}
	}
	if !used {
		problems = append(problems, entry+": no codexadapter.IsDecisionBearing result selects a branch in the parse-failure scope")
	}
	return problems
}

func nodes(fds []*ast.FuncDecl) []ast.Node {
	out := make([]ast.Node, 0, len(fds))
	for _, fd := range fds {
		out = append(out, fd.Body)
	}
	return out
}

// checkB2 returns every construct listing two or more decision-event
// identifiers: EventType slice/array/map literals, a switch's case lists, and
// ||-chained equality comparisons.
func checkB2(src pkgSource) []string {
	var problems []string
	report := func(n ast.Node, what string, ids []string) {
		if distinct(ids) >= 2 {
			sort.Strings(ids)
			problems = append(problems, src.fset.Position(n.Pos()).String()+": "+what+" lists "+strings.Join(ids, ","))
		}
	}
	isEventType := func(e ast.Expr) bool { return isSel(e, "hook", "EventType") }
	for _, f := range src.files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CompositeLit:
				switch tt := v.Type.(type) {
				case *ast.ArrayType:
					if isEventType(tt.Elt) {
						report(v, "EventType list literal", decisionIdentsIn(v))
					}
				case *ast.MapType:
					if isEventType(tt.Key) {
						var ids []string
						for _, el := range v.Elts {
							if kv, ok := el.(*ast.KeyValueExpr); ok {
								ids = append(ids, decisionIdentsIn(kv.Key)...)
							}
						}
						report(v, "EventType-keyed map literal", ids)
					}
				}
			case *ast.SwitchStmt:
				var ids []string
				for _, c := range v.Body.List {
					for _, e := range c.(*ast.CaseClause).List {
						ids = append(ids, decisionIdentsIn(e)...)
					}
				}
				report(v, "switch case lists", ids)
			case *ast.BinaryExpr:
				if v.Op == token.LOR {
					var ids []string
					var walk func(e ast.Expr)
					walk = func(e ast.Expr) {
						switch b := e.(type) {
						case *ast.BinaryExpr:
							if b.Op == token.LOR {
								walk(b.X)
								walk(b.Y)
								return
							}
							if b.Op == token.EQL {
								ids = append(ids, decisionIdentsIn(b)...)
							}
						case *ast.ParenExpr:
							walk(b.X)
						}
					}
					walk(v)
					report(v, "|| comparison chain", ids)
					return false
				}
			}
			return true
		})
	}
	return problems
}

// forbiddenReads are the environment and file primitives the parse-failure
// path must not call (acceptance.md AC-HSF-003(b3)).
var forbiddenReads = [][2]string{
	{"os", "Getenv"}, {"os", "LookupEnv"}, {"os", "Environ"}, {"syscall", "Getenv"},
	{"os", "ReadFile"}, {"os", "Open"}, {"os", "ReadDir"},
}

// b3AllowedRoot is the one same-package call the scope may make without being
// followed: the sink's project-root resolution.
const b3AllowedRoot = "resolveHookProjectRoot"

// b3StopParseCapFile is the one file whose functions may read inside the
// parse-failure path, and only what SPEC-HOOK-STOP-PARSE-CAP-001 requires:
// the two environment variables that choose the counting key (REQ-SPC-014 —
// they pick WHICH record, never the deny/release threshold, REQ-SPC-004) and
// the state area's own records (REQ-SPC-001/007/008). Any other read there,
// and every read anywhere else, is still a violation. The threshold side of
// REQ-SPC-004 is pinned separately by TestStopParseCap_StaticChecks.
const b3StopParseCapFile = "hook_stop_parse_cap.go"

var (
	b3StopParseCapEnv   = map[string]bool{"EnvClaudeCodeSessionID": true, "EnvMoaiSessionPID": true}
	b3StopParseCapFiles = map[string]bool{"ReadFile": true, "ReadDir": true}
)

// b3StopParseCapAllows reports whether call, found in file, is one of the
// reads b3StopParseCapFile is allowed.
func b3StopParseCapAllows(file string, call *ast.CallExpr, fr [2]string) bool {
	if filepath.Base(file) != b3StopParseCapFile || fr[0] != "os" {
		return false
	}
	if fr[1] == "Getenv" {
		if len(call.Args) != 1 {
			return false
		}
		sel, ok := call.Args[0].(*ast.SelectorExpr)
		return ok && isSel(sel, "config", "") && b3StopParseCapEnv[sel.Sel.Name]
	}
	return b3StopParseCapFiles[fr[1]]
}

// checkB3 walks both entry points' parse-failure blocks and the transitive
// closure of same-package function calls from them.
func checkB3(src pkgSource) []string {
	var problems []string
	visited := map[string]bool{}
	var queue []ast.Node
	for _, entry := range []string{"runHookEvent", "runAgentHook"} {
		blk := readInputErrorBlock(src.funcs[entry])
		if blk == nil {
			return []string{entry + ": no ReadInput error block found"}
		}
		queue = append(queue, blk)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		ast.Inspect(n, func(x ast.Node) bool {
			call, ok := x.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); ok {
				if id.Name == b3AllowedRoot {
					return true
				}
				if fd := src.funcs[id.Name]; fd != nil && !visited[id.Name] {
					visited[id.Name] = true
					queue = append(queue, fd.Body)
				}
				return true
			}
			for _, fr := range forbiddenReads {
				if !isSel(call.Fun, fr[0], fr[1]) {
					continue
				}
				if fr[0] == "os" && fr[1] == "Getenv" && len(call.Args) == 1 && isSel(call.Args[0], "config", "EnvClaudeProjectDir") {
					continue
				}
				if b3StopParseCapAllows(src.fset.Position(call.Pos()).Filename, call, fr) {
					continue
				}
				problems = append(problems, src.fset.Position(call.Pos()).String()+": "+fr[0]+"."+fr[1]+" in the parse-failure path")
			}
			return true
		})
	}
	return problems
}

// TestStdinFailClosed_SingleDecisionListAST is AC-HSF-003(b1)(b2)(b3).
func TestStdinFailClosed_SingleDecisionListAST(t *testing.T) {
	src := loadPackageSource(t)
	for _, entry := range []string{"runHookEvent", "runAgentHook"} {
		for _, p := range checkB1(src, entry) {
			t.Errorf("(b1) %s", p)
		}
	}
	for _, p := range checkB2(src) {
		t.Errorf("(b2) %s", p)
	}
	for _, p := range checkB3(src) {
		t.Errorf("(b3) %s", p)
	}
}

// TestStdinFailClosed_ASTCheckerDetectsItsTargets proves each check can fail:
// it runs the same checkers over fixture sources that carry the violation.
func TestStdinFailClosed_ASTCheckerDetectsItsTargets(t *testing.T) {
	fixtureNamed := func(t *testing.T, name, body string) pkgSource {
		t.Helper()
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, "package cli\n"+body, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse fixture: %v", err)
		}
		src := pkgSource{fset: fset, files: []*ast.File{f}, funcs: map[string]*ast.FuncDecl{}}
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				src.funcs[fd.Name.Name] = fd
			}
		}
		return src
	}
	fixture := func(t *testing.T, body string) pkgSource {
		t.Helper()
		return fixtureNamed(t, "fixture.go", body)
	}
	entries := `
func runHookEvent() error {
	input, err := deps.HookProtocol.ReadInput(stdin)
	if err != nil { return handle(event, err) }
	_ = input
	return nil
}
func runAgentHook() error {
	input, err := deps.HookProtocol.ReadInput(stdin)
	if err != nil { return handle(event, err) }
	_ = input
	return nil
}
`
	good := fixture(t, entries+`
func handle(event hook.EventType, err error) error {
	if !codexadapter.IsDecisionBearing(event) { return nil }
	root := resolveHookProjectRoot()
	_ = os.Getenv(config.EnvClaudeProjectDir)
	_ = root
	return nil
}
`)
	if p := append(checkB1(good, "runHookEvent"), checkB3(good)...); len(p) != 0 {
		t.Fatalf("clean fixture reported violations: %v", p)
	}

	discarded := fixture(t, entries+`
func handle(event hook.EventType, err error) error {
	_ = codexadapter.IsDecisionBearing(event)
	if event != hook.EventPostToolUse && event != hook.EventSubagentStop { return nil }
	return nil
}
`)
	if len(checkB1(discarded, "runAgentHook")) == 0 {
		t.Fatal("(b1) missed an IsDecisionBearing result that selects no branch")
	}

	compared := fixture(t, entries+`
func handle(event hook.EventType, err error) error {
	if codexadapter.IsDecisionBearing(event) || event == hook.EventPreToolUse { return nil }
	return nil
}
`)
	if len(checkB1(compared, "runHookEvent")) == 0 {
		t.Fatal("(b1) missed a decision-event comparison in the parse-failure scope")
	}

	dup := fixture(t, `
func f(event hook.EventType) {
	switch event {
	case hook.EventPreToolUse:
	case hook.EventStop:
	}
}
`)
	if len(checkB2(dup)) == 0 {
		t.Fatal("(b2) missed a multi-line switch listing two decision events")
	}

	env := fixture(t, entries+`
func handle(event hook.EventType, err error) error {
	if codexadapter.IsDecisionBearing(event) { return helper() }
	return nil
}
func helper() error {
	if os.Getenv("SOME_SWITCH") != "" { return nil }
	return nil
}
`)
	if len(checkB3(env)) == 0 {
		t.Fatal("(b3) missed an environment read in a transitive callee")
	}

	// The SPEC-HOOK-STOP-PARSE-CAP-001 carve-out admits exactly the key
	// selection and the state-area reads, and only in its own file.
	capBody := func(reads string) string {
		return entries + `
func handle(event hook.EventType, err error) error {
	if codexadapter.IsDecisionBearing(event) { return helper() }
	return nil
}
func helper() error {
` + reads + `
	return nil
}
`
	}
	allowed := `	_ = os.Getenv(config.EnvClaudeCodeSessionID)
	_ = os.Getenv(config.EnvMoaiSessionPID)
	_, _ = os.ReadFile("x")
	_, _ = os.ReadDir("x")`
	if p := checkB3(fixtureNamed(t, "hook_stop_parse_cap.go", capBody(allowed))); len(p) != 0 {
		t.Fatalf("(b3) carve-out rejected its own reads: %v", p)
	}
	if len(checkB3(fixture(t, capBody(allowed)))) != 4 {
		t.Fatal("(b3) the carve-out leaked outside hook_stop_parse_cap.go")
	}
	for _, read := range []string{
		`	_ = os.Getenv(config.EnvClaudeCodeStopHookBlockCap)`,
		`	_ = os.Getenv("CLAUDE_CODE_SESSION_ID")`,
		`	_, _ = os.Open("x")`,
		`	_, _ = os.LookupEnv(config.EnvClaudeCodeSessionID)`,
	} {
		if len(checkB3(fixtureNamed(t, "hook_stop_parse_cap.go", capBody(read)))) == 0 {
			t.Fatalf("(b3) carve-out admitted %q", read)
		}
	}
}
