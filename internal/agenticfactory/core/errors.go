package core

import "errors"

// Ownership and dispatch errors. Held (the board says another owner has it)
// stays distinct from Busy (the mutation was contended; retry), mirroring the
// internal/factory slot-lease split.
var (
	ErrHeld                = errors.New("agenticfactory: issue is owned by another session")
	ErrBusy                = errors.New("agenticfactory: claim store busy, retry")
	ErrNotHeld             = errors.New("agenticfactory: no claim held for this issue")
	ErrForeign             = errors.New("agenticfactory: claim belongs to a different holder")
	ErrStaleHead           = errors.New("agenticfactory: result is bound to a stale HEAD")
	ErrDuplicateDispatch   = errors.New("agenticfactory: same work already dispatched")
	ErrOwnershipUnverified = errors.New("agenticfactory: ownership could not be verified, write refused")
	ErrStoreUnavailable    = errors.New("agenticfactory: claim store unavailable")
	ErrDuplicateHandoff    = errors.New("agenticfactory: a handoff receipt for this HEAD already exists")
	ErrRepairOnly          = errors.New("agenticfactory: issue has a published PR, only a repair claim is allowed")
	ErrBacklogFull         = errors.New("agenticfactory: open PR backlog is at its limit, new claim held")
)

// Validation errors.
var (
	ErrIDInvalid        = errors.New("agenticfactory: invalid identifier")
	ErrHeadInvalid      = errors.New("agenticfactory: head must be a full 40-char lowercase sha")
	ErrIssueInvalid     = errors.New("agenticfactory: invalid issue number")
	ErrNotPauseState    = errors.New("agenticfactory: state is not blocked or on-hold")
	ErrReasonUnknown    = errors.New("agenticfactory: unknown block reason")
	ErrReasonNotAllowed = errors.New("agenticfactory: reason not allowed for this state")
)

// Flow errors.
var (
	ErrTransitionInvalid  = errors.New("agenticfactory: transition not allowed from the current state")
	ErrEvidenceIncomplete = errors.New("agenticfactory: required evidence is missing")
)

// ReasonFor maps an error to the block reason it records when the Issue must
// pause. Busy and Held are not blocks: Busy is a retry and Held is a
// statement about another owner.
func ReasonFor(err error) (BlockReason, bool) {
	switch {
	case errors.Is(err, ErrOwnershipUnverified):
		return ReasonOwnershipUnverified, true
	case errors.Is(err, ErrStoreUnavailable):
		return ReasonStoreUnavailable, true
	case errors.Is(err, ErrBacklogFull):
		return ReasonCapacityShort, true
	}
	return "", false
}
