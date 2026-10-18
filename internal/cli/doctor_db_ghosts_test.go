// doctor_db_ghosts_test.go — SPEC-MOAI-DB-PATHS-001 (card t1827) M5: the
// "DB Path Ghosts" doctor check. Fixtures live under t.TempDir() with an
// absolute MOAI_HOME inside it; nothing reaches the operator's home.
package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// ghostFixture builds the committed scratch project plus the eight observed
// guessed paths (spec § A.1) and returns the root, the home, and those paths.
func ghostFixture(t *testing.T) (root, home string, ghosts []string) {
	t.Helper()
	root, home = pathsHomeFixture(t)
	projDir, err := homestate.ProjectDir(root)
	if err != nil {
		t.Fatal(err)
	}
	ghosts = []string{
		filepath.Join(home, "todo.db"),
		filepath.Join(home, "backlog.db"),
		filepath.Join(home, "db", "todo.db"),
		filepath.Join(home, "db", "factory.db"),
		filepath.Join(projDir, "backlog.db"),
		filepath.Join(root, ".moai", "state", "backlog.db"),
		filepath.Join(root, ".moai", "state", "factory.db"),
		filepath.Join(root, ".moai", "state", "factory", "factory.db"),
	}
	for _, g := range ghosts {
		writeSized(t, g, 0)
	}
	return root, home, ghosts
}

func writeSized(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", size)), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDBPathGhosts_EightObserved(t *testing.T) {
	root, _, ghosts := ghostFixture(t)
	check := checkDBPathGhosts(root, true)
	if check.Name != dbPathGhostCheckName {
		t.Errorf("name = %q, want %q", check.Name, dbPathGhostCheckName)
	}
	if check.Status != uikit.CheckWarn {
		t.Fatalf("status = %v, want warning (message %q)", check.Status, check.Message)
	}
	for _, g := range ghosts {
		if !strings.Contains(check.Detail, g) {
			t.Errorf("the detail does not name the ghost %s:\n%s", g, check.Detail)
		}
	}
	// A clean project passes.
	clean, _ := pathsHomeFixture(t)
	if c := checkDBPathGhosts(clean, true); c.Status != uikit.CheckOK {
		t.Errorf("a project without ghosts: status %v (%s)", c.Status, c.Message)
	}
}

func TestDBPathGhosts_CanonicalExcluded(t *testing.T) {
	root, _ := pathsHomeFixture(t)
	queue := factory.QueueLocationForRoot(factory.ResolveTodoQueueRoot(root)).DBPath
	factoryDB, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	writeSized(t, queue, 0)
	writeSized(t, factoryDB, 0)
	if c := checkDBPathGhosts(root, true); c.Status != uikit.CheckOK {
		t.Errorf("zero-byte canonical databases must not be reported: %v %s\n%s", c.Status, c.Message, c.Detail)
	}
}

func TestDBPathGhosts_NonEmptyIgnored(t *testing.T) {
	root, _, ghosts := ghostFixture(t)
	for _, g := range ghosts {
		writeSized(t, g, 3)
	}
	if c := checkDBPathGhosts(root, true); c.Status != uikit.CheckOK {
		t.Errorf("non-empty files at guessed paths must not be reported: %v\n%s", c.Status, c.Detail)
	}
}

func TestDBPathGhosts_UnreadableNotFail(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permission bits are not enforced here")
	}
	root, home, _ := ghostFixture(t)
	locked := filepath.Join(home, "db")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
	c := checkDBPathGhosts(root, true)
	if c.Status == uikit.CheckFail {
		t.Errorf("an unreadable directory must not fail the check: %s", c.Message)
	}
	if c.Status != uikit.CheckWarn {
		t.Errorf("the readable directories still hold ghosts, status = %v", c.Status)
	}
}

func TestDBPathGhosts_ReadOnly(t *testing.T) {
	root, home, _ := ghostFixture(t)
	before := pathsTreeSnapshot(t, root, home)
	_ = checkDBPathGhosts(root, true)
	if diff := snapshotDiff(before, pathsTreeSnapshot(t, root, home)); len(diff) > 0 {
		t.Errorf("the check changed the filesystem: %v", diff)
	}
}
