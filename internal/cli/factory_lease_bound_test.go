// factory_lease_bound_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-007 (a busy record or a busy queue lock bounds the lease and does not
// starve the queue), AC-FAL-008 (the cap is derived from the queue lock's
// budget) and the re-entrancy guard of AC-FAL-011 (a public Mutate from inside
// a lease section times out instead of deadlocking or silently succeeding).
package cli

import (
	"context"
	"database/sql"
	"errors"
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

// flOverheadCeil is the stated ceiling on harness overhead — fixture setup, the
// in-process CLI exec, and teardown — that these record-stall arms may add above
// the product's own claim cap C (= factoryLeaseClaimWaitCap, 1 s). Measured
// (t1506): elapsed − C ranged 0.15–0.65 s across the 20 arms of the reproduction
// run (.moai/reports/t1506/repro-run2.log, -race -count=10, nominated max
// 0.652 s, bare max 0.275 s) and hit 0.657 s on CI run 37188480785. 700 ms is
// that observed maximum rounded up to a stated figure; the hang guard below
// spends four of them, so a runner needs overheads nearly 4x the worst
// measurement before the bound trips. The fix's own verification pass (a loaded
// machine, -race -count=10) later saw overhead up to 1.14 s — 1.6x this ceiling
// and still 2.5x inside the guard's allowance, which is the margin the multiple
// exists to provide; under the old C + 500 ms bound that pass would have failed
// 17 of its 20 arms.
const flOverheadCeil = 700 * time.Millisecond

// TestFactoryLeaseRecordStallBounded — AC-FAL-007 (a) and its bare-form arm (c)
// clause: a record write transaction held by a second connection for three
// times the cap C. The nominated lease gives up at C with exit 4 and a `raced`
// refusal whose detail says the record or the queue lock was busy (not
// "another lane"), the queue item is `queued` again, a queue writer started at
// the same instant completes without a lock-held error, and the queue's lock is
// acquirable right afterwards. The bare form gives up at C with exit 1 and
// leaves the promoted item `picked`.
//
// The wall-clock assertions here are hang guards, not the AC bound. The precise
// deadline is the product's own and lives in the code: the claim runs under
// context.WithTimeout(factoryLeaseClaimDeadline) (factoryClaimContext,
// factory_card.go) on a record connection whose busy timeout is
// factoryLeaseClaimBusyTimeout, and a busy refusal is marked only once that
// deadline has passed (factoryClaimFailure) — so the give-up is bounded by C and
// cannot happen before the deadline. Everything above C in a wall-clock reading
// is harness overhead, measured at 0.15–0.65 s (repro log) and 0.657 s (CI).
// AC-FAL-007's C + 500 ms margin budgets product-side delay only — one lock
// retry (50 ms), one queue mutation (33 ms), the claim's overshoot (~30 ms) and
// scheduler delay — not test-harness time, which is why it does not serve as the
// process wall-clock bound here. (Divergence from the AC's literal `C + 500 ms`
// wording noted for card t1506; the AC's semantic invariant is asserted
// unchanged below.)
func TestFactoryLeaseRecordStallBounded(t *testing.T) {
	limit := factoryLeaseClaimWaitCap + 4*flOverheadCeil
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
		if elapsed < factoryLeaseClaimDeadline {
			t.Errorf("the nominated lease returned after %s, below the claim's own %s deadline — with the record held the whole time the claim cannot have given up before waiting it out", elapsed, factoryLeaseClaimDeadline)
		}
		if elapsed > limit {
			t.Errorf("the nominated lease returned after %s with the record held, want within %s (C + 4x the measured overhead ceiling, hang guard)", elapsed, limit)
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
		if elapsed < factoryLeaseClaimDeadline {
			t.Errorf("the bare lease returned after %s, below the claim's own %s deadline — with the record held the whole time the claim cannot have given up before waiting it out", elapsed, factoryLeaseClaimDeadline)
		}
		if elapsed > limit {
			t.Errorf("the bare lease returned after %s with the record held, want within %s (C + 4x the measured overhead ceiling, hang guard)", elapsed, limit)
		}
		if code := nmExit(err); code != 1 && code != -1 {
			t.Errorf("the bare lease exited %d (%v), want an error (exit 1)", code, err)
		}
		if q != factory.BacklogStatePicked {
			t.Errorf("queue state after the bare stall = %s, want picked (spec §F R4, third shape)", q)
		}
	})
}

// flClaimSegmentMargin is the margin above the wait cap C allowed to the claim
// segment TestFactoryLeaseClaimRespectsWaitCap measures. The segment is nearly
// pure product code — the record is opened before the clock starts and the
// measured call is the claim itself — so the time above the deadline it waits
// out is only the cancellation propagating through one blocked write:
// measured (t1506) at 0.896–0.930 s across twenty -race runs on a loaded
// machine (two ten-run passes), i.e. 0.10–0.13 s above the 800 ms deadline and
// 0.07–0.10 s inside the cap. 250 ms is ~2x that worst observed overshoot,
// keeps the bound at over 1.3x the worst observed segment, and is a quarter of
// C, so the review's overlay mutation (the deadline alone moved to 2 s,
// segment ~2.1 s) fails the bound by ~0.85 s — a detector margin, not a
// measured ceiling.
const flClaimSegmentMargin = 250 * time.Millisecond

// TestFactoryLeaseClaimRespectsWaitCap — the mutation-detector for the claim
// cap itself (card-review round 1 P2, card t1506; AC-FAL-007's give-up bound).
// The end-to-end arms in TestFactoryLeaseRecordStallBounded are hang guards on
// the whole process wall-clock: a claim deadline loosened by overlay (the
// review's mutation moved only factoryClaimContext's limit, 800 ms → 2 s)
// gives up at ~2.4 s, inside the 3.8 s guard and above the 800 ms lower bound,
// so nothing failed. This test measures the claim segment directly and
// narrowly: it invokes factoryNextRecordAndClaim in-process, exactly as the
// nominated lease does (homestate.OpenFactoryBounded with the claim's busy
// timeout, factory_card.go factoryNextNominate), against a record write
// transaction held by a second connection (flHoldRecord), reading the clock
// immediately around the call — no CLI startup or teardown inside the measured
// segment. It then asserts AC-FAL-007's semantic invariant on that segment:
// the claim under a permanently held record returns no earlier than its own
// deadline (it waited the record out) and no later than the wait cap
// C = factoryLeaseClaimWaitCap plus the segment margin above — so a give-up
// past the declared cap fails here even while every wall-clock guard passes.
func TestFactoryLeaseClaimRespectsWaitCap(t *testing.T) {
	root, _ := nmQueuedNominee(t)
	db, err := homestate.OpenFactoryBounded(root, factoryLeaseClaimBusyTimeout)
	if err != nil {
		t.Fatalf("open the factory record with the claim's busy timeout: %v", err)
	}
	defer func() { _ = db.Close() }()
	end := flHoldRecord(t, root)
	start := time.Now()
	_, leased, retry, claimErr := factoryNextRecordAndClaim(context.Background(), db, root, fcRun, "t1", "lane-1", homestate.CardFields{})
	segment := time.Since(start)
	end()
	t.Logf("claim segment under a held record: %s busy=%v leased=%v retry=%v", segment, errors.Is(claimErr, errFactoryRecordBusy), leased, retry)
	if !errors.Is(claimErr, errFactoryRecordBusy) {
		t.Fatalf("the claim under a held record returned %v (leased=%v retry=%v), want the busy refusal errFactoryRecordBusy — without it the segment did not measure the give-up at the cap", claimErr, leased, retry)
	}
	if leased || retry {
		t.Errorf("the claim under a held record reported leased=%v retry=%v, want both false (the give-up is a hard busy refusal, not a race to retry)", leased, retry)
	}
	if segment < factoryLeaseClaimDeadline {
		t.Errorf("the claim segment returned after %s, below the claim's own %s deadline — with the record held the whole time the claim cannot have given up before waiting it out", segment, factoryLeaseClaimDeadline)
	}
	if segment > factoryLeaseClaimWaitCap+flClaimSegmentMargin {
		t.Errorf("the claim segment returned after %s, want within %s (the wait cap %s + %s segment margin) — the claim gave up past its declared cap", segment, factoryLeaseClaimWaitCap+flClaimSegmentMargin, factoryLeaseClaimWaitCap, flClaimSegmentMargin)
	}
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

// flQueueReleaseSlack is the slack past the queue lock's wait budget B that the
// release deadline carries. It is a hang/mutant bound, not a measurement: no
// pass/fail rides on elapsed wall-clock. The release is a state event the test
// controls — a context that starts counting at the lease's first queue-lock
// wait and expires B + flQueueReleaseSlack later — so a lease that gives up
// within its budget returns with the lock still held, with ~4 s of allowance
// above the worst measured harness overhead (1.14 s, the t1506 verification
// pass), while a lease that ignores its budget only returns after the
// deadline-driven release.
const flQueueReleaseSlack = 4 * time.Second

// flVerbResult carries one qasRunNext call's outcome from the goroutine that
// ran it to the test body.
type flVerbResult struct {
	out, stderr string
	err         error
	returned    time.Time
}

// flRunNextAsync runs qasRunNext on a goroutine and returns the channel its
// result arrives on, so the test body can check the order "the call returned ->
// the release signal" against the goroutine's send.
func flRunNextAsync(t *testing.T, args ...string) <-chan flVerbResult {
	t.Helper()
	done := make(chan flVerbResult, 1)
	go func() {
		out, stderr, err := qasRunNext(t, args...)
		done <- flVerbResult{out, stderr, err, time.Now()}
	}()
	return done
}

// flWaitStartRelease installs the lease's queue-lock wait seam
// (factoryLeaseWaitStart, AC-FAL-007 (c)) and arms the release of the held
// queue lock at a state event the test controls (card t1807's pattern): a
// context created at the seam's first firing — the instant the call reached its
// queue-lock wait — that expires releaseAfter later, at which point
// releaseSignal is closed and the queue lock is released. The release is
// deadline-driven, not a midpoint timer: nothing decides pass/fail on elapsed
// wall-clock, and the caller checks the order "the call returned -> the release
// signal" with a non-blocking receive. Creating the context at the wait, rather
// than before the verb, keeps the lock held past the release deadline during
// the wait whatever the pre-lock latency, which varies with the host's load.
// The getter returns the instant the seam fired and fails the test unless the
// seam fired exactly once: each lease call reads the clock once, immediately
// before its first queue-lock wait (for the bare form this is the one-attempt
// premise). The instant is read from the real clock inside the seam, not taken
// from its argument: the argument is factoryCardNow's read, and fcFixture pins
// factoryCardNow to a fixed 2026-09-26 date, so the argument cannot serve as a
// wall-clock start.
func flWaitStartRelease(t *testing.T, releaseAfter time.Duration, release func()) (func() time.Time, <-chan struct{}) {
	t.Helper()
	var (
		mu    sync.Mutex
		at    time.Time
		fired int
	)
	releaseSignal := make(chan struct{})
	releaserDone := make(chan struct{})
	prev := factoryLeaseWaitStart
	factoryLeaseWaitStart = func(time.Time) {
		mu.Lock()
		defer mu.Unlock()
		fired++
		if fired == 1 {
			at = time.Now()
			releaseCtx, cancel := context.WithTimeout(context.Background(), releaseAfter)
			// Cleanups run LIFO: cancel first (the releaser wakes), then the
			// done-wait, then the seam restore registered at the helper's call.
			t.Cleanup(func() { <-releaserDone })
			t.Cleanup(cancel)
			go func() {
				defer close(releaserDone)
				<-releaseCtx.Done()
				close(releaseSignal)
				release()
			}()
		}
	}
	t.Cleanup(func() {
		factoryLeaseWaitStart = prev
	})
	return func() time.Time {
		t.Helper()
		mu.Lock()
		defer mu.Unlock()
		switch {
		case fired == 0:
			t.Fatal("factoryLeaseWaitStart never fired: the lease reached no queue-lock wait, so the contract has no start (AC-FAL-007 (c))")
		case fired != 1:
			t.Fatalf("factoryLeaseWaitStart fired %d times, want exactly 1: the lease reads the clock once, immediately before its queue-lock wait (AC-FAL-007 (c))", fired)
		}
		return at
	}, releaseSignal
}

// TestFactoryLeaseQueueLockStallBounded — AC-FAL-007 (c): the queue lock held by
// the test for longer than the budget B. The contract is order, not elapsed
// wall-clock (card t1746, following card t1807's pattern): the lock is released
// at a state event the test controls — a release context that starts counting at
// the lease's first queue-lock wait (flWaitStartRelease) and expires
// B + flQueueReleaseSlack later — and the lease must return BEFORE that release
// signal fires: a lease that gives up within its budget returns with the lock
// still held, while a lease that ignores its budget only returns after the
// deadline-driven release. The verb runs on a goroutine (flRunNextAsync) and the
// order is checked with a non-blocking receive after the result arrives;
// elapsed time is logged for diagnostics only. The nominated lease then exits 4
// with a `raced` refusal whose detail names the queue lock and writes no record
// row; the bare form — one attempt, not five (the seam fired exactly once) —
// exits 1 with the lock's timeout error on standard error, nothing on standard
// output (in particular not `no card is available`), and no record row written.
func TestFactoryLeaseQueueLockStallBounded(t *testing.T) {
	t.Run("nominated", func(t *testing.T) {
		budget := flBudget(t)
		root, _ := nmQueuedNominee(t)
		nmIsolatedWorktrees(t, "t1")
		end := flHoldQueue(t, root)
		waitStart, released := flWaitStartRelease(t, budget+flQueueReleaseSlack, end)
		res := <-flRunNextAsync(t, "--run", fcRun, "--card", "t1")
		// Order check: the lease must return before the release signal fires.
		heldAtReturn := true
		select {
		case <-released:
			heldAtReturn = false
		default:
		}
		end()
		elapsed := res.returned.Sub(waitStart())
		t.Logf("nominated under a held queue lock: elapsed from the queue-lock wait=%s exit=%d err=%v stderr=%q", elapsed, nmExit(res.err), res.err, res.stderr)
		if !heldAtReturn {
			t.Errorf("the nominated lease returned only after the release signal fired (elapsed %s from the queue-lock wait, release deadline B + %s): the lease ignored its %s queue-lock wait budget", elapsed, flQueueReleaseSlack, budget)
		}
		nmAssertRefused(t, res.out, res.stderr, res.err, "raced")
		if !strings.Contains(res.stderr, "queue lock") {
			t.Errorf("refusal detail = %q, want it to name the queue lock", res.stderr)
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
		waitStart, released := flWaitStartRelease(t, budget+flQueueReleaseSlack, end)
		res := <-flRunNextAsync(t, "--run", fcRun)
		// Order check: the lease must return before the release signal fires.
		heldAtReturn := true
		select {
		case <-released:
			heldAtReturn = false
		default:
		}
		end()
		elapsed := res.returned.Sub(waitStart())
		t.Logf("bare under a held queue lock: elapsed from the queue-lock wait=%s exit=%d err=%v stdout=%q stderr=%q", elapsed, nmExit(res.err), res.err, res.out, res.stderr)
		if !heldAtReturn {
			t.Errorf("the bare lease returned only after the release signal fired (elapsed %s from the queue-lock wait, release deadline B + %s): the lease ignored its %s queue-lock wait budget, one attempt", elapsed, flQueueReleaseSlack, budget)
		}
		if err := res.err; err == nil || nmExit(err) == 3 || nmExit(err) == 0 {
			t.Errorf("the bare lease returned %v (exit %d), want an error with exit 1", err, nmExit(err))
		}
		if res.err != nil && !strings.Contains(res.stderr+res.err.Error(), "lock") {
			t.Errorf("the bare lease's error %q (stderr %q) does not carry the queue lock's timeout", res.err, res.stderr)
		}
		if res.out != "" {
			t.Errorf("the bare lease wrote stdout %q, want nothing (in particular not `no card is available`)", res.out)
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
