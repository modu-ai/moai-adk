// Package harness — HRN-003 M3 scoring engine tests.
// EvaluatorRunner.Score(), aggregation (min/mean), must-pass firewall, WriteContract().
// REQ-HRN-003-004: EvaluatorRunner.Score().
// REQ-HRN-003-007: aggregation rules.
// REQ-HRN-003-008: must-pass firewall.
// REQ-HRN-003-009: rubric citation enforcement.
// REQ-HRN-003-011: Sprint Contract sub-criterion persistence.
package harness

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// makeScoreCard builds a 2 Dimension × 3 Criterion × 2 SubCriterion ScoreCard fixture.
// AC-HRN-003-02: 12 SubCriterionScore entries.
func makeScoreCard(t *testing.T) *ScoreCard {
	t.Helper()
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "SPEC-V3R2-HRN-003",
		Dimensions:    make(map[Dimension]DimensionScore),
	}
	dims := []Dimension{Functionality, Security}
	for _, dim := range dims {
		ds := DimensionScore{Criteria: make(map[string]CriterionScore)}
		for ci := 1; ci <= 3; ci++ {
			critID := "crit-" + string(rune('0'+ci))
			cs := CriterionScore{SubCriteria: make(map[string]SubCriterionScore)}
			for si := 1; si <= 2; si++ {
				subID := critID + ".sub" + string(rune('0'+si))
				cs.SubCriteria[subID] = SubCriterionScore{
					Score:        0.75,
					RubricAnchor: "0.75",
					Evidence:     "fixture evidence",
					Dimension:    dim,
				}
			}
			ds.Criteria[critID] = cs
		}
		card.Dimensions[dim] = ds
	}
	return card
}

// TestEvaluatorRunner_ScoreAggregatesHierarchically verifies that EvaluatorRunner.Score
// hierarchically aggregates a 2×3×2 ScoreCard fixture and returns the correct Aggregate.
// REQ-HRN-003-004, AC-HRN-003-02.
func TestEvaluatorRunner_ScoreAggregatesHierarchically(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test",
		PassThreshold: 0.60,
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	card := makeScoreCard(t)
	runner := NewEvaluatorRunner(rubric)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("AggregateScoreCard() error = %v, want nil", err)
	}

	// Each Dimension's Aggregate must not remain at 0.
	for dim, ds := range result.Dimensions {
		if ds.Aggregate <= 0 {
			t.Errorf("dimension %v Aggregate = %f, want > 0", dim, ds.Aggregate)
		}
		for critID, cs := range ds.Criteria {
			if cs.Aggregate <= 0 {
				t.Errorf("dimension %v criterion %q Aggregate = %f, want > 0", dim, critID, cs.Aggregate)
			}
		}
	}
}

// TestAggregateMin verifies that aggregateMin returns the minimum value of a slice.
// REQ-HRN-003-007, AC-HRN-003-03.a.
func TestAggregateMin(t *testing.T) {
	scores := []float64{0.8, 0.5, 0.9}
	got := aggregateMin(scores)
	want := 0.5
	if got != want {
		t.Errorf("aggregateMin(%v) = %f, want %f", scores, got, want)
	}
}

// TestAggregateMean verifies that aggregateMean returns the mean of a slice.
// REQ-HRN-003-015, AC-HRN-003-03.b.
func TestAggregateMean(t *testing.T) {
	scores := []float64{0.8, 0.5}
	got := aggregateMean(scores)
	want := 0.65
	if got-want > 1e-9 || want-got > 1e-9 {
		t.Errorf("aggregateMean(%v) = %f, want %f", scores, got, want)
	}
}

// TestAggregateMin_Empty verifies that aggregateMin returns 0 for an empty slice.
// Edge Case E6.
func TestAggregateMin_Empty(t *testing.T) {
	got := aggregateMin(nil)
	if got != 0.0 {
		t.Errorf("aggregateMin(nil) = %f, want 0.0", got)
	}
}

// TestAggregateMean_Single verifies that aggregateMean returns the single value unchanged.
func TestAggregateMean_Single(t *testing.T) {
	got := aggregateMean([]float64{0.75})
	want := 0.75
	if got != want {
		t.Errorf("aggregateMean([0.75]) = %f, want %f", got, want)
	}
}

// TestMustPassFirewall_SecurityFails verifies that applyMustPassFirewall returns "fail"
// when Security.Aggregate < threshold.
// REQ-HRN-003-008, AC-HRN-003-04.a.i.
func TestMustPassFirewall_SecurityFails(t *testing.T) {
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.55}, // below the 0.60 floor
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	rubric := &Rubric{
		PassThreshold: 0.60,
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	verdict, rationale := applyMustPassFirewall(card, rubric)
	if verdict != VerdictFail {
		t.Errorf("applyMustPassFirewall() verdict = %q, want %q", verdict, VerdictFail)
	}
	// AC-HRN-003-04.a.ii: Rationale must include the failing dimension info.
	if rationale == "" {
		t.Error("applyMustPassFirewall() rationale is empty, want non-empty failure message")
	}
	// Rationale must contain "Security" and the failing score.
	if !containsAll(rationale, "Security", "0.55") {
		t.Errorf("applyMustPassFirewall() rationale = %q, want to contain 'Security' and '0.55'", rationale)
	}
}

// TestMustPassFirewall_FunctionalityFails verifies that applyMustPassFirewall returns "fail"
// when Functionality.Aggregate < threshold.
// AC-HRN-003-04.b.
func TestMustPassFirewall_FunctionalityFails(t *testing.T) {
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 0.55}, // this time Functionality fails
			Security:      {Aggregate: 0.95},
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	rubric := &Rubric{
		PassThreshold: 0.60,
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	verdict, rationale := applyMustPassFirewall(card, rubric)
	if verdict != VerdictFail {
		t.Errorf("applyMustPassFirewall() verdict = %q, want %q", verdict, VerdictFail)
	}
	if !containsAll(rationale, "Functionality") {
		t.Errorf("applyMustPassFirewall() rationale = %q, want to contain 'Functionality'", rationale)
	}
}

// TestMustPassFirewall_AllPass verifies that applyMustPassFirewall returns "pass"
// when every must-pass dimension is at or above the threshold.
// AC-HRN-003-04.c.
func TestMustPassFirewall_AllPass(t *testing.T) {
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 0.85},
			Security:      {Aggregate: 0.90},
			Craft:         {Aggregate: 0.70},
			Consistency:   {Aggregate: 0.80},
		},
	}
	rubric := &Rubric{
		PassThreshold: 0.60,
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	verdict, rationale := applyMustPassFirewall(card, rubric)
	if verdict != VerdictPass {
		t.Errorf("applyMustPassFirewall() verdict = %q, want %q", verdict, VerdictPass)
	}
	_ = rationale // rationale may be empty when pass.
}

// TestMustPassFirewall_StrictProfile verifies that with a rubric using a strict.md-level
// high threshold (0.85), 0.84 fails and 0.86 passes.
// REQ-HRN-003-014, AC-HRN-003-12.
func TestMustPassFirewall_StrictProfile(t *testing.T) {
	strictDimRubric := DimensionRubric{
		Weight:        0.30,
		PassThreshold: 0.85, // strict level
		Anchors: map[float64]string{
			0.25: "low",
			0.50: "medium",
			0.75: "high",
			1.00: "perfect",
		},
	}
	rubric := &Rubric{
		PassThreshold: 0.85,
		MustPass:      []Dimension{Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      strictDimRubric,
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	// 0.84: must fail.
	cardFail := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test-fail",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.84}, // below 0.85
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	verdict, _ := applyMustPassFirewall(cardFail, rubric)
	if verdict != VerdictFail {
		t.Errorf("strict profile 0.84: applyMustPassFirewall() = %q, want %q", verdict, VerdictFail)
	}

	// 0.86: must pass.
	cardPass := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test-pass",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.86}, // at or above 0.85
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	verdict, _ = applyMustPassFirewall(cardPass, rubric)
	if verdict != VerdictPass {
		t.Errorf("strict profile 0.86: applyMustPassFirewall() = %q, want %q", verdict, VerdictPass)
	}
}

// TestAggregateScoreCard_MinAggregation verifies that in "min" aggregation mode
// the criterion aggregate uses the minimum value.
// AC-HRN-003-03.a.
func TestAggregateScoreCard_MinAggregation(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-min",
		PassThreshold: 0.60,
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	// Functionality: crit-1: sub1=0.8, sub2=0.5 → criterion aggregate = 0.5 (min)
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {
				Criteria: map[string]CriterionScore{
					"crit-1": {
						SubCriteria: map[string]SubCriterionScore{
							"sub-1": {Score: 0.8, RubricAnchor: "0.75", Evidence: "e", Dimension: Functionality},
							"sub-2": {Score: 0.5, RubricAnchor: "0.50", Evidence: "e", Dimension: Functionality},
						},
					},
				},
			},
			Security:    {Criteria: map[string]CriterionScore{}},
			Craft:       {Criteria: map[string]CriterionScore{}},
			Consistency: {Criteria: map[string]CriterionScore{}},
		},
	}

	runner := NewEvaluatorRunner(rubric)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("AggregateScoreCard() error = %v", err)
	}

	funcDim := result.Dimensions[Functionality]
	crit := funcDim.Criteria["crit-1"]
	if crit.Aggregate != 0.5 {
		t.Errorf("min aggregation: crit-1 Aggregate = %f, want 0.5", crit.Aggregate)
	}
}

// TestAggregateScoreCard_MeanAggregation verifies that in "mean" aggregation mode
// the criterion aggregate uses the mean.
// AC-HRN-003-03.b.
func TestAggregateScoreCard_MeanAggregation(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-mean",
		PassThreshold: 0.60,
		Aggregation:   "mean",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	// Functionality: crit-1: sub1=0.8, sub2=0.5 → criterion aggregate = 0.65 (mean)
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {
				Criteria: map[string]CriterionScore{
					"crit-1": {
						SubCriteria: map[string]SubCriterionScore{
							"sub-1": {Score: 0.8, RubricAnchor: "0.75", Evidence: "e", Dimension: Functionality},
							"sub-2": {Score: 0.5, RubricAnchor: "0.50", Evidence: "e", Dimension: Functionality},
						},
					},
				},
			},
			Security:    {Criteria: map[string]CriterionScore{}},
			Craft:       {Criteria: map[string]CriterionScore{}},
			Consistency: {Criteria: map[string]CriterionScore{}},
		},
	}

	runner := NewEvaluatorRunner(rubric)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("AggregateScoreCard() error = %v", err)
	}

	funcDim := result.Dimensions[Functionality]
	crit := funcDim.Criteria["crit-1"]
	want := 0.65
	if diff := crit.Aggregate - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("mean aggregation: crit-1 Aggregate = %f, want %f", crit.Aggregate, want)
	}
}

// TestAggregateScoreCard_PerDimOverride verifies that per-dimension aggregation overrides
// are applied correctly.
// AC-HRN-003-03.c.
func TestAggregateScoreCard_PerDimOverride(t *testing.T) {
	// Functionality: mean, Security: (profile default = min)
	rubric := &Rubric{
		ProfileName:   "test-override",
		PassThreshold: 0.60,
		Aggregation:   "min", // profile default
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: {
				Weight:        0.25,
				PassThreshold: 0.60,
				Aggregation:   "mean", // per-dimension override
				Anchors:       map[float64]string{0.25: "low", 0.50: "med", 0.75: "high", 1.00: "perfect"},
			},
			Security:    makeValidDimensionRubric(0.60),
			Craft:       makeValidDimensionRubric(0.60),
			Consistency: makeValidDimensionRubric(0.60),
		},
	}

	// Functionality: sub1=0.8, sub2=0.5 → mean=0.65
	// Security: sub1=0.8, sub2=0.5 → min=0.5
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {
				Criteria: map[string]CriterionScore{
					"crit-1": {
						SubCriteria: map[string]SubCriterionScore{
							"sub-1": {Score: 0.8, RubricAnchor: "0.75", Evidence: "e", Dimension: Functionality},
							"sub-2": {Score: 0.5, RubricAnchor: "0.50", Evidence: "e", Dimension: Functionality},
						},
					},
				},
			},
			Security: {
				Criteria: map[string]CriterionScore{
					"crit-1": {
						SubCriteria: map[string]SubCriterionScore{
							"sub-1": {Score: 0.8, RubricAnchor: "0.75", Evidence: "e", Dimension: Security},
							"sub-2": {Score: 0.5, RubricAnchor: "0.50", Evidence: "e", Dimension: Security},
						},
					},
				},
			},
			Craft:       {Criteria: map[string]CriterionScore{}},
			Consistency: {Criteria: map[string]CriterionScore{}},
		},
	}

	runner := NewEvaluatorRunner(rubric)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("AggregateScoreCard() error = %v", err)
	}

	// Functionality: mean → 0.65
	funcCrit := result.Dimensions[Functionality].Criteria["crit-1"]
	wantFunc := 0.65
	if diff := funcCrit.Aggregate - wantFunc; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("Functionality (mean) crit aggregate = %f, want %f", funcCrit.Aggregate, wantFunc)
	}

	// Security: min → 0.5
	secCrit := result.Dimensions[Security].Criteria["crit-1"]
	wantSec := 0.5
	if secCrit.Aggregate != wantSec {
		t.Errorf("Security (min) crit aggregate = %f, want %f", secCrit.Aggregate, wantSec)
	}
}

// TestWriteContract verifies that WriteContract persists the sub-criterion status
// correctly as YAML.
// REQ-HRN-003-011, AC-HRN-003-10.
func TestWriteContract(t *testing.T) {
	tmpDir := t.TempDir()
	contractPath := filepath.Join(tmpDir, "contract.yaml")

	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "SPEC-TEST-001",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {
				Aggregate: 0.75,
				Criteria: map[string]CriterionScore{
					"AC-001": {
						Aggregate: 0.75,
						SubCriteria: map[string]SubCriterionScore{
							"AC-001.a": {Score: 0.75, RubricAnchor: "0.75", Evidence: "passed", Dimension: Functionality},
							"AC-001.b": {Score: 0.50, RubricAnchor: "0.50", Evidence: "partial", Dimension: Functionality},
						},
					},
				},
			},
			Security: {
				Aggregate: 1.00,
				Criteria: map[string]CriterionScore{
					"AC-002": {
						Aggregate: 1.00,
						SubCriteria: map[string]SubCriterionScore{
							"AC-002.a": {Score: 1.00, RubricAnchor: "1.00", Evidence: "complete", Dimension: Security},
						},
					},
				},
			},
			Craft:       {Criteria: map[string]CriterionScore{}},
			Consistency: {Criteria: map[string]CriterionScore{}},
		},
		Verdict:   VerdictPass,
		Rationale: "all must-pass dimensions passed",
	}

	if err := WriteContract(card, contractPath); err != nil {
		t.Fatalf("WriteContract() error = %v, want nil", err)
	}

	// Verify the file was created.
	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", contractPath, err)
	}
	content := string(data)

	// YAML must contain the required fields.
	for _, want := range []string{"spec_id", "SPEC-TEST-001", "verdict", "pass", "acceptance_checklist"} {
		if !containsStr(content, want) {
			t.Errorf("WriteContract YAML missing %q in:\n%s", want, content)
		}
	}
}

// TestWriteContract_NoLeakDetection verifies that the YAML produced by WriteContract
// does not trigger the HRN-002 leak-detection pattern.
// AC-HRN-003-10: contract carries criterion state, not evaluator rationale.
func TestWriteContract_NoLeakDetection(t *testing.T) {
	tmpDir := t.TempDir()
	contractPath := filepath.Join(tmpDir, "contract.yaml")

	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "SPEC-LEAK-TEST",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Criteria: map[string]CriterionScore{}},
			Security:      {Criteria: map[string]CriterionScore{}},
			Craft:         {Criteria: map[string]CriterionScore{}},
			Consistency:   {Criteria: map[string]CriterionScore{}},
		},
		Verdict:   VerdictPass,
		Rationale: "all dimensions passed",
	}

	if err := WriteContract(card, contractPath); err != nil {
		t.Fatalf("WriteContract() error = %v", err)
	}

	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// HRN-002 leak detection: YAML contents must not contain forbidden substrings.
	// "Score:", "Feedback:", "Verdict:" are evaluator-judgment traces and are forbidden.
	// However, "verdict:" (lowercase) is allowed as a state field.
	if err := DetectPriorJudgmentLeak(string(data)); err != nil {
		t.Errorf("WriteContract YAML triggers leak detection: %v", err)
	}
}

// TestValidateCitation_EmptyAnchor_M3 re-verifies, at the M3 level, that ValidateCitation
// returns ErrRubricCitationMissing for an empty RubricAnchor.
// AC-HRN-003-05.a — also verified in M2; re-verified in the M3 engine context.
func TestValidateCitation_EmptyAnchor_M3(t *testing.T) {
	rubric := &Rubric{
		ProfileName: "test",
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}
	score := SubCriterionScore{
		Score:        0.75,
		RubricAnchor: "",
		Evidence:     "test",
		Dimension:    Functionality,
	}
	if err := rubric.ValidateCitation(score); !errors.Is(err, config.ErrRubricCitationMissing) {
		t.Errorf("ValidateCitation(empty anchor) = %v, want ErrRubricCitationMissing", err)
	}
}

// containsAll is a helper that checks whether s contains every substr.
func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !containsStr(s, sub) {
			return false
		}
	}
	return true
}

// containsStr is a helper that checks whether s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
