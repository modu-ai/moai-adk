// Package harness — HRN-003 Rubric struct 및 Validate() 메서드 테스트.
// REQ-HRN-003-003: 4개 anchor level 검증 (0.25, 0.50, 0.75, 1.00).
// REQ-HRN-003-013: anchor level FROZEN.
package harness

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestRubric_AnchorLevelsValid는 4개의 canonical anchor level을 가진 Rubric이
// Validate()를 통과하는지 검증합니다.
// REQ-HRN-003-003, REQ-HRN-003-013.
func TestRubric_AnchorLevelsValid(t *testing.T) {
	rubric := &Rubric{
		ProfileName: "test",
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

// TestRubric_AnchorLevelsRejectFifth는 5개 anchor를 가진 경우 ErrInvalidConfig를
// 반환하는지 검증합니다.
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
				// 5번째 anchor 추가 — 거부되어야 합니다.
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

// TestRubric_AnchorLevelsRejectNonCanonical는 non-canonical anchor 값이 포함된 경우
// ErrInvalidConfig를 반환하는지 검증합니다.
// REQ-HRN-003-013: {0.20, 0.40, 0.60, 0.80} 등은 거부됩니다.
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
				// non-canonical anchor 값들
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

// TestParseRubricMarkdown_DefaultProfile는 default.md 프로필을 로드하여
// 4개의 차원과 각 차원에 4개의 anchor가 있는지 검증합니다.
// REQ-HRN-003-005, AC-HRN-003-07.a.
func TestParseRubricMarkdown_DefaultProfile(t *testing.T) {
	rubric, err := ParseRubricMarkdown("../../.moai/config/evaluator-profiles/default.md")
	if err != nil {
		t.Fatalf("ParseRubricMarkdown(default.md) = %v, want nil", err)
	}

	// 4개의 차원이 있어야 합니다.
	if len(rubric.Dimensions) != 4 {
		t.Errorf("rubric.Dimensions count = %d, want 4", len(rubric.Dimensions))
	}

	// 각 차원에 4개의 anchor가 있어야 합니다.
	for dim, dr := range rubric.Dimensions {
		if len(dr.Anchors) != 4 {
			t.Errorf("dimension %v has %d anchors, want 4", dim, len(dr.Anchors))
		}
		// canonical anchor 검증.
		for anchor := range dr.Anchors {
			if !canonicalAnchors[anchor] {
				t.Errorf("dimension %v has non-canonical anchor %.2f", dim, anchor)
			}
		}
	}

	// MustPass가 Functionality와 Security를 포함해야 합니다.
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

	// PassThreshold floor 검증.
	if rubric.PassThreshold < 0.60 {
		t.Errorf("PassThreshold = %.2f, want >= 0.60", rubric.PassThreshold)
	}
}

// TestParseRubricMarkdown_AllProfiles는 4개의 shipping 프로필 파일이 모두
// 오류 없이 로드되는지 검증합니다.
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

// TestRubric_ValidateCitation은 SubCriterionScore의 RubricAnchor 필드 검증을
// 확인합니다.
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

	// AC-HRN-003-05.a: RubricAnchor 필드가 없으면 ErrRubricCitationMissing.
	emptyAnchor := SubCriterionScore{
		Score:        0.75,
		RubricAnchor: "", // 빈 값
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

	// AC-HRN-003-05.b: non-canonical anchor 값은 ErrRubricCitationMissing.
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

	// AC-HRN-003-05.c: canonical anchor 값은 nil (성공).
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

// TestRubric_PassThresholdFloor는 pass_threshold < 0.60이면 ErrInvalidConfig를
// 반환하는지 검증합니다.
// REQ-HRN-003-014, AC-HRN-003-12.
func TestRubric_PassThresholdFloor(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-low-threshold",
		PassThreshold: 0.55, // 0.60 미만 — 거부되어야 합니다.
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

// TestRubric_MustPassNonNarrowing는 MustPass에서 Security를 제거하면
// ErrMustPassBypassProhibited를 반환하는지 검증합니다.
// REQ-HRN-003-018, AC-HRN-003-11.
func TestRubric_MustPassNonNarrowing(t *testing.T) {
	rubric := &Rubric{
		ProfileName:   "test-bypass",
		PassThreshold: 0.60,
		Aggregation:   "min",
		// Security를 제거하여 must-pass floor 위반을 시도합니다.
		MustPass: []Dimension{Functionality}, // Security 누락 — 거부되어야 합니다.
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

// makeValidDimensionRubric는 4개의 canonical anchor를 가진 유효한 DimensionRubric을
// 생성하는 헬퍼 함수입니다.
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
