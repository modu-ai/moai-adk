package mx

import (
	"fmt"
	"regexp"
	"strings"
)

// specIDRegex is the regex for extracting SPEC ID from tag body (REQ-SPC-004-006).
var specIDRegex = regexp.MustCompile(`SPEC-[A-Z0-9][A-Z0-9-]*`)

// SpecAssociator connects @MX TAG with SPEC IDs.
//  1. When the tag's file path is under the module paths listed in SPEC's frontmatter
//  2. When the tag's Body contains a SPEC ID pattern (body-based association)
//  3. When the tag carries a captured @MX:SPEC sub-line (sub-line source; additive)
//
// @MX:NOTE: [AUTO] SpecAssociator — unifies three SPEC-association sources: path-based, body-based, and sub-line (@MX:SPEC).
// Path-based: matches the SPEC frontmatter `module:` field against tag.File prefixes.
// Body-based: SPEC-[A-Z0-9-]+ regex over tag.Body.
// Sub-line: the captured tag.SpecRef field (additive third source; de-duped via the seen map).
type SpecAssociator struct {
	// specModules is the specID → []modulePath mapping.
	specModules map[string][]string
	// subLineDisabled, when true, skips the (c) sub-line source loop so the
	// associator reproduces the pre-SPEC-MX-ASSOCIATION-001 behavior. It exists
	// for the AC-002 measurement harness (on vs off delta) and defaults to
	// false (sub-line source enabled) in production.
	subLineDisabled bool
}

// NewSpecAssociator creates a SpecAssociator with SPEC ID → module path mapping.
func NewSpecAssociator(specModules map[string][]string) *SpecAssociator {
	return &SpecAssociator{
		specModules: specModules,
	}
}

// SetSubLineSourceEnabled toggles the additive sub-line (@MX:SPEC) association
// source. It defaults to enabled (false argument). The off state reproduces the
// pre-SPEC-MX-ASSOCIATION-001 association behavior, used by the AC-002 coverage
// measurement harness to compute the on-vs-off delta.
func (a *SpecAssociator) SetSubLineSourceEnabled(enabled bool) {
	a.subLineDisabled = !enabled
}

// Associate returns a list of SPEC IDs connected to the tag (REQ-SPC-004-006).
//
// It is the backward-compatible entry point: it delegates to
// AssociateWithDiagnostics and discards the diagnostics. Callers that need the
// UnresolvedSpecRef warnings (REQ-MX-ASSOC-003) MUST use
// AssociateWithDiagnostics instead.
func (a *SpecAssociator) Associate(tag Tag) []string {
	associations, _ := a.AssociateWithDiagnostics(tag)
	return associations
}

// AssociateWithDiagnostics returns the SPEC IDs connected to the tag plus a
// slice of diagnostic warning strings. Source order is deterministic and
// additive: path → body → sub-line (@MX:SPEC), de-duplicated via the seen map
// so a SPEC ID named by two sources appears once (REQ-MX-ASSOC-002).
//
// Validation (REQ-MX-ASSOC-003, flag-but-keep): a captured @MX:SPEC ID that
// does not resolve to any SPEC in the known-SPEC set (the specModules keys) is
// kept in the associations slice AND produces an UnresolvedSpecRef warning. The
// associator never panics on an unresolved ID. The validation binds ONLY the
// sub-line source (path-based and body-based sources remain unvalidated, per
// spec.md §G Out of Scope).
//
// @MX:NOTE: [AUTO] AssociateWithDiagnostics — additive sub-line source + flag-but-keep validation; de-duped via seen map
func (a *SpecAssociator) AssociateWithDiagnostics(tag Tag) ([]string, []string) {
	seen := make(map[string]bool)
	var result []string
	var diagnostics []string

	// (a) path-based connection: when the tag's file path is under the SPEC's module paths
	for specID, modules := range a.specModules {
		if isFileUnderModules(tag.File, modules) && !seen[specID] {
			seen[specID] = true
			result = append(result, specID)
		}
	}

	// (b) body-based connection: SPEC ID tokens in the tag Body text.
	for _, specID := range ExtractSpecIDs(tag.Body) {
		if !seen[specID] {
			seen[specID] = true
			result = append(result, specID)
		}
	}

	// (c) sub-line connection (REQ-MX-ASSOC-002): the captured @MX:SPEC ID.
	// Additive, de-duped via the existing seen map. Validation is flag-but-keep
	// (REQ-MX-ASSOC-003): an unresolved ID is kept AND warned, never dropped.
	// Skipped entirely when subLineDisabled (AC-002 measurement harness).
	if !a.subLineDisabled && tag.SpecRef != "" {
		if !seen[tag.SpecRef] {
			seen[tag.SpecRef] = true
			result = append(result, tag.SpecRef)
		}
		if _, resolved := a.specModules[tag.SpecRef]; !resolved {
			diagnostics = append(diagnostics,
				fmt.Sprintf("UnresolvedSpecRef: %s:%d - @MX:SPEC %q is not a known SPEC",
					tag.File, tag.Line, tag.SpecRef))
		}
	}

	return result, diagnostics
}

// "ANCHOR for SPEC-AUTH-001 handler" → ["SPEC-AUTH-001"] (REQ-SPC-004-006 (b))
func ExtractSpecIDs(body string) []string {
	matches := specIDRegex.FindAllString(body, -1)
	if len(matches) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}
	return result
}

// isFileUnderModules verifies if the file path is under one of the module paths.
// Uses path prefix matching (REQ-SPC-004-006 (a)).
func isFileUnderModules(filePath string, modulePaths []string) bool {
	for _, modulePath := range modulePaths {
		if strings.HasPrefix(filePath, modulePath) {
			return true
		}
	}
	return false
}
