package mergeq

// N66 (final audit of a36ea8c368): the check that a commit with a failing newest run is not landed (N54) ran once, before
// the gates, and a gate takes minutes; a failure recorded while it ran was landed all the same. Merger.LastCheck is asked
// again with the landing lock held, after the row says landing and immediately before the compare-and-swap of the trunk
// ref. A refusal moves nothing: the trunk stays, the row goes back to waiting, the lock is released, and the error is
// returned.

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

var errNotLandable = errors.New("the newest gate run of the commit failed")

func TestLand_RealLedger_ARefusedLastCheckMovesNothingAndLeavesTheRowWaiting(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	trunk := git(t, dir, "rev-parse", "refs/heads/trunk")
	m := realMerger(t, dir, s)
	var asked []string
	m.LastCheck = func(_ context.Context, cardID, candidateSHA string) error {
		asked = append(asked, cardID+"@"+candidateSHA)
		return errNotLandable
	}
	res, err := m.Land(ctx, "t1", cand)
	if !errors.Is(err, errNotLandable) {
		t.Fatalf("Land = %+v, %v; want the refusal of the last check returned", res, err)
	}
	if res.Outcome == OutcomeLanded {
		t.Errorf("outcome %q", res.Outcome)
	}
	if len(asked) != 1 || asked[0] != "t1@"+cand {
		t.Errorf("the last check was asked %v, want once, about the card and its candidate", asked)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got != trunk {
		t.Errorf("the trunk moved from %s to %s although the last check refused", trunk, got)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the row is %s, want waiting (the holder, or a successor, measures again)", got)
	}
	// the landing lock was given back
	epoch, err := NewLedgerQueue(s).AcquireLock(ctx, LandingLock, Holder{Session: "other", PID: os.Getpid(), PIDStart: "start-other"}, time.Minute)
	if err != nil {
		t.Fatalf("the landing lock is still held after the refusal: %v", err)
	}
	_ = NewLedgerQueue(s).ReleaseLock(ctx, LandingLock, epoch)
}

// The pair: a last check that agrees lands the card, and it is asked inside the landing lock with the trunk still where
// the merge was built (so the swap is the very next write).
func TestLand_RealLedger_ALastCheckThatAgreesIsAskedUnderTheLockBeforeTheSwap(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	trunk := git(t, dir, "rev-parse", "refs/heads/trunk")
	m := realMerger(t, dir, s)
	calls := 0
	m.LastCheck = func(ctx context.Context, _, _ string) error {
		calls++
		if _, err := NewLedgerQueue(s).AcquireLock(ctx, LandingLock, Holder{Session: "other", PID: os.Getpid(), PIDStart: "start-other"}, time.Minute); !errors.Is(err, ErrLockHeld) {
			t.Errorf("the last check ran without the landing lock held: %v", err)
		}
		if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got != trunk {
			t.Errorf("the trunk was already moved when the last check ran: %s", got)
		}
		if got := rowState(t, s, "t1"); got != ledger.QueueLanding {
			t.Errorf("the row is %s when the last check runs, want landing", got)
		}
		return nil
	}
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("Land = %+v, %v; want landed", res, err)
	}
	if calls != 1 {
		t.Errorf("the last check ran %d times, want once", calls)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got == trunk {
		t.Error("the trunk did not move")
	}
}

// A Merger with no last check lands as before (the library's other callers).
func TestLand_RealLedger_NoLastCheckLandsAsBefore(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	res, err := realMerger(t, dir, s).Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("Land = %+v, %v; want landed", res, err)
	}
}

// A last check that refuses on the second pass of a trunk that moved: the first pass was rebuilt, the second is refused,
// and the row is waiting, not landing.
func TestLand_RealLedger_ALastCheckThatRefusesAfterARebuildLeavesTheRowWaiting(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	moved := false
	m := realMerger(t, dir, s)
	m.Verify = func(_ context.Context, _ string) error {
		if !moved {
			moved = true
			advanceTrunk(t, dir, "other.txt", "o\n")
		}
		return nil
	}
	n := 0
	m.LastCheck = func(context.Context, string, string) error {
		n++
		return errNotLandable
	}
	if _, err := m.Land(ctx, "t1", cand); !errors.Is(err, errNotLandable) || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("Land: %v", err)
	}
	if n != 1 {
		t.Errorf("asked %d times, want once (the first pass was rebuilt before the check)", n)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the row is %s, want waiting", got)
	}
}
