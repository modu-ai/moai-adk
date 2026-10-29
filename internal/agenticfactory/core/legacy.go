package core

import "fmt"

// Legacy is a record in the contract 0.4.0 model (State with blocked/on-hold and
// a return stage), with the two facts the old State alone cannot say: whether a
// PR receipt exists and was accepted by the leader, and whether the merge was seen.
type Legacy struct {
	State      State
	Return     State
	Reason     BlockReason
	Resume     string
	HasReceipt bool
	// Repair is true while the owner repairs a PR it handed off (canon E06 evidence
	// is recorded): the accepted receipt then reads as BUILDING/FIX, not REVIEWING.
	Repair bool
	// Revision is true while an approved contract revision is applied to a new
	// attempt (canon E18 adopt): claimed reads as BUILDING/SPEC, later states as
	// BUILDING/IMPLEMENT, whatever receipt the Issue holds.
	Revision       bool
	LeaderAccepted bool
	Merged         bool
}

func unmappable(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrLegacyUnmappable, fmt.Sprintf(format, a...))
}

// FromLegacy maps an old record to a position. S46 has no stage for the old
// verifying and ready (before the PR handoff E03 everything is BUILDING), so
// the stage follows the evidence the record carries (see legacyByEvidence) and
// the check evidence stays in the record.
// A paused record keeps its stage and gains the overlay; on-hold maps to the
// operator-hold reason, which S46 does not name (canon unconfirmed, OD-12).
func FromLegacy(l Legacy) (Position, error) {
	switch l.State {
	case StateBlocked, StateOnHold:
		return legacyPaused(l)
	}
	return legacyStage(l)
}

func legacyStage(l Legacy) (Position, error) {
	switch l.State {
	case StateQueued:
		return Position{Stage: StageReady}, nil
	case StateDone:
		return Position{Stage: StageDone}, nil
	case StateClaimed, StateImplementing, StateVerifying, StateReady, StateLanding:
		return legacyByEvidence(l), nil
	}
	return Position{}, unmappable("state %q", l.State)
}

// legacyByEvidence decides the stage of an in-flight record from what it
// carries, not from its old name (S46 E03 judges REVIEWING by the handoff
// evidence): a merge seen is MERGING, a receipt the leader accepted is
// REVIEWING, anything else is still BUILDING. A merge seen is checked first
// because it implies the receipt was accepted.
func legacyByEvidence(l Legacy) Position {
	switch {
	case l.Merged:
		return Position{Stage: StageMerging, RemoteIntegrated: true}
	case l.Revision && l.State == StateClaimed:
		return Position{Stage: StageBuilding, Sub: SubSpec}
	case l.Revision:
		return Position{Stage: StageBuilding, Sub: SubImplement}
	case l.HasReceipt && l.Repair:
		return Position{Stage: StageBuilding, Sub: SubFix}
	case l.HasReceipt && l.LeaderAccepted:
		return Position{Stage: StageReviewing}
	case l.HasReceipt:
		return Position{Stage: StageBuilding, Sub: SubFix}
	case l.State == StateClaimed:
		return Position{Stage: StageBuilding, Sub: SubSpec}
	}
	return Position{Stage: StageBuilding, Sub: SubImplement}
}

func legacyPaused(l Legacy) (Position, error) {
	switch l.Return {
	case "", StateBlocked, StateOnHold, StateDone:
		return Position{}, unmappable("%s needs an in-flight return stage, got %q", l.State, l.Return)
	}
	base := l
	base.State = l.Return
	pos, err := legacyStage(base)
	if err != nil {
		return Position{}, err
	}
	reason := l.Reason
	if reason == "" && l.State == StateOnHold {
		reason = ReasonOperatorHold
	}
	if l.Resume == "" {
		return Position{}, unmappable("%s has no resume condition", l.State)
	}
	p := Park{Reason: reason, Evidence: "migrated from a legacy " + string(l.State) + " record; no evidence was stored", Resume: l.Resume}
	if err := p.Validate(); err != nil {
		return Position{}, unmappable("%v", err)
	}
	pos.Park = &p
	return pos, nil
}
