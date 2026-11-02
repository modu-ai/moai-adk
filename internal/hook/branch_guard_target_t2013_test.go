package hook

// branch_guard_target_t2013_test.go — the BranchGuard judges a git command at
// the path the command itself names (`git -C <abs path>` or a leading
// `cd <abs path> &&`), not only at the hook input cwd. Real git fixtures and
// the real preToolHandler; the hook cwd is always the primary checkout, so a
// verdict that follows the explicit target is distinguishable from one that
// follows the cwd.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func newTargetGuardHandler(fx affGuardFixture) *preToolHandler {
	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = true
	cfg.Workflow.BranchGuard.DenyCommitsOn = []string{"main"}
	return &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: fx.primary,
	}
}

func TestBranchGuardExplicitTargetDeniesPrimary(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	// Hook cwd is the integration worktree (not primary): only the explicit
	// target can make these primary-checkout violations.
	for _, command := range []string{
		"git -C " + fx.primary + " commit -m x",
		"cd " + fx.primary + " && git commit -m x",
		"git -C " + fx.primary + " switch other",
	} {
		decision, reason := affObserve(t, handler, fx.release, command)
		if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Errorf("%q from a worktree cwd = %q / %q, want deny with the sentinel (target is the primary checkout)", command, decision, reason)
		}
	}
}

func TestBranchGuardExplicitTargetAllowsWorktree(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	// Hook cwd is the primary checkout on main.
	for _, command := range []string{
		"git -C " + fx.release + " commit -m x",
		"cd " + fx.release + " && git commit -m x",
		"git -C " + fx.release + " switch other",
	} {
		_, reason := affObserve(t, handler, fx.primary, command)
		if strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Errorf("%q with primary cwd carries the sentinel: %q, want no BranchGuard deny (target is a worktree)", command, reason)
		}
	}
}

func TestBranchGuardExplicitTargetUnchangedForeignForms(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	// Primary cwd on main: real commits stay denied; data-only and relative or
	// unresolvable targets get no new allow.
	for _, command := range []string{
		"git commit -m x",
		"git revert HEAD",
		"git cherry-pick abc1234",
		`git commit -m "git -C ` + fx.release + `"`,
		"git -C " + fx.release + " status && git commit -m x",
	} {
		decision, reason := affObserve(t, handler, fx.primary, command)
		if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Errorf("%q = %q / %q, want deny with the sentinel", command, decision, reason)
		}
	}

	// Quoted data is not a target: the command is a foreign one, still allowed.
	for _, command := range []string{
		`moai todo add "cd ` + fx.release + ` && git commit"`,
	} {
		_, reason := affObserve(t, handler, fx.primary, command)
		if strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Errorf("%q carries the sentinel: %q, want no BranchGuard deny", command, reason)
		}
	}

	// Relative and unresolvable -C paths fall back to today's behavior (the
	// verb is not right after `git`, so no pattern matches: allow, no new deny
	// and no new allow beyond that).
	for _, command := range []string{
		"git -C ../rel commit -m x",
		"git -C /no/such/dir/t2013 commit -m x",
	} {
		_, reason := affObserve(t, handler, fx.primary, command)
		if strings.HasPrefix(reason, branchGuardViolationPrefix+":") && strings.Contains(command, "../rel") {
			t.Errorf("%q carries the sentinel: %q, want today's behavior (no match)", command, reason)
		}
	}
}

// TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects pins the
// can-only-deny promise of the fallback: a leading `cd <abs> &&` is the git
// target only when nothing after it can change the directory or the
// repository. From the primary checkout on main, every redirected form below
// must stay denied (the hook-cwd judgment), as it was before the explicit
// target judgment existed.
func TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	for _, command := range []string{
		"cd " + fx.release + " && cd " + fx.primary + " && git commit -m x",
		"cd " + fx.release + " && GIT_DIR=" + fx.primary + "/.git git commit -m x",
		"cd " + fx.release + " && GIT_WORK_TREE=" + fx.primary + " git commit -m x",
		"cd " + fx.release + " && pushd " + fx.primary + " && git commit -m x",
		"cd " + fx.release + " && (cd " + fx.primary + " && git commit -m x)",
		"cd " + fx.release + " && git --git-dir=" + fx.primary + "/.git commit -m x",
		"cd " + fx.release + " && git --work-tree=" + fx.primary + " commit -m x",
		"git -C " + fx.release + " -C " + fx.primary + " commit -m x",
	} {
		decision, reason := affObserve(t, handler, fx.primary, command)
		if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Errorf("%q from the primary cwd = %q / %q, want deny with the sentinel (a later segment redirects, so the explicit target must not be adopted)", command, decision, reason)
		}
	}
}
