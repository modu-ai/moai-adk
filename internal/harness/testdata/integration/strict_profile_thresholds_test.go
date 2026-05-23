// Package integration_test — HRN-003 integration tests: strict profile threshold verification.
// AC-HRN-003-12: strict profile per-dimension threshold (0.80) consistency.
// REQ-HRN-003-008: must-pass firewall — every must-pass dimension must be at or above the threshold to pass.
package integration_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// makeStrictRubric constructs a Rubric matching the strict profile directly.
// strict profile: all 4 dimensions individually must be at or above 0.80 to pass.
func makeStrictRubric() *harness.Rubric {
	anchors := map[float64]string{
		0.25: "기준 미달",
		0.50: "기준 부분 충족",
		0.75: "기준 대부분 충족",
		1.00: "기준 완전 충족",
	}
	return &harness.Rubric{
		ProfileName: "strict",
		Dimensions: map[harness.Dimension]harness.DimensionRubric{
			harness.Functionality: {Weight: 0.35, PassThreshold: 0.80, Anchors: anchors, Aggregation: "min"},
			harness.Security:      {Weight: 0.35, PassThreshold: 0.80, Anchors: anchors, Aggregation: "min"},
			harness.Craft:         {Weight: 0.20, PassThreshold: 0.80, Anchors: anchors, Aggregation: "min"},
			harness.Consistency:   {Weight: 0.10, PassThreshold: 0.80, Anchors: anchors, Aggregation: "min"},
		},
		PassThreshold: 0.80,
		MustPass:      []harness.Dimension{harness.Functionality, harness.Security},
		Aggregation:   "min",
	}
}

// makeCardWithDimScore builds a ScoreCard where every dimension has the same score.
func makeCardWithDimScore(score float64) *harness.ScoreCard {
	dims := []harness.Dimension{
		harness.Functionality,
		harness.Security,
		harness.Craft,
		harness.Consistency,
	}
	dimensions := make(map[harness.Dimension]harness.DimensionScore, len(dims))
	for _, d := range dims {
		dimensions[d] = harness.DimensionScore{
			Criteria: map[string]harness.CriterionScore{
				"crit-main": {
					SubCriteria: map[string]harness.SubCriterionScore{
						"sub-1": {
							Score:        score,
							RubricAnchor: "0.75",
							Evidence:     "테스트 픽스처 sub-criterion",
							Dimension:    d,
						},
					},
				},
			},
		}
	}
	return &harness.ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "TEST-STRICT-THRESHOLD",
		Dimensions:    dimensions,
	}
}

// TestStrictProfile_BelowThreshold verifies that under the strict profile a score of 0.79 fails.
// AC-HRN-003-12.a: score 0.79 → fail (must-pass Functionality + Security both below 0.80).
func TestStrictProfile_BelowThreshold(t *testing.T) {
	rubric := makeStrictRubric()
	runner := harness.NewEvaluatorRunner(rubric)

	card := makeCardWithDimScore(0.79)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != harness.VerdictFail {
		t.Errorf("score 0.79 below threshold 0.80: expected VerdictFail, got %q (rationale: %s)",
			result.Verdict, result.Rationale)
	}
}

// TestStrictProfile_AtThreshold verifies that under the strict profile a score of 0.80 passes.
// AC-HRN-003-12.b: score 0.80 → pass (must-pass dimensions exactly at the threshold).
func TestStrictProfile_AtThreshold(t *testing.T) {
	rubric := makeStrictRubric()
	runner := harness.NewEvaluatorRunner(rubric)

	card := makeCardWithDimScore(0.80)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != harness.VerdictPass {
		t.Errorf("score 0.80 at threshold 0.80: expected VerdictPass, got %q (rationale: %s)",
			result.Verdict, result.Rationale)
	}
}

// TestStrictProfile_AboveThreshold verifies that under the strict profile a score of 0.90 passes.
// AC-HRN-003-12.c: score 0.90 → pass (must-pass dimensions above the threshold).
func TestStrictProfile_AboveThreshold(t *testing.T) {
	rubric := makeStrictRubric()
	runner := harness.NewEvaluatorRunner(rubric)

	card := makeCardWithDimScore(0.90)
	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != harness.VerdictPass {
		t.Errorf("score 0.90 above threshold 0.80: expected VerdictPass, got %q (rationale: %s)",
			result.Verdict, result.Rationale)
	}
}
