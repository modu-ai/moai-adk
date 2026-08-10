package delegationmap

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// patternNamespace is the reserved pattern-key prefix for this analyzer's
// proposals (REQ-HLA-010).
//
// It is deliberately NOT a member of harness.PatternBearingEventTypes(), so the
// existing proposalgen format regex — which is derived from that SSOT — rejects
// every key emitted here by construction. That is the isolation the design
// wants: widening the event-type SSOT to make the existing mapper accept these
// keys would silently widen the OBSERVER's taxonomy too, which is a different
// subject entirely (REQ-HLA-011, plan.md §B1).
const patternNamespace = "delegation_map"

// approvalGateNote is carried on every emitted proposal. The analyzer produces
// an amendment REQUEST; three independent surfaces make applying it a human
// decision, and the draft should say so on its face rather than reading as an
// instruction.
const approvalGateNote = "Tier-4 approval gate — user approval required; the delegation map carries auto_apply: false and sits on a frozen-allowlist path"

// BuildCandidates converts findings into proposal candidates for the EXISTING
// proposalgen writer. It constructs candidates directly rather than routing
// through proposalgen.MapPromotions: the mapper's subject is a promotion record
// from the usage-log tier ladder, which this analyzer neither reads nor
// produces (REQ-HLA-011).
//
// The output order follows the finding order, which is already deterministic,
// so two runs over the same ledger produce the same candidate slice.
func BuildCandidates(res Result) []proposalgen.ProposalCandidate {
	out := make([]proposalgen.ProposalCandidate, 0, len(res.Findings))
	sourceTs := parseTS(res.LatestTS)

	for _, f := range res.Findings {
		patternKey := buildPatternKey(f)
		out = append(out, proposalgen.ProposalCandidate{
			PatternKey:       patternKey,
			ObservationCount: f.ObservationCount,
			Confidence:       f.SupportRatio,
			Tier:             harness.TierRule.String(),
			SourceTs:         sourceTs,
			DraftID:          buildDraftID(patternKey),
			Evidence:         evidenceFor(f),
		})
	}
	return out
}

// buildPatternKey renders the reserved key
// delegation_map:<subcommand>:<sha256(kind|subcommand|agent)[:8]>.
//
// The discriminator hashes all three components rather than the agent alone, so
// two findings differing only in kind cannot collide onto one draft.
func buildPatternKey(f Finding) string {
	sum := sha256.Sum256([]byte(string(f.Kind) + "|" + f.Subcommand + "|" + f.Agent))
	return patternNamespace + ":" + f.Subcommand + ":" + hex.EncodeToString(sum[:])[:8]
}

// buildDraftID derives the draft directory name from the pattern key, matching
// the existing PROPOSAL-<sha256[:8]> convention so one pattern yields exactly
// one draft across runs and dates.
func buildDraftID(patternKey string) string {
	sum := sha256.Sum256([]byte(patternKey))
	return "PROPOSAL-" + hex.EncodeToString(sum[:])[:8]
}

// evidenceFor assembles the reviewer-facing evidence carried on the draft
// (REQ-HLA-012): enough to judge the amendment without re-reading the ledger.
func evidenceFor(f Finding) map[string]any {
	return map[string]any{
		"kind":               string(f.Kind),
		"subcommand":         f.Subcommand,
		"agent":              f.Agent,
		"observation_count":  f.ObservationCount,
		"support_ratio":      f.SupportRatio,
		"qualifying_rows":    f.QualifyingRows,
		"unattributed_share": f.UnattributedShare,
		"approval_gate":      approvalGateNote,
	}
}

// parseTS converts the analyzer's latest observed row timestamp into a time.
// An absent or unparseable value yields the zero time, which still renders
// deterministically — the requirement is that two runs over the same input
// agree, not that the stamp be a wall clock.
func parseTS(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}
