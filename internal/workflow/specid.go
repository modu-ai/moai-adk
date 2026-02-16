package workflow

import "fmt"

// ValidateSpecID checks whether specID matches the expected SPEC-ISSUE-{number}
// format. Returns ErrInvalidSPECID wrapped with context if the format is invalid.
func ValidateSpecID(specID string) error {
	if !specIDPattern.MatchString(specID) {
		return fmt.Errorf("SPEC ID %q: %w", specID, ErrInvalidSPECID)
	}
	return nil
}
