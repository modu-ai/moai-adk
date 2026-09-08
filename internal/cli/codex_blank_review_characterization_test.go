package cli

import (
	"errors"
	"testing"
)

// SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (card t551) — characterization.
//
// These tests were written in M1 against the UNREPAIRED tree and pinned its
// behavior verbatim, so the repair shows up as a visible diff in test
// expectations rather than as an unexplained new test (plan.md §D M1). The rows
// the repair changed carry the value they replaced in a comment.
//
// One of them was never a pre-repair snapshot: TestCharacterize_UnavailableBackend
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
	return runCodexReviewRPC(t.Context(), "/fake/codex", codexMethodReviewStart,
		map[string]any{"target": codexTargetUncommitted})
}

// codexTurnOutput is runCodexTurnWithLines for the criteria that assert on the
// output alone. The fail-open cause is logged rather than discarded, so a
// surprising error is visible in the test log instead of silently dropped.
func codexTurnOutput(t *testing.T, lines []string) ReviewOutput {
	t.Helper()
	out, err := runCodexTurnWithLines(t, lines)
	if err != nil {
		t.Logf("fail-open cause: %v", err)
	}
	return out
}

// runUnavailableCodexTurn drives state C: the backend cannot be started.
func runUnavailableCodexTurn(t *testing.T) (ReviewOutput, error) {
	t.Helper()
	withCodexSessionStartErr(t, errCharacterizeStartFailed)
	return runCodexReviewRPC(t.Context(), "/fake/codex", codexMethodReviewStart,
		map[string]any{"target": codexTargetUncommitted})
}

// TestCharacterize_UnavailableBackend is the PERMANENT control (AC-CBR-004):
// when codex cannot be reached, the audit returns the fail-open inconclusive
// whose Summary names codex unavailability. This SPEC changes nothing here, so
// the pin below is a before/after equality, not a snapshot to be updated.
func TestCharacterize_UnavailableBackend(t *testing.T) {
	out, err := runUnavailableCodexTurn(t)
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

// TestCharacterize_ReviewTextPath pins the verdict/summary of the review-text
// path for each fixture class. The rows marked REPAIRED are the ones whose
// expectations the repair commit rewrote in place; each records the pre-repair
// value it replaced, so the behavior change is readable here rather than only in
// the diff. The rows marked UNCHANGED are the controls.
func TestCharacterize_ReviewTextPath(t *testing.T) {
	cases := []struct {
		name        string
		review      string
		wantVerdict string
		wantSummary string
	}{
		// REPAIRED by SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001. The pre-repair
		// snapshot these three rows carried was {"pass", ""}: a body with no
		// non-whitespace character passed the exact-equality guard, reached the
		// synthesizer, and the native review mode's unrecognized-body default
		// turned it into `pass` with an empty Summary — a review that never
		// produced a verdict, reported as one that found nothing wrong.
		{"space-only", " ", VerdictInconclusive, codexBlankReviewSummary},
		{"newline-only", "\n", VerdictInconclusive, codexBlankReviewSummary},
		{"mixed-whitespace", "\n\t  \n", VerdictInconclusive, codexBlankReviewSummary},
		// REPAIRED: exactly-empty was already inconclusive, but its Summary was
		// "codex unavailable: codex review produced no verdict text" — the
		// UNAVAILABLE wording for a state that is not unavailability. It now
		// shares the blank-output wording with the rows above (acceptance.md §C)
		// and is therefore distinguishable from state C (AC-CBR-005).
		{"exactly-empty", "", VerdictInconclusive, codexBlankReviewSummary},
		// UNCHANGED by this SPEC — the control (state B).
		{"real-clean-review", "The change introduces no blocking issues.", "pass",
			"The change introduces no blocking issues."},
		// UNCHANGED by this SPEC — findings still outrank everything.
		{"finding-bullet", "- [P1] injection at vuln.go:5", "fail",
			"- [P1] injection at vuln.go:5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := codexTurnOutput(t, codexSessionScript(tc.review))
			if out.Verdict != tc.wantVerdict {
				t.Errorf("verdict = %q, want %q", out.Verdict, tc.wantVerdict)
			}
			if out.Summary != tc.wantSummary {
				t.Errorf("summary = %q, want %q", out.Summary, tc.wantSummary)
			}
		})
	}
}
