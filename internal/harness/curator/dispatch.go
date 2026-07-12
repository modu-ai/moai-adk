// Package curator — tier→surface dispatch (SPEC-HARNESS-EVOLVE-003 M2).
//
// REQ-HEV3-001: registers auto_detection as a Tier-4 Evolvable surface.
// REQ-HEV3-003: tier→surface dispatch (Tier 3 → CLAUDE.local.md, Tier 4 → CLAUDE.md).
// REQ-HEV3-004: cross-surface leak guard.
//
// H-4 resolved: this is a separate Dispatch surface in curator/dispatch.go; the
// Applier is PRESERVE for the existing harness-applier path. The full production
// wiring (TierGatedWrite + WriteManagedBlockGated callers + L5 round) lands in M5;
// M2 ships the tier→surface map + the surface registry + the cross-surface guard.
package curator

import (
	"errors"
	"fmt"
)

// ErrCrossSurfaceLeak is returned when a tier-qualified proposal targets the
// wrong surface (REQ-HEV3-004 — no cross-surface leak). A Tier-3 proposal must
// land in CLAUDE.local.md; a Tier-4 proposal must land in CLAUDE.md. Forcing the
// wrong pairing is a reward-hacking shape (REQ-HEV2-027 extended to the dispatch layer).
var ErrCrossSurfaceLeak = errors.New("curator: cross-surface leak blocked")

// ErrInvalidCuratorTier is returned when a tier has no registered Curator surface
// (only Tiers 3 and 4 have managed-block surfaces; Tiers 1-2 land in auto-memory).
var ErrInvalidCuratorTier = errors.New("curator: tier has no Curator managed-block surface")

// SurfaceTarget describes the write target for a tier-qualified Curator proposal.
type SurfaceTarget struct {
	// Path is the relative file path (project-root-relative).
	Path string

	// BlockType is the managed-block marker type for this surface.
	BlockType BlockType

	// Tier is the evidence tier this surface corresponds to.
	Tier int
}

// TierSurfaceMap is the canonical tier → SurfaceTarget mapping (REQ-HEV3-003,
// design.md §B.1). Tier 3 → CLAUDE.local.md append-only; Tier 4 → CLAUDE.md digest.
var TierSurfaceMap = map[int]SurfaceTarget{
	3: {Path: "CLAUDE.local.md", BlockType: BlockTypeLearnedLocal, Tier: 3},
	4: {Path: "CLAUDE.md", BlockType: BlockTypeLearnedWorkflow, Tier: 4},
}

// AutoDetectionSurfacePath identifies the harness.yaml auto_detection block as
// the Tier-4 Evolvable editable surface (REQ-HEV3-001). This is a config-edit
// surface (NOT a managed-block write); it is registered separately so the
// dispatch layer can route auto_detection threshold proposals through the
// value-range validator (internal/config.ValidateAutoDetectionConditions).
const AutoDetectionSurfacePath = "auto_detection"

// EvolvableSurface describes an Evolvable-zone surface registered for Curator
// edits (design SSOT §4 3-Zone contract — Evolvable-zone A6 row).
type EvolvableSurface struct {
	// Path identifies the surface (a relative file path or a config block key).
	Path string

	// Tier is the evidence tier required to edit this surface.
	Tier int

	// IsConfigEdit is true when the surface is a YAML config block edit (not a
	// managed-block write). auto_detection is the canonical IsConfigEdit=true surface.
	IsConfigEdit bool
}

// EvolvableSurfaces is the registry of Evolvable-zone surfaces the Curator MAY
// propose bounded edits to (REQ-HEV3-001). Seeded at M2 with auto_detection.
var EvolvableSurfaces = map[string]EvolvableSurface{
	AutoDetectionSurfacePath: {
		Path:         AutoDetectionSurfacePath,
		Tier:         4,
		IsConfigEdit: true,
	},
}

// SurfaceForTier returns the target managed-block surface for a tier
// (REQ-HEV3-003). Only Tiers 3 and 4 have Curator managed-block surfaces;
// Tiers 1-2 land in auto-memory (NOT a Curator concern) and return ErrInvalidCuratorTier.
func SurfaceForTier(tier int) (SurfaceTarget, error) {
	surface, ok := TierSurfaceMap[tier]
	if !ok {
		return SurfaceTarget{}, fmt.Errorf("%w: tier %d (only tiers 3 and 4 have managed-block surfaces)",
			ErrInvalidCuratorTier, tier)
	}
	return surface, nil
}

// ValidateSurfaceForTier verifies a proposed (tier, BlockType) pairing is correct
// (REQ-HEV3-004 — no cross-surface leak). A Tier-3 proposal MUST target
// BlockTypeLearnedLocal; a Tier-4 proposal MUST target BlockTypeLearnedWorkflow.
func ValidateSurfaceForTier(tier int, blockType BlockType) error {
	surface, ok := TierSurfaceMap[tier]
	if !ok {
		return fmt.Errorf("%w: tier %d has no Curator surface", ErrInvalidCuratorTier, tier)
	}
	if surface.BlockType != blockType {
		return fmt.Errorf("%w: tier %d requires BlockType %d, got %d",
			ErrCrossSurfaceLeak, tier, surface.BlockType, blockType)
	}
	return nil
}

// CuratorProposalInput is the curator-local input for tier dispatch preparation.
// It carries only primitive + curator types (no harness import) so the harness-
// layer Dispatch wrapper can construct it without an import cycle.
type CuratorProposalInput struct {
	// TargetPath is the relative file path the proposal targets.
	TargetPath string

	// PatternKey is the structural pattern token (for A7 consult by the caller).
	PatternKey string

	// Observations is the pattern's evidence count (drives tier classification).
	Observations int

	// BlockType is the requested managed-block surface type.
	BlockType BlockType

	// Content is the proposed block payload.
	Content BlockContent
}

// PrepareTierDispatch classifies the proposal's tier from the observation count,
// selects the target surface, and validates the tier↔surface pairing
// (REQ-HEV3-003/004/010). Returns the resolved SurfaceTarget with the explicit
// Tier field set on the content (REQ-HEV3-010 — exercises ErrTierNotQualified).
func PrepareTierDispatch(input CuratorProposalInput) (SurfaceTarget, BlockContent, error) {
	surface, err := SurfaceForTier(tierForBlockType(input.BlockType))
	if err != nil {
		return SurfaceTarget{}, BlockContent{}, err
	}

	// REQ-HEV3-010: set the explicit Tier field so the writer's
	// ErrTierNotQualified self-tier-escalation guard fires on every cycle.
	content := input.Content
	content.Tier = surface.Tier

	// REQ-HEV3-004: validate the tier↔surface pairing (cross-surface leak guard).
	if err := ValidateSurfaceForTier(surface.Tier, input.BlockType); err != nil {
		return SurfaceTarget{}, BlockContent{}, err
	}

	return surface, content, nil
}
