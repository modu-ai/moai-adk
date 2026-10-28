// jevmeasure_cmd_test.go — SPEC-FETCH-WEB-001 M6 (AC-FW-012). The
// `moai jevmeasure bench` command: compare mode over recorded runs (pure
// data, no network), baseline write-then-diff, and the gated run mode that
// refuses with zero network while workflow.jev.enabled is false.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		Output: "# passthrough\n## (no heading)\nThe port is 3041.\n", KeptText: "The port is 3041.", OutputTokens: 3,
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

// Live run mode with the gate ON: the pipeline executes against the injected
// answerer, --run-out records the run, and the plain comparison report is
// emitted.
func TestJevmeasureBench_RunModeExecutesAndRecords(t *testing.T) {
	gateOn := jevProject(t, true)
	dir := t.TempDir()
	cases := []jevmeasure.BenchCase{{
		ID:            "case-1",
		Question:      "what is the port?",
		ExpectedQuote: "The port is 3041.",
		PageSnapshot:  "<html><body><p>The port is 3041.</p></body></html>",
	}}
	raw, _ := json.Marshal(cases)
	casesPath := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	runOut := filepath.Join(dir, "run.json")

	origRoot, origAsker := findProjectRootFn, fetchWebAskerFn
	findProjectRootFn = func() (string, error) { return gateOn, nil }
	fetchWebAskerFn = func() fetchweb.Answerer {
		return stubAsker{script: map[string]float64{"b0": 0.95, "c0": 0.9}}
	}
	t.Cleanup(func() { findProjectRootFn, fetchWebAskerFn = origRoot, origAsker })

	out, _, code := runBenchCmd(t, "--cases", casesPath, "--run-out", runOut)
	if code != 0 {
		t.Fatalf("live run exit = %d, want 0\n%s", code, out)
	}
	recorded, err := os.ReadFile(runOut)
	if err != nil {
		t.Fatalf("run result not recorded: %v", err)
	}
	var runs []jevmeasure.CaseRun
	if err := json.Unmarshal(recorded, &runs); err != nil || len(runs) != 1 {
		t.Fatalf("recorded run unparsable or wrong size: %v %s", err, recorded)
	}
	if !strings.Contains(out, "contained") {
		t.Errorf("comparison report missing:\n%s", out)
	}
}

// R8 (card t1968 review): a baseline-diff case that FAILS without being a
// regression (a case with no baseline entry) still drives exit 1 — the stated
// compare contract is "exit 0 when every verdict matches", not "exit 1 only
// on regressions".
func TestJevmeasureBench_NonRegressionFailStillExitsOne(t *testing.T) {
	dir := t.TempDir()
	casesPath, runPath := benchFixture(t, dir)
	baseline := filepath.Join(dir, "baseline.json")

	if _, _, code := runBenchCmd(t, "--cases", casesPath, "--run", runPath, "--baseline", baseline); code != 0 {
		t.Fatalf("baseline write exit = %d, want 0", code)
	}

	// Grow the case set: case-2 has no baseline entry, so it FAILS the diff
	// without being a containment regression.
	cases := []jevmeasure.BenchCase{
		{ID: "case-1", Question: "what is the port?", ExpectedQuote: "The port is 3041.", PageSnapshot: "<html><body><p>The port is 3041.</p></body></html>"},
		{ID: "case-2", Question: "what is the name?", ExpectedQuote: "The name is moai.", PageSnapshot: "<html><body><p>The name is moai.</p></body></html>"},
	}
	raw, _ := json.Marshal(cases)
	grown := filepath.Join(dir, "cases-grown.json")
	if err := os.WriteFile(grown, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	runs := []jevmeasure.CaseRun{
		{CaseID: "case-1", Status: "passthrough", Output: "# passthrough\n## (no heading)\nThe port is 3041.\n", KeptText: "The port is 3041.", OutputTokens: 3},
		{CaseID: "case-2", Status: "passthrough", Output: "# passthrough\n## (no heading)\nThe name is moai.\n", KeptText: "The name is moai.", OutputTokens: 3},
	}
	rawRuns, _ := json.Marshal(runs)
	run2 := filepath.Join(dir, "run2.json")
	if err := os.WriteFile(run2, rawRuns, 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, code := runBenchCmd(t, "--cases", grown, "--run", run2, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("new-case FAIL exit = %d, want 1 (any Pass==false fails the compare)\n%s", code, out)
	}
	if !strings.Contains(out, "FAIL case-2") {
		t.Errorf("report must FAIL the baseline-less case:\n%s", out)
	}
}

// C3 (card t1968 review round 3): a run result file holding no entry for a
// baseline case must exit 1. The zero-value lookup made the missing case read
// not-contained, which matched a not-contained baseline row and passed the
// compare silently (exit 0) despite no run result existing.
func TestJevmeasureBench_MissingRunResultFailsCompare(t *testing.T) {
	dir := t.TempDir()
	// case-1's quote is contained in its run output; case-2's is not (its
	// baseline verdict is not-contained — exactly the row the zero value
	// silently matched).
	cases := []jevmeasure.BenchCase{
		{ID: "case-1", Question: "what is the port?", ExpectedQuote: "The port is 3041.", PageSnapshot: "<html><body><p>The port is 3041.</p></body></html>"},
		{ID: "case-2", Question: "what license?", ExpectedQuote: "MIT LICENSE GRANTS", PageSnapshot: "<html><body><p>The license section is elsewhere.</p></body></html>"},
	}
	raw, _ := json.Marshal(cases)
	casesPath := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	fullRuns := []jevmeasure.CaseRun{
		{CaseID: "case-1", Status: "passthrough", Output: "# passthrough\n## (no heading)\nThe port is 3041.\n", KeptText: "The port is 3041.", OutputTokens: 3},
		{CaseID: "case-2", Status: "no_match", Output: "# no_match\n", KeptText: "", OutputTokens: 1},
	}
	rawRuns, _ := json.Marshal(fullRuns)
	fullRunPath := filepath.Join(dir, "run-full.json")
	if err := os.WriteFile(fullRunPath, rawRuns, 0o600); err != nil {
		t.Fatal(err)
	}
	baseline := filepath.Join(dir, "baseline.json")
	if _, _, code := runBenchCmd(t, "--cases", casesPath, "--run", fullRunPath, "--baseline", baseline); code != 0 {
		t.Fatalf("baseline write exit = %d, want 0", code)
	}

	// The partial run holds case-1's entry only — case-2 has no run result.
	partialRuns := fullRuns[:1]
	rawPartial, _ := json.Marshal(partialRuns)
	partialPath := filepath.Join(dir, "run-partial.json")
	if err := os.WriteFile(partialPath, rawPartial, 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, code := runBenchCmd(t, "--cases", casesPath, "--run", partialPath, "--baseline", baseline)
	if code != 1 {
		t.Fatalf("missing-run exit = %d, want 1 (never a silent zero-value pass)\n%s", code, out)
	}
	if !strings.Contains(out, "FAIL case-2") {
		t.Errorf("report must FAIL the run-less case:\n%s", out)
	}
}

// F1 (card t1968 iter18 codex gate, GLM-ruled F3 write-seam extension): a run
// record written by a PRE-kept-text build — a bare run list whose entries
// carry output but NO kept_text — launders through the --run import into
// WriteBaseline, which stamps format kept-text-1 on Contained=false verdicts
// the old maskable logic computed; a later no_match run of the same case then
// PASSes the equality, masking a real containment regression. The baseline
// side of the seam is gated (DiffAgainstBaseline refuses marker-less
// baselines); the import must REFUSE a pre-marker run record BEFORE anything
// is written, on the exit-1 channel, so WriteBaseline is never reached with
// unverifiable provenance.
func TestJevmeasureBench_RunImportRefusesPreMarkerRunRecord(t *testing.T) {
	dir := t.TempDir()
	cases := []jevmeasure.BenchCase{{
		ID:            "case-1",
		Question:      "what is the port?",
		ExpectedQuote: "The port is 3041.",
		PageSnapshot:  "<html><body><p>The port is 3041.</p></body></html>",
	}}
	raw, _ := json.Marshal(cases)
	casesPath := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	// The pre-marker build's on-disk shape: a bare run list, kept_text ABSENT,
	// output carrying the expected quote — the entry the old maskable measure
	// judged contained from the rendered markdown.
	oldRun := filepath.Join(dir, "run-old.json")
	oldRaw := `[{"case_id":"case-1","status":"passthrough","output":"# passthrough\n## (no heading)\nThe port is 3041.\n","output_tokens":3}]`
	if err := os.WriteFile(oldRun, []byte(oldRaw), 0o600); err != nil {
		t.Fatal(err)
	}
	// The later run this laundering masks: a legitimate no_match (kept_text
	// present and empty — the empty/absent distinction the gate must keep).
	noMatch := filepath.Join(dir, "run-nomatch.json")
	noMatchRaw := `[{"case_id":"case-1","status":"no_match","output":"# no_match\n","kept_text":"","output_tokens":0}]`
	if err := os.WriteFile(noMatch, []byte(noMatchRaw), 0o600); err != nil {
		t.Fatal(err)
	}
	baseline := filepath.Join(dir, "baseline.json")

	cmd := newJevmeasureCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"bench", "--cases", casesPath, "--run", oldRun, "--baseline", baseline})
	err := cmd.Execute()

	// Evidence block — runs only at a tree where the import launders: capture
	// the stamped baseline and the later no_match run's verdict against it.
	laundered := ""
	if err == nil {
		if rawBase, readErr := os.ReadFile(baseline); readErr == nil {
			out2, _, code2 := runBenchCmd(t, "--cases", casesPath, "--run", noMatch, "--baseline", baseline)
			laundered = fmt.Sprintf("baseline written: %s; later no_match run exit=%d, out=%q", rawBase, code2, out2)
		}
	}

	// The contract under test: refuse BEFORE anything is written.
	if err == nil {
		t.Fatalf("the --run import must REFUSE a pre-marker run record (F1 iter18 write-seam laundering); instead it laundered: exit 0, %s", laundered)
	}
	code := 1
	if c, ok := ResolveExitCode(err); ok {
		code = c
	}
	if code != 1 {
		t.Errorf("the refusal must ride the existing exit-1 channel, got exit %d (%v)", code, err)
	}
	if !strings.Contains(err.Error(), "predates kept-text containment") {
		t.Errorf("the refusal must name the format gap, got %q", err.Error())
	}
	if _, statErr := os.Stat(baseline); !os.IsNotExist(statErr) {
		rawBase, _ := os.ReadFile(baseline)
		t.Errorf("a refused import must write no baseline, stat err = %v, content %s", statErr, rawBase)
	}
}

// Argument and file errors surface as command errors (exit 1 via cobra).
func TestJevmeasureBench_InputErrors(t *testing.T) {
	if _, _, code := runBenchCmd(t); code != 1 {
		t.Errorf("missing --cases must be an error, exit = %d", code)
	}
	if _, _, code := runBenchCmd(t, "--cases", "/nonexistent/cases.json"); code != 1 {
		t.Errorf("unreadable cases file must be an error, exit = %d", code)
	}
	dir := t.TempDir()
	casesPath, _ := benchFixture(t, dir)
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, code := runBenchCmd(t, "--cases", casesPath, "--run", bad); code != 1 {
		t.Errorf("unparsable run file must be an error, exit = %d", code)
	}
}
