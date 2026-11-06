package hook

// branch_guard_gitc_redirect_t2049_test.go — card t2049 (rc.40 blocker): a
// redirection written directly against the `-C` operand, as in
// `git -C <primary></dev/null commit`, was read as part of the path. The path
// lookup then failed, the failed lookup fell back to the hook cwd (a linked
// worktree), and a commit on the primary checkout's protected branch was
// allowed. The repair ends the operand at a redirection character. The rows run
// on the same real git fixtures as the t2013 table, with the hook cwd at the
// linked worktree.

import "testing"

func TestBranchGuardGitCOperandEndsAtRedirection(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	rows := []struct{ name, command string }{
		{"R1 input redirect glued to operand", "git -C " + P + "</dev/null commit -m x"},
		{"R2 input redirect glued, quoted message", "git -C " + P + "</dev/null commit -m 'message'"},
		{"R3 output redirect glued to operand", "git -C " + P + ">/dev/null commit -m x"},
		{"R4 glued redirect on a second -C", "git -C " + W + " -C " + P + "</dev/null commit -m x"},
		{"R5 glued redirect, switch verb", "git -C " + P + "</dev/null switch other"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, row.command)
		if !denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at a worktree cwd, want deny with the sentinel (the operand names the primary checkout)", row.name, decision, reason, row.command)
		}
	}
}

// TestBranchGuardGitCOperandRedirectionKeepsWorktreeAllowed keeps the repair
// deny-only: the same glued-redirection shape aimed at the linked worktree, or
// carrying no commit verb, is judged where it was before.
func TestBranchGuardGitCOperandRedirectionKeepsWorktreeAllowed(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	for _, command := range []string{
		"git -C " + W + "</dev/null commit -m x",
		"git -C " + W + ">/dev/null commit -m x",
		"git -C " + P + "</dev/null status",
	} {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, command)
		if denies {
			t.Errorf("%q at a worktree cwd = %q / %q, want no BranchGuard deny", command, decision, reason)
		}
	}
}
