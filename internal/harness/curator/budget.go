package curator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Digest-layer budget constants (REQ-HEV2-008 / REQ-HEV2-009). These are
// load-bearing: the MOAI:LEARNED-WORKFLOW block is the always-loaded digest
// surface (design SSOT §4 "DIGEST LAYER (always-loaded, ≤3K chars, ≤20
// bullets)"), so its size is bounded to protect the context-window budget.
const (
	// MaxDigestBlockChars is the maximum number of bytes permitted in the
	// rendered bullet body of a BlockTypeLearnedWorkflow block. The body is
	// the text strictly between the start and end markers (the distilled
	// bullets); heading + marker overhead is not counted.
	MaxDigestBlockChars = 3000
	// MaxDigestBullets is the maximum number of bullets permitted in a
	// BlockTypeLearnedWorkflow block.
	MaxDigestBullets = 20
)

// Sentinel errors for digest-layer budget/cap/content enforcement.
var (
	// ErrDigestBudgetExceeded is returned by WriteManagedBlock when the
	// rendered bullet body of a LearnedWorkflow block would exceed
	// MaxDigestBlockChars. The file is NOT touched.
	ErrDigestBudgetExceeded = errors.New("curator: digest budget exceeded")
	// ErrBulletCapExceeded is returned by WriteManagedBlock when the proposed
	// bullet count for a LearnedWorkflow block exceeds MaxDigestBullets. The
	// file is NOT touched.
	ErrBulletCapExceeded = errors.New("curator: bullet cap exceeded")
	// ErrForbiddenContent is returned by WriteManagedBlock when a bullet's
	// text matches an anti-fabrication forbidden pattern (internal SPEC IDs,
	// REQ/AC tokens, ISO dates, commit SHAs). The file is NOT touched.
	ErrForbiddenContent = errors.New("curator: forbidden content in bullet text")
)

// forbiddenTokenPatterns matches the internal-specific tokens that MUST NOT
// appear in a digest bullet (REQ-HEV2-011, §25 template-internal-content
// isolation):
//
//   - SPEC / REQ / AC tokens with a multi-segment domain and a 3+ digit
//     suffix (e.g. SPEC-HARNESS-EVOLVE-002, REQ-HEV2-008, AC-HEV2-014).
//   - ISO dates in YYYY-MM-DD form (e.g. 2026-07-12).
//
// Commit SHAs are matched separately by shaLikePattern because they need a
// digit filter to avoid false positives on pure-alpha words.
var forbiddenTokenPatterns = []*regexp.Regexp{
	// (?:SPEC|REQ|AC) - domain segments ([A-Z0-9]+, hyphen-separated) - 3+ digits.
	regexp.MustCompile(`(?:SPEC|REQ|AC)-[A-Z0-9]+(?:-[A-Z0-9]+)*-[0-9]{3,}`),
	// ISO date YYYY-MM-DD.
	regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}`),
}

// shaLikePattern matches a word-bounded hex run of 7-40 characters (git short
// to full SHA length). The digit filter is applied separately in
// containsForbiddenContent because RE2 lacks lookahead: a match is only
// forbidden when it contains at least one digit, which eliminates pure-alpha
// false positives like "defaced" or "accede" while catching real SHAs like
// "d4edb5fb5".
var shaLikePattern = regexp.MustCompile(`(?i)\b[0-9a-f]{7,40}\b`)

// containsForbiddenContent reports whether text carries any anti-fabrication
// forbidden pattern (REQ-HEV2-011). When forbidden, returns the matched
// substring and true; otherwise returns "" and false.
//
// The check is locale-neutral (edge case E7): it matches only the forbidden
// patterns (SPEC IDs, REQ/AC tokens, ISO dates, SHAs) and does NOT reject CJK
// or other non-ASCII prose.
func containsForbiddenContent(text string) (matched string, forbidden bool) {
	for _, re := range forbiddenTokenPatterns {
		if m := re.FindString(text); m != "" {
			return m, true
		}
	}
	for _, m := range shaLikePattern.FindAllString(text, -1) {
		// Require at least one digit so pure-alpha hex-letter words
		// ("defaced", "accede", "deedee") are not flagged.
		if strings.ContainsAny(m, "0123456789") {
			return m, true
		}
	}
	return "", false
}

// validateDigestBlock runs the three digest-layer enforcement checks
// (REQ-HEV2-008 budget, REQ-HEV2-009 cap, REQ-HEV2-011 anti-fabrication)
// against the proposed LearnedWorkflow block content. It returns a typed error
// identifying the first violation, or nil when the content is admissible.
//
// The checks run BEFORE the file is touched: on violation the writer returns
// the error without writing, so the file's bytes are unchanged.
func validateDigestBlock(content BlockContent) error {
	if len(content.Bullets) > MaxDigestBullets {
		return ErrBulletCapExceeded
	}
	for _, bullet := range content.Bullets {
		if m, forbidden := containsForbiddenContent(bullet.Text); forbidden {
			return fmt.Errorf("%w: %s", ErrForbiddenContent, m)
		}
	}
	if bodyLen := len(renderBullets(content.Bullets)); bodyLen > MaxDigestBlockChars {
		return fmt.Errorf("%w: %d > %d", ErrDigestBudgetExceeded, bodyLen, MaxDigestBlockChars)
	}
	return nil
}
