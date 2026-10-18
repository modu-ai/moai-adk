package testguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
//     exec.Command or exec.CommandContext whose literal arguments include
//     "config" and "--global".
//   - A seam is a variable node. A test isolates a seam when it, or a function
//     it reaches in a test file, assigns it a replacement: a restore inside a
//     Cleanup closure does not count, and an assignment from a parameter field
//     (seam = s.field) counts only when reachable test code sets that field to
//     a non-nil value in a composite literal. RedirectHome, or setting
//     GIT_CONFIG_GLOBAL, isolates every seam.
//
// A top-level Test, Benchmark or Fuzz function that reaches a sink along a path
// with no isolated seam is reported once, on its declaration line, unless the
// line or the line above carries `home-guard:allow <reason>`. The graph is
// conservative about what it follows (every name match is an edge) and blind to
// what it cannot name: method calls, calls through struct fields and
// interfaces, sinks in another package that are not reached through an
// assigned variable, and anything behind a dispatch table (dispatchFanOut).

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

// gnode is one node of the name graph.
type gnode struct {
	refs    map[string]bool
	assigns map[string]bool // names assigned a replacement in this body; test files only
	// guarded maps a seam to the parameter fields it is assigned from
	// (seam = s.field); passes holds the struct fields this body sets to a
	// non-nil value in a composite literal. Test files only.
	guarded  map[string]map[string]bool
	passes   map[string]bool
	sink     bool
	varLike  bool
	isolates bool // test files only
}

type pkgGraph struct{ nodes map[string]*gnode }

func (g *pkgGraph) node(name string) *gnode {
	n := g.nodes[name]
	if n == nil {
		n = &gnode{refs: map[string]bool{}, assigns: map[string]bool{}, guarded: map[string]map[string]bool{}, passes: map[string]bool{}}
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
		// cleanups holds the extent of every closure handed to a Cleanup call.
		var cleanups [][2]token.Pos
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Cleanup" {
					for _, a := range call.Args {
						if lit, ok := a.(*ast.FuncLit); ok {
							cleanups = append(cleanups, [2]token.Pos{lit.Pos(), lit.End()})
						}
					}
				}
			}
			return true
		})
		inCleanup := func(p token.Pos) bool {
			for _, r := range cleanups {
				if p >= r[0] && p < r[1] {
					return true
				}
			}
			return false
		}
		lines := strings.Split(string(files[name]), "\n")
		markers := map[int]bool{}
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if allowComment.MatchString(c.Text) {
					markers[fset.Position(c.Slash).Line] = true
				}
			}
		}
		// fill walks root on behalf of owner. scope bounds the declaration that
		// holds root, so an identifier resolved inside it is a local value.
		var fill func(owner *gnode, root, scope ast.Node)
		fill = func(owner *gnode, root, scope ast.Node) {
			local := func(id *ast.Ident) bool {
				if id.Obj == nil {
					return false
				}
				d, ok := id.Obj.Decl.(ast.Node)
				return ok && d.Pos() >= scope.Pos() && d.End() <= scope.End()
			}
			qualified := func(x ast.Expr) bool {
				id, ok := x.(*ast.Ident)
				if !ok {
					return false
				}
				_, imp := imports[id.Name]
				return imp && id.Obj == nil
			}
			// paramField names the field when assignment x stores a field of a
			// local value (s.field), the shape of a helper that swaps a seam
			// only for the fields its caller passes.
			paramField := func(x *ast.AssignStmt, i int) string {
				if len(x.Rhs) != len(x.Lhs) {
					return ""
				}
				sel, ok := x.Rhs[i].(*ast.SelectorExpr)
				if !ok {
					return ""
				}
				if id, ok := sel.X.(*ast.Ident); ok && id.Obj != nil && local(id) {
					return sel.Sel.Name
				}
				return ""
			}
			var visit func(ast.Node) bool
			visit = func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					if x.Name != "_" && !local(x) {
						owner.refs[x.Name] = true
						// The guard's own package calls RedirectHome unqualified.
						if isTest && x.Name == "RedirectHome" {
							owner.isolates = true
						}
					}
				case *ast.SelectorExpr:
					ast.Inspect(x.X, visit)
					if qualified(x.X) {
						owner.refs[x.Sel.Name] = true
						if isTest && x.Sel.Name == "RedirectHome" {
							owner.isolates = true
						}
					}
					return false
				case *ast.KeyValueExpr:
					// A struct field key is a field name, not a reference.
					if key, isIdent := x.Key.(*ast.Ident); !isIdent {
						ast.Inspect(x.Key, visit)
					} else if isTest {
						// nil passes nothing, and a package variable as the value is
						// the capture of a seam's current value for a later restore.
						v, isIdent := x.Value.(*ast.Ident)
						if !isIdent || (v.Name != "nil" && !pkgVars[v.Name]) {
							owner.passes[key.Name] = true
						}
					}
					ast.Inspect(x.Value, visit)
					return false
				case *ast.CallExpr:
					if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
						if id, ok := sel.X.(*ast.Ident); ok && id.Obj == nil && imports[id.Name] == "os/exec" &&
							(sel.Sel.Name == "Command" || sel.Sel.Name == "CommandContext") &&
							hasStringArg(x, "config") && hasStringArg(x, "--global") {
							owner.sink = true
						}
						if isTest && sel.Sel.Name == "Setenv" && hasStringArg(x, gitGlobalConfigEnv) {
							owner.isolates = true
						}
					}
				case *ast.AssignStmt:
					if x.Tok != token.ASSIGN {
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
						if isTest {
							switch {
							case inCleanup(x.Pos()):
								// A restore to the original value swaps nothing.
							case paramField(x, i) != "":
								if owner.guarded[assigned] == nil {
									owner.guarded[assigned] = map[string]bool{}
								}
								owner.guarded[assigned][paramField(x, i)] = true
							default:
								owner.assigns[assigned] = true
							}
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

		for _, d := range file.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				if decl.Recv != nil || decl.Body == nil {
					continue
				}
				fill(g.node(decl.Name.Name), decl, decl)
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
	for _, src := range files {
		s := string(src)
		if strings.Contains(s, `"--global"`) || strings.Contains(s, "`--global`") {
			return true
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
	// Everything the test can reach decides what it swaps and whether it
	// redirects the global config.
	swapped := map[string]bool{}
	passed := map[string]bool{}
	guarded := map[string]map[string]bool{}
	seen := map[string]bool{start: true}
	for stack := []string{start}; len(stack) > 0; {
		n := g.nodes[stack[len(stack)-1]]
		stack = stack[:len(stack)-1]
		if n.isolates {
			return false
		}
		for a := range n.assigns {
			swapped[a] = true
		}
		for k := range n.passes {
			passed[k] = true
		}
		for seam, fields := range n.guarded {
			if guarded[seam] == nil {
				guarded[seam] = map[string]bool{}
			}
			for f := range fields {
				guarded[seam][f] = true
			}
		}
		for r := range n.refs {
			if g.nodes[r] != nil && !seen[r] {
				seen[r] = true
				stack = append(stack, r)
			}
		}
	}
	// A seam a helper replaces only when its parameter field is set counts only
	// when some reachable test code passes that field.
	for seam, fields := range guarded {
		for f := range fields {
			if passed[f] {
				swapped[seam] = true
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
