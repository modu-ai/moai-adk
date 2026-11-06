package claim

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

// The store has per-Issue locks, a documents lock, a requests lock and an agent
// registry lock. The order is fixed: an Issue lock may be held while the
// documents lock is taken (a bind or a handoff reserves a worktree inside its
// update), and the documents lock is a leaf: nothing run under it may take any
// other lock. This reads the source and fails when code reachable from a
// withDocs body can take another lock, which is the half of a cycle the other
// direction would need.
func TestTheDocumentsLockIsALeaf(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]*ast.File{} // the non-test files of package claim, by file name
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name.Name == "claim" {
			files[name] = f
		}
	}
	if len(files) == 0 {
		t.Fatal("package claim not found")
	}
	calls := map[string]map[string]bool{} // function name -> names it calls
	var docBodies []*ast.FuncLit
	for _, f := range files {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			name := fn.Name.Name
			if calls[name] == nil {
				calls[name] = map[string]bool{}
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				c, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				callee := calleeName(c)
				if callee == "" {
					return true
				}
				if callee == "withDocs" && name != "withDocs" {
					for _, a := range c.Args {
						if lit, ok := a.(*ast.FuncLit); ok {
							docBodies = append(docBodies, lit)
						}
					}
					return true // its body is walked as a docs body, not as part of this function
				}
				calls[name][callee] = true
				return true
			})
		}
	}
	if len(docBodies) < 2 {
		t.Fatalf("expected the withDocs bodies of PutDoc and DeleteDoc, found %d", len(docBodies))
	}
	lockTakers := map[string]bool{"withLock": true, "withRequests": true, "withLockFile": true, "withAgents": true}
	for i, lit := range docBodies {
		seen := map[string]bool{}
		var queue []string
		ast.Inspect(lit.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if cn := calleeName(c); cn != "" {
					queue = append(queue, cn)
				}
			}
			return true
		})
		var bad []string
		for len(queue) > 0 {
			cn := queue[0]
			queue = queue[1:]
			if seen[cn] {
				continue
			}
			seen[cn] = true
			if lockTakers[cn] {
				bad = append(bad, cn)
				continue
			}
			for next := range calls[cn] {
				queue = append(queue, next)
			}
		}
		sort.Strings(bad)
		if len(bad) > 0 {
			t.Errorf("withDocs body #%d can reach another lock through %v: the documents lock must stay a leaf", i, bad)
		}
	}
}

// calleeName is the name of the function or method a call names, "" for anything
// else (a func value, a conversion).
func calleeName(c *ast.CallExpr) string {
	switch f := c.Fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}
