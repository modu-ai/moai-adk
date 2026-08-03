package mx

import (
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
}

// NewSpecAssociator creates a SpecAssociator with SPEC ID → module path mapping.
func NewSpecAssociator(specModules map[string][]string) *SpecAssociator {
	return &SpecAssociator{
		specModules: specModules,
	}
}

// Associate returns a list of SPEC IDs connected to the tag (REQ-SPC-004-006).
//
// Source order is deterministic and additive: path → body → sub-line
// (@MX:SPEC), de-duplicated via the seen map so a SPEC ID named by two sources
// appears once (REQ-MX-ASSOC-002). The sub-line source is the captured
// tag.SpecRef field populated by the scanner's @MX:SPEC arm.
//
// @MX:NOTE: [AUTO] Associate sub-line source — additive third loop over tag.SpecRef, de-duped via seen map
func (a *SpecAssociator) Associate(tag Tag) []string {
	seen := make(map[string]bool)
	var result []string

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
	// Additive, de-duped via the existing seen map.
	if tag.SpecRef != "" && !seen[tag.SpecRef] {
		seen[tag.SpecRef] = true
		result = append(result, tag.SpecRef)
	}

	return result
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
