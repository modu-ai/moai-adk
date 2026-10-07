package worktree

// done_landing_test.go — the origin-landing machine check on the disposal
// path (SPEC-FACTORY-LANE-AUTONOMY-001 M4, REQ-FLA-012..015 / AC-FLA-012..015).
//
// Every cell runs against a REAL repository under t.TempDir() with a LOCAL
// bare remote standing in as "origin" — the tests never touch the network or
// the project's own checkout. The card tree is created OUTSIDE the repo's
// .claude/.moai worktrees prefixes (the L2 shape), so the L1 tier guard
// passes and the landing check is the precondition under test. Guard-order
// cells reuse the landed L1/anchor fixtures to pin that both existing
// refusals fire exactly as before (AC-FLA-013).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// landingGit runs a git subcommand in dir, failing the test on error.
func landingGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// landingFixture holds one sandbox: a bare remote ("origin"), a working
// repository on develop, and a card worktree on a WT- branch carrying one
// committed-but-unlanded card work commit.
type landingFixture struct {
	base   string // the t.TempDir() root
	origin string // the bare remote
	repo   string // the working repository (develop)
	tree   string // the card worktree (L2 shape, outside L1 prefixes)
	branch string // the card branch, WT- prefixed per the gitflow discipline
}

// newLandingFixture builds origin.git, a repo whose develop is pushed to it,
// and the card worktree with one unlanded commit.
func newLandingFixture(t *testing.T) *landingFixture {
	t.Helper()
	base := t.TempDir()
	f := &landingFixture{
		base:   base,
		origin: filepath.Join(base, "origin.git"),
		repo:   filepath.Join(base, "repo"),
		tree:   filepath.Join(base, "card-wt"),
		branch: "WT-lane-card",
	}
	landingGit(t, base, "init", "-q", "--bare", f.origin)
	// The repo dir must exist before git init: landingGit anchors the git
	// process on cmd.Dir, and chdir into a missing directory fails.
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.repo, "init", "-q", "-b", "develop")
	landingGit(t, f.repo, "config", "user.email", "landing-test@example.com")
	landingGit(t, f.repo, "config", "user.name", "Landing Test")
	landingGit(t, f.repo, "config", "commit.gpgsign", "false")
	landingGit(t, f.repo, "remote", "add", "origin", f.origin)
	if err := os.WriteFile(filepath.Join(f.repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.repo, "add", ".")
	landingGit(t, f.repo, "commit", "-q", "-m", "seed")
	landingGit(t, f.repo, "push", "-q", "-u", "origin", "develop")
	// Card worktree on its own branch (the gitflow card-tree shape), carrying
	// one committed card work commit that neither develop nor origin has.
	landingGit(t, f.repo, "worktree", "add", "-q", "-b", f.branch, f.tree)
	if err := os.WriteFile(filepath.Join(f.tree, "work.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	landingGit(t, f.tree, "add", ".")
	landingGit(t, f.tree, "commit", "-q", "-m", "card work")
	// The landing base derives from the configured integration target (card
	// t1453): seed the git-flow configuration these cells exercise. Written
	// untracked after the commits so rows can replace or remove it.
	installBaseRow(t, f.repo, gitFlowBaseRow())
	// Wire the real WorktreeProvider for repo; the launch-ledger prune is
	// stubbed so removal never touches the developer's ~/.moai state.
	withTierTestEnv(t, f.repo)
	return f
}

// landCard integrates the card branch the way a lane does — a --no-ff merge
// into local develop — and pushes develop to the origin remote.
func (f *landingFixture) landCard(t *testing.T) {
	t.Helper()
	landingGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card", f.branch)
	landingGit(t, f.repo, "push", "-q", "origin", "develop")
}

// assertTreeSurvives asserts the refused tree is still on disk (AC-FLA-015).
func assertTreeSurvives(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("refused tree must survive the attempt, stat %s: %v", path, err)
	}
}

// TestDoneLandingCheck_RefusesUnlandedCardWorktree (AC-FLA-012 + AC-FLA-015):
// while the card's merge commit is not confirmed on origin/develop, done
// refuses on EVERY path — manual and --auto — showing the machine-check
// output, and the tree survives every refused attempt.
func TestDoneLandingCheck_RefusesUnlandedCardWorktree(t *testing.T) {
	t.Run("manual_unpushed_branch", func(t *testing.T) {
		f := newLandingFixture(t)
		err := executeDoneForTierGuard(t, f.branch)
		if err == nil {
			t.Fatal("done must refuse an unlanded card worktree (nil error = removal proceeded)")
		}
		if !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
			t.Errorf("refusal must carry the MERGE_NOT_ON_ORIGIN sentinel, got: %v", err)
		}
		if !strings.Contains(err.Error(), "rev-list --count --left-right") || !strings.Contains(err.Error(), "origin/develop") {
			t.Errorf("refusal must show the machine-check output, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})

	t.Run("manual_local_only_merge", func(t *testing.T) {
		f := newLandingFixture(t)
		// Merged into LOCAL develop but never pushed: still unlanded.
		landingGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "local merge", f.branch)
		err := executeDoneForTierGuard(t, f.branch)
		if err == nil {
			t.Fatal("done must refuse a local-only merge (nil error = removal proceeded)")
		}
		if !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
			t.Errorf("refusal must carry the MERGE_NOT_ON_ORIGIN sentinel, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})

	t.Run("auto_unpushed_branch_refused", func(t *testing.T) {
		// AC-FLA-015: the --auto path is refused exactly like the manual one.
		f := newLandingFixture(t)
		err := executeDoneForTierGuard(t, "--auto", f.branch)
		if err == nil {
			t.Fatal("--auto disposal must refuse an unlanded card worktree (nil error = removal proceeded)")
		}
		if !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
			t.Errorf("refusal must carry the MERGE_NOT_ON_ORIGIN sentinel, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})

	t.Run("fetch_failure_refused_fail_closed", func(t *testing.T) {
		// AC-FLA-012 edge (acceptance §D.2): origin fetch fails -> the
		// landing cannot be confirmed -> refuse, even for fully-merged work.
		f := newLandingFixture(t)
		f.landCard(t)
		landingGit(t, f.repo, "remote", "set-url", "origin", filepath.Join(f.base, "missing.git"))
		err := executeDoneForTierGuard(t, "--auto", f.branch)
		if err == nil {
			t.Fatal("done must refuse when the origin fetch fails (nil error = removal proceeded)")
		}
		if !strings.Contains(err.Error(), "ORIGIN_LANDING_UNCONFIRMED") {
			t.Errorf("fetch-failure refusal must carry the ORIGIN_LANDING_UNCONFIRMED sentinel, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})
}

// TestDoneLandingCheck_LandedCardDisposalCompletes (AC-FLA-014): once the
// machine check confirms the merge commit on origin/develop, disposal
// completes unattended (exit 0 + removal observable) on both paths. The
// CI-status absence is grep-provable (E4), not testable here.
func TestDoneLandingCheck_LandedCardDisposalCompletes(t *testing.T) {
	t.Run("auto", func(t *testing.T) {
		f := newLandingFixture(t)
		f.landCard(t)
		if err := executeDoneForTierGuard(t, "--auto", f.branch); err != nil {
			t.Fatalf("landed card worktree must dispose unattended via --auto, got error: %v", err)
		}
		if _, statErr := os.Stat(f.tree); !os.IsNotExist(statErr) {
			t.Errorf("expected %s removed after --auto disposal, stat error: %v", f.tree, statErr)
		}
	})

	t.Run("manual", func(t *testing.T) {
		f := newLandingFixture(t)
		f.landCard(t)
		if err := executeDoneForTierGuard(t, f.branch); err != nil {
			t.Fatalf("landed card worktree must dispose manually, got error: %v", err)
		}
		if _, statErr := os.Stat(f.tree); !os.IsNotExist(statErr) {
			t.Errorf("expected %s removed after manual disposal, stat error: %v", f.tree, statErr)
		}
	})
}

// TestDoneLandingCheck_GuardsOutrankLandingCheck (AC-FLA-013): the existing
// guards fire exactly as before this SPEC — an unlanded card tree that is
// ALSO an L1 tree gets the L1 refusal, and an unlanded card tree with a live
// anchored session gets the ANCHORED_SESSIONS_PRESENT refusal. The landing
// check never fires before either guard (acceptance §D.2: anchored wins
// over --auto).
func TestDoneLandingCheck_GuardsOutrankLandingCheck(t *testing.T) {
	t.Run("l1_card_tree_refused_by_tier_guard", func(t *testing.T) {
		f := newLandingFixture(t) // card work unlanded
		l1 := filepath.Join(f.repo, ".claude", "worktrees", "card-l1")
		landingGit(t, f.repo, "worktree", "add", "-q", "-b", "WT-card-l1", l1)
		if err := os.WriteFile(filepath.Join(l1, "l1work.txt"), []byte("unlanded\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		landingGit(t, l1, "add", ".")
		landingGit(t, l1, "commit", "-q", "-m", "unlanded l1 work")
		// assertL1Refusal pins the L1 sentinel AND the surviving tree.
		assertL1Refusal(t, executeDoneForTierGuard(t, "WT-card-l1"), l1)
	})

	t.Run("anchored_card_tree_refused_by_anchor_guard", func(t *testing.T) {
		f := newLandingFixture(t) // card work unlanded
		writeTreeRegistry(t, f.tree, []session.Entry{anchoredEntry(t, f.tree, os.Getpid())})
		err := executeDoneForTierGuard(t, f.branch)
		if err == nil {
			t.Fatal("anchor guard must refuse a live-anchored card tree")
		}
		if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
			t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})
}
