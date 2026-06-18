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
	// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M2 (mechanism ②): 4-phase model에서
	// sync-phase = implemented이며 completed는 close-infix(ClassifyPRTitle에서 먼저
	// 검사됨)로만 도달한다. legacy {sync/docs(sync) → completed} 규칙을 implemented로 정정.
	// 이 두 규칙이 completed를 추론하면 frontmatter가 (정확히) implemented인 SPEC을
	// drift로 오탐한다 (mechanism ②). close-infix가 유일한 positive completed 신호다 (AP-2).
	{"docs(sync)", transition{"sync-merge", "implemented"}},
	{"sync", transition{"sync-merge", "implemented"}},

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

// 정규 close convention infix 리터럴 (모두 소문자 — ClassifyPRTitle/shouldSkipCommitTitle은
// ToLower 후 비교한다). Status Transition Ownership Matrix의 close subject
// `chore(SPEC-{ID}): Mx-phase audit-ready signal + 3-phase close`에서 verbatim 추출.
//
// SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-020 (D4): close infix가 "4-phase close"에서
// "3-phase close"로 개명되었다. closeInfix4Phase는 git history의 과거 close commit들이
// "4-phase close"를 carry하므로 backward-compat를 위해 RETAIN된다 (drift walker가 과거
// close를 여전히 인식해야 한다). closeInfix3Phase가 신규 close commit의 정규 infix다.
// 두 infix 모두 closeInfixMatch에서 OR된다.
//
// @MX:NOTE: [AUTO] close-infix는 walker의 유일한 positive `completed` 신호다.
// @MX:REASON: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 — sync/feat에서 completed를 추론하면
//
//	genuine incomplete-close SPEC을 마스킹하므로 (AP-2) 금지.
const (
	closeInfix3Phase = "3-phase close" // new canonical infix (3-phase lifecycle, REQ-LR-020)
	closeInfix4Phase = "4-phase close" // legacy infix retained for git-history close commits
	closeInfixMx     = "mx-phase audit-ready"
)

// closeInfixMatch는 (이미 소문자화된) commit title이 정규 close convention infix를
// 포함하는지 검사한다. drift walker와 classifier가 공유한다. 신규 ("3-phase close")와
// legacy ("4-phase close") infix 모두 인식한다 (REQ-LR-020/021, AC-LR-012).
func closeInfixMatch(lowerTitle string) bool {
	return strings.Contains(lowerTitle, closeInfix3Phase) ||
		strings.Contains(lowerTitle, closeInfix4Phase) ||
		strings.Contains(lowerTitle, closeInfixMx)
}

// syncPhaseDocsPattern은 정규 sync commit subject `docs(SPEC-XXX-NNN): ... sync-phase ...`를
// 좁게 매칭한다 (이미 소문자화된 title 대상). generic `docs`→in-progress 규칙을 broaden하지
// 않도록 SPEC-ID-scoped + `sync-phase` infix를 모두 요구한다.
//
// @MX:NOTE: [AUTO] sync convention 매칭 — Status Transition Ownership Matrix sync subject 정합.
// @MX:REASON: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 — 일반 docs(non-sync)는 여전히 in-progress.
var syncPhaseDocsPattern = regexp.MustCompile(`^docs\(spec-[a-z0-9-]+-[0-9]+\):.*sync-phase`)

// isSyncPhaseDocs는 (소문자) title이 정규 sync-phase docs commit인지 검사한다.
func isSyncPhaseDocs(lowerTitle string) bool {
	return syncPhaseDocsPattern.MatchString(lowerTitle)
}

// ClassifyPRTitle analyzes a PR/commit title and returns the transition category and target status
//
// The function handles 7 main cases:
// 1. plan(spec): → (plan-merge, "planned", nil)
// 2. feat(SPEC-XXX): → (run-complete, "implemented", nil) or (run-partial, "in-progress", nil)
// 3. docs(sync): → (sync-merge, "completed", nil)
// 4. chore(spec): → (skip-meta, "", nil) for loop prevention
// 5. revert: → (no-op, "", nil) reverts don't auto-roll-back
// 6. ... 4-phase close / Mx-phase audit-ready infix → (mx-close, "completed", nil)
// 7. docs(SPEC-XXX): ...sync-phase... → (sync-merge, "implemented", nil)
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

	// 정규 close convention infix 검사 (prefix loop보다 먼저).
	// close commit subject는 `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close`이므로
	// prefix(`chore(spec-`)로는 backfill/sync `chore(SPEC-...)` 변형까지 잘못 잡는다.
	// 따라서 verbatim close infix로 매칭한다 (generic `chore` prefix 규칙보다 우선).
	if closeInfixMatch(lowerTitle) {
		return "mx-close", "completed", nil
	}

	// 정규 sync convention 검사 (prefix loop보다 먼저, close-infix 다음).
	// sync commit subject는 `docs(SPEC-{ID}): sync-phase artifacts` (in-progress→implemented)이며,
	// generic `docs`→in-progress 규칙으로 떨어지면 안 된다. SPEC-ID-scoped + `sync-phase` infix로
	// 좁게 매칭한다 (generic `docs` rule은 그대로 유지 — 일반 docs는 여전히 in-progress).
	// @MX:NOTE: [AUTO] sync-phase 매칭은 Status Transition Ownership Matrix sync subject에 정합.
	// @MX:REASON: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 AC-DCA-008 — backfill skip 후 노출된 sync docs는 implemented여야 한다.
	if isSyncPhaseDocs(lowerTitle) {
		return "sync-merge", "implemented", nil
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
