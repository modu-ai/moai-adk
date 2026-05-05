package mx

import (
	"regexp"
	"strings"
)

// specIDRegex is the regex for extracting SPEC ID from tag body (REQ-SPC-004-006).
var specIDRegex = regexp.MustCompile(`SPEC-[A-Z0-9][A-Z0-9-]*`)

// SpecAssociator connects @MX TAG with SPEC IDs.
//  1. When the tag's file path is under the module paths listed in SPEC's frontmatter
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

	for _, specID := range ExtractSpecIDs(tag.Body) {
		if !seen[specID] {
			seen[specID] = true
			result = append(result, specID)
		}
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
