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

// TestNormalizeCodexReviewGatePrimaryScope pins the primary_scope value rule
// (SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-004, the §F.2 disposition table): only
// an explicit "review", ignoring case and surrounding whitespace, restores the
// pre-SPEC whole-tree review; every other value — empty, unknown, a prefix or
// suffix variant, the sibling policy's own "skip" — reads as skip, because the
// fail direction is REVERSED from tree_scope: skip is the default, and only an
// explicit review wins it.
func TestNormalizeCodexReviewGatePrimaryScope(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"review", CodexReviewGatePrimaryScopeReview},
		{" Review ", CodexReviewGatePrimaryScopeReview},
		{"REVIEW", CodexReviewGatePrimaryScopeReview},
		{"skip", CodexReviewGatePrimaryScopeSkip},
		{"", CodexReviewGatePrimaryScopeSkip},
		{"never", CodexReviewGatePrimaryScopeSkip},
		{"reviewx", CodexReviewGatePrimaryScopeSkip},
		{"re view", CodexReviewGatePrimaryScopeSkip},
	} {
		if got := NormalizeCodexReviewGatePrimaryScope(tc.in); got != tc.want {
			t.Errorf("NormalizeCodexReviewGatePrimaryScope(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestDefaultConfig_CodexReviewGatePrimaryScopeIsSkip pins the distributed
// default: primary_scope is skip (REQ-CGSC-002), so a project that never sets
// the key skips a primary-checkout tree session.
func TestDefaultConfig_CodexReviewGatePrimaryScopeIsSkip(t *testing.T) {
	if got := NewDefaultConfig().Workflow.Codex.ReviewGate.PrimaryScope; got != CodexReviewGatePrimaryScopeSkip {
		t.Errorf("default primary_scope = %q, want %q", got, CodexReviewGatePrimaryScopeSkip)
	}
}
