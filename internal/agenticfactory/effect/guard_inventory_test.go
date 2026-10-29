package effect

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Any non-test source under effect/ that names a write method (or gh's "-X" flag) must go through
// CheckRequest. A new write path added without the guard fails here, so the refusal stays in one place.
func TestEveryWritePathUnderEffectUsesTheGuard(t *testing.T) {
	writes := map[string]bool{`"POST"`: true, `"PATCH"`: true, `"PUT"`: true, `"DELETE"`: true, `"-X"`: true}
	checked := 0
	err := filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		af, perr := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if perr != nil {
			return perr
		}
		var names, calls bool
		ast.Inspect(af, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.BasicLit:
				if x.Kind == token.STRING && writes[x.Value] {
					names = true
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "MethodPost" || x.Sel.Name == "MethodPatch" || x.Sel.Name == "MethodPut" || x.Sel.Name == "MethodDelete" {
					names = true
				}
			case *ast.Ident:
				if x.Name == "CheckRequest" {
					calls = true
				}
			}
			return true
		})
		if names {
			checked++
			if !calls {
				t.Errorf("%s names a write method but never calls CheckRequest", p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 3 {
		t.Errorf("only %d write-capable files found; the inventory is not seeing the senders", checked)
	}
}
