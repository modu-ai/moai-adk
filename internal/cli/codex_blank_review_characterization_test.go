package cli

import (
	"context"
	"errors"
	"testing"
)

// SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (card t551) — M1 characterization.
//
// These tests pin the behavior of the codex review-text path BEFORE the
// blank-output repair, so that the repair shows up as a visible diff in test
// expectations rather than as an unexplained new test (plan.md §D M1).
//
// One of them is NOT a pre-repair snapshot: TestCharacterize_UnavailableBackend
// is the permanent control (acceptance.md AC-CBR-004 state C). It pins the
// unavailable-backend fail-open path, which this SPEC must leave byte-identical.
// If it ever changes, the repair reached a path it was required to leave alone.

// errCharacterizeStartFailed drives the unavailable-backend state (state C).
var errCharacterizeStartFailed = errors.New("codex binary could not be started")

// withCodexSessionStartErr installs a fake session whose start always fails —
// the "backend unavailable" state (acceptance.md §A state C).
func withCodexSessionStartErr(t *testing.T, err error) {
	t.Helper()
	sess := withCodexSession(t, nil)
	sess.startErr = err
}

// runCodexTurnWithLines drives the REAL production path (runCodexReviewRPC) over
// a canned transcript. Every fixture in this SPEC's suite goes through here, so
// no assertion is made about a layer production does not traverse
// (plan.md §F AP-4).
func runCodexTurnWithLines(t *testing.T, lines []string) (ReviewOutput, error) {
	t.Helper()
	withCodexSession(t, lines)
	return runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart,
		map[string]any{"target": codexTargetUncommitted})
}

// TestCharacterize_UnavailableBackend is the PERMANENT control (AC-CBR-004):
// when codex cannot be reached, the audit returns the fail-open inconclusive
// whose Summary names codex unavailability. This SPEC changes nothing here, so
// the pin below is a before/after equality, not a snapshot to be updated.
func TestCharacterize_UnavailableBackend(t *testing.T) {
	withCodexSessionStartErr(t, errCharacterizeStartFailed)
	out, err := runCodexReviewRPC(context.Background(), "/fake/codex", codexMethodReviewStart,
		map[string]any{"target": codexTargetUncommitted})
	if err == nil {
		t.Fatal("unavailable backend must still surface its cause alongside the fail-open output")
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q", out.Verdict, VerdictInconclusive)
	}
	const wantSummary = "codex unavailable: codex session start failed: codex binary could not be started"
	if out.Summary != wantSummary {
		t.Errorf("summary = %q, want %q (byte-identical pin — REQ-CBR-006)", out.Summary, wantSummary)
	}
	if len(out.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(out.Findings))
	}
	if len(out.NextSteps) != 1 || out.NextSteps[0] != "fall back to the active auditor (claude)" {
		t.Errorf("next_steps = %v, want the claude-fallback step", out.NextSteps)
	}
}

// TestCharacterize_ReviewTextPathPreRepair pins the pre-repair verdict/summary of
// the review-text path for each fixture class. The rows marked PRE-REPAIR
// SNAPSHOT are the ones whose expectations the repair commit rewrites in place.
func TestCharacterize_ReviewTextPathPreRepair(t *testing.T) {
	cases := []struct {
		name        string
		review      string
		wantVerdict string
		wantSummary string
	}{
		// PRE-REPAIR SNAPSHOT (rewritten by the repair commit of
		// SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001): a body carrying no
		// non-whitespace character passes the exact-equality guard at
		// internal/cli/mcp_codex.go:817, reaches the synthesizer, and the
		// native review mode's unrecognized-body default turns it into `pass`
		// with an empty Summary — a review that never produced a verdict,
		// reported as a review that found nothing wrong.
		{"space-only", " ", "pass", ""},
		{"newline-only", "\n", "pass", ""},
		{"mixed-whitespace", "\n\t  \n", "pass", ""},
		// PRE-REPAIR SNAPSHOT: exactly-empty is already inconclusive, but its
		// Summary carries the UNAVAILABLE wording for a state that is not
		// unavailability. The repair gives it the blank-output wording it
		// shares with the rows above (acceptance.md §C).
		{"exactly-empty", "", VerdictInconclusive,
			"codex unavailable: codex review produced no verdict text"},
		// UNCHANGED by this SPEC — the control (state B).
		{"real-clean-review", "The change introduces no blocking issues.", "pass",
			"The change introduces no blocking issues."},
		// UNCHANGED by this SPEC — findings still outrank everything.
		{"finding-bullet", "- [P1] injection at vuln.go:5", "fail",
			"- [P1] injection at vuln.go:5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := runCodexTurnWithLines(t, codexSessionScript(tc.review))
			if out.Verdict != tc.wantVerdict {
				t.Errorf("verdict = %q, want %q", out.Verdict, tc.wantVerdict)
			}
			if out.Summary != tc.wantSummary {
				t.Errorf("summary = %q, want %q", out.Summary, tc.wantSummary)
			}
		})
	}
}
