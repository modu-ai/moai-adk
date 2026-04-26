package constitution

import "fmt"

// ValidateRuleReferences validates that each Rule ID in refs exists in the registry.
// Returns warning strings for IDs that are not found.
//
// Implements REQ-CON-001-041. CLI wiring and SPEC frontmatter scanning will be added in SPEC-V3R2-SPC-003.
//
// @MX:NOTE: [AUTO] CLI wiring to be added in SPEC-V3R2-SPC-003
// @MX:SPEC: SPEC-V3R2-CON-001 REQ-CON-001-041
func ValidateRuleReferences(registry *Registry, refs []string) []string {
	var warnings []string
	for _, ref := range refs {
		if _, ok := registry.Get(ref); !ok {
			warnings = append(warnings,
				fmt.Sprintf("dangling reference: %s not found in registry", ref))
		}
	}
	return warnings
}
