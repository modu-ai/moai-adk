// Package integration_test — HRN-003 통합 테스트: flat ScoreCard 거부 검증.
// REQ-HRN-003-017: ScoreCard must be hierarchical; flat shape rejected.
// AC-HRN-003-02: 12개 SubCriterionScore 항목 계층 구조 필수.
package integration_test

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness"
)

// makeTestRubric은 통합 테스트용 최소 유효 Rubric을 생성합니다.
func makeTestRubric() *harness.Rubric {
	return &harness.Rubric{
		ProfileName: "test",
		Dimensions: map[harness.Dimension]harness.DimensionRubric{
			harness.Functionality: {
				Weight:        0.40,
				PassThreshold: 0.75,
				Anchors: map[float64]string{
					0.25: "기능 미흡",
					0.50: "기능 부분 충족",
					0.75: "기능 대부분 충족",
					1.00: "기능 완전 충족",
				},
				Aggregation: "min",
			},
			harness.Security: {
				Weight:        0.40,
				PassThreshold: 0.75,
				Anchors: map[float64]string{
					0.25: "보안 취약",
					0.50: "보안 부분 충족",
					0.75: "보안 대부분 충족",
					1.00: "보안 완전 충족",
				},
				Aggregation: "min",
			},
		},
		PassThreshold: 0.75,
		MustPass:      []harness.Dimension{harness.Functionality, harness.Security},
		Aggregation:   "min",
	}
}

// TestFlatScoreCardRejected는 flat ScoreCard 입력 시 ErrFlatScoreCardProhibited가
// 반환되는지 검증합니다.
// REQ-HRN-003-017: flat-scorecard-prohibited 강제.
func TestFlatScoreCardRejected(t *testing.T) {
	// flat ScoreCard: Dimensions 맵은 있지만 SubCriteria가 없는 구조.
	flatCard := &harness.ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "TEST-FLAT-001",
		Dimensions:    map[harness.Dimension]harness.DimensionScore{},
		Verdict:       harness.VerdictFail,
		Rationale:     "flat shape — no sub-criteria",
	}

	rubric := makeTestRubric()
	runner := harness.NewEvaluatorRunner(rubric)

	_, err := runner.AggregateScoreCard(flatCard)
	if err == nil {
		t.Fatal("expected ErrFlatScoreCardProhibited, got nil")
	}
	if !errors.Is(err, config.ErrFlatScoreCardProhibited) {
		t.Errorf("expected errors.Is(err, ErrFlatScoreCardProhibited), got: %v", err)
	}
}

// TestHierarchicalScoreCardAccepted는 올바른 계층 ScoreCard가 수락되는지 검증합니다.
// REQ-HRN-003-017 positive case.
func TestHierarchicalScoreCardAccepted(t *testing.T) {
	// 계층 ScoreCard: Dimensions → Criteria → SubCriteria 구조 완비.
	card := &harness.ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "TEST-HIER-001",
		Dimensions: map[harness.Dimension]harness.DimensionScore{
			harness.Functionality: {
				Criteria: map[string]harness.CriterionScore{
					"crit-a": {
						SubCriteria: map[string]harness.SubCriterionScore{
							"sub-1": {
								Score:        0.85,
								RubricAnchor: "0.75",
								Evidence:     "기능 구현 확인",
								Dimension:    harness.Functionality,
							},
						},
					},
				},
			},
			harness.Security: {
				Criteria: map[string]harness.CriterionScore{
					"crit-sec": {
						SubCriteria: map[string]harness.SubCriterionScore{
							"sub-sec-1": {
								Score:        0.90,
								RubricAnchor: "1.00",
								Evidence:     "보안 검증 완료",
								Dimension:    harness.Security,
							},
						},
					},
				},
			},
		},
		Verdict:   harness.VerdictPass,
		Rationale: "hierarchical shape accepted",
	}

	rubric := makeTestRubric()
	runner := harness.NewEvaluatorRunner(rubric)

	result, err := runner.AggregateScoreCard(card)
	if err != nil {
		t.Fatalf("unexpected error for hierarchical card: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil ScoreCard result")
	}
}
