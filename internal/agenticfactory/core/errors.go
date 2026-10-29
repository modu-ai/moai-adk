package core

import "errors"

// Ownership and dispatch errors. Held (the board says another owner has it)
// stays distinct from Busy (the mutation was contended; retry), mirroring the
// internal/factory slot-lease split.
var (
	ErrHeld                 = errors.New("agenticfactory: issue is owned by another session")
	ErrBusy                 = errors.New("agenticfactory: claim store busy, retry")
	ErrNotHeld              = errors.New("agenticfactory: no claim held for this issue")
	ErrForeign              = errors.New("agenticfactory: claim belongs to a different holder")
	ErrStaleHead            = errors.New("agenticfactory: result is bound to a stale HEAD")
	ErrDuplicateDispatch    = errors.New("agenticfactory: same work already dispatched")
	ErrOwnershipUnverified  = errors.New("agenticfactory: ownership could not be verified, write refused")
	ErrStoreUnavailable     = errors.New("agenticfactory: claim store unavailable")
	ErrDuplicateHandoff     = errors.New("agenticfactory: a handoff receipt for this HEAD already exists")
	ErrRepairOnly           = errors.New("agenticfactory: issue has a published PR, only a repair claim is allowed")
	ErrIdentityMismatch     = errors.New("agenticfactory: session id, agent number and owned claims do not match")
	ErrBacklogUnset         = errors.New("agenticfactory: open PR backlog limit is not configured, new claim refused")
	ErrLimitInvalid         = errors.New("agenticfactory: open PR backlog limit must be at least 1")
	ErrRevisionConflict     = errors.New("agenticfactory: the authoritative copy moved on, write refused")
	ErrNotIssueAuthor       = errors.New("agenticfactory: the authenticated user is not the Issue's author for this operation")
	ErrIdentityUnknown      = errors.New("agenticfactory: the user, the author or the comment author could not be established")
	ErrRepoNotAllowed       = errors.New("agenticfactory: the repository is not on the collaboration allow list, or the host answered for another one")
	ErrPermissionNotGranted = errors.New("agenticfactory: collaboration does not grant this write")
	ErrOwnerUnconfirmed     = errors.New("agenticfactory: the Issue's current owner is not confirmed handed over or absent")
	ErrForeignClosing       = errors.New("agenticfactory: the PR text would close an Issue by another author")
	ErrEffectUnknown        = errors.New("agenticfactory: an effect may have happened and must be looked up, not repeated")
	ErrDeferred             = errors.New("agenticfactory: call budget is short, report write deferred")
	ErrHistoryRewritten     = errors.New("agenticfactory: the control ref history was rewritten, writes refused")
	ErrConflict             = errors.New("agenticfactory: the same request key carries a different payload")
	ErrIntakeFull           = errors.New("agenticfactory: intake is full, request rejected")
	ErrNoSlot               = errors.New("agenticfactory: no free worktree slot, request parked")
	ErrBacklogFull          = errors.New("agenticfactory: open PR backlog is at its limit, new claim held")
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
	case errors.Is(err, ErrBacklogFull), errors.Is(err, ErrIntakeFull), errors.Is(err, ErrNoSlot):
		return ReasonCapacityShort, true
	case errors.Is(err, ErrBacklogUnset):
		return ReasonNotConfigured, true
	case errors.Is(err, ErrNotIssueAuthor):
		return ReasonNotIssueAuthor, true
	case errors.Is(err, ErrIdentityUnknown):
		return ReasonIdentityUnknown, true
	case errors.Is(err, ErrRepoNotAllowed):
		return ReasonRepoNotAllowed, true
	case errors.Is(err, ErrPermissionNotGranted):
		return ReasonPermissionNotGranted, true
	case errors.Is(err, ErrOwnerUnconfirmed):
		return ReasonOwnerUnconfirmed, true
	case errors.Is(err, ErrForeignClosing):
		return ReasonForeignClosing, true
	}
	return "", false
}
