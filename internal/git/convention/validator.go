package convention

import (
	"fmt"
	"slices"
	"strings"
)

// Validate checks a commit message against a convention.
// If conv is nil the message is considered valid.
func Validate(message string, conv *Convention) ValidationResult {
	if conv == nil {
		return ValidationResult{Valid: true, Message: message}
	}

	result := ValidationResult{Message: message}

	// Extract first line (header) for validation.
	header := strings.SplitN(message, "\n", 2)[0]
	header = strings.TrimSpace(header)

	if header == "" {
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationRequired,
			Field:    "header",
			Expected: "non-empty commit message",
			Actual:   "",
		})
		result.Valid = false
		return result
	}

	// Check max length.
	if conv.MaxLength > 0 && len(header) > conv.MaxLength {
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationMaxLength,
			Field:    "header",
			Expected: fmt.Sprintf("max %d characters", conv.MaxLength),
			Actual:   fmt.Sprintf("%d characters", len(header)),
		})
	}

	// Check pattern match.
	if !conv.Pattern.MatchString(header) {
		result.Violations = append(result.Violations, Violation{
			Type:       ViolationPattern,
			Field:      "header",
			Expected:   conv.Pattern.String(),
			Actual:     header,
			Suggestion: suggestFix(header, conv),
		})
	} else {
		// Pattern matches; check semantic rules.
		validateSemantics(header, conv, &result)
	}

	result.Valid = len(result.Violations) == 0
	return result
}

// validateSemantics checks type and scope against allowed lists.
func validateSemantics(header string, conv *Convention, result *ValidationResult) {
	commitType := extractType(header)

	// Check type validity.
	if len(conv.Types) > 0 && commitType != "" {
		found := slices.Contains(conv.Types, commitType)
		if !found {
			result.Violations = append(result.Violations, Violation{
				Type:     ViolationInvalidType,
				Field:    "type",
				Expected: strings.Join(conv.Types, ", "),
				Actual:   commitType,
			})
		}
	}

	// Check scope validity (only when scopes are defined).
	scope := extractScope(header)
	if len(conv.Scopes) > 0 && scope != "" {
		found := slices.Contains(conv.Scopes, scope)
		if !found {
			result.Violations = append(result.Violations, Violation{
				Type:     ViolationInvalidScope,
				Field:    "scope",
				Expected: strings.Join(conv.Scopes, ", "),
				Actual:   scope,
			})
		}
	}
}

// extractType extracts the commit type from the header.
// e.g., "feat(auth): add JWT" -> "feat"
func extractType(header string) string {
	for i, c := range header {
		if c == '(' || c == ':' || c == '!' {
			return header[:i]
		}
	}
	return ""
}

// extractScope extracts the scope from the header.
// e.g., "feat(auth): add JWT" -> "auth"
func extractScope(header string) string {
	start := strings.IndexByte(header, '(')
	if start < 0 {
		return ""
	}
	end := strings.IndexByte(header[start:], ')')
	if end < 0 {
		return ""
	}
	return header[start+1 : start+end]
}

// suggestFix attempts to create a valid message suggestion.
func suggestFix(header string, conv *Convention) string {
	lower := strings.ToLower(header)

	suggestedType := "chore"
	switch {
	case strings.Contains(lower, "fix") || strings.Contains(lower, "bug"):
		suggestedType = "fix"
	case strings.Contains(lower, "add") || strings.Contains(lower, "feat") || strings.Contains(lower, "new"):
		suggestedType = "feat"
	case strings.Contains(lower, "doc") || strings.Contains(lower, "readme"):
		suggestedType = "docs"
	case strings.Contains(lower, "test"):
		suggestedType = "test"
	case strings.Contains(lower, "refactor") || strings.Contains(lower, "clean"):
		suggestedType = "refactor"
	}

	desc := strings.TrimSpace(header)
	if len(desc) > 0 {
		// Lowercase first character.
		desc = strings.ToLower(desc[:1]) + desc[1:]
	}

	return suggestedType + ": " + desc
}
