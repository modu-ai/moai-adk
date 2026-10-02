package contract

import (
	"errors"
	"reflect"
	"strings"
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
	mc, err := ProjectToMission(signedValidMissionFixture(), "develop")
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

// TestProjectionMergeTargetFollowsIntegrationTarget is the AC-GFD-024 table.
// The merge target is the caller's configured integration target: a non-empty
// target (trimmed) becomes the projected MissionContract.MergeTarget with no
// develop literal anywhere on the path, and an empty or whitespace-only target
// is refused as ErrNotProjectable naming merge_target, returning the zero
// MissionContract so no usable contract leaks out of a refusal.
func TestProjectionMergeTargetFollowsIntegrationTarget(t *testing.T) {
	cases := []struct {
		name       string
		target     string
		wantTarget string // projected MergeTarget when the row is accepted
		wantHash   string // pinned sealed hash, only for the git-flow row
		refused    bool
	}{
		{name: "git-flow develop keeps the pinned sealed hash", target: "develop", wantTarget: "develop", wantHash: preChangeProjectionSealHash},
		{name: "github-flow main", target: "main", wantTarget: "main"},
		{name: "gitlab-flow style staging", target: "staging", wantTarget: "staging"},
		{name: "padded target is trimmed", target: "  main\t", wantTarget: "main"},
		{name: "empty target is refused", target: "", refused: true},
		{name: "whitespace-only target is refused", target: "   \t\n", refused: true},
	}
	visited := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			visited++
			mc, err := ProjectToMission(signedValidMissionFixture(), tc.target)
			if tc.refused {
				if err == nil {
					t.Fatalf("target %q: expected a refusal, got MergeTarget %q", tc.target, mc.MergeTarget)
				}
				if !errors.Is(err, ErrNotProjectable) {
					t.Fatalf("target %q: err = %v, want errors.Is ErrNotProjectable", tc.target, err)
				}
				if !strings.Contains(err.Error(), "merge_target") {
					t.Fatalf("target %q: error %q does not name merge_target", tc.target, err.Error())
				}
				if !reflect.DeepEqual(mc, mission.MissionContract{}) {
					t.Fatalf("target %q: a refusal returned a non-zero contract: %+v", tc.target, mc)
				}
				return
			}
			if err != nil {
				t.Fatalf("target %q: unexpected error: %v", tc.target, err)
			}
			if mc.MergeTarget != tc.wantTarget {
				t.Fatalf("target %q: MergeTarget = %q, want %q", tc.target, mc.MergeTarget, tc.wantTarget)
			}
			sealed, err := mission.SealMissionContract(mc)
			if err != nil {
				t.Fatalf("target %q: mission.SealMissionContract: %v", tc.target, err)
			}
			if sealed.Contract.MergeTarget != tc.wantTarget {
				t.Fatalf("target %q: sealed MergeTarget = %q, want %q", tc.target, sealed.Contract.MergeTarget, tc.wantTarget)
			}
			if tc.wantHash != "" && sealed.Hash != tc.wantHash {
				t.Fatalf("target %q: sealed hash = %s, want the pinned pre-change literal %s", tc.target, sealed.Hash, tc.wantHash)
			}
		})
	}
	if visited != len(cases) {
		t.Fatalf("visited %d of %d rows", visited, len(cases))
	}
}
