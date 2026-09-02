// backlog_concurrency_test.go — the lost-update and id-collision guards under
// contention (SPEC-TODO-SQLITE-001 AC-TOSQ-009, REQ-TOSQ-005/008; M2).
//
// Factory mode runs up to ten lanes against ONE queue. The storage swap must
// not weaken that: the outer advisory lock still serializes the whole
// read-modify-write, and the engine's UNIQUE(id) plus the single-transaction
// last_seq advance are the backstop underneath it. Both are exercised here by
// contention, not by inspection.
package kanban

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// stressWriters and stressAddsPerWriter bound this file's stress fan-out. They
// are package-level with a SINGLE definition so the lock-wait budget guard
// (TestBoardLockWaitBudgetCoversSerializedMutations, board_lock_wait_test.go)
// reads the very figures TestConcurrencyStress serializes, rather than a
// hand-copied second pair (SPEC-STRESS-INVARIANT-VERDICT-001 REQ-SIV-011).
const (
	stressWriters       = 8
	stressAddsPerWriter = 6
)

// stressAddClass names the class one Add attempt's outcome falls into. The
// class is decided SOLELY by IsBoardLockHeld (board_lock.go) — never by
// matching error text, by an error count, or by elapsed time
// (SPEC-STRESS-INVARIANT-VERDICT-001 REQ-SIV-003).
type stressAddClass int

const (
	stressAddSucceeded stressAddClass = iota
	stressAddStarved
	stressAddHardFailed
)

func (c stressAddClass) String() string {
	switch c {
	case stressAddSucceeded:
		return "succeeded"
	case stressAddStarved:
		return "starved"
	default:
		return "hard-failed"
	}
}

// classifyStressAdd classifies one Add attempt's error (REQ-SIV-001/002/003).
func classifyStressAdd(err error) stressAddClass {
	switch {
	case err == nil:
		return stressAddSucceeded
	case IsBoardLockHeld(err):
		return stressAddStarved
	default:
		return stressAddHardFailed
	}
}

// stressTally accounts for every attempted add in exactly one class
// (REQ-SIV-008).
//
// successes counts Add calls that returned a NIL ERROR. It is deliberately NOT
// len(issued) / issuedCount, which counts DISTINCT ISSUED IDS and is the
// quantity the four invariants are anchored to. The two differ precisely when
// an id is issued twice: reading successes as len(issued) would collapse two
// successes into one id and break conservation for a reason that is not an
// accounting fault — that is invariant (a)'s job to report, not this identity's.
type stressTally struct {
	successes    int
	starved      int
	hardFailures []error
}

// record classifies err, folds it into the tally, and returns the class.
func (s *stressTally) record(err error) stressAddClass {
	class := classifyStressAdd(err)
	switch class {
	case stressAddSucceeded:
		s.successes++
	case stressAddStarved:
		s.starved++
	default:
		s.hardFailures = append(s.hardFailures, err)
	}
	return class
}

// attempts is the number of add attempts accounted for across all three
// classes.
func (s *stressTally) attempts() int {
	return s.successes + s.starved + len(s.hardFailures)
}

// zeroProgressVerdict returns the failure message for a total-starvation
// outcome, or "" when at least one add landed (REQ-SIV-007).
//
// It is a predicate over OUTCOMES — no wall clock, no fraction, no percentage
// success threshold. A fractional floor would be a load sensor wearing an
// accounting label, and would recreate on the next slower runner exactly the
// flake this SPEC removes.
func zeroProgressVerdict(s *stressTally) string {
	if s.successes > 0 {
		return ""
	}
	return fmt.Sprintf("0 of %d attempted adds succeeded (starved=%d, hard failures=%d) — "+
		"total starvation is a broken lock, not tolerable contention",
		s.attempts(), s.starved, len(s.hardFailures))
}

// AC-TOSQ-009 / REQ-TOSQ-008: N concurrent adders over one seeded root yield
// unique sequential ids and zero lost updates. Run with -race.
//
// The verdict criterion (SPEC-STRESS-INVARIANT-VERDICT-001, card t372). This
// test used ONE criterion for TWO unrelated properties: it failed when the
// queue invariants broke — correct — and failed identically when a contender
// exhausted the machine-speed-sensitive boardLockWaitBudget, which measures the
// runner rather than the code. Card t370 measured the consequence: 12 of 14
// non-cancelled CI runs red, every one of them at the lock-acquisition gate,
// and the invariants themselves broken in NONE of them.
//
// The criteria are now separate. A lock-acquisition failure is recorded as
// STARVED and tolerated; every other error class still fails hard immediately
// (REQ-SIV-002/003). The invariants below are what decides the verdict, they
// are anchored to the ids actually issued rather than to a static 48, and NONE
// of them is conditional on the starved count (REQ-SIV-006). Acquisition
// latency moved to its own machine-independent guard,
// TestBoardLockWaitBudgetCoversSerializedMutations (board_lock_wait_test.go).
func TestConcurrencyStress(t *testing.T) {
	t.Parallel()
	const attempted = stressWriters * stressAddsPerWriter
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))

	var mu sync.Mutex
	issued := make(map[string]int, attempted)
	var tally stressTally

	begin := time.Now()
	var start sync.WaitGroup
	start.Add(1)
	var done sync.WaitGroup
	for w := 0; w < stressWriters; w++ {
		done.Add(1)
		go func(w int) {
			defer done.Done()
			start.Wait() // release every writer into the lock at once
			for i := 0; i < stressAddsPerWriter; i++ {
				item, _, err := store.Add("card")
				mu.Lock()
				if tally.record(err) == stressAddSucceeded {
					issued[item.ID]++
				}
				mu.Unlock()
			}
		}(w)
	}
	start.Done()
	done.Wait()
	elapsed := time.Since(begin)

	// Any error class other than the contention sentinel still fails hard,
	// immediately, naming the returned error (REQ-SIV-002/003).
	if n := len(tally.hardFailures); n != 0 {
		t.Fatalf("%d/%d adds failed with an error that is not the contention sentinel; first: %v",
			n, attempted, tally.hardFailures[0])
	}

	// Zero progress is a broken lock, not tolerable contention (REQ-SIV-007).
	if msg := zeroProgressVerdict(&tally); msg != "" {
		t.Fatalf("%s", msg)
	}

	// Every attempted add is accounted for in exactly one class (REQ-SIV-008).
	// This replaces the retired `len(issued) != wantTotal` check: it counts
	// outcomes, never milliseconds, so it does not reintroduce the load-sensor
	// verdict, and it catches an accounting bug that silently drops successes —
	// which every invariant below would otherwise absorb by staying
	// self-consistent against a smaller issued set.
	if got := tally.attempts(); got != attempted {
		t.Fatalf("attempt conservation broken: successes(%d) + starved(%d) + hardFailures(%d) = %d, want %d",
			tally.successes, tally.starved, len(tally.hardFailures), got, attempted)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after stress: %v", err)
	}
	issuedCount := len(issued)
	stored := make(map[string]bool, len(rec.Items))
	for _, it := range rec.Items {
		stored[it.ID] = true
	}

	// --- The invariant block (REQ-SIV-005 (a)-(d)) ---
	// These four assertions, and only these four, are the verdict. Each runs
	// unconditionally: no `if starved == 0` guard, no t.Skip (REQ-SIV-006).
	// The Load error check above sits outside the block — a failure there
	// evidences a broken read path, not a broken invariant.

	// (a) no id collision: every issued id appears exactly once.
	for id, n := range issued {
		if n != 1 {
			t.Errorf("invariant (a) no id collision: id %s issued %d times — the mark advanced outside the lock",
				id, n)
		}
	}
	// (b) no lost update: every issued id is present in the loaded queue. A
	// lost update is invisible in the id set — it shows up only here, as an id
	// handed out and then overwritten by a concurrent whole-record write.
	for id := range issued {
		if !stored[id] {
			t.Errorf("invariant (b) no lost update: id %s was issued but is not in the queue", id)
		}
	}
	// (c) count consistency: stored item count equals the distinct issued count.
	if len(rec.Items) != issuedCount {
		t.Errorf("invariant (c) count consistency: stored items = %d, want %d (distinct issued ids) — %d updates were lost",
			len(rec.Items), issuedCount, issuedCount-len(rec.Items))
	}
	// (d) mark consistency: last_seq equals the distinct issued count.
	if rec.LastSeq != issuedCount {
		t.Errorf("invariant (d) mark consistency: last_seq = %d, want %d (distinct issued ids)",
			rec.LastSeq, issuedCount)
	}

	// Observability only — neither figure participates in the verdict
	// (REQ-SIV-012). The derived cost is what a latency regression looks like
	// in the CI stream now that latency no longer fails the test.
	t.Logf("SPEC-STRESS-INVARIANT-VERDICT-001: %d writers x %d adds = %d attempts; "+
		"%d succeeded, %d starved (tolerated), %d hard failures; "+
		"%d distinct ids, %d stored items, last_seq %d; "+
		"back-derived per-mutation cost %v (elapsed %v / %d successful mutations)",
		stressWriters, stressAddsPerWriter, attempted,
		tally.successes, tally.starved, len(tally.hardFailures),
		issuedCount, len(rec.Items), rec.LastSeq,
		elapsed/time.Duration(tally.successes), elapsed, tally.successes)
}

// AC-SIV-001 / REQ-SIV-001: a starved add is tolerated, produced
// DETERMINISTICALLY rather than by load.
//
// The holder is acquired in-process and released by a t.Cleanup-registered
// function — the pattern already in the tree at
// TestBacklogLockStuckHolderSurfacesBoundedNamedError. No background process is
// spawned, no machine load is generated, and no CI run is needed. Only two adds
// are attempted, so the bounded boardLockWaitBudget wait is paid twice rather
// than 48 times.
func TestStressAddClassificationToleratesStarvation(t *testing.T) {
	t.Parallel()

	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	held, err := acquireBoardLockImpl(store.LockPath())
	if err != nil {
		t.Fatalf("seeding the stuck holder: %v", err)
	}
	t.Cleanup(func() {
		if rerr := held.release(); rerr != nil {
			t.Logf("releasing the seeded holder: %v", rerr)
		}
	})

	const attempts = 2
	var tally stressTally
	for i := 0; i < attempts; i++ {
		_, _, addErr := store.Add("card")
		if addErr == nil {
			t.Fatalf("attempt %d succeeded while the lock was held by the seeded holder", i)
		}
		if !IsBoardLockHeld(addErr) {
			t.Fatalf("attempt %d error does not satisfy IsBoardLockHeld: %v", i, addErr)
		}
		if class := tally.record(addErr); class != stressAddStarved {
			t.Fatalf("attempt %d classified as %s, want starved", i, class)
		}
	}

	if tally.starved != attempts {
		t.Errorf("starved = %d, want %d", tally.starved, attempts)
	}
	if len(tally.hardFailures) != 0 {
		t.Errorf("hard failures = %d, want 0 — a contention error must not be classified as a hard failure",
			len(tally.hardFailures))
	}
	if tally.successes != 0 {
		t.Errorf("successes = %d, want 0", tally.successes)
	}
	if got := tally.attempts(); got != attempts {
		t.Errorf("conservation: accounted attempts = %d, want %d", got, attempts)
	}

	// The counterpart the tolerance must NOT swallow: any error class other
	// than the contention sentinel is still a hard failure (REQ-SIV-002/003).
	if class := classifyStressAdd(errors.New("some other failure")); class != stressAddHardFailed {
		t.Errorf("a non-sentinel error classified as %s, want hard-failed", class)
	}
	t.Logf("AC-SIV-001: %d/%d adds starved under a seeded holder, all satisfying IsBoardLockHeld, 0 hard failures",
		tally.starved, attempts)
}

// AC-SIV-005 / REQ-SIV-007: the zero-progress floor fails total starvation.
//
// Same seeded-holder construction as AC-SIV-001, with the holder held for the
// whole duration so that EVERY add is starved and zero succeed. The floor's
// predicate is exercised directly against that outcome, so the parent stress
// test need not itself be made to fail.
func TestStressZeroProgressFloorFailsTotalStarvation(t *testing.T) {
	t.Parallel()

	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	held, err := acquireBoardLockImpl(store.LockPath())
	if err != nil {
		t.Fatalf("seeding the stuck holder: %v", err)
	}
	t.Cleanup(func() {
		if rerr := held.release(); rerr != nil {
			t.Logf("releasing the seeded holder: %v", rerr)
		}
	})

	var tally stressTally
	_, _, addErr := store.Add("card")
	if addErr == nil {
		t.Fatalf("the add succeeded while the lock was held by the seeded holder")
	}
	if class := tally.record(addErr); class != stressAddStarved {
		t.Fatalf("the add classified as %s, want starved", class)
	}
	if tally.successes != 0 {
		t.Fatalf("successes = %d, want 0 — the zero-success outcome was not produced", tally.successes)
	}

	msg := zeroProgressVerdict(&tally)
	if msg == "" {
		t.Fatal("zeroProgressVerdict reported no failure for a zero-success outcome — " +
			"total starvation must fail, not be tolerated")
	}
	if !strings.Contains(msg, "broken lock") {
		t.Errorf("zero-progress message does not name total starvation as a broken lock: %q", msg)
	}

	// The floor admits any run that made progress — it is a zero-progress
	// floor, never a fractional or percentage success threshold.
	progressed := stressTally{successes: 1, starved: 47}
	if got := zeroProgressVerdict(&progressed); got != "" {
		t.Errorf("zeroProgressVerdict failed a run that made progress (1 success, 47 starved): %q — "+
			"the floor must not be a success-rate threshold", got)
	}
	t.Logf("AC-SIV-005: zero-success outcome rejected (%q); 1-success outcome admitted", msg)
}

// REQ-TOSQ-005: id integrity is enforced at the STORAGE layer, not only by the
// issuing code. A mutation that mints a duplicate id must abort the whole
// transaction and leave the prior state intact — fired here by handing the
// store a callback that deliberately duplicates an existing id, which is the
// shape a mis-ported issuer or a hand-edited import would take.
func TestDuplicateIDRejectedByStorage(t *testing.T) {
	t.Parallel()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if _, _, err := store.Add("original"); err != nil {
		t.Fatalf("seed Add: %v", err)
	}

	err := store.Mutate(func(rec *BacklogRecord) error {
		dup := rec.Items[0]
		dup.Text = "duplicate id smuggled in"
		rec.Items = append(rec.Items, dup)
		return nil
	})
	if err == nil {
		t.Fatal("Mutate(duplicate id) err = nil, want a storage-layer rejection")
	}
	if !IsBacklogIDConflict(err) {
		t.Errorf("err = %v, want it to satisfy IsBacklogIDConflict", err)
	}

	rec, loadErr := store.Load()
	if loadErr != nil {
		t.Fatalf("Load after rejected mutation: %v", loadErr)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("items = %d, want 1 — the aborted transaction must leave prior state intact", len(rec.Items))
	}
	if rec.Items[0].Text != "original" {
		t.Errorf("surviving item text = %q, want %q", rec.Items[0].Text, "original")
	}
}

// REQ-TOSQ-004: a `todo move`-style reorder round-trips. This is the case the
// design's "seq mirrors t<N>" note would have broken — after a move, array
// order and id order differ, and array order is the contract.
func TestReorderedItemsRoundTrip(t *testing.T) {
	t.Parallel()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	for _, text := range []string{"first", "second", "third"} {
		if _, _, err := store.Add(text); err != nil {
			t.Fatalf("Add(%s): %v", text, err)
		}
	}

	// Move the last card to the front, exactly as `todo move` rewrites the slice.
	if err := store.Mutate(func(rec *BacklogRecord) error {
		last := rec.Items[len(rec.Items)-1]
		rec.Items = append([]BacklogItem{last}, rec.Items[:len(rec.Items)-1]...)
		return nil
	}); err != nil {
		t.Fatalf("reorder Mutate: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after reorder: %v", err)
	}
	got := []string{rec.Items[0].ID, rec.Items[1].ID, rec.Items[2].ID}
	want := []string{"t3", "t1", "t2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order after reorder = %v, want %v (array position is the contract)", got, want)
		}
	}
}
