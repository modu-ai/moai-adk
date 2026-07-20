package curator

import (
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/harness/tier"
)

// ErrTierNotQualified is returned by the tier-gated write entry point when a
// Curator write is attempted at a higher tier than the pattern's evidence count
// qualifies for, OR when the requested tier does not match the target surface
// (REQ-HEV2-027 — no self-tier-escalation). A 6-observation pattern qualifies
// for Tier 3 only; a Tier-4 write attempt for it is rejected WITHOUT touching
// the file. Self-tier-escalation is a reward-hacking shape and is blocked here.
var ErrTierNotQualified = errors.New("curator: tier not qualified for requested surface")

// qualifiedTier maps an observation count to the highest tier level [1..4] it
// qualifies for. It reuses the canonical [1,3,5,10] learning ladder via
// tier.ClassifyStatus (the threshold SSOT) rather than duplicating the
// thresholds here — REQ-HEV2-025/026 bind only the surface selection, not the
// threshold value (which the existing learner owns).
func qualifiedTier(observations int) int {
	switch tier.ClassifyStatus(observations) {
	case tier.StatusHighConfidence:
		return 4
	case tier.StatusRule:
		return 3
	case tier.StatusHeuristic:
		return 2
	default:
		// StatusObservation and any lower/undefined state.
		return 1
	}
}

// tierForBlockType returns the evidence tier a managed-block surface corresponds
// to per the §D.2 marker contract: BlockTypeLearnedWorkflow → Tier 4 (CLAUDE.md
// digest surface), BlockTypeLearnedLocal → Tier 3 (CLAUDE.local.md append
// surface). Any other BlockType (the legacy HarnessGenerated block) is not a
// tier-gated Learned surface and returns 0.
func tierForBlockType(blockType BlockType) int {
	switch blockType {
	case BlockTypeLearnedWorkflow:
		return 4
	case BlockTypeLearnedLocal:
		return 3
	default:
		return 0
	}
}

// TierGatedWrite is the tier-gated Curator write entry point (REQ-HEV2-025/026/
// 027). It performs surface selection and self-tier-escalation blocking, then
// delegates to the surface's writer:
//
//   - Tier 4 (BlockTypeLearnedWorkflow) → WriteManagedBlock (CLAUDE.md digest,
//     atomic block replace).
//   - Tier 3 (BlockTypeLearnedLocal)    → AppendLearnedLocal per bullet
//     (CLAUDE.local.md append-only surface).
//
// It rejects — with ErrTierNotQualified and WITHOUT touching the file — when:
//
//   - blockType is not a tier-gated Learned surface (tierForBlockType == 0), or
//   - the explicit content.Tier does not match the surface's tier, or
//   - the pattern's observation count does not qualify for the surface's tier
//     (a 6-observation pattern cannot write to the Tier-4 surface).
//
// The observation count is a machine signal derived from the existing learner
// aggregation (REQ-HEV2-031 machine-signal-only); it is never model self-report.
func TierGatedWrite(path string, blockType BlockType, observations int, content BlockContent) error {
	surfaceTier := tierForBlockType(blockType)
	if surfaceTier == 0 {
		return fmt.Errorf("curator: %w: BlockType %d is not a tier-gated Learned surface",
			ErrTierNotQualified, blockType)
	}
	// Tier/surface consistency: an explicit content.Tier (non-zero) MUST match
	// the surface's tier. content.Tier == 0 means "unspecified — infer from the
	// surface" and is admitted.
	if content.Tier != 0 && content.Tier != surfaceTier {
		return fmt.Errorf("curator: %w: explicit tier %d does not match surface tier %d",
			ErrTierNotQualified, content.Tier, surfaceTier)
	}
	// Evidence qualification: the observation count MUST qualify for the surface
	// tier. This blocks self-tier-escalation (REQ-HEV2-027 / AP-HEV2-009).
	if got := qualifiedTier(observations); got < surfaceTier {
		return fmt.Errorf("curator: %w: %d observations qualify for tier %d, not tier %d",
			ErrTierNotQualified, observations, got, surfaceTier)
	}

	switch blockType {
	case BlockTypeLearnedLocal:
		// Tier-3 append-only surface: append each bullet without touching
		// existing entries (append.go dedups on ledger_key).
		for _, b := range content.Bullets {
			if err := AppendLearnedLocal(path, b); err != nil {
				return err
			}
		}
		return nil
	default:
		// Tier-4 digest surface: atomic managed-block replace.
		return WriteManagedBlock(path, blockType, content)
	}
}
