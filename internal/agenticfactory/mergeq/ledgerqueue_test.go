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
		epoch       ledger.QueueEpoch
		state       ledger.QueueState
		trunk, fail string
	}
	note      ledger.Note
	updateErr error
	items     []ledger.QueueItem
	cards     []ledger.Card
	leases    []ledger.Lease
}

func (f *fakeLedger) Queue(context.Context) ([]ledger.QueueItem, error) { return f.items, nil }
func (f *fakeLedger) Cards(context.Context) ([]ledger.Card, error)      { return f.cards, nil }
func (f *fakeLedger) Leases(context.Context) ([]ledger.Lease, error)    { return f.leases, nil }

func (f *fakeLedger) AcquireLock(_ context.Context, name string, h ledger.LockHolder, ttl time.Duration) (int64, error) {
	f.gotName, f.gotHolder, f.gotTTL = name, h, ttl
	return 7, f.lockErr
}
func (f *fakeLedger) ReleaseLock(context.Context, string, int64) error { return nil }
func (f *fakeLedger) QueueUpdate(_ context.Context, card string, epoch ledger.QueueEpoch, st ledger.QueueState, trunk, fail string) error {
	f.update.card, f.update.epoch, f.update.state, f.update.trunk, f.update.fail = card, epoch, st, trunk, fail
	return f.updateErr
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
	f.items = []ledger.QueueItem{{CardID: "t1", CardEpoch: 5, CandidateSHA: "c1"}}
	// A write holding the landing lock (lockEpoch 9) is presented as a lock epoch, one without it (0) as the
	// queue row's own CardEpoch (5); both name the candidate the caller verified (c1).
	for name, c := range map[string]struct {
		lock  int64
		state State
		want  ledger.QueueState
		epoch ledger.QueueEpoch
	}{
		"first waiting":         {0, StateWaiting, ledger.QueueWaiting, ledger.CardEpoch(5, "c1")},
		"verifying":             {0, StateVerifying, ledger.QueueVerifying, ledger.CardEpoch(5, "c1")},
		"failed outside lock":   {0, StateFailed, ledger.QueueFailed, ledger.CardEpoch(5, "c1")},
		"reconcile waiting":     {9, StateWaiting, ledger.QueueWaiting, ledger.LockEpoch(9, "c1")},
		"landing":               {9, StateLanding, ledger.QueueLanding, ledger.LockEpoch(9, "c1")},
		"landed":                {9, StateLanded, ledger.QueueLanded, ledger.LockEpoch(9, "c1")},
		"stale landing settled": {9, StateFailed, ledger.QueueFailed, ledger.LockEpoch(9, "c1")},
	} {
		if err := q.SetState(ctx, "t1", "c1", c.lock, c.state, "abc", "why"); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if f.update.state != c.want || f.update.card != "t1" || f.update.epoch != c.epoch || f.update.trunk != "abc" || f.update.fail != "why" {
			t.Errorf("%s -> %+v, want state %s epoch %+v with the other fields passed through", name, f.update, c.want, c.epoch)
		}
	}
	if err := q.SetState(ctx, "ghost", "c1", 0, StateVerifying, "abc", ""); err == nil {
		t.Error("a card that is not in the queue must be refused, not written with a guessed epoch")
	}
	f.updateErr = ledger.ErrStaleEpoch
	if err := q.SetState(ctx, "t1", "c1", 9, StateLanded, "abc", ""); !errors.Is(err, ErrStaleRow) {
		t.Errorf("a ledger ErrStaleEpoch came back as %v, want ErrStaleRow", err)
	}
	f.updateErr = nil
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
