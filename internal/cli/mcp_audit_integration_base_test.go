package cli

// Card t1426 — codex_audit / glm_audit (target=baseBranch) review the card diff
// against the configured integration base.
//
// A git-flow repository cuts card branches from an integration branch
// (`develop`, named by git_strategy.worktree_base_branch) while the remote
// default head stays `main`. Resolving the base as "remote default head first"
// made a baseBranch audit review everything develop carries that main lacks,
// instead of the card's own change. The resolution order is now: the configured
// integration base when it is set AND resolves, then the remote default head,
// then main — and the audit result names the base it used.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newIntegrationBaseRepo builds: main (also origin/main = origin/HEAD) →
// develop one commit ahead (develop-only.txt) → card branch cut from develop
// with one change (card.txt). configBase, when non-empty, is written as
// git_strategy.worktree_base_branch and committed on main before develop forks.
// Returns the repo and the expected develop merge base.
func newIntegrationBaseRepo(t *testing.T, configBase string) (string, string) {
	t.Helper()
	repo := newReviewTargetRepo(t)
	reviewTargetGit(t, repo, "branch", "-M", "main")
	if configBase != "" {
		writeWorktreeBaseBranchConfig(t, repo, configBase)
		reviewTargetGit(t, repo, "add", ".moai")
		reviewTargetGit(t, repo, "commit", "-m", "config")
	}
	seedRemoteMain(t, repo)

	reviewTargetGit(t, repo, "checkout", "-b", "develop")
	writeBaseFixtureFile(t, repo, "develop-only.txt", "integrated by another card\n")
	reviewTargetGit(t, repo, "add", "develop-only.txt")
	reviewTargetGit(t, repo, "commit", "-m", "develop advances past main")
	developTip := reviewTargetGitOut(t, repo, "rev-parse", "HEAD")

	reviewTargetGit(t, repo, "checkout", "-b", "WT-card-fixture")
	writeBaseFixtureFile(t, repo, "card.txt", "the card's own change\n")
	reviewTargetGit(t, repo, "add", "card.txt")
	reviewTargetGit(t, repo, "commit", "-m", "card change")
	return repo, developTip
}

func writeBaseFixtureFile(t *testing.T, repo, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestAuditBaseBranch_ConfiguredIntegrationBaseWins(t *testing.T) {
	repo, developTip := newIntegrationBaseRepo(t, "develop")

	t.Run("merge_base_is_develop", func(t *testing.T) {
		got, err := resolveReviewMergeBase(repo)
		if err != nil {
			t.Fatalf("resolveReviewMergeBase: %v", err)
		}
		if got != developTip {
			t.Errorf("merge base = %s, want develop merge base %s (not main's)", got, developTip)
		}
	})

	t.Run("branch_name_is_develop", func(t *testing.T) {
		got, err := resolveReviewBaseBranchName(repo)
		if err != nil {
			t.Fatalf("resolveReviewBaseBranchName: %v", err)
		}
		if got != "develop" {
			t.Errorf("base branch = %q, want %q", got, "develop")
		}
	})

	t.Run("glm_material_is_card_diff_only", func(t *testing.T) {
		diff, err := collectReviewDiff(repo, codexTargetBaseBranch)
		if err != nil {
			t.Fatalf("collectReviewDiff: %v", err)
		}
		if !strings.Contains(diff, "card.txt") {
			t.Errorf("diff must carry the card change; got:\n%s", diff)
		}
		if strings.Contains(diff, "develop-only.txt") {
			t.Errorf("diff must not carry develop's integrated work; got:\n%s", diff)
		}
	})

	t.Run("codex_audit_sends_and_reports_develop", func(t *testing.T) {
		sess, res := runNativeAudit(t, repo, codexTargetBaseBranch)
		if got := sentTargetBranch(t, sess); got != "develop" {
			t.Errorf("target.branch = %q, want %q", got, "develop")
		}
		base := reviewOutputField(res, "review_base")
		if !strings.Contains(base, "develop") || !strings.Contains(base, developTip) {
			t.Errorf("review_base = %q, want it to name develop and merge base %s", base, developTip)
		}
	})
}

// No worktree_base_branch ⇒ behavior unchanged: the remote default head (main).
func TestAuditBaseBranch_NoConfigKeepsRemoteDefaultHead(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "")
	mainSHA := reviewTargetGitOut(t, repo, "rev-parse", "main")

	got, err := resolveReviewMergeBase(repo)
	if err != nil {
		t.Fatalf("resolveReviewMergeBase: %v", err)
	}
	if got != mainSHA {
		t.Errorf("merge base = %s, want main %s", got, mainSHA)
	}
	if name, _ := resolveReviewBaseBranchName(repo); name != "main" {
		t.Errorf("base branch = %q, want %q", name, "main")
	}
}

// A configured base that does not resolve in this tree falls through to the
// remote default head rather than failing the audit.
func TestAuditBaseBranch_UnresolvableConfigFallsThrough(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "no-such-branch")
	mainSHA := reviewTargetGitOut(t, repo, "rev-parse", "main")

	if got, _ := resolveReviewMergeBase(repo); got != mainSHA {
		t.Errorf("merge base = %s, want main %s", got, mainSHA)
	}
	if name, _ := resolveReviewBaseBranchName(repo); name != "main" {
		t.Errorf("base branch = %q, want %q", name, "main")
	}
}
