package router

import (
	"regexp"
	"strings"
	"sync"
)

// securityKeywords is the list of security-related keywords that trigger force-thorough.
// REQ-HRN-001-008, spec.md §4 Assumptions.
//
// @MX:WARN: [AUTO] FROZEN keyword set — modifications require a CON-002 amendment
// @MX:REASON: design-constitution §5 FROZEN floor; changes to the security keyword set are treated as schema changes
var securityKeywords = []string{
	"auth", "crypto", "encrypt", "oauth", "jwt", "session", "password", "rbac", "acl",
}

// paymentKeywords is the list of payment-related keywords that trigger force-thorough.
// REQ-HRN-001-008, spec.md §4 Assumptions.
var paymentKeywords = []string{
	"payment", "billing", "subscription", "invoice", "charge", "stripe", "paypal",
}

// matchForceThoroughKeywords matches security/payment keywords against
// the SPEC title and Requirements section body.
// REQ-HRN-001-008: force-thorough if a keyword appears in title OR any requirement body.
// Note: the tags field is excluded from keyword matching (false-positive prevention).
// Word-boundary matching prevents false positives such as "author" -> "auth".
// Returns the list of matched keywords (empty slice when none match).
func matchForceThoroughKeywords(doc *SPECInput) []string {
	// Match target: title + Requirements section body (case-insensitive).
	// tags are excluded (false-positive prevention).
	searchText := strings.ToLower(doc.Title + " " + extractRequirementsSection(doc.Body))

	var matched []string

	for _, kw := range securityKeywords {
		if matchKeywordBoundary(searchText, kw) {
			matched = append(matched, kw)
		}
	}
	for _, kw := range paymentKeywords {
		if matchKeywordBoundary(searchText, kw) {
			matched = append(matched, kw)
		}
	}

	if matched == nil {
		return []string{}
	}
	return matched
}

// keywordPatternCache is the cache of compiled regex patterns.
// sync.Map provides a goroutine-safe cache.
var keywordPatternCache sync.Map

// matchKeywordBoundary checks whether the text contains the keyword, honoring word boundaries.
// Allows "auth" -> "authentication" while preventing a false positive on "author".
func matchKeywordBoundary(lowerText, keyword string) bool {
	// Look up the cached pattern.
	if v, ok := keywordPatternCache.Load(keyword); ok {
		return v.(*regexp.Regexp).MatchString(lowerText)
	}

	// Compile the pattern.
	patStr := `\b` + regexp.QuoteMeta(keyword) + `\b`
	compiled, err := regexp.Compile(patStr)
	if err != nil {
		// Fall back to a simple contains search on compile failure.
		return strings.Contains(lowerText, keyword)
	}

	// Store in the cache (concurrent writes are safe).
	keywordPatternCache.Store(keyword, compiled)
	return compiled.MatchString(lowerText)
}

// extractRequirementsSection extracts the Requirements section text from the SPEC body.
// Returns content from "## 5. Requirements" up to the next H2 section.
// Returns the entire body when no such section exists.
func extractRequirementsSection(body string) string {
	if body == "" {
		return ""
	}

	// Find the Requirements section.
	lines := strings.Split(body, "\n")
	inReqs := false
	var reqLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && strings.Contains(strings.ToLower(line), "requirement") {
			inReqs = true
			continue
		}
		if inReqs {
			// Stop at the start of the next H2 section.
			if strings.HasPrefix(line, "## ") {
				break
			}
			reqLines = append(reqLines, line)
		}
	}

	if len(reqLines) == 0 {
		return body
	}
	return strings.Join(reqLines, "\n")
}
