package auditreceipt

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A cwd inside a git repository resolves to the repository toplevel, not the cwd.
func TestTreeRootFromCWD_ResolvesGitToplevel(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v (%s)", err, out)
	}
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	want := Canonical(root)
	if got := TreeRootFromCWD(sub); got != want {
		t.Errorf("TreeRootFromCWD(%q) = %q, want the repository toplevel %q", sub, got, want)
	}
}

// Outside a repository the canonicalized cwd is the tree root, and an empty cwd
// resolves to nothing rather than to the process working directory.
func TestTreeRootFromCWD_FallsBackToCanonicalCWD(t *testing.T) {
	dir := t.TempDir()
	if got, want := TreeRootFromCWD(dir), Canonical(dir); got != want {
		t.Errorf("TreeRootFromCWD(%q) = %q, want %q", dir, got, want)
	}
	if got := TreeRootFromCWD(""); got != "" {
		t.Errorf("TreeRootFromCWD(\"\") = %q, want empty", got)
	}
}
