// Package curator — tier→surface dispatch tests (SPEC-HARNESS-EVOLVE-003 M2).
// REQ-HEV3-001/003/004: auto_detection Tier-4 surface registration, tier→surface dispatch,
// cross-surface leak guard.
package curator

import (
	"testing"
)

// TestSurfaceForTier_Tier3 verifies REQ-HEV3-003: Tier 3 → CLAUDE.local.md (BlockTypeLearnedLocal).
func TestSurfaceForTier_Tier3(t *testing.T) {
	t.Parallel()

	surface, err := SurfaceForTier(3)
	if err != nil {
		t.Fatalf("SurfaceForTier(3) error: %v", err)
	}
	if surface.BlockType != BlockTypeLearnedLocal {
		t.Errorf("Tier 3 BlockType = %d, want BlockTypeLearnedLocal (%d)", surface.BlockType, BlockTypeLearnedLocal)
	}
	if surface.Path != "CLAUDE.local.md" {
		t.Errorf("Tier 3 Path = %q, want CLAUDE.local.md", surface.Path)
	}
}

// TestSurfaceForTier_Tier4 verifies REQ-HEV3-003: Tier 4 → CLAUDE.md (BlockTypeLearnedWorkflow).
func TestSurfaceForTier_Tier4(t *testing.T) {
	t.Parallel()

	surface, err := SurfaceForTier(4)
	if err != nil {
		t.Fatalf("SurfaceForTier(4) error: %v", err)
	}
	if surface.BlockType != BlockTypeLearnedWorkflow {
		t.Errorf("Tier 4 BlockType = %d, want BlockTypeLearnedWorkflow (%d)", surface.BlockType, BlockTypeLearnedWorkflow)
	}
	if surface.Path != "CLAUDE.md" {
		t.Errorf("Tier 4 Path = %q, want CLAUDE.md", surface.Path)
	}
}

// TestSurfaceForTier_InvalidTier verifies that tiers outside [3,4] return an error.
func TestSurfaceForTier_InvalidTier(t *testing.T) {
	t.Parallel()

	for _, tier := range []int{0, 1, 2, 5, -1} {
		if _, err := SurfaceForTier(tier); err == nil {
			t.Errorf("SurfaceForTier(%d) = nil error, want error for non-Curator-surface tier", tier)
		}
	}
}

// TestValidateSurfaceForTier_CrossSurfaceLeak verifies REQ-HEV3-004:
// a Tier-3 proposal with a Tier-4 BlockType is rejected (cross-surface leak).
func TestValidateSurfaceForTier_CrossSurfaceLeak(t *testing.T) {
	t.Parallel()

	// Tier 3 forced to Tier 4 surface → ErrCrossSurfaceLeak (AC-HEV3-004 shape)
	err := ValidateSurfaceForTier(3, BlockTypeLearnedWorkflow)
	if err == nil {
		t.Error("ValidateSurfaceForTier(3, BlockTypeLearnedWorkflow) = nil, want ErrCrossSurfaceLeak")
	}
	// Tier 4 forced to Tier 3 surface → ErrCrossSurfaceLeak
	err = ValidateSurfaceForTier(4, BlockTypeLearnedLocal)
	if err == nil {
		t.Error("ValidateSurfaceForTier(4, BlockTypeLearnedLocal) = nil, want ErrCrossSurfaceLeak")
	}
}

// TestValidateSurfaceForTier_MatchingSurface verifies REQ-HEV3-004:
// matching tier↔surface combinations pass.
func TestValidateSurfaceForTier_MatchingSurface(t *testing.T) {
	t.Parallel()

	if err := ValidateSurfaceForTier(3, BlockTypeLearnedLocal); err != nil {
		t.Errorf("ValidateSurfaceForTier(3, BlockTypeLearnedLocal) = %v, want nil", err)
	}
	if err := ValidateSurfaceForTier(4, BlockTypeLearnedWorkflow); err != nil {
		t.Errorf("ValidateSurfaceForTier(4, BlockTypeLearnedWorkflow) = %v, want nil", err)
	}
}

// TestAutoDetectionSurfaceRegistered verifies REQ-HEV3-001 / AC-HEV3-001:
// the dispatch layer references the auto_detection surface as a Tier-4 Evolvable surface.
func TestAutoDetectionSurfaceRegistered(t *testing.T) {
	t.Parallel()

	if AutoDetectionSurfacePath == "" {
		t.Fatal("AutoDetectionSurfacePath is empty — auto_detection surface not registered (REQ-HEV3-001)")
	}
	if surface, ok := EvolvableSurfaces[AutoDetectionSurfacePath]; !ok {
		t.Errorf("EvolvableSurfaces missing %q entry (REQ-HEV3-001 registration)", AutoDetectionSurfacePath)
	} else if surface.Tier != 4 {
		t.Errorf("auto_detection surface Tier = %d, want 4 (Tier-4 Evolvable surface per REQ-HEV3-001)", surface.Tier)
	}
}

// TestPrepareTierDispatch_Tier4 verifies REQ-HEV3-010: tier classification from
// observations + content.Tier set explicitly + cross-surface validated.
func TestPrepareTierDispatch_Tier4(t *testing.T) {
	t.Parallel()

	input := CuratorProposalInput{
		TargetPath:   "CLAUDE.md",
		PatternKey:   "feature+plan+autopilot+success",
		Observations: 10,
		BlockType:    BlockTypeLearnedWorkflow,
		Content:      BlockContent{Bullets: []Bullet{{LedgerKey: "lk-1", Text: "pattern"}}},
	}

	surface, content, err := PrepareTierDispatch(input)
	if err != nil {
		t.Fatalf("PrepareTierDispatch error: %v", err)
	}
	if surface.Tier != 4 {
		t.Errorf("surface.Tier = %d, want 4", surface.Tier)
	}
	if content.Tier != 4 {
		t.Errorf("content.Tier = %d, want 4 (REQ-HEV3-010 explicit Tier set)", content.Tier)
	}
}

// TestPrepareTierDispatch_Tier3 verifies the Tier-3 path (CLAUDE.local.md).
func TestPrepareTierDispatch_Tier3(t *testing.T) {
	t.Parallel()

	input := CuratorProposalInput{
		TargetPath:   "CLAUDE.local.md",
		PatternKey:   "feature+run+autopilot+success",
		Observations: 5,
		BlockType:    BlockTypeLearnedLocal,
		Content:      BlockContent{Bullets: []Bullet{{LedgerKey: "lk-1", Text: "local pattern"}}},
	}

	surface, content, err := PrepareTierDispatch(input)
	if err != nil {
		t.Fatalf("PrepareTierDispatch error: %v", err)
	}
	if surface.Tier != 3 {
		t.Errorf("surface.Tier = %d, want 3", surface.Tier)
	}
	if content.Tier != 3 {
		t.Errorf("content.Tier = %d, want 3", content.Tier)
	}
}

// TestPrepareTierDispatch_InvalidBlockType verifies the error path when a
// non-Curator BlockType is passed.
func TestPrepareTierDispatch_InvalidBlockType(t *testing.T) {
	t.Parallel()

	input := CuratorProposalInput{
		BlockType: BlockTypeHarnessGenerated,
	}
	_, _, err := PrepareTierDispatch(input)
	if err == nil {
		t.Error("PrepareTierDispatch with BlockTypeHarnessGenerated = nil, want error")
	}
}
