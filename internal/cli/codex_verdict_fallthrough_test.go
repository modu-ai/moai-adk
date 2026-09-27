package cli

import "testing"

// SPEC-CODEX-VERDICT-SYNTH-001 §0 — an unrecognized format is NOT a pass.
//
// The corpus below is the WITNESS of that property, never the requirement
// itself. acceptance.md §B is explicit about why: an AC that enumerates the
// formats an implementation was built to read verifies nothing — it repeats the
// implementation. So the assertions here iterate the corpus and apply ONE
// assertion; adding a member must never require editing an assertion.

// codexFormatCorpus is the AC-CVS-001 witness set (acceptance.md §B, initial 8).
// Members are deliberately unlike one another: when one regex covers two of
// them, the corpus has lost a witness rather than gained coverage.
var codexFormatCorpus = []struct {
	name string
	body string
}{
	{"C1 markdown blocking table", "| check | status |\n| --- | --- |\n| secrets | Blocking |\n| tests | ok |\n\nmerge_status: blocked"},
	{"C2 numbered findings", "1. Missing input validation at api.go:31\n2. Unchecked error at cli.go:80"},
	{"C3 json blob", `{"result":"blocked","issues":[{"file":"db.go","kind":"sqli"}]}`},
	{"C4 truncated heading", "## Review Summary"},
	{"C5 one line of prose", "I walked the diff and moved on."},
	{"C6 korean prose", "차단 사유 2건을 확인했습니다."},
	{"C7 empty", ""},
	{"C8 scored verdict", "FAIL 0.75 / 1.00\n\nBlocking issues: 2\n1. unchecked error at cli.go:80\n2. missing input validation at api.go:31"},
}

// TestSynthesizeReviewOutput_AdversarialNeverPassesUnknownFormat is AC-CVS-001.
//
// One assertion, applied across the corpus: in adversarial mode NOTHING the
// synthesizer fails to recognize may come back as "pass". Not recognizing a
// format is having observed nothing, which is not evidence of a clean review.
func TestSynthesizeReviewOutput_AdversarialNeverPassesUnknownFormat(t *testing.T) {
	for _, c := range codexFormatCorpus {
		if got := synthesizeReviewOutput(c.body, codexMethodTurnStart).Verdict; got == "pass" {
			t.Errorf("%s: adversarial synthesis returned %q; an unrecognized format must never synthesize a pass\nbody:\n%s", c.name, got, c.body)
		}
	}
}

// TestSynthesizeReviewOutput_UnknownMethodIsConservative pins the third row of
// plan.md §C.2: a method the synthesizer does not recognize is not a licence to
// assume the lenient default.
func TestSynthesizeReviewOutput_UnknownMethodIsConservative(t *testing.T) {
	for _, method := range []string{"", "thread/start", "review/resume"} {
		if got := synthesizeReviewOutput("I walked the diff and moved on.", method).Verdict; got != VerdictInconclusive {
			t.Errorf("method %q: Verdict = %q, want %q", method, got, VerdictInconclusive)
		}
	}
}

// TestSynthesizeReviewOutput_NativeNoSignalBodyInconclusive supersedes
// AC-CVS-003.
//
// AC-CVS-003 read the native path's bullet-less body as codex SAYING there is
// nothing to block on — the observation that justified the native "pass"
// default. SPEC-CODEX-PARSER-SHAPE-001 M4 (AC-CPS-004) revises exactly that
// reading: byte-level indistinguishability (verdict.md §E4) means a body
// alone cannot carry that observation, so the native request now pins an
// output format (REQ-CPS-005 as amended) and a body carrying no recognized
// signal is downgraded to inconclusive. A genuinely clean review states the
// pinned `Verdict: pass` line — TestCodexNativePinnedPassStaysPass witnesses
// that class; runTurn still short-circuits a blank body before the
// synthesizer (REQ-CBR-004), so this unit-level empty-body row documents the
// synthesizer's own contract only.
func TestSynthesizeReviewOutput_NativeNoSignalBodyInconclusive(t *testing.T) {
	cases := map[string]string{
		"no-signal review prose": "The change introduces no blocking issues.",
		"empty body":             "",
	}
	for name, body := range cases {
		if got := synthesizeReviewOutput(body, codexMethodReviewStart).Verdict; got != VerdictInconclusive {
			t.Errorf("%s: native Verdict = %q, want %q", name, got, VerdictInconclusive)
		}
	}
}

// TestSynthesizeReviewOutput_UnrecognizedBodyInconclusiveOnBothPaths supersedes
// the mode-split wiring witness (acceptance.md AC-CVS-003, which named C5).
// C5 still never reaches the synthesizer in production as a blank body, but
// since M4 the two paths no longer split on an unrecognized body — both report
// inconclusive. The split that REMAINS is in the requests, not the
// synthesizer: each path pins an output format of its own (native:
// TestCodexNativeRequest_CarriesFormatPin; adversarial:
// TestCodexAdversarialPrompt_PinsOutputFormat), and bodies following either
// pin are recognized on both paths (TestCodexAdversarialFormat_IsRecognized,
// TestCodexNativeFormat_IsRecognized).
func TestSynthesizeReviewOutput_UnrecognizedBodyInconclusiveOnBothPaths(t *testing.T) {
	const c5 = "I walked the diff and moved on."

	if got := synthesizeReviewOutput(c5, codexMethodReviewStart).Verdict; got != VerdictInconclusive {
		t.Errorf("native: Verdict = %q, want %q", got, VerdictInconclusive)
	}
	if got := synthesizeReviewOutput(c5, codexMethodTurnStart).Verdict; got != VerdictInconclusive {
		t.Errorf("adversarial: Verdict = %q, want %q", got, VerdictInconclusive)
	}
}
