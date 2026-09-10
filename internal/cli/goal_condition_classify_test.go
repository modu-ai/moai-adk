package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// canonicalAcConvergeProse is the run.md § Run-phase Autonomy (ac_converge)
// condition text, authored entirely as model conditions ("every predicate
// references a line the orchestrator surfaces in the conversation"). It is the
// exact shape the orchestrator arms at the Implementation Kickoff Approval gate,
// and the shape reported in issue #1660.
//
// Routed to the mechanical path it is a shell syntax error, so the evaluator can
// never observe the expected exit 0 — the block persists to the turn ceiling.
const canonicalAcConvergeProse = "Every blocking acceptance criterion in " +
	".moai/specs/SPEC-XXX/acceptance.md has its PASS evidence surfaced in " +
	"the conversation (test output, build exit 0, or explicit AC-id: PASS " +
	"line); AND `go test ./...` exit 0 is surfaced; AND no test file outside " +
	"the SPEC scope was modified (surfaced via git status). Stop when all hold."

// TestParseCondition_CanonicalAcConvergeIsModel is the t288 regression: the
// canonical ac_converge prose MUST classify as a model condition. Before the
// fix, parseCondition keyed on the single literal token "transcript", which the
// canonical text does not carry (it says "in the conversation"), so the whole
// paragraph became a mechanical condition — a shell command that can never exit
// 0, blocking every turn-end until the ceiling.
//
// This is a classifier defect, NOT an MCP-wrapper defect: handleGoalArm
// (mcp_server.go) has called parseCondition since #1378, so the CLI and MCP
// paths misclassify identically.
func TestParseCondition_CanonicalAcConvergeIsModel(t *testing.T) {
	t.Parallel()

	cond := parseCondition(canonicalAcConvergeProse)
	if cond.Type != goal.ConditionModel {
		t.Fatalf("canonical ac_converge prose classified as %q (cmd=%q), want %q\n"+
			"a prose claim routed to the mechanical path is executed as a shell "+
			"command and blocks every turn-end", cond.Type, cond.Cmd, goal.ConditionModel)
	}
	if cond.Claim != canonicalAcConvergeProse {
		t.Errorf("model condition claim = %q, want the full condition text", cond.Claim)
	}
	if cond.Cmd != "" {
		t.Errorf("model condition carries cmd = %q, want empty", cond.Cmd)
	}
}

// TestParseCondition_ShellCommandsStayMechanical guards the other direction: the
// broadened discriminator MUST NOT pull runnable commands into the model path.
func TestParseCondition_ShellCommandsStayMechanical(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		in         string
		wantCmd    string
		wantExpect int
	}{
		{"plain command", "go test ./internal/cli/...", "go test ./internal/cli/...", 0},
		{"trailing exits clause", "go build ./... exits 0", "go build ./...", 0},
		{"non-zero expected exit", "grep -q TODO main.go exits 1", "grep -q TODO main.go", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cond := parseCondition(tc.in)
			if cond.Type != goal.ConditionMechanical {
				t.Fatalf("%q classified as %q, want mechanical", tc.in, cond.Type)
			}
			if cond.Cmd != tc.wantCmd {
				t.Errorf("cmd = %q, want %q", cond.Cmd, tc.wantCmd)
			}
			if cond.ExpectExit != tc.wantExpect {
				t.Errorf("expect_exit = %d, want %d", cond.ExpectExit, tc.wantExpect)
			}
		})
	}
}

// TestParseCondition_TranscriptReferentsAreModel pins both referents named by
// REQ-GLE-032 ("a natural-language claim that references the conversation
// transcript"): the implementation must accept BOTH words, not only one.
func TestParseCondition_TranscriptReferentsAreModel(t *testing.T) {
	t.Parallel()

	for _, in := range []string{
		"all AC rows show PASS in the transcript",
		"all AC rows show PASS in the conversation",
		"Every AC has evidence surfaced in the Conversation",
	} {
		cond := parseCondition(in)
		if cond.Type != goal.ConditionModel {
			t.Errorf("%q classified as %q, want model", in, cond.Type)
		}
	}
}

// TestRunmdAcConvergeProseMatchesFixture keeps the fixture above honest: if the
// canonical run.md text drifts, this test says so rather than letting the
// regression silently stop covering the real condition.
func TestRunmdAcConvergeProseMatchesFixture(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../.claude/skills/moai/workflows/run.md")
	if err != nil {
		t.Skipf("run.md unavailable: %v", err)
	}
	body := string(data)
	for _, frag := range []string{
		"has its PASS evidence surfaced in",
		"the conversation (test output, build exit 0, or explicit AC-id: PASS",
	} {
		if !strings.Contains(body, frag) {
			t.Errorf("run.md no longer contains %q — update canonicalAcConvergeProse", frag)
		}
	}
}
