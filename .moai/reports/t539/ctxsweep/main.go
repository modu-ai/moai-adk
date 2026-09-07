// ctxsweep — mechanical sweep for t539.
//
// Lists every function in *_test.go files under the given root that RECEIVES
// a context-bearing parameter (context.Context or *http.Request) and reports
// whether the body ever CONSULTS that parameter. A test double that receives a
// context and never reads it is "context-blind": it cannot fail the way the
// real dependency fails when the context is dead, so a production defect in
// context derivation passes straight through it.
//
// This tool answers question (b) of the card's discriminator — "does the
// double consult the context?" — mechanically. Questions (a) "does the real
// thing consult it?" and (c) "do they disagree?" are judged per interface in
// the sweep report; they are not derivable from the test file alone.
//
// "Consults" is judged per parameter type:
//   context.Context  — any use of the identifier in the body (reading
//                      Done/Err/Deadline/Value, or passing it downstream).
//   *http.Request    — a call of <name>.Context() somewhere in the body.
//                      Reading req.URL or req.Body is NOT consulting the
//                      context; a double can read the body and still answer
//                      a dead request.
//
// Usage: go run main.go [-prod] <root>
//   default : scan *_test.go only (test doubles)
//   -prod   : scan non-test .go files only (the real implementations), so the
//             two tables can be joined by method name for question (a)
// Output: TSV  kind  file:line  receiver  method  param  consults
//   kind     = method | funclit
//   consults = yes | no | "" (parameter is unnamed or named _)
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type row struct{ kind, pos, owner, method, param, consults string }

func main() {
	args := os.Args[1:]
	prod := false
	if len(args) > 0 && args[0] == "-prod" {
		prod = true
		args = args[1:]
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: ctxsweep [-prod] <root>")
		os.Exit(2)
	}
	root := args[0]
	var rows []row
	fset := token.NewFileSet()
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") {
			return nil
		}
		if strings.HasSuffix(p, "_test.go") == prod {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		ast.Inspect(f, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil || d.Body == nil {
					return true
				}
				owner := recvName(d.Recv)
				for _, r := range inspectParams(d.Type, d.Body) {
					r.kind, r.owner, r.method = "method", owner, d.Name.Name
					r.pos = fmt.Sprintf("%s:%d", rel, fset.Position(d.Pos()).Line)
					rows = append(rows, r)
				}
			case *ast.FuncLit:
				for _, r := range inspectParams(d.Type, d.Body) {
					r.kind, r.owner, r.method = "funclit", "-", "-"
					r.pos = fmt.Sprintf("%s:%d", rel, fset.Position(d.Pos()).Line)
					rows = append(rows, r)
				}
			}
			return true
		})
		return nil
	})
	sort.Slice(rows, func(i, j int) bool { return rows[i].pos < rows[j].pos })
	for _, r := range rows {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", r.kind, r.pos, r.owner, r.method, r.param, r.consults)
	}
}

func recvName(fl *ast.FieldList) string {
	if len(fl.List) == 0 {
		return "?"
	}
	t := fl.List[0].Type
	if s, ok := t.(*ast.StarExpr); ok {
		t = s.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name
	}
	return "?"
}

func typeStr(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return "*" + typeStr(t.X)
	case *ast.SelectorExpr:
		return typeStr(t.X) + "." + t.Sel.Name
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// inspectParams returns one row per context-bearing parameter of the function.
func inspectParams(ft *ast.FuncType, body *ast.BlockStmt) []row {
	var out []row
	if ft.Params == nil {
		return nil
	}
	for _, f := range ft.Params.List {
		ts := typeStr(f.Type)
		if ts != "context.Context" && ts != "*http.Request" {
			continue
		}
		if len(f.Names) == 0 {
			out = append(out, row{param: ts + " (unnamed)", consults: ""})
			continue
		}
		for _, nm := range f.Names {
			if nm.Name == "_" {
				out = append(out, row{param: ts + " _", consults: ""})
				continue
			}
			used := false
			ast.Inspect(body, func(n ast.Node) bool {
				if used {
					return false
				}
				if ts == "*http.Request" {
					// only <name>.Context() counts
					sel, ok := n.(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "Context" {
						return true
					}
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == nm.Name && id.Obj == nm.Obj {
						used = true
					}
					return true
				}
				if id, ok := n.(*ast.Ident); ok && id.Name == nm.Name && id.Obj == nm.Obj {
					used = true
				}
				return true
			})
			c := "no"
			if used {
				c = "yes"
			}
			out = append(out, row{param: ts + " " + nm.Name, consults: c})
		}
	}
	return out
}
