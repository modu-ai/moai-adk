package hook

// env_literal_diff_base_test.go — RED reproduction for card t1856:
// srlDiffBase resolved the diff base only through develop → origin/develop.
// origin/develop was deleted (2026-10-17) and local develop is a stale
// branch, so the merge-base landed on the stale fork point and every
// MOAI_KANBAN_* literal landed on main since then read as "added by this
// card" — a false positive (13271 swept added lines, base=966f1b158, on a
// clean main-absorbed tree).
//
// These tests construct the defect's precondition in isolated throwaway git
// repos under /private/tmp (outside $HOME and outside this repository, no
// card id in the path) and pin the agentic-trunk contract: the base resolves
// from local main first, falling back to origin/main when local main is
// absent — never from develop.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// envDiffGit runs a git command in repo, failing the test on error.
func envDiffGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo,
		"-c", "user.name=env-diff-base-test",
		"-c", "user.email=env-diff-base-test@example.invalid"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, repo, err, out)
	}
	return strings.TrimSpace(string(out))
}

// envDiffCommit writes a file and commits it in repo.
func envDiffCommit(t *testing.T, repo, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	envDiffGit(t, repo, "add", name)
	envDiffGit(t, repo, "commit", "-q", "-m", "add "+name)
}

// envDiffTempRepo creates an isolated throwaway git repo under /private/tmp
// (never under $HOME or this repository; no card id in the path).
func envDiffTempRepo(t *testing.T, initialBranch string) string {
	t.Helper()
	repo, err := os.MkdirTemp("/private/tmp", "env-diff-base-")
	if err != nil {
		t.Fatalf("temp repo: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(repo) })
	envDiffGit(t, repo, "init", "-q", "-b", initialBranch)
	return repo
}

// TestSrlDiffBaseResolvesFromMainNotStaleDevelop is the t1856 RED cell: a
// stale develop branch exists, main has advanced past the develop fork point
// carrying a MOAI_KANBAN_* literal, and the card branch starts from main's
// tip. The base the resolver picks MUST be main's merge-base; the develop
// merge-base sweeps main-side history into the card's diff (the false
// positive).
func TestSrlDiffBaseResolvesFromMainNotStaleDevelop(t *testing.T) {
	repo := envDiffTempRepo(t, "main")
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "branch", "develop") // stale develop frozen at the fork point
	// This literal landed on MAIN after the fork point — it is not this
	// card's change and must not be swept.
	envDiffCommit(t, repo, "literal.txt", "MOAI_KANBAN_CARD=stale-side\n")
	envDiffGit(t, repo, "checkout", "-q", "-b", "card-branch")
	envDiffCommit(t, repo, "card.txt", "card work, no env literal\n")

	want := envDiffGit(t, repo, "merge-base", "HEAD", "main")
	if want == "" {
		t.Fatalf("merge-base HEAD main empty in temp repo")
	}
	if got := srlDiffBase(t, repo); got != want {
		t.Fatalf("srlDiffBase picked %s; want merge-base with local main %s (stale-develop false positive, card t1856)", got, want)
	}
}

// TestSrlDiffBaseFallsBackToOriginMain pins the fallback: where local main
// is absent, the base resolves from origin/main.
func TestSrlDiffBaseFallsBackToOriginMain(t *testing.T) {
	repo := envDiffTempRepo(t, "card-branch") // no local main in this repo
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	envDiffCommit(t, repo, "card.txt", "card work\n")

	want := envDiffGit(t, repo, "merge-base", "HEAD", "origin/main")
	if want == "" {
		t.Fatalf("merge-base HEAD origin/main empty in temp repo")
	}
	if got := srlDiffBase(t, repo); got != want {
		t.Fatalf("srlDiffBase picked %s; want merge-base with origin/main %s (fallback, card t1856)", got, want)
	}
}
