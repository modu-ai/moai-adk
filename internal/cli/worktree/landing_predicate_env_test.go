package worktree

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRunLandingGitIgnoresCallerRepoScopingEnv (card t2009, S11): the landing
// predicate asks git about a named directory. GIT_DIR and GIT_WORK_TREE outrank
// `git -C`, so a caller whose process carries another repository's location
// would get that repository's answer. runLandingGit must run git with the
// gitenv-scrubbed environment.
func TestRunLandingGitIgnoresCallerRepoScopingEnv(t *testing.T) {
	target := initGuardTestRepo(t)
	other := initGuardTestRepo(t)

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	out, exit, err := runLandingGit(target, "", "rev-parse", "--show-toplevel")
	if err != nil || exit != 0 {
		t.Fatalf("rev-parse under foreign GIT_DIR: exit=%d err=%v", exit, err)
	}
	got, err := filepath.EvalSymlinks(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("resolve answer %q: %v", out, err)
	}
	want, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	if got != want {
		t.Errorf("S11_WRONG_REPO: runLandingGit(%s) answered about %s, want the target repository", want, got)
	}
}
