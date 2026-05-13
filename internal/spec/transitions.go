package spec

import (
	"fmt"
	"regexp"
	"strings"
)

// PrefixToStatus maps commit title prefixes to lifecycle transitions
// Each entry maps a prefix pattern (category:prefix) to (category, targetStatus)
//
// IMPORTANT: Longer prefixes must be checked before shorter ones to avoid
// false matches (e.g., "docs(sync)" before "docs"). Map iteration order
// in Go is random, so we use a slice for deterministic prefix matching.
var transitionRules = []struct {
	prefix string
	transition
}{
	// Plan phase transitions
	{"plan(spec)", transition{"plan-merge", "planned"}},
	{"plan(specs)", transition{"plan-merge", "planned"}},
	{"chore(spec)", transition{"skip-meta", ""}}, // Loop prevention
	{"docs(spec-plan)", transition{"plan-merge", "planned"}},

	// Sync phase transitions (must come before generic "docs")
	{"docs(sync)", transition{"sync-merge", "completed"}},
	{"sync", transition{"sync-merge", "completed"}},

	// Run phase transitions
	{"feat", transition{"run-complete", "implemented"}},
	{"fix", transition{"run-complete", "implemented"}},
	{"refactor", transition{"run-complete", "implemented"}},
	{"perf", transition{"run-complete", "implemented"}},
	{"test", transition{"run-complete", "implemented"}},
	{"chore", transition{"run-partial", "in-progress"}}, // Partial AC completion
	{"ci", transition{"run-partial", "in-progress"}},
	{"style", transition{"run-partial", "in-progress"}},
	{"docs", transition{"run-partial", "in-progress"}}, // Non-sync docs

	// Direct status updates (for manual overrides)
	{"status(draft)", transition{"status-update", "draft"}},
	{"status(planned)", transition{"status-update", "planned"}},
	{"status(in-progress)", transition{"status-update", "in-progress"}},
	{"status(implemented)", transition{"status-update", "implemented"}},
	{"status(completed)", transition{"status-update", "completed"}},
	{"status(superseded)", transition{"status-update", "superseded"}},
	{"status(archived)", transition{"status-update", "archived"}},
	{"status(rejected)", transition{"status-update", "rejected"}},

	// Reverts - no automatic rollback
	{"revert", transition{"no-op", ""}},
}

// transition represents a status transition
type transition struct {
	category string
	status   string
}

// ClassifyPRTitle analyzes a PR/commit title and returns the transition category and target status
//
// The function handles 5 main cases:
// 1. plan(spec): → (plan-merge, "planned", nil)
// 2. feat(SPEC-XXX): → (run-complete, "implemented", nil) or (run-partial, "in-progress", nil)
// 3. docs(sync): → (sync-merge, "completed", nil)
// 4. chore(spec): → (skip-meta, "", nil) for loop prevention
// 5. revert: → (no-op, "", nil) reverts don't auto-roll-back
//
// Returns (category, targetStatus, error)
func ClassifyPRTitle(title string) (string, string, error) {
	if title == "" {
		return "", "", fmt.Errorf("empty title")
	}

	// Convert to lowercase for prefix matching
	lowerTitle := strings.ToLower(strings.TrimSpace(title))

	// Check revert first (special case: no-op)
	if strings.HasPrefix(lowerTitle, "revert:") {
		return "no-op", "", nil
	}

	// Check each prefix pattern in order (longer prefixes first)
	for _, rule := range transitionRules {
		if strings.HasPrefix(lowerTitle, rule.prefix) {
			return rule.category, rule.status, nil
		}
	}

	// Unknown prefix
	return "unknown", "", nil
}

// TransitionSPECIDPattern detects SPEC-ID patterns in PR/commit titles
// This is a more permissive pattern than the strict specIDPattern in lint.go
// because it needs to match historical SPEC IDs that may not conform to current format
var TransitionSPECIDPattern = regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]+`)

// ExtractSPECIDs finds all SPEC identifiers in text
// Uses a permissive pattern to match both current and historical SPEC ID formats
// Examples: SPEC-HOOK-001, SPEC-STATUS-AUTO-001, SPEC-V3R2-CON-001
func ExtractSPECIDs(text string) []string {
	matches := TransitionSPECIDPattern.FindAllString(text, -1)

	// Deduplicate while preserving order
	seen := make(map[string]bool)
	var result []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			result = append(result, match)
		}
	}

	return result
}
