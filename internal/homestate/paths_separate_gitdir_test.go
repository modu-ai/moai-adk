package homestate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// separateGitDirFixture builds `git init --separate-git-dir meta repo` plus one
// linked worktree wt, all under t.TempDir(), and returns their resolved paths.
// In this shape `git worktree list` names meta — the metadata directory — as
// the main worktree and nothing records where repo is, so a linked worktree
// cannot be traced back to the primary checkout (t1221, t1208 audit F2).
func separateGitDirFixture(t *testing.T) (repo, meta, wt string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp: %v", err)
	}
	repo, meta, wt = filepath.Join(base, "repo"), filepath.Join(base, "meta"), filepath.Join(base, "wt")
	for _, args := range [][]string{
		{"init", "--initial-branch=main", "--separate-git-dir", meta, repo},
		{"-C", repo, "commit", "--allow-empty", "-m", "seed"},
		{"-C", repo, "worktree", "add", wt, "-b", "linked"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Env = append(cmd.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("separate-git-dir fixture: git %v: %v: %s", args, err, out)
		}
	}
	return repo, meta, wt
}

// TestCanonicalProjectRootSeparateGitDirLinkedWorktree pins lead decision A for
// t1221: a linked worktree of a separate-git-dir repository resolves to its own
// checkout, never to the metadata directory git names as the main worktree.
// The primary checkout keeps resolving to itself, so its existing key is
// unchanged. The two still do NOT converge on one key — git keeps no record
// that would allow it — which is a documented limitation, not a regression.
func TestCanonicalProjectRootSeparateGitDirLinkedWorktree(t *testing.T) {
	repo, meta, wt := separateGitDirFixture(t)

	if got := homestate.CanonicalProjectRoot(wt); got != wt {
		t.Errorf("CanonicalProjectRoot(linked worktree) = %q, want its own checkout %q (meta dir is %q)", got, wt, meta)
	}
	if got := homestate.CanonicalProjectRoot(repo); got != repo {
		t.Errorf("CanonicalProjectRoot(primary checkout) = %q, want %q (existing key must not move)", got, repo)
	}
}

// TestEnsureProjectLayoutSeparateGitDirDoesNotWriteIntoMetadata shows the harm
// the wrong root caused: under a temp root, ProjectDir places state at
// <root>/.moai/db/<key>, so a root of meta wrote .moai into git's own metadata
// directory. Not parallel: t.Setenv.
func TestEnsureProjectLayoutSeparateGitDirDoesNotWriteIntoMetadata(t *testing.T) {
	_, meta, wt := separateGitDirFixture(t)
	t.Setenv(paths.EnvHome, "")
	t.Setenv("HOME", t.TempDir())

	if err := homestate.EnsureProjectLayout(wt); err != nil {
		t.Fatalf("EnsureProjectLayout(linked worktree): %v", err)
	}
	if _, err := os.Stat(filepath.Join(meta, ".moai")); err == nil {
		t.Fatalf("EnsureProjectLayout wrote .moai into the git metadata directory %s", meta)
	}
	if _, err := os.Stat(filepath.Join(wt, ".moai", "db")); err != nil {
		t.Fatalf("expected project state under the linked worktree %s: %v", wt, err)
	}
}
