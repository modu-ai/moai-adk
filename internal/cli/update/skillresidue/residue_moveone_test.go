package skillresidue

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestMoveFileReferencesArePinned parses the package's non-test sources and counts
// references to the moveFile seam by enclosing declaration. The registry row for
// moveOne (internal/cli/update_destructive_registry.go) counts the single os.Rename
// behind that seam; a new reference to moveFile is a new route to a rename, so it
// must fail here and force the row to be reviewed.
func TestMoveFileReferencesArePinned(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	got := map[string]int{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			owner := "package-level"
			if fn, ok := decl.(*ast.FuncDecl); ok {
				owner = fn.Name.Name
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == "moveFile" {
					got[owner]++
				}
				return true
			})
		}
	}

	want := map[string]int{"package-level": 1, "renameSkillFile": 3}
	for owner, n := range want {
		if got[owner] != n {
			t.Errorf("moveFile references in %s = %d, want %d", owner, got[owner], n)
		}
	}
	for owner, n := range got {
		if _, ok := want[owner]; !ok {
			t.Errorf("new moveFile reference in %s (%d): review the moveOne registry row", owner, n)
		}
	}
}
