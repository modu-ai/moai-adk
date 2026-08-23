package web

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestListenAndServeRegistersSignalsBeforeBinding locks the ordering the
// t199 fix established: signal.NotifyContext MUST run before the listener is
// bound.
//
// Why the invariant is load-bearing: a caller learns the server is up by
// observing the listener address, so anything that acts on that observation —
// the self-SIGTERM shutdown test, a supervisor, an operator — races a
// registration performed afterwards. Until signal.NotifyContext runs, SIGTERM
// keeps its default disposition and kills the process. With the bind first,
// that window is real; it killed the whole `go test ./internal/web/` binary on
// Linux with "signal: terminated" and no test-failure line to read.
//
// The check is structural rather than textual: the file is parsed and the two
// call sites are located as AST nodes inside ListenAndServe's body, so a
// reordering is caught by position, and a rename or a reformat that preserves
// the ordering does not produce a false alarm. A behavioural test cannot cover
// this — the window it guards is exactly the one where the process dies rather
// than fails.
func TestListenAndServeRegistersSignalsBeforeBinding(t *testing.T) {
	const file = "server.go"

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}

	body := methodBody(t, parsed, "Server", "ListenAndServe")

	notifyPos := callPos(body, isSelectorCall("signal", "NotifyContext"))
	if !notifyPos.IsValid() {
		t.Fatal("ListenAndServe does not call signal.NotifyContext; the shutdown path is gone")
	}
	bindPos := callPos(body, isMethodCall("bind"))
	if !bindPos.IsValid() {
		t.Fatal("ListenAndServe does not call s.bind(); the listener is bound somewhere unexpected")
	}

	if notifyPos > bindPos {
		t.Errorf(
			"signal.NotifyContext at %s runs AFTER s.bind() at %s; "+
				"a SIGTERM arriving between the bind and the registration kills the process",
			fset.Position(notifyPos), fset.Position(bindPos),
		)
	}
}

// TestBindIsTheOnlyListenSite pins the premise the ordering test rests on:
// net.Listen is reached through bind and nowhere else, so guarding the bind
// call guards every path that opens the listener.
func TestBindIsTheOnlyListenSite(t *testing.T) {
	const file = "server.go"

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}

	var enclosing []string
	ast.Inspect(parsed, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if callPos(fn.Body, isSelectorCall("net", "Listen")).IsValid() {
			enclosing = append(enclosing, fn.Name.Name)
		}
		return true
	})

	if len(enclosing) != 1 || enclosing[0] != "bind" {
		t.Errorf("net.Listen is called from %v, want only from bind; "+
			"the signal-ordering guard only covers the bind path", enclosing)
	}
}

// methodBody returns the body of the named method on the named receiver type.
func methodBody(t *testing.T, f *ast.File, recv, name string) *ast.BlockStmt {
	t.Helper()
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		if ident, ok := star.X.(*ast.Ident); ok && ident.Name == recv {
			return fn.Body
		}
	}
	t.Fatalf("method (*%s).%s not found", recv, name)
	return nil
}

// callPos reports the position of the first call in body matching want, or an
// invalid position when there is none.
func callPos(body *ast.BlockStmt, want func(*ast.CallExpr) bool) token.Pos {
	found := token.NoPos
	if body == nil {
		return found
	}
	ast.Inspect(body, func(n ast.Node) bool {
		if found.IsValid() {
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok && want(call) {
			found = call.Pos()
			return false
		}
		return true
	})
	return found
}

// isSelectorCall matches a package-qualified call such as signal.NotifyContext.
func isSelectorCall(pkg, name string) func(*ast.CallExpr) bool {
	return func(call *ast.CallExpr) bool {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != name {
			return false
		}
		ident, ok := sel.X.(*ast.Ident)
		return ok && ident.Name == pkg
	}
}

// isMethodCall matches a call on the receiver, such as s.bind().
func isMethodCall(name string) func(*ast.CallExpr) bool {
	return func(call *ast.CallExpr) bool {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != name {
			return false
		}
		_, ok = sel.X.(*ast.Ident)
		return ok
	}
}
