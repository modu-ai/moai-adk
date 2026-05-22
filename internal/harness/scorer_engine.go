// Package harness — HRN-003 scoring-engine implementation.
// EvaluatorRunner: hierarchical aggregation + must-pass firewall + Sprint Contract persistence.
// REQ-HRN-003-004: EvaluatorRunner.AggregateScoreCard().
// REQ-HRN-003-007: aggregation rules (min default, mean opt-in).
// REQ-HRN-003-008: must-pass firewall.
// REQ-HRN-003-011: WriteContract() persists Sprint Contract sub-criterion status.
package harness

// @MX:NOTE: [AUTO] aggregation default = min per REQ-HRN-003-007; mean opt-in per REQ-HRN-003-015; per-dimension override supported
// @MX:WARN: [AUTO] must-pass firewall FROZEN per design-constitution §12 Mechanism 3; bypass requires CON-002 amendment
// @MX:REASON: design-constitution §12 Mechanism 3 — "Must-pass criteria cannot be compensated by high scores in other areas" (FROZEN per zone-registry CONST-V3R2-154)

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// EvaluatorRunner performs ScoreCard hierarchical aggregation + must-pass firewall enforcement.
// REQ-HRN-003-004: scored via Score() or AggregateScoreCard().
//
// @MX:ANCHOR: [AUTO] EvaluatorRunner.AggregateScoreCard — fan_in=3 (scorer_engine_test + SKILL.md Phase 3b + future HRN-001 router)
// @MX:REASON: SPEC-V3R2-HRN-003 REQ-004 — single entry point of the scoring engine; referenced by scorer_test.go, SKILL.md, and future routers
type EvaluatorRunner struct {
	rubric *Rubric
}

// NewEvaluatorRunner creates a new EvaluatorRunner with the given Rubric.
func NewEvaluatorRunner(rubric *Rubric) *EvaluatorRunner {
	return &EvaluatorRunner{rubric: rubric}
}

// AggregateScoreCard performs hierarchical aggregation
// (SubCriterionScore -> CriterionScore -> DimensionScore -> ScoreCard) and decides the Verdict.
// REQ-HRN-003-004, REQ-HRN-003-007, REQ-HRN-003-008.
//
// Aggregation rules:
//   - CriterionScore.Aggregate = aggregateMin/Mean(SubCriterionScore.Score) per dim.Aggregation or rubric.Aggregation
//   - DimensionScore.Aggregate = aggregateMin/Mean(CriterionScore.Aggregate) — dimension-level aggregation
//   - Verdict = applyMustPassFirewall(card, rubric)
func (r *EvaluatorRunner) AggregateScoreCard(card *ScoreCard) (*ScoreCard, error) {
	if card == nil {
		return nil, fmt.Errorf("AggregateScoreCard: card is nil")
	}

	// REQ-HRN-003-017: reject flat ScoreCard.
	// flat = the Dimensions map exists but no Dimension has Criteria,
	// or Criteria exist but no Criterion has SubCriteria.
	if isFlat(card) {
		return nil, config.ErrFlatScoreCardProhibited
	}

	result := &ScoreCard{
		SchemaVersion: card.SchemaVersion,
		SpecID:        card.SpecID,
		Dimensions:    make(map[Dimension]DimensionScore, len(card.Dimensions)),
	}

	for dim, ds := range card.Dimensions {
		// Determine per-dimension aggregation.
		agg := r.effectiveAggregation(dim)

		newDS := DimensionScore{
			Criteria: make(map[string]CriterionScore, len(ds.Criteria)),
		}

		// Aggregate CriterionScore.
		critAggregates := make([]float64, 0, len(ds.Criteria))
		for critID, cs := range ds.Criteria {
			subScores := make([]float64, 0, len(cs.SubCriteria))
			for _, sub := range cs.SubCriteria {
				subScores = append(subScores, sub.Score)
			}

			aggregator := selectAggregator(agg)
			critAggregate := aggregator.Aggregate(subScores)

			newCS := CriterionScore{
				Aggregate:   critAggregate,
				SubCriteria: cs.SubCriteria,
			}
			newDS.Criteria[critID] = newCS
			critAggregates = append(critAggregates, critAggregate)
		}

		// Aggregate DimensionScore.
		aggregator := selectAggregator(agg)
		dimAggregate := aggregator.Aggregate(critAggregates)
		newDS.Aggregate = dimAggregate
		result.Dimensions[dim] = newDS
	}

	// Apply the must-pass firewall.
	verdict, rationale := applyMustPassFirewall(result, r.rubric)
	result.Verdict = verdict
	result.Rationale = rationale

	return result, nil
}

// effectiveAggregation returns the effective aggregation mode, honoring per-dimension overrides.
// AC-HRN-003-03.c: uses the per-dimension override when present, otherwise the profile default.
func (r *EvaluatorRunner) effectiveAggregation(dim Dimension) string {
	if r.rubric == nil {
		return "min"
	}
	if dr, ok := r.rubric.Dimensions[dim]; ok && dr.Aggregation != "" {
		return dr.Aggregation
	}
	if r.rubric.Aggregation != "" {
		return r.rubric.Aggregation
	}
	return "min" // REQ-HRN-003-007 default.
}

// Aggregator is the strategy interface that reduces a score slice to a single aggregate value.
// T3.13 REFACTOR: extracts the aggregation strategy into an explicit interface to allow extension.
type Aggregator interface {
	Aggregate(scores []float64) float64
}

// MinAggregator is the minimum aggregation strategy.
// REQ-HRN-003-007 default aggregation.
type MinAggregator struct{}

// Aggregate returns the minimum value of the slice.
func (MinAggregator) Aggregate(scores []float64) float64 { return aggregateMin(scores) }

// MeanAggregator is the mean aggregation strategy.
// REQ-HRN-003-015 opt-in aggregation.
type MeanAggregator struct{}

// Aggregate returns the mean of the slice.
func (MeanAggregator) Aggregate(scores []float64) float64 { return aggregateMean(scores) }

// selectAggregator returns an Aggregator based on the aggregation mode string.
func selectAggregator(agg string) Aggregator {
	if agg == "mean" {
		return MeanAggregator{}
	}
	return MinAggregator{} // Default "min".
}

// aggregateMin returns the minimum value of the slice.
// Returns 0.0 for an empty slice (Edge Case E6).
// REQ-HRN-003-007, AC-HRN-003-03.a.
func aggregateMin(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	min := scores[0]
	for _, s := range scores[1:] {
		if s < min {
			min = s
		}
	}
	return min
}

// aggregateMean returns the mean of the slice.
// Returns 0.0 for an empty slice.
// REQ-HRN-003-015, AC-HRN-003-03.b.
func aggregateMean(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	var sum float64
	for _, s := range scores {
		sum += s
	}
	return sum / float64(len(scores))
}

// applyMustPassFirewall verifies that each must-pass dimension's Aggregate
// is at or above the per-dimension PassThreshold and returns the Verdict.
// REQ-HRN-003-008, AC-HRN-003-04.
//
// Rules:
//   - Aggregate < DimensionRubric.PassThreshold for any must-pass dimension -> VerdictFail.
//   - All must-pass dimensions pass -> VerdictPass.
//   - Rationale specifies the failed dimension and score (R6 user-friendliness easing).
func applyMustPassFirewall(card *ScoreCard, rubric *Rubric) (Verdict, string) {
	if rubric == nil || len(rubric.MustPass) == 0 {
		return VerdictPass, ""
	}

	var failedDims []string

	for _, dim := range rubric.MustPass {
		ds, exists := card.Dimensions[dim]
		if !exists {
			// Treat missing dimension as 0 -> failure.
			failedDims = append(failedDims, fmt.Sprintf("%s (missing, 0.00 < threshold)", dim))
			continue
		}

		// Check the per-dimension PassThreshold.
		threshold := rubric.PassThreshold // Default: profile-wide threshold.
		if dr, ok := rubric.Dimensions[dim]; ok && dr.PassThreshold > 0 {
			threshold = dr.PassThreshold
		}

		if ds.Aggregate < threshold {
			failedDims = append(failedDims, fmt.Sprintf(
				"must-pass dimension %s failed (%.2f < %.2f)",
				dim, ds.Aggregate, threshold,
			))
		}
	}

	if len(failedDims) > 0 {
		rationale := strings.Join(failedDims, "; ")
		return VerdictFail, rationale
	}

	return VerdictPass, ""
}

// ─────────────────────────────────────────────
// Sprint Contract persistence
// ─────────────────────────────────────────────

// ContractItem is an acceptance_checklist entry in the Sprint Contract YAML.
// REQ-HRN-003-011: includes the status field.
type ContractItem struct {
	ID        string  `yaml:"id"`
	Criterion string  `yaml:"criterion"`
	Evidence  string  `yaml:"evidence,omitempty"`
	Score     float64 `yaml:"score"`
	Anchor    string  `yaml:"rubric_anchor"`
	// Status is the sub-criterion status: "passed | failed | refined | new".
	// HRN-002 §11.4.1 Sprint Contract durability — passed items carry forward to the next iteration.
	Status string `yaml:"status"`
}

// SprintContractYAML is the structure of the YAML document produced by WriteContract.
// Adds a sub-criterion status field on top of the HRN-002 Sprint Contract shape.
// Remains compatible with the existing SKILL.md acceptance_checklist format.
type SprintContractYAML struct {
	SpecID             string         `yaml:"spec_id"`
	SchemaVersion      string         `yaml:"schema_version"`
	Verdict            string         `yaml:"verdict"`
	Rationale          string         `yaml:"rationale,omitempty"`
	AcceptanceChecklist []ContractItem `yaml:"acceptance_checklist"`
}

// WriteContract persists the ScoreCard's sub-criterion scores and statuses as a Sprint Contract YAML.
// REQ-HRN-003-011, AC-HRN-003-10.
//
// Auto-creates the directory of the file path when it does not exist.
// AC-HRN-003-10: the generated YAML does not trigger the HRN-002 DetectPriorJudgmentLeak pattern.
// (Only includes criterion state; no evaluator judgment rationale.)
func WriteContract(card *ScoreCard, path string) error {
	if card == nil {
		return fmt.Errorf("WriteContract: card is nil")
	}

	// Auto-create the directory.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("WriteContract: mkdir %q: %w", dir, err)
	}

	// Build acceptance_checklist: iterate by dimension > criterion > sub-criterion.
	var items []ContractItem

	// Sort dimensions for deterministic output.
	dims := sortedDimensions(card.Dimensions)
	for _, dim := range dims {
		ds := card.Dimensions[dim]
		critIDs := sortedStringKeys(ds.Criteria)
		for _, critID := range critIDs {
			cs := ds.Criteria[critID]
			subIDs := sortedStringKeys(cs.SubCriteria)
			for _, subID := range subIDs {
				sub := cs.SubCriteria[subID]
				status := subCriterionStatus(sub.Score, sub.RubricAnchor)
				items = append(items, ContractItem{
					ID:        subID,
					Criterion: critID,
					Evidence:  sub.Evidence,
					Score:     sub.Score,
					Anchor:    sub.RubricAnchor,
					Status:    status,
				})
			}
		}
	}

	// YAML serialization (simple direct formatting without the standard library).
	// Implements YAML serialization without external dependencies.
	content := buildContractYAML(card, items)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("WriteContract: write %q: %w", path, err)
	}

	return nil
}

// subCriterionStatus determines the status from a sub-criterion's score and anchor.
// status: "passed" (>= 0.75), "failed" (< 0.50), "refined" (0.50 <= score < 0.75).
func subCriterionStatus(score float64, _ string) string {
	switch {
	case score >= 0.75:
		return "passed"
	case score < 0.50:
		return "failed"
	default:
		return "refined"
	}
}

// buildContractYAML produces the SprintContract YAML content as a string.
// Performs direct formatting without an external YAML library.
// HARD (CLAUDE.local.md §14): hardcoding forbidden — all values are injected via parameters.
func buildContractYAML(card *ScoreCard, items []ContractItem) string {
	var sb strings.Builder
	sb.WriteString("# Sprint Contract — HRN-003 sub-criterion status\n")
	sb.WriteString("# format: HRN-003 WriteContract v1\n")
	sb.WriteString("# HRN-002 §11.4.1: passed criteria carry forward; evaluator memory is per-iteration\n\n")
	fmt.Fprintf(&sb, "spec_id: %q\n", card.SpecID)
	fmt.Fprintf(&sb, "schema_version: %q\n", card.SchemaVersion)
	fmt.Fprintf(&sb, "verdict: %q\n", string(card.Verdict))
	if card.Rationale != "" {
		fmt.Fprintf(&sb, "rationale: %q\n", card.Rationale)
	}
	sb.WriteString("\nacceptance_checklist:\n")
	for _, item := range items {
		fmt.Fprintf(&sb, "  - id: %q\n", item.ID)
		fmt.Fprintf(&sb, "    criterion: %q\n", item.Criterion)
		if item.Evidence != "" {
			fmt.Fprintf(&sb, "    evidence: %q\n", item.Evidence)
		}
		fmt.Fprintf(&sb, "    score: %.2f\n", item.Score)
		fmt.Fprintf(&sb, "    rubric_anchor: %q\n", item.Anchor)
		fmt.Fprintf(&sb, "    status: %q\n", item.Status)
	}
	return sb.String()
}

// sortedDimensions returns the ScoreCard's Dimension keys in sorted order.
// Used to produce deterministic YAML output.
func sortedDimensions(dims map[Dimension]DimensionScore) []Dimension {
	result := make([]Dimension, 0, len(dims))
	for d := range dims {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return int(result[i]) < int(result[j])
	})
	return result
}

// sortedStringKeys returns the keys of map[string]T in sorted order.
func sortedStringKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// isFlat checks whether the ScoreCard is in flat (non-hierarchical) form.
// REQ-HRN-003-017: flat ScoreCard is forbidden.
// Flat condition: no SubCriterion exists anywhere in the ScoreCard.
//   - Dimensions map is empty, or
//   - no Criterion in any Dimension contains SubCriteria.
// "Partially populated" cards with SubCriteria in only some dimensions are permitted.
func isFlat(card *ScoreCard) bool {
	if len(card.Dimensions) == 0 {
		return true
	}
	for _, ds := range card.Dimensions {
		for _, cs := range ds.Criteria {
			if len(cs.SubCriteria) > 0 {
				// Treated as hierarchical as soon as at least one SubCriterion exists.
				return false
			}
		}
	}
	// Flat when no Criterion contains SubCriteria.
	return true
}
