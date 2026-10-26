package jevmeasure

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	// Regression: a run whose output lost the quote flips install-req to
	// FAIL while absent-quote stays at its baseline verdict.
	degraded := append([]CaseRun(nil), runs...)
	for i := range degraded {
		if degraded[i].CaseID == "install-req" {
			degraded[i].Output = "# no_match\n"
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
