package skillresidue

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestMoveFileReferencesArePinned parses the package's non-test sources and counts every
// reference to the root-bound steps of the held-back rename and to their seams, by
// enclosing declaration. The registry rows for moveOne and dropOne
// (internal/cli/update_destructive_registry.go) cover the one Rename and the one Remove
// behind them. A new reference to any of these names is a new route to a move or a
// removal: the test fails and forces those rows to be reviewed. The declaration name of a
// function itself is not a reference, so it is not counted.
func TestMoveFileReferencesArePinned(t *testing.T) {
	pinned := map[string]bool{
		"mkdirFile": true, "mkdirOne": true,
		"moveFile": true, "moveOne": true,
		"linkFile": true, "linkOne": true,
		"dropFile": true, "dropOne": true,
	}
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
					if id, ok := n.(*ast.Ident); ok && pinned[id.Name] {
						got[id.Name+"@"+owner]++
					}
					return true
				})
			}
		}
	}

	// The only allowed references. At package level each seam variable is declared once
	// and bound to its function once. renameSkillFile makes the move, the final link and
	// the drops of the staging names; claimStage claims the staging folder; putBack returns
	// the file to its own name.
	want := map[string]int{
		"mkdirFile@package-level": 1, "mkdirOne@package-level": 1,
		"moveFile@package-level": 1, "moveOne@package-level": 1,
		"linkFile@package-level": 1, "linkOne@package-level": 1,
		"dropFile@package-level": 1, "dropOne@package-level": 1,

		"mkdirFile@claimStage":     1,
		"moveFile@renameSkillFile": 1,
		"linkFile@renameSkillFile": 1,
		"dropFile@renameSkillFile": 4,
		"linkFile@putBack":         1,
		"dropFile@putBack":         1,
		"moveFile@putBack":         1,
	}
	for key, n := range want {
		if got[key] != n {
			t.Errorf("%s references = %d, want %d", key, got[key], n)
		}
	}
	for key, n := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("new reference %s (%d): review the moveOne and dropOne registry rows", key, n)
		}
	}
}
