package hook

// branch_guard_target_expand_t2013_test.go — sync audit 5 findings K4 and K5.
// The explicit git target (`git -C <abs>` or a leading `cd <abs> &&`) is adopted
// only when the shell and git are certain to agree on the directory it names.
// Every case runs from the primary checkout on main, so any command that is
// ALLOWED here was allowed only because a target was adopted: the cwd judgment
// the fallback uses can deny, never newly allow.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBranchGuardExplicitTargetNotAdoptedWhenPathIsShellExpanded pins K4: a
// target the shell expands or normalizes differently from git is not the
// directory the shell actually enters, and a failed lookup of a literal target
// must not turn a protected commit into an allow.
func TestBranchGuardExplicitTargetNotAdoptedWhenPathIsShellExpanded(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)

	// A symlink inside the primary checkout: the shell's logical `cd` resolves
	// `<primary>/lnk/..` back to the primary checkout, while git follows the link
	// physically to the parent of the worktree, which is no repository.
	link := filepath.Join(fx.primary, "lnk")
	if err := os.Symlink(fx.release, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	for _, command := range []string{
		"cd /${PWD#/} && git commit -m x",
		"git -C /${PWD#/} commit -m x",
		"cd /$PWD && git commit -m x",
		"cd " + link + "/.. && git commit -m x",
		"git -C " + link + "/.. commit -m x",
		"cd /no/such/dir/t2013 && git commit -m x",
	} {
		command := command
		t.Run(command, func(t *testing.T) {
			decision, reason := affObserve(t, handler, fx.primary, command)
			if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("%q from the primary cwd = %q / %q, want deny with the sentinel (the target is expanded or unresolvable, so it must not be adopted)", command, decision, reason)
			}
		})
	}
}

// TestBranchGuardExplicitTargetNotAdoptedWhenCommandWordIsObscured pins K5: a
// leading `cd <abs> &&` is the git target only when the remainder is exactly one
// git invocation. Spellings the shell reads as eval, cd, or dot-source but a
// spelling-based scan does not (backslash, quotes, `$'..'`, `then .`,
// `command .`) leave something else between the cd and the git command.
func TestBranchGuardExplicitTargetNotAdoptedWhenCommandWordIsObscured(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)
	wt, primary := fx.release, fx.primary

	for _, command := range []string{
		"cd " + wt + " && \\eval 'cd " + primary + "' && git commit -m x",
		"cd " + wt + " && \"eval\" 'cd " + primary + "' && git commit -m x",
		"cd " + wt + " && \"cd\" " + primary + " && git commit -m x",
		"cd " + wt + " && c\\d " + primary + " && git commit -m x",
		"cd " + wt + " && $'cd' " + primary + " && git commit -m x",
		"cd " + wt + " && if true; then . ./x.sh; fi && git commit -m x",
		"cd " + wt + " && command . ./x.sh && git commit -m x",
	} {
		command := command
		t.Run(command, func(t *testing.T) {
			decision, reason := affObserve(t, handler, primary, command)
			if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("%q from the primary cwd = %q / %q, want deny with the sentinel (something other than one git invocation follows the cd, so the target must not be adopted)", command, decision, reason)
			}
		})
	}
}
