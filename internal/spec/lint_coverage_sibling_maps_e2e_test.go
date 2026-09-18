package spec_test

import "testing"

// Card t913 — the end-to-end bracket. The earlier work measured the absorption
// at the capture function; these two cases decide it through the real linter, on
// the finding the defect actually silenced.
//
// The pair is the control. Both fixtures use the numeric-tail shorthand; they
// differ only in what follows the tail, so the discriminator is the boundary and
// nothing else.
//
// MUTATION: delete the truncation branch in siblingMapsREQIDs — the disclaimer
// fixture drops from one CoverageIncomplete to zero (the silence this card
// removes) while the closed fixture stays at zero. A narrowing that overshoots
// instead breaks the closed fixture, which is the other half of the bracket.

// TestCoverageSibling_MapsProseDisclaimerStillFires is the "after" half: a REQ
// the sibling's prose explicitly disclaims is reported as uncovered.
func TestCoverageSibling_MapsProseDisclaimerStillFires(t *testing.T) {
	got := coverageSiblingFindings(t, "maps-prose-disclaimer")
	if len(got) != 1 {
		t.Fatalf("CoverageIncomplete findings = %d, want 1: %+v", len(got), got)
	}
	if got[0].Message == "" {
		t.Fatalf("finding carries no message: %+v", got[0])
	}
}

// TestCoverageSibling_MapsShorthandClosedIsCovered is the control: the same
// shorthand, closed by a token, still covers both REQs end to end.
func TestCoverageSibling_MapsShorthandClosedIsCovered(t *testing.T) {
	if got := coverageSiblingFindings(t, "maps-shorthand-closed"); len(got) != 0 {
		t.Fatalf("CoverageIncomplete findings = %d, want 0: %+v", len(got), got)
	}
}
