package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// liveProbeBodyPath is the review body codex actually returned during the t229
// investigation, kept as a report artifact and read here verbatim rather than
// retyped. A fixture paraphrased from an incident stops describing the incident.
const liveProbeBodyPath = "../../.moai/reports/t229/live-probe-body.txt"

// TestSynthesizeReviewOutput_LiveProbeBodyStaysInconclusive pins a body that
// already synthesizes correctly. That is the point: it states its verdict in the
// label form, so it reaches none of the paths this SPEC changed, and it is here
// to notice if a later edit to the recognizers or the fall-through moves it.
//
// It is a REGRESSION pin, not a detection. Nothing about it was red.
func TestSynthesizeReviewOutput_LiveProbeBodyStaysInconclusive(t *testing.T) {
	body, err := os.ReadFile(filepath.FromSlash(liveProbeBodyPath))
	if err != nil {
		t.Fatalf("read live probe body: %v", err)
	}
	if got := synthesizeReviewOutput(string(body), codexMethodTurnStart).Verdict; got != VerdictInconclusive {
		t.Errorf("live probe body synthesized Verdict = %q, want %q", got, VerdictInconclusive)
	}
}

// TestCodexTask_OutputTextUnchangedByVerdictSynthesis guards the seam this SPEC
// runs through but does not serve.
//
// codex_task shares runTurn with the review paths, so the fall-through change
// passes through its call too — but a task produces output, not a verdict, and
// it reads Summary rather than Verdict. The body below carries no verdict label,
// no score line, and no finding bullet, which is exactly the shape whose
// synthesized verdict this SPEC altered. The returned text must be indifferent
// to that.
func TestCodexTask_OutputTextUnchangedByVerdictSynthesis(t *testing.T) {
	const output = "walked the parser, no verdict to give"

	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, codexTaskScript("trn-verdictless", output))

	res := callCodexTask(t, map[string]any{"prompt": "walk the parser"})
	if res.IsError {
		t.Fatalf("unexpected IsError result: %+v", res)
	}
	got := structuredMap(t, res)
	if out, _ := got["output"].(string); out != output {
		t.Errorf("output = %q, want %q — codex_task returns the task's text, and the verdict synthesis must not reach it", out, output)
	}
}
