package cli

// SPEC-CODEX-GATE-SCOPING-001 — the card-review repair round 2 (card t1404).
//
// The round-1 repair commits (2c25df957, d2f3b1d1e, 12152cd5b) were
// re-reviewed by codex and five NEW defects were returned, all in the
// path-resolution / reclassification-target family. Each test below pins one
// finding through an EXISTING entry point, authored BEFORE its repair and
// observed RED on the pre-repair tree:
//
//	N1  codexFindingsOf takes the FIRST path:line in a finding message as
//	    File, so a headline naming the config surface while the actual
//	    location is a source file reclassifies a real defect as drift and
//	    silently ALLOWs — the anchor must stay unset when the message carries
//	    several distinct path candidates (REQ-CGSC-008 reclassifies only
//	    findings targeting ONLY those surfaces)
//	N3  finding-path normalization anchors on the session tree, so a session
//	    sitting in a subdirectory relativizes an absolute repo-root config
//	    path to a "../" form that escapes the exclusion sets and a config-only
//	    FAIL stays BLOCK — the comparison anchors on the GIT REPOSITORY ROOT
//	N4  isPrimaryCheckoutGit compares spellings, and through a symlinked
//	    subdirectory git reports --git-dir as the REAL absolute path while
//	    --git-common-dir comes back relative — the same primary directory
//	    judges not-primary through the link and the default skip is missed
//	N5  porcelain collapses a fully-untracked .moai/ tree to "?? .moai/", so
//	    the config-only probe cannot see .moai/config/ inside the collapsed
//	    entry and a config-only untracked change counts as reviewable — the
//	    probe collects untracked files at file level (--untracked-files=all)
//	(N2 is the sync gate script's scan/key parity; its repro lives in
//	internal/template/hook_gate_reports_scan_parity_test.go beside the M4
//	tests it repairs.)

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// --- N1: the parser claims no anchor it cannot defend ----------------------

// multiPathDriftReviewText is the N1 shape: the finding's HEADLINE names the
// runtime-config surface while its ACTUAL location is a source file — two
// distinct path:line candidates in one message. Taking the first candidate
// misreads the finding as config drift and silently ALLOWs a source defect.
const multiPathDriftReviewText = "- [P1] `.moai/config/sections/workflow.yaml:25` gate key drifted — fix the loader guard at `internal/config/loader.go:125`\n"

// TestCodexFindingsOfAmbiguousMultiPathMessageLeavesAnchorUnset pins N1 at the
// parser: a message carrying SEVERAL distinct path candidates has no
// defensible single location, so the anchor stays unset and consumers that
// require an unambiguous target (the REQ-CGSC-008 reclassification) keep the
// strict disposition. A reference inside a title is not the target; ambiguity
// is not resolved by position.
//
// Controls pin the two shapes that stay anchored: exactly one candidate (the
// message's location), and one DISTINCT candidate repeated (the same path at
// two lines is still one location).
func TestCodexFindingsOfAmbiguousMultiPathMessageLeavesAnchorUnset(t *testing.T) {
	out := synthesizeReviewOutput(multiPathDriftReviewText, codexMethodReviewStart)
	if out.Verdict != "fail" || len(out.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", out.Verdict, out.Findings)
	}
	f := out.Findings[0]
	if f.File != "" || f.Line != 0 {
		t.Errorf("a message with two distinct path candidates has no defensible anchor — File:Line = %q:%d, want unset so the reclassification never claims it", f.File, f.Line)
	}

	single := synthesizeReviewOutput("- [P1] `.claude/settings.json:13` personal PATH entry drifted\n", codexMethodReviewStart)
	if len(single.Findings) != 1 || single.Findings[0].File != ".claude/settings.json" || single.Findings[0].Line != 13 {
		t.Errorf("control: an exactly-one-candidate message keeps its anchor, got %+v", single.Findings)
	}
	repeat := synthesizeReviewOutput("- [P1] `main.go:10` duplicated key, also at `main.go:44`\n", codexMethodReviewStart)
	if len(repeat.Findings) != 1 || repeat.Findings[0].File != "main.go" {
		t.Errorf("control: one DISTINCT candidate repeated keeps its anchor (first occurrence's line), got %+v", repeat.Findings)
	}
}

// TestCodexReviewGateMultiPathFindingKeepsStrictDisposition pins N1 end to end
// on the Claude Stop path: the ambiguous-anchor finding must keep the gate's
// only block path, and no reclassification row may be recorded for it.
func TestCodexReviewGateMultiPathFindingKeepsStrictDisposition(t *testing.T) {
	prem := synthesizeReviewOutput(multiPathDriftReviewText, codexMethodReviewStart)
	if prem.Verdict != "fail" || len(prem.Findings) != 1 {
		t.Fatalf("premise: the fixture must synthesize fail with one finding, got verdict %q findings %+v", prem.Verdict, prem.Findings)
	}

	withChangeDetector(t, true)
	withCodexSession(t, codexSessionScript(multiPathDriftReviewText))

	read := captureGateDiagnostics(t)
	out, err := HandleCodexReviewGate(gateInput(false), true, "/proj")
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	diagnostics := read()
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Errorf("a fail finding whose message cannot pin an unambiguous config target must keep the gate's block, got %+v", out)
	}
	if strings.Contains(diagnostics, "runtime-managed") {
		t.Errorf("no reclassification may be recorded for an ambiguous-anchor finding; diagnostics: %q", diagnostics)
	}
}

