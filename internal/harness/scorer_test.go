// Package harness — HRN-003 Hierarchical Acceptance Scoring tests.
// REQ-HRN-003-001: 4-dimension enumeration (Dimension enum).
// REQ-HRN-003-002: hierarchical scorecard structure (ScoreCard tree shape).
package harness

import (
	"testing"
)

// TestDimensionEnum_FrozenSet verifies that the Dimension enum has exactly 4 values
// and that IsValid() returns false for out-of-range values.
// REQ-HRN-003-001, AC-HRN-003-01.
func TestDimensionEnum_FrozenSet(t *testing.T) {
	// Only the 4 canonical dimensions must be valid.
	canonicals := []Dimension{Functionality, Security, Craft, Consistency}
	for _, d := range canonicals {
		if !d.IsValid() {
			t.Errorf("canonical dimension %v IsValid() = false, want true", d)
		}
	}

	// Out-of-range values must be invalid.
	if Dimension(99).IsValid() {
		t.Error("Dimension(99).IsValid() = true, want false")
	}
	if Dimension(0).IsValid() {
		t.Error("Dimension(0).IsValid() = true, want false (iota starts at 1)")
	}

	// Verify the String() method.
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

// TestScoreCard_HierarchicalShape verifies that ScoreCard has the hierarchical
// structure to hold 2dim × 3crit × 2subcrit = 12 SubCriterionScore entries.
// REQ-HRN-003-002, AC-HRN-003-02.
func TestScoreCard_HierarchicalShape(t *testing.T) {
	// Build a ScoreCard that can hold 2 dimension × 3 criteria × 2 sub-criteria
	// = 12 SubCriterionScore entries.
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

	// There must be 12 SubCriterionScore entries in total.
	total := 0
	for _, ds := range card.Dimensions {
		for _, cs := range ds.Criteria {
			total += len(cs.SubCriteria)
		}
	}
	if total != 12 {
		t.Errorf("total SubCriterionScore entries = %d, want 12", total)
	}

	// Verify the SchemaVersion field.
	if card.SchemaVersion != "v1" {
		t.Errorf("ScoreCard.SchemaVersion = %q, want %q", card.SchemaVersion, "v1")
	}

	// Verify DefaultMustPassDimensions.
	if len(DefaultMustPassDimensions) != 2 {
		t.Errorf("DefaultMustPassDimensions len = %d, want 2", len(DefaultMustPassDimensions))
	}
}
