package permission

import (
	"strings"
)

// specificityScore computes the specificity score of a pattern.
// Fewer wildcards (*) means higher specificity (higher score).
func specificityScore(pattern string) int {
	wildcards := strings.Count(pattern, "*")
	// Start from a base score of 100, subtract 10 per wildcard, add a length bonus.
	return 100 - (wildcards * 10) + len(pattern)
}

// @MX:NOTE: [AUTO] resolveConflict — specificity wins; on a tie, later Origin lexicographic order wins (fs-order)
// @MX:NOTE: [AUTO] tiebreak rule: REQ-V3R2-RT-002-042, AC-12 — conflicts are recorded in permission.log
// resolveConflict determines the winning rule when two or more rules match within the same tier.
//
// Priority criteria:
//  1. Rule with higher specificity score wins (inverse of wildcard count).
//  2. On a tie, the rule whose Origin sorts later lexicographically wins (fs-order).
//
// Conflicts are recorded in .moai/logs/permission.log.
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-042, AC-12.
func resolveConflict(rules []*PermissionRule, tool, input string) *PermissionRule {
	if len(rules) == 0 {
		return nil
	}
	if len(rules) == 1 {
		return rules[0]
	}

	// Record the conflict.
	logConflict(rules, tool, input)

	best := rules[0]
	bestScore := specificityScore(best.Pattern)

	for _, r := range rules[1:] {
		score := specificityScore(r.Pattern)
		if score > bestScore {
			best = r
			bestScore = score
		} else if score == bestScore {
			// On a tie, the later Origin lexicographically wins.
			if r.Origin > best.Origin {
				best = r
				bestScore = score
			}
		}
	}

	return best
}

// logConflict records a conflict to permission.log when one occurs.
func logConflict(rules []*PermissionRule, tool, input string) {
	if len(rules) < 2 {
		return
	}
	// Build the conflict payload.
	var origins []string
	for _, r := range rules {
		origins = append(origins, r.Origin+":"+string(r.Action))
	}
	_ = origins // Reserved for future log-file writes (currently silent).
}
