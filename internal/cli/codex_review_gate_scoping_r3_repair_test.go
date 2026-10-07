package cli

// SPEC-CODEX-GATE-SCOPING-001 — the card-review repair round 3 (card t1404),
// a single-occurrence focused repair (M1+M2) whose evidence standard is the
// mutation contrast: each repro test is observed RED pre-repair, GREEN
// post-repair, and RED AGAIN with the repair temporarily reversed in the
// working tree (restored from HEAD — never stashed), replacing the codex
// re-review this round does not run.
//
//	M1  codexFindingAnchorOf decided the anchor from the finding's HEADLINE
//	    alone, so a headline naming the config surface while the body's
//	    continuation names the actual source location still anchored to the
//	    config path and reclassified a real defect as drift — the anchor
//	    candidates must be collected over the COMPLETE body (title + joined
//	    continuations), keeping N1's strict disposition
//	(M2 is the sync gate's surviving-roots sweep; its repro lives in
//	internal/template/hook_gate_reports_scan_parity_test.go.)

import (
	"context"
	"strings"
	"testing"

)

// titleBodyDriftReviewText is the M1 shape: the finding's HEADLINE names the
// runtime-config surface (one path candidate — unambiguous in isolation) while
// the body's indented continuation names the ACTUAL source location. The
// title-only decision anchored to the config path and silently ALLOWed the
// source defect.
const titleBodyDriftReviewText = "- [P1] `.moai/config/sections/workflow.yaml:25` gate key drifted\n" +
	"  the guard that must carry the fix is at `internal/config/loader.go:125`\n"

// TestCodexFindingsOfTitleBodyCrossCandidatesLeaveAnchorUnset pins M1 at the
// parser: the anchor is decided after the continuations join, over the
// COMPLETE body — title and body together carry two distinct file candidates,
// so the anchor stays unset and the reclassification keeps the strict
// disposition.
func TestCodexFindingsOfTitleBodyCrossCandidatesLeaveAnchorUnset(t *testing.T) {
	out := synthesizeReviewOutput(titleBodyDriftReviewText, codexMethodReviewStart)
	if out.Verdict != "fail" || len(out.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", out.Verdict, out.Findings)
	}
	f := out.Findings[0]
	if !strings.Contains(f.Body, "internal/config/loader.go:125") {
		t.Fatalf("premise: the continuation must join the body before the anchor decision, got body %q", f.Body)
	}
	if f.File != "" || f.Line != 0 {
		t.Errorf("a headline naming one file while the body names the actual location carries two distinct candidates — File:Line = %q:%d, want unset so the reclassification never claims it", f.File, f.Line)
	}

	// Control: one distinct candidate across title+body keeps its anchor —
	// the body-only path is an anchor source, not an ambiguity source.
	single := synthesizeReviewOutput("- [P1] `main.go:10` secret committed\n  rotate before shipping\n", codexMethodReviewStart)
	if len(single.Findings) != 1 || single.Findings[0].File != "main.go" || single.Findings[0].Line != 10 {
		t.Errorf("control: a single candidate spanning title+body keeps its anchor, got %+v", single.Findings)
	}
}

// TestCodexReviewGateTitleBodyCrossCandidateKeepsStrictDisposition pins M1 end
// to end at its M2 home, the receipt producer (the gate's live arm moved with
// the review — REQ-GBN-002): the cross-candidate finding must keep the strict
// disposition in the recorded receipt, and no reclassification row may be
// recorded for it.
func TestCodexReviewGateTitleBodyCrossCandidateKeepsStrictDisposition(t *testing.T) {
	prem := synthesizeReviewOutput(titleBodyDriftReviewText, codexMethodReviewStart)
	if prem.Verdict != "fail" || len(prem.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", prem.Verdict, prem.Findings)
	}

	root := cacheTestRoot(t)
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{stdoutByCmd: map[string]string{"--version": "9.9.9\n"}})
	withCodexSession(t, codexSessionScript(titleBodyDriftReviewText))

	read := captureGateDiagnostics(t)
	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("receipt producer error: %v", err)
	}
	diagnostics := read()
	if r.Verdict != codexReviewVerdictFail {
		t.Errorf("a fail finding whose headline and body pin different files must keep the strict disposition, got %q", r.Verdict)
	}
	if strings.Contains(diagnostics, "runtime-managed") {
		t.Errorf("no reclassification may be recorded for a cross-candidate finding; diagnostics: %q", diagnostics)
	}
}
