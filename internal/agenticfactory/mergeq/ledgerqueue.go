package mergeq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// ledgerAPI is the part of *ledger.Store the merge queue uses; a fake of it tests the mapping.
type ledgerAPI interface {
	AcquireLock(ctx context.Context, name string, holder ledger.LockHolder, ttl time.Duration) (int64, error)
	ReleaseLock(ctx context.Context, name string, epoch int64) error
	QueueUpdate(ctx context.Context, cardID string, epoch ledger.QueueEpoch, state ledger.QueueState, trunkBefore, failReason string) error
	AddNote(ctx context.Context, n ledger.Note) error
	Queue(ctx context.Context) ([]ledger.QueueItem, error)
	Cards(ctx context.Context) ([]ledger.Card, error)
	Leases(ctx context.Context) ([]ledger.Lease, error)
}

var _ ledgerAPI = (*ledger.Store)(nil)

// LedgerQueue adapts the ledger to the Queue the merge queue codes against. The mapping is mechanical: the
// states and holder fields have the same meaning on both sides, and a held lock is the one error that is
// translated (ledger.ErrLockHeld to ErrLockHeld) because Land treats it as "wait", not "fail".
type LedgerQueue struct{ L ledgerAPI }

var _ Queue = LedgerQueue{}

// NewLedgerQueue wraps an open ledger store.
func NewLedgerQueue(s *ledger.Store) LedgerQueue { return LedgerQueue{L: s} }

// AcquireLock takes the named lock.
func (q LedgerQueue) AcquireLock(ctx context.Context, name string, h Holder, ttl time.Duration) (int64, error) {
	epoch, err := q.L.AcquireLock(ctx, name, ledger.LockHolder{SessionID: h.Session, PID: h.PID, PIDStart: h.PIDStart}, ttl)
	if errors.Is(err, ledger.ErrLockHeld) {
		return 0, ErrLockHeld
	}
	return epoch, err
}

// ReleaseLock gives the lock up.
func (q LedgerQueue) ReleaseLock(ctx context.Context, name string, epoch int64) error {
	return q.L.ReleaseLock(ctx, name, epoch)
}

type startEpochKey struct{ cardID string }

// withStartEpoch pins the card epoch a lander read from the queue row before it began verifying cardID.
func withStartEpoch(ctx context.Context, cardID string, epoch int64) context.Context {
	return context.WithValue(ctx, startEpochKey{cardID}, epoch)
}

func startEpochFrom(ctx context.Context, cardID string) (int64, bool) {
	n, ok := ctx.Value(startEpochKey{cardID}).(int64)
	return n, ok
}

// SetState records the queue row's state. A write that holds the landing lock (lockEpoch != 0: landing, landed,
// a reconcile's waiting, a stale landing row's settlement) presents it as a ledger.LockEpoch; one that does not
// (verifying, failed, a first waiting) presents the row's own CardEpoch, read from the queue row and never from
// cards.epoch: it is unverified here against the lease, and the ledger fences a stale row itself. Both kinds name
// the card's candidate commit the caller verified, so the two epoch numbers can never be mistaken for each other
// and a lander whose row was restarted with another candidate cannot move it. The ledger's ErrStaleEpoch is
// returned as ErrStaleRow.
func (q LedgerQueue) SetState(ctx context.Context, cardID, candidate string, lockEpoch int64, st State, trunkBefore, failReason string) error {
	epoch := ledger.LockEpoch(lockEpoch, candidate)
	if started, ok := startEpochFrom(ctx, cardID); ok && lockEpoch == 0 {
		// The caller pinned the card epoch it started verifying at; presenting that one (not the row's current one)
		// makes a write by a fenced lander stale instead of landing on the new holder's row.
		epoch = ledger.CardEpoch(started, candidate)
	} else if lockEpoch == 0 {
		items, err := q.L.Queue(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, it := range items {
			if it.CardID == cardID {
				epoch, found = ledger.CardEpoch(it.CardEpoch, candidate), true
				break
			}
		}
		if !found {
			return fmt.Errorf("mergeq: %s is not in the queue", cardID)
		}
	}
	err := q.L.QueueUpdate(ctx, cardID, epoch, ledger.QueueState(st), trunkBefore, failReason)
	if errors.Is(err, ledger.ErrStaleEpoch) {
		return fmt.Errorf("%w: %v", ErrStaleRow, err)
	}
	return err
}

// RowState reads the card's queue row state.
func (q LedgerQueue) RowState(ctx context.Context, cardID string) (State, error) {
	items, err := q.L.Queue(ctx)
	if err != nil {
		return "", err
	}
	for _, it := range items {
		if it.CardID == cardID {
			return State(it.State), nil
		}
	}
	return "", fmt.Errorf("mergeq: %s is not in the queue", cardID)
}

// Items reads the queue rows for Reconcile, each joined with its card's current epoch and whether the card holds
// a live lease at that epoch (a lease whose expiry has not passed): the facts that decide whether a verifying
// row may still belong to a live lander.
func (q LedgerQueue) Items(ctx context.Context, now time.Time) ([]Item, error) {
	rows, err := q.L.Queue(ctx)
	if err != nil {
		return nil, err
	}
	cards, err := q.L.Cards(ctx)
	if err != nil {
		return nil, err
	}
	leases, err := q.L.Leases(ctx)
	if err != nil {
		return nil, err
	}
	current := map[string]int64{}
	for _, c := range cards {
		current[c.ID] = c.Epoch
	}
	live := map[string]bool{}
	for _, l := range leases {
		if l.Epoch == current[l.CardID] && l.ExpiresAt.After(now) {
			live[l.CardID] = true
		}
	}
	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		out = append(out, Item{CardID: r.CardID, CandidateSHA: r.CandidateSHA, State: State(r.State),
			CardEpoch: r.CardEpoch, CurrentEpoch: current[r.CardID], LeaseLive: live[r.CardID]})
	}
	return out, nil
}

// Item reads one queue row joined with its card's epoch and lease (see Items).
func (q LedgerQueue) Item(ctx context.Context, cardID string, now time.Time) (Item, error) {
	items, err := q.Items(ctx, now)
	if err != nil {
		return Item{}, err
	}
	for _, it := range items {
		if it.CardID == cardID {
			return it, nil
		}
	}
	return Item{}, fmt.Errorf("mergeq: %s is not in the queue", cardID)
}

// AddNote appends a note of the given kind.
func (q LedgerQueue) AddNote(ctx context.Context, kind, cardID, text string) error {
	return q.L.AddNote(ctx, ledger.Note{Kind: ledger.NoteKind(kind), CardID: cardID, Text: text})
}
