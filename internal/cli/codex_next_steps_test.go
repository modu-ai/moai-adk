package cli

import (
	"strings"
	"testing"
)

// codexNextStepShapedBody is a review body whose prose is as next-step-shaped as
// codex output gets: an explicit "Next steps:" heading over imperative bullets,
// alongside the severity-tagged findings. If the synthesis path were ever to
// harvest steps from prose, THIS is the body it would harvest them from.
const codexNextStepShapedBody = `Verdict: fail

Full review comments:
- [P1] Remove the hardcoded token in internal/auth/keys.go:42 before merge
- [P2] Reconcile the report's conflicting totals

Next steps:
- Rotate the leaked credential and invalidate the old one
- Re-run the audit once the rotation has landed
`

// TestSynthesizeReviewOutput_NextStepsAreNotManufacturedFromProse pins the
// boundary this package declares at the NextSteps field: the synthesis path does
// NOT invent next_steps out of review prose.
//
// What this test asserts, precisely: no element of NextSteps is derived from the
// review body. It does NOT assert that NextSteps must be empty forever — a
// future tool-known routing instruction (the shape inconclusiveReviewWithSummary
// and codex_task's timeout branch already use) would be a legitimate non-empty
// value and this test would still pass. What it refuses is the other direction:
// copying codex's own words across, which is the move codexFindingsOf's
// contract rules out for this parser.
func TestSynthesizeReviewOutput_NextStepsAreNotManufacturedFromProse(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"finding bullets", issue1632ReviewBody},
		{"explicit next-steps section", codexNextStepShapedBody},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := synthesizeReviewOutput(tc.body, codexMethodTurnStart)

			// Positive control: this body DID yield structure, so an unharvested
			// NextSteps below is a boundary being honoured — not a parser that
			// silently did nothing.
			if len(out.Findings) == 0 {
				t.Fatalf("precondition: want findings parsed from this body, got none (%+v)", out)
			}

			if out.NextSteps == nil {
				t.Errorf("NextSteps: want a non-nil slice for schema uniformity, got nil")
			}
			for i, step := range out.NextSteps {
				if strings.Contains(tc.body, strings.TrimSpace(step)) {
					t.Errorf("NextSteps[%d] = %q was taken from the review prose; "+
						"steps must come from this package, never from codex's words", i, step)
				}
			}
		})
	}
}
