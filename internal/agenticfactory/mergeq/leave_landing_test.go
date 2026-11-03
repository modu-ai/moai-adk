package mergeq

// Defect answer (final audit F4): leaveLanding's comment promised that a row whose waiting write failed would go to
// failed, but the failed write it tried used the lock-kind epoch, which the ledger refuses for failed, so the
// fallback never worked and its error muddied the report. A row whose waiting write fails stays landing, which the
// next Land resets under the lock; no failed write is attempted, the error says what failed, and the next Land lands.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// breakSwapOnce fails the first swap with an error that is not "the trunk moved".
type breakSwapOnce struct {
	GitCLI
	done bool
}

func (b *breakSwapOnce) UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error {
	if !b.done {
		b.done = true
		return errors.New("update-ref: cannot lock ref (simulated lock failure)")
	}
	return b.GitCLI.UpdateRef(ctx, ref, newSHA, oldSHA)
}

// failWaitingOnce fails the first waiting write and counts the failed writes that are attempted.
type failWaitingOnce struct {
	Queue
	waitingFailed bool
	failedWrites  int
}

func (q *failWaitingOnce) SetState(ctx context.Context, cardID, cand string, lock int64, st State, trunk, reason string) error {
	switch {
	case st == StateWaiting && !q.waitingFailed:
		q.waitingFailed = true
		return errors.New("simulated ledger write failure")
	case st == StateFailed:
		q.failedWrites++
	}
	return q.Queue.SetState(ctx, cardID, cand, lock, st, trunk, reason)
}

func TestLand_RealLedger_AFailedWaitingWriteLeavesTheRowLandingForTheNextLand(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	q := &failWaitingOnce{Queue: NewLedgerQueue(s)}
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}

	m := realMerger(t, dir, s)
	m.Queue = q
	m.Git = &breakSwapOnce{GitCLI: g}
	_, err := m.Land(ctx, "t1", cand)
	if err == nil || !strings.Contains(err.Error(), "cannot lock ref") || !strings.Contains(err.Error(), "simulated ledger write failure") {
		t.Fatalf("the error must name both failures: %v", err)
	}
	if q.failedWrites != 0 {
		t.Errorf("%d failed writes were attempted: that fallback cannot succeed (the ledger wants a card-kind epoch there)", q.failedWrites)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanding {
		t.Errorf("the row is %s, want landing (left for the next Land to reset)", got)
	}

	m.Queue = NewLedgerQueue(s)
	m.Git = g
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("the next Land: %+v %v, want landed", res, err)
	}
}
