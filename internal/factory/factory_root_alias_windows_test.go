package factory

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func factoryRootAlias(t *testing.T, target string) string {
	t.Helper()
	alias := filepath.Join(t.TempDir(), "project-alias")
	// A directory junction exercises native alias traversal without requiring
	// a symbolic-link privilege or changing the runner's Developer Mode.
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput(); err != nil {
		t.Fatalf("create native directory junction: %v\n%s", err, out)
	}
	left, err := os.Stat(alias)
	if err != nil {
		t.Fatal(err)
	}
	right, err := os.Stat(target)
	if err != nil || !os.SameFile(left, right) {
		t.Fatalf("junction does not identify the original project directory: %v", err)
	}
	return alias
}
