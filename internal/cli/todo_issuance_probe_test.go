package cli

// todo_issuance_probe_test.go — SPEC-GITHUB-FLOW-CI-RESIDUE-001 M1,
// REQ-GFC-017 / AC-GFC-017. The issuance presentation's lane-files probe ran
// `git merge-base develop <branch>` against a LITERAL develop: in a main-only
// repository the merge-base fails and the probe reports unmeasured, and in a
// stale-develop repository it measures against the wrong fork point. The
// merge-base operand must be the project's resolved integration base — the
// same configured-first chain the landing surfaces use (REQ-GFC-001).
//
// The fixture is a REAL main-only scratch repository (no develop branch, no
// origin/HEAD) with a card worktree at the .moai/worktrees/<card-id>
// convention carrying one committed change.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func probeGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestProductionLaneFilesProbeFollowsConfiguredBase is AC-GFC-017: in a
// main-based repository the probe measures the card's changed files against
// the resolved base — the literal develop is gone.
func TestProductionLaneFilesProbeFollowsConfiguredBase(t *testing.T) {
	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	repo := filepath.Join(base, "repo")
	tree := filepath.Join(repo, ".moai", "worktrees", "probe-card")
	probeGit(t, base, "init", "-q", "--bare", origin)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	probeGit(t, repo, "init", "-q", "-b", "main")
	probeGit(t, repo, "config", "user.email", "probe-test@example.com")
	probeGit(t, repo, "config", "user.name", "Probe Test")
	probeGit(t, repo, "config", "commit.gpgsign", "false")
	probeGit(t, repo, "remote", "add", "origin", origin)
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	probeGit(t, repo, "add", ".")
	probeGit(t, repo, "commit", "-q", "-m", "seed")
	probeGit(t, repo, "push", "-q", "-u", "origin", "main")
	// The card worktree (the .moai/worktrees/<card-id> convention) with one
	// committed change.
	probeGit(t, repo, "worktree", "add", "-q", "-b", "WT-probe-card", tree)
	if err := os.WriteFile(filepath.Join(tree, "changed.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	probeGit(t, tree, "add", ".")
	probeGit(t, tree, "commit", "-q", "-m", "card work")
	// No origin/HEAD (push -u does not create it) and no git-strategy.yaml:
	// the chain falls to its default — which, unlike the literal develop,
	// EXISTS in this repository.
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	files, ok := productionLaneFilesProbe("probe-card", "lane-1")
	if !ok {
		t.Fatal("AC-GFC-017: the probe must measure against the resolved base in a main-only repository, got unmeasured")
	}
	found := false
	for _, f := range files {
		if f == "changed.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("changed.txt must appear in the measured files, got %v", files)
	}
}

// TestProductionLaneFilesProbeConfiguredBaseWins pins the level-1 key: with
// worktree_base_branch configured, the probe measures against THAT branch,
// not against the flow's develop.
func TestProductionLaneFilesProbeConfiguredBaseWins(t *testing.T) {
	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	repo := filepath.Join(base, "repo")
	tree := filepath.Join(repo, ".moai", "worktrees", "probe-card2")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	probeGit(t, base, "init", "-q", "--bare", origin)
	probeGit(t, repo, "init", "-q", "-b", "main")
	probeGit(t, repo, "config", "user.email", "probe-test@example.com")
	probeGit(t, repo, "config", "user.name", "Probe Test")
	probeGit(t, repo, "config", "commit.gpgsign", "false")
	probeGit(t, repo, "remote", "add", "origin", origin)
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	probeGit(t, repo, "add", ".")
	probeGit(t, repo, "commit", "-q", "-m", "seed")
	// The chain answers origin/<configured base>, so the remote-tracking ref
	// must exist for the level-1 answer to be measurable — same fixture step
	// as the sibling test above.
	probeGit(t, repo, "push", "-q", "-u", "origin", "main")
	// develop exists but is UNRELATED to the card's fork point (an orphan
	// commit on a fresh branch): merge-base with it would measure nothing.
	probeGit(t, repo, "checkout", "-q", "--orphan", "develop")
	probeGit(t, repo, "commit", "-q", "-m", "orphan develop", "seed.txt")
	probeGit(t, repo, "checkout", "-q", "main")
	probeGit(t, repo, "worktree", "add", "-q", "-b", "WT-probe-card2", tree)
	if err := os.WriteFile(filepath.Join(tree, "changed.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	probeGit(t, tree, "add", ".")
	probeGit(t, tree, "commit", "-q", "-m", "card work")
	sections := filepath.Join(repo, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n    mode: manual\n    worktree_base_branch: main\n"
	if err := os.WriteFile(filepath.Join(sections, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	files, ok := productionLaneFilesProbe("probe-card2", "lane-1")
	if !ok {
		t.Fatal("the configured worktree_base_branch must give the probe a measurable base")
	}
	found := false
	for _, f := range files {
		if f == "changed.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("changed.txt must appear in the measured files, got %v", files)
	}
}
