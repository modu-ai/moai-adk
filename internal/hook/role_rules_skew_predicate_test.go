// role_rules_skew_predicate_test.go — SPEC-ROLE-INJECTION-BUDGET-001
// AC-RIB-005 (REQ-RIB-007): the pure version-skew predicate over fixtures.
// Held out of the M1 commit (package must stay compilable across
// milestones); its compile-failure RED is observed at M1 and the file
// lands together with the M3 implementation (plan M1 item 3 sanctioned form).
package hook

import "testing"

// TestRoleRulesVersionSkewPredicate drives the pure skew predicate over
// fixtures in three cases: equal stamps (no skew), differing stamps (skew,
// both versions named), and an unreadable stamp (fail-open silent skip —
// REQ-RIB-007 fixture-only, no live-repo self-assertion).
func TestRoleRulesVersionSkewPredicate(t *testing.T) {
	t.Run("equal stamps are not skewed", func(t *testing.T) {
		skewed, _, ok := detectRoleRuleVersionSkew("template_version: v3.2.0\n", "v3.2.0")
		if !ok {
			t.Fatal("equal readable stamps must resolve (ok=true)")
		}
		if skewed {
			t.Fatal("equal stamps must not read as skewed")
		}
	})
	t.Run("differing stamps are skewed", func(t *testing.T) {
		skewed, ruleVersion, ok := detectRoleRuleVersionSkew("template_version: v3.1.3\n", "v3.2.0-rc.29")
		if !ok {
			t.Fatal("both stamps readable must resolve (ok=true)")
		}
		if !skewed {
			t.Fatal("differing stamps must read as skewed")
		}
		if ruleVersion != "v3.1.3" {
			t.Errorf("rule-side version must be reported for the detail, got %q", ruleVersion)
		}
	})
	t.Run("unreadable stamp skips silently", func(t *testing.T) {
		skewed, _, ok := detectRoleRuleVersionSkew("not: a version file\n", "v3.2.0")
		if ok {
			t.Fatal("unreadable rule-side stamp must fail open (ok=false, no warning)")
		}
		if skewed {
			t.Fatal("unreadable stamp must not read as skewed")
		}
	})
}
