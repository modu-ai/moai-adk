package mergeq

// Defect answer: these tests catch a state or holder that changes meaning crossing the adapter, a held lock
// reported as a generic error (Land would then fail a card that only had to wait), and any other error being
// swallowed or rewritten.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

type fakeLedger struct {
	lockErr   error
	gotHolder ledger.LockHolder
	gotTTL    time.Duration
	gotName   string
	update    struct {
		card        string
		epoch       int64
		state       ledger.QueueState
		trunk, fail string
	}
	note ledger.Note
}

func (f *fakeLedger) AcquireLock(_ context.Context, name string, h ledger.LockHolder, ttl time.Duration) (int64, error) {
	f.gotName, f.gotHolder, f.gotTTL = name, h, ttl
	return 7, f.lockErr
}
func (f *fakeLedger) ReleaseLock(context.Context, string, int64) error { return nil }
func (f *fakeLedger) QueueUpdate(_ context.Context, card string, epoch int64, st ledger.QueueState, trunk, fail string) error {
	f.update.card, f.update.epoch, f.update.state, f.update.trunk, f.update.fail = card, epoch, st, trunk, fail
	return nil
}
func (f *fakeLedger) AddNote(_ context.Context, n ledger.Note) error { f.note = n; return nil }

func TestLedgerQueue_MapsStatesHoldersAndNotes(t *testing.T) {
	f := &fakeLedger{}
	q := LedgerQueue{L: f}
	ctx := context.Background()

	epoch, err := q.AcquireLock(ctx, LandingLock, Holder{Session: "s1", PID: 42, PIDStart: "t0"}, DefaultLockTTL)
	if err != nil || epoch != 7 {
		t.Fatalf("epoch %d err %v", epoch, err)
	}
	if LandingLock != ledger.LockMerge {
		t.Errorf("the queue's lock name %q and the ledger's %q must be the same string", LandingLock, ledger.LockMerge)
	}
	if f.gotName != LandingLock {
		t.Errorf("the adapter asked the ledger for lock %q, want %q", f.gotName, LandingLock)
	}
	if f.gotHolder != (ledger.LockHolder{SessionID: "s1", PID: 42, PIDStart: "t0"}) || f.gotTTL != DefaultLockTTL {
		t.Errorf("holder %+v ttl %v", f.gotHolder, f.gotTTL)
	}
	for in, want := range map[State]ledger.QueueState{
		StateWaiting: ledger.QueueWaiting, StateVerifying: ledger.QueueVerifying, StateLanding: ledger.QueueLanding,
		StateLanded: ledger.QueueLanded, StateFailed: ledger.QueueFailed,
	} {
		if err := q.SetState(ctx, "t1", 9, in, "abc", "why"); err != nil {
			t.Fatal(err)
		}
		if f.update.state != want || f.update.card != "t1" || f.update.epoch != 9 || f.update.trunk != "abc" || f.update.fail != "why" {
			t.Errorf("%s -> %+v, want state %s with the other fields passed through", in, f.update, want)
		}
	}
	if err := q.AddNote(ctx, "landed", "t1", "text"); err != nil {
		t.Fatal(err)
	}
	if f.note.Kind != ledger.NoteLanded || f.note.CardID != "t1" || f.note.Text != "text" {
		t.Errorf("note %+v", f.note)
	}
}

func TestLedgerQueue_OnlyAHeldLockIsTranslated(t *testing.T) {
	f := &fakeLedger{lockErr: ledger.ErrLockHeld}
	q := LedgerQueue{L: f}
	if _, err := q.AcquireLock(context.Background(), LandingLock, Holder{}, time.Minute); !errors.Is(err, ErrLockHeld) {
		t.Errorf("a held lock = %v, want ErrLockHeld", err)
	}
	other := errors.New("disk I/O")
	f.lockErr = other
	if _, err := q.AcquireLock(context.Background(), LandingLock, Holder{}, time.Minute); !errors.Is(err, other) || errors.Is(err, ErrLockHeld) {
		t.Errorf("another error = %v, want it passed through and not read as a held lock", err)
	}
}

func TestStatesMatchTheLedgersQueueStates(t *testing.T) {
	for _, p := range [][2]string{
		{string(StateWaiting), string(ledger.QueueWaiting)}, {string(StateVerifying), string(ledger.QueueVerifying)},
		{string(StateLanding), string(ledger.QueueLanding)}, {string(StateLanded), string(ledger.QueueLanded)},
		{string(StateFailed), string(ledger.QueueFailed)},
	} {
		if p[0] != p[1] {
			t.Errorf("queue state %q is %q on the ledger side: the strings must be equal because the adapter converts by cast", p[0], p[1])
		}
	}
}
