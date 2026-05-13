// Package harness — HRN-003 Hierarchical Acceptance Scoring 구현.
// 4-차원(Functionality / Security / Craft / Consistency) × sub-criteria 계층 채점.
// design-constitution §12 Mechanism 1 (Rubric Anchoring) + Mechanism 3 (Must-Pass Firewall).
package harness

// @MX:NOTE: [AUTO] FROZEN at 4 values per spec.md REQ-HRN-003-001; 5th dimension requires CON-002 amendment
// @MX:WARN: [AUTO] FROZEN-zone constraint; addition of 5th dimension requires CON-002 amendment per zone-registry CONST-V3R2-154
// @MX:REASON: design-constitution §12 Mechanism 3 + SPEC-V3R2-HRN-003 REQ-001 — Dimension enum is FROZEN at {Functionality, Security, Craft, Consistency} for v3.0

// Dimension은 evaluator-active의 4-차원 채점 열거형입니다.
// FROZEN: v3.0 기간 동안 정확히 4개의 값을 가집니다.
// REQ-HRN-003-001.
type Dimension int

const (
	// Functionality는 기능 완성도 차원입니다.
	Functionality Dimension = iota + 1
	// Security는 보안 차원입니다.
	Security
	// Craft는 코드 품질 차원입니다.
	Craft
	// Consistency는 일관성 차원입니다.
	Consistency
)

// String은 Dimension의 canonical 이름을 반환합니다.
func (d Dimension) String() string {
	switch d {
	case Functionality:
		return "Functionality"
	case Security:
		return "Security"
	case Craft:
		return "Craft"
	case Consistency:
		return "Consistency"
	default:
		return "Unknown"
	}
}

// IsValid는 Dimension이 4개의 canonical 값 중 하나인지 검증합니다.
// FROZEN 4-dimension set 외의 값은 false를 반환합니다.
func (d Dimension) IsValid() bool {
	return d >= Functionality && d <= Consistency
}

// DefaultMustPassDimensions는 기본 must-pass 차원 목록입니다.
// OQ3 default: [Functionality, Security]; floor-only [Security] (REQ-018 prevents narrowing).
var DefaultMustPassDimensions = []Dimension{Functionality, Security}

// Verdict는 ScoreCard의 최종 판정 결과입니다.
type Verdict string

const (
	// VerdictPass는 모든 평가 기준을 통과한 경우입니다.
	VerdictPass Verdict = "pass"
	// VerdictFail은 하나 이상의 평가 기준을 통과하지 못한 경우입니다.
	VerdictFail Verdict = "fail"
)

// SubCriterionScore는 단일 sub-criterion에 대한 채점 결과입니다.
// REQ-HRN-003-002: {Score, RubricAnchor, Evidence, Dimension} 필드 포함.
type SubCriterionScore struct {
	// Score는 0.0~1.0 범위의 점수입니다.
	Score float64
	// RubricAnchor는 rubric anchor 참조값입니다 (canonical: "0.25", "0.50", "0.75", "1.00").
	// REQ-HRN-003-009: 빈 값 또는 non-canonical 값은 ErrRubricCitationMissing을 반환합니다.
	RubricAnchor string
	// Evidence는 채점 근거 설명입니다.
	Evidence string
	// Dimension은 이 sub-criterion이 속하는 차원입니다.
	Dimension Dimension
}

// CriterionScore는 단일 criterion에 대한 채점 결과입니다.
// SubCriteria map의 값들을 집계하여 Aggregate를 계산합니다.
type CriterionScore struct {
	// Aggregate는 sub-criterion 점수들의 집계값입니다 (min 또는 mean).
	Aggregate float64
	// SubCriteria는 sub-criterion ID → SubCriterionScore 맵입니다.
	SubCriteria map[string]SubCriterionScore
}

// DimensionScore는 단일 Dimension에 대한 채점 결과입니다.
// Criteria map의 값들을 집계하여 Aggregate를 계산합니다.
type DimensionScore struct {
	// Aggregate는 criterion 점수들의 집계값입니다 (min 또는 mean).
	Aggregate float64
	// Criteria는 criterion ID → CriterionScore 맵입니다.
	Criteria map[string]CriterionScore
}

// ScoreCard는 하나의 SPEC 아티팩트에 대한 전체 채점 결과를 담는 계층 구조입니다.
// REQ-HRN-003-002: Dimensions → Criteria → SubCriteria 계층 구조.
// OQ4 default: SchemaVersion = "v1" (mirrors HRN-002 LogSchemaVersion pattern).
type ScoreCard struct {
	// SchemaVersion은 채점 스키마 버전입니다 ("v1").
	SchemaVersion string
	// SpecID는 평가 대상 SPEC ID입니다.
	SpecID string
	// Dimensions는 Dimension → DimensionScore 맵입니다.
	Dimensions map[Dimension]DimensionScore
	// Verdict는 최종 판정 결과입니다 (pass/fail).
	Verdict Verdict
	// Rationale는 판정 근거 설명입니다.
	// must-pass firewall이 트리거된 경우 실패 차원과 임계값을 명시합니다.
	Rationale string
}
