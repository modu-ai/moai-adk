package core

import (
	"errors"
	"testing"
)

func parked(r BlockReason) *Park { p := park(r); return &p }

// aClassified is a reason E19 accepts.
func aClassified() BlockReason { return ClassifiedReasons()[0] }

func TestCheckMoveAllowsStayAndTableEdges(t *testing.T) {
	cases := []struct {
		name     string
		from, to Position
	}{
		{"stay", Position{Stage: StageBuilding, Sub: SubImplement}, Position{Stage: StageBuilding, Sub: SubImplement}},
		{"stay with a refreshed park text", Position{Stage: StageBuilding, Sub: SubImplement, Park: parked(ReasonDependency)}, Position{Stage: StageBuilding, Sub: SubImplement, Park: &Park{Reason: ReasonDependency, Evidence: "new", Resume: "again"}}},
		{"E01 claim", Position{Stage: StageReady}, Position{Stage: StageBuilding, Sub: SubSpec}},
		{"E23 spec passes", Position{Stage: StageBuilding, Sub: SubSpec}, Position{Stage: StageBuilding, Sub: SubImplement}},
		{"E03 handoff", Position{Stage: StageBuilding, Sub: SubImplement}, Position{Stage: StageReviewing}},
		{"E06 repair", Position{Stage: StageReviewing}, Position{Stage: StageBuilding, Sub: SubFix}},
		{"E08 merge", Position{Stage: StageReviewing}, Position{Stage: StageMerging}},
		{"E12 integrated", Position{Stage: StageMerging}, Position{Stage: StageMerging, RemoteIntegrated: true}},
		{"E13 done", Position{Stage: StageMerging, RemoteIntegrated: true}, Position{Stage: StageDone}},
		{"E15 cancel", Position{Stage: StageBuilding, Sub: SubImplement}, Position{Stage: StageCancelled}},
		{"E19 park", Position{Stage: StageReviewing}, Position{Stage: StageReviewing, Park: parked(aClassified())}},
		{"landing to done is E08, E12, E13 in one write", Position{Stage: StageReviewing}, Position{Stage: StageDone}},
		{"a claimed record holding a receipt is E03, E06", Position{Stage: StageBuilding, Sub: SubSpec}, Position{Stage: StageBuilding, Sub: SubFix}},
		{"claimed to done in one write", Position{Stage: StageBuilding, Sub: SubSpec}, Position{Stage: StageDone}},
		{"E18 adopt releases a contract-change park into SPEC", Position{Stage: StageBuilding, Sub: SubFix, Park: parked(ReasonContractChange)}, Position{Stage: StageBuilding, Sub: SubSpec}},
		{"a parked position can be cancelled", Position{Stage: StageReviewing, Park: parked(ReasonDependency)}, Position{Stage: StageCancelled}},
		{"E20 resume", Position{Stage: StageReviewing, Park: parked(aClassified())}, Position{Stage: StageReviewing}},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to); err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}

func TestCheckMoveRefusesWhatNoEdgeReaches(t *testing.T) {
	cases := []struct {
		name     string
		from, to Position
	}{
		{"ready straight to done is five edges", Position{Stage: StageReady}, Position{Stage: StageDone}},
		{"implement back to spec", Position{Stage: StageBuilding, Sub: SubImplement}, Position{Stage: StageBuilding, Sub: SubSpec}},
		{"done reopens", Position{Stage: StageDone}, Position{Stage: StageBuilding, Sub: SubImplement}},
		{"cancelled revives", Position{Stage: StageCancelled}, Position{Stage: StageReady}},
		{"a park of the unknown family is not replaced by another reason", Position{Stage: StageReviewing, Park: parked(ReasonUnknownEffect)}, Position{Stage: StageReviewing, Park: parked(aClassified())}},
		{"parked moves on", Position{Stage: StageReviewing, Park: parked(aClassified())}, Position{Stage: StageMerging, Park: parked(aClassified())}},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to); !errors.Is(err, ErrTransitionInvalid) {
			t.Errorf("%s: want ErrTransitionInvalid, got %v", c.name, err)
		}
	}
}
