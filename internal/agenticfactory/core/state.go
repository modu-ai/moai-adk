package core

import "fmt"

// State is an Issue execution state (design §2.1).
type State string

const (
	StateQueued       State = "queued"
	StateClaimed      State = "claimed"
	StateImplementing State = "implementing"
	StateVerifying    State = "verifying"
	StateReady        State = "ready"
	StateLanding      State = "landing"
	StateDone         State = "done"
	StateBlocked      State = "blocked"
	StateOnHold       State = "on-hold"
)

var allStates = []State{
	StateQueued, StateClaimed, StateImplementing, StateVerifying, StateReady,
	StateLanding, StateDone, StateBlocked, StateOnHold,
}

// AllStates returns every state in declaration order.
func AllStates() []State { return append([]State(nil), allStates...) }

// Valid reports whether s is a member of the closed state set.
func (s State) Valid() bool {
	for _, v := range allStates {
		if s == v {
			return true
		}
	}
	return false
}

// IsPause reports whether s is blocked or on-hold. Neither is a progress
// stage; both record a return_stage and a resume condition.
func (s State) IsPause() bool { return s == StateBlocked || s == StateOnHold }

// IsProgress reports whether s is a valid forward stage (not a pause state).
func (s State) IsProgress() bool { return s.Valid() && !s.IsPause() }

// IsTerminal reports whether no transition leaves s.
func (s State) IsTerminal() bool { return s == StateDone }

// BlockReason is the closed reason set shared by blocked and on-hold
// (design §2.1a).
type BlockReason string

const (
	ReasonNotInstalled        BlockReason = "not-installed"
	ReasonConnectionFailed    BlockReason = "connection-failed"
	ReasonNotLoggedIn         BlockReason = "not-logged-in"
	ReasonNotConfigured       BlockReason = "not-configured"
	ReasonAuthFailed          BlockReason = "auth-failed"
	ReasonPermissionDenied    BlockReason = "permission-denied"
	ReasonRateLimited         BlockReason = "rate-limited"
	ReasonCapacityShort       BlockReason = "capacity-short"
	ReasonBudgetExhausted     BlockReason = "budget-exhausted"
	ReasonCheckNotPassed      BlockReason = "check-not-passed"
	ReasonDebtPresent         BlockReason = "debt-present"
	ReasonVerdictNotPass      BlockReason = "verdict-not-pass"
	ReasonProducerNotAllowed  BlockReason = "producer-not-allowed"
	ReasonOwnershipUnverified BlockReason = "ownership-unverified"
	ReasonStoreUnavailable    BlockReason = "store-unavailable"
	ReasonUserControl         BlockReason = "user-control"
	ReasonOperatorAckMissing  BlockReason = "operator-ack-missing"
	// ReasonOperatorHold is valid for on-hold only.
	ReasonOperatorHold BlockReason = "operator-hold"
)

var allReasons = []BlockReason{
	ReasonNotInstalled, ReasonConnectionFailed, ReasonNotLoggedIn, ReasonNotConfigured,
	ReasonAuthFailed, ReasonPermissionDenied, ReasonRateLimited, ReasonCapacityShort,
	ReasonBudgetExhausted, ReasonCheckNotPassed, ReasonDebtPresent, ReasonVerdictNotPass,
	ReasonProducerNotAllowed, ReasonOwnershipUnverified, ReasonStoreUnavailable,
	ReasonUserControl, ReasonOperatorAckMissing, ReasonOperatorHold,
}

// AllReasons returns every reason in declaration order.
func AllReasons() []BlockReason { return append([]BlockReason(nil), allReasons...) }

// Valid reports whether r is a member of the closed reason set.
func (r BlockReason) Valid() bool {
	for _, v := range allReasons {
		if r == v {
			return true
		}
	}
	return false
}

// ValidatePause checks that (state, reason) is a legal pause record: the state
// must be blocked or on-hold, the reason must be in the closed set, and
// operator-hold must not be used with blocked.
func ValidatePause(s State, r BlockReason) error {
	if !s.IsPause() {
		return fmt.Errorf("%w: %q", ErrNotPauseState, s)
	}
	if !r.Valid() {
		return fmt.Errorf("%w: %q", ErrReasonUnknown, r)
	}
	if r == ReasonOperatorHold && s != StateOnHold {
		return fmt.Errorf("%w: %q on %q", ErrReasonNotAllowed, r, s)
	}
	return nil
}

// Delivery is the command delivery stage (design §3). The four stages are
// recorded separately: sent is never evidence of received, and so on.
type Delivery string

const (
	DeliverySent      Delivery = "sent"
	DeliveryReceived  Delivery = "received"
	DeliveryExecuted  Delivery = "executed"
	DeliveryCompleted Delivery = "completed"
)

var deliveryOrder = []Delivery{DeliverySent, DeliveryReceived, DeliveryExecuted, DeliveryCompleted}

func (d Delivery) rank() int {
	for i, v := range deliveryOrder {
		if d == v {
			return i
		}
	}
	return -1
}

// Valid reports whether d is one of the four stages.
func (d Delivery) Valid() bool { return d.rank() >= 0 }

// AtLeast reports whether d has reached stage min. Invalid values never do.
func (d Delivery) AtLeast(min Delivery) bool {
	return d.Valid() && min.Valid() && d.rank() >= min.rank()
}
