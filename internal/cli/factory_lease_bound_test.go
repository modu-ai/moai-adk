// factory_lease_bound_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-007 (a busy record or a busy queue lock bounds the lease and does not
// starve the queue), AC-FAL-008 (the cap is derived from the queue lock's
// budget) and the re-entrancy guard of AC-FAL-011 (a public Mutate from inside
// a lease section times out instead of deadlocking or silently succeeding).
package cli

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flHoldRecord starts a write transaction on a second connection to the
// factory record and returns the function that ends it. The DSN takes the
// write lock at BEGIN, so every other record write waits while it is open.
func flHoldRecord(t *testing.T, root string) func() {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open the factory record for the holder: %v", err)
	}
	tx, err := db.DB.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		_ = db.Close()
		t.Fatalf("begin the holder's write transaction: %v", err)
	}
	// Take the write lock for real (a deferred BEGIN takes it at the first write).
	if _, err := tx.Exec(`UPDATE cards SET version = version WHERE 0`); err != nil {
		_ = tx.Rollback()
		_ = db.Close()
		t.Fatalf("take the holder's write lock: %v", err)
	}
	var once sync.Once
	end := func() {
		once.Do(func() {
			_ = tx.Rollback()
			_ = db.Close()
		})
	}
	t.Cleanup(end)
	return end
}

// flHoldQueue holds the queue's lock through an open public Mutate on another
// store value and returns the function that ends it.
func flHoldQueue(t *testing.T, root string) func() {
	t.Helper()
	held := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- todoStoreAt(root).Mutate(func(*factory.BacklogRecord) error {
			close(held)
			<-release
			return nil
		})
	}()
	select {
	case <-held:
	case err := <-finished:
		t.Fatalf("the test could not take the queue lock: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("the test never took the queue lock")
	}
	var once sync.Once
	end := func() {
		once.Do(func() {
			close(release)
			<-finished
		})
	}
	t.Cleanup(end)
	return end
}

// flQueueFreeWithin reports whether a no-op public Mutate completes within d
// (the queue's lock is acquirable).
func flQueueFreeWithin(store *factory.BacklogStore, d time.Duration) (bool, error) {
	op := flStartNoop(store)
	ok := op.within(d)
	return ok, op.err
}

// TestFactoryLeaseRecordStallBounded — AC-FAL-007 (a) and its bare-form arm (c)
// clause: a record write transaction held by a second connection for three
// times the cap C. The nominated lease returns within C + 500 ms with exit 4 and
// a `raced` refusal whose detail says the record or the queue lock was busy (not
// "another lane"), the queue item is `queued` again, a queue writer started at
// the same instant completes without a lock-held error, and the queue's lock is
// acquirable right afterwards. The bare form returns within C + 500 ms with exit
// 1 and leaves the promoted item `picked`.
func TestFactoryLeaseRecordStallBounded(t *testing.T) {
	limit := factoryLeaseClaimWaitCap + flMargin
	t.Run("nominated", func(t *testing.T) {
		root, store := nmQueuedNominee(t)
		nmIsolatedWorktrees(t, "t1")
		end := flHoldRecord(t, root)
		time.AfterFunc(3*factoryLeaseClaimWaitCap, end)
		writer := flStartNoop(store)
		start := time.Now()
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
		elapsed := time.Since(start)
		writerDone := writer.within(10 * time.Second)
		end()
		q := nmQueueState(t, store, "t1")
		t.Logf("nominated under a record stall: elapsed=%s exit=%d stderr=%q queue=%s", elapsed, nmExit(err), stderr, q)
		if elapsed > limit {
			t.Errorf("the nominated lease returned after %s with the record held, want within %s (C + 500 ms)", elapsed, limit)
		}
		nmAssertRefused(t, out, stderr, err, "raced")
		if !strings.Contains(stderr, "busy") || strings.Contains(stderr, "another lane") {
			t.Errorf("refusal detail = %q, want it to say the record or the queue lock was busy, and not \"another lane\"", stderr)
		}
		if q != factory.BacklogStateQueued {
			t.Errorf("queue state after the refused lease = %s, want queued (the compensation restores it)", q)
		}
		if !writerDone || writer.err != nil {
			t.Errorf("the queue writer started with the lease completed=%v err=%v, want completed without a lock-held error", writerDone, writer.err)
		}
		if ok, werr := flQueueFreeWithin(store, flMargin); !ok || werr != nil {
			t.Errorf("the queue's lock was not acquirable right after the lease (completed=%v err=%v)", ok, werr)
		}
	})
	t.Run("bare arm (c)", func(t *testing.T) {
		root, store := nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		nmIsolatedWorktrees(t, "t1")
		end := flHoldRecord(t, root)
		time.AfterFunc(3*factoryLeaseClaimWaitCap, end)
		start := time.Now()
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		elapsed := time.Since(start)
		end()
		q := nmQueueState(t, store, "t1")
		t.Logf("bare under a record stall: elapsed=%s exit=%d stdout=%q stderr=%q queue=%s", elapsed, nmExit(err), out, stderr, q)
		if elapsed > limit {
			t.Errorf("the bare lease returned after %s with the record held, want within %s (C + 500 ms)", elapsed, limit)
		}
		if code := nmExit(err); code != 1 && code != -1 {
			t.Errorf("the bare lease exited %d (%v), want an error (exit 1)", code, err)
		}
		if q != factory.BacklogStatePicked {
			t.Errorf("queue state after the bare stall = %s, want picked (spec §F R4, third shape)", q)
		}
	})
}

// TestFactoryLeaseMidClaimStallBounded — AC-FAL-007 (b): the same stall started
// only when the claim's third record write is about to begin (after the assign
// edge committed). The first invocation returns within C + 500 ms of the stall's
// start with exit 4 and the busy detail; the row is `assigned` to the lane and
// the queue item stays `picked`; a second invocation after the stall leases.
func TestFactoryLeaseMidClaimStallBounded(t *testing.T) {
	root, store := nmQueuedNominee(t)
	nmIsolatedWorktrees(t, "t1")
	var mu sync.Mutex
	writes := 0
	var stallStart time.Time
	var end func()
	prevNow := factoryCardNow
	factoryCardNow = func() time.Time {
		if flAtClaimWrite() {
			mu.Lock()
			writes++
			third := writes == 3
			mu.Unlock()
			if third {
				e := flHoldRecord(t, root)
				mu.Lock()
				stallStart, end = time.Now(), e
				mu.Unlock()
				time.AfterFunc(3*factoryLeaseClaimWaitCap, e)
			}
		}
		return prevNow()
	}
	t.Cleanup(func() { factoryCardNow = prevNow })

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	returned := time.Now()
	mu.Lock()
	started, stop := stallStart, end
	mu.Unlock()
	if stop == nil {
		t.Fatalf("the claim's third record write was never reached (writes=%d, err=%v stderr=%q)", writes, err, stderr)
	}
	stop()
	factoryCardNow = prevNow
	elapsed := returned.Sub(started)
	state, holder := flRow(t, root, "t1")
	q := nmQueueState(t, store, "t1")
	t.Logf("mid-claim stall: elapsed since stall start=%s exit=%d stderr=%q record=%s holder=%s queue=%s", elapsed, nmExit(err), stderr, state, holder, q)
	if limit := factoryLeaseClaimWaitCap + flMargin; elapsed > limit {
		t.Errorf("the lease returned %s after the stall began, want within %s (C + 500 ms)", elapsed, limit)
	}
	nmAssertRefused(t, out, stderr, err, "raced")
	if !strings.Contains(stderr, "busy") {
		t.Errorf("refusal detail = %q, want it to say the record was busy", stderr)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAssigned || c.OwnerLabel != "lane-1" {
		t.Errorf("t1 = %s owner=%q, want assigned to lane-1 (spec §F R4, second shape)", c.State, c.OwnerLabel)
	}
	if q != factory.BacklogStatePicked {
		t.Errorf("queue state = %s, want picked", q)
	}
	if _, stderr2, err2 := qasRunNext(t, "--run", fcRun, "--card", "t1"); err2 != nil {
		t.Errorf("the re-nomination after the stall: %v (stderr %q), want it to lease", err2, stderr2)
	} else {
		nmAssertLeased(t, root, "t1", "lane-1")
	}
}

// TestFactoryLeaseQueueLockStallBounded — AC-FAL-007 (c): the queue lock held by
// the test for longer than the budget B. The nominated lease returns within B +
// 500 ms with exit 4 and a detail naming the queue lock; the bare form returns
// within B + 500 ms — one attempt, not five — with exit 1, the lock's timeout
// error on standard error, nothing on standard output (in particular not `no
// card is available`), and no record row written.
func TestFactoryLeaseQueueLockStallBounded(t *testing.T) {
	t.Run("nominated", func(t *testing.T) {
		budget := flBudget(t)
		root, _ := nmQueuedNominee(t)
		nmIsolatedWorktrees(t, "t1")
		end := flHoldQueue(t, root)
		time.AfterFunc(budget+2*time.Second, end)
		start := time.Now()
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
		elapsed := time.Since(start)
		end()
		t.Logf("nominated under a held queue lock: elapsed=%s exit=%d err=%v stderr=%q", elapsed, nmExit(err), err, stderr)
		if limit := budget + flMargin; elapsed > limit {
			t.Errorf("the nominated lease returned after %s, want within %s (B + 500 ms)", elapsed, limit)
		}
		nmAssertRefused(t, out, stderr, err, "raced")
		if !strings.Contains(stderr, "queue lock") {
			t.Errorf("refusal detail = %q, want it to name the queue lock", stderr)
		}
		if n := nmRowCount(t, root, "t1"); n != 0 {
			t.Errorf("t1 has %d record rows, want 0", n)
		}
	})
	t.Run("bare", func(t *testing.T) {
		budget := flBudget(t)
		root, _ := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		nmIsolatedWorktrees(t, "t1", "t2")
		end := flHoldQueue(t, root)
		time.AfterFunc(budget+2*time.Second, end)
		start := time.Now()
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		elapsed := time.Since(start)
		end()
		t.Logf("bare under a held queue lock: elapsed=%s exit=%d err=%v stdout=%q stderr=%q", elapsed, nmExit(err), err, out, stderr)
		if limit := budget + flMargin; elapsed > limit {
			t.Errorf("the bare lease returned after %s, want within %s (B + 500 ms, one attempt)", elapsed, limit)
		}
		if err == nil || nmExit(err) == 3 || nmExit(err) == 0 {
			t.Errorf("the bare lease returned %v (exit %d), want an error with exit 1", err, nmExit(err))
		}
		if err != nil && !strings.Contains(stderr+err.Error(), "lock") {
			t.Errorf("the bare lease's error %q (stderr %q) does not carry the queue lock's timeout", err, stderr)
		}
		if out != "" {
			t.Errorf("the bare lease wrote stdout %q, want nothing (in particular not `no card is available`)", out)
		}
		for _, id := range []string{"t1", "t2"} {
			if n := nmRowCount(t, root, id); n != 0 {
				t.Errorf("%s has %d record rows, want 0", id, n)
			}
		}
	})
}

// TestFactoryLeaseCapWithinBoardBudget — AC-FAL-008: three times the claim's
// wait cap is not greater than the queue lock's wait budget as
// factory.LockWaitBudget() returns it; the guard reads the accessor, never a
// literal.
func TestFactoryLeaseCapWithinBoardBudget(t *testing.T) {
	budget := factory.LockWaitBudget()
	if factoryLeaseClaimWaitCap != factoryLeaseClaimDeadline+factoryLeaseClaimBusyTimeout {
		t.Errorf("factoryLeaseClaimWaitCap = %s, want the deadline (%s) plus the busy timeout (%s)", factoryLeaseClaimWaitCap, factoryLeaseClaimDeadline, factoryLeaseClaimBusyTimeout)
	}
	if 3*factoryLeaseClaimWaitCap > budget {
		t.Errorf("3 x factoryLeaseClaimWaitCap = %s > factory.LockWaitBudget() = %s; the claim cap must stay within one third of the queue lock's wait budget", 3*factoryLeaseClaimWaitCap, budget)
	}
}

// TestFactoryLeaseSectionRejectsNestedMutate — AC-FAL-011's re-entrancy guard
// (plan D4): a public Mutate called from inside a lease section (here, from the
// nomination seam) contends with the section's own descriptor and returns the
// lock-held timeout error within the budget plus 500 ms — so a future edit that
// calls the public Mutate inside the section is caught.
func TestFactoryLeaseSectionRejectsNestedMutate(t *testing.T) {
	budget := flBudget(t)
	_, store := nmQueuedNominee(t)
	nmIsolatedWorktrees(t, "t1")
	reached := false
	var nestedErr error
	var nestedElapsed time.Duration
	nmSetSeam(t, func(string) error {
		reached = true
		start := time.Now()
		nestedErr = store.Mutate(func(*factory.BacklogRecord) error { return nil })
		nestedElapsed = time.Since(start)
		return nil
	})
	_, _, verbErr := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if !reached {
		t.Fatalf("the nomination seam was never reached (verb err %v)", verbErr)
	}
	t.Logf("in-section Mutate: elapsed=%s err=%v", nestedElapsed, nestedErr)
	if nestedErr == nil || !factory.IsStateLockHeld(nestedErr) {
		t.Errorf("a public Mutate inside the lease section returned %v, want the lock-held timeout error (the section must hold the queue lock)", nestedErr)
	}
	if limit := budget + flMargin; nestedElapsed > limit {
		t.Errorf("the in-section Mutate returned after %s, want within %s (budget + 500 ms)", nestedElapsed, limit)
	}
}
