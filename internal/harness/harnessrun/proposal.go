package harnessrun

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// approvalGateNote is carried on every emitted proposal. The producer surfaces
// an amendment REQUEST; applying it is a Tier-4 human decision, and the draft
// should say so on its face rather than read as an instruction. This mirrors
// delegationmap's approval-gate note (delegationmap/proposal.go).
const approvalGateNote = "Tier-4 approval gate — user approval required; the harness_run producer carries auto_apply:false intent"

// BuildHarnessRunCandidates converts run-time findings into proposal
// candidates for the EXISTING proposalgen writer. It constructs candidates
// directly rather than routing through proposalgen.MapPromotions: the mapper's
// subject is a promotion record from the usage-log tier ladder, which this
// producer neither reads nor produces (sibling of
// delegationmap.BuildCandidates, REQ-HRR-007 / plan.md §A.3).
//
// The output order follows the finding order, which is already deterministic,
// so two calls over the same findings slice produce the same candidate slice.
//
// An empty or nil findings slice returns a non-nil empty slice (REQ-HRR-003:
// the orchestrator distinguishes "field absent" from "no signal" — this
// producer never returns nil).
func BuildHarnessRunCandidates(findings []Finding) []proposalgen.ProposalCandidate {
	out := make([]proposalgen.ProposalCandidate, 0, len(findings))
	for _, f := range findings {
		patternKey := buildPatternKey(f)
		out = append(out, proposalgen.ProposalCandidate{
			PatternKey:       patternKey,
			ObservationCount: 1, // harness-run is a single-shot run-time signal
			Confidence:       f.Confidence,
			Tier:             f.SuggestedTier,
			// SourceTs is the zero time: the run's wall-clock stamp is applied
			// by the caller after the run returns (dynamic-workflow
			// determinism — no clock read inside the producer). Zero renders
			// deterministically, so two runs over the same findings agree.
			SourceTs: time.Time{},
			DraftID:  buildDraftID(patternKey),
			Evidence: evidenceFor(f),
		})
	}
	return out
}

// buildPatternKey renders the reserved key
// harness_run:<sha256(surface)[:8]>:<kind> (plan.md §D-D2).
//
// The discriminator hashes the surface so two findings on different surfaces
// cannot collide onto one draft. The kind is carried verbatim as the suffix so
// two findings on the same surface but differing in kind also stay distinct.
func buildPatternKey(f Finding) string {
	sum := sha256.Sum256([]byte(f.Surface))
	return PatternNamespace + ":" + hex.EncodeToString(sum[:])[:8] + ":" + f.Kind
}

// buildDraftID derives the draft directory name from the pattern key, matching
// the existing PROPOSAL-<sha256[:8]> convention so one pattern yields exactly
// one draft across runs (sibling of delegationmap.buildDraftID and
// proposalgen.buildDraftID).
func buildDraftID(patternKey string) string {
	sum := sha256.Sum256([]byte(patternKey))
	return "PROPOSAL-" + hex.EncodeToString(sum[:])[:8]
}

// evidenceFor assembles the reviewer-facing evidence carried on the draft. The
// producer-specific fields {surface, kind, summary, confidence, suggested_tier}
// flow through the ProposalCandidate.Evidence map seam
// (proposalgen/types.go:71), and the approval-gate note states plainly that
// application is gated.
func evidenceFor(f Finding) map[string]any {
	return map[string]any{
		"surface":        f.Surface,
		"kind":           f.Kind,
		"summary":        f.Summary,
		"confidence":     f.Confidence,
		"suggested_tier": f.SuggestedTier,
		"approval_gate":  approvalGateNote,
	}
}
