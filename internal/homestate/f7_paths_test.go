package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The canonical-root key is what ProjectKey computes for the same root, it
// is deterministic, and the human-readable prefix sanitizes characters that
// do not belong in a directory name.
func TestProjectKeyForCanonicalRoot(t *testing.T) {
	root := "/Users/x/proj"
	if got := ProjectKeyForCanonicalRoot(root); got != ProjectKeyForCanonicalRoot(root) {
		t.Fatalf("key not deterministic: %q vs %q", got, ProjectKeyForCanonicalRoot(root))
	}
	if got := ProjectKeyForCanonicalRoot(root); got != ProjectKey(root) {
		t.Fatalf("canonical-root key = %q, ProjectKey = %q, want equivalence", got, ProjectKey(root))
	}
	if got := ProjectKeyForCanonicalRoot("/x/한글이름"); strings.Contains(got, "한글") {
		t.Fatalf("key %q carries unsanitized characters", got)
	}
	if got := ProjectKeyForCanonicalRoot("/"); !strings.HasPrefix(got, "project-") {
		t.Fatalf("key for root \"/\" = %q, want project- prefix", got)
	}
}

// TodoDir and BacklogDBPath hang off the project state directory like the
// factory paths do.
func TestTodoAndBacklogPaths(t *testing.T) {
	root := factorySandbox(t)
	todo, err := TodoDir(root)
	if err != nil || !strings.HasSuffix(todo, "todo") {
		t.Fatalf("todo dir = %q err=%v", todo, err)
	}
	db, err := BacklogDBPath(root)
	if err != nil || db != filepath.Join(todo, "backlog.db") {
		t.Fatalf("backlog db = %q err=%v, want todo/backlog.db", db, err)
	}
}

// ProjectRootFromDBPath reads the manifest adjacent to todo/ and factory/:
// a valid manifest round-trips, and missing, corrupt, or empty manifests are
// each refused.
func TestProjectRootFromDBPath(t *testing.T) {
	root := factorySandbox(t)
	db, err := BacklogDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectRootFromDBPath(db); err == nil {
		t.Fatal("missing manifest: err = nil, want read failure")
	}
	stateDir := filepath.Dir(filepath.Dir(db))
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectRootFromDBPath(db); err == nil {
		t.Fatal("garbage manifest: err = nil, want unmarshal failure")
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte(`{"project_root":""}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectRootFromDBPath(db); err == nil || !strings.Contains(err.Error(), "empty project_root") {
		t.Fatalf("empty manifest: err = %v, want empty project_root failure", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte(`{"project_root":"/the/project"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := ProjectRootFromDBPath(db); err != nil || got != "/the/project" {
		t.Fatalf("valid manifest = %q err=%v, want /the/project", got, err)
	}
}
