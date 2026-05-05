// Package safety — Layer 3: Contradiction Detector (REQ-HL-008).
// Detects trigger keyword overlaps and chaining rules contradictions.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// SkillTriggers is the trigger list for a single skill.
// Used as input type for DetectOverlappingTriggers.
type SkillTriggers struct {
	// SkillPath is the skill file path.
	SkillPath string

	// Keywords is the trigger keyword list for this skill.
	Keywords []string
}

// DetectOverlappingTriggers detects trigger keyword overlaps across multiple skills.
// REQ-HL-008: Records overlapping keywords in ContradictionReport.
//
// Overlaps within the same SkillPath are not considered conflicts.
//
// @MX:ANCHOR: [AUTO] DetectOverlappingTriggers is the trigger overlap detection entry point.
// @MX:REASON: [AUTO] fan_in >= 3: contradiction_test.go, pipeline.go, Phase 4 coordinator
func DetectOverlappingTriggers(skillTriggers []SkillTriggers) harness.ContradictionReport {
	if len(skillTriggers) == 0 {
		return harness.ContradictionReport{}
	}

	// keyword → list of skill paths using that keyword
	keywordToSkills := make(map[string][]string)

	for _, st := range skillTriggers {
		for _, kw := range st.Keywords {
			keywordToSkills[kw] = append(keywordToSkills[kw], st.SkillPath)
		}
	}

	var items []harness.ContradictionItem

	for kw, paths := range keywordToSkills {
		// Remove duplicate paths (same skill entered multiple times)
		uniquePaths := deduplicatePaths(paths)
		if len(uniquePaths) <= 1 {
			// No conflict if only one skill uses it
			continue
		}

		items = append(items, harness.ContradictionItem{
			Type: harness.ContradictionOverlappingTriggers,
			Description: fmt.Sprintf(
				"trigger keyword '%s' used in %d skills overlapping: %v",
				kw, len(uniquePaths), uniquePaths,
			),
			ConflictingPaths:  uniquePaths,
			ConflictingValues: []string{kw},
		})
	}

	return harness.ContradictionReport{Items: items}
}

// deduplicatePaths removes duplicates from a path slice.
func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	var result []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}

// DetectChainRuleContradictions detects contradictions between existing and proposed chaining rules.
// REQ-HL-008: Detects as contradiction if different insert_before/insert_after exist for the same phase.
//
// @MX:ANCHOR: [AUTO] DetectChainRuleContradictions is the chaining rule contradiction detection entry point.
// @MX:REASON: [AUTO] fan_in >= 3: contradiction_test.go, pipeline.go, Phase 4 coordinator
func DetectChainRuleContradictions(existing, proposed harness.ChainingRules) harness.ContradictionReport {
	// Index existing rules by phase → entry
	existingByPhase := make(map[string]harness.ChainEntry, len(existing.Chains))
	for _, entry := range existing.Chains {
		existingByPhase[entry.Phase] = entry
	}

	var items []harness.ContradictionItem

	for _, proposedEntry := range proposed.Chains {
		existingEntry, ok := existingByPhase[proposedEntry.Phase]
		if !ok {
			// No existing phase → no conflict
			continue
		}

		// Check insert_before conflict in the same phase
		if conflict := findChainConflict(existingEntry, proposedEntry); conflict != "" {
			items = append(items, harness.ContradictionItem{
				Type:              harness.ContradictionChainRules,
				Description:       conflict,
				ConflictingPaths:  []string{"chaining-rules.yaml"},
				ConflictingValues: []string{existingEntry.Phase},
			})
		}
	}

	return harness.ContradictionReport{Items: items}
}

// findChainConflict detects conflicts between two ChainEntry.
// Considers as conflict if different agents are in insert_before for the same phase.
func findChainConflict(existing, proposed harness.ChainEntry) string {
	// insert_before conflict: both existing and proposed are non-empty and different
	if len(existing.InsertBefore) > 0 && len(proposed.InsertBefore) > 0 {
		if !stringSlicesEqual(existing.InsertBefore, proposed.InsertBefore) {
			return fmt.Sprintf(
				"phase '%s' insert_before conflict: existing=%v, proposed=%v",
				existing.Phase, existing.InsertBefore, proposed.InsertBefore,
			)
		}
	}

	// insert_after conflict: both existing and proposed are non-empty and different
	if len(existing.InsertAfter) > 0 && len(proposed.InsertAfter) > 0 {
		if !stringSlicesEqual(existing.InsertAfter, proposed.InsertAfter) {
			return fmt.Sprintf(
				"phase '%s' insert_after conflict: existing=%v, proposed=%v",
				existing.Phase, existing.InsertAfter, proposed.InsertAfter,
			)
		}
	}

	return ""
}

// stringSlicesEqual checks whether two string slices are identical including order.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
