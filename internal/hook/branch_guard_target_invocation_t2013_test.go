package hook

// branch_guard_target_invocation_t2013_test.go — F8 / F4 / F10 / F12 of the rc.40
// final-candidate audit (4d4bd00c91): the git `-C` target is judged PER
// INVOCATION. The `-C` operands that decide where a guarded verb lands are the
// ones on the invocation that carries the verb, in order (a later absolute path
// replaces the earlier one, a relative path is joined onto the previous one). A
// `-C` that sits on a different invocation, in a quoted mention, in a comment, in
// a heredoc body or in an `echo` argument must not change another invocation's
// verdict. Before the repair, the scan collected every `-C` in the command and
// paired none of them with a verb, so a harmless `git -C <primary> status` next
// to a worktree commit denied the commit (F8), and a verb behind `-C` in a chain
// of two invocations was never seen at all (F4). The tables run on the same real
// git fixture as the other t2013 guard tests (primary on main, deny_commits_on
// [main], one linked worktree) with the hook cwd at the linked worktree unless a
// test says otherwise.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBranchGuardPerInvocationTargetIsAllowedFromWorktreeCwd: every row's verb
// really runs in the linked worktree, so nothing may deny it, whatever a `-C`
// elsewhere in the command text names.
func TestBranchGuardPerInvocationTargetIsAllowedFromWorktreeCwd(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	rows := []struct{ name, command string }{
		{"N11 -C on a read-only invocation before the commit", "git -C " + P + " rev-parse HEAD && git commit -m 'msg'"},
		{"N12 -C status then commit", "git -C " + P + " status --short; git commit -m x"},
		{"N13 -C in a heredoc body", "git commit -F - <<EOF\nnote git -C " + P + " status\nEOF"},
		{"N14 -C in a trailing comment", "git commit -m x # git -C " + P + " status"},
		{"N15 -C fetch then switch", "git -C " + P + " fetch origin && git switch other"},
		{"N16 -C log with a redirect then commit", "git -C " + P + " log -1 > /tmp/x && git commit -m 'msg'"},
		{"N17 -C commit inside a heredoc body, later commit", "cat <<EOF > f\ngit -C " + P + " commit -m x\nEOF\ngit commit -m y"},
		{"N18 -C in an echo argument", "echo git -C " + P + " status && git commit -m x"},
		{"N27 later -C names the worktree", "git -C " + P + " -C " + W + " commit -m 'msg'"},
		{"N31 later -C names the worktree, bare message", "git -C " + P + " -C " + W + " commit -m msg"},
		{"N30 commit then -C status", "git commit -m 'msg' && git -C " + P + " status"},
		{"N32 bare -C in a trailing comment", "git commit -m x # git -C " + P},
		{"Q5 -C inside an escaped-quote message", `git commit -m "see \"git -C ` + P + ` status\" here"`},
		{"Q8 commit then -C log", "git commit --allow-empty -m 'wt' && git -C " + P + " log --oneline -1"},
		{"Q11 switch then -C log", "git switch -c topic && git -C " + P + " log -1"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, row.command)
		if denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at a worktree cwd, want no deny (the verb runs in the worktree)", row.name, decision, reason, row.command)
		}
	}
}

// TestBranchGuardPerInvocationRowsStillDenyAtPrimaryCwd is the non-vacuity pair
// of the allow table: the same commands, run from the primary checkout, land
// their verb there and must still be denied. It proves the allow rows above are
// judged by a guard that sees the verb, not by one that stopped looking.
func TestBranchGuardPerInvocationRowsStillDenyAtPrimaryCwd(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P := fx.primary

	rows := []struct{ name, command string }{
		{"N11", "git -C " + P + " rev-parse HEAD && git commit -m 'msg'"},
		{"N12", "git -C " + P + " status --short; git commit -m x"},
		{"N13", "git commit -F - <<EOF\nnote git -C " + P + " status\nEOF"},
		{"N15", "git -C " + P + " fetch origin && git switch other"},
		{"N16", "git -C " + P + " log -1 > /tmp/x && git commit -m 'msg'"},
		{"N17", "cat <<EOF > f\ngit -C " + P + " commit -m x\nEOF\ngit commit -m y"},
		{"N18", "echo git -C " + P + " status && git commit -m x"},
		{"N30", "git commit -m 'msg' && git -C " + P + " status"},
		{"Q5", `git commit -m "see \"git -C ` + P + ` status\" here"`},
		{"Q8", "git commit --allow-empty -m 'wt' && git -C " + P + " log --oneline -1"},
		{"Q11", "git switch -c topic && git -C " + P + " log -1"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, P, row.command)
		if !denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at the primary cwd, want deny with the sentinel (the verb lands in the primary checkout)", row.name, decision, reason, row.command)
		}
	}
}

// TestBranchGuardPerInvocationTargetNamingPrimaryIsDeniedFromWorktreeCwd: the
// invocation that carries the verb has a `-C` that resolves to the primary
// checkout, so the verb lands there whatever else the command holds.
func TestBranchGuardPerInvocationTargetNamingPrimaryIsDeniedFromWorktreeCwd(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release
	if err := os.MkdirAll(filepath.Join(P, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	relToPrimary, err := filepath.Rel(W, P)
	if err != nil {
		t.Fatal(err)
	}

	rows := []struct{ name, command string }{
		{"X1 later -C names the primary", "git -C " + W + " -C " + P + " commit -m 'message'"},
		{"F4a -C status then -C commit", "git -C " + P + " status && git -C " + P + " commit --allow-empty -m landed"},
		{"F4b -C add then -C commit", "git -C " + P + " add f.txt && git -C " + P + " commit -m 'msg'"},
		{"F4c -C commit then -C push", "git -C " + P + " commit -m 'msg' && git -C " + P + " push"},
		{"F4d -C switch then -C status", "git -C " + P + " switch other && git -C " + P + " status"},
		{"F4e -C reset --hard then -C status", "git -C " + P + " reset --hard && git -C " + P + " status"},
		{"F4f -C commit then a plain status", "git -C " + P + " commit -m 'msg' && git status"},
		{"F10 uncleaned path that resolves to the primary", "git -C " + P + "/sub/.. commit -m 'm'"},
		{"F10b same, bare message", "git -C " + P + "/sub/.. commit -m m"},
		{"relative -C joined onto the previous operand", "git -C " + W + " -C " + relToPrimary + " commit -m x"},
	}
	for _, row := range rows {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, row.command)
		if !denies {
			t.Errorf("%s: BranchGuard alone = %q / %q for %q at a worktree cwd, want deny with the sentinel (the verb lands in the primary checkout)", row.name, decision, reason, row.command)
		}
	}
}

// TestBranchGuardPerInvocationUnrelatedRepositoryChainStaysAllowed: seeing a
// verb through `-C` must not add a deny for a repository that is not the hook
// cwd's own (F9 is a separate, unrepaired item). Q3 was allowed before the
// per-invocation repair and stays allowed.
func TestBranchGuardPerInvocationUnrelatedRepositoryChainStaysAllowed(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	other := mainRepoFixture(t)

	command := "git -C " + other + " add . && git -C " + other + " commit -m 'second'"
	denies, decision, reason := gitcGuardVerdict(t, fx, fx.release, command)
	if denies {
		t.Errorf("Q3: BranchGuard alone = %q / %q for %q at a worktree cwd, want no deny (the target is another repository)", decision, reason, command)
	}
}

// TestBranchGuardPerInvocationQuotedMentionDoesNotChangeVerdict (F12): a `-C`
// that only appears inside a quoted argument is data. The commit runs in the
// worktree, so it is allowed. This row is what makes the quoted-span blanking of
// the per-invocation scan observable: unblank the quoted span and it turns red.
func TestBranchGuardPerInvocationQuotedMentionDoesNotChangeVerdict(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	P, W := fx.primary, fx.release

	for _, command := range []string{
		"git commit -m 'see git -C " + P + " commit -m x'",
		`git commit -m "see git -C ` + P + ` commit -m x"`,
	} {
		denies, decision, reason := gitcGuardVerdict(t, fx, W, command)
		if denies {
			t.Errorf("%q at a worktree cwd = %q / %q, want no deny (a quoted mention is data, the commit runs in the worktree)", command, decision, reason)
		}
	}
}
