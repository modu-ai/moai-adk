// Package tui provides Bubble Tea TUI components for MoAI-ADK.
package tui

// WizardResult holds the final wizard results from the TUI.
// This is a simplified version for the PoC.
type WizardResult struct {
	// ProjectName is the user's project name
	ProjectName string
}

// ValidationError represents a validation error for form fields.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Validator is a function that validates a value.
type Validator func(string) error

// RequiredValidator creates a validator that checks for non-empty values.
func RequiredValidator(fieldName string) Validator {
	return func(value string) error {
		if value == "" {
			return NewValidationError(fieldName, fieldName+" is required")
		}
		return nil
	}
}

// MinLengthValidator creates a validator that checks minimum length.
func MinLengthValidator(fieldName string, min int) Validator {
	return func(value string) error {
		if len(value) < min {
			return NewValidationError(fieldName,
				fieldName+" must be at least "+string(rune('0'+min))+" characters")
		}
		return nil
	}
}

// MaxLengthValidator creates a validator that checks maximum length.
func MaxLengthValidator(fieldName string, max int) Validator {
	return func(value string) error {
		if len(value) > max {
			return NewValidationError(fieldName,
				fieldName+" must be at most "+string(rune('0'+max))+" characters")
		}
		return nil
	}
}

// PatternValidator creates a validator that checks a regex pattern.
func PatternValidator(fieldName string, pattern string) Validator {
	return func(value string) error {
		// Simple implementation for PoC
		// In production, use regexp
		for _, r := range value {
			isValid := (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-' || r == '_'
			if !isValid {
				return NewValidationError(fieldName,
					fieldName+" can only contain letters, numbers, hyphens, and underscores")
			}
		}
		return nil
	}
}
