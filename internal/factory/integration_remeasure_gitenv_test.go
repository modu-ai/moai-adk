package factory

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// t1589 — the integration-tree cleanliness check reads the repository named by
// an inherited GIT_DIR instead of the tree it was given. A dirty primary then
// reads as clean and the merge step proceeds.
func TestGitIntegrationWorktreeCleanIgnoresInheritedGitEnv(t *testing.T) {
	dirty := t1589FactoryRepo(t)
	if err := os.WriteFile(filepath.Join(dirty, "untracked.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}
	clean := t1589FactoryRepo(t)
	t.Setenv("GIT_DIR", filepath.Join(clean, ".git"))
	t.Setenv("GIT_WORK_TREE", clean)

	ok, status, err := gitIntegrationWorktreeClean(dirty)
	if err != nil {
		t.Fatalf("gitIntegrationWorktreeClean() error = %v", err)
	}
	if ok {
		t.Errorf("gitIntegrationWorktreeClean(dirty) reported clean (status %q); the inherited GIT_DIR read the clean repository", status)
	}
}

// t1589FactoryRepo makes a one-commit repository in a temp directory.
func t1589FactoryRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "init"},
	} {
		full := append([]string{"-C", dir}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}
