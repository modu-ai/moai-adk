// Package harness — HRN-003 Rubric struct and Validate() method tests.
// REQ-HRN-003-003: 4 anchor levels (0.25, 0.50, 0.75, 1.00).
// REQ-HRN-003-013: anchor levels FROZEN.
package harness

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestRubric_AnchorLevelsValid verifies that a Rubric with the 4 canonical anchor
// levels passes Validate().
// REQ-HRN-003-003, REQ-HRN-003-013.
func TestRubric_AnchorLevelsValid(t *testing.T) {
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
	if err := rubric.Validate(); err != nil {
		t.Errorf("Rubric.Validate() = %v, want nil", err)
	}
}

// TestRubric_AnchorLevelsRejectFifth verifies that ErrInvalidConfig is returned
// when there are 5 anchors.
// REQ-HRN-003-013.
func TestRubric_AnchorLevelsRejectFifth(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test",
		PassThreshold: 0.60,
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: {
				Weight:        0.40,
				PassThreshold: 0.60,
				// Add a 5th anchor — must be rejected.
				Anchors: map[float64]string{
					0.25: "low",
					0.50: "medium",
					0.75: "high",
					1.00: "perfect",
					0.10: "extra anchor — invalid",
				},
			},
			Security:    makeValidDimensionRubric(0.60),
			Craft:       makeValidDimensionRubric(0.60),
			Consistency: makeValidDimensionRubric(0.60),
		},
	}
	err := rubric.Validate()
	if err == nil {
		t.Fatal("Rubric.Validate() = nil, want error for 5 anchors")
	}
	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Errorf("Rubric.Validate() error = %v, want wrapping ErrInvalidConfig", err)
	}
}

// TestRubric_AnchorLevelsRejectNonCanonical verifies that ErrInvalidConfig is
// returned when non-canonical anchor values are included.
// REQ-HRN-003-013: values such as {0.20, 0.40, 0.60, 0.80} are rejected.
func TestRubric_AnchorLevelsRejectNonCanonical(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test",
		PassThreshold: 0.60,
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: {
				Weight:        0.40,
				PassThreshold: 0.60,
				// Non-canonical anchor values
				Anchors: map[float64]string{
					0.20: "low",
					0.40: "medium",
					0.60: "high",
					0.80: "perfect",
				},
			},
			Security:    makeValidDimensionRubric(0.60),
			Craft:       makeValidDimensionRubric(0.60),
			Consistency: makeValidDimensionRubric(0.60),
		},
	}
	err := rubric.Validate()
	if err == nil {
		t.Fatal("Rubric.Validate() = nil, want error for non-canonical anchors")
	}
	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Errorf("Rubric.Validate() error = %v, want wrapping ErrInvalidConfig", err)
	}
}

// TestParseRubricMarkdown_DefaultProfile verifies that loading the default.md
// profile yields 4 dimensions, each with 4 anchors.
// REQ-HRN-003-005, AC-HRN-003-07.a.
func TestParseRubricMarkdown_DefaultProfile(t *testing.T) {
	rubric, err := ParseRubricMarkdown("../../.moai/config/evaluator-profiles/default.md")
	if err != nil {
		t.Fatalf("ParseRubricMarkdown(default.md) = %v, want nil", err)
	}

	// Must have 4 dimensions.
	if len(rubric.Dimensions) != 4 {
		t.Errorf("rubric.Dimensions count = %d, want 4", len(rubric.Dimensions))
	}

	// Each dimension must have 4 anchors.
	for dim, dr := range rubric.Dimensions {
		if len(dr.Anchors) != 4 {
			t.Errorf("dimension %v has %d anchors, want 4", dim, len(dr.Anchors))
		}
		// Verify canonical anchors.
		for anchor := range dr.Anchors {
			if !canonicalAnchors[anchor] {
				t.Errorf("dimension %v has non-canonical anchor %.2f", dim, anchor)
			}
		}
	}

	// MustPass must include Functionality and Security.
	hasFunctionality, hasSecurity := false, false
	for _, d := range rubric.MustPass {
		if d == Functionality {
			hasFunctionality = true
		}
		if d == Security {
			hasSecurity = true
		}
	}
	if !hasFunctionality || !hasSecurity {
		t.Errorf("MustPass = %v, want to include Functionality and Security", rubric.MustPass)
	}

	// Verify PassThreshold floor.
	if rubric.PassThreshold < 0.60 {
		t.Errorf("PassThreshold = %.2f, want >= 0.60", rubric.PassThreshold)
	}
}

// TestParseRubricMarkdown_AllProfiles verifies that all 4 shipping profile files
// load without error.
// AC-HRN-003-07.b.
func TestParseRubricMarkdown_AllProfiles(t *testing.T) {
	profiles := []string{
		"../../.moai/config/evaluator-profiles/default.md",
		"../../.moai/config/evaluator-profiles/strict.md",
		"../../.moai/config/evaluator-profiles/lenient.md",
		"../../.moai/config/evaluator-profiles/frontend.md",
	}
	for _, path := range profiles {
		t.Run(path, func(t *testing.T) {
			rubric, err := ParseRubricMarkdown(path)
			if err != nil {
				t.Fatalf("ParseRubricMarkdown(%q) = %v, want nil", path, err)
			}
			if rubric == nil {
				t.Fatal("ParseRubricMarkdown returned nil rubric")
			}
		})
	}
}

// TestRubric_ValidateCitation verifies validation of the RubricAnchor field on
// SubCriterionScore.
// REQ-HRN-003-009, AC-HRN-003-05.
func TestRubric_ValidateCitation(t *testing.T) {
	rubric := &Rubric{
		ProfileName: "test",
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}

	// AC-HRN-003-05.a: missing RubricAnchor → ErrRubricCitationMissing.
	emptyAnchor := SubCriterionScore{
		Score:        0.75,
		RubricAnchor: "", // empty value
		Evidence:     "test",
		Dimension:    Functionality,
	}
	err := rubric.ValidateCitation(emptyAnchor)
	if err == nil {
		t.Fatal("ValidateCitation(empty anchor) = nil, want ErrRubricCitationMissing")
	}
	if !errors.Is(err, config.ErrRubricCitationMissing) {
		t.Errorf("ValidateCitation(empty anchor) = %v, want ErrRubricCitationMissing", err)
	}

	// AC-HRN-003-05.b: non-canonical anchor → ErrRubricCitationMissing.
	nonCanonical := SubCriterionScore{
		Score:        0.65,
		RubricAnchor: "0.65", // non-canonical
		Evidence:     "test",
		Dimension:    Functionality,
	}
	err = rubric.ValidateCitation(nonCanonical)
	if err == nil {
		t.Fatal("ValidateCitation(non-canonical anchor 0.65) = nil, want ErrRubricCitationMissing")
	}
	if !errors.Is(err, config.ErrRubricCitationMissing) {
		t.Errorf("ValidateCitation(non-canonical anchor) = %v, want ErrRubricCitationMissing", err)
	}

	// AC-HRN-003-05.c: canonical anchor → nil (success).
	valid := SubCriterionScore{
		Score:        0.75,
		RubricAnchor: "0.75", // canonical
		Evidence:     "test",
		Dimension:    Functionality,
	}
	if err = rubric.ValidateCitation(valid); err != nil {
		t.Errorf("ValidateCitation(canonical anchor 0.75) = %v, want nil", err)
	}
}

// TestRubric_PassThresholdFloor verifies that ErrInvalidConfig is returned
// when pass_threshold < 0.60.
// REQ-HRN-003-014, AC-HRN-003-12.
func TestRubric_PassThresholdFloor(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-low-threshold",
		PassThreshold: 0.55, // below 0.60 — must be rejected.
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security},
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}
	err := rubric.Validate()
	if err == nil {
		t.Fatal("Rubric.Validate() = nil, want error for pass_threshold < 0.60")
	}
	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Errorf("Rubric.Validate() error = %v, want wrapping ErrInvalidConfig", err)
	}
}

// TestRubric_MustPassNonNarrowing verifies that removing Security from MustPass
// returns ErrMustPassBypassProhibited.
// REQ-HRN-003-018, AC-HRN-003-11.
func TestRubric_MustPassNonNarrowing(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-bypass",
		PassThreshold: 0.60,
		Aggregation:   "min",
		// Attempt to violate the must-pass floor by removing Security.
		MustPass: []Dimension{Functionality}, // Security missing — must be rejected.
		Dimensions: map[Dimension]DimensionRubric{
			Functionality: makeValidDimensionRubric(0.60),
			Security:      makeValidDimensionRubric(0.60),
			Craft:         makeValidDimensionRubric(0.60),
			Consistency:   makeValidDimensionRubric(0.60),
		},
	}
	err := rubric.Validate()
	if err == nil {
		t.Fatal("Rubric.Validate() = nil, want error for MustPass narrowing (Security removed)")
	}
	if !errors.Is(err, config.ErrMustPassBypassProhibited) {
		t.Errorf("Rubric.Validate() error = %v, want wrapping ErrMustPassBypassProhibited", err)
	}
}

// makeValidDimensionRubric is a helper that builds a valid DimensionRubric with
// the 4 canonical anchors.
func makeValidDimensionRubric(passThreshold float64) DimensionRubric {
	return DimensionRubric{
		Weight:        0.25,
		PassThreshold: passThreshold,
		Anchors: map[float64]string{
			0.25: "low quality",
			0.50: "medium quality",
			0.75: "high quality",
			1.00: "perfect quality",
		},
	}
}
