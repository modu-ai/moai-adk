// Package integration_test — HRN-003 통합 테스트: strict 프로필 임계값 검증.
// AC-HRN-003-12: strict 프로필 per-dimension 임계값 (0.80) 정합성.
// REQ-HRN-003-008: must-pass firewall — 모든 must-pass 차원이 임계값 이상이어야 통과.
package integration_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// makeStrictRubric은 strict 프로필에 상응하는 Rubric을 직접 구성합니다.
// strict 프로필: 모든 4개 차원이 개별적으로 0.80 이상이어야 통과.
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

// makeCardWithDimScore는 모든 차원에 동일한 점수를 가진 ScoreCard를 생성합니다.
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

// TestStrictProfile_BelowThreshold는 strict 프로필에서 0.79 점수가 fail인지 검증합니다.
// AC-HRN-003-12.a: score 0.79 → fail (must-pass Functionality + Security 모두 0.80 미만).
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

// TestStrictProfile_AtThreshold는 strict 프로필에서 0.80 점수가 pass인지 검증합니다.
// AC-HRN-003-12.b: score 0.80 → pass (must-pass 차원이 정확히 임계값 도달).
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

// TestStrictProfile_AboveThreshold는 strict 프로필에서 0.90 점수가 pass인지 검증합니다.
// AC-HRN-003-12.c: score 0.90 → pass (must-pass 차원이 임계값 초과).
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
