package hook

// branch_guard_aff_integration_t2013_test.go — the BranchGuard contract the
// agentic factory flow (separate lane worktree -> dedicated local integration
// branch -> verification -> release path) depends on, observed against a real
// git repository and the real preToolHandler.
//
// The flow's git work happens in two places. The moai verbs (`moai factory
// complete`, `moai integration merge --card`, `moai factory checkpoint`, `moai
// factory next`) run git as a subprocess of the moai process, never as a Bash
// tool call, so the PreToolUse guard sees only the verb's command line, which
// carries no git verb. Direct git commands an agent types are judged by the
// guard on two facts: whether the command's cwd is the PRIMARY checkout (git
// dir == common dir) and, for commit-creating verbs, whether that checkout's
// HEAD is on a deny_commits_on branch.
//
// Contract, in one fixture (primary checkout on main, plus one real dedicated
// integration worktree on the branch release/main-batch), with the guard on and
// deny_commits_on = [main]:
//
//   - at the dedicated integration worktree the integration work is allowed:
//     merge --no-ff, commit, revert, cherry-pick, and the plumbing commit-tree;
//   - at the shared primary checkout on main the same verbs are denied with the
//     BRANCH_GUARD_VIOLATION sentinel, together with the other ref-moving
//     branch-state verbs the guard has always denied there (switch, checkout,
//     reset --hard, branch -f, rebase, stash);
//   - the moai verbs of the flow are allowed at the primary checkout.
//
// What this test does NOT show, so a green run is not read as full AF/Git-AFF
// compatibility: it does not run any moai verb; it does not exercise
// `git -C <path> ...` (the matchers require the verb right after `git`, so a
// -C form is not classified at all); and ref-moving commands the guard has never
// matched (`git update-ref`, `git push . <sha>:main`, `git fetch . <sha>:main`)
// are outside the existing contract and are not asserted either way.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// affGuardFixture is the real repository both halves of the contract run in.
type affGuardFixture struct {
	primary string // primary checkout, HEAD on main
	release string // dedicated integration worktree, HEAD on release/main-batch
}

func newAFFGuardFixture(t *testing.T) affGuardFixture {
	t.Helper()
	primary := mainRepoFixture(t)
	release := filepath.Join(t.TempDir(), "main-batch")
	mustRunGit(t, primary, "worktree", "add", release, "-b", "release/main-batch")

	// Non-vacuity: the two directories must classify differently, and the
	// branches must be the ones the contract is stated for. Without this a
	// mislocated fixture would turn every allow/deny below into noise.
	isPrimary, err := isPrimaryCheckout(primary)
	if err != nil || !isPrimary {
		t.Fatalf("fixture primary: isPrimaryCheckout = %t, err = %v, want true", isPrimary, err)
	}
	isPrimary, err = isPrimaryCheckout(release)
	if err != nil || isPrimary {
		t.Fatalf("fixture integration worktree: isPrimaryCheckout = %t, err = %v, want false", isPrimary, err)
	}
	if got := currentBranch(t, primary); got != "main" {
		t.Fatalf("fixture primary branch = %q, want main", got)
	}
	if got := currentBranch(t, release); got != "release/main-batch" {
		t.Fatalf("fixture integration worktree branch = %q, want release/main-batch", got)
	}
	return affGuardFixture{primary: primary, release: release}
}

// affObserve sends one Bash command through the real handler as a main-thread
// payload (no AgentType, so the identity exemption cannot apply) with its cwd
// at cwd. The sentinel environment exemption is cleared by the caller.
func affObserve(t *testing.T, handler *preToolHandler, cwd, command string) (decision, reason string) {
	t.Helper()
	toolInput, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	out, err := handler.Handle(context.Background(), &HookInput{
		SessionID:     "sess-t2013-aff-guard",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		CWD:           cwd,
		ToolInput:     json.RawMessage(toolInput),
	})
	if err != nil {
		t.Fatalf("Handle(%q at %q) err = %v", command, cwd, err)
	}
	return decisionOf(out), reasonOf(out)
}

// TestBranchGuardAFFDedicatedIntegrationContract is the headline: dedicated
// integration allowed, shared main still protected. Not parallel: the parent
// calls t.Setenv, which mutates process-global environment.
func TestBranchGuardAFFDedicatedIntegrationContract(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)

	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = true
	cfg.Workflow.BranchGuard.DenyCommitsOn = []string{"main"}
	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: fx.primary,
	}

	t.Run("dedicated integration worktree allows the integration work", func(t *testing.T) {
		for _, command := range []string{
			"git merge --no-ff abc1234",
			`git commit -m "[t2013] forward repair"`,
			"git revert abc1234",
			"git cherry-pick abc1234",
			"git commit-tree abc1234 -p def5678 -m probe",
			"git status && git merge --no-ff abc1234",
		} {
			// The assertion is BranchGuard-specific: another PreToolUse guard (for
			// example the commit identity guard) may judge a fixture commit for its
			// own reasons, and that must not read as a BranchGuard failure.
			_, reason := affObserve(t, handler, fx.release, command)
			if strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("at the dedicated integration worktree %q carries the %s sentinel: %q, want no BranchGuard deny", command, branchGuardViolationPrefix, reason)
			}
		}
	})

	t.Run("shared main still denies the ref-moving and commit-creating verbs", func(t *testing.T) {
		for _, command := range []string{
			"git commit -m x",
			"git commit --amend",
			"git revert HEAD",
			"git cherry-pick abc1234",
			"git merge --no-ff abc1234",
			"git switch release/main-batch",
			"git checkout other-branch",
			"git reset --hard HEAD~1",
			"git branch -f main abc1234",
			"git rebase main",
			"git stash",
			"git status && git commit -m x",
			"git commit-tree abc1234 -m probe && git commit -m x",
		} {
			decision, reason := affObserve(t, handler, fx.primary, command)
			// The dangerous-pattern check runs before BranchGuard and turns
			// `reset --hard` into an ask; that still stops the command, so an
			// ask is accepted there and the sentinel check is skipped.
			if decision == DecisionAsk && strings.Contains(command, "reset --hard") {
				continue
			}
			if decision != DecisionDeny {
				t.Errorf("at the shared primary checkout on main %q = %q, want %q", command, decision, DecisionDeny)
				continue
			}
			if !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("at the shared primary checkout %q denied without the sentinel: %q", command, reason)
			}
		}
	})

	t.Run("the moai verbs of the flow are not git commands and pass at the primary checkout", func(t *testing.T) {
		for _, command := range []string{
			"moai factory complete t2013",
			"moai factory next --card t2013",
			"moai factory checkpoint t2013",
			"moai integration merge --card t2013",
			"moai integration status",
		} {
			_, reason := affObserve(t, handler, fx.primary, command)
			if strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("%q carries the %s sentinel at the primary checkout: %q, want no BranchGuard deny", command, branchGuardViolationPrefix, reason)
			}
		}
	})
}

// TestBranchGuardAFFProtectionIsPerHeadBranch pins the second half of the
// commit-creating deny: it follows the HEAD branch of the command's cwd. A
// primary checkout whose HEAD is NOT on a deny_commits_on branch may commit,
// while the branch-state verbs stay denied there; this is what keeps the deny
// from reaching a dedicated clone or a release branch by name alone.
func TestBranchGuardAFFProtectionIsPerHeadBranch(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	// A second, independent primary checkout (a plain repository, not a worktree
	// of the fixture) whose HEAD is the integration branch name.
	other := mainRepoFixture(t)
	mustRunGit(t, other, "branch", "-m", "release/main-batch")

	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = true
	cfg.Workflow.BranchGuard.DenyCommitsOn = []string{"main"}
	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: fx.primary,
	}

	if _, reason := affObserve(t, handler, other, `git commit -m "[t2013] on a non-protected primary"`); strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
		t.Errorf("git commit at a primary checkout on release/main-batch carries the %s sentinel: %q, want no BranchGuard deny (deny_commits_on is [main])", branchGuardViolationPrefix, reason)
	}
	if decision, _ := affObserve(t, handler, other, "git switch main"); decision != DecisionDeny {
		t.Errorf("git switch at a primary checkout = %q, want %q (branch-state verbs are denied at any primary checkout)", decision, DecisionDeny)
	}
	if decision, _ := affObserve(t, handler, fx.primary, "git commit -m x"); decision != DecisionDeny {
		t.Errorf("git commit at the primary checkout on main = %q, want %q", decision, DecisionDeny)
	}
}
