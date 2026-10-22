package privatefile

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// Linux can check the native call boundary even though executing its Windows
// long-path behavior requires the native Windows suite below.
func TestNativeWindowsPathsNormalizeBeforeUTF16(t *testing.T) {
	conversions := 0
	for _, path := range []string{"privatefile_windows.go", "existing_windows.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "UTF16PtrFromString" {
				return true
			}
			conversions++
			if len(call.Args) != 1 {
				t.Errorf("%s: invalid native path conversion", path)
				return true
			}
			normalized, ok := call.Args[0].(*ast.CallExpr)
			if !ok {
				t.Errorf("%s: native UTF16 path has no extended-path normalization", path)
				return true
			}
			name, ok := normalized.Fun.(*ast.Ident)
			if !ok || name.Name != "nativePath" {
				t.Errorf("%s: native conversion does not use the shared path policy", path)
			}
			return true
		})
	}
	if conversions != 5 {
		t.Fatalf("inspected %d native path boundaries, want5", conversions)
	}
}
