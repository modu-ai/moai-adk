package dtcg

import (
	"fmt"
	"strings"
)

// frozenPrefixes: List of FROZEN zone file path prefixes.
// [HARD] Hardcoded in source - cannot be bypassed via config.
// REQ-DPL-011: Design constitution §2, §3.1-3.3, §5, §11, §12 + brand context protection.
//
// @MX:ANCHOR: [AUTO] Core protection boundary of Validate call chain - fan_in increase expected
// @MX:REASON: FROZEN zone bypass requirement (REQ-DPL-011, design constitution §2)
var frozenPrefixes = []string{
	// Design constitution file (includes entire §2 Frozen vs Evolvable Zones)
	".claude/rules/moai/design/constitution.md",
	// Brand context directory (§3.1 Brand Context)
	".moai/project/brand/",
}

// frozenSectionKeywords: FROZEN section identifiers within constitution file.
// Used to block references in file path + section anchor format.
var frozenSectionKeywords = []string{
	"#2-frozen-vs-evolvable-zones",
	"#3-brand-context-and-design-brief",
	"#31-brand-context",
	"#32-design-brief",
	"#33-relationship",
	"#5-safety-architecture",
	"#safety-architecture",
	"#11-gan-loop-contract",
	"#gan-loop-contract",
	"#12-evaluator-leniency-prevention",
	"#evaluator-leniency-prevention",
}

// IsFrozen: Checks if the given file path belongs to FROZEN zone.
// [HARD] FROZEN determination depends only on hardcoded list in source - cannot change via config.
func IsFrozen(path string) bool {
	// Block paths starting with frozenPrefixes
	for _, prefix := range frozenPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// If constitution.md file contains section anchor (e.g., constitution.md#5-...)
	if strings.Contains(path, "constitution.md") {
		for _, keyword := range frozenSectionKeywords {
			if strings.Contains(path, keyword) {
				return true
			}
		}
	}

	return false
}

// BlockWrite: Returns error if the given path is in FROZEN zone.
// Returns nil for allowed paths.
//
// path: Target file path for write attempt
// reason: Write attempt reason (for audit logging)
func BlockWrite(path, reason string) error {
	if !IsFrozen(path) {
		return nil
	}

	return fmt.Errorf("frozen zone write blocked: '%s' belongs to FROZEN zone (reason: %s). "+
		"Refer to design constitution §2: Constitution files and brand context cannot be directly modified", path, reason)
}
