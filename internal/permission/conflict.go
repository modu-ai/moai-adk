package permission

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// conflictLogDir is the directory where the conflict audit log (permission.log)
// is written. It defaults to the project-root-relative ".moai/logs" (matching the
// existing logUnreachablePrompt path in resolver.go). It is a package-level seam so
// that tests can redirect the write to a t.TempDir() root without touching the real
// project tree (SPEC-SEC-HARDEN-001 §M2 / D3).
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-001 M2 — injectable log-dir seam for the conflict audit write; tests override to a t.TempDir() root.
var conflictLogDir = filepath.Join(".moai", "logs")

// specificityScore computes the specificity score of a pattern.
// Fewer wildcards (*) means higher specificity (higher score).
func specificityScore(pattern string) int {
	wildcards := strings.Count(pattern, "*")
	// Start from a base score of 100, subtract 10 per wildcard, add a length bonus.
	return 100 - (wildcards * 10) + len(pattern)
}

// @MX:NOTE: [AUTO] resolveConflict — higher specificity wins; on an equal-specificity tie, deny wins (SPEC-SEC-HARDEN-001 M2), else later Origin lexicographic order (fs-order)
// @MX:NOTE: [AUTO] tiebreak rule: REQ-V3R2-RT-002-042 AC-12 + REQ-SEC-M2-001..003 — conflicts are recorded in .moai/logs/permission.log
// resolveConflict determines the winning rule when two or more rules match within the same tier.
//
// Priority criteria:
//  1. Rule with higher specificity score wins (inverse of wildcard count).
//  2. On an equal-specificity tie, if any tied rule is a deny, a deny wins
//     (SPEC-SEC-HARDEN-001 §M2 REQ-SEC-M2-001 — deny-precedence on tie).
//  3. Otherwise (all-allow tie), the rule whose Origin sorts later lexicographically
//     wins (fs-order — preserved existing behavior, REQ-SEC-M2-002).
//
// Conflicts are recorded in .moai/logs/permission.log (best-effort; REQ-SEC-M2-003/004).
//
// Reference: SPEC-V3R2-RT-002 REQ-V3R2-RT-002-042 AC-12; SPEC-SEC-HARDEN-001 §M2.
func resolveConflict(rules []*PermissionRule, tool, input string) *PermissionRule {
	if len(rules) == 0 {
		return nil
	}
	if len(rules) == 1 {
		return rules[0]
	}

	// Record the conflict (best-effort; never affects the decision).
	logConflict(rules, tool, input)

	// SPEC-SEC-HARDEN-001 §M2 (D1 max-specificity-set form): compute the maximum
	// specificity among ALL matched rules; if that top-specificity set contains a
	// deny, restrict the tiebreak to the deny rules so deny wins on an
	// equal-specificity tie. Across specificity tiers a higher-specificity rule
	// still wins regardless of action (AC-SEC-M2-004), because the candidate set is
	// scoped to the max-specificity rules only.
	maxScore := specificityScore(rules[0].Pattern)
	for _, r := range rules[1:] {
		if s := specificityScore(r.Pattern); s > maxScore {
			maxScore = s
		}
	}

	candidates := rules
	denyInTopTier := false
	for _, r := range rules {
		if specificityScore(r.Pattern) == maxScore && r.Action == DecisionDeny {
			denyInTopTier = true
			break
		}
	}
	if denyInTopTier {
		denies := make([]*PermissionRule, 0, len(rules))
		for _, r := range rules {
			if specificityScore(r.Pattern) == maxScore && r.Action == DecisionDeny {
				denies = append(denies, r)
			}
		}
		candidates = denies
	}

	best := candidates[0]
	bestScore := specificityScore(best.Pattern)

	for _, r := range candidates[1:] {
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

// logConflict records a conflict to .moai/logs/permission.log when one occurs.
//
// Best-effort (SPEC-SEC-HARDEN-001 §M2 REQ-SEC-M2-003/004): any I/O error is
// swallowed so the conflict-log write never changes the allow/deny decision and
// never surfaces an error to the caller. The log directory is resolved from the
// conflictLogDir package seam (default ".moai/logs"; tests override it).
func logConflict(rules []*PermissionRule, tool, input string) {
	if len(rules) < 2 {
		return
	}
	// Build the conflict payload: each candidate's origin and action.
	origins := make([]string, 0, len(rules))
	for _, r := range rules {
		origins = append(origins, r.Origin+":"+string(r.Action))
	}

	// Best-effort write; any failure is silently dropped (must not affect the decision).
	if err := os.MkdirAll(conflictLogDir, 0o755); err != nil {
		return
	}
	logPath := filepath.Join(conflictLogDir, "permission.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	entry := fmt.Sprintf("[%s] Conflict resolved: tool=%s input=%s candidates=[%s]\n",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		tool, truncate(input, 200), strings.Join(origins, ", "))
	_, _ = f.WriteString(entry)
}
