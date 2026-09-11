package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Card t560 — the core.bare flip of GH #1691, measured under the shape that
// actually fires it.
//
// The sibling file's section C left the reporter's defect-1 attribution
// UNVERIFIED: it drove `git init` through a leaked GIT_DIR and observed that
// core.bare did not move on git 2.50.1 (Apple Git-155). That observation was
// correct and its conclusion was too weak, because the leak was modelled with a
// host repository that has NO linked worktree, so GIT_DIR named a plain `.git`.
//
// Isolated on the same git, against a host that HAS one linked worktree:
//
//	GIT_DIR = <host>/.git/worktrees/<wt>       -> host core.bare becomes true
//	GIT_DIR = <host>/.git                      -> unchanged
//	GIT_INDEX_FILE = <host>/.git/worktrees/... -> unchanged
//
// Only the first shape fires. "Did not reproduce" and "the firing shape was not
// measured yet" produce the same output — no change — and section C recorded
// the first when it had observed the second.
//
// The flip is worse than a stray key. `bare = true` lands in <host>/.git/config,
// which every linked worktree shares, so one fixture breaks the whole repository
// and every later git command fails with "fatal: this operation must be run in a
// work tree" — the error the original report opens with.
//
// This file pins the fix against that shape. It passes because runStep scrubs
// (t516); deleting the scrub turns it red AND flips core.bare, which is how the
// guard was shown to be non-vacuous rather than merely present.

// newRepoWithLinkedWorktree creates a repository with one linked worktree and
// returns (repoPath, worktreeGitDir). The second value is the per-worktree
// gitdir — <repo>/.git/worktrees/<name> — which is the shape that fires.
func newRepoWithLinkedWorktree(t *testing.T, base, name string) (string, string) {
	t.Helper()
	repo := newRepo(t, filepath.Join(base, name))
	wtPath := filepath.Join(base, name+"-wt")
	gitIn(t, repo, "worktree", "add", "-q", wtPath, "-b", name+"-br")

	wtGitDir := filepath.Join(repo, ".git", "worktrees", filepath.Base(wtPath))
	if _, err := os.Stat(wtGitDir); err != nil {
		t.Fatalf("fixture broken: per-worktree gitdir %s: %v", wtGitDir, err)
	}
	return repo, wtGitDir
}

// leakWorktreePreCommitGitEnv is leakPreCommitGitEnv for a commit made FROM a
// linked worktree: git exports that worktree's own gitdir, not the shared one.
// This distinction is the whole finding — the sibling helper models a commit
// from the primary checkout and cannot reach the flip.
func leakWorktreePreCommitGitEnv(t *testing.T, worktreeGitDir string) {
	t.Helper()
	t.Setenv("GIT_DIR", worktreeGitDir)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(worktreeGitDir, "index"))
}

// A gate step's child must not re-initialise, or re-configure, the repository
// the caller was committing to — including when the commit was made from a
// linked worktree, which is the only shape under which `git init` rewrites the
// host's shared config.
func TestRunStep_ChildGitInitFromWorktreeDoesNotFlipOuterCoreBare(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host, wtGitDir := newRepoWithLinkedWorktree(t, base, "host")
	stepDir := filepath.Join(base, "step")
	if err := os.MkdirAll(stepDir, 0o755); err != nil {
		t.Fatalf("mkdir step: %v", err)
	}

	// core.bare is read as the repository reports it, not as a remembered
	// literal: an unset key and "false" are both legitimate starting states
	// depending on git version, and hard-coding either would make this assert
	// something the fixture never established.
	bareBefore := gitInAllowFail(host, "config", "--get", "core.bare")

	report := filepath.Join(base, "init-output.txt")
	leakWorktreePreCommitGitEnv(t, wtGitDir)
	t.Setenv(helperGitInitEnv, "1")
	t.Setenv(helperGitEnvReportEnv, report)

	name, args := helperStep("TestHelperGitInit")
	if ok, msg := stepGate(stepDir).runStep(context.Background(), "init", 60*time.Second, name, args...); !ok {
		t.Fatalf("init step failed: %s", msg)
	}

	out, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read child init output: %v", err)
	}
	t.Logf("child `git init` reported:\n%s", out)

	if strings.Contains(string(out), host) {
		t.Errorf("the gate step's child reached the OUTER repository (%s) through the leaked worktree gitdir; its git init reported:\n%s", host, out)
	}

	bareAfter := gitInAllowFail(host, "config", "--get", "core.bare")
	if bareAfter != bareBefore {
		t.Errorf("a gate step's child flipped the OUTER repository's core.bare: %q before, %q after.\n"+
			"This key lives in %s/.git/config, which every linked worktree shares — the whole repository is now unusable.",
			bareBefore, bareAfter, host)
	}
}

// The discriminator itself, asserted rather than described.
//
// Without this, the test above could pass for the wrong reason: if a future git
// stopped flipping core.bare under ANY shape, the guard would go quiet while
// still reading as a live check. This test drives `git init` through each of
// the three leak shapes directly — no production code involved — and fails if
// the shape the guard is built around stops being the one that fires.
//
// It asserts a property of git, deliberately: the guard above is only meaningful
// while that property holds, so the guard's premise is checked next to it rather
// than assumed. A future git that changes this should fail HERE, loudly, instead
// of silently hollowing out the guard.
func TestGitInitFlipsCoreBareOnlyThroughWorktreeGitDir(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host, wtGitDir := newRepoWithLinkedWorktree(t, base, "host")

	cases := []struct {
		name     string
		env      map[string]string
		wantFlip bool
	}{
		{
			name:     "worktree gitdir",
			env:      map[string]string{"GIT_DIR": wtGitDir},
			wantFlip: true,
		},
		{
			name:     "shared gitdir",
			env:      map[string]string{"GIT_DIR": filepath.Join(host, ".git")},
			wantFlip: false,
		},
		{
			name:     "index file only",
			env:      map[string]string{"GIT_INDEX_FILE": filepath.Join(wtGitDir, "index")},
			wantFlip: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each case starts from a known value and restores it, so the cases
			// do not depend on each other's ordering.
			gitIn(t, host, "config", "core.bare", "false")

			fixture := filepath.Join(base, "fixture", tc.name)
			if err := os.MkdirAll(fixture, 0o755); err != nil {
				t.Fatalf("mkdir fixture: %v", err)
			}

			env := cleanGitEnv()
			for k, v := range tc.env {
				env = append(env, k+"="+v)
			}
			runGitInitWithEnv(t, fixture, env)

			got := gitInAllowFail(host, "config", "--get", "core.bare")
			flipped := got == "true"
			if flipped != tc.wantFlip {
				t.Errorf("host core.bare after `git init` with %v: got %q (flipped=%v), want flipped=%v",
					tc.env, got, flipped, tc.wantFlip)
			}
		})
	}
}

// runGitInitWithEnv runs `git init` in dir with exactly env, failing the test on
// a non-zero exit. It is deliberately not gitIn: gitIn scrubs every GIT_*
// variable, which is precisely what these cases must NOT do.
func runGitInitWithEnv(t *testing.T, dir string, env []string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init in %s: %v\n%s", dir, err, out)
	}
}
