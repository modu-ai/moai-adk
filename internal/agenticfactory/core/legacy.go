package core

import "fmt"

// Legacy is a record in the contract 0.4.0 model (State with blocked/on-hold and
// a return stage), with the two facts the old State alone cannot say: whether a
// PR receipt exists and was accepted by the leader, and whether the merge was seen.
type Legacy struct {
	State          State
	Return         State
	Reason         BlockReason
	Resume         string
	HasReceipt     bool
	LeaderAccepted bool
	Merged         bool
}

func unmappable(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrLegacyUnmappable, fmt.Sprintf(format, a...))
}

// FromLegacy maps an old record to a position. S46 has no stage for the old
// verifying and ready (before the PR handoff E03 everything is BUILDING), so
// both map to BUILDING/IMPLEMENT and the check evidence stays in the record.
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
	building := func(sub SubStage) Position { return Position{Stage: StageBuilding, Sub: sub} }
	implementing := SubImplement
	if l.HasReceipt {
		implementing = SubFix
	}
	switch l.State {
	case StateQueued:
		return Position{Stage: StageReady}, nil
	case StateClaimed:
		if l.HasReceipt {
			return building(SubFix), nil
		}
		return building(SubSpec), nil
	case StateImplementing, StateVerifying, StateReady:
		return building(implementing), nil
	case StateLanding:
		switch {
		case l.Merged:
			return Position{Stage: StageMerging, RemoteIntegrated: true}, nil
		case l.HasReceipt && l.LeaderAccepted:
			return Position{Stage: StageReviewing}, nil
		}
		return building(SubImplement), nil
	case StateDone:
		return Position{Stage: StageDone}, nil
	}
	return Position{}, unmappable("state %q", l.State)
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
