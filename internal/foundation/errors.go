package foundation

import (
	"errors"
	"fmt"
)

// Sentinel errors for the foundation package.
var (
	// ErrUnsupportedLanguage indicates a programming language that is not supported.
	ErrUnsupportedLanguage = errors.New("unsupported language")
)

// LanguageNotFoundError is returned when a language lookup fails.
type LanguageNotFoundError struct {
	Query string
}

// Error returns the error message including the query that failed.
func (e *LanguageNotFoundError) Error() string {
	return fmt.Sprintf("language not found: %s", e.Query)
}
