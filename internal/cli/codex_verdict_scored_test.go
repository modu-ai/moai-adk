package cli

import "testing"

// TestSynthesizeReviewOutput_ScoredVerdictIsRead is AC-CVS-002.
//
// codex states a verdict in a score form too — "FAIL 0.75 / 1.00" at the head of
// a line. Reading it is REQ-CVS-002; the reason it is a narrow recognizer rather
// than a widening of the existing label regex is plan.md §C.3: PASS and FAIL are
// ordinary prose words, and a loose recognizer would read a sentence about a
// test suite as a verdict about the change.
func TestSynthesizeReviewOutput_ScoredVerdictIsRead(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"C8 scored fail", codexFormatCorpus[7].body, "fail"},
		{"scored pass", "PASS 0.88 / 1.00\n\nNothing to block on.", "pass"},
		{"scored inconclusive", "INCONCLUSIVE 0.50 / 1.00\n\nCould not reach the diff.", VerdictInconclusive},
		{"emphasised scored fail", "**FAIL 0.20 / 1.00**\n\nreasons follow", "fail"},
	}
	for _, c := range cases {
		if got := synthesizeReviewOutput(c.body, codexMethodTurnStart).Verdict; got != c.want {
			t.Errorf("%s: Verdict = %q, want %q\nbody:\n%s", c.name, got, c.want, c.body)
		}
	}
}

// TestSynthesizeReviewOutput_ScoredVerdictDoesNotMatchProse is the other half of
// AC-CVS-002. A recognizer that reads this sentence as a verdict has widened the
// hole rather than closed it: the corpus rule then makes this body a pass, which
// is exactly what §0 forbids for a body carrying no verdict at all.
func TestSynthesizeReviewOutput_ScoredVerdictDoesNotMatchProse(t *testing.T) {
	const prose = "the suite reported PASS 12 times before the regression"

	if got := synthesizeReviewOutput(prose, codexMethodTurnStart).Verdict; got == "pass" {
		t.Errorf("prose mentioning PASS was read as a verdict: Verdict = %q\nbody: %s", got, prose)
	}
}

// codexSignalCorpus is the AC-CVS-006 witness set (acceptance.md §B-2, initial
// 8 rows). Each row is (body, the signal SET that body produces, the expected
// adopted verdict), and the expectation is ALWAYS the most conservative member
// of that set — P-CONS, spec.md §A.5. No row carries a hand-made exception.
//
// [HARD] K3 and K7 are RED before M2 lands (baseline-rebased.md: both synthesize
// "pass"). Their expectations are derived from P-CONS, NOT from observed
// behavior, and must never be lowered to match what the tree currently does.
// Lowering them would silently delete the two rows' detecting power, which is
// the failure this whole card is about.
var codexSignalCorpus = []struct {
	name, body, want string
}{
	{"K1 stated fail then scored pass", "Verdict: fail — merge blocked.\n\nPASS 0.95 / 1.00", "fail"},
	{"K2 scored pass then stated fail", "PASS 0.95 / 1.00\n\nVerdict: fail — merge blocked.", "fail"},
	{"K3 stated pass then scored fail", "Verdict: pass — nothing to block on.\n\nFAIL 0.20 / 1.00", "fail"},
	{"K4 stated inconclusive then scored pass", "Verdict: inconclusive — could not reach the diff.\n\nPASS 0.95 / 1.00", VerdictInconclusive},
	{"K5 scored pass plus finding bullet", "PASS 0.95 / 1.00\n\n- [P1] path traversal at fs.go:44", "fail"},
	{"K6 three signals", "Verdict: pass — nothing to block on.\n\nINCONCLUSIVE 0.50 / 1.00\n\n- [P2] weak hash at auth.go:88", "fail"},
	{"K7 scored inconclusive then stated pass", "INCONCLUSIVE 0.50 / 1.00\n\nVerdict: pass — nothing to block on.", VerdictInconclusive},
	{"K8 signals agree", "Verdict: pass — nothing to block on.\n\nPASS 0.99 / 1.00", "pass"},
}

// TestSynthesizeReviewOutput_AdoptsMostConservativeSignal is AC-CVS-006.
//
// ONE assertion over the whole corpus. Adding a row — or adding a fourth signal
// — must never require editing it: P-CONS is stated over the signal SET, and a
// set has no order, so the rule does not change as signals accumulate. An
// assertion phrased as "a later signal does not overwrite an earlier one" would
// be an ordering rule, true only while there are three signals.
//
// This AC asserts the ADOPTED VALUE only. It requires nothing of the internal
// shape — rank table, assignment chain, or branch — so an implementation is free
// to satisfy P-CONS however it likes.
func TestSynthesizeReviewOutput_AdoptsMostConservativeSignal(t *testing.T) {
	for _, c := range codexSignalCorpus {
		if got := synthesizeReviewOutput(c.body, codexMethodTurnStart).Verdict; got != c.want {
			t.Errorf("%s: adopted Verdict = %q, want %q (the most conservative member of the signal set)\nbody:\n%s", c.name, got, c.want, c.body)
		}
	}
}

// TestSynthesizeReviewOutput_SignalOrderDoesNotMatter is the order-independence
// witness called out in acceptance.md §B-2: K1 and K2 carry the SAME signal set
// in opposite textual order, so they must adopt the same verdict. An
// implementation that layers assignments splits here.
func TestSynthesizeReviewOutput_SignalOrderDoesNotMatter(t *testing.T) {
	k1 := synthesizeReviewOutput(codexSignalCorpus[0].body, codexMethodTurnStart).Verdict
	k2 := synthesizeReviewOutput(codexSignalCorpus[1].body, codexMethodTurnStart).Verdict
	if k1 != k2 {
		t.Errorf("same signal set in opposite order adopted different verdicts: K1 = %q, K2 = %q", k1, k2)
	}
}
