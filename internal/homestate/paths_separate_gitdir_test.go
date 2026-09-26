package homestate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// Repositories whose `git worktree list` names a git directory as the main
// worktree (t1221, t1208 audit F2): --separate-git-dir (the metadata dir), bare
// (the bare repository), submodule (.git/modules/<name>). For these,
// CanonicalProjectRoot roots a linked worktree at that git directory. That is
// the key every sibling worktree already shares, so it is kept (lead decision
// D); what must not happen is state being WRITTEN inside the git directory.

func runFixtureGit(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Env = append(cmd.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("fixture: git %v: %v: %s", args, err, out)
	}
}

func resolvedTemp(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp: %v", err)
	}
	return base
}

// separateGitDirFixture: `git init --separate-git-dir meta repo` plus two
// linked worktrees wa and wb.
func separateGitDirFixture(t *testing.T) (repo, meta, wa, wb string) {
	t.Helper()
	base := resolvedTemp(t)
	repo, meta = filepath.Join(base, "repo"), filepath.Join(base, "meta")
	wa, wb = filepath.Join(base, "wa"), filepath.Join(base, "wb")
	runFixtureGit(t, "init", "--initial-branch=main", "--separate-git-dir", meta, repo)
	runFixtureGit(t, "-C", repo, "commit", "--allow-empty", "-m", "seed")
	runFixtureGit(t, "-C", repo, "worktree", "add", wa, "-b", "la")
	runFixtureGit(t, "-C", repo, "worktree", "add", wb, "-b", "lb")
	return repo, meta, wa, wb
}

// bareFixture: a bare repository x.git (cloned from a seeded repo so it has a
// commit) plus two linked worktrees w1 and w2.
func bareFixture(t *testing.T) (bare, w1, w2 string) {
	t.Helper()
	base := resolvedTemp(t)
	seed := filepath.Join(base, "seed")
	bare, w1, w2 = filepath.Join(base, "x.git"), filepath.Join(base, "w1"), filepath.Join(base, "w2")
	runFixtureGit(t, "init", "--initial-branch=main", seed)
	runFixtureGit(t, "-C", seed, "commit", "--allow-empty", "-m", "seed")
	runFixtureGit(t, "clone", "-q", "--bare", seed, bare)
	runFixtureGit(t, "-C", bare, "worktree", "add", w1, "-b", "l1")
	runFixtureGit(t, "-C", bare, "worktree", "add", w2, "-b", "l2")
	return bare, w1, w2
}

// submoduleFixture: a superproject with submodule sub and one linked worktree
// of the submodule. Returns the linked worktree and the submodule's git
// directory (<super>/.git/modules/sub).
func submoduleFixture(t *testing.T) (worktree, gitDir string) {
	t.Helper()
	base := resolvedTemp(t)
	src, super := filepath.Join(base, "subsrc"), filepath.Join(base, "super")
	worktree = filepath.Join(base, "swa")
	runFixtureGit(t, "init", "--initial-branch=main", src)
	runFixtureGit(t, "-C", src, "commit", "--allow-empty", "-m", "seed")
	runFixtureGit(t, "init", "--initial-branch=main", super)
	runFixtureGit(t, "-C", super, "commit", "--allow-empty", "-m", "seed")
	runFixtureGit(t, "-C", super, "-c", "protocol.file.allow=always", "submodule", "add", "-q", src, "sub")
	runFixtureGit(t, "-C", filepath.Join(super, "sub"), "worktree", "add", worktree, "-b", "la")
	return worktree, filepath.Join(super, ".git", "modules", "sub")
}

// Sibling linked worktrees keep sharing one key, and the primary checkout of a
// separate-git-dir repository keeps its own key: no existing key moves.
func TestGitDirRootedWorktreesKeepSharedKey(t *testing.T) {
	t.Run("separate_git_dir", func(t *testing.T) {
		repo, meta, wa, wb := separateGitDirFixture(t)
		if ka, kb := homestate.ProjectKey(wa), homestate.ProjectKey(wb); ka != kb {
			t.Errorf("sibling linked worktrees split: %s -> %q, %s -> %q", wa, ka, wb, kb)
		}
		if got := homestate.CanonicalProjectRoot(wa); got != meta {
			t.Errorf("CanonicalProjectRoot(linked worktree) = %q, want the shared root %q (existing key)", got, meta)
		}
		if got := homestate.CanonicalProjectRoot(repo); got != repo {
			t.Errorf("CanonicalProjectRoot(primary checkout) = %q, want %q", got, repo)
		}
	})
	t.Run("bare", func(t *testing.T) {
		bare, w1, w2 := bareFixture(t)
		if k1, k2 := homestate.ProjectKey(w1), homestate.ProjectKey(w2); k1 != k2 {
			t.Errorf("sibling linked worktrees split: %s -> %q, %s -> %q", w1, k1, w2, k2)
		}
		if got := homestate.CanonicalProjectRoot(w1); got != bare {
			t.Errorf("CanonicalProjectRoot(linked worktree) = %q, want the bare repository %q (existing key)", got, bare)
		}
	})
}

// A linked worktree of a bare repository is not the canonical tree, so a
// migration-class mutation from it is still refused.
func TestRefuseMutationFromBareLinkedWorktree(t *testing.T) {
	_, w1, _ := bareFixture(t)
	if err := homestate.RefuseMutationFromNonCanonicalTree(w1); err == nil {
		t.Fatalf("RefuseMutationFromNonCanonicalTree(%s) = nil, want a refusal", w1)
	}
}

// Under a temp root, state goes to <root>/.moai/db/<key>. When the root is a
// git directory that would write inside git's own metadata; state must go to
// the home layout instead. A plain temp repository keeps the root layout, so
// the temp branch is narrowed, not disabled. Not parallel: t.Setenv.
func TestEnsureProjectLayoutNeverWritesIntoGitDir(t *testing.T) {
	cases := []struct {
		name  string
		build func(t *testing.T) (worktree, gitDir string)
	}{
		{"separate_git_dir", func(t *testing.T) (string, string) {
			_, meta, wa, _ := separateGitDirFixture(t)
			return wa, meta
		}},
		{"bare", func(t *testing.T) (string, string) {
			bare, w1, _ := bareFixture(t)
			return w1, bare
		}},
		// A submodule's git directory always carries core.worktree, under which
		// `rev-parse --is-inside-git-dir` answers false (t1221 re-audit F1).
		{"submodule", func(t *testing.T) (string, string) {
			return submoduleFixture(t)
		}},
		{"separate_git_dir_with_core_worktree", func(t *testing.T) (string, string) {
			repo, meta, wa, _ := separateGitDirFixture(t)
			runFixtureGit(t, "-C", repo, "config", "core.worktree", repo)
			return wa, meta
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			worktree, gitDir := tc.build(t)
			t.Setenv(paths.EnvHome, "")
			t.Setenv("HOME", t.TempDir())
			if err := homestate.EnsureProjectLayout(worktree); err != nil {
				t.Fatalf("EnsureProjectLayout(%s): %v", worktree, err)
			}
			if _, err := os.Stat(filepath.Join(gitDir, ".moai")); err == nil {
				t.Fatalf("EnsureProjectLayout wrote .moai into the git directory %s", gitDir)
			}
			dir, err := homestate.ProjectDir(worktree)
			if err != nil {
				t.Fatalf("ProjectDir: %v", err)
			}
			if _, err := os.Stat(dir); err != nil {
				t.Fatalf("project state dir %s not created: %v", dir, err)
			}
			// The home layout is complete, not just the project dir: all
			// three call sites must agree the root layout does not apply.
			runDir, err := homestate.RunProjectDir(worktree)
			if err != nil {
				t.Fatalf("RunProjectDir: %v", err)
			}
			searchPath, err := homestate.SearchDBPath(worktree)
			if err != nil {
				t.Fatalf("SearchDBPath: %v", err)
			}
			for _, p := range []string{filepath.Join(runDir, "locks"), filepath.Dir(searchPath)} {
				if _, err := os.Stat(p); err != nil {
					t.Fatalf("home layout dir %s not created: %v", p, err)
				}
			}
		})
	}

	t.Run("plain_temp_repo_keeps_root_layout", func(t *testing.T) {
		base := resolvedTemp(t)
		repo := filepath.Join(base, "plain")
		runFixtureGit(t, "init", "--initial-branch=main", repo)
		t.Setenv(paths.EnvHome, "")
		t.Setenv("HOME", t.TempDir())
		if err := homestate.EnsureProjectLayout(repo); err != nil {
			t.Fatalf("EnsureProjectLayout(%s): %v", repo, err)
		}
		if _, err := os.Stat(filepath.Join(repo, ".moai", "db")); err != nil {
			t.Fatalf("plain temp repo lost the root layout: %v", err)
		}
	})
}
