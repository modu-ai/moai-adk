// Package harness — HRN-003 Hierarchical Acceptance Scoring implementation.
// 4-dimensional (Functionality / Security / Craft / Consistency) × sub-criteria hierarchical scoring.
// design-constitution §12 Mechanism 1 (Rubric Anchoring) + Mechanism 3 (Must-Pass Firewall).
package harness

// @MX:NOTE: [AUTO] FROZEN at 4 values per spec.md REQ-HRN-003-001; 5th dimension requires CON-002 amendment
// @MX:WARN: [AUTO] FROZEN-zone constraint; addition of 5th dimension requires CON-002 amendment per zone-registry CONST-V3R2-154
// @MX:REASON: design-constitution §12 Mechanism 3 + SPEC-V3R2-HRN-003 REQ-001 — Dimension enum is FROZEN at {Functionality, Security, Craft, Consistency} for v3.0

// Dimension is the 4-dimensional scoring enum for evaluator-active.
// FROZEN: holds exactly four values throughout v3.0.
// REQ-HRN-003-001.
type Dimension int

const (
	// Functionality is the feature-completeness dimension.
	Functionality Dimension = iota + 1
	// Security is the security dimension.
	Security
	// Craft is the code-quality dimension.
	Craft
	// Consistency is the consistency dimension.
	Consistency
)

// String returns the canonical name of the Dimension.
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

// IsValid verifies that the Dimension is one of the four canonical values.
// Returns false for any value outside the FROZEN 4-dimension set.
func (d Dimension) IsValid() bool {
	return d >= Functionality && d <= Consistency
}

// DefaultMustPassDimensions is the default must-pass dimension list.
// OQ3 default: [Functionality, Security]; floor-only [Security] (REQ-018 prevents narrowing).
var DefaultMustPassDimensions = []Dimension{Functionality, Security}

// Verdict is the final ScoreCard verdict.
type Verdict string

const (
	// VerdictPass indicates that every evaluation criterion has passed.
	VerdictPass Verdict = "pass"
	// VerdictFail indicates that at least one evaluation criterion did not pass.
	VerdictFail Verdict = "fail"
)

// SubCriterionScore is the scoring result for a single sub-criterion.
// REQ-HRN-003-002: includes the {Score, RubricAnchor, Evidence, Dimension} fields.
type SubCriterionScore struct {
	// Score is a value in the range 0.0~1.0.
	Score float64
	// RubricAnchor is the rubric anchor reference value (canonical: "0.25", "0.50", "0.75", "1.00").
	// REQ-HRN-003-009: empty or non-canonical values return ErrRubricCitationMissing.
	RubricAnchor string
	// Evidence is the rationale text supporting the score.
	Evidence string
	// Dimension is the dimension to which this sub-criterion belongs.
	Dimension Dimension
}

// CriterionScore is the scoring result for a single criterion.
// Aggregate is computed by aggregating the values in the SubCriteria map.
type CriterionScore struct {
	// Aggregate is the aggregated value of sub-criterion scores (min or mean).
	Aggregate float64
	// SubCriteria is the sub-criterion ID -> SubCriterionScore map.
	SubCriteria map[string]SubCriterionScore
}

// DimensionScore is the scoring result for a single Dimension.
// Aggregate is computed by aggregating the values in the Criteria map.
type DimensionScore struct {
	// Aggregate is the aggregated value of criterion scores (min or mean).
	Aggregate float64
	// Criteria is the criterion ID -> CriterionScore map.
	Criteria map[string]CriterionScore
}

// ScoreCard is the hierarchical struct holding the full scoring result for one SPEC artifact.
// REQ-HRN-003-002: Dimensions -> Criteria -> SubCriteria hierarchy.
// OQ4 default: SchemaVersion = "v1" (mirrors HRN-002 LogSchemaVersion pattern).
type ScoreCard struct {
	// SchemaVersion is the scoring-schema version ("v1").
	SchemaVersion string
	// SpecID is the SPEC ID being evaluated.
	SpecID string
	// Dimensions is the Dimension -> DimensionScore map.
	Dimensions map[Dimension]DimensionScore
	// Verdict is the final verdict (pass/fail).
	Verdict Verdict
	// Rationale is the rationale text for the verdict.
	// When the must-pass firewall triggers, specifies the failed dimension and threshold.
	Rationale string
}
