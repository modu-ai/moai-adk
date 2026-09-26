package cli

import "testing"

// SPEC-CODEX-PARSER-SHAPE-001 M3 — preserved-behaviour guards. These pin
// properties no candidate may break; they describe behaviour that existed
// before the run phase, so they pass on arrival and are proven live by
// mutation (see progress.md §E.2), not by a RED run.

// codexV9CleanNativeBody is the clean native review: codex found nothing to
// block on and said so in prose.
const codexV9CleanNativeBody = "The change introduces no identifiable correctness or blocking issues."

// TestGuard_CleanNativeReviewStaysPass is AC-CPS-008 / REQ-CPS-009: none of the
// widened recognizers and neither the contradiction report may make a genuinely
// clean native review loud.
func TestGuard_CleanNativeReviewStaysPass(t *testing.T) {
	for _, body := range []string{codexV9CleanNativeBody, "clean change, no findings", "Looks good to me."} {
		out := synthesizeReviewOutput(body, codexMethodReviewStart)
		if out.Verdict != "pass" || len(out.Findings) != 0 || out.Contradiction != "" {
			t.Errorf("clean native body %q: got %s/%d contradiction=%q, want pass/0 with no contradiction",
				body, out.Verdict, len(out.Findings), out.Contradiction)
		}
	}
}

// TestGuard_AdversarialUnrecognizedStaysInconclusive is AC-CPS-009 /
// REQ-CPS-010 (kept as written by the operator): a body matching no recognized
// signal still returns inconclusive on the adversarial path. A candidate may
// change WHICH bodies are recognized; it may not change this handling.
func TestGuard_AdversarialUnrecognizedStaysInconclusive(t *testing.T) {
	bodies := map[string]string{
		"N1": readCodex1718Fixture(t, "N1.txt"),
		"N2": readCodex1718Fixture(t, "N2.txt"),
	}
	for _, c := range codexFormatCorpus {
		if len(codexVerdictSignalsOf(c.body)) == 0 {
			bodies[c.name] = c.body
		}
	}
	if len(bodies) < 5 {
		t.Fatalf("guard swept only %d unrecognized bodies — the corpus lost its witnesses", len(bodies))
	}
	for name, body := range bodies {
		if len(codexVerdictSignalsOf(body)) != 0 {
			t.Fatalf("%s: precondition: body carries a recognized signal", name)
		}
		out := synthesizeReviewOutput(body, codexMethodTurnStart)
		if out.Verdict != VerdictInconclusive || out.Contradiction != "" {
			t.Errorf("%s: adversarial unrecognized body got %s contradiction=%q, want inconclusive with no contradiction",
				name, out.Verdict, out.Contradiction)
		}
	}
}

// TestGuard_NextStepsUntouchedOnNewShapes is AC-CPS-010 / REQ-CPS-011: the
// next_steps axis card t1052 closed stays an empty non-nil list, including for
// the bodies whose findings the widened recognizers now structure.
func TestGuard_NextStepsUntouchedOnNewShapes(t *testing.T) {
	for _, name := range []string{"S1.txt", "S2.txt", "S2p.txt"} {
		body := readCodex1718Fixture(t, name)
		for _, method := range []string{codexMethodTurnStart, codexMethodReviewStart} {
			out := synthesizeReviewOutput(body, method)
			if out.NextSteps == nil || len(out.NextSteps) != 0 {
				t.Errorf("%s on %s: NextSteps = %#v, want an empty non-nil slice", name, method, out.NextSteps)
			}
		}
	}
}
