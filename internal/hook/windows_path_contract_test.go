package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A missing project leaf must still share its resolved ancestor with a new
// file. Windows short-name aliases and POSIX symlink aliases have this shape.
func TestFileAccessMissingProjectBelowAlias(t *testing.T) {
	root := t.TempDir()
	realRoot := filepath.Join(root, "real")
	if err := os.Mkdir(realRoot, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(realRoot, alias); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(alias, "new-project")
	h := &preToolHandler{policy: &SecurityPolicy{}, projectDir: project}
	for _, tc := range []struct{ path, want string }{
		{filepath.Join(project, "new.txt"), ""},
		{filepath.Join(realRoot, "outside", "new.txt"), DecisionDeny},
	} {
		raw, err := json.Marshal(map[string]string{"file_path": tc.path})
		if err != nil {
			t.Fatal(err)
		}
		decision, reason := h.checkFileAccess(raw, "Write")
		if decision != tc.want {
			t.Errorf("path %q: decision=%q reason=%q want=%q", tc.path, decision, reason, tc.want)
		}
	}
}

// Both security walks must keep the native volume and canonical ancestor
// when their leaf does not exist. This exercises drive-letter paths on Windows.
func TestSecurityWalkNativeNewLeaf(t *testing.T) {
	root := t.TempDir()
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "missing", "leaf.txt")
	want := filepath.Join(canonical, "missing", "leaf.txt")
	for name, walk := range map[string]func(string) (string, bool){
		"file-access":    resolveThroughExistingParent,
		"protected-zone": zoneResolve,
	} {
		t.Run(name, func(t *testing.T) {
			got, ok := walk(target)
			if !ok || got != want {
				t.Fatalf("walk(%q) = %q, %v; want %q, true", target, got, ok, want)
			}
		})
	}
}
