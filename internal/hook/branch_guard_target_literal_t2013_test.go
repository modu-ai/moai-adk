package hook

// branch_guard_target_literal_t2013_test.go — sync audit 6 finding K6. The
// explicit git target (`git -C <abs>` or a leading `cd <abs> &&`) is adopted
// only for a LITERAL command: the quote-collapse the scan uses reads `\"` and
// `\'` as quote starts although the shell does not, so a command carrying any
// quote, backslash, expansion character, or command separator cannot be trusted
// to hold exactly one git invocation. Every case runs from the primary checkout
// on main, so an ALLOW here exists only because a target was adopted.

import (
	"strings"
	"testing"
)

// TestBranchGuardExplicitTargetNotAdoptedWhenCommandIsNotLiteral pins K6: the
// escaped-quote pair hides a real command separator and a second commit from the
// quote collapse, and the fail-closed allow-list refuses to adopt the target for
// any command that is not a plain literal.
func TestBranchGuardExplicitTargetNotAdoptedWhenCommandIsNotLiteral(t *testing.T) {
	skipWindowsPOSIXCommandText(t)
	t.Setenv(branchGuardExemptEnv, "")
	fx := newAFFGuardFixture(t)
	handler := newTargetGuardHandler(fx)
	wt, primary := fx.release, fx.primary

	for _, command := range []string{
		// The audit's three observed bypasses: `\"`, `\'`, and a newline variant.
		"cd " + wt + ` && git commit -m \" ; cd ` + primary + ` ; git commit -m \"`,
		"cd " + wt + ` && git commit -m \' ; cd ` + primary + ` ; git commit -m \'`,
		"cd " + wt + " && git commit -m \\\"\ncd " + primary + "\ngit commit -m \\\"",
		// The same shape through `git -C`.
		"git -C " + wt + ` commit -m \" ; git -C ` + primary + ` commit -m \"`,
		// Fail closed beyond the observed spellings: quotes, expansion, and
		// separators all keep the target from being adopted.
		"cd " + wt + ` && git commit -m "msg"`,
		"git -C " + wt + ` commit -m 'msg'`,
		"cd " + wt + " && git commit -m $MSG",
		"cd " + wt + " && git commit -m x # note",
		"git -C " + wt + " commit -m x > /dev/null",
	} {
		command := command
		t.Run(command, func(t *testing.T) {
			decision, reason := affObserve(t, handler, primary, command)
			if decision != DecisionDeny || !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
				t.Errorf("%q from the primary cwd = %q / %q, want deny with the sentinel (not a literal command, so the explicit target must not be adopted)", command, decision, reason)
			}
		})
	}
}
