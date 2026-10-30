// Package mergeq lands a verified card onto the local trunk (Issue #1948, local-only AF redesign).
//
// A card is done when its merge result has passed the gates in a candidate worktree and the trunk ref has been
// moved to that result by a compare-and-swap. The queue never touches main or a remote: the trunk is a
// configured local ref. Verification runs outside the landing lock; the lock is held only for the re-check and
// the one `git update-ref`.
package mergeq

import (
	"context"
	"errors"
	"time"
)

// State is a merge-queue row state.
type State string

const (
	StateWaiting   State = "waiting"
	StateVerifying State = "verifying"
	StateLanding   State = "landing"
	StateLanded    State = "landed"
	StateFailed    State = "failed"
)

// LandingLock is the name of the one lock that serializes ref updates.
const LandingLock = "merge-queue:landing"

// Defaults, named so no call site carries a bare number (design §6).
const (
	DefaultLockTTL    = 10 * time.Minute
	DefaultMaxRetries = 3
)

// ErrLockHeld is what AcquireLock returns when another holder has a live lock.
var ErrLockHeld = errors.New("mergeq: the landing lock is held")

// Holder identifies who holds a lock: the session and the process (PID and its start fingerprint).
type Holder struct {
	Session  string
	PID      int
	PIDStart string
}

// Queue is what the merge queue needs from the ledger. The ledger package adapts to it.
type Queue interface {
	AcquireLock(ctx context.Context, name string, h Holder, ttl time.Duration) (epoch int64, err error)
	ReleaseLock(ctx context.Context, name string, epoch int64) error
	// SetState records a queue row's state. lockEpoch is the landing lock's epoch for landing and landed, and
	// zero for the states that are set without the lock.
	SetState(ctx context.Context, cardID string, lockEpoch int64, st State, trunkBefore, failReason string) error
	AddNote(ctx context.Context, kind, cardID, text string) error
}

// Verify runs the gates in dir, a worktree holding the merge result. A non-nil error fails the card.
type Verify func(ctx context.Context, dir string) error

// Outcome is how one Land call ended.
type Outcome string

const (
	OutcomeLanded  Outcome = "landed"
	OutcomeFailed  Outcome = "failed"
	OutcomeWaiting Outcome = "waiting" // another lander holds the lock; try again later, nothing failed
)

// Result is what Land reports.
type Result struct {
	Outcome     Outcome
	TrunkBefore string // the trunk commit the landed merge was built on
	TrunkAfter  string // the trunk commit after landing (the merge result), empty unless landed
	Reason      string
	Verified    int // how many times the gates ran (more than one means the trunk moved and it was redone)
}

// Merger lands cards.
type Merger struct {
	Git        Git
	Queue      Queue
	Trunk      string // the trunk ref, e.g. refs/heads/<integration branch>
	Verify     Verify
	Holder     Holder
	LockTTL    time.Duration
	MaxRetries int
	Message    func(cardID string) string // the merge commit message; must carry the card id
}

var errNotImplemented = errors.New("mergeq: not implemented")

// Land verifies candidateSHA's merge into the trunk and moves the trunk to it.
func (m *Merger) Land(ctx context.Context, cardID, candidateSHA string) (Result, error) {
	return Result{}, errNotImplemented
}

// Item is a queue row as the ledger holds it, for restart reconciliation.
type Item struct {
	CardID       string
	CandidateSHA string
	State        State
}

// Action is what Reconcile decided for one row.
type Action struct {
	CardID string
	From   State
	To     State
	Why    string
}

// Reconcile compares queue rows left in verifying or landing with git ancestry after a restart: a candidate
// already reachable from the trunk landed; one that is not goes back to waiting. It never moves a ref.
func (m *Merger) Reconcile(ctx context.Context, items []Item) ([]Action, error) {
	return nil, errNotImplemented
}
