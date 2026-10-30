package core

import (
	"errors"
	"testing"
)

func parked(r BlockReason) *Park { p := park(r); return &p }

// aClassified is a reason E19 accepts.
func aClassified() BlockReason { return ClassifiedReasons()[0] }

var (
	bSpec  = Position{Stage: StageBuilding, Sub: SubSpec}
	bImpl  = Position{Stage: StageBuilding, Sub: SubImplement}
	bFix   = Position{Stage: StageBuilding, Sub: SubFix}
	review = Position{Stage: StageReviewing}
	merge  = Position{Stage: StageMerging}
	mergeI = Position{Stage: StageMerging, RemoteIntegrated: true}
	done   = Position{Stage: StageDone}
)

func with(p Position, r BlockReason) Position { p.Park = parked(r); return p }

func TestCheckMoveAllowsStayAndSingleEdges(t *testing.T) {
	cases := []struct {
		name     string
		from, to Position
		have     []string
	}{
		{"stay", bImpl, bImpl, nil},
		{"stay with a refreshed park text", with(bImpl, ReasonDependency), Position{Stage: StageBuilding, Sub: SubImplement, Park: &Park{Reason: ReasonDependency, Evidence: "new", Resume: "again"}}, nil},
		{"E01 claim", Position{Stage: StageReady}, bSpec, nil},
		{"E23 spec passes", bSpec, bImpl, nil},
		{"E03 accepted handoff", bImpl, review, []string{FactReceipt, FactReceiptAccepted}},
		{"E06 repair", review, bFix, nil},
		{"E08 merge", review, merge, nil},
		{"E12 integrated", merge, mergeI, []string{FactMerged}},
		{"E13 done", mergeI, done, []string{FactMerged, FactClosed, FactSettled}},
		{"E15 cancel", bImpl, Position{Stage: StageCancelled}, nil},
		{"E19 park", review, with(review, aClassified()), nil},
		{"E20 resume in place", with(review, aClassified()), review, []string{"resume-guard:" + string(aClassified())}},
		{"E20 resume of unknown-effect", with(merge, ReasonUnknownEffect), merge, []string{"resume-guard:unknown-effect"}},
		{"E18 adopt releases a contract-change park into SPEC", with(bFix, ReasonContractChange), bSpec, nil},
		{"a parked position can be cancelled", with(review, ReasonDependency), Position{Stage: StageCancelled}, nil},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to, c.have); err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}

// The old State packs several edges into one write. Only these compositions are
// allowed, and each needs the facts of the edges it absorbed on the record written.
func TestCheckMoveAllowsOnlyTheNamedLegacyCompositionsAndTheirFacts(t *testing.T) {
	all := []string{FactReceipt, FactReceiptAccepted, FactMerged, FactClosed, FactSettled}
	cases := []struct {
		name     string
		from, to Position
		need     []string
	}{
		{"an unaccepted receipt reads as FIX (E03, E06) from implement", bImpl, bFix, []string{FactReceipt}},
		{"an unaccepted receipt reads as FIX from spec", bSpec, bFix, []string{FactReceipt}},
		{"landing to done (E08, E12, E13)", review, done, all},
		{"landing with the merge seen (E08, E12)", review, mergeI, []string{FactReceipt, FactReceiptAccepted, FactMerged}},
		{"finish without a handoff (E03, E08, E12, E13), the local path", bImpl, done, []string{FactMerged, FactClosed, FactSettled}},
		{"merge seen without a handoff (E03, E08, E12)", bImpl, mergeI, []string{FactMerged}},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to, c.need); err != nil {
			t.Errorf("%s with its facts: %v", c.name, err)
		}
		for i := range c.need {
			short := append(append([]string(nil), c.need[:i]...), c.need[i+1:]...)
			if err := CheckMove(c.from, c.to, short); !errors.Is(err, ErrGuardMissing) {
				t.Errorf("%s without %q: want ErrGuardMissing, got %v", c.name, c.need[i], err)
			}
		}
	}
}

func TestCheckMoveRefusesWhatNoEdgeOrNamedCompositionReaches(t *testing.T) {
	all := []string{FactReceipt, FactReceiptAccepted, FactMerged, FactClosed, FactSettled, "resume-guard:" + string(ReasonDependency)}
	cases := []struct {
		name     string
		from, to Position
	}{
		{"ready straight to merging", Position{Stage: StageReady}, mergeI},
		{"ready straight to done", Position{Stage: StageReady}, done},
		{"implement back to spec", bImpl, bSpec},
		{"done reopens", done, bImpl},
		{"cancelled revives", Position{Stage: StageCancelled}, Position{Stage: StageReady}},
		{"fix straight to done", bFix, done},
		{"parked moves on", with(review, aClassified()), merge},
		{"parked finishes", with(bImpl, ReasonDependency), done},
		{"resume and move in one write", with(review, ReasonDependency), merge},
		{"adopt reaches SPEC only", with(bFix, ReasonContractChange), bImpl},
		{"a stage change that also parks", review, with(merge, aClassified())},
		{"a park of the unknown family is not replaced by another reason", with(review, ReasonUnknownEffect), with(review, aClassified())},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to, all); !errors.Is(err, ErrTransitionInvalid) && !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s: want a refusal, got %v", c.name, err)
		}
	}
}

// E20 is the one edge whose evidence the writer carries: the guard for the reason it
// resolves. The table checks that the key is present, so a missing or mismatched key
// is a refusal, including for the unknown family.
func TestCheckMoveResumeNeedsTheGuardOfItsOwnReason(t *testing.T) {
	for _, r := range []BlockReason{ReasonDependency, ReasonStopUnknown, ReasonUnknownEffect, ReasonPendingEffect, ReasonContractChange, ReasonStartUnknown, aClassified()} {
		from := with(review, r)
		if err := CheckMove(from, review, nil); !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s resumed without a guard: %v", r, err)
		}
		if err := CheckMove(from, review, []string{"resume-guard:other"}); !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s resumed with another reason's guard: %v", r, err)
		}
		if err := CheckMove(from, review, []string{"resume-guard:" + string(r)}); err != nil {
			t.Errorf("%s resumed with its guard: %v", r, err)
		}
	}
}

// E15 cleanup-unknown parks with a reason about the cleanup, not any reason.
func TestE15CleanupUnknownTakesACleanupReasonOnly(t *testing.T) {
	ev := func(r BlockReason) Event {
		return Event{ID: "E15", Outcome: "cleanup-unknown", Evidence: []string{"new-grants-blocked"}, Park: park(r)}
	}
	for _, r := range []BlockReason{ReasonUnknownEffect, ReasonPendingEffect, ReasonStopUnknown} {
		if _, err := Transition(bImpl, ev(r)); err != nil {
			t.Errorf("%s: %v", r, err)
		}
	}
	for _, r := range []BlockReason{ReasonDependency, ReasonContractChange, ReasonStartUnknown, aClassified()} {
		if _, err := Transition(bImpl, ev(r)); !errors.Is(err, ErrReasonNotAllowed) {
			t.Errorf("%s: want ErrReasonNotAllowed, got %v", r, err)
		}
	}
}

// S46 has no edge for the operator's hold (OD-12, unconfirmed): the old State has
// on-hold. It is a named exception, not a way round the table: in place, from a
// position that is not finished or parked, and only with the writer carrying
// "operator-hold".
func TestCheckMoveOperatorHoldIsANamedInPlaceParkWithItsEvidence(t *testing.T) {
	for _, from := range []Position{Position{Stage: StageReady}, bSpec, bImpl, bFix, review, merge} {
		to := with(from, ReasonOperatorHold)
		if err := CheckMove(from, to, []string{"operator-hold"}); err != nil {
			t.Errorf("%s: %v", spot(from), err)
		}
		if err := CheckMove(from, to, nil); !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s without the evidence: %v", spot(from), err)
		}
	}
	for _, from := range []Position{done, {Stage: StageCancelled}, with(review, ReasonDependency)} {
		if err := CheckMove(from, with(from, ReasonOperatorHold), []string{"operator-hold"}); err == nil {
			t.Errorf("%s must not take an operator hold", spot(from))
		}
	}
	if err := CheckMove(review, with(merge, ReasonOperatorHold), []string{"operator-hold"}); err == nil {
		t.Error("an operator hold does not move the stage")
	}
}

// E03, E12 and E13 are single edges, and the record must show what they stand on.
func TestCheckMoveSingleEdgesNeedTheFactsTheRecordHolds(t *testing.T) {
	cases := []struct {
		name     string
		from, to Position
		need     []string
	}{
		{"E03", bImpl, review, []string{FactReceipt, FactReceiptAccepted}},
		{"E12", merge, mergeI, []string{FactMerged}},
		{"E13", mergeI, done, []string{FactMerged, FactClosed, FactSettled}},
	}
	for _, c := range cases {
		for i := range c.need {
			short := append(append([]string(nil), c.need[:i]...), c.need[i+1:]...)
			if err := CheckMove(c.from, c.to, short); !errors.Is(err, ErrGuardMissing) {
				t.Errorf("%s without %q: want ErrGuardMissing, got %v", c.name, c.need[i], err)
			}
		}
		if err := CheckMove(c.from, c.to, nil); !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s with no facts: %v", c.name, err)
		}
	}
}
