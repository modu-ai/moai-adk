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
// enclosing declaration. The registry row for dropOne
// (internal/cli/update_destructive_registry.go) covers the one Remove behind it; the
// package has no Rename of its own (the moveFile seam and its row were removed, card t1733
// F2); the no-replace rename and the rmdir of the staging folder are pinned beside it. A new
// reference to any of these names is a new route to a move, a link or a removal: the test
// fails and forces that row to be reviewed. The declaration name of a function itself is not a
// reference, so it is not counted.
func TestMoveFileReferencesArePinned(t *testing.T) {
	pinned := map[string]bool{
		"mkdirFile": true, "mkdirOne": true,
		"moveFile": true, "moveOne": true, // removed seams: any reappearance is a new route to a Rename
		"linkFile": true, "linkOne": true,
		"dropFile": true, "dropOne": true,
		// The no-replace rename and the rmdir behind the staging folder (card t1733 round 4),
		// and since round 5 the per-platform calls themselves: the registry scan in
		// internal/cli does not count a syscall.Rmdir or a platform rename (F19), so this
		// test is where a new reference inside this package is caught.
		"renameNoReplace": true, "noReplaceRenameAt": true, "renameatNoReplace": true,
		"Renameat2": true, "RenameatxNp": true, "MoveFileEx": true,
		"removeDirAt": true, "Unlinkat": true, "Rmdir": true,
		"openMoveEntry": true, "NtCreateFile": true, "NtSetInformationFile": true,
		"FileRenameInformation": true, "FileDispositionInformation": true,
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
	// and bound to its function once. moveNoReplace is the one caller of the no-replace
	// rename and removeStage the one caller of removeDirAt. renameByLinks (the fallback) makes the
	// first link, the SKILL.md link and the restore link, and drops skill.md itself; dropOurs
	// drops every other name (only while it is still ours); claimStage claims the staging
	// folder.
	want := map[string]int{
		"mkdirFile@package-level": 1, "mkdirOne@package-level": 1,
		"linkFile@package-level": 1, "linkOne@package-level": 1,
		"dropFile@package-level": 1, "dropOne@package-level": 1,
		"renameNoReplace@package-level": 1, "noReplaceRenameAt@package-level": 1,

		"mkdirFile@claimStage":          1,
		"renameNoReplace@moveNoReplace": 1,
		"removeDirAt@removeStage":       1,
		// The platform calls, one owner each: the descriptor-relative renames (linux,
		// darwin), the native handle-relative Windows calls, and rmdir elsewhere.
		"renameatNoReplace@noReplaceRenameAt":     1,
		"Renameat2@renameatNoReplace":             1,
		"RenameatxNp@renameatNoReplace":           1,
		"MoveFileEx@noReplaceRenameAt":            0, // path-based Windows rename must not return
		"openMoveEntry@noReplaceRenameAt":         1,
		"openMoveEntry@removeDirAt":               1,
		"NtCreateFile@openMoveEntry":              1,
		"NtSetInformationFile@noReplaceRenameAt":  1,
		"NtSetInformationFile@removeDirAt":        1,
		"FileRenameInformation@noReplaceRenameAt": 1,
		"FileDispositionInformation@removeDirAt":  1,
		"Unlinkat@removeDirAt":                    1,
		"Rmdir@removeDirAt":                       1, // only platforms without a native rename primitive
		"linkFile@renameByLinks":                  3,
		"dropFile@renameByLinks":                  1,
		"dropFile@dropOurs":                       1,
	}
	for key, n := range want {
		if got[key] != n {
			t.Errorf("%s references = %d, want %d", key, got[key], n)
		}
	}
	for key, n := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("new reference %s (%d): review the dropOne registry row", key, n)
		}
	}
}
