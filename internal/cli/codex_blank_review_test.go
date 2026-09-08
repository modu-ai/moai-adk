package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (card t551, GitHub #1632 blank-output
// axis) — the acceptance suite.
//
// The defect: three emptiness checks on the codex review-text path test EXACT
// equality with "", so a review body carrying only whitespace counts as content,
// passes the guard, reaches synthesizeReviewOutput, and the native review mode's
// unrecognized-body default turns it into `pass`. A review that never produced a
// verdict is reported as a review that found nothing wrong.
//
// acceptance.md §A: three states must be observed in the SAME run and must
// produce three DISTINGUISHABLE signals —
//
//	A blank            → inconclusive, Summary names blank review output
//	B reviewed, clean  → pass,         Summary carries the review prose
//	C backend absent   → inconclusive, Summary names codex unavailability
//
// A and C share a verdict and are separated by their Summary alone; that
// separation is itself a criterion (AC-CBR-005).

// --- transcript builders (multi-item; extends, never rewrites, the single-item
// codexSessionScript in mcp_codex_test.go) ---

// codexItemLine wraps one item object in an item/completed notification on the
// canned thread.
func codexItemLine(itemJSON string) string {
	return `{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","completedAtMs":1,"item":` + itemJSON + `}}`
}

// codexReviewItem is an exitedReviewMode item carrying the structured review.
func codexReviewItem(id, review string) string {
	return `{"type":"exitedReviewMode","id":` + jsonString(id) + `,"review":` + jsonString(review) + `}`
}

// codexAgentItem is an agentMessage item carrying free-form text.
func codexAgentItem(id, text string) string {
	return `{"type":"agentMessage","id":` + jsonString(id) + `,"text":` + jsonString(text) + `}`
}

// codexMultiItemScript drives the handshake, then the given item/completed
// notifications IN ORDER, then turn/completed. Ordering is the point: the
// collection loop reassigns on every matching item, so the order in which blank
// and real items arrive is what AC-CBR-009 measures.
func codexMultiItemScript(items ...string) []string {
	lines := []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
	}
	for _, item := range items {
		lines = append(lines, codexItemLine(item))
	}
	return append(lines,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`)
}

const realCleanReview = "The change introduces no blocking issues."

// blankFixtures are the bodies acceptance.md AC-CBR-001 enumerates: three that
// carry only whitespace, and the exactly-empty body that already returned
// inconclusive before this SPEC and must now share that treatment (§C).
var blankFixtures = map[string]string{
	"space-only":       " ",
	"newline-only":     "\n",
	"mixed-whitespace": "\n\t  \n",
	"exactly-empty":    "",
}

// --- the blankness discriminator (plan.md §B.2) ---

// TestCodexBlankReview_BlankDiscriminator pins the ONE discriminator all four
// sites share, including the U+00A0 decision plan.md §B.2 leaves open:
// strings.TrimSpace cuts on unicode.IsSpace, which DOES include the
// non-breaking space, so a body of non-breaking spaces alone is blank. The
// decision is asserted here rather than left implicit (acceptance.md §C).
func TestCodexBlankReview_BlankDiscriminator(t *testing.T) {
	blank := []string{"", " ", "\n", "\t", "\n\t  \n", " ", "   ", "\r\n"}
	notBlank := []string{realCleanReview, "- [P1] injection", ".", " x ", "0"}
	for _, s := range blank {
		if !codexReviewTextIsBlank(s) {
			t.Errorf("codexReviewTextIsBlank(%q) = false, want true", s)
		}
	}
	for _, s := range notBlank {
		if codexReviewTextIsBlank(s) {
			t.Errorf("codexReviewTextIsBlank(%q) = true, want false", s)
		}
	}
}

// --- AC-CBR-001 (state A) ---

// TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass: a body with no
// non-whitespace character, driven through the production runCodexReviewRPC
// path, yields inconclusive — never pass. Covers REQ-CBR-004.
func TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass(t *testing.T) {
	for name, body := range blankFixtures {
		t.Run(name, func(t *testing.T) {
			out, err := runCodexTurnWithLines(t, codexSessionScript(body))
			if out.Verdict == "pass" {
				t.Errorf("verdict = pass for a blank body — a review that produced no verdict was reported as one that found nothing wrong")
			}
			if out.Verdict != VerdictInconclusive {
				t.Errorf("verdict = %q, want %q", out.Verdict, VerdictInconclusive)
			}
			if err == nil {
				t.Error("blank output must surface its cause alongside the fail-open struct")
			}
		})
	}
}

// --- AC-CBR-002 (state B, CONTROL) ---

// TestCodexBlankReview_AC002_RealCleanReviewStillPasses is the control that
// proves AC-CBR-001 blocked the right thing rather than everything: a real clean
// review still passes and its prose still reaches the Summary verbatim.
// Covers REQ-CBR-007. Green BOTH before and after the repair.
func TestCodexBlankReview_AC002_RealCleanReviewStillPasses(t *testing.T) {
	out, err := runCodexTurnWithLines(t, codexSessionScript(realCleanReview))
	if err != nil {
		t.Fatalf("a completed clean review must not error: %v", err)
	}
	if out.Verdict != "pass" {
		t.Errorf("verdict = %q, want pass", out.Verdict)
	}
	if out.Summary != realCleanReview {
		t.Errorf("summary = %q, want the review prose verbatim %q", out.Summary, realCleanReview)
	}
}

// --- AC-CBR-003 (selection site) ---

// TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage:
// a whitespace-only exitedReviewMode.review must NOT shadow a real
// agentMessage.text. Covers REQ-CBR-003 only.
//
// The load-bearing assertion is the SELECTED TEXT, not the verdict: the verdict
// half is already `pass` on the pre-repair tree (the blank structured review
// reaches the synthesizer and the native-mode default produces pass), so a test
// asserting the verdict alone would be green before AND after and would prove
// nothing (acceptance.md §D).
func TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage(t *testing.T) {
	// (i) the selector itself — the site AC-CBR-003 pins (:1253).
	if got := bestCodexReviewText("   \n", realCleanReview); got != realCleanReview {
		t.Errorf("bestCodexReviewText(blank, real) = %q, want the agent text %q — a blank structured review must fall through", got, realCleanReview)
	}
	// The preference ordering itself is UNCHANGED and is pinned here so making
	// :1253 blank-aware cannot silently reverse it (progress.md residual risk F4).
	if got := bestCodexReviewText("structured wins", "agent loses"); got != "structured wins" {
		t.Errorf("bestCodexReviewText(real, real) = %q, want the structured review — the preference ordering is pre-existing behavior this SPEC preserves", got)
	}

	// (ii) end to end through the production path: the selected text is what the
	// Summary carries, so the Summary is the observable form of the selection.
	out, err := runCodexTurnWithLines(t, codexMultiItemScript(
		codexReviewItem("e1", "  \n\t "),
		codexAgentItem("a1", realCleanReview),
	))
	if err != nil {
		t.Fatalf("a real agent message is a completed review: %v", err)
	}
	if out.Summary != realCleanReview {
		t.Errorf("summary = %q, want %q — the real agent text must be the SELECTED text", out.Summary, realCleanReview)
	}
	if out.Verdict != "pass" {
		t.Errorf("verdict = %q, want pass — real review content is available, so this is not a blank-output state", out.Verdict)
	}
}

// --- AC-CBR-004 (state C, CONTROL) ---

// TestCodexBlankReview_AC004_UnavailableBackendStillFailsOpen: the
// unavailable-backend path is unchanged by this SPEC. The byte-identical
// comparison against the pre-SPEC behavior lives in
// TestCharacterize_UnavailableBackend, measured on the pre-repair tree in M1;
// this criterion asserts the contract in the acceptance suite's own terms.
// Covers REQ-CBR-006.
func TestCodexBlankReview_AC004_UnavailableBackendStillFailsOpen(t *testing.T) {
	out, err := runUnavailableCodexTurn(t)
	if err == nil {
		t.Fatal("an unreachable backend must surface its cause")
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q — codex is OPTIONAL and its absence fails open", out.Verdict, VerdictInconclusive)
	}
	if !strings.Contains(out.Summary, "codex unavailable") {
		t.Errorf("summary = %q, want it to name codex unavailability", out.Summary)
	}
}

// --- AC-CBR-005 (A and C are distinguishable) ---

// TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable: states A and
// C share the inconclusive verdict, so their Summary is the only thing that can
// separate them. Covers REQ-CBR-005. Asserts on the distinguishing substring,
// not full-string equality, so incidental wording changes do not break it.
func TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable(t *testing.T) {
	blankOut := codexTurnOutput(t, codexSessionScript(" \n "))

	// state C in its own sub-scope so its session double is torn down cleanly.
	var unavailableOut ReviewOutput
	t.Run("state-c", func(t *testing.T) {
		out, err := runUnavailableCodexTurn(t)
		if err == nil {
			t.Error("an unreachable backend must surface its cause")
		}
		unavailableOut = out
	})

	if blankOut.Verdict != VerdictInconclusive || unavailableOut.Verdict != VerdictInconclusive {
		t.Fatalf("both states must be inconclusive; blank=%q unavailable=%q", blankOut.Verdict, unavailableOut.Verdict)
	}
	if blankOut.Summary == unavailableOut.Summary {
		t.Fatalf("state A and state C carry the same Summary %q — an inconclusive that cannot be told apart from \"codex is missing\" is a new silence, not a repair", blankOut.Summary)
	}
	if !strings.Contains(blankOut.Summary, "blank") {
		t.Errorf("state-A summary = %q, want it to name blank review output", blankOut.Summary)
	}
	if !strings.Contains(unavailableOut.Summary, "codex unavailable") {
		t.Errorf("state-C summary = %q, want it to name codex unavailability", unavailableOut.Summary)
	}
}

// --- AC-CBR-006 (no blank body is reported as fail) ---

// TestCodexBlankReview_AC006_BlankBodyIsNeverFail pins spec.md §C against a
// later drift to fail-closed: codex is optional, and a blank body is one symptom
// of a backend that is present but not answering. Covers REQ-CBR-008. Green BOTH
// before and after the repair.
func TestCodexBlankReview_AC006_BlankBodyIsNeverFail(t *testing.T) {
	for name, body := range blankFixtures {
		t.Run(name, func(t *testing.T) {
			if got := codexTurnOutput(t, codexSessionScript(body)).Verdict; got == "fail" {
				t.Errorf("verdict = fail for a blank body — spec.md §C decides inconclusive, not fail: blocking on an optional backend that is installed but misbehaving is a worse answer than reporting that we could not tell")
			}
		})
	}
}

// --- AC-CBR-007 (required gate visibility) ---

// TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput: with
// workflow.audit.gates.codex=required, a blank-output inconclusive carries the
// GateUnmet annotation, so the unmet gate is visible to a machine consumer.
// Covers REQ-CBR-009. Exercises applyGateUnmet without modifying it.
func TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput(t *testing.T) {
	root := newProbeProject(t, "SPEC-BLANKGATE-904")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript("   \n  "))

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("fail-open stays a structured result, never a tool error")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q", v, VerdictInconclusive)
	}
	if g, _ := got["gate_unmet"].(string); g == "" {
		t.Error("gate_unmet is empty — a required gate whose backend produced no verdict text must be recorded as unmet")
	}
}

// --- AC-CBR-008 (the repair does not widen into the mode-keyed policy) ---

// TestCodexBlankReview_AC008_NonBlankUnrecognizedBodyUnchanged: a body that is
// PRESENT but matches no known verdict signal keeps the native review mode's
// documented `pass` default. Covers REQ-CBR-007. This is the criterion that
// stops the repair from widening into codexUnrecognizedVerdict's policy — only
// ABSENCE is reclassified. Green BOTH before and after the repair; its
// pre-repair value is recorded in
// .moai/reports/t551/probe-synthesizer-20260908.txt (prose-no-signal
// verdict="pass").
func TestCodexBlankReview_AC008_NonBlankUnrecognizedBodyUnchanged(t *testing.T) {
	const prose = "I looked at the diff."
	if got := synthesizeReviewOutput(prose, codexMethodReviewStart).Verdict; got != "pass" {
		t.Errorf("synthesize(%q, review/start).Verdict = %q, want pass — the mode-keyed default for a PRESENT unrecognized body is deliberate and documented", prose, got)
	}
	out, err := runCodexTurnWithLines(t, codexSessionScript(prose))
	if err != nil {
		t.Fatalf("a present body is a completed review: %v", err)
	}
	if out.Verdict != "pass" {
		t.Errorf("end-to-end verdict = %q, want pass", out.Verdict)
	}
}

// --- AC-CBR-009 (collection sites: a later blank item must not clobber) ---

// TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne covers REQ-CBR-001
// and REQ-CBR-002 — the ONLY behavioral difference the collection sites make.
// The collection loop REASSIGNS on every matching item/completed, so a later
// whitespace-only item today passes the != "" test and overwrites an earlier
// real one.
//
// The reversed orderings (blank arriving BEFORE real) are here to kill a
// first-wins mutant: an implementation that keeps the first non-blank value and
// ignores later ones satisfies the two real-then-blank rows while recording the
// blank in these, and no other criterion catches it.
func TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne(t *testing.T) {
	cases := []struct {
		name  string
		items []string
	}{
		{"review: real then blank", []string{
			codexReviewItem("e1", realCleanReview),
			codexReviewItem("e2", "  \n "),
		}},
		{"review: blank then real (kills first-wins)", []string{
			codexReviewItem("e1", "  \n "),
			codexReviewItem("e2", realCleanReview),
		}},
		{"agent: real then blank", []string{
			codexAgentItem("a1", realCleanReview),
			codexAgentItem("a2", "\t\n"),
		}},
		{"agent: blank then real (kills first-wins)", []string{
			codexAgentItem("a1", "\t\n"),
			codexAgentItem("a2", realCleanReview),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCodexTurnWithLines(t, codexMultiItemScript(tc.items...))
			if err != nil {
				t.Fatalf("the real text is a completed review: %v", err)
			}
			if out.Summary != realCleanReview {
				t.Errorf("summary = %q, want the surviving real text %q", out.Summary, realCleanReview)
			}
			if out.Verdict != "pass" {
				t.Errorf("verdict = %q, want pass — the real review survived, so this is not a blank-output state", out.Verdict)
			}
		})
	}
}

// --- acceptance.md §A: the three-state control matrix, in ONE run ---

// TestCodexBlankReview_ThreeStateControlMatrix observes states A, B and C in the
// same test and asserts they produce three DISTINGUISHABLE signals. Separating
// "no review happened" from "a review happened and found nothing" IS the repair;
// a suite that observes only the blank case cannot tell a correct repair from
// one that blocks everything.
func TestCodexBlankReview_ThreeStateControlMatrix(t *testing.T) {
	var stateA, stateB, stateC ReviewOutput

	t.Run("A-blank", func(t *testing.T) {
		stateA = codexTurnOutput(t, codexSessionScript(" \n\t "))
	})
	t.Run("B-reviewed-zero-findings", func(t *testing.T) {
		stateB = codexTurnOutput(t, codexSessionScript(realCleanReview))
	})
	t.Run("C-backend-unavailable", func(t *testing.T) {
		out, err := runUnavailableCodexTurn(t)
		if err == nil {
			t.Error("an unreachable backend must surface its cause")
		}
		stateC = out
	})

	if stateA.Verdict == "pass" {
		t.Errorf("state A verdict = pass, want NOT pass")
	}
	if stateA.Verdict != VerdictInconclusive {
		t.Errorf("state A verdict = %q, want %q", stateA.Verdict, VerdictInconclusive)
	}
	if stateB.Verdict != "pass" {
		t.Errorf("state B verdict = %q, want pass", stateB.Verdict)
	}
	if stateC.Verdict != VerdictInconclusive {
		t.Errorf("state C verdict = %q, want %q", stateC.Verdict, VerdictInconclusive)
	}
	// Three distinguishable signals: B by verdict, A and C by Summary.
	signals := map[string]string{
		"A": stateA.Verdict + "|" + stateA.Summary,
		"B": stateB.Verdict + "|" + stateB.Summary,
		"C": stateC.Verdict + "|" + stateC.Summary,
	}
	seen := map[string]string{}
	for state, sig := range signals {
		if other, dup := seen[sig]; dup {
			t.Errorf("states %s and %s produce the SAME signal %q — the three states are not distinguishable", other, state, sig)
		}
		seen[sig] = state
	}
	if !strings.Contains(stateA.Summary, "blank") {
		t.Errorf("state A summary = %q, want it to name blank review output", stateA.Summary)
	}
	if stateB.Summary != realCleanReview {
		t.Errorf("state B summary = %q, want the review prose", stateB.Summary)
	}
	if !strings.Contains(stateC.Summary, "codex unavailable") {
		t.Errorf("state C summary = %q, want it to name codex unavailability", stateC.Summary)
	}
}
