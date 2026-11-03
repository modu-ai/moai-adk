package mergeq

// Defect answer: the merge queue's value is in what it does when something under the lock goes wrong, and the
// happy-path tests never reached those lines. A ledger write that fails while the lock is held must return the
// error and still release the lock; a refused or unreadable repository must stop the landing before it writes;
// the output limit must stop a runaway git call. Each test names the failure it injects and checks what was left
// behind (the trunk, the lock), not only that an error came back.

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// errQueue is the fake queue with one injected failure: SetState to failState returns err.
type errQueue struct {
	*fakeQueue
	failState State
	err       error
	noteErr   error
}

func (q *errQueue) SetState(ctx context.Context, card, cand string, lock int64, st State, trunk, reason string) error {
	if st == q.failState {
		return q.err
	}
	return q.fakeQueue.SetState(ctx, card, cand, lock, st, trunk, reason)
}

func (q *errQueue) AddNote(ctx context.Context, kind, card, text string) error {
	if q.noteErr != nil {
		return q.noteErr
	}
	return q.fakeQueue.AddNote(ctx, kind, card, text)
}

func TestLand_AFailedLandingWriteReturnsTheErrorAndReleasesTheLock(t *testing.T) {
	boom := errors.New("ledger down")
	for _, st := range []State{StateVerifying, StateLanding, StateLanded} {
		dir := repo(t)
		cand := candidate(t, dir, "c1", "a.txt", "a\n")
		before := git(t, dir, "rev-parse", "refs/heads/trunk")
		fq := &fakeQueue{}
		m := newMerger(t, dir, fq, passing)
		m.Queue = &errQueue{fakeQueue: fq, failState: st, err: boom}
		_, err := m.Land(context.Background(), "t1", cand)
		if !errors.Is(err, boom) {
			t.Fatalf("%s write failing: %v, want the ledger error", st, err)
		}
		if fq.lockHeld() || fq.locks != fq.unlocks {
			t.Errorf("%s write failing: the lock leaked (acquired %d released %d)", st, fq.locks, fq.unlocks)
		}
		after := git(t, dir, "rev-parse", "refs/heads/trunk")
		// a failed `landed` write comes after the swap; the earlier states fail before it
		if st != StateLanded && after != before {
			t.Errorf("%s write failing: the trunk moved (%s -> %s)", st, before, after)
		}
	}
}

func TestLand_AFailedFailWriteIsReturned(t *testing.T) {
	boom := errors.New("ledger down")
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	fq := &fakeQueue{}
	m := newMerger(t, dir, fq, func(context.Context, string) error { return errors.New("gate red") })
	m.Queue = &errQueue{fakeQueue: fq, failState: StateFailed, err: boom}
	if _, err := m.Land(context.Background(), "t1", cand); !errors.Is(err, boom) {
		t.Fatalf("failed gates with a failing failed-write: %v, want the ledger error", err)
	}
}

func TestLand_ALandedNoteThatCannotBeWrittenIsReturnedAndTheLockReleased(t *testing.T) {
	boom := errors.New("notes down")
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	fq := &fakeQueue{}
	m := newMerger(t, dir, fq, passing)
	m.Queue = &errQueue{fakeQueue: fq, noteErr: boom}
	if _, err := m.Land(context.Background(), "t1", cand); !errors.Is(err, boom) {
		t.Fatalf("a failing landed note: %v, want that error", err)
	}
	if fq.lockHeld() {
		t.Error("the lock was not released after the note failed")
	}
}

// A candidate already in the trunk settles without a swap; each write along the way can fail and must be reported.
func TestSettleLanded_EveryWriteFailureIsReturnedAndTheLockReleased(t *testing.T) {
	boom := errors.New("ledger down")
	for _, tc := range []struct {
		from State
		fail State
	}{
		{StateWaiting, StateVerifying},
		{StateWaiting, StateLanding},
		{StateVerifying, StateLanding},
		{StateLanding, StateLanded},
	} {
		dir := repo(t)
		cand := advanceTrunk(t, dir, "a.txt", "a\n")
		fq := &fakeQueue{}
		m := newMerger(t, dir, fq, passing)
		m.Queue = &errQueue{fakeQueue: fq, failState: tc.fail, err: boom}
		if _, err := m.settleLanded(context.Background(), "t1", cand, tc.from, git(t, dir, "rev-parse", "refs/heads/trunk"), "test"); !errors.Is(err, boom) {
			t.Errorf("from %s with %s failing: %v, want the ledger error", tc.from, tc.fail, err)
		}
		if fq.lockHeld() {
			t.Errorf("from %s with %s failing: the lock leaked", tc.from, tc.fail)
		}
	}
}

func TestSettleLanded_ALandedRowAnswersLandedWithoutTakingTheLock(t *testing.T) {
	dir := repo(t)
	fq := &fakeQueue{}
	m := newMerger(t, dir, fq, passing)
	res, err := m.settleLanded(context.Background(), "t1", "c1", StateLanded, "trunk-sha", "retry")
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed", res, err)
	}
	if fq.locks != 0 {
		t.Errorf("an already-landed row took the lock %d times", fq.locks)
	}
}

func TestMerger_TheLockTTLDefaultsWhenUnsetAndIsKeptWhenSet(t *testing.T) {
	m := &Merger{}
	if got := m.lockTTL(); got != DefaultLockTTL {
		t.Errorf("unset TTL = %v, want %v", got, DefaultLockTTL)
	}
	m.LockTTL = 3 * time.Second
	if got := m.lockTTL(); got != 3*time.Second {
		t.Errorf("set TTL = %v, want 3s", got)
	}
}

// A repository git cannot read stops the landing before anything is written.
func TestLand_AnUnreadableRepositoryStopsBeforeAnyWrite(t *testing.T) {
	fq := &fakeQueue{}
	m := newMerger(t, t.TempDir(), fq, passing) // not a git repository
	if _, err := m.Land(context.Background(), "t1", "c1"); err == nil {
		t.Fatal("Land in a directory that is not a repository succeeded")
	}
	if len(fq.states) != 0 || fq.locks != 0 {
		t.Errorf("a queue write happened (states %v, locks %d) although git could not be read", fq.states, fq.locks)
	}
}

func TestGitCLI_ReadsAreErrorsOutsideARepository(t *testing.T) {
	g := GitCLI{Repo: t.TempDir(), ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	ctx := context.Background()
	if _, err := g.IsSymbolicRef(ctx, "refs/heads/trunk"); err == nil {
		t.Error("IsSymbolicRef outside a repository returned no error")
	}
	if _, err := g.CheckedOutAt(ctx, "refs/heads/trunk"); err == nil {
		t.Error("CheckedOutAt outside a repository returned no error")
	}
	if _, err := g.PrepareCandidate(ctx, "0000000000000000000000000000000000000000", "1111111111111111111111111111111111111111", "m"); err == nil {
		t.Error("PrepareCandidate outside a repository returned no error")
	}
}

func TestGitCLI_ASymbolicRefIsReportedAndAPlainOneIsNot(t *testing.T) {
	dir := repo(t)
	git(t, dir, "symbolic-ref", "refs/heads/link", "refs/heads/work")
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	ctx := context.Background()
	for ref, want := range map[string]bool{"refs/heads/link": true, "refs/heads/trunk": false, "refs/heads/absent": false} {
		got, err := g.IsSymbolicRef(ctx, ref)
		if err != nil || got != want {
			t.Errorf("IsSymbolicRef(%s) = %v, %v; want %v", ref, got, err, want)
		}
	}
}

func TestGitCLI_DiscardOfAnEmptyCandidateIsHarmless(t *testing.T) {
	g := GitCLI{Repo: repo(t), ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	if err := g.Discard(context.Background(), Candidate{}); err != nil {
		t.Errorf("Discard of a candidate that was never made: %v", err)
	}
}

func TestLimited_StopsWritingPastTheOutputLimit(t *testing.T) {
	var b bytes.Buffer
	l := &limited{b: &b}
	if _, err := l.Write(bytes.Repeat([]byte("x"), 16)); err != nil {
		t.Fatalf("a small write failed: %v", err)
	}
	if _, err := l.Write(make([]byte, maxGitOutput)); err == nil {
		t.Error("a write past the limit was accepted")
	}
	if b.Len() != 16 {
		t.Errorf("the buffer holds %d bytes after the refused write, want 16", b.Len())
	}
}

// RowState and Item answer for a card that is in the queue and refuse one that is not.
func TestLedgerQueue_RealLedger_RowStateAndItemRefuseACardNotInTheQueue(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	q := NewLedgerQueue(s)
	if st, err := q.RowState(ctx, "t1"); err != nil || st != State(ledger.QueueWaiting) {
		t.Errorf("RowState t1 = %q, %v; want waiting", st, err)
	}
	if it, err := q.Item(ctx, "t1", time.Now()); err != nil || it.CandidateSHA != cand || !it.LeaseLive {
		t.Errorf("Item t1 = %+v, %v; want the candidate and a live lease", it, err)
	}
	if _, err := q.RowState(ctx, "ghost"); err == nil {
		t.Error("RowState of a card that is not queued returned no error")
	}
	if _, err := q.Item(ctx, "ghost", time.Now()); err == nil {
		t.Error("Item of a card that is not queued returned no error")
	}
}
