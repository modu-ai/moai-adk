package contract

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/mission"
)

// preChangeProjectionSealHash is the mission.SealMissionContract hash of the
// signedValidMissionFixture projection as measured on the tree that still
// carried the literal develop merge target (card t1453, M2a start commit
// 5bc398dd8). It is a characterization pin: the projection with the git-flow
// integration target must keep producing exactly this value (AC-GFD-024 (3)),
// so a change to the projection's merge-target handling cannot silently move
// the sealed bytes of a git-flow contract.
const preChangeProjectionSealHash = "5f030074f0bc0957a4bc8df6cea9068692502987fcc9e1744badf6a67bf64229"

// TestProjectionSealHashPinnedForGitFlowTarget characterizes the projection of
// the fixture contract before the merge target becomes a parameter. It seals
// the projected contract through mission's own sealer and compares the hash
// with the pinned literal.
func TestProjectionSealHashPinnedForGitFlowTarget(t *testing.T) {
	mc, err := ProjectToMission(signedValidMissionFixture())
	if err != nil {
		t.Fatalf("ProjectToMission: unexpected error: %v", err)
	}
	if mc.MergeTarget != "develop" {
		t.Fatalf("projected MergeTarget = %q, want the git-flow target %q", mc.MergeTarget, "develop")
	}
	sealed, err := mission.SealMissionContract(mc)
	if err != nil {
		t.Fatalf("mission.SealMissionContract: %v", err)
	}
	if sealed.Hash != preChangeProjectionSealHash {
		t.Fatalf("sealed projection hash = %s, want the pinned pre-change literal %s", sealed.Hash, preChangeProjectionSealHash)
	}
}
