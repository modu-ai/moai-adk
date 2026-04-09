package workflow

import "fmt"

// ValidateSpecID checks whether specID matches the expected SPEC-ISSUE-{number}
// format. Returns ErrInvalidSPECID wrapped with context if the format is invalid.
func ValidateSpecID(specID string) error {
	if !specIssueIDPattern.MatchString(specID) {
		return fmt.Errorf("SPEC ID %q: %w", specID, ErrInvalidSPECID)
	}
	return nil
}

// ParseSpecID extracts the first SPEC ID from text using the canonical pattern.
// Returns empty string if no SPEC ID is found.
func ParseSpecID(text string) string {
	match := SpecIDPattern.FindString(text)
	return match
}
