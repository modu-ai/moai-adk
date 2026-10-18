package testguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"path"
	"sort"
	"strconv"
	"strings"
)

// The indirect check finds tests that reach `git config --global` through
// production code, which the two direct patterns in ScanSource cannot see. It is
// a static, stdlib-only, name-based call graph over one package directory:
//
//   - A node is a package-level function, a package-level variable, or a
//     variable of another package that this package assigns (wtcmd.X = f).
//     Methods are not nodes.
//   - An edge is any reference by name from a node's body or initializer to
//     another node: a call, a function value in a composite literal, or an
//     import-qualified variable. A reference to a local value that shadows a
//     node name is not an edge.
//   - A sink is a function or variable initializer that holds a call to
//     exec.Command or exec.CommandContext whose program is the string literal
//     "git" (or a path whose base is git) and whose literal arguments include
//     "config" and "--global".
//   - A seam is a variable node. A test isolates a seam only by an executed
//     replacement: a call statement, on the path from the test function, to a
//     test-file function that assigns the seam a replacement value, or to a
//     helper that stores a parameter, or a field of a parameter struct, in the
//     seam while the call passes a replacement at that argument position. A
//     replacement is a func literal that does not use the seam's own value
//     (directly or through a local derived from it), a function, or a call
//     result that does not take the seam's value. RedirectHome, or a Setenv
//     whose first argument is the literal "GIT_CONFIG_GLOBAL", isolates every
//     seam.
//   - Executed means run before the sink: statements outside defer, go, a
//     Cleanup argument, a closure that is not invoked on the spot, passed to
//     Run, or called through a local variable at a call that is itself executed,
//     an `if` or `for` whose condition is the literal false, a range over an
//     empty composite literal or the empty string, and a `case false` of a
//     tagless switch. A restore is not a replacement: the seam itself, a local
//     whose every assignment derives from the seam (through any chain of locals
//     or a call that takes it), a func literal that uses the seam's value, or a
//     field read of a local.
//
// A top-level Test, Benchmark or Fuzz function that reaches a sink along a path
// with no isolated seam is reported once, on its declaration line, unless the
// line or the line above carries `home-guard:allow <reason>`.
//
// Limits. The check is conservative about what it follows (every name match is
// an edge) and blind to the following; a test that depends on one is reported
// or passed as stated, and `home-guard:allow <reason>` is the escape hatch for
// a false report:
//   - Method calls, and calls through struct fields, interfaces and
//     function-typed parameters, are neither followed nor credited.
//   - A helper reached by a qualified selector (pkg.Helper) or a method is
//     neither followed nor credited; only a variable of another package that
//     this package assigns becomes a node.
//   - Edges out of a node that references more than dispatchFanOut (40) nodes
//     are not followed, so a sink behind a larger dispatch table is not
//     reported, and a helper reached only through such a node is not credited.
//   - The program must be a string literal: a const or var holding "git" (or a
//     call result) is not a sink. No production sink is affected, because all
//     four use the literal "git".
//   - "config" and "--global" must be string-literal arguments; a flag built by
//     concatenation, a const, or a slice spread is not seen.
//   - Writes to the global git config by other routes (writing .gitconfig
//     directly, an environment other than GIT_CONFIG_GLOBAL) are not covered.
//   - Isolation is credited from calls on the executed path only; a helper
//     stored in a package-level variable or a struct field and called through
//     it, a helper that forwards its parameter to a second helper, and a field
//     assigned on a local struct (a.f = stub) are not credited, so the test is
//     reported.
//   - A func literal that calls the real function by name (not through the
//     seam's value) is credited as a replacement although it reaches the sink.
//   - RedirectHome is credited for any executed call to a function or qualified
//     selector with that name, and Setenv for any call whose first argument is
//     the literal "GIT_CONFIG_GLOBAL"; the arguments of RedirectHome and the
//     value given to GIT_CONFIG_GLOBAL are not inspected.
//   - Order inside a function is not modeled: a swap or RedirectHome written
//     after the call that reaches the sink still isolates, and an early return,
//     panic or Fatal before the isolating call is not seen.
//   - Only the constant-false forms named above are read as not executed; any
//     other condition (1 > 2, a named constant, a loop that cannot iterate for a
//     reason other than a false condition or an empty literal, a `case false`
//     of a switch with a tag) is treated as possibly true, so isolation under
//     it is credited.
//   - A local is judged by every assignment to it in its function, whatever the
//     order; a local with any assignment that is not a replacement is not one.
//   - All .go files of a directory form one package: build tags are ignored and
//     same-named identifiers of package x and x_test merge into one node.
//   - Only the directory's own code is analysed; a sink in another package is
//     seen only through a variable this package assigns.

// dispatchFanOut is the number of nodes above which a node is read as a dispatch
// table: a command tree or a server that registers many handlers. One test runs
// one branch of such a table and the name graph cannot tell which, so following
// every branch reports nearly every test of a command package (measured on
// internal/cli: 648 reports, none of 41 sampled reached git). Edges out of a
// node with more references than this are not followed; the test function
// itself is never pruned. The bound sits between the largest single-command
// node (12 references) and the smallest command tree (48) seen in this
// repository.
const dispatchFanOut = 40

// effect is what a helper does to a seam when called: it stores its idx-th
// parameter (field "") or a field of it in seam.
type effect struct {
	seam  string
	idx   int
	field string
}

// callSite is an executed call to the node fn; ok decides whether the
// arguments of the call satisfy an effect of fn.
type callSite struct {
	fn string
	ok func(effect) bool
}

// gnode is one node of the name graph.
type gnode struct {
	refs map[string]bool
	// Test-file facts, taken from the executed path of the body only.
	execCalls map[string]bool // functions called
	repl      map[string]bool // seams assigned a replacement
	effects   []effect        // seams stored from a parameter
	sites     []callSite      // calls made, with their arguments
	isolates  bool            // RedirectHome or GIT_CONFIG_GLOBAL
	sink      bool
	varLike   bool
}

type pkgGraph struct{ nodes map[string]*gnode }

func (g *pkgGraph) node(name string) *gnode {
	n := g.nodes[name]
	if n == nil {
		n = &gnode{refs: map[string]bool{}, execCalls: map[string]bool{}, repl: map[string]bool{}}
		g.nodes[name] = n
	}
	return n
}

// testFunc is one reportable top-level test function.
type testFunc struct {
	file   string
	name   string
	line   int
	text   string
	marked bool
}

// importNames maps each local import name of file to its import path. A blank
// or dot import contributes nothing.
func importNames(file *ast.File) map[string]string {
	names := map[string]string{}
	for _, spec := range file.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			name = spec.Name.Name
		}
		names[name] = p
	}
	return names
}

// hasStringArg reports whether call has the string literal want among its arguments.
func hasStringArg(call *ast.CallExpr, want string) bool {
	for _, a := range call.Args {
		if lit, ok := a.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil && s == want {
				return true
			}
		}
	}
	return false
}

// gitProgram reports whether call, a Command (ctxArg false) or CommandContext
// (ctxArg true), runs a program whose literal name has the base git.
func gitProgram(call *ast.CallExpr, ctxArg bool) bool {
	i := 0
	if ctxArg {
		i = 1
	}
	if len(call.Args) <= i {
		return false
	}
	lit, ok := call.Args[i].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}
	return strings.TrimSuffix(path.Base(strings.ReplaceAll(s, `\`, "/")), ".exe") == "git"
}

// scanIndirect builds the name graph of one package directory (file name to
// source, test and non-test files alike) and reports the tests that reach a
// sink. The filename keys are used verbatim in the violations.
func scanIndirect(files map[string][]byte) ([]Violation, error) {
	g, tests, err := buildGraph(files)
	if err != nil {
		return nil, err
	}
	var out []Violation
	for _, tf := range tests {
		if tf.marked || !g.exposed(tf.name) {
			continue
		}
		out = append(out, Violation{File: tf.file, Line: tf.line, Kind: KindIndirectGlobalGitConfig, Text: tf.text})
	}
	return out, nil
}

// buildGraph parses files and returns the name graph with the reportable test
// functions in file order.
func buildGraph(files map[string][]byte) (*pkgGraph, []testFunc, error) {
	fset := token.NewFileSet()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	parsed := map[string]*ast.File{}
	pkgVars := map[string]bool{}
	for _, name := range names {
		f, err := parser.ParseFile(fset, name, files[name], parser.ParseComments)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", name, err)
		}
		parsed[name] = f
		for _, d := range f.Decls {
			if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.VAR {
				for _, s := range gd.Specs {
					for _, id := range s.(*ast.ValueSpec).Names {
						pkgVars[id.Name] = true
					}
				}
			}
		}
	}

	g := &pkgGraph{nodes: map[string]*gnode{}}
	var tests []testFunc
	for _, name := range names {
		file := parsed[name]
		isTest := strings.HasSuffix(name, "_test.go")
		imports := importNames(file)
		lines := strings.Split(string(files[name]), "\n")
		markers := map[int]bool{}
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if allowComment.MatchString(c.Text) {
					markers[fset.Position(c.Slash).Line] = true
				}
			}
		}
		qualified := func(x ast.Expr) bool {
			id, ok := x.(*ast.Ident)
			if !ok {
				return false
			}
			_, imp := imports[id.Name]
			return imp && id.Obj == nil
		}
		// fill walks root on behalf of owner and records the references and the
		// sinks. scope bounds the declaration that holds root, so an identifier
		// resolved inside it is a local value.
		var fill func(owner *gnode, root, scope ast.Node)
		fill = func(owner *gnode, root, scope ast.Node) {
			local := func(id *ast.Ident) bool {
				if id.Obj == nil {
					return false
				}
				d, ok := id.Obj.Decl.(ast.Node)
				return ok && d.Pos() >= scope.Pos() && d.End() <= scope.End()
			}
			var visit func(ast.Node) bool
			visit = func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					if x.Name != "_" && !local(x) {
						owner.refs[x.Name] = true
					}
				case *ast.SelectorExpr:
					ast.Inspect(x.X, visit)
					if qualified(x.X) {
						owner.refs[x.Sel.Name] = true
					}
					return false
				case *ast.KeyValueExpr:
					// A struct field key is a field name, not a reference.
					if _, isIdent := x.Key.(*ast.Ident); !isIdent {
						ast.Inspect(x.Key, visit)
					}
					ast.Inspect(x.Value, visit)
					return false
				case *ast.CallExpr:
					if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
						if id, ok := sel.X.(*ast.Ident); ok && id.Obj == nil && imports[id.Name] == "os/exec" &&
							(sel.Sel.Name == "Command" || sel.Sel.Name == "CommandContext") &&
							gitProgram(x, sel.Sel.Name == "CommandContext") &&
							hasStringArg(x, "config") && hasStringArg(x, "--global") {
							owner.sink = true
						}
					}
				case *ast.AssignStmt:
					// Production wiring (X = f, pkg.X = f) gives the variable its
					// default value; assignments in test files are swaps and are
					// read by execFacts.
					if isTest || x.Tok != token.ASSIGN {
						break
					}
					for i, lhs := range x.Lhs {
						assigned := ""
						switch l := lhs.(type) {
						case *ast.Ident:
							if pkgVars[l.Name] && !local(l) {
								assigned = l.Name
							}
						case *ast.SelectorExpr:
							if qualified(l.X) {
								assigned = l.Sel.Name
							}
						}
						if assigned == "" {
							continue
						}
						target := g.node(assigned)
						target.varLike = true
						rhs := ast.Node(x.Rhs[0])
						if len(x.Rhs) == len(x.Lhs) {
							rhs = x.Rhs[i]
						}
						fill(target, rhs, scope)
					}
				}
				return true
			}
			ast.Inspect(root, visit)
		}

		// execFacts records, for the test-file function fd, what its executed
		// path does to the seams: see the header for what executed means.
		execFacts := func(owner *gnode, fd *ast.FuncDecl) {
			local := func(id *ast.Ident) bool {
				if id.Obj == nil {
					return false
				}
				d, ok := id.Obj.Decl.(ast.Node)
				return ok && d.Pos() >= fd.Pos() && d.End() <= fd.End()
			}
			params := map[*ast.Object]int{}
			if fd.Type.Params != nil {
				i := 0
				for _, f := range fd.Type.Params.List {
					if len(f.Names) == 0 {
						i++
						continue
					}
					for _, nm := range f.Names {
						if nm.Obj != nil {
							params[nm.Obj] = i
						}
						i++
					}
				}
			}
			// run holds the closures that execute on the spot: invoked at once, or
			// passed to Run.
			run := map[*ast.FuncLit]bool{}
			ast.Inspect(fd, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					if lit, ok := c.Fun.(*ast.FuncLit); ok {
						run[lit] = true
					}
					if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Run" {
						for _, a := range c.Args {
							if lit, ok := a.(*ast.FuncLit); ok {
								run[lit] = true
							}
						}
					}
				}
				return true
			})
			// defsOf returns every value assigned to the local obj in fd. ok is
			// false when there is none or one cannot be read (a parameter, a zero
			// value, a tuple assignment).
			defsOf := func(obj *ast.Object) (vals []ast.Expr, ok bool) {
				ok = true
				ast.Inspect(fd, func(n ast.Node) bool {
					switch d := n.(type) {
					case *ast.AssignStmt:
						if d.Tok != token.ASSIGN && d.Tok != token.DEFINE {
							return true
						}
						for j, l := range d.Lhs {
							if li, isID := l.(*ast.Ident); isID && li.Obj == obj {
								if len(d.Lhs) == len(d.Rhs) {
									vals = append(vals, d.Rhs[j])
								} else {
									ok = false
								}
							}
						}
					case *ast.ValueSpec:
						for j, nm := range d.Names {
							if nm.Obj == obj {
								if len(d.Values) == len(d.Names) {
									vals = append(vals, d.Values[j])
								} else {
									ok = false
								}
							}
						}
					}
					return true
				})
				return vals, ok && len(vals) > 0
			}
			// inLit is the set of closures already walked from a call through a
			// local variable, so a self-calling closure ends.
			inLit := map[*ast.FuncLit]bool{}
			// fromSeam reports whether e is derived from the seam's own value.
			var fromSeam func(e ast.Expr, seam string, seen map[*ast.Object]bool) bool
			fromSeam = func(e ast.Expr, seam string, seen map[*ast.Object]bool) bool {
				switch v := e.(type) {
				case *ast.ParenExpr:
					return fromSeam(v.X, seam, seen)
				case *ast.Ident:
					if v.Name == seam {
						return true
					}
					if v.Obj == nil || !local(v) || seen[v.Obj] {
						return false
					}
					seen[v.Obj] = true
					vals, _ := defsOf(v.Obj)
					for _, d := range vals {
						if fromSeam(d, seam, seen) {
							return true
						}
					}
				case *ast.CallExpr:
					for _, a := range v.Args {
						if fromSeam(a, seam, seen) {
							return true
						}
					}
				}
				return false
			}
			// isRepl reports whether e is a replacement for the seam on every
			// assignment that can reach it.
			var isRepl func(e ast.Expr, seam string, seen map[*ast.Object]bool) bool
			isRepl = func(e ast.Expr, seam string, seen map[*ast.Object]bool) bool {
				switch v := e.(type) {
				case *ast.ParenExpr:
					return isRepl(v.X, seam, seen)
				case *ast.FuncLit:
					// A literal that uses the seam's own value (directly, or through a
					// local derived from it) still reaches the real function.
					uses := false
					ast.Inspect(v.Body, func(n ast.Node) bool {
						if id, ok := n.(*ast.Ident); ok && fromSeam(id, seam, map[*ast.Object]bool{}) {
							uses = true
						}
						return !uses
					})
					return !uses
				case *ast.Ident:
					if v.Name == "nil" || v.Name == seam {
						return false
					}
					if v.Obj != nil && local(v) {
						if seen[v.Obj] {
							return false
						}
						seen[v.Obj] = true
						vals, ok := defsOf(v.Obj)
						if !ok {
							return false
						}
						for _, d := range vals {
							if !isRepl(d, seam, seen) {
								return false
							}
						}
						return true
					}
					return !pkgVars[v.Name]
				case *ast.CallExpr:
					return !fromSeam(v, seam, map[*ast.Object]bool{})
				}
				return false
			}
			// litSets reports whether e is a composite literal, in place or in a
			// local, that sets field to a replacement on every assignment.
			var litSets func(e ast.Expr, field, seam string, seen map[*ast.Object]bool) bool
			litSets = func(e ast.Expr, field, seam string, seen map[*ast.Object]bool) bool {
				switch v := e.(type) {
				case *ast.ParenExpr:
					return litSets(v.X, field, seam, seen)
				case *ast.UnaryExpr:
					return litSets(v.X, field, seam, seen)
				case *ast.CompositeLit:
					for _, el := range v.Elts {
						if kv, ok := el.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok && key.Name == field {
								return isRepl(kv.Value, seam, map[*ast.Object]bool{})
							}
						}
					}
				case *ast.Ident:
					if v.Obj == nil || !local(v) || seen[v.Obj] {
						return false
					}
					seen[v.Obj] = true
					vals, ok := defsOf(v.Obj)
					if !ok {
						return false
					}
					for _, d := range vals {
						if !litSets(d, field, seam, seen) {
							return false
						}
					}
					return true
				}
				return false
			}
			isFalse := func(e ast.Expr) bool {
				id, ok := e.(*ast.Ident)
				return ok && id.Name == "false"
			}
			var walk func(ast.Node) bool
			walk = func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncLit:
					return run[x]
				case *ast.DeferStmt, *ast.GoStmt:
					return false
				case *ast.IfStmt:
					if isFalse(x.Cond) {
						if x.Init != nil {
							ast.Inspect(x.Init, walk)
						}
						if x.Else != nil {
							ast.Inspect(x.Else, walk)
						}
						return false
					}
				case *ast.ForStmt:
					if isFalse(x.Cond) {
						if x.Init != nil {
							ast.Inspect(x.Init, walk)
						}
						return false
					}
				case *ast.RangeStmt:
					switch c := x.X.(type) {
					case *ast.CompositeLit:
						if len(c.Elts) == 0 {
							ast.Inspect(x.X, walk)
							return false
						}
					case *ast.BasicLit:
						if c.Kind == token.STRING && (c.Value == `""` || c.Value == "``") {
							return false
						}
					}
				case *ast.SwitchStmt:
					if x.Tag != nil {
						break
					}
					if x.Init != nil {
						ast.Inspect(x.Init, walk)
					}
					for _, s := range x.Body.List {
						cc := s.(*ast.CaseClause)
						off := len(cc.List) > 0
						for _, e := range cc.List {
							if !isFalse(e) {
								off = false
							}
						}
						if off {
							continue
						}
						for _, e := range cc.List {
							ast.Inspect(e, walk)
						}
						for _, st := range cc.Body {
							ast.Inspect(st, walk)
						}
					}
					return false
				case *ast.AssignStmt:
					if x.Tok != token.ASSIGN || len(x.Lhs) != len(x.Rhs) {
						break
					}
					for i, lhs := range x.Lhs {
						assigned := ""
						switch l := lhs.(type) {
						case *ast.Ident:
							if pkgVars[l.Name] && !local(l) {
								assigned = l.Name
							}
						case *ast.SelectorExpr:
							if qualified(l.X) {
								assigned = l.Sel.Name
							}
						}
						if assigned == "" {
							continue
						}
						switch r := x.Rhs[i].(type) {
						case *ast.Ident:
							if idx, isParam := params[r.Obj]; isParam && r.Obj != nil {
								owner.effects = append(owner.effects, effect{seam: assigned, idx: idx})
								continue
							}
						case *ast.SelectorExpr:
							if id, ok := r.X.(*ast.Ident); ok && id.Obj != nil {
								if idx, isParam := params[id.Obj]; isParam {
									owner.effects = append(owner.effects, effect{seam: assigned, idx: idx, field: r.Sel.Name})
									continue
								}
							}
						}
						if isRepl(x.Rhs[i], assigned, map[*ast.Object]bool{}) {
							owner.repl[assigned] = true
						}
					}
				case *ast.CallExpr:
					switch f := x.Fun.(type) {
					case *ast.Ident:
						if local(f) {
							// A closure held in a local runs here, at this executed call.
							vals, _ := defsOf(f.Obj)
							for _, v := range vals {
								if lit, ok := v.(*ast.FuncLit); ok && !inLit[lit] {
									inLit[lit] = true
									ast.Inspect(lit.Body, walk)
								}
							}
							break
						}
						if f.Name == "RedirectHome" {
							owner.isolates = true
						}
						owner.execCalls[f.Name] = true
						args := x.Args
						owner.sites = append(owner.sites, callSite{fn: f.Name, ok: func(e effect) bool {
							if e.idx >= len(args) {
								return false
							}
							if e.field == "" {
								return isRepl(args[e.idx], e.seam, map[*ast.Object]bool{})
							}
							return litSets(args[e.idx], e.field, e.seam, map[*ast.Object]bool{})
						}})
					case *ast.SelectorExpr:
						if qualified(f.X) && f.Sel.Name == "RedirectHome" {
							owner.isolates = true
						}
						if f.Sel.Name == "Setenv" && len(x.Args) > 0 {
							if lit, ok := x.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
								if s, err := strconv.Unquote(lit.Value); err == nil && s == gitGlobalConfigEnv {
									owner.isolates = true
								}
							}
						}
					}
				}
				return true
			}
			ast.Inspect(fd.Body, walk)
		}

		for _, d := range file.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				if decl.Recv != nil || decl.Body == nil {
					continue
				}
				owner := g.node(decl.Name.Name)
				fill(owner, decl, decl)
				if isTest {
					execFacts(owner, decl)
				}
				if !isTest || !isTestEntry(decl.Name.Name) {
					continue
				}
				line := fset.Position(decl.Pos()).Line
				tests = append(tests, testFunc{
					file: name, name: decl.Name.Name, line: line,
					text:   strings.TrimSpace(lineAt(lines, line)),
					marked: markers[line] || markers[line-1],
				})
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				for _, s := range decl.Specs {
					spec := s.(*ast.ValueSpec)
					for i, id := range spec.Names {
						n := g.node(id.Name)
						n.varLike = true
						switch {
						case len(spec.Values) == len(spec.Names):
							fill(n, spec.Values[i], spec)
						default:
							for _, v := range spec.Values {
								fill(n, v, spec)
							}
						}
					}
				}
			}
		}
	}

	return g, tests, nil
}

// mentionsGlobalFlag reports whether any file spells the "--global" literal. A
// sink needs that literal, so a directory without it has nothing to report and
// is not parsed again.
func mentionsGlobalFlag(files map[string][]byte) bool {
	fset := token.NewFileSet()
	for _, src := range files {
		var sc scanner.Scanner
		sc.Init(fset.AddFile("", fset.Base(), len(src)), src, nil, 0)
		for {
			_, tok, lit := sc.Scan()
			if tok == token.EOF {
				break
			}
			if tok != token.STRING {
				continue
			}
			// Unquote exactly as hasStringArg does, so an escaped or raw spelling
			// of the flag is found too.
			if s, err := strconv.Unquote(lit); err == nil && s == "--global" {
				return true
			}
		}
	}
	return false
}

// isTestEntry reports whether name is a top-level function the go tool runs.
func isTestEntry(name string) bool {
	for _, p := range []string{"Test", "Benchmark", "Fuzz"} {
		if strings.HasPrefix(name, p) && name != "TestMain" {
			return true
		}
	}
	return false
}

// exposed reports whether the test function start reaches a sink along a path
// whose seams the test does not isolate.
func (g *pkgGraph) exposed(start string) bool {
	root := g.nodes[start]
	if root == nil {
		return false
	}
	// What the test executes decides what it swaps and whether it redirects the
	// global config: follow executed calls only.
	swapped := map[string]bool{}
	var sites []callSite
	seen := map[string]bool{start: true}
	for stack := []string{start}; len(stack) > 0; {
		n := g.nodes[stack[len(stack)-1]]
		stack = stack[:len(stack)-1]
		if n.isolates {
			return false
		}
		for s := range n.repl {
			swapped[s] = true
		}
		sites = append(sites, n.sites...)
		// A dispatch table is not expanded: a helper only it references is not
		// something this test calls.
		if n != root && g.fanOut(n) > dispatchFanOut {
			continue
		}
		for c := range n.execCalls {
			if g.nodes[c] != nil && !seen[c] {
				seen[c] = true
				stack = append(stack, c)
			}
		}
	}
	// A helper that stores a parameter in a seam isolates it only when the call
	// passes a replacement at that position.
	for _, cs := range sites {
		h := g.nodes[cs.fn]
		if h == nil {
			continue
		}
		for _, e := range h.effects {
			if cs.ok(e) {
				swapped[e.seam] = true
			}
		}
	}
	seen = map[string]bool{start: true}
	for stack := []string{start}; len(stack) > 0; {
		n := g.nodes[stack[len(stack)-1]]
		stack = stack[:len(stack)-1]
		if n.sink {
			return true
		}
		if n != root && g.fanOut(n) > dispatchFanOut {
			continue
		}
		for r := range n.refs {
			t := g.nodes[r]
			if t == nil || seen[r] || (t.varLike && swapped[r]) {
				continue
			}
			seen[r] = true
			stack = append(stack, r)
		}
	}
	return false
}

// fanOut counts the nodes n references.
func (g *pkgGraph) fanOut(n *gnode) int {
	c := 0
	for r := range n.refs {
		if g.nodes[r] != nil {
			c++
		}
	}
	return c
}
