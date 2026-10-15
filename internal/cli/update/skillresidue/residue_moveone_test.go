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
// every reference to the rename seam (moveFile) and to the single rename site
// (moveOne), by enclosing declaration. The registry row for moveOne
// (internal/cli/update_destructive_registry.go) counts the one os.Rename behind
// them. A new reference to either name is a new route to a rename: the test fails
// and forces that row to be reviewed. The declaration name of moveOne itself is not
// a reference, so it is not counted.
func TestMoveFileReferencesArePinned(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	// got maps "name@owner" to a count, e.g. "moveFile@renameSkillFile".
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
			// Walk the function's signature and body, never its name: the name of
			// moveOne is its declaration, not a reference to it.
			nodes := []ast.Node{decl}
			if fn, ok := decl.(*ast.FuncDecl); ok {
				owner = fn.Name.Name
				nodes = []ast.Node{fn.Type, fn.Body}
				if fn.Recv != nil {
					nodes = append(nodes, fn.Recv)
				}
			}
			for _, root := range nodes {
				if root == nil {
					continue
				}
				ast.Inspect(root, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok && (id.Name == "moveFile" || id.Name == "moveOne") {
						got[id.Name+"@"+owner]++
					}
					return true
				})
			}
		}
	}

	// The only allowed references: moveFile is declared at package level (its name is
	// the declaration, counted once) and renameSkillFile makes its three moves through
	// it; moveOne is referenced once, as the value of the package-level declaration.
	want := map[string]int{
		"moveFile@package-level":   1,
		"moveFile@renameSkillFile": 3,
		"moveOne@package-level":    1,
	}
	for key, n := range want {
		if got[key] != n {
			t.Errorf("%s references = %d, want %d", key, got[key], n)
		}
	}
	for key, n := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("new reference %s (%d): review the moveOne registry row", key, n)
		}
	}
}
