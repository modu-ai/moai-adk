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
	"fmt"
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

func (m *Merger) lockTTL() time.Duration {
	if m.LockTTL > 0 {
		return m.LockTTL
	}
	return DefaultLockTTL
}

// Land verifies candidateSHA's merge into the trunk and moves the trunk to it.
//
// The gates run outside the lock. The lock is taken only after they pass; under it the trunk is read again, and
// if it moved the merge is rebuilt and the gates run again (at most MaxRetries times) instead of overwriting
// what landed meanwhile. The one write is `git update-ref <trunk> <new> <old>`, which itself refuses if the
// trunk moved after the re-check.
func (m *Merger) Land(ctx context.Context, cardID, candidateSHA string) (Result, error) {
	maxTries := m.MaxRetries
	if maxTries <= 0 {
		maxTries = DefaultMaxRetries
	}
	msg := fmt.Sprintf("merge(%s): land", cardID)
	if m.Message != nil {
		msg = m.Message(cardID)
	}
	settle := context.WithoutCancel(ctx) // releasing the lock and removing the worktree must survive a cancelled ctx

	trunk, err := m.Git.RevParse(ctx, m.Trunk)
	if err != nil {
		return Result{}, fmt.Errorf("mergeq: read the trunk %s: %w", m.Trunk, err)
	}
	in, err := m.Git.IsAncestor(ctx, candidateSHA, trunk)
	if err != nil {
		return Result{}, err
	}
	if in {
		return m.settleLanded(ctx, cardID, trunk, "already in the trunk")
	}

	res := Result{TrunkBefore: trunk}
	for attempt := 0; attempt < maxTries; attempt++ {
		if err := m.Queue.SetState(ctx, cardID, 0, StateVerifying, trunk, ""); err != nil {
			return res, err
		}
		cand, err := m.Git.PrepareCandidate(ctx, trunk, candidateSHA, msg)
		if err != nil {
			if errors.Is(err, ErrMerge) {
				return m.fail(ctx, res, cardID, trunk, err.Error())
			}
			return res, err
		}
		res.Verified++
		if err := m.Verify(ctx, cand.Dir); err != nil {
			_ = m.Git.Discard(settle, cand)
			return m.fail(ctx, res, cardID, trunk, "gates failed: "+err.Error())
		}

		epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
		if errors.Is(err, ErrLockHeld) {
			_ = m.Git.Discard(settle, cand)
			if err := m.Queue.SetState(ctx, cardID, 0, StateWaiting, trunk, ""); err != nil {
				return res, err
			}
			res.Outcome, res.Reason = OutcomeWaiting, "the landing lock is held by another lander"
			return res, nil
		}
		if err != nil {
			_ = m.Git.Discard(settle, cand)
			return res, err
		}

		next, landed, err := m.landUnderLock(ctx, cardID, epoch, trunk, cand)
		_ = m.Queue.ReleaseLock(settle, LandingLock, epoch)
		_ = m.Git.Discard(settle, cand)
		if err != nil {
			return res, err
		}
		if landed {
			res.Outcome, res.TrunkAfter = OutcomeLanded, cand.Result
			return res, nil
		}
		trunk, res.TrunkBefore = next, next // the trunk moved: rebuild on where it is now
	}
	return m.fail(ctx, res, cardID, trunk, fmt.Sprintf("the trunk moved %d times while the candidate was being verified", maxTries))
}

// landUnderLock runs with the landing lock held. It re-reads the trunk; if the trunk is where the merge was
// built it records landing, swaps the ref, records landed and the note. When the trunk moved (at the re-check
// or between it and the swap) it writes nothing and returns the trunk's new commit so the caller rebuilds.
func (m *Merger) landUnderLock(ctx context.Context, cardID string, epoch int64, trunk string, cand Candidate) (next string, landed bool, err error) {
	now, err := m.Git.RevParse(ctx, m.Trunk)
	if err != nil {
		return "", false, err
	}
	if now != trunk {
		return now, false, nil
	}
	if err := m.Queue.SetState(ctx, cardID, epoch, StateLanding, trunk, ""); err != nil {
		return "", false, err
	}
	if err := m.Git.UpdateRef(ctx, m.Trunk, cand.Result, trunk); err != nil {
		if !errors.Is(err, ErrRefMoved) {
			return "", false, err
		}
		now, rerr := m.Git.RevParse(ctx, m.Trunk)
		if rerr != nil {
			return "", false, rerr
		}
		return now, false, nil
	}
	if err := m.Queue.SetState(ctx, cardID, epoch, StateLanded, trunk, ""); err != nil {
		return "", false, err
	}
	if err := m.Queue.AddNote(ctx, "landed", cardID, fmt.Sprintf("%s -> %s on %s", trunk, cand.Result, m.Trunk)); err != nil {
		return "", false, err
	}
	return cand.Result, true, nil
}

func (m *Merger) fail(ctx context.Context, res Result, cardID, trunk, reason string) (Result, error) {
	if err := m.Queue.SetState(ctx, cardID, 0, StateFailed, trunk, reason); err != nil {
		return res, err
	}
	res.Outcome, res.Reason = OutcomeFailed, reason
	return res, nil
}

// settleLanded records a card whose candidate is already reachable from the trunk as landed. The ref is not
// touched; the lock is taken only because the landed state is written under it.
func (m *Merger) settleLanded(ctx context.Context, cardID, trunk, why string) (Result, error) {
	epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
	if errors.Is(err, ErrLockHeld) {
		return Result{Outcome: OutcomeWaiting, TrunkBefore: trunk, Reason: "the landing lock is held by another lander"}, nil
	}
	if err != nil {
		return Result{}, err
	}
	settle := context.WithoutCancel(ctx)
	defer func() { _ = m.Queue.ReleaseLock(settle, LandingLock, epoch) }()
	if err := m.Queue.SetState(ctx, cardID, epoch, StateLanded, trunk, ""); err != nil {
		return Result{}, err
	}
	if err := m.Queue.AddNote(ctx, "landed", cardID, why+" ("+trunk+")"); err != nil {
		return Result{}, err
	}
	return Result{Outcome: OutcomeLanded, TrunkBefore: trunk, TrunkAfter: trunk, Reason: why}, nil
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
	trunk, err := m.Git.RevParse(ctx, m.Trunk)
	if err != nil {
		return nil, fmt.Errorf("mergeq: read the trunk %s: %w", m.Trunk, err)
	}
	var acts []Action
	for _, it := range items {
		if it.State != StateVerifying && it.State != StateLanding {
			continue
		}
		in, err := m.Git.IsAncestor(ctx, it.CandidateSHA, trunk)
		if err != nil {
			return acts, err
		}
		if in {
			res, err := m.settleLanded(ctx, it.CardID, trunk, "found in the trunk after a restart")
			if err != nil {
				return acts, err
			}
			if res.Outcome == OutcomeLanded {
				acts = append(acts, Action{CardID: it.CardID, From: it.State, To: StateLanded, Why: "the candidate is reachable from the trunk"})
			}
			continue
		}
		if err := m.Queue.SetState(ctx, it.CardID, 0, StateWaiting, trunk, ""); err != nil {
			return acts, err
		}
		acts = append(acts, Action{CardID: it.CardID, From: it.State, To: StateWaiting, Why: "the candidate is not in the trunk; it is verified again"})
	}
	return acts, nil
}
