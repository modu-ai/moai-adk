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
// repos (t.TempDir, so the caller's TMPDIR decides where they live) and pin
// the agentic-trunk contract: the base resolves from local main first,
// falling back to origin/main when local main is absent — never from develop.

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
		"-c", "commit.gpgSign=false",
		"-c", "user.name=env-diff-base-test",
		"-c", "user.email=env-diff-base-test@example.invalid"}, args...)...)
	// The fixture must not depend on the caller's git configuration.
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
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

// envDiffTempRepo creates an isolated throwaway git repo under t.TempDir().
// The helper hard-codes no path: keeping the repo outside $HOME and this
// repository is the caller's TMPDIR discipline.
func envDiffTempRepo(t *testing.T, initialBranch string) string {
	t.Helper()
	repo := t.TempDir()
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

// TestSrlDiffBasePrefersLocalMainOverStaleOriginMain pins the main ->
// origin/main order: local main has advanced past a stale origin/main ref, so
// the base MUST be the merge-base with local main. Resolving origin/main first
// lands on the older commit.
func TestSrlDiffBasePrefersLocalMainOverStaleOriginMain(t *testing.T) {
	repo := envDiffTempRepo(t, "main")
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD") // stale remote ref
	envDiffCommit(t, repo, "advanced.txt", "main moved on\n")
	envDiffGit(t, repo, "checkout", "-q", "-b", "card-branch")
	envDiffCommit(t, repo, "card.txt", "card work\n")

	want := envDiffGit(t, repo, "merge-base", "HEAD", "main")
	stale := envDiffGit(t, repo, "merge-base", "HEAD", "origin/main")
	if want == "" || want == stale {
		t.Fatalf("fixture invalid: main base %q must be non-empty and differ from origin/main base %q", want, stale)
	}
	if got := srlDiffBase(t, repo); got != want {
		t.Fatalf("srlDiffBase picked %s; want merge-base with local main %s, not stale origin/main %s (card t1856)", got, want, stale)
	}
}

// TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain pins that develop never
// sits between main and origin/main: with no local main, a stale develop, and
// an origin/main ahead of develop's fork point, the base MUST be the
// merge-base with origin/main.
func TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain(t *testing.T) {
	repo := envDiffTempRepo(t, "card-branch") // no local main in this repo
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "branch", "develop") // stale develop frozen at the fork point
	envDiffCommit(t, repo, "landed.txt", "landed on origin/main\n")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	envDiffCommit(t, repo, "card.txt", "card work\n")

	want := envDiffGit(t, repo, "merge-base", "HEAD", "origin/main")
	viaDevelop := envDiffGit(t, repo, "merge-base", "HEAD", "develop")
	if want == "" || want == viaDevelop {
		t.Fatalf("fixture invalid: origin/main base %q must be non-empty and differ from develop base %q", want, viaDevelop)
	}
	if got := srlDiffBase(t, repo); got != want {
		t.Fatalf("srlDiffBase picked %s; want merge-base with origin/main %s, not stale develop %s (card t1856)", got, want, viaDevelop)
	}
}

// TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain pins that no develop-family
// ref resolves the base: with neither local main nor origin/main, a stale
// develop and a stale origin/develop are both present, and the resolver MUST
// fall through to its documented HEAD fallback.
func TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain(t *testing.T) {
	repo := envDiffTempRepo(t, "card-branch") // no local main, no origin/main
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "branch", "develop")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/develop", "HEAD")
	envDiffCommit(t, repo, "card.txt", "card work\n")

	for _, ref := range []string{"develop", "origin/develop"} {
		if base := envDiffGit(t, repo, "merge-base", "HEAD", ref); base == "" || base == "HEAD" {
			t.Fatalf("fixture invalid: merge-base with %s is %q, must be a commit that differs from the HEAD fallback", ref, base)
		}
	}
	if got := srlDiffBase(t, repo); got != "HEAD" {
		t.Fatalf("srlDiffBase picked %s; want the HEAD fallback, a develop-family ref must never resolve the base (card t1856)", got)
	}
}

// TestSrlDiffBaseSkipsStaleOriginDevelopBeforeOriginMain pins that
// origin/develop never sits ahead of origin/main: with no local main and a
// stale origin/develop older than origin/main, the base MUST be the
// merge-base with origin/main.
func TestSrlDiffBaseSkipsStaleOriginDevelopBeforeOriginMain(t *testing.T) {
	repo := envDiffTempRepo(t, "card-branch") // no local main in this repo
	envDiffCommit(t, repo, "base.txt", "base\n")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/develop", "HEAD") // stale
	envDiffCommit(t, repo, "landed.txt", "landed on origin/main\n")
	envDiffGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	envDiffCommit(t, repo, "card.txt", "card work\n")

	want := envDiffGit(t, repo, "merge-base", "HEAD", "origin/main")
	viaDevelop := envDiffGit(t, repo, "merge-base", "HEAD", "origin/develop")
	if want == "" || want == viaDevelop {
		t.Fatalf("fixture invalid: origin/main base %q must be non-empty and differ from origin/develop base %q", want, viaDevelop)
	}
	if got := srlDiffBase(t, repo); got != want {
		t.Fatalf("srlDiffBase picked %s; want merge-base with origin/main %s, not stale origin/develop %s (card t1856)", got, want, viaDevelop)
	}
}
