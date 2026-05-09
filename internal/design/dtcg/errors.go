// Package dtcg: W3C Design Tokens Community Group (DTCG) 2025.10 spec-based token validator.
//
// Source: https://tr.designtokens.org/format/ (Editor's Draft 2025-10)
// SPEC: SPEC-V3R3-DESIGN-PIPELINE-001 Phase 3
//
// Usage example:
//
//	tokens := map[string]any{
//	    "color-primary": map[string]any{
//	        "$type":  "color",
//	        "$value": "#0066ff",
//	    },
//	}
//	report, err := dtcg.Validate(tokens)
//	if err != nil { ... }
//	if !report.Valid { ... report.Errors ... }
package dtcg

import "fmt"

// ValidationError: Validation error at a specific token path.
// Represents violation of DTCG 2025.10 §9 validation rules.
type ValidationError struct {
	// TokenPath: Token path where error occurred (e.g., "group.subgroup.token-name")
	TokenPath string
	// Category: $type value of the token (e.g., "color", "dimension")
	Category string
	// Rule: Description of violated rule (e.g., "invalid hex color format")
	Rule string
	// Value: Actual value that caused error (for debugging)
	Value any
}

// Error: String representation of ValidationError.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("token '%s' (%s): %s (value: %v)", e.TokenPath, e.Category, e.Rule, e.Value)
	}
	return fmt.Sprintf("token '%s' (%s): %s", e.TokenPath, e.Category, e.Rule)
}

// ValidationWarning: Non-blocking warning requiring attention but not stopping token usage.
// Example: Color value conflicting with brand context.
type ValidationWarning struct {
	// TokenPath: Token path where warning occurred
	TokenPath string
	// Category: $type value of the token
	Category string
	// Message: Warning message
	Message string
}

// Warning: String representation of ValidationWarning.
func (w *ValidationWarning) Warning() string {
	return fmt.Sprintf("warning: token '%s' (%s): %s", w.TokenPath, w.Category, w.Message)
}

// Report: Validation result report for entire token set.
// Top-level result type returned by DTCG validator.
type Report struct {
	// Valid: Whether all tokens passed validation without errors
	Valid bool
	// Errors: List of validation errors (at least 1 when Valid == false)
	Errors []*ValidationError
	// Warnings: List of non-blocking warnings
	Warnings []*ValidationWarning
	// TokenCount: Number of validated tokens
	TokenCount int
}

// HasErrors: Checks if validation errors exist.
func (r *Report) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings: Checks if warnings exist.
func (r *Report) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// AddError: Adds validation error to report and sets Valid to false.
func (r *Report) AddError(err *ValidationError) {
	r.Errors = append(r.Errors, err)
	r.Valid = false
}

// AddWarning: Adds warning to report.
func (r *Report) AddWarning(w *ValidationWarning) {
	r.Warnings = append(r.Warnings, w)
}
