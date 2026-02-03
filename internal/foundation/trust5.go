package foundation

import (
	"fmt"
	"math"
)

// Pillar represents one of the five TRUST 5 quality pillars.
type Pillar string

const (
	// Tested ensures comprehensive test coverage (85%+ target).
	Tested Pillar = "tested"

	// Readable ensures clear naming, structure, and documentation.
	Readable Pillar = "readable"

	// Understandable ensures documentation completeness and manageable code complexity.
	Understandable Pillar = "understandable"

	// Secured ensures OWASP compliance, input validation, and security scanning.
	Secured Pillar = "secured"

	// Trackable ensures traceable changes via conventional commits and SPEC references.
	Trackable Pillar = "trackable"
)

// String returns the string representation of the Pillar.
func (p Pillar) String() string {
	return string(p)
}

// IsValid checks whether the Pillar is one of the defined constants.
func (p Pillar) IsValid() bool {
	switch p {
	case Tested, Readable, Understandable, Secured, Trackable:
		return true
	}
	return false
}

// AllPillars returns all five TRUST 5 pillars in canonical order.
func AllPillars() []Pillar {
	return []Pillar{Tested, Readable, Understandable, Secured, Trackable}
}

// PillarStatus represents the evaluation status of a pillar.
type PillarStatus string

const (
	// StatusPass indicates the pillar meets or exceeds the threshold.
	StatusPass PillarStatus = "pass"

	// StatusWarning indicates the pillar is below threshold but above critical.
	StatusWarning PillarStatus = "warning"

	// StatusCritical indicates the pillar is critically below threshold.
	StatusCritical PillarStatus = "critical"
)

// String returns the string representation of the PillarStatus.
func (s PillarStatus) String() string {
	return string(s)
}

// passThreshold is the minimum score (inclusive) for a pillar to pass.
const passThreshold = 0.85

// warningThreshold is the minimum score (inclusive) to avoid critical status.
const warningThreshold = 0.5

// PillarScore holds the evaluation result for a single TRUST 5 pillar.
type PillarScore struct {
	Pillar Pillar       `json:"pillar"`
	Score  float64      `json:"score"`
	Status PillarStatus `json:"status"`
	Issues []string     `json:"issues,omitempty"`
}

// Assessment aggregates all five TRUST 5 pillar scores.
type Assessment struct {
	Scores map[Pillar]*PillarScore `json:"scores"`
}

// NewAssessment creates a new Assessment with all five pillars initialized
// to zero score and critical status.
func NewAssessment() *Assessment {
	scores := make(map[Pillar]*PillarScore, 5)
	for _, p := range AllPillars() {
		scores[p] = &PillarScore{
			Pillar: p,
			Score:  0.0,
			Status: StatusCritical,
			Issues: []string{},
		}
	}
	return &Assessment{Scores: scores}
}

// SetScore sets the score and issues for a specific pillar.
// Score must be between 0.0 and 1.0 inclusive.
// Returns ErrInvalidPillar if the pillar is not recognized.
func (a *Assessment) SetScore(pillar Pillar, score float64, issues []string) error {
	if !pillar.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidPillar, pillar)
	}
	if score < 0.0 || score > 1.0 {
		return fmt.Errorf("score must be between 0.0 and 1.0, got %f", score)
	}
	if issues == nil {
		issues = []string{}
	}
	a.Scores[pillar] = &PillarScore{
		Pillar: pillar,
		Score:  score,
		Issues: issues,
	}
	return nil
}

// Evaluate updates the status of each pillar based on score thresholds.
// A score >= 0.85 is Pass, >= 0.50 is Warning, below 0.50 is Critical.
func (a *Assessment) Evaluate() {
	for _, ps := range a.Scores {
		switch {
		case ps.Score >= passThreshold:
			ps.Status = StatusPass
		case ps.Score >= warningThreshold:
			ps.Status = StatusWarning
		default:
			ps.Status = StatusCritical
		}
	}
}

// OverallStatus returns the worst (most severe) status across all pillars.
// Critical > Warning > Pass in severity.
func (a *Assessment) OverallStatus() PillarStatus {
	worst := StatusPass
	for _, ps := range a.Scores {
		if ps.Status == StatusCritical {
			return StatusCritical
		}
		if ps.Status == StatusWarning {
			worst = StatusWarning
		}
	}
	return worst
}

// IsPass returns true if all pillars have a score >= 0.85.
func (a *Assessment) IsPass() bool {
	for _, ps := range a.Scores {
		if ps.Score < passThreshold {
			return false
		}
	}
	return true
}

// OverallScore returns the average score across all five pillars.
// Returns 0.0 if no scores are present.
func (a *Assessment) OverallScore() float64 {
	if len(a.Scores) == 0 {
		return 0.0
	}
	var total float64
	for _, ps := range a.Scores {
		total += ps.Score
	}
	result := total / float64(len(a.Scores))
	// Round to 4 decimal places to avoid floating-point noise.
	return math.Round(result*10000) / 10000
}
