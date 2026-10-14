// second_review_effective_scope_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card
// t1632, REQ-MN-019): the second-review record states each backend's
// EFFECTIVE scope — a baseBranch target whose leg carried no review base is
// recorded as "target not applied", never as baseBranch coverage.
package cli

import "testing"

func TestEffectiveScopeFor(t *testing.T) {
	cases := []struct {
		name, target, reviewBase, want string
	}{
		{"baseBranch applied", codexTargetBaseBranch, "base X", "baseBranch (base X)"},
		{"baseBranch NOT applied", codexTargetBaseBranch, "", "working-tree (target not applied)"},
		{"ordinary tree review", codexTargetUncommitted, "", "working-tree"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := effectiveScopeFor(tc.target, tc.reviewBase); got != tc.want {
				t.Errorf("effectiveScopeFor(%q, %q) = %q, want %q", tc.target, tc.reviewBase, got, tc.want)
			}
		})
	}
}
