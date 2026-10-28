package jevmeasure

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// bench_test.go — SPEC-FETCH-WEB-001 M6 (AC-FW-012, REQ-FW-014). Bench
// cases in the jevfetch shape (question, expected quote, page snapshot), the
// comparison report (containment + token-savings), and the deterministic
// baseline.json diff. No network: the runner takes the Answerer seam and the
// golden test supplies a stub (REQ-FW-010).

// stubBenchAsker scores by the question's embedded passage: a passage
// containing the marker scores high, everything else low. This is what makes
// the savings limb observable — a flat score would keep (or pass through)
// the whole snapshot and demonstrate nothing.
type stubBenchAsker struct {
	marker string
}

func (s stubBenchAsker) Ask(_ context.Context, req jev.Request) jev.Result {
	answers := make([]jev.Answer, 0, len(req.Questions))
	for _, q := range req.Questions {
		p := 0.05
		if strings.Contains(q.Text, s.marker) {
			p = 0.95
		}
		answers = append(answers, jev.Answer{
			QuestionID: q.ID, Kind: jev.KindNoul, Probability: p, Model: jev.VerifiedInterpretedModel,
		})
	}
	return jev.Result{Availability: jev.Available, Answers: answers}
}

// paddedSnapshot pads the install page past the ~2,000-token passthrough
// bound with filler paragraphs so the pipeline actually filters.
func paddedSnapshot() string {
	var b strings.Builder
	b.WriteString("<html><body><h1>Install</h1><p>You need Go 1.23 or later.</p>")
	for i := 0; i < 150; i++ {
		b.WriteString("<p>Filler paragraph with unrelated maintenance notes that no install question needs.</p>")
	}
	b.WriteString("</body></html>")
	return b.String()
}

func benchCases() []BenchCase {
	return []BenchCase{
		{
			ID:            "install-req",
			Question:      "what are the install requirements?",
			ExpectedQuote: "You need Go 1.23 or later.",
			PageSnapshot:  paddedSnapshot(),
		},
		{
			ID:            "absent-quote",
			Question:      "what license?",
			ExpectedQuote: "MIT LICENSE GRANTS EVERYTHING",
			PageSnapshot:  "<html><body><p>The license section is elsewhere.</p></body></html>",
		},
	}
}

// Case loading reads the jevfetch shape from a JSON file.
func TestLoadBenchCases_JSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")
	raw, err := json.Marshal(benchCases())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	cases, err := LoadBenchCases(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 || cases[0].ID != "install-req" || cases[0].ExpectedQuote == "" {
		t.Fatalf("loaded %+v", cases)
	}
}

// Running the bench executes the fetchweb pipeline over the snapshot with
// the injected answerer and records each case's output.
func TestRunBench_RecordsOutput(t *testing.T) {
	runs := RunBench(context.Background(), benchCases(), stubBenchAsker{marker: "You need Go"})
	if len(runs) != 2 {
		t.Fatalf("runs = %d, want 2", len(runs))
	}
	if runs[0].CaseID != "install-req" || runs[0].Output == "" {
		t.Fatalf("run[0] = %+v", runs[0])
	}
	if !strings.Contains(runs[0].Output, "You need Go 1.23 or later.") {
		t.Errorf("high-relevance run must keep the expected block:\n%s", runs[0].Output)
	}
}

// P2 (card t1968 codex review, repair round): containment must measure the
// KEPT BLOCK TEXT, not the rendered markdown. RenderMarkdown's first line is
// the "# <status>" status line, so an ExpectedQuote equal to a status name
// ("no_match") counted as contained with ZERO excerpts kept. A status echo
// alone is not containment.
func TestCompareRun_ContainmentIgnoresStatusLine(t *testing.T) {
	cases := []BenchCase{{
		ID:            "status-echo",
		Question:      "what is the port?",
		ExpectedQuote: "no_match", // equals the pipeline status name
		PageSnapshot:  paddedSnapshot(),
	}}
	// No passage carries the marker: every stage-1 score is low, no candidate
	// exists, and the pipeline ends in no_match with an empty kept set.
	runs := RunBench(context.Background(), cases, stubBenchAsker{marker: "zz-no-passage-carries-this-zz"})
	if runs[0].Status != string(fetchweb.StatusNoMatch) {
		t.Fatalf("fixture premise: status = %q, want no_match (zero candidates)", runs[0].Status)
	}
	report := CompareRun(cases, runs)
	if report.Cases[0].Contained {
		t.Errorf("Containment counted the STATUS LINE: ExpectedQuote %q matched the status render line while the kept set is empty (run.Output = %q)",
			cases[0].ExpectedQuote, runs[0].Output)
	}
}

// The comparison report: per-case containment and token-savings rate versus
// the snapshot's token count (approximate counts), plus a set summary.
func TestCompareRun_ContainmentAndSavings(t *testing.T) {
	runs := RunBench(context.Background(), benchCases(), stubBenchAsker{marker: "You need Go"})
	report := CompareRun(benchCases(), runs)

	byID := map[string]CaseComparison{}
	for _, c := range report.Cases {
		byID[c.CaseID] = c
	}
	if !byID["install-req"].Contained {
		t.Error("install-req quote must be contained in the output")
	}
	if byID["absent-quote"].Contained {
		t.Error("absent-quote must not read as contained")
	}
	if byID["install-req"].SnapshotTokens <= 0 {
		t.Errorf("snapshot tokens = %d, want > 0", byID["install-req"].SnapshotTokens)
	}
	if byID["install-req"].OutputTokens <= 0 || byID["install-req"].OutputTokens > byID["install-req"].SnapshotTokens {
		t.Errorf("output tokens %d out of (0, snapshot %d]", byID["install-req"].OutputTokens, byID["install-req"].SnapshotTokens)
	}
	if byID["install-req"].SavingsRate <= 0 || byID["install-req"].SavingsRate >= 1 {
		t.Errorf("savings rate %v out of (0,1)", byID["install-req"].SavingsRate)
	}
	if report.Summary.N != 2 || report.Summary.Contained != 1 {
		t.Errorf("summary = %+v, want N=2 Contained=1", report.Summary)
	}
}

// The baseline contract: write a baseline.json, then a second comparison of
// the same run reports PASS per case deterministically; a mutated run (quote
// gone where the baseline said contained) reports FAIL for that case.
func TestBaseline_WriteThenDeterministicDiff(t *testing.T) {
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "baseline.json")
	cases := benchCases()
	runs := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})

	// First comparison writes the baseline.
	if _, err := WriteBaseline(baselinePath, cases, runs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(baselinePath); err != nil {
		t.Fatalf("baseline not written: %v", err)
	}

	// Same inputs again: deterministic per-case PASS.
	diff, err := DiffAgainstBaseline(baselinePath, cases, runs)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range diff.Cases {
		if c.CaseID == "install-req" && !c.Pass {
			t.Errorf("stable run must PASS against its own baseline: %+v", c)
		}
	}

	// Regression: a run whose kept text lost the quote flips install-req to
	// FAIL while absent-quote stays at its baseline verdict.
	degraded := append([]CaseRun(nil), runs...)
	for i := range degraded {
		if degraded[i].CaseID == "install-req" {
			degraded[i].Output = "# no_match\n"
			degraded[i].KeptText = ""
		}
	}
	diff2, err := DiffAgainstBaseline(baselinePath, cases, degraded)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]CaseDiff{}
	for _, c := range diff2.Cases {
		byID[c.CaseID] = c
	}
	if byID["install-req"].Pass {
		t.Error("a lost quote must FAIL against a baseline that recorded it contained")
	}
	if !byID["absent-quote"].Pass {
		t.Error("absent-quote's stable verdict must stay PASS")
	}
	if !diff2.AnyRegression {
		t.Error("AnyRegression must be true when a case flipped to not-contained")
	}
}

// C3 (card t1968 review round 3): a case with NO run result must FAIL the
// baseline diff explicitly. The run map lookup yields the zero value
// (not-contained, zero output tokens), which matched a baseline row that
// recorded not-contained — so a missing run entry passed the comparison
// silently.
func TestBaselineDiff_MissingRunResultFails(t *testing.T) {
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "baseline.json")
	cases := benchCases()
	runs := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})
	if _, err := WriteBaseline(baselinePath, cases, runs); err != nil {
		t.Fatal(err)
	}
	// Drop the absent-quote entry: its case stays in cases and in the
	// baseline (recorded not-contained), but no run result exists for it.
	partial := runs[:0:0]
	for _, r := range runs {
		if r.CaseID != "absent-quote" {
			partial = append(partial, r)
		}
	}
	diff, err := DiffAgainstBaseline(baselinePath, cases, partial)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]CaseDiff{}
	for _, c := range diff.Cases {
		byID[c.CaseID] = c
	}
	if byID["absent-quote"].Pass {
		t.Error("a case with no run result must FAIL the diff, not pass on the zero value")
	}
	if byID["absent-quote"].Reason == "" {
		t.Error("the missing-run FAIL must name its reason")
	}
	if !byID["install-req"].Pass {
		t.Errorf("the case with a real run result must keep its verdict: %+v", byID["install-req"])
	}
}

// C4 (t1968 sync-audit iter6): a fetch_error run is not comparable to any
// baseline verdict. The old diff scored it PASS against a Contained=false
// baseline row (an error output contains nothing), and the savings formula
// turned its zero output tokens into a bogus 100% savings inside the summary
// mean. A failed run now FAILs the comparison with its reason named, and its
// savings are excluded from the case row and the mean.
func TestBaselineDiff_FetchErrorRunFailsAndExcludedFromSavings(t *testing.T) {
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "baseline.json")
	cases := benchCases()
	// Baseline rows recorded by a healthy run (absent-quote: Contained=false).
	healthy := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})
	if _, err := WriteBaseline(baselinePath, cases, healthy); err != nil {
		t.Fatal(err)
	}
	// The current run for absent-quote ends in fetch_error (an unavailable
	// scoring call, say): empty output, no kept text, zero kept tokens.
	failed := append([]CaseRun(nil), healthy...)
	for i := range failed {
		if failed[i].CaseID == "absent-quote" {
			failed[i].Status = string(fetchweb.StatusFetchError)
			failed[i].Output = ""
			failed[i].KeptText = ""
			failed[i].OutputTokens = 0
		}
	}

	// The diff: a failed run FAILs — containment equality against a
	// not-contained baseline row is not a pass for an error.
	diff, err := DiffAgainstBaseline(baselinePath, cases, failed)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]CaseDiff{}
	for _, c := range diff.Cases {
		byID[c.CaseID] = c
	}
	if byID["absent-quote"].Pass {
		t.Error("a fetch_error run must FAIL the comparison, not pass against a Contained=false baseline row")
	}
	if byID["absent-quote"].Reason == "" {
		t.Error("the fetch_error FAIL must name its reason")
	}
	if !byID["install-req"].Pass {
		t.Errorf("the healthy case must keep its verdict: %+v", byID["install-req"])
	}

	// Savings exclusion: the error run's row carries no savings figure and
	// the summary mean averages only the comparable cases — a zero-token
	// error run must not read as 100% savings.
	report := CompareRun(cases, failed)
	byCase := map[string]CaseComparison{}
	for _, c := range report.Cases {
		byCase[c.CaseID] = c
	}
	if s := byCase["absent-quote"].SavingsRate; s != 0 {
		t.Errorf("a fetch_error run's savings must stay 0, got %v", s)
	}
	if report.Summary.MeanSavingsRate != byCase["install-req"].SavingsRate {
		t.Errorf("mean %v must exclude the fetch_error case (want %v, the sole comparable rate)",
			report.Summary.MeanSavingsRate, byCase["install-req"].SavingsRate)
	}
}

// P2-a (card t1968 codex review, repair round): WriteBaseline must refuse a
// run set that is not a measurement, mirroring the CompareRun/L3 semantics.
// A run that ended in fetch_error stored its empty output like a legit
// measurement (Contained=false), and a later run of the same case then
// compared PASS against the poisoned baseline; a case with no run result (or
// a run naming an unknown case) leaves the same hole L3 names in the
// comparison. Refusal returns an error and writes NO file.
func TestWriteBaseline_RefusesFailedOrMissingRuns(t *testing.T) {
	cases := benchCases()
	healthy := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})
	dir := t.TempDir()

	// Control: a healthy full run still writes.
	okPath := filepath.Join(dir, "ok-baseline.json")
	if _, err := WriteBaseline(okPath, cases, healthy); err != nil {
		t.Fatalf("healthy run must write: %v", err)
	}
	if _, err := os.Stat(okPath); err != nil {
		t.Fatalf("healthy baseline not written: %v", err)
	}

	// A run ending in fetch_error is refused and nothing is written.
	failed := append([]CaseRun(nil), healthy...)
	for i := range failed {
		if failed[i].CaseID == "absent-quote" {
			failed[i].Status = string(fetchweb.StatusFetchError)
			failed[i].Output = ""
			failed[i].OutputTokens = 0
		}
	}
	errPath := filepath.Join(dir, "failed-baseline.json")
	if _, err := WriteBaseline(errPath, cases, failed); err == nil {
		t.Error("WriteBaseline must refuse a run containing fetch_error")
	}
	if _, statErr := os.Stat(errPath); !os.IsNotExist(statErr) {
		t.Errorf("refused baseline must write nothing, stat err = %v", statErr)
	}

	// A case with no run result is refused (L3 existence semantics).
	partial := healthy[:0:0]
	for _, r := range healthy {
		if r.CaseID != "absent-quote" {
			partial = append(partial, r)
		}
	}
	missPath := filepath.Join(dir, "missing-baseline.json")
	if _, err := WriteBaseline(missPath, cases, partial); err == nil {
		t.Error("WriteBaseline must refuse a case with no run result")
	}
	if _, statErr := os.Stat(missPath); !os.IsNotExist(statErr) {
		t.Errorf("refused baseline must write nothing, stat err = %v", statErr)
	}

	// A run naming a case that does not exist is refused: the diff walks
	// cases, so such an entry could never be compared — storing it bakes an
	// unverifiable verdict into the baseline.
	stray := append(append([]CaseRun(nil), healthy...), CaseRun{
		CaseID: "no-such-case", Status: string(fetchweb.StatusFiltered), Output: "x",
	})
	strayPath := filepath.Join(dir, "stray-baseline.json")
	if _, err := WriteBaseline(strayPath, cases, stray); err == nil {
		t.Error("WriteBaseline must refuse a run with no matching bench case")
	}
	if _, statErr := os.Stat(strayPath); !os.IsNotExist(statErr) {
		t.Errorf("refused baseline must write nothing, stat err = %v", statErr)
	}
}

// F3 (card t1968 iter15 codex gate, operator scope-widened): a baseline
// written by an OLDER build (pre-23d02c65e3) stores CaseComparison verdicts
// computed against the rendered markdown — the maskable containment the P2
// repair replaced — and carries no kept-text provenance. The auditor's
// scenario, mirrored below: an old-format row recorded Contained=false for a
// case, and a later no_match run of the same case (Contained=false under the
// current semantics) PASSED the equality — a real containment regression
// crossing the semantics change passed silently. DiffAgainstBaseline must
// REFUSE such a baseline outright (fail-closed): error, empty diff, no
// comparison. The refusal covers an absent marker (a pre-format build) and an
// unrecognized one (a build this diff cannot attribute semantics to).
func TestBaselineDiff_RefusesOldFormatBaseline(t *testing.T) {
	dir := t.TempDir()
	cases := benchCases()
	// The later run: install-req keeps its quote; absent-quote ends in
	// no_match with an empty kept set — the run whose masked equality against
	// the old Contained=false row is the regression the refusal must close.
	runs := []CaseRun{
		{CaseID: "install-req", Status: string(fetchweb.StatusFiltered), Output: "# filtered\nYou need Go 1.23 or later.\n", KeptText: "You need Go 1.23 or later.", OutputTokens: 8},
		{CaseID: "absent-quote", Status: string(fetchweb.StatusNoMatch), Output: "# no_match\n", KeptText: "", OutputTokens: 0},
	}
	// The pre-marker rows an older build stored for this case set: a
	// Contained=false entry for absent-quote is the poisoned verdict.
	rows := `{"case_id":"install-req","contained":true,"output_tokens":8,"snapshot_tokens":2100,"savings_rate":0.99},` +
		`{"case_id":"absent-quote","contained":false,"output_tokens":9,"snapshot_tokens":12,"savings_rate":0.25}`

	t.Run("no format marker", func(t *testing.T) {
		path := filepath.Join(dir, "baseline-no-marker.json")
		if err := os.WriteFile(path, []byte(`{"cases":[`+rows+`]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		diff, err := DiffAgainstBaseline(path, cases, runs)
		if err == nil {
			t.Fatalf("DiffAgainstBaseline must REFUSE a baseline with no format marker (F3 fail-closed); the comparison ran and reported %+v — the no_match run PASSes against the old Contained=false row, masking the regression", diff.Cases)
		}
		if !strings.Contains(err.Error(), "re-baseline") {
			t.Errorf("the refusal must tell the operator to re-baseline, got %q", err.Error())
		}
		if len(diff.Cases) != 0 || diff.AnyRegression {
			t.Errorf("a refused baseline must yield an empty diff, got %+v", diff)
		}
	})

	t.Run("unrecognized format marker", func(t *testing.T) {
		path := filepath.Join(dir, "baseline-odd-marker.json")
		if err := os.WriteFile(path, []byte(`{"format":"kept-text-2","cases":[`+rows+`]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		diff, err := DiffAgainstBaseline(path, cases, runs)
		if err == nil {
			t.Fatalf("DiffAgainstBaseline must REFUSE a baseline whose format marker it does not recognize (F3 fail-closed); the comparison ran and reported %+v", diff.Cases)
		}
		if !strings.Contains(err.Error(), "re-baseline") {
			t.Errorf("the refusal must tell the operator to re-baseline, got %q", err.Error())
		}
		if len(diff.Cases) != 0 || diff.AnyRegression {
			t.Errorf("a refused baseline must yield an empty diff, got %+v", diff)
		}
	})
}

// F3 (card t1968 iter15): WriteBaseline stamps the kept-text format marker on
// every baseline it writes, and DiffAgainstBaseline accepts its own build's
// baseline — the write/diff round-trip that makes refusing every other format
// safe. The marker value is pinned here as an on-disk contract: a build that
// changes it must treat existing baselines as old-format and refuse them.
func TestWriteBaseline_StampsKeptTextFormat(t *testing.T) {
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "baseline.json")
	cases := benchCases()
	runs := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})
	if _, err := WriteBaseline(baselinePath, cases, runs); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Format string `json:"format"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Format != "kept-text-1" {
		t.Fatalf("written baseline format = %q, want kept-text-1 (raw: %s)", doc.Format, raw)
	}
	diff, err := DiffAgainstBaseline(baselinePath, cases, runs)
	if err != nil {
		t.Fatalf("a baseline written by this build must diff without refusal, got %v", err)
	}
	for _, c := range diff.Cases {
		if !c.Pass {
			t.Errorf("round-trip diff must PASS every case, got %+v", c)
		}
	}
}

// L3 (t1968 leader audit): a case with NO run result must not aggregate as
// a perfect run. The run map lookup yielded a zero CaseRun whose empty
// status is not fetch_error, so the row computed SavingsRate=1 (100%
// savings from zero output tokens) and pulled the summary mean up. A
// missing run is a comparison failure, not a 100%-savings success
// (C4/iter6 semantics: absent = not comparable, never silent 100%): the row
// names the case missing, carries no savings figure, and the overall
// verdict FAILs it per the existing compare contract.
func TestCompareRun_MissingRunExcludedFromSavings(t *testing.T) {
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "baseline.json")
	cases := benchCases()
	runs := RunBench(context.Background(), cases, stubBenchAsker{marker: "You need Go"})
	if _, err := WriteBaseline(baselinePath, cases, runs); err != nil {
		t.Fatal(err)
	}
	// A run file that carries only install-req: absent-quote has no result.
	partial := runs[:0:0]
	for _, r := range runs {
		if r.CaseID == "install-req" {
			partial = append(partial, r)
		}
	}
	report := CompareRun(cases, partial)
	byID := map[string]CaseComparison{}
	for _, c := range report.Cases {
		byID[c.CaseID] = c
	}
	if !byID["absent-quote"].Missing {
		t.Error("the case with no run result must be NAMED missing in its row")
	}
	if got := byID["absent-quote"].SavingsRate; got != 0 {
		t.Errorf("a case with no run result must carry no savings figure, got %v (bogus 100%% from zero output tokens)", got)
	}
	present := byID["install-req"]
	if present.Missing {
		t.Error("the case with a run result must not read as missing")
	}
	if report.Summary.MeanSavingsRate != present.SavingsRate {
		t.Errorf("the mean must exclude the missing case (want exactly install-req's %v), got %v",
			present.SavingsRate, report.Summary.MeanSavingsRate)
	}
	// The overall verdict: the existing compare contract FAILs a missing run.
	diff, err := DiffAgainstBaseline(baselinePath, cases, partial)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range diff.Cases {
		if c.CaseID == "absent-quote" && c.Pass {
			t.Error("a missing run must FAIL the overall compare per the existing contract")
		}
	}
}
