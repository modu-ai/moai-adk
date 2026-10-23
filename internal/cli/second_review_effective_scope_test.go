// second_review_effective_scope_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card
// t1632, REQ-MN-019): the second-review record states each backend's
// EFFECTIVE scope — a baseBranch target whose leg carried no review base is
// recorded as "target not applied", never as baseBranch coverage.
package cli

import (
	"context"
	"strings"
	"testing"
)

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

// TestPerformGLMAudit_BaseBranchCarriesReviewBase pins the l57-r7 finding —
// the GLM twin of the codex REQ-MN-019/t1426 chain. The audit_multi GLM leg
// with target=baseBranch resolves its base, measures the diff from that exact
// merge base, and reports the SAME value as review_base. Pre-fix the leg left
// review_base empty, so the second-review record read a baseBranch audit that
// WAS applied as "working-tree (target not applied)" — a false target report.
func TestPerformGLMAudit_BaseBranchCarriesReviewBase(t *testing.T) {
	d := newDeltaRepo(t) // origin/main resolves as the base
	doer := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "ok"})}
	withGLMSeams(t, "test-key", doer)

	t.Run("resolved base names the base it measured", func(t *testing.T) {
		out := performGLMAudit(context.Background(), codexTargetBaseBranch, "", d.root)
		if out.Verdict == VerdictInconclusive {
			t.Fatalf("expected a verdict; got inconclusive: %s", out.Summary)
		}
		if strings.TrimSpace(out.ReviewBase) == "" {
			t.Fatalf("baseBranch GLM leg carried no review_base — the record reads %q for an audit that was applied at the base (l57-r7 finding 2)",
				effectiveScopeFor(codexTargetBaseBranch, out.ReviewBase))
		}
		if scope := effectiveScopeFor(codexTargetBaseBranch, out.ReviewBase); !strings.HasPrefix(scope, "baseBranch (") {
			t.Errorf("effective scope = %q, want a baseBranch scope naming the resolved base", scope)
		}
	})

	t.Run("delta scope names the delta base it measured", func(t *testing.T) {
		out := performGLMAudit(withDeltaBase(context.Background(), d.shaB), codexTargetBaseBranch, "", d.root)
		if out.Verdict == VerdictInconclusive {
			t.Fatalf("expected a verdict; got inconclusive: %s", out.Summary)
		}
		if !strings.Contains(out.ReviewBase, d.shaB) {
			t.Errorf("review_base = %q, want it to name the delta base %s the diff was measured from", out.ReviewBase, d.shaB)
		}
	})
}
