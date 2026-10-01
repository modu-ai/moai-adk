package config

import "testing"

// TestClaudeTierConstants pins the three agent-tier constants byte-exact
// (SPEC-AGENT-TIER-001 REQ-TIER-002 / AC-TIER-001):
//
//	max    → {sonnet-5-5, max}    (70.6% @ ~$11 — accuracy-first)
//	medium → {sonnet-5-5, high}   (45% @ ~$2.3 — cost-efficiency sweet spot)
//	low    → {sonnet-5-5, medium} (29% @ ~$0.8)
//
// The tier tokens are CONFIGURATION-KEY names, not effort values (REQ-TIER-001
// Q2): the pair's Effort field keeps the chart-axis vocabulary, the constant
// name is keyed by tier token. RED-now (run start): the symbols do not exist
// anywhere in the package — this test failed to COMPILE naming them (R1).
func TestClaudeTierConstants(t *testing.T) {
	if DefaultClaudeTierMax != (ModelEffort{Model: "sonnet-5-5", Effort: "max"}) {
		t.Errorf("DefaultClaudeTierMax = %+v, want {sonnet-5-5 max}", DefaultClaudeTierMax)
	}
	if DefaultClaudeTierMedium != (ModelEffort{Model: "sonnet-5-5", Effort: "high"}) {
		t.Errorf("DefaultClaudeTierMedium = %+v, want {sonnet-5-5 high}", DefaultClaudeTierMedium)
	}
	if DefaultClaudeTierLow != (ModelEffort{Model: "sonnet-5-5", Effort: "medium"}) {
		t.Errorf("DefaultClaudeTierLow = %+v, want {sonnet-5-5 medium}", DefaultClaudeTierLow)
	}
}

// TestClaudeTierMaxFallbackRecorded pins the max-tier fallback pair
// (REQ-TIER-003 / AC-TIER-008): {claude-opus-5-5, xhigh} — the availability/
// dispersion alternative (65% @ ~$5 on the chart). The SPEC requires the
// RECORD only, not automatic failover; see the constant's doc comment for the
// chart grounding. RED-now (run start): compile failure — the symbol did not
// exist (authored with TestClaudeTierConstants, R1).
func TestClaudeTierMaxFallbackRecorded(t *testing.T) {
	if DefaultClaudeTierMaxFallback != (ModelEffort{Model: "claude-opus-5-5", Effort: "xhigh"}) {
		t.Errorf("DefaultClaudeTierMaxFallback = %+v, want {claude-opus-5-5 xhigh}", DefaultClaudeTierMaxFallback)
	}
}
