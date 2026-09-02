package cli

import (
	"strings"
	"testing"
)

// issue1632ReviewBody is the review body SHAPE GitHub #1632 reports: codex
// returned every finding as a severity-tagged bullet inside the prose, while
// the structured findings/next_steps arrays came back empty — "any consumer
// that reads the structured fields sees a clean review". Keeping the issue's
// own example as the fixture pins the parser to the shape the reporter
// actually saw, not to a paraphrase of it.
const issue1632ReviewBody = `Full review comments:
- [P1] Remove the GainUp credentials in internal/auth/keys.go:42 before merge
- [P1] Remove the Telegram token in services/notify/bot.go:7
- [P2] Remove the obsolete gate in tools/report.go:118
- [P2] Reconcile the report's conflicting totals`

// TestSynthesizeReviewOutput_Issue1632FindingsArePopulated is the headline
// acceptance of #1632 axis 1: each review comment must parse into findings[]
// with its severity, and the verdict must stay the conservative fail the
// finding bullets already produce (P-CONS, unchanged by this work).
func TestSynthesizeReviewOutput_Issue1632FindingsArePopulated(t *testing.T) {
	out := synthesizeReviewOutput(issue1632ReviewBody, codexMethodTurnStart)

	if len(out.Findings) != 4 {
		t.Fatalf("Findings: want 4 parsed findings, got %d (%+v)", len(out.Findings), out.Findings)
	}
	wantSev := []string{"P1", "P1", "P2", "P2"}
	for i, f := range out.Findings {
		if f.Severity != wantSev[i] {
			t.Errorf("Findings[%d].Severity = %q, want %q", i, f.Severity, wantSev[i])
		}
		if f.Title == "" {
			t.Errorf("Findings[%d].Title is empty — the message text must survive the parse", i)
		}
	}
	if out.Verdict != "fail" {
		t.Errorf("Verdict = %q, want fail — concrete findings keep the conservative verdict", out.Verdict)
	}
}

// TestSynthesizeReviewOutput_FindingCarriesPathAndLine covers the file:line
// anchors the issue names as part of the expected finding shape.
func TestSynthesizeReviewOutput_FindingCarriesPathAndLine(t *testing.T) {
	out := synthesizeReviewOutput(issue1632ReviewBody, codexMethodTurnStart)

	if len(out.Findings) < 2 {
		t.Fatalf("Findings: want ≥2 to inspect anchors, got %d — the arrays are empty (#1632 axis 1)", len(out.Findings))
	}
	f := out.Findings[0]
	if f.File != "internal/auth/keys.go" {
		t.Errorf("File = %q, want %q", f.File, "internal/auth/keys.go")
	}
	if f.Line != 42 {
		t.Errorf("Line = %d, want 42", f.Line)
	}
	second := out.Findings[1]
	if second.File != "services/notify/bot.go" || second.Line != 7 {
		t.Errorf("second finding File:Line = %q:%d, want services/notify/bot.go:7", second.File, second.Line)
	}
}

// TestSynthesizeReviewOutput_FindingWithoutPathLeavesFileEmpty — a bullet with
// no file:line anchor (the issue's fourth comment) still parses as a finding;
// the File/Line fields just stay empty rather than being filled with a guess.
func TestSynthesizeReviewOutput_FindingWithoutPathLeavesFileEmpty(t *testing.T) {
	out := synthesizeReviewOutput(issue1632ReviewBody, codexMethodTurnStart)

	if len(out.Findings) < 4 {
		t.Fatalf("Findings: want ≥4 to inspect the anchor-less comment, got %d", len(out.Findings))
	}
	f := out.Findings[3]
	if f.File != "" || f.Line != 0 {
		t.Errorf("File:Line = %q:%d, want empty — no anchor exists in the message", f.File, f.Line)
	}
	if !strings.Contains(f.Title, "Reconcile the report's conflicting totals") {
		t.Errorf("Title = %q, want the verbatim message text", f.Title)
	}
}

// TestSynthesizeReviewOutput_BulletlessBodyKeepsFindingsEmpty pins the clean
// case: a review with no finding bullets synthesizes an empty (non-nil) slice,
// exactly as before — the parser must not invent structure from prose.
func TestSynthesizeReviewOutput_BulletlessBodyKeepsFindingsEmpty(t *testing.T) {
	out := synthesizeReviewOutput("clean change, no findings", codexMethodReviewStart)

	if out.Findings == nil || len(out.Findings) != 0 {
		t.Errorf("Findings = %+v, want an empty non-nil slice", out.Findings)
	}
}

// TestSynthesizeReviewOutput_IndentedContinuationJoinsBody — codex often
// continues a finding across the following indented lines. Those lines belong
// to the finding's body; reading them as body-less prose would truncate the
// finding to its headline.
func TestSynthesizeReviewOutput_IndentedContinuationJoinsBody(t *testing.T) {
	body := "- [P1] Secret leakage in main.go:10\n  the key is committed in plaintext\n  rotate it before shipping"
	out := synthesizeReviewOutput(body, codexMethodTurnStart)

	if len(out.Findings) != 1 {
		t.Fatalf("Findings: want 1, got %d (%+v)", len(out.Findings), out.Findings)
	}
	f := out.Findings[0]
	if !strings.Contains(f.Body, "rotate it before shipping") {
		t.Errorf("Body does not carry the continuation lines: %q", f.Body)
	}
	if f.Line != 10 || f.File != "main.go" {
		t.Errorf("File:Line = %q:%d, want main.go:10", f.File, f.Line)
	}
}

// TestSynthesizeReviewOutput_VerbatimSummaryUnchangedByParsing guards the
// t229 contract the parser rides on: Summary keeps the verbatim review text,
// so the operator still sees codex's own words alongside the parsed findings.
func TestSynthesizeReviewOutput_VerbatimSummaryUnchangedByParsing(t *testing.T) {
	out := synthesizeReviewOutput(issue1632ReviewBody, codexMethodTurnStart)

	if out.Summary != strings.TrimSpace(issue1632ReviewBody) {
		t.Errorf("Summary was rewritten by findings parsing — it must stay the verbatim body")
	}
}
