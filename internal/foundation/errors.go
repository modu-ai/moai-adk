package foundation

import (
	"errors"
	"fmt"
)

// Sentinel errors for the foundation package.
var (
	// ErrInvalidRequirementType indicates an unrecognized EARS requirement type.
	ErrInvalidRequirementType = errors.New("invalid requirement type")

	// ErrInvalidPillar indicates an unrecognized TRUST 5 pillar.
	ErrInvalidPillar = errors.New("invalid TRUST 5 pillar")

	// ErrAssessmentFailed indicates that the quality assessment did not meet thresholds.
	ErrAssessmentFailed = errors.New("assessment failed: not all pillars meet threshold")

	// ErrInvalidPhaseTransition indicates an invalid methodology phase transition.
	ErrInvalidPhaseTransition = errors.New("invalid phase transition")

	// ErrUnsupportedLanguage indicates a programming language that is not supported.
	ErrUnsupportedLanguage = errors.New("unsupported language")
)

// RequirementNotFoundError is returned when a requirement ID cannot be found.
type RequirementNotFoundError struct {
	ID string
}

// Error returns the error message including the requirement ID.
func (e *RequirementNotFoundError) Error() string {
	return fmt.Sprintf("requirement not found: %s", e.ID)
}

// LanguageNotFoundError is returned when a language lookup fails.
type LanguageNotFoundError struct {
	Query string
}

// Error returns the error message including the query that failed.
func (e *LanguageNotFoundError) Error() string {
	return fmt.Sprintf("language not found: %s", e.Query)
}
