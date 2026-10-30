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
	"strings"
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

// DefaultTrunk is the dedicated trunk ref: no worktree has it checked out (leader decision). Lane trees follow it
// by merging it.
const DefaultTrunk = "refs/heads/af/trunk"

// ErrTrunkCheckedOut is returned, before anything is verified or locked, when a worktree has the trunk ref
// checked out: moving it with update-ref would leave that tree's index and files behind the ref.
var ErrTrunkCheckedOut = errors.New("mergeq: the trunk ref is checked out in a worktree; landing is refused")

// ErrBadTrunk is returned for a trunk that is not a full refs/heads/ branch name, or that names the project's
// default branch (main or master): this queue never moves the default branch.
var ErrBadTrunk = errors.New("mergeq: the trunk must be a full refs/heads/ name other than the default branch")

// ErrLockHeld is what AcquireLock returns when another holder has a live lock.
var ErrLockHeld = errors.New("mergeq: the landing lock is held")

// ErrStaleRow means the ledger refused a write because the row is not the caller's to move: the card was claimed
// again at a newer epoch, or the row was restarted with another candidate. A lander that gets it stops; Reconcile
// reads it as "leave this row to the card's holder", except for an older landing row, which it settles.
var ErrStaleRow = errors.New("mergeq: the queue row is not this caller's to move")

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
	// SetState records a queue row's state. candidate is the CARD's candidate commit the caller verified (the
	// Land argument or the row's candidate), never the merge-result commit on the trunk; the ledger refuses a
	// write that names another (ErrStaleRow), so a lander whose row was restarted with a new candidate cannot move
	// it, and once a landed settlement has made the row final a second candidate cannot be queued for the card.
	// lockEpoch is the landing lock's epoch for landing, landed, a reconcile's waiting and a stale landing row's
	// settlement, and zero for the states that are set without the lock.
	SetState(ctx context.Context, cardID, candidate string, lockEpoch int64, st State, trunkBefore, failReason string) error
	AddNote(ctx context.Context, kind, cardID, text string) error
	// RowState reads the card's queue row state, for the one path that settles a row without verifying it.
	RowState(ctx context.Context, cardID string) (State, error)
	// Item reads the card's queue row again with its card epoch and lease, as of now: Reconcile reads it under
	// the landing lock, so a lease renewed since the items were collected is seen.
	Item(ctx context.Context, cardID string, now time.Time) (Item, error)
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
	Now        func() time.Time           // the clock Reconcile judges leases by; nil means time.Now
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

	if err := m.validTrunk(ctx); err != nil {
		return Result{}, err
	}
	if err := m.noCheckout(ctx); err != nil {
		return Result{}, err
	}
	trunk, err := m.Git.RevParse(ctx, m.Trunk)
	if err != nil {
		return Result{}, fmt.Errorf("mergeq: read the trunk %s: %w", m.Trunk, err)
	}
	in, err := m.Git.IsAncestor(ctx, candidateSHA, trunk)
	if err != nil {
		return Result{}, err
	}
	if in {
		// The row's state decides which steps remain: a row already landing must not be written verifying.
		from, err := m.Queue.RowState(ctx, cardID)
		if err != nil {
			return Result{}, err
		}
		return m.settleLanded(ctx, cardID, candidateSHA, from, trunk, "already in the trunk")
	}

	res := Result{TrunkBefore: trunk}
	// A row a dead lander left in landing (its candidate is not in the trunk, checked above) cannot be written
	// verifying; reset it under the lock first.
	from, err := m.Queue.RowState(ctx, cardID)
	if err != nil {
		return res, err
	}
	if from == StateLanding {
		waiting, err := m.resetLanding(ctx, cardID, candidateSHA, trunk)
		if err != nil {
			return res, err
		}
		if waiting {
			res.Outcome, res.Reason = OutcomeWaiting, "the landing lock is held by another lander"
			return res, nil
		}
	}
	for attempt := 0; attempt < maxTries; attempt++ {
		if err := m.Queue.SetState(ctx, cardID, candidateSHA, 0, StateVerifying, trunk, ""); err != nil {
			return res, err
		}
		cand, err := m.Git.PrepareCandidate(ctx, trunk, candidateSHA, msg)
		if err != nil {
			if errors.Is(err, ErrMerge) {
				return m.fail(ctx, res, cardID, candidateSHA, trunk, err.Error())
			}
			return res, err
		}
		res.Verified++
		if err := m.Verify(ctx, cand.Dir); err != nil {
			_ = m.Git.Discard(settle, cand)
			return m.fail(ctx, res, cardID, candidateSHA, trunk, "gates failed: "+err.Error())
		}

		epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
		if errors.Is(err, ErrLockHeld) {
			_ = m.Git.Discard(settle, cand)
			// The row stays verifying: the ledger lets only a lock holder send verifying back to waiting, and a
			// later Land writes verifying again at the card epoch (verifying -> verifying is allowed).
			res.Outcome, res.Reason = OutcomeWaiting, "the landing lock is held by another lander"
			return res, nil
		}
		if err != nil {
			_ = m.Git.Discard(settle, cand)
			return res, err
		}

		next, landed, err := m.landUnderLock(ctx, cardID, candidateSHA, epoch, trunk, cand)
		relErr := m.Queue.ReleaseLock(settle, LandingLock, epoch)
		_ = m.Git.Discard(settle, cand)
		if relErr != nil {
			relErr = fmt.Errorf("mergeq: releasing the landing lock failed: %w", relErr)
		}
		if landed {
			// the trunk moved and the card is landed; a lock that could not be released is reported next to that
			res.Outcome, res.TrunkAfter = OutcomeLanded, cand.Result
			return res, relErr
		}
		if err != nil {
			return res, errors.Join(err, relErr)
		}
		if relErr != nil {
			return res, relErr
		}
		trunk, res.TrunkBefore = next, next // the trunk moved: rebuild on where it is now
	}
	return m.fail(ctx, res, cardID, candidateSHA, trunk, fmt.Sprintf("the trunk moved %d times while the candidate was being verified", maxTries))
}

// landUnderLock runs with the landing lock held. It re-reads the trunk; if the trunk is where the merge was
// built it records landing, swaps the ref, records landed and the note. When the trunk moved (at the re-check
// or between it and the swap) it writes nothing and returns the trunk's new commit so the caller rebuilds.
func (m *Merger) landUnderLock(ctx context.Context, cardID, candidateSHA string, epoch int64, trunk string, cand Candidate) (next string, landed bool, err error) {
	if err := m.noCheckout(ctx); err != nil {
		return "", false, err
	}
	now, err := m.Git.RevParse(ctx, m.Trunk)
	if err != nil {
		return "", false, err
	}
	if now != trunk {
		return now, false, nil
	}
	if err := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateLanding, trunk, ""); err != nil {
		return "", false, err
	}
	if err := m.Git.UpdateRef(ctx, m.Trunk, cand.Result, trunk); err != nil {
		// Whatever the error says, the trunk says what happened: read it.
		now, rerr := m.Git.RevParse(ctx, m.Trunk)
		switch {
		case rerr != nil:
			// The trunk cannot be read, so nothing can be judged; the row stays landing for Reconcile.
			return "", false, errors.Join(err, rerr)
		case now == cand.Result:
			// The swap took effect and only reported an error (git killed after it wrote the ref): the card
			// is landed, and the writes below say so.
		case errors.Is(err, ErrRefMoved):
			// The row is landing and the retry will write verifying, which the ledger refuses from landing: send
			// it back to waiting now, while the lock is still held (only the lock holder may).
			return now, false, m.leaveLanding(ctx, cardID, candidateSHA, epoch, trunk, nil)
		default:
			// The swap failed for another reason (a stale .lock file, a full disk): the trunk did not move, the
			// row must not stay landing.
			return "", false, m.leaveLanding(ctx, cardID, candidateSHA, epoch, trunk, err)
		}
	}
	if err := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateLanded, trunk, ""); err != nil {
		return "", false, err
	}
	if err := m.Queue.AddNote(ctx, "landed", cardID, fmt.Sprintf("%s -> %s on %s", trunk, cand.Result, m.Trunk)); err != nil {
		return "", false, err
	}
	return cand.Result, true, nil
}

// leaveLanding takes a landing row off landing while the lock is held: back to waiting, or, when even that write
// fails, to failed with the reason, so the row is never left where the next Land cannot start from. cause is the
// error that made the landing give up (nil when the trunk simply moved); it is returned, joined with any write
// error.
func (m *Merger) leaveLanding(ctx context.Context, cardID, candidateSHA string, epoch int64, trunk string, cause error) error {
	werr := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateWaiting, trunk, "")
	if werr == nil {
		return cause
	}
	reason := "the landing could not be undone: " + werr.Error()
	if cause != nil {
		reason = "the landing failed (" + cause.Error() + ") and could not be undone: " + werr.Error()
	}
	if ferr := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateFailed, trunk, reason); ferr != nil {
		return errors.Join(cause, werr, ferr)
	}
	return errors.Join(cause, werr)
}

// resetLanding takes a row left in landing by a lander that is gone back to waiting, under the lock, so that the
// verifying write that starts a new landing is allowed. It reports waiting=true when another lander holds the lock.
func (m *Merger) resetLanding(ctx context.Context, cardID, candidateSHA, trunk string) (waiting bool, err error) {
	epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
	if errors.Is(err, ErrLockHeld) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	defer func() {
		if relErr := m.Queue.ReleaseLock(context.WithoutCancel(ctx), LandingLock, epoch); relErr != nil {
			err = errors.Join(err, fmt.Errorf("mergeq: releasing the landing lock failed: %w", relErr))
		}
	}()
	return false, m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateWaiting, trunk, "")
}

// validTrunk refuses a trunk that is not a full branch name or that is the project's default branch.
func (m *Merger) validTrunk(ctx context.Context) error {
	t := m.Trunk
	switch {
	case !strings.HasPrefix(t, "refs/heads/") || len(t) == len("refs/heads/"):
		return fmt.Errorf("%w: %q", ErrBadTrunk, t)
	case t == "refs/heads/main" || t == "refs/heads/master":
		return fmt.Errorf("%w: %q is the default branch", ErrBadTrunk, t)
	}
	// A symbolic trunk would make every later check look at the link and every write land on its target.
	sym, err := m.Git.IsSymbolicRef(ctx, t)
	if err != nil {
		return fmt.Errorf("mergeq: look at whether %s is a symbolic ref: %w", t, err)
	}
	if sym {
		return fmt.Errorf("%w: %q is a symbolic ref", ErrBadTrunk, t)
	}
	def, err := m.Git.OriginHeadBranch(ctx)
	if err != nil {
		return fmt.Errorf("mergeq: read origin/HEAD: %w", err)
	}
	if def != "" && def == t {
		return fmt.Errorf("%w: %q is the branch origin/HEAD points at", ErrBadTrunk, t)
	}
	return nil
}

// noCheckout fails when a worktree has the trunk checked out. It runs before anything is verified and again
// under the lock, because a tree can be made after the first look.
func (m *Merger) noCheckout(ctx context.Context) error {
	holder, err := m.Git.CheckedOutAt(ctx, m.Trunk)
	if err != nil {
		return fmt.Errorf("mergeq: look for a worktree holding %s: %w", m.Trunk, err)
	}
	if holder != "" {
		return fmt.Errorf("%w: %s holds %s", ErrTrunkCheckedOut, holder, m.Trunk)
	}
	return nil
}

func (m *Merger) fail(ctx context.Context, res Result, cardID, candidateSHA, trunk, reason string) (Result, error) {
	if err := m.Queue.SetState(ctx, cardID, candidateSHA, 0, StateFailed, trunk, reason); err != nil {
		return res, err
	}
	res.Outcome, res.Reason = OutcomeFailed, reason
	return res, nil
}

// settleLanded records a card whose candidate is already reachable from the trunk as landed. The ref is not
// touched; the lock is taken only because the landed state is written under it.
func (m *Merger) settleLanded(ctx context.Context, cardID, candidateSHA string, from State, trunk, why string) (res Result, err error) {
	if from == StateLanded {
		// Settled already (a retry after a landed result that came back with a lock-release error): nothing to
		// write, and the table would refuse landed -> landing.
		return Result{Outcome: OutcomeLanded, TrunkBefore: trunk, TrunkAfter: trunk, Reason: "already settled as landed"}, nil
	}
	epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
	if errors.Is(err, ErrLockHeld) {
		return Result{Outcome: OutcomeWaiting, TrunkBefore: trunk, Reason: "the landing lock is held by another lander"}, nil
	}
	if err != nil {
		return Result{}, err
	}
	settle := context.WithoutCancel(ctx)
	defer func() {
		if relErr := m.Queue.ReleaseLock(settle, LandingLock, epoch); relErr != nil {
			err = errors.Join(err, fmt.Errorf("mergeq: releasing the landing lock failed: %w", relErr))
		}
	}()
	// The ledger moves a row to landed only from landing: a row that is still waiting or verifying goes through
	// verifying (a card-epoch write, allowed again from verifying) and landing first.
	if from == StateWaiting {
		if err := m.Queue.SetState(ctx, cardID, candidateSHA, 0, StateVerifying, trunk, ""); err != nil {
			return Result{}, err
		}
	}
	if from != StateLanding {
		if err := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateLanding, trunk, ""); err != nil {
			return Result{}, err
		}
	}
	if err := m.Queue.SetState(ctx, cardID, candidateSHA, epoch, StateLanded, trunk, ""); err != nil {
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
	// CardEpoch is the card epoch the row was queued at; CurrentEpoch is the card's epoch now; LeaseLive says the
	// card holds a live lease at CurrentEpoch. Reconcile sends a verifying row back to waiting only when the row is
	// from an older epoch or the card has no live lease: verifying is outside the landing lock, so a free lock does
	// not say no lander is alive. Zero values (no lease information) read as "no live lease".
	// Liveness is judged from the caller's clock against the lease expiry (not the database clock), and a dead
	// lane's verifying row stays as it is until its lease expires or is reaped: intended.
	CardEpoch, CurrentEpoch int64
	LeaseLive               bool
}

// landerMayBeAlive says a verifying row still belongs to a live card holder at the same epoch: its lander may be
// about to write landing, and sending the row back would make that write fail.
func (it Item) landerMayBeAlive() bool {
	return it.State == StateVerifying && it.CardEpoch == it.CurrentEpoch && it.LeaseLive
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
	var errs []error // one failing row does not stop the others
	for _, it := range items {
		if it.State != StateVerifying && it.State != StateLanding {
			continue
		}
		in, err := m.Git.IsAncestor(ctx, it.CandidateSHA, trunk)
		if err != nil {
			errs = append(errs, fmt.Errorf("mergeq: %s: %w", it.CardID, err))
			continue
		}
		if in {
			res, err := m.settleLanded(ctx, it.CardID, it.CandidateSHA, it.State, trunk, "found in the trunk after a restart")
			switch {
			case errors.Is(err, ErrStaleRow):
				continue // not this queue's to move: left for the card's holder
			case err != nil:
				errs = append(errs, fmt.Errorf("mergeq: %s: %w", it.CardID, err))
				continue
			}
			if res.Outcome == OutcomeLanded {
				acts = append(acts, Action{CardID: it.CardID, From: it.State, To: StateLanded, Why: "the candidate is reachable from the trunk"})
			}
			continue
		}
		if it.landerMayBeAlive() {
			continue
		}
		to, why, err := m.backToWaiting(ctx, it, trunk)
		if err != nil {
			errs = append(errs, fmt.Errorf("mergeq: %s: %w", it.CardID, err))
			continue
		}
		if to != "" {
			acts = append(acts, Action{CardID: it.CardID, From: it.State, To: to, Why: why})
		}
	}
	return acts, errors.Join(errs...)
}

// backToWaiting returns a row that is not in the trunk to waiting, and only while holding the landing lock: a
// row in verifying or landing may belong to a lander that is alive and about to finish, and the lock is what says
// no lander is. When the lock is held the row is left as it is.
//
// A row of an older card epoch (the card was claimed again after the row was written) is not this queue's to
// send back: the ledger refuses the write (ErrStaleRow). A landing row of that kind still has the lock holder's
// settlement, and its candidate is not in the trunk, so it is settled as failed; any other stale row is left for
// the card's holder, whose Enqueue restarts it.
func (m *Merger) backToWaiting(ctx context.Context, it Item, trunk string) (to State, why string, err error) {
	epoch, err := m.Queue.AcquireLock(ctx, LandingLock, m.Holder, m.lockTTL())
	if errors.Is(err, ErrLockHeld) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	defer func() {
		if relErr := m.Queue.ReleaseLock(context.WithoutCancel(ctx), LandingLock, epoch); relErr != nil {
			err = errors.Join(err, fmt.Errorf("mergeq: releasing the landing lock failed: %w", relErr))
		}
	}()
	if it.State == StateVerifying {
		// The items were collected before the lock; a lander may have renewed its lease since. Judge by a read
		// taken now, under the lock.
		now := time.Now
		if m.Now != nil {
			now = m.Now
		}
		fresh, ierr := m.Queue.Item(ctx, it.CardID, now())
		if ierr != nil {
			return "", "", ierr
		}
		if fresh.landerMayBeAlive() {
			return "", "", nil
		}
	}
	switch err := m.Queue.SetState(ctx, it.CardID, it.CandidateSHA, epoch, StateWaiting, trunk, ""); {
	case err == nil:
		return StateWaiting, "the candidate is not in the trunk; it is verified again", nil
	case errors.Is(err, ErrStaleRow) && it.State == StateLanding:
		reason := "candidate " + it.CandidateSHA + " is not an ancestor of the trunk " + trunk
		if err := m.Queue.SetState(ctx, it.CardID, it.CandidateSHA, epoch, StateFailed, trunk, reason); err != nil {
			return "", "", err
		}
		return StateFailed, "an older landing row whose candidate is not in the trunk is settled as failed", nil
	case errors.Is(err, ErrStaleRow):
		return "", "", nil
	default:
		return "", "", err
	}
}
