package router

import (
	"regexp"
	"strings"
)

// ComplexitySignals is the collection of signals used to estimate SPEC complexity.
// REQ-HRN-001-007: serves as input to the auto-detection rules.
type ComplexitySignals struct {
	// FileCount is the estimated number of related files in the SPEC.
	// Estimated from the number of internal/... path mentions in acceptance.md + plan.md.
	FileCount int
	// DomainCount is the number of related domains in the SPEC.
	// Estimated from the number of comma-separated tokens in the tags field.
	DomainCount int
	// SpecType is the estimated SPEC type.
	// Estimated as one of {bugfix, docs, config, feature, refactor} from tags or title.
	SpecType string
	// HasSecurityKeyword indicates whether a security keyword is present.
	// REQ-HRN-001-008: case-insensitive match against the security_keywords list.
	HasSecurityKeyword bool
	// HasPaymentKeyword indicates whether a payment keyword is present.
	// REQ-HRN-001-008: case-insensitive match against the payment_keywords list.
	HasPaymentKeyword bool
}

// internalPathPattern is the path pattern for internal/ or .moai/.
// Used to estimate file_count.
var internalPathPattern = regexp.MustCompile(`(?i)\b(internal/|\.moai/)\S+\.(go|yaml|yml|md)`)

// ExtractSignals extracts complexity signals from a SPECInput.
// REQ-HRN-001-007/008/012.
func ExtractSignals(doc *SPECInput) ComplexitySignals {
	signals := ComplexitySignals{}

	// file_count: number of internal/ or .moai/ path mentions in the body.
	if doc.Body != "" {
		matches := internalPathPattern.FindAllString(doc.Body, -1)
		// Count unique file paths.
		uniquePaths := make(map[string]bool)
		for _, m := range matches {
			uniquePaths[m] = true
		}
		signals.FileCount = len(uniquePaths)
	}

	// domain_count: number of comma-separated tokens in the tags field.
	if doc.Tags != "" {
		parts := strings.Split(doc.Tags, ",")
		count := 0
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				count++
			}
		}
		signals.DomainCount = count
	}
	if signals.DomainCount == 0 {
		signals.DomainCount = 1 // Default: single domain.
	}

	// spec_type: estimated from tags or title.
	signals.SpecType = inferSpecType(doc.Tags, doc.Title)

	// Keyword matching (title + Requirements section body, tags excluded).
	searchText := strings.ToLower(doc.Title + " " + extractRequirementsSection(doc.Body))
	signals.HasSecurityKeyword = hasAnyKeyword(searchText, securityKeywords)
	signals.HasPaymentKeyword = hasAnyKeyword(searchText, paymentKeywords)

	return signals
}

// inferSpecType estimates the SPEC type from tags and title.
func inferSpecType(tags, title string) string {
	combined := strings.ToLower(tags + " " + title)

	// bugfix type
	if containsAny(combined, []string{"bugfix", "fix", "bug", "hotfix"}) {
		return "bugfix"
	}
	// docs type
	if containsAny(combined, []string{"docs", "documentation", "readme"}) {
		return "docs"
	}
	// config type
	if containsAny(combined, []string{"config", "yaml", "configuration", "settings"}) {
		return "config"
	}
	// refactor type
	if containsAny(combined, []string{"refactor", "refactoring", "cleanup", "clean"}) {
		return "refactor"
	}
	// feature type
	if containsAny(combined, []string{"feature", "feat", "impl", "implement", "add"}) {
		return "feature"
	}

	return "other"
}

// hasAnyKeyword checks whether text contains at least one of the keywords.
// Case-insensitive (text must already be lowercased).
// Uses word-boundary matching to prevent false positives.
func hasAnyKeyword(lowerText string, keywords []string) bool {
	for _, kw := range keywords {
		if matchKeywordBoundary(lowerText, kw) {
			return true
		}
	}
	return false
}

// containsAny checks whether s contains at least one of the items.
func containsAny(s string, items []string) bool {
	for _, item := range items {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
