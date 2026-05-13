// Package harness — HRN-003 M3 채점 엔진 테스트.
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

// makeScoreCard는 2 Dimension × 3 Criterion × 2 SubCriterion ScoreCard 픽스처를 생성합니다.
// AC-HRN-003-02: 12개 SubCriterionScore 항목.
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

// TestEvaluatorRunner_ScoreAggregatesHierarchically는 EvaluatorRunner.Score가
// 2×3×2 픽스처 ScoreCard를 계층 집계하여 올바른 Aggregate를 반환하는지 검증합니다.
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

	// 각 Dimension의 Aggregate가 0으로 남지 않아야 합니다.
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

// TestAggregateMin는 aggregateMin이 슬라이스의 최솟값을 반환하는지 검증합니다.
// REQ-HRN-003-007, AC-HRN-003-03.a.
func TestAggregateMin(t *testing.T) {
	scores := []float64{0.8, 0.5, 0.9}
	got := aggregateMin(scores)
	want := 0.5
	if got != want {
		t.Errorf("aggregateMin(%v) = %f, want %f", scores, got, want)
	}
}

// TestAggregateMean는 aggregateMean이 슬라이스의 평균을 반환하는지 검증합니다.
// REQ-HRN-003-015, AC-HRN-003-03.b.
func TestAggregateMean(t *testing.T) {
	scores := []float64{0.8, 0.5}
	got := aggregateMean(scores)
	want := 0.65
	if got-want > 1e-9 || want-got > 1e-9 {
		t.Errorf("aggregateMean(%v) = %f, want %f", scores, got, want)
	}
}

// TestAggregateMin_Empty는 aggregateMin이 빈 슬라이스에 대해 0을 반환하는지 검증합니다.
// Edge Case E6.
func TestAggregateMin_Empty(t *testing.T) {
	got := aggregateMin(nil)
	if got != 0.0 {
		t.Errorf("aggregateMin(nil) = %f, want 0.0", got)
	}
}

// TestAggregateMean_Single는 aggregateMean이 단일 값에 대해 그 값을 그대로 반환하는지 검증합니다.
func TestAggregateMean_Single(t *testing.T) {
	got := aggregateMean([]float64{0.75})
	want := 0.75
	if got != want {
		t.Errorf("aggregateMean([0.75]) = %f, want %f", got, want)
	}
}

// TestMustPassFirewall_SecurityFails는 Security.Aggregate < threshold일 때
// applyMustPassFirewall이 "fail"을 반환하는지 검증합니다.
// REQ-HRN-003-008, AC-HRN-003-04.a.i.
func TestMustPassFirewall_SecurityFails(t *testing.T) {
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.55}, // 0.60 floor 이하
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
	// AC-HRN-003-04.a.ii: Rationale에 실패 차원 정보가 포함되어야 합니다.
	if rationale == "" {
		t.Error("applyMustPassFirewall() rationale is empty, want non-empty failure message")
	}
	// Rationale에 "Security"와 실패 점수가 포함되어야 합니다.
	if !containsAll(rationale, "Security", "0.55") {
		t.Errorf("applyMustPassFirewall() rationale = %q, want to contain 'Security' and '0.55'", rationale)
	}
}

// TestMustPassFirewall_FunctionalityFails는 Functionality.Aggregate < threshold일 때
// applyMustPassFirewall이 "fail"을 반환하는지 검증합니다.
// AC-HRN-003-04.b.
func TestMustPassFirewall_FunctionalityFails(t *testing.T) {
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 0.55}, // 이번엔 Functionality가 실패
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

// TestMustPassFirewall_AllPass는 모든 must-pass 차원이 threshold 이상일 때
// applyMustPassFirewall이 "pass"를 반환하는지 검증합니다.
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
	_ = rationale // pass 시 rationale은 비어 있어도 됩니다.
}

// TestMustPassFirewall_StrictProfile은 strict.md 수준의 높은 threshold (0.85)를
// 사용하는 rubric에서 0.84가 실패하고 0.86이 통과하는지 검증합니다.
// REQ-HRN-003-014, AC-HRN-003-12.
func TestMustPassFirewall_StrictProfile(t *testing.T) {
	strictDimRubric := DimensionRubric{
		Weight:        0.30,
		PassThreshold: 0.85, // strict 수준
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

	// 0.84: 실패해야 합니다.
	cardFail := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test-fail",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.84}, // 0.85 미만
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	verdict, _ := applyMustPassFirewall(cardFail, rubric)
	if verdict != VerdictFail {
		t.Errorf("strict profile 0.84: applyMustPassFirewall() = %q, want %q", verdict, VerdictFail)
	}

	// 0.86: 통과해야 합니다.
	cardPass := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "test-pass",
		Dimensions: map[Dimension]DimensionScore{
			Functionality: {Aggregate: 1.00},
			Security:      {Aggregate: 0.86}, // 0.85 이상
			Craft:         {Aggregate: 1.00},
			Consistency:   {Aggregate: 1.00},
		},
	}
	verdict, _ = applyMustPassFirewall(cardPass, rubric)
	if verdict != VerdictPass {
		t.Errorf("strict profile 0.86: applyMustPassFirewall() = %q, want %q", verdict, VerdictPass)
	}
}

// TestAggregateScoreCard_MinAggregation은 "min" aggregation 모드에서
// criterion aggregate가 최솟값을 사용하는지 검증합니다.
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

// TestAggregateScoreCard_MeanAggregation은 "mean" aggregation 모드에서
// criterion aggregate가 평균을 사용하는지 검증합니다.
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

// TestAggregateScoreCard_PerDimOverride는 차원별 aggregation override가
// 올바르게 적용되는지 검증합니다.
// AC-HRN-003-03.c.
func TestAggregateScoreCard_PerDimOverride(t *testing.T) {
	// Functionality: mean, Security: (profile default = min)
	rubric := &Rubric{
		ProfileName:   "test-override",
		PassThreshold: 0.60,
		Aggregation:   "min", // 프로필 기본값
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: {
				Weight:        0.25,
				PassThreshold: 0.60,
				Aggregation:   "mean", // 차원별 override
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

// TestWriteContract는 WriteContract가 sub-criterion status를 YAML로 올바르게
// 저장하는지 검증합니다.
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

	// 파일이 생성되었는지 확인합니다.
	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", contractPath, err)
	}
	content := string(data)

	// YAML에 필수 필드가 포함되어야 합니다.
	for _, want := range []string{"spec_id", "SPEC-TEST-001", "verdict", "pass", "acceptance_checklist"} {
		if !containsStr(content, want) {
			t.Errorf("WriteContract YAML missing %q in:\n%s", want, content)
		}
	}
}

// TestWriteContract_NoLeakDetection는 WriteContract가 생성한 YAML이
// HRN-002 leak detection 패턴을 트리거하지 않는지 검증합니다.
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

	// HRN-002 leak detection: YAML 내용이 forbidden substring을 포함하지 않아야 합니다.
	// "Score:", "Feedback:", "Verdict:" 는 evaluator 판단 흔적이므로 금지됩니다.
	// 단, "verdict:" (lowercase)는 상태 필드로 허용됩니다.
	if err := DetectPriorJudgmentLeak(string(data)); err != nil {
		t.Errorf("WriteContract YAML triggers leak detection: %v", err)
	}
}

// TestValidateCitation_EmptyAnchor_M3은 M3 레벨에서 ValidateCitation이
// 빈 RubricAnchor에 대해 ErrRubricCitationMissing을 반환하는지 재확인합니다.
// AC-HRN-003-05.a — M2에서도 검증되었으나 M3 engine 맥락에서 재확인.
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

// containsAll는 s가 모든 substr을 포함하는지 확인하는 헬퍼입니다.
func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !containsStr(s, sub) {
			return false
		}
	}
	return true
}

// containsStr는 s가 substr을 포함하는지 확인하는 헬퍼입니다.
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
