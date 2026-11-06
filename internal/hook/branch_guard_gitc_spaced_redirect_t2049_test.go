package hook

// branch_guard_gitc_spaced_redirect_t2049_test.go — card t2049, second repair:
// a redirection written as its own word between `git` / the `-C` operand and the
// verb (`git -C <primary> </dev/null commit`, `>`, `2>`, `2>&1`, `>>`) ended the
// option run the `-C` judgment reads, so the verb was not recognised behind it
// and a commit on the primary checkout's protected branch was allowed from a
// linked-worktree cwd. The repair removes redirection words at one point of the
// `-C` scan normalization (normalizeForGitCScan). The deny rows and the allow
// pins run on the same real git fixtures as the t2013 table, hook cwd at the
// linked worktree.

import "testing"

func TestBranchGuardGitCSpacedRedirectionStillDeniedForPrimary(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P := fx.primary

	rows := []struct{ name, command string }{
		{"S1 input redirect", "git -C " + P + " </dev/null commit -m x"},
		{"S2 output redirect", "git -C " + P + " >/dev/null commit -m x"},
		{"S3 output redirect, spaced target", "git -C " + P + " > /dev/null commit -m x"},
		{"S4 stderr redirect", "git -C " + P + " 2>/dev/null commit -m x"},
		{"S5 stderr to stdout", "git -C " + P + " 2>&1 commit -m x"},
		{"S6 append redirect", "git -C " + P + " >>/dev/null commit -m x"},
		{"S7 redirect before -C", "git </dev/null -C " + P + " commit -m x"},
		{"S8 two redirects", "git -C " + P + " </dev/null 2>&1 commit -m x"},
		{"S9 quoted message holding < and >", "git -C " + P + " </dev/null commit -m 'a<b>c'"},
		{"S10 switch verb, input redirect", "git -C " + P + " </dev/null switch other"},
		{"S11 switch verb, 2>&1", "git -C " + P + " 2>&1 switch other"},
		{"S12 reset verb, output redirect", "git -C " + P + " >/dev/null reset --hard HEAD"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, fx.release, row.command)
		if !denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at a worktree cwd, want deny with the sentinel (the -C operand names the primary checkout)", row.name, decision, reason, row.command)
		}
	}
}

// TestBranchGuardGitCSpacedRedirectionKeepsAllowedForms keeps the repair
// deny-only: forms that do not land a guarded verb on the primary checkout are
// judged where they were before.
func TestBranchGuardGitCSpacedRedirectionKeepsAllowedForms(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	for _, command := range []string{
		"git </dev/null commit -m x",
		"git -C " + W + " </dev/null commit -m x",
		"git -C " + W + " 2>&1 commit -m x",
		"git -C " + W + " > /dev/null commit -m x",
		"git -C " + P + " </dev/null status",
		"git -C " + P + " 2>&1 log",
		"git -C " + P + " > /dev/null status",
		"git -C " + P + " </dev/null diff",
		"git -C " + W + " commit -m 'a<b>c'",
		"git -C " + W + ` commit -m "a > b"`,
	} {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, command)
		if denies {
			t.Errorf("%q at a worktree cwd = %q / %q, want no BranchGuard deny (nothing lands a guarded verb on the primary checkout)", command, decision, reason)
		}
	}
}
