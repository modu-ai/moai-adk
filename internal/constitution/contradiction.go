package constitution

import (
	"fmt"
	"regexp"
	"strings"
)

// contradictionDetector is an implementation of the ContradictionDetector interface.
// Detects contradictions between rules.
type contradictionDetector struct{}

// NewContradictionDetector creates a ContradictionDetector.
func NewContradictionDetector() ContradictionDetector {
	return &contradictionDetector{}
}

// Scan checks if the proposal contradicts other rules in the registry.
// SPEC-V3R2-CON-002 REQ-CON-002-006 Layer 3 implementation.
//
// Contradiction detection strategy:
// 1. ID conflict: Reusing the same ID → block
// 2. Zone contradiction: Relaxing Frozen constraints without Frozen→Evolvable demotion → block
// 3. Clause contradiction: "MUST X" and "MUST NOT X" coexist → block
// 4. Semantic contradiction: Opposite meaning keyword pairs → warning
func (d *contradictionDetector) Scan(proposal *AmendmentProposal, registry *Registry) (*ContradictionResult, error) {
	result := &ContradictionResult{
		Conflicts: []ConflictDetail{},
	}

	// Compare with all rules in the registry
	for _, rule := range registry.Entries {
		// Skip self
		if rule.ID == proposal.RuleID {
			continue
		}

		// ID conflict (different rule using the same ID - not allowed)
		// This is already blocked in the loader, so no additional check here

		// Zone contradiction detection
		if conflict := d.detectZoneContradiction(proposal, &rule); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			if conflict.IsBlocking {
				result.HasBlockingContradiction = true
			}
		}

		// Clause contradiction detection
		if conflict := d.detectClauseContradiction(proposal, &rule); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			if conflict.IsBlocking {
				result.HasBlockingContradiction = true
			}
		}
	}

	// Return error if there are contradictions
	if result.HasBlockingContradiction {
		conflictingIDs := make([]string, 0, len(result.Conflicts))
		descriptions := make([]string, 0, len(result.Conflicts))
		for _, c := range result.Conflicts {
			conflictingIDs = append(conflictingIDs, c.ConflictingRuleID)
			descriptions = append(descriptions, c.Description)
		}
		return result, &ErrContradictionDetected{
			NewRuleID:      proposal.RuleID,
			ConflictingIDs: conflictingIDs,
			Conflicts:      descriptions,
		}
	}

	return result, nil
}

// detectZoneContradiction detects zone change contradictions.
func (d *contradictionDetector) detectZoneContradiction(proposal *AmendmentProposal, rule *Rule) *ConflictDetail {
	// Extract zone hints from proposal's After clause
	afterUpper := strings.ToUpper(proposal.After)

	// Detect Frozen constraint relaxation: Removing "MUST", adding "MAY" or "SHOULD"
	beforeUpper := strings.ToUpper(proposal.Before)
	hadMust := strings.Contains(beforeUpper, "MUST") || strings.Contains(beforeUpper, "SHALL") || strings.Contains(beforeUpper, "REQUIRED")
	lostMust := hadMust && !strings.Contains(afterUpper, "MUST") && !strings.Contains(afterUpper, "SHALL") && !strings.Contains(afterUpper, "REQUIRED")
	addedMay := strings.Contains(afterUpper, "MAY") || strings.Contains(afterUpper, "SHOULD") || strings.Contains(afterUpper, "OPTIONAL")

	if lostMust && addedMay {
		return &ConflictDetail{
			ConflictingRuleID: rule.ID,
			Description:       fmt.Sprintf("Constraint relaxation: %s's mandatory clause changed to recommendation", rule.ID),
			IsBlocking:        true, // Frozen zone blocks constraint relaxation
		}
	}

	return nil
}

// detectClauseContradiction detects clause content contradictions.
func (d *contradictionDetector) detectClauseContradiction(proposal *AmendmentProposal, rule *Rule) *ConflictDetail {
	// Compare proposal's After clause with existing rule's clause

	// 1. Detect opposite meaning keyword pairs
	afterWords := strings.Fields(strings.ToUpper(proposal.After))
	ruleWords := strings.Fields(strings.ToUpper(rule.Clause))

	// Detect "MUST NOT" + "MUST" combination
	proposalMustNot := containsSequence(afterWords, "MUST", "NOT")
	ruleMust := containsWord(ruleWords, "MUST") || containsWord(ruleWords, "SHALL") || containsWord(ruleWords, "REQUIRED")

	if proposalMustNot && ruleMust {
		return &ConflictDetail{
			ConflictingRuleID: rule.ID,
			Description:       fmt.Sprintf("Semantic contradiction: %s is a requirement but proposal adds 'MUST NOT' prohibition", rule.ID),
			IsBlocking:        true,
		}
	}

	// 2. Detect opposing action patterns
	// Example: "always X" vs "never X"
	if extractAction(proposal.After) != "" && extractAction(proposal.After) == extractAction(rule.Clause) {
		proposalMod := extractModifier(proposal.After)
		ruleMod := extractModifier(rule.Clause)

		if isOppositeModifier(proposalMod, ruleMod) {
			return &ConflictDetail{
				ConflictingRuleID: rule.ID,
				Description:       fmt.Sprintf("Action contradiction: %s('%s') vs proposal('%s')", rule.ID, ruleMod, proposalMod),
				IsBlocking:        true,
			}
		}
	}

	return nil
}

// containsSequence finds a consecutive word sequence in a slice.
func containsSequence(words []string, seq ...string) bool {
	for i := 0; i <= len(words)-len(seq); i++ {
		match := true
		for j, word := range seq {
			if i+j >= len(words) || words[i+j] != word {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// containsWord finds a word in a slice.
func containsWord(words []string, target string) bool {
	for _, w := range words {
		if w == target {
			return true
		}
	}
	return false
}

// extractAction extracts the action/object from a clause.
// Simple implementation: Content in quotes or the first noun phrase.
func extractAction(clause string) string {
	// Extract content in quotes
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(clause)
	if len(matches) >= 2 {
		return matches[1]
	}
	// TODO: More sophisticated NLP-based estimation (SPEC-V3R2-CON-003)
	return ""
}

// extractModifier extracts modifiers (MUST/SHOULD/MAY/NOT, etc.) from a clause.
func extractModifier(clause string) string {
	upper := strings.ToUpper(clause)
	modifiers := []string{"MUST NOT", "MUST", "SHALL NOT", "SHALL", "REQUIRED", "SHOULD", "MAY", "OPTIONAL"}
	for _, mod := range modifiers {
		if strings.Contains(upper, mod) {
			return mod
		}
	}
	return ""
}

// isOppositeModifier determines if two modifiers are in an opposing relationship.
func isOppositeModifier(mod1, mod2 string) bool {
	opposites := map[string][]string{
		"MUST NOT": {"MUST", "SHALL", "REQUIRED"},
		"MUST":     {"MUST NOT", "NEVER"},
		"SHALL":    {"SHALL NOT", "MUST NOT"},
		"NEVER":    {"MUST", "SHALL", "REQUIRED", "ALWAYS"},
		"ALWAYS":   {"NEVER", "MUST NOT"},
	}

	for _, opposite := range opposites[mod1] {
		if opposite == mod2 {
			return true
		}
	}
	return false
}

// contradictionDetector satisfies the ContradictionDetector interface.
var _ ContradictionDetector = (*contradictionDetector)(nil)
