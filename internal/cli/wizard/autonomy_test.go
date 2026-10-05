package wizard

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-AUTONOMY-TIERS-001 M7 — interactive autonomy-tier wizard page (AC-001).
// The flag (M4) validates the closed set fail-loud; this surface adds the
// interactive selector PAGE so `moai init` prompts the user to pick a tier.
// semi-auto is pre-selected (REQ-006: no fully-autonomous default); the
// fully-autonomous option is gated at apply time via the existing
// EffectiveTierWithGates core (sandbox proof + kill-switch).

func TestInitQuestions_HasAutonomyTierPage(t *testing.T) {
	questions := InitQuestions("/tmp/test-project")
	q := QuestionByID(questions, "autonomy_tier")
	if q == nil {
		t.Fatal("InitQuestions must include an autonomy_tier question (AC-001 wizard page)")
	}
	if q.Type != QuestionTypeSelect {
		t.Errorf("autonomy_tier question must be a select, got %v", q.Type)
	}
	// REQ-006: semi-auto pre-selected — no fully-autonomous default ships.
	if q.Default != config.AutonomyTierSemiAuto {
		t.Errorf("autonomy_tier default must be %q, got %q", config.AutonomyTierSemiAuto, q.Default)
	}
	// The page must OFFER the 3-tier closed set (values, not labels).
	values := make(map[string]bool, len(q.Options))
	for _, opt := range q.Options {
		values[opt.Value] = true
	}
	for _, tier := range []string{
		config.AutonomyTierSemiAuto,
		config.AutonomyTierAutomatic,
		config.AutonomyTierFullyAutonomous,
	} {
		if !values[tier] {
			t.Errorf("autonomy_tier options must include %q; got %v", tier, q.Options)
		}
	}
	// A dedicated page label (NOT empty) so the question renders on its own page.
	if q.Group == "" {
		t.Error("autonomy_tier question must carry a Group label so it renders as a page")
	}
}

func TestAutonomyTierQuestion_FullyAutonomousNotRecommended(t *testing.T) {
	// AC-006: the selector copy MUST NOT pre-pick / recommend fully-autonomous.
	// The (Recommended) label lives on the acceptEdits-backed option
	// (semi-auto), never on fully-autonomous (SPEC-AUT-PERMMODES-001 REQ-002).
	// REQ-001: the labels speak Claude Code permission-mode vocabulary.
	questions := InitQuestions("/tmp/test-project")
	q := QuestionByID(questions, "autonomy_tier")
	if q == nil {
		t.Fatal("autonomy_tier question missing")
	}
	wantLabels := map[string]string{
		config.AutonomyTierSemiAuto:        "Accept edits on",
		config.AutonomyTierAutomatic:       "Auto mode",
		config.AutonomyTierFullyAutonomous: "Bypass permissions",
	}
	for _, opt := range q.Options {
		if want, ok := wantLabels[opt.Value]; ok && !contains(opt.Label, want) {
			t.Errorf("option %q label must carry the CC permission-mode vocabulary %q (REQ-001): %q", opt.Value, want, opt.Label)
		}
		if opt.Value == config.AutonomyTierSemiAuto && !contains(opt.Label, "(Recommended)") {
			t.Errorf("the acceptEdits-backed option (semi-auto) MUST carry the Recommended signal (REQ-002): %q", opt.Label)
		}
	}
	for _, opt := range q.Options {
		if opt.Value == config.AutonomyTierFullyAutonomous {
			for _, marker := range []string{"(Recommended)", "(recommended)"} {
				if contains(opt.Label, marker) {
					t.Errorf("fully-autonomous option MUST NOT carry %q: %q", marker, opt.Label)
				}
			}
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
