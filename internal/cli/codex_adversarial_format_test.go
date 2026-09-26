package cli

import (
	"strings"
	"testing"
)

// SPEC-CODEX-PARSER-SHAPE-001 candidate (d): the adversarial request pins an
// output format for the verdict and the findings that the recognizers accept
// (REQ-CPS-012).
//
// What these tests establish is the OFFLINE half only: the request asks for
// the format, and the format it asks for is one the parser reads. Whether live
// codex follows the request against a target project's own instructions is
// AC-CPS-014, which only a recorded live observation can decide — no test in
// this tree can satisfy it.

func TestCodexAdversarialPrompt_PinsOutputFormat(t *testing.T) {
	for _, focus := range []string{"", "the auth layer"} {
		p := codexAdversarialReviewPrompt(focus)
		for _, want := range []string{
			codexAdversarialVerdictFormat,
			codexAdversarialFindingFormat,
			"Verdict: pass", "Verdict: fail", "Verdict: inconclusive",
		} {
			if !strings.Contains(p, want) {
				t.Errorf("focus %q: prompt does not carry %q\nprompt: %s", focus, want, p)
			}
		}
		if focus != "" && !strings.Contains(p, "Focus area: "+focus+".") {
			t.Errorf("focus %q: focus clause lost", focus)
		}
	}
}

// TestCodexAdversarialFormat_IsRecognized ties the pinned format to the
// recognizers mechanically: a body written exactly in the format the request
// asks for synthesizes with the stated verdict and every finding structured.
func TestCodexAdversarialFormat_IsRecognized(t *testing.T) {
	fill := func(sev, msg, path, line string) string {
		return strings.NewReplacer("P1", sev, "<message>", msg, "<path>", path, "<line>", line).
			Replace(codexAdversarialFindingFormat)
	}
	body := strings.Replace(codexAdversarialVerdictFormat, "<pass|fail|inconclusive>", "fail", 1) + "\n" +
		fill("P1", "unchecked error", "internal/x.go", "12") + "\n" +
		"  the error from Close is dropped\n" +
		fill("P2", "missing input validation", "cmd/y.go", "7") + "\n"

	for _, method := range []string{codexMethodTurnStart, codexMethodReviewStart} {
		out := synthesizeReviewOutput(body, method)
		if out.Verdict != "fail" {
			t.Errorf("%s: verdict = %q, want fail\nbody:\n%s", method, out.Verdict, body)
		}
		if len(out.Findings) != 2 {
			t.Fatalf("%s: want exactly 2 findings, got %d (%+v)\nbody:\n%s", method, len(out.Findings), out.Findings, body)
		}
		if f := out.Findings[0]; f.Severity != "P1" || f.File != "internal/x.go" || f.Line != 12 || !strings.Contains(f.Body, "Close is dropped") {
			t.Errorf("%s: finding[0] = %+v", method, f)
		}
		if f := out.Findings[1]; f.Severity != "P2" || f.File != "cmd/y.go" || f.Line != 7 {
			t.Errorf("%s: finding[1] = %+v", method, f)
		}
		if out.Contradiction != "" {
			t.Errorf("%s: a conforming body reported as contradictory: %q", method, out.Contradiction)
		}
	}
}
