package config

import "testing"

// TestNormalizeCodexReviewGateTreeScope pins REQ-CRO-001's value rule
// (SPEC-CODEX-REVIEW-OWNERSHIP-001): only "skip", ignoring case and surrounding
// whitespace, reads as skip; every other value — empty, unknown, a prefix or
// suffix variant — reads as review, so a mistyped value never silently reviews
// less.
func TestNormalizeCodexReviewGateTreeScope(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"skip", CodexReviewGateTreeScopeSkip},
		{" Skip ", CodexReviewGateTreeScopeSkip},
		{"SKIP", CodexReviewGateTreeScopeSkip},
		{"review", CodexReviewGateTreeScopeReview},
		{"", CodexReviewGateTreeScopeReview},
		{"never", CodexReviewGateTreeScopeReview},
		{"skipx", CodexReviewGateTreeScopeReview},
		{"no-skip", CodexReviewGateTreeScopeReview},
		{"sk ip", CodexReviewGateTreeScopeReview},
	} {
		if got := NormalizeCodexReviewGateTreeScope(tc.in); got != tc.want {
			t.Errorf("NormalizeCodexReviewGateTreeScope(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestDefaultConfig_CodexReviewGateTreeScopeIsReview pins the distributed
// default: tree_scope is review (today's whole-tree behavior), so a project that
// never sets the key is unchanged.
func TestDefaultConfig_CodexReviewGateTreeScopeIsReview(t *testing.T) {
	if got := NewDefaultConfig().Workflow.Codex.ReviewGate.TreeScope; got != CodexReviewGateTreeScopeReview {
		t.Errorf("default tree_scope = %q, want %q", got, CodexReviewGateTreeScopeReview)
	}
}
