package mergeq

import (
	"context"
	"errors"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// ledgerAPI is the part of *ledger.Store the merge queue uses; a fake of it tests the mapping.
type ledgerAPI interface {
	AcquireLock(ctx context.Context, name string, holder ledger.LockHolder, ttl time.Duration) (int64, error)
	ReleaseLock(ctx context.Context, name string, epoch int64) error
	QueueUpdate(ctx context.Context, cardID string, lockEpoch int64, state ledger.QueueState, trunkBefore, failReason string) error
	AddNote(ctx context.Context, n ledger.Note) error
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

// SetState records the queue row's state.
func (q LedgerQueue) SetState(ctx context.Context, cardID string, lockEpoch int64, st State, trunkBefore, failReason string) error {
	return q.L.QueueUpdate(ctx, cardID, lockEpoch, ledger.QueueState(st), trunkBefore, failReason)
}

// AddNote appends a note of the given kind.
func (q LedgerQueue) AddNote(ctx context.Context, kind, cardID, text string) error {
	return q.L.AddNote(ctx, ledger.Note{Kind: ledger.NoteKind(kind), CardID: cardID, Text: text})
}
