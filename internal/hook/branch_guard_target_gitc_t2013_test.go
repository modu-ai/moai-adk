package hook

// branch_guard_target_gitc_t2013_test.go — F1 of the rc.40 final-candidate
// audit: from a hook cwd that is a linked worktree, a `git -C <primary>` command
// whose text is not literal (a quote, `$VAR`, a trailing comment, a redirect)
// had its explicit target dropped, and with the target dropped the hook cwd (the
// worktree) decided, so a commit landed on the primary checkout's protected
// branch. The repair keeps the allow-list (a non-literal command never ADOPTS a
// target to allow) but lets the primary-checkout scan also see `git -C <abs>`,
// so such a command is still judged as primary. The table runs on real git
// fixtures (primary on main, deny_commits_on [main], one linked worktree) with
// the hook cwd at the linked worktree.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gitcGuardVerdict is the BranchGuard's own verdict for one command at one hook
// cwd: branch-state verbs through checkBranchState, commit-creating verbs through
// checkProtectedCommit (the same order the real handler uses).
func gitcGuardVerdict(t *testing.T, fx affGuardFixture, cwd, command string) (denies bool, decision, reason string) {
	t.Helper()
	toolInput, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatal(err)
	}
	in := &HookInput{
		SessionID: "sess-t2013-gitc", HookEventName: "PreToolUse", ToolName: "Bash",
		CWD: cwd, ToolInput: json.RawMessage(toolInput),
	}
	decision, reason = checkBranchState(in, fx.primary)
	if decision == "" {
		decision, reason = checkProtectedCommit(in, fx.primary, []string{"main"})
	}
	return decision == DecisionDeny && strings.HasPrefix(reason, branchGuardViolationPrefix+":"), decision, reason
}

func TestBranchGuardGitCTargetNamingPrimaryIsDeniedFromWorktreeCwd(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	link := filepath.Join(t.TempDir(), "primary-link")
	if err := os.Symlink(fx.primary, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	P, W := fx.primary, fx.release

	rows := []struct{ name, command string }{
		{"A1 unquoted message", "git -C " + P + " commit -m x"},
		{"A2 single-quoted message", "git -C " + P + " commit -m 'message'"},
		{"A3 double-quoted message", "git -C " + P + ` commit -m "message"`},
		{"A4 quoted switch operand", "git -C " + P + " switch 'other'"},
		{"A5 quoted reset operand", "git -C " + P + " reset --hard 'HEAD'"},
		{"A6 quoted cherry-pick operand", "git -C " + P + ` cherry-pick "abc1234"`},
		{"A10 cd worktree then git -C primary", "cd " + W + " && git -C " + P + " commit -m x"},
		{"A16 symlink to primary", "git -C " + link + " commit -m 'message'"},
		{"A17 trailing comment", "git -C " + P + " commit -m x # note"},
		{"A18 redirect", "git -C " + P + " commit -m x > /dev/null"},
		{"A19 variable message", "git -C " + P + " commit -m $MSG"},
		{"X1 second -C names primary", "git -C " + W + " -C " + P + " commit -m 'message'"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, row.command)
		if !denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at a worktree cwd, want deny with the sentinel (the target names the primary checkout)", row.name, decision, reason, row.command)
		}
		handled, handledReason := affObserve(t, handler, W, row.command)
		if handled != DecisionDeny && handled != DecisionAsk {
			t.Errorf("%s: handler = %q / %q for %q at a worktree cwd, want deny or ask", row.name, handled, handledReason, row.command)
		}
	}
}

// TestBranchGuardGitCTargetNamingWorktreeStaysAllowedFromWorktreeCwd keeps the
// repair deny-only and non-vacuous: a non-literal command whose -C names the
// linked worktree (or no -C at all) is still judged where the hook cwd says.
func TestBranchGuardGitCTargetNamingWorktreeStaysAllowedFromWorktreeCwd(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	for _, command := range []string{
		"git -C " + W + " commit -m 'message'",
		"git -C " + W + ` switch "other"`,
		"git -C " + W + " commit -m x # note",
		"git commit -m 'message'",
		"git -C " + W + " status && git -C " + W + " commit -m 'message'",
		`moai todo add "git -C ` + P + ` commit -m x"`,
	} {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, command)
		if denies {
			t.Errorf("%q at a worktree cwd = %q / %q, want no BranchGuard deny (nothing names the primary checkout)", command, decision, reason)
		}
	}
}

// TestBranchGuardGitCTargetPrimaryCwdUnchanged pins F3: from the primary
// checkout, a non-literal `git -C <worktree> commit -m 'msg'` keeps its
// existing (fall-back to the hook cwd) verdict, which is deny.
func TestBranchGuardGitCTargetPrimaryCwdUnchanged(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	W := fx.release

	denies, decision, reason := gitcGuardVerdict(t, fx, fx.primary, "git -C "+W+" commit -m 'msg'")
	if !denies {
		t.Errorf("primary cwd, git -C <worktree> commit -m 'msg' = %q / %q, want the existing deny with the sentinel", decision, reason)
	}
}
