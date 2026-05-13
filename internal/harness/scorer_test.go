// Package harness — HRN-003 Hierarchical Acceptance Scoring 테스트.
// REQ-HRN-003-001: 4-차원 열거형 (Dimension enum).
// REQ-HRN-003-002: 계층 스코어카드 구조 (ScoreCard tree shape).
package harness

import (
	"testing"
)

// TestDimensionEnum_FrozenSet는 Dimension enum이 정확히 4개의 값을 가지고
// 범위 외의 값에 대해 IsValid()가 false를 반환하는지 검증합니다.
// REQ-HRN-003-001, AC-HRN-003-01.
func TestDimensionEnum_FrozenSet(t *testing.T) {
	// 4개의 canonical dimension만 유효해야 합니다.
	canonicals := []Dimension{Functionality, Security, Craft, Consistency}
	for _, d := range canonicals {
		if !d.IsValid() {
			t.Errorf("canonical dimension %v IsValid() = false, want true", d)
		}
	}

	// 범위 외의 값은 유효하지 않아야 합니다.
	if Dimension(99).IsValid() {
		t.Error("Dimension(99).IsValid() = true, want false")
	}
	if Dimension(0).IsValid() {
		t.Error("Dimension(0).IsValid() = true, want false (iota starts at 1)")
	}

	// String() 메서드 검증.
	if got := Functionality.String(); got != "Functionality" {
		t.Errorf("Functionality.String() = %q, want %q", got, "Functionality")
	}
	if got := Security.String(); got != "Security" {
		t.Errorf("Security.String() = %q, want %q", got, "Security")
	}
	if got := Craft.String(); got != "Craft" {
		t.Errorf("Craft.String() = %q, want %q", got, "Craft")
	}
	if got := Consistency.String(); got != "Consistency" {
		t.Errorf("Consistency.String() = %q, want %q", got, "Consistency")
	}
}

// TestScoreCard_HierarchicalShape는 ScoreCard가 2dim × 3crit × 2subcrit = 12개의
// SubCriterionScore 항목을 담을 수 있는 계층 구조를 가지는지 검증합니다.
// REQ-HRN-003-002, AC-HRN-003-02.
func TestScoreCard_HierarchicalShape(t *testing.T) {
	// 2 dimension × 3 criteria × 2 sub-criteria = 12 SubCriterionScore 항목을
	// 담을 수 있는 ScoreCard를 생성합니다.
	card := &ScoreCard{
		SchemaVersion: "v1",
		SpecID:        "SPEC-V3R2-HRN-003",
		Dimensions:    make(map[Dimension]DimensionScore),
	}

	dims := []Dimension{Functionality, Security}
	for _, dim := range dims {
		ds := DimensionScore{
			Aggregate: 0.0,
			Criteria:  make(map[string]CriterionScore),
		}
		for ci := 1; ci <= 3; ci++ {
			critID := "AC-HRN-003-0" + string(rune('0'+ci))
			cs := CriterionScore{
				Aggregate:    0.0,
				SubCriteria:  make(map[string]SubCriterionScore),
			}
			for si := 1; si <= 2; si++ {
				subID := critID + ".sub" + string(rune('0'+si))
				cs.SubCriteria[subID] = SubCriterionScore{
					Score:       0.75,
					RubricAnchor: "0.75",
					Evidence:    "fixture evidence",
					Dimension:   dim,
				}
			}
			ds.Criteria[critID] = cs
		}
		card.Dimensions[dim] = ds
	}

	// 총 12개 SubCriterionScore 항목이 있어야 합니다.
	total := 0
	for _, ds := range card.Dimensions {
		for _, cs := range ds.Criteria {
			total += len(cs.SubCriteria)
		}
	}
	if total != 12 {
		t.Errorf("total SubCriterionScore entries = %d, want 12", total)
	}

	// SchemaVersion 필드 검증.
	if card.SchemaVersion != "v1" {
		t.Errorf("ScoreCard.SchemaVersion = %q, want %q", card.SchemaVersion, "v1")
	}

	// DefaultMustPassDimensions 검증.
	if len(DefaultMustPassDimensions) != 2 {
		t.Errorf("DefaultMustPassDimensions len = %d, want 2", len(DefaultMustPassDimensions))
	}
}
