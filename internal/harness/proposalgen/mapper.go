// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-004..005 — 어휘·스키마 조항은
// SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-001/002가 부분 supersede한다 (아래 계약 반영).
//
// mapper.go transforms internal/harness.Promotion records into
// ProposalCandidate values according to the actionable-pattern contract:
//
//   - pattern_key MUST match the buildPatternKey(learner.go) shape
//     <event_type>:<subject>:<context_hash>, where <event_type> is one of
//     harness.PatternBearingEventTypes() (the EventType enum SSOT, excluding
//     apply_outcome) and <subject>/<context_hash> MAY be empty. The regex is a
//     FORMAT gate only — the actual filtering is done by event-type exclusion +
//     tier + confidence.
//   - Event-type MUST NOT be agent_invocation (REQ-HLR-009 / AC-HLR-008, M4). A
//     bare tool name carries no learnable decision and is not routable to a
//     hook/rule/lesson, so agent_invocation promotions are excluded regardless
//     of tier, confidence, or format. tool_failure / user_prompt /
//     subagent_stop / session_stop / moai_subcommand / spec_reference /
//     feedback / test_fail remain promotable.
//   - Confidence MUST be >= ConfidenceThreshold (0.70)
//   - ToTier MUST be one of {"rule", "auto_update"} (the two actionable Tier
//     vocabulary values per Tier.String() SSOT; "observation"/"heuristic" are
//     pre-actionable and excluded).
//
// Supersede rationale: the pre-repair contract keyed on a hand-maintained
// 4-prefix list and a 2-value tier vocabulary — neither of which the classifier
// ever emits (the observer path only produces EventType-prefixed keys and the
// Tier enum only emits observation/heuristic/rule/auto_update). That mismatch
// filtered ALL real promotions to zero candidates (the "structural 0" defect).
// The repair aligns the mapper to the two SSOTs it consumes: the EventType enum
// (schema) and the Tier vocabulary (tier).
//
// Promotions failing any condition are silently filtered. Multiple promotions
// sharing the same pattern_key are deduplicated, keeping the record with the
// most recent Ts (which carries the latest ObservationCount + Confidence).
//
// The mapper never panics on malformed input — it is the second tolerance
// layer after the reader. Callers receive zero candidates with no diagnostic
// noise when nothing is actionable.
package proposalgen

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// actionablePatternRE enforces the pattern_key FORMAT emitted by the learner's
// buildPatternKey: <event_type>:<subject>:<context_hash>. The event_type prefix
// alternation is DERIVED from harness.PatternBearingEventTypes() (the EventType
// enum SSOT) — NOT a hand-maintained parallel list. <subject> and <context_hash>
// MAY be empty ("user_prompt::", "agent_invocation:Bash:") because the learner
// emits empty segments for system events. This regex is a format gate only; the
// real filtering (tier + confidence) lives in isActionable.
var actionablePatternRE = buildActionablePatternRE()

// buildActionablePatternRE constructs the format regex from the EventType SSOT.
// apply_outcome is excluded via PatternBearingEventTypes() (it carries no
// pattern_key). Empty subject/context_hash segments are permitted ([^:]*).
func buildActionablePatternRE() *regexp.Regexp {
	prefixes := harness.PatternBearingEventTypes()
	quoted := make([]string, 0, len(prefixes))
	for _, et := range prefixes {
		quoted = append(quoted, regexp.QuoteMeta(string(et)))
	}
	return regexp.MustCompile(`^(` + strings.Join(quoted, "|") + `):[^:]*:[^:]*$`)
}

// actionableTiers names the ToTier values that admit a candidate, derived from
// the Tier vocabulary SSOT (Tier.String()): only "rule" (5+ observations) and
// "auto_update" (10+) are proposal-eligible. The pre-actionable tiers
// ("observation", "heuristic") and any unknown tier are excluded.
var actionableTiers = map[string]struct{}{
	harness.TierRule.String():       {},
	harness.TierAutoUpdate.String(): {},
}

// excludedEventTypes names the pattern_key event-type prefixes whose promotions
// MUST be excluded from candidates regardless of tier, confidence, or format —
// the narrowing gate added by REQ-HLR-009 / AC-HLR-008 (M4). Derivation uses the
// EventType enum SSOT (string(harness.EventTypeAgentInvocation) etc.), never a
// hand-maintained parallel string.
//
// agent_invocation is excluded because its subject is a bare tool name that
// carries no learnable decision (spec.md §A.5: 94.5% of observations are
// bare-tool-name agent_invocation) and is not routable to a hook/rule/lesson.
// The remaining pattern-bearing event types (tool_failure, user_prompt,
// subagent_stop, session_stop, moai_subcommand, spec_reference, feedback,
// test_fail) remain promotable subject to the existing format + tier +
// confidence gates. This is independent of the actionableTiers observation-tier
// exclusion: an agent_invocation promotion with to_tier=auto_update + high
// confidence + valid format is rejected here even though it passes every other
// gate.
var excludedEventTypes = map[string]struct{}{
	string(harness.EventTypeAgentInvocation): {},
}

// hasExcludedEventType reports whether the pattern_key's event-type prefix (the
// segment before the first ':') is in excludedEventTypes. A malformed key with no
// colon yields the whole string as the prefix, which is not in the set; such keys
// are separately rejected by the format gate below.
func hasExcludedEventType(patternKey string) bool {
	prefix, _, _ := strings.Cut(patternKey, ":")
	_, excluded := excludedEventTypes[prefix]
	return excluded
}

// MapPromotions reduces a Promotion list to ProposalCandidate values per
// REQ-PGN-004. The returned slice is deterministically ordered:
//  1. By Confidence descending (high-confidence first), then
//  2. By PatternKey ascending (alphabetical) for tie-break stability.
//
// An empty input or a fully filtered input returns an empty (non-nil) slice.
func MapPromotions(promotions []harness.Promotion) []ProposalCandidate {
	// Dedup by pattern_key, keeping the record with the latest Ts. Two
	// passes are simpler than a single in-place sort and remain O(n).
	latest := make(map[string]harness.Promotion, len(promotions))
	for _, p := range promotions {
		if !isActionable(p) {
			continue
		}
		existing, ok := latest[p.PatternKey]
		if !ok || p.Ts.After(existing.Ts) {
			latest[p.PatternKey] = p
		}
	}

	out := make([]ProposalCandidate, 0, len(latest))
	for _, p := range latest {
		out = append(out, ProposalCandidate{
			PatternKey:       p.PatternKey,
			ObservationCount: p.ObservationCount,
			Confidence:       p.Confidence,
			Tier:             p.ToTier,
			SourceTs:         p.Ts,
			DraftID:          buildDraftID(p),
		})
	}
	SortCandidatesByConfidenceDesc(out)
	return out
}

// isActionable applies the actionability filter. The clauses are checked in
// narrowing order: event-type exclusion (REQ-HLR-009, the outermost narrowing),
// then confidence (REQ-PGN-004), then tier (actionableTiers SSOT), then format
// (actionablePatternRE). A promotion must pass every clause; the event-type
// exclusion is independent of the tier gate — it rejects agent_invocation
// promotions even when to_tier=auto_update and confidence is high.
func isActionable(p harness.Promotion) bool {
	if hasExcludedEventType(p.PatternKey) {
		return false
	}
	if p.Confidence < ConfidenceThreshold {
		return false
	}
	if _, ok := actionableTiers[p.ToTier]; !ok {
		return false
	}
	if !actionablePatternRE.MatchString(p.PatternKey) {
		return false
	}
	return true
}

// buildDraftID computes the canonical draft identifier:
//
//	PROPOSAL-<sha256(pattern_key)[:8]>
//
// REQ-HLR-011 / AC-HLR-017 (M4): the date segment was dropped so that one
// pattern_key yields exactly one draft ID across dates. Previously the ID was
// PROPOSAL-<YYYYMMDD>-<sha256(pattern_key)[:8]>, taking the date from the
// promotion's Ts; that meant the same pattern promoted on another day produced
// a new ID and a duplicate draft directory (spec.md §A.6: 7 duplicate pairs).
// The 7 existing duplicate pairs are grandfathered (forward-looking only).
// The hash alone is the idempotent identity the scaffolder relies on for
// overwrite-safe directory creation.
func buildDraftID(p harness.Promotion) string {
	sum := sha256.Sum256([]byte(p.PatternKey))
	short := hex.EncodeToString(sum[:])[:8]
	return "PROPOSAL-" + short
}

// SortCandidatesByConfidenceDesc reorders candidates in place by Confidence
// descending, with PatternKey ascending as the tie-break. Exported so the
// CLI surface can re-sort after applying --limit truncation if needed.
func SortCandidatesByConfidenceDesc(candidates []ProposalCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Confidence != candidates[j].Confidence {
			return candidates[i].Confidence > candidates[j].Confidence
		}
		return candidates[i].PatternKey < candidates[j].PatternKey
	})
}
