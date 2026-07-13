package settings

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
)

// Tests for the settings-package re-export of the sub-agent 4-color tier
// accessors (SPEC-WEBCONF-SIMPLIFY-001 M1). The canonical table lives in
// internal/harness/v4manifest; these accessors surface it next to
// V4EffortValues / V4ModelValues so the web layer imports the tier surface
// from the same package it already imports for effort/model closed sets.

// TestTierForAgent_Delegates verifies the settings accessor agrees with the
// v4manifest canonical table for a representative agent per tier.
func TestTierForAgent_Delegates(t *testing.T) {
	cases := map[string]v4manifest.Tier{
		"manager-spec":         v4manifest.TierRed,
		"manager-develop":      v4manifest.TierOrange,
		"manager-docs":         v4manifest.TierBlue,
		"hns-github-specialist": v4manifest.TierLightBlue,
	}
	for name, want := range cases {
		got, ok := TierForAgent(name)
		if !ok {
			t.Errorf("TierForAgent(%q) returned ok=false; want an entry", name)
			continue
		}
		if got != want {
			t.Errorf("TierForAgent(%q) = %q, want %q", name, got, want)
		}
	}
}

// TestTierForAgent_UnknownReturnsNotOK verifies the ok=false contract is
// preserved through the re-export.
func TestTierForAgent_UnknownReturnsNotOK(t *testing.T) {
	if _, ok := TierForAgent("nonexistent-future-agent"); ok {
		t.Error(`TierForAgent("nonexistent-future-agent") returned ok=true; want false`)
	}
}

// TestTierSuggestedModelEffort_Delegates verifies the settings accessor agrees
// with the v4manifest canonical tier → suggested-(model, effort) table
// (AC-WC-006).
func TestTierSuggestedModelEffort_Delegates(t *testing.T) {
	cases := []struct {
		tier       v4manifest.Tier
		wantModel  string
		wantEffort string
	}{
		{v4manifest.TierRed, v4manifest.ModelOpus, v4manifest.EffortXhigh},
		{v4manifest.TierOrange, v4manifest.ModelOpus, v4manifest.EffortHigh},
		{v4manifest.TierBlue, v4manifest.ModelSonnet, v4manifest.EffortMedium},
		{v4manifest.TierLightBlue, v4manifest.ModelHaiku, v4manifest.EffortLow},
	}
	for _, tc := range cases {
		gotModel, gotEffort := TierSuggestedModelEffort(tc.tier)
		if gotModel != tc.wantModel || gotEffort != tc.wantEffort {
			t.Errorf("TierSuggestedModelEffort(%q) = (%q, %q), want (%q, %q)",
				tc.tier, gotModel, gotEffort, tc.wantModel, tc.wantEffort)
		}
	}
}
