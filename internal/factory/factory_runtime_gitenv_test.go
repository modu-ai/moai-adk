package factory

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// t1589 — the completion provenance records the commit of the project it was
// asked about, not the commit of the repository an inherited GIT_DIR names.
func TestCaptureFactoryProvenanceIgnoresInheritedGitEnv(t *testing.T) {
	root := t1589FactoryRepo(t)
	wantCommit := t1589FactoryGit(t, root, "rev-parse", "HEAD")

	other := t1589FactoryRepo(t)
	t1589FactoryGit(t, other, "-c", "user.name=test", "-c", "user.email=test@example.invalid",
		"commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "second")

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got := captureFactoryProvenance(root, "")
	if got.GitCommit != wantCommit {
		t.Errorf("captureFactoryProvenance(root).GitCommit = %q, want the project's commit %q (the inherited GIT_DIR named %q)", got.GitCommit, wantCommit, other)
	}
}

// t1589FactoryGit runs one git command in dir and returns its trimmed stdout.
func t1589FactoryGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}
