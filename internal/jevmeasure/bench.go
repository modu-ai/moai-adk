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
	// blocks) — the display surface.
	Output string `json:"output"`
	// KeptText is the join of the pipeline's kept block texts — what the
	// containment comparison measures. The markdown render's status line and
	// per-block heading lines are presentation, not kept text: an
	// ExpectedQuote equal to a status name ("no_match") must not count as
	// contained with zero excerpts (P2, card t1968 codex review).
	KeptText string `json:"kept_text"`
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
		keptTexts := make([]string, len(res.Blocks))
		for i, b := range res.Blocks {
			keptTexts[i] = b.Text
		}
		runs = append(runs, CaseRun{
			CaseID:       c.ID,
			Status:       string(res.Status),
			Output:       fetchweb.RenderMarkdown(res),
			KeptText:     strings.Join(keptTexts, "\n\n"),
			OutputTokens: res.KeptTokens,
		})
	}
	return runs
}

// CaseComparison is one case's containment + savings outcome.
type CaseComparison struct {
	CaseID string `json:"case_id"`
	// Contained reports whether the expected quote appears in the run's KEPT
	// BLOCK TEXT (CaseRun.KeptText) — not in the rendered markdown, whose
	// status line an ExpectedQuote equal to a status name would otherwise
	// match with zero excerpts kept (P2, card t1968 codex review).
	Contained bool `json:"contained"`
	// Missing marks a case with NO run result (L3, t1968 leader audit): the
	// run map lookup would otherwise yield a zero CaseRun — not
	// fetch_error, zero output tokens — that aggregates a bogus 100%
	// savings into the row and the mean. A missing run is a comparison
	// failure named by id, never a perfect run; it carries no savings
	// figure and is excluded from the summary mean (C4/iter6 semantics).
	Missing bool `json:"missing,omitempty"`
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
// snapshot's token count, plus the set summary. A run that ended in
// fetch_error demonstrates no savings — its zero kept tokens would otherwise
// read as a bogus 100% rate — so it carries no savings figure and is
// excluded from the summary mean (C4, t1968 sync-audit iter6). A case with
// NO run result is the same class of non-comparison: its row is marked
// missing and excluded from the mean (L3, t1968 leader audit) — the
// baseline-diff layer additionally FAILs it ("no run result for this case").
func CompareRun(cases []BenchCase, runs []CaseRun) BenchReport {
	runByID := make(map[string]CaseRun, len(runs))
	for _, r := range runs {
		runByID[r.CaseID] = r
	}
	var report BenchReport
	sum := 0.0
	comparable := 0
	for _, c := range cases {
		run, has := runByID[c.ID]
		snapTokens := fetchweb.DocTokens(fetchweb.SplitBlocks(c.PageSnapshot))
		// A missing run result is a failed comparison, not a perfect one:
		// the zero CaseRun's empty status is not fetch_error and its zero
		// output tokens would read as a bogus 100% savings (L3).
		failed := !has || run.Status == string(fetchweb.StatusFetchError)
		cc := CaseComparison{
			CaseID:         c.ID,
			Missing:        !has,
			Contained:      c.ExpectedQuote != "" && strings.Contains(run.KeptText, c.ExpectedQuote),
			OutputTokens:   run.OutputTokens,
			SnapshotTokens: snapTokens,
		}
		if snapTokens > 0 && !failed {
			cc.SavingsRate = 1 - float64(run.OutputTokens)/float64(snapTokens)
		}
		report.Cases = append(report.Cases, cc)
		report.Summary.N++
		if cc.Contained {
			report.Summary.Contained++
		}
		if !failed {
			sum += cc.SavingsRate
			comparable++
		}
	}
	if comparable > 0 {
		report.Summary.MeanSavingsRate = sum / float64(comparable)
	}
	return report
}

// baselineFormatKeptText is the format marker WriteBaseline stamps on every
// baseline it writes. DiffAgainstBaseline refuses a baseline whose marker is
// absent (written by a pre-marker build) or unrecognized: those verdicts were
// computed under the pre-23d02c65e3 containment semantics — rendered markdown,
// the maskable measure — and carry no kept-text provenance, so an equality
// PASS against them can mask a real containment regression (F3, card t1968
// iter15 codex gate; operator-approved fail-closed direction).
const baselineFormatKeptText = "kept-text-1"

// baselineDoc is the stored baseline.json: the comparison a recorded run
// produced, against which later runs diff deterministically.
type baselineDoc struct {
	Format string           `json:"format"`
	Cases  []CaseComparison `json:"cases"`
}

// WriteBaseline compares a run and stores it as the baseline at path,
// stamped with the kept-text format marker (F3, card t1968) so a later diff
// can tell this build's provenance from an older build's. A baseline is a
// recorded measurement, so a run set that is not one is refused outright
// (P2, card t1968 codex review): a run that ended in fetch_error stored its
// empty output like a legit measurement (Contained=false), and a later run
// of the same case then compared PASS against the poisoned baseline.
// Refusal mirrors the CompareRun/L3 semantics — failed or missing is not
// comparable, never a verdict — and covers both directions of the
// case↔run correspondence: a case with no run result (L3) and a run naming
// no bench case (an entry the diff, which walks cases, could never compare).
// On refusal NOTHING is written and the returned report is the zero value.
func WriteBaseline(path string, cases []BenchCase, runs []CaseRun) (BenchReport, error) {
	runByID := make(map[string]CaseRun, len(runs))
	for _, r := range runs {
		runByID[r.CaseID] = r
	}
	caseByID := make(map[string]bool, len(cases))
	for _, c := range cases {
		caseByID[c.ID] = true
		if _, has := runByID[c.ID]; !has {
			return BenchReport{}, fmt.Errorf("jevmeasure: baseline refused: case %s has no run result", c.ID)
		}
	}
	for _, r := range runs {
		if !caseByID[r.CaseID] {
			return BenchReport{}, fmt.Errorf("jevmeasure: baseline refused: run %s has no bench case", r.CaseID)
		}
		if r.Status == string(fetchweb.StatusFetchError) {
			return BenchReport{}, fmt.Errorf("jevmeasure: baseline refused: run %s ended in fetch_error", r.CaseID)
		}
	}
	report := CompareRun(cases, runs)
	raw, err := json.MarshalIndent(baselineDoc{Format: baselineFormatKeptText, Cases: report.Cases}, "", "  ")
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

// DiffAgainstBaseline compares a run with the baseline at path. The baseline
// must carry this build's kept-text format marker: a baseline whose marker is
// absent or unrecognized is REFUSED outright — error, empty diff, no
// comparison (F3, card t1968 iter15 codex gate) — because its verdicts were
// computed under containment semantics this diff cannot attribute, and an
// equality PASS against them can mask a real containment regression. A case
// PASSES iff its containment verdict equals the baseline's verdict; a flip
// from contained to not-contained is a regression; a flip the other way
// still fails the equality (the baseline is stale for that case) without
// counting as a regression. A case missing from the baseline fails with the
// reason named, and so does a case with NO run result: the run map lookup
// would otherwise yield the zero value (not-contained, zero output tokens),
// which silently matched a baseline row that recorded not-contained. A run
// that ended in fetch_error FAILs unconditionally (C4, t1968 sync-audit
// iter6): an error output contains nothing, so its Contained=false would
// otherwise pass the equality against a not-contained baseline row — "could
// not judge" is not a verdict match.
func DiffAgainstBaseline(path string, cases []BenchCase, runs []CaseRun) (BaselineDiff, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return BaselineDiff{}, fmt.Errorf("jevmeasure: baseline unreadable: %w", err)
	}
	var base baselineDoc
	if err := json.Unmarshal(raw, &base); err != nil {
		return BaselineDiff{}, fmt.Errorf("jevmeasure: baseline not a JSON baseline: %w", err)
	}
	if base.Format != baselineFormatKeptText {
		return BaselineDiff{}, fmt.Errorf("jevmeasure: baseline format predates kept-text containment; re-baseline with the current build")
	}
	baseByID := make(map[string]CaseComparison, len(base.Cases))
	for _, c := range base.Cases {
		baseByID[c.CaseID] = c
	}
	runByID := make(map[string]CaseRun, len(runs))
	for _, r := range runs {
		runByID[r.CaseID] = r
	}
	current := CompareRun(cases, runs)
	diff := BaselineDiff{}
	for _, cc := range current.Cases {
		d := CaseDiff{CaseID: cc.CaseID}
		b, ok := baseByID[cc.CaseID]
		run, has := runByID[cc.CaseID]
		switch {
		case !has:
			d.Pass = false
			d.Reason = "no run result for this case"
		case !ok:
			d.Pass = false
			d.Reason = "no baseline entry for this case"
		case run.Status == string(fetchweb.StatusFetchError):
			d.Pass = false
			d.Reason = "run ended in fetch_error; containment is not comparable (could not judge, not a no-match)"
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
