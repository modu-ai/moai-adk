package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CODEX-PARSER-SHAPE-001 candidate (c): a blocking verdict with an empty
// findings list and no GateUnmet annotation is self-contradictory — the verdict
// survived while the review's finding content was lost (the V8 shape). The
// adapter must say so rather than emit it as a clean review (REQ-CPS-006,
// AC-CPS-005), and must NOT say so for an unmet required gate, which produces
// the same verdict/findings pair legitimately (REQ-CPS-006a).

// codexV8Body is the V8 shape: the stated verdict still parses, while the
// finding markers (a numbered list here) sit outside every findings
// recognizer, so the findings list comes back empty.
const codexV8Body = "Verdict: fail\n\n1. [P1] Missing input validation at api.go:31\n2. [P2] Unchecked error at cli.go:80\n"

func TestSynthesizeReviewOutput_V8ContradictionIsReported(t *testing.T) {
	for _, method := range []string{codexMethodTurnStart, codexMethodReviewStart} {
		out := synthesizeReviewOutput(codexV8Body, method)
		if out.Verdict != "fail" || len(out.Findings) != 0 {
			t.Fatalf("%s: precondition: want the V8 output fail/0, got %s/%d", method, out.Verdict, len(out.Findings))
		}
		if out.Contradiction == "" {
			t.Errorf("%s: fail with zero findings and no GateUnmet was emitted without a contradiction report — a consumer reading findings sees a clean review", method)
		}
	}
}

// TestFlagVerdictFindingsContradiction_Predicate pins the three-term predicate
// verdict == fail && len(findings) == 0 && GateUnmet == "", including the
// mandatory unmet-gate control case.
func TestFlagVerdictFindingsContradiction_Predicate(t *testing.T) {
	cases := []struct {
		name string
		in   ReviewOutput
		want bool
	}{
		{"fail, no findings, no gate annotation (V8)", ReviewOutput{Verdict: "fail", Findings: []Finding{}}, true},
		{"fail, nil findings, no gate annotation", ReviewOutput{Verdict: "fail"}, true},
		{"CONTROL: fail, no findings, unmet required gate", ReviewOutput{Verdict: "fail", Findings: []Finding{}, GateUnmet: "workflow.audit.gates.codex is `required`, but this audit returned no verdict (fail-open inconclusive)"}, false},
		{"fail with findings", ReviewOutput{Verdict: "fail", Findings: []Finding{{Severity: "P1", Title: "x"}}}, false},
		{"pass, no findings", ReviewOutput{Verdict: "pass", Findings: []Finding{}}, false},
		{"inconclusive, no findings", ReviewOutput{Verdict: VerdictInconclusive, Findings: []Finding{}}, false},
	}
	for _, c := range cases {
		got := flagVerdictFindingsContradiction(c.in).Contradiction != ""
		if got != c.want {
			t.Errorf("%s: flagged = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestContradiction_UnmetRequiredGateIsNotFlagged runs the control case through
// the real producers: an unrecognized adversarial body synthesizes as
// inconclusive, applyGateUnmet turns it into fail + GateUnmet under a declared
// required gate, and neither step may report a parser contradiction.
func TestContradiction_UnmetRequiredGateIsNotFlagged(t *testing.T) {
	root := newProbeProject(t, "SPEC-CPS-CONTROL-001")
	writeCodexAuditGate(t, root, config.AuditGateRequired)

	out := applyGateUnmet(synthesizeReviewOutput("I walked the diff and moved on.", codexMethodTurnStart), root)
	if out.Verdict != "fail" || out.GateUnmet == "" || len(out.Findings) != 0 {
		t.Fatalf("precondition: want fail/0 with GateUnmet, got %s/%d gate=%q", out.Verdict, len(out.Findings), out.GateUnmet)
	}
	if out.Contradiction != "" {
		t.Errorf("unmet required gate reported as a contradiction: %q", out.Contradiction)
	}
	if again := flagVerdictFindingsContradiction(out); again.Contradiction != "" {
		t.Errorf("predicate flags the unmet-gate output: %q", again.Contradiction)
	}
}

// TestContradiction_1718Fixtures is AC-CPS-013 on the sanitized #1718
// reductions, evaluated against whatever the parser yields for them: a
// fixture is reported as contradictory exactly when its synthesized output is
// fail with zero findings. S1 and S2 carry no surviving blocking verdict on
// the parser before candidate (a), so (c) alone leaves them unflagged.
func TestContradiction_1718Fixtures(t *testing.T) {
	for _, name := range []string{"S1.txt", "S2.txt", "S2p.txt", "N1.txt", "N2.txt"} {
		body := readCodex1718Fixture(t, name)
		for _, method := range []string{codexMethodTurnStart, codexMethodReviewStart} {
			out := synthesizeReviewOutput(body, method)
			want := out.Verdict == "fail" && len(out.Findings) == 0
			if got := out.Contradiction != ""; got != want {
				t.Errorf("%s on %s (%s/%d): flagged = %v, want %v", name, method, out.Verdict, len(out.Findings), got, want)
			}
			t.Logf("CONTRADICTION %s %s verdict=%s findings=%d flagged=%v", name, method, out.Verdict, len(out.Findings), out.Contradiction != "")
		}
	}
}
