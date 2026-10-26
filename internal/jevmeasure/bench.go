package jevmeasure

// bench.go — SPEC-FETCH-WEB-001 M6 (REQ-FW-014). Bench cases in the
// jevfetch shape — a question, an expected quote, a page snapshot — plus the
// run/compare/baseline surface the `moai jevmeasure bench` command rides.
// The runner reuses the fetchweb pipeline over the SNAPSHOT (no page fetch)
// with an injected Answerer, so no test or offline run can reach a real
// endpoint (REQ-FW-010). Upstream's numeric bench claims are NOT asserted
// anywhere (spec.md §D8): this harness reports containment and approximate
// token savings; it promises no figure.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
)

// BenchCase is one jevfetch-shaped bench case.
type BenchCase struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	ExpectedQuote string `json:"expected_quote"`
	PageSnapshot  string `json:"page_snapshot"`
}

// CaseRun is one case's recorded pipeline outcome.
type CaseRun struct {
	CaseID string `json:"case_id"`
	Status string `json:"status"`
	// Output is the pipeline's markdown render (status line + verbatim kept
	// blocks).
	Output string `json:"output"`
	// OutputTokens is the kept set's APPROXIMATE token count, recorded from
	// the pipeline result at run time (re-deriving it from the rendered
	// markdown would re-parse prose as HTML).
	OutputTokens int `json:"output_tokens"`
}

// LoadBenchCases reads a JSON file holding a case list.
func LoadBenchCases(path string) ([]BenchCase, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("jevmeasure: bench cases unreadable: %w", err)
	}
	var cases []BenchCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		return nil, fmt.Errorf("jevmeasure: bench cases not a JSON case list: %w", err)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("jevmeasure: bench case file %s holds no cases", path)
	}
	return cases, nil
}

// RunBench executes the fetchweb pipeline over each case's snapshot. Token
// figures in the comparison are APPROXIMATE (jev.EstimateTokens via
// fetchweb; see that package's note) — the savings rate is a ratio of two
// approximations, not a measured cost figure.
func RunBench(ctx context.Context, cases []BenchCase, ask fetchweb.Answerer) []CaseRun {
	runs := make([]CaseRun, 0, len(cases))
	for _, c := range cases {
		blocks := fetchweb.SplitBlocks(c.PageSnapshot)
		res := fetchweb.Process(ctx, blocks, c.Question, ask)
		runs = append(runs, CaseRun{
			CaseID:       c.ID,
			Status:       string(res.Status),
			Output:       fetchweb.RenderMarkdown(res),
			OutputTokens: res.KeptTokens,
		})
	}
	return runs
}

// CaseComparison is one case's containment + savings outcome.
type CaseComparison struct {
	CaseID string `json:"case_id"`
	// Contained reports whether the expected quote appears in the output.
	Contained bool `json:"contained"`
	// OutputTokens / SnapshotTokens are APPROXIMATE token counts.
	OutputTokens   int `json:"output_tokens"`
	SnapshotTokens int `json:"snapshot_tokens"`
	// SavingsRate is 1 - output/snapshot when the snapshot has tokens; 0 when
	// it has none (an empty snapshot cannot demonstrate savings).
	SavingsRate float64 `json:"savings_rate"`
}

// BenchReport is the set-level comparison over a run.
type BenchReport struct {
	Cases   []CaseComparison `json:"cases"`
	Summary struct {
		N               int     `json:"n"`
		Contained       int     `json:"contained"`
		MeanSavingsRate float64 `json:"mean_savings_rate"`
	} `json:"summary"`
}

// CompareRun computes per-case containment and token-savings versus each
// snapshot's token count, plus the set summary.
func CompareRun(cases []BenchCase, runs []CaseRun) BenchReport {
	runByID := make(map[string]CaseRun, len(runs))
	for _, r := range runs {
		runByID[r.CaseID] = r
	}
	var report BenchReport
	sum := 0.0
	for _, c := range cases {
		run := runByID[c.ID]
		snapTokens := fetchweb.DocTokens(fetchweb.SplitBlocks(c.PageSnapshot))
		cc := CaseComparison{
			CaseID:         c.ID,
			Contained:      c.ExpectedQuote != "" && strings.Contains(run.Output, c.ExpectedQuote),
			OutputTokens:   run.OutputTokens,
			SnapshotTokens: snapTokens,
		}
		if snapTokens > 0 {
			cc.SavingsRate = 1 - float64(run.OutputTokens)/float64(snapTokens)
		}
		report.Cases = append(report.Cases, cc)
		report.Summary.N++
		if cc.Contained {
			report.Summary.Contained++
		}
		sum += cc.SavingsRate
	}
	if n := len(report.Cases); n > 0 {
		report.Summary.MeanSavingsRate = sum / float64(n)
	}
	return report
}

// baselineDoc is the stored baseline.json: the comparison a recorded run
// produced, against which later runs diff deterministically.
type baselineDoc struct {
	Cases []CaseComparison `json:"cases"`
}

// WriteBaseline compares a run and stores it as the baseline at path.
func WriteBaseline(path string, cases []BenchCase, runs []CaseRun) (BenchReport, error) {
	report := CompareRun(cases, runs)
	raw, err := json.MarshalIndent(baselineDoc{Cases: report.Cases}, "", "  ")
	if err != nil {
		return report, err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return report, fmt.Errorf("jevmeasure: baseline write failed: %w", err)
	}
	return report, nil
}

// CaseDiff is one case's verdict against the stored baseline.
type CaseDiff struct {
	CaseID string `json:"case_id"`
	Pass   bool   `json:"pass"`
	// Regression marks a case the baseline recorded contained that the run
	// lost.
	Regression bool   `json:"regression"`
	Reason     string `json:"reason,omitempty"`
}

// BaselineDiff is the deterministic per-case diff of a run against a stored
// baseline. Determinism: the diff is a pure function of (baseline file, run)
// — no clock, no network, no float fuzz.
type BaselineDiff struct {
	Cases         []CaseDiff `json:"cases"`
	AnyRegression bool       `json:"any_regression"`
}

// DiffAgainstBaseline compares a run with the baseline at path. A case
// PASSES iff its containment verdict equals the baseline's verdict; a flip
// from contained to not-contained is a regression; a flip the other way
// still fails the equality (the baseline is stale for that case) without
// counting as a regression. A case missing from the baseline fails with the
// reason named, and so does a case with NO run result: the run map lookup
// would otherwise yield the zero value (not-contained, zero output tokens),
// which silently matched a baseline row that recorded not-contained.
func DiffAgainstBaseline(path string, cases []BenchCase, runs []CaseRun) (BaselineDiff, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return BaselineDiff{}, fmt.Errorf("jevmeasure: baseline unreadable: %w", err)
	}
	var base baselineDoc
	if err := json.Unmarshal(raw, &base); err != nil {
		return BaselineDiff{}, fmt.Errorf("jevmeasure: baseline not a JSON baseline: %w", err)
	}
	baseByID := make(map[string]CaseComparison, len(base.Cases))
	for _, c := range base.Cases {
		baseByID[c.CaseID] = c
	}
	hasRun := make(map[string]bool, len(runs))
	for _, r := range runs {
		hasRun[r.CaseID] = true
	}
	current := CompareRun(cases, runs)
	diff := BaselineDiff{}
	for _, cc := range current.Cases {
		d := CaseDiff{CaseID: cc.CaseID}
		b, ok := baseByID[cc.CaseID]
		switch {
		case !hasRun[cc.CaseID]:
			d.Pass = false
			d.Reason = "no run result for this case"
		case !ok:
			d.Pass = false
			d.Reason = "no baseline entry for this case"
		case cc.Contained == b.Contained:
			d.Pass = true
		case b.Contained && !cc.Contained:
			d.Pass = false
			d.Regression = true
			d.Reason = "expected quote was contained at baseline and is now absent"
			diff.AnyRegression = true
		default:
			d.Pass = false
			d.Reason = "containment appeared since baseline; re-baseline to accept it"
		}
		diff.Cases = append(diff.Cases, d)
	}
	return diff, nil
}
