// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-004..005.
//
// mapper.go transforms internal/harness.Promotion records into
// ProposalCandidate values according to the actionable-pattern contract:
//
//   - pattern_key MUST match `^(code_change|error_pattern|tool_failure|repeated_edit):[a-z_]+:[^:]+$`
//   - Confidence MUST be >= ConfidenceThreshold (0.70)
//   - ToTier MUST be one of {"recommendation", "approval_required"}
//
// Promotions failing any condition are silently filtered. Multiple promotions
// sharing the same pattern_key are deduplicated, keeping the record with the
// most recent Ts (which carries the latest ObservationCount + Confidence).
//
// The mapper never panics on malformed input — it is the second tolerance
// layer after the reader. Callers receive zero candidates with no diagnostic
// noise when nothing is actionable, which is the canonical no-op path on the
// current learning-history data.
package proposalgen

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// actionablePatternRE enforces the REQ-PGN-004 pattern_key shape:
//
//	<prefix>:<subject>:<scope>
//
// where prefix is one of the 4 actionable types, subject is snake_case
// (lowercase letters and underscores), and scope is any non-colon string
// (allows pattern-specific identifiers like module paths or test names).
var actionablePatternRE = regexp.MustCompile(`^(code_change|error_pattern|tool_failure|repeated_edit):[a-z_]+:[^:]+$`)

// actionableTiers names the ToTier values that admit a candidate. The two
// pre-actionable tiers ("observation", "auto_update") and any unknown tier
// are excluded.
var actionableTiers = map[string]struct{}{
	"recommendation":    {},
	"approval_required": {},
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

// isActionable applies the 3-clause filter defined by REQ-PGN-004.
func isActionable(p harness.Promotion) bool {
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
//	PROPOSAL-<YYYYMMDD>-<sha256(pattern_key)[:8]>
//
// The date segment uses the promotion's Ts (UTC) so identical inputs always
// produce identical IDs — the scaffolder relies on this for idempotent
// overwrite-safe directory creation.
func buildDraftID(p harness.Promotion) string {
	date := p.Ts.UTC().Format("20060102")
	sum := sha256.Sum256([]byte(p.PatternKey))
	short := hex.EncodeToString(sum[:])[:8]
	return "PROPOSAL-" + date + "-" + short
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
