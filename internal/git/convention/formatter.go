package convention

import (
	"fmt"
	"strings"
)

// FormatError creates a user-friendly error message from a ValidationResult.
// Returns an empty string when the result is valid.
func FormatError(result ValidationResult, conv *Convention) string {
	if result.Valid {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Commit message violates %s convention:\n", conv.Name)
	fmt.Fprintf(&b, "  Message: %q\n", result.Message)
	fmt.Fprintf(&b, "\n")

	for _, v := range result.Violations {
		fmt.Fprintf(&b, "  %s %s\n", violationIcon(v.Type), FormatViolation(v))
	}

	if len(conv.Examples) > 0 {
		fmt.Fprintf(&b, "\n  Examples of valid messages:\n")
		for _, ex := range conv.Examples {
			fmt.Fprintf(&b, "    - %s\n", ex)
		}
	}

	return b.String()
}

// FormatViolation formats a single violation into a human-readable string.
func FormatViolation(v Violation) string {
	switch v.Type {
	case ViolationPattern:
		s := fmt.Sprintf("Pattern mismatch (expected: %s)", v.Expected)
		if v.Suggestion != "" {
			s += fmt.Sprintf("\n    Suggestion: %s", v.Suggestion)
		}
		return s
	case ViolationInvalidType:
		return fmt.Sprintf("Invalid type %q (allowed: %s)", v.Actual, v.Expected)
	case ViolationInvalidScope:
		return fmt.Sprintf("Invalid scope %q (allowed: %s)", v.Actual, v.Expected)
	case ViolationMaxLength:
		return fmt.Sprintf("Header too long (%s, %s)", v.Actual, v.Expected)
	case ViolationRequired:
		return fmt.Sprintf("Missing required field: %s", v.Field)
	default:
		return fmt.Sprintf("%s: expected %s, got %s", v.Type, v.Expected, v.Actual)
	}
}

// FormatBatchSummary creates a summary for multiple validation results.
func FormatBatchSummary(results []ValidationResult, conv *Convention) string {
	if len(results) == 0 {
		return "No commits to validate."
	}

	valid := 0
	for _, r := range results {
		if r.Valid {
			valid++
		}
	}

	total := len(results)
	if valid == total {
		return fmt.Sprintf("All %d commits follow %s convention.", total, conv.Name)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d of %d commits violate %s convention:\n\n", total-valid, total, conv.Name)

	for _, r := range results {
		if !r.Valid {
			b.WriteString(FormatError(r, conv))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// violationIcon returns a label for the violation type.
func violationIcon(vt ViolationType) string {
	switch vt {
	case ViolationPattern:
		return "[PATTERN]"
	case ViolationInvalidType:
		return "[TYPE]"
	case ViolationInvalidScope:
		return "[SCOPE]"
	case ViolationMaxLength:
		return "[LENGTH]"
	case ViolationRequired:
		return "[REQUIRED]"
	default:
		return "[ERROR]"
	}
}
