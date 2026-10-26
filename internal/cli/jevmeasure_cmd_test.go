// jevmeasure_cmd_test.go — SPEC-FETCH-WEB-001 M6 (AC-FW-012). The
// `moai jevmeasure bench` command: compare mode over recorded runs (pure
// data, no network), baseline write-then-diff, and the gated run mode that
// refuses with zero network while workflow.jev.enabled is false.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
	"github.com/modu-ai/moai-adk/internal/jevmeasure"
)

func runBenchCmd(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := newJevmeasureCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"bench"}, args...))
	err := cmd.Execute()
	code := 0
	if err != nil {
		if c, ok := ResolveExitCode(err); ok {
			code = c
		} else {
			code = 1
		}
	}
	return out.String(), errBuf.String(), code
}

// benchFixture writes cases.json and a run.json produced against a stub
// answerer (no network), returning their paths.
func benchFixture(t *testing.T, dir string) (casesPath, runPath string) {
	t.Helper()
	cases := []jevmeasure.BenchCase{{
		ID:            "case-1",
		Question:      "what is the port?",
		ExpectedQuote: "The port is 3041.",
		PageSnapshot:  "<html><body><p>The port is 3041.</p></body></html>",
	}}
	raw, _ := json.Marshal(cases)
	casesPath = filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	runs := []jevmeasure.CaseRun{{
		CaseID: "case-1", Status: "passthrough",
		Output: "# passthrough\n## (no heading)\nThe port is 3041.\n", OutputTokens: 3,
	}}
	rawRuns, _ := json.Marshal(runs)
	runPath = filepath.Join(dir, "run.json")
	if err := os.WriteFile(runPath, rawRuns, 0o600); err != nil {
		t.Fatal(err)
	}
	return casesPath, runPath
}

func TestJevmeasureBench_CommandSurface(t *testing.T) {
	cmd := newJevmeasureCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"bench", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "--baseline") || !strings.Contains(out.String(), "--cases") {
		t.Errorf("help must document the flags:\n%s", out.String())
	}
}

// Compare mode: a first invocation writes the baseline and passes; the same
// run again passes deterministically; a degraded run fails with exit 1.
func TestJevmeasureBench_CompareAndBaseline(t *testing.T) {
	dir := t.TempDir()
	casesPath, runPath := benchFixture(t, dir)
	baseline := filepath.Join(dir, "baseline.json")

	out, _, code := runBenchCmd(t, "--cases", casesPath, "--run", runPath, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("first compare exit = %d, want 0 (baseline write)\n%s", code, out)
	}
	if _, err := os.Stat(baseline); err != nil {
		t.Fatalf("baseline not written: %v", err)
	}

	out, _, code = runBenchCmd(t, "--cases", casesPath, "--run", runPath, "--baseline", baseline)
	if code != 0 {
		t.Fatalf("stable re-run exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "case-1") || !strings.Contains(out, "PASS") {
		t.Errorf("report must carry per-case PASS verdicts:\n%s", out)
	}

	// Degrade the run: the quote disappears.
	runs := []jevmeasure.CaseRun{{
		CaseID: "case-1", Status: "no_match", Output: "# no_match\n", OutputTokens: 0,
	}}
	rawRuns, _ := json.Marshal(runs)
	degraded := filepath.Join(dir, "degraded.json")
	if err := os.WriteFile(degraded, rawRuns, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, code = runBenchCmd(t, "--cases", casesPath, "--run", degraded, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("regression exit = %d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "FAIL") {
		t.Errorf("regression report must FAIL the case:\n%s", out)
	}
}

// Run mode is gated: with workflow.jev.enabled=false the command performs
// ZERO network requests, prints the fallback notice, exits 2.
func TestJevmeasureBench_RunModeGatedOff(t *testing.T) {
	gateOff := jevProject(t, false)
	dir := t.TempDir()
	casesPath, _ := benchFixture(t, dir)

	origRoot, origAsker := findProjectRootFn, fetchWebAskerFn
	findProjectRootFn = func() (string, error) { return gateOff, nil }
	trip := &tripwireAsker{t: t}
	fetchWebAskerFn = func() fetchweb.Answerer { return trip }
	t.Cleanup(func() { findProjectRootFn, fetchWebAskerFn = origRoot, origAsker })

	out, errBuf, code := runBenchCmd(t, "--cases", casesPath)
	if code != 2 {
		t.Fatalf("gated-off bench exit = %d, want 2\n%s", code, out)
	}
	if !strings.Contains(errBuf, "workflow.jev.enabled") {
		t.Errorf("stderr must name the gate:\n%s", errBuf)
	}
	if trip.calls != 0 {
		t.Errorf("%d Jev calls escaped the gate", trip.calls)
	}
}
