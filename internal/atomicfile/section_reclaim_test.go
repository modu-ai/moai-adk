package atomicfile

// The breaker-marker reclaim tests (review-gate finding, P1): reclaiming an
// orphaned .breaking marker was a non-atomic read→delete — two processes
// reclaiming the same dead marker could interleave so that one's late
// delete removed the other's LIVE re-acquired marker, reopening the
// breaker-vs-breaker race the marker exists to close. The reclaim now
// takes its own .reclaim marker first and disposes the breaker marker only
// while that creation holds: delete only on creation success.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMarkerReclaimRefusesWhileAReclaimerHoldsTheReclaimMarker pins the
// mutual exclusion: while a rival reclaimer holds <marker>.reclaim, this
// breaker must refuse to touch the breaker marker at all — the rival owns
// its disposal, and this caller's claim loop retries later.
func TestMarkerReclaimRefusesWhileAReclaimerHoldsTheReclaimMarker(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "queue.lock.breaking")
	previousBootFixture(t, markerPath) // a dead breaker's orphaned marker

	// A rival reclaimer holds the .reclaim marker: it owns this disposal.
	release, err := ClaimSection(context.Background(), markerPath+reclaimSuffix, 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("the test rival could not claim the reclaim marker: %v", err)
	}
	defer func() { _ = release() }()

	if BreakStaleLock(markerPath) {
		t.Fatal("the reclaim disposed a breaker marker while a rival reclaimer held the .reclaim marker — the disposal is not mutually excluded")
	}
	// The orphaned marker is untouched: the rival reclaimer will dispose it.
	if _, serr := os.Stat(markerPath); serr != nil {
		t.Fatal("the orphaned breaker marker was disposed while a rival held the reclaim marker")
	}
}

// TestReclaimMarkerReclaimIsGuardedToo pins the residual: the .reclaim
// marker's OWN reclaim was a bare verify-and-delete, so a reclaimer pausing
// between its check and its delete could remove a rival's LIVE re-acquired
// guard — the same defect the breaker-marker guard closed, one level down.
// The .reclaim reclaim takes its own guard first, exactly like the main
// path.
func TestReclaimMarkerReclaimIsGuardedToo(t *testing.T) {
	dir := t.TempDir()
	reclaimPath := filepath.Join(dir, "queue.lock.breaking.reclaim")
	previousBootFixture(t, reclaimPath) // a dead reclaimer's orphaned guard

	// A rival holds the reclaim marker's OWN guard: it owns this disposal.
	release, err := ClaimSection(context.Background(), reclaimPath+reclaimSuffix, 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("the test rival could not claim the guard: %v", err)
	}
	defer func() { _ = release() }()

	if BreakStaleLock(reclaimPath) {
		t.Fatal("the reclaim disposed a .reclaim marker while a rival reclaimer held its guard — the disposal is not mutually excluded")
	}
	if _, serr := os.Stat(reclaimPath); serr != nil {
		t.Fatal("the orphaned reclaim marker was disposed while a rival held its guard")
	}
}

// TestDeepDeadReclaimChainIsReclaimable pins the wedge repair (card
// t1606): the guard claim is a non-recursive bounded primitive, so a dead
// guard chain of ANY depth reclaims — each level is broken in turn and the
// walk never spawns a deeper guard. The former depth cap refused a chain
// at 3, and that refusal was permanent: no later walk could collect the
// chain, so a single dead reclaimer wedged the lock past the cap for the
// life of the boot — the rigidity this repair removes.
func TestDeepDeadReclaimChainIsReclaimable(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "queue.lock.breaking")
	previousBootFixture(t, markerPath)
	chain := []string{markerPath}
	p := markerPath
	for range 5 {
		p += reclaimSuffix
		previousBootFixture(t, p)
		chain = append(chain, p)
	}

	if !BreakStaleLock(markerPath) {
		t.Fatal("a deep dead guard chain wedged the reclaim — the guard claim must not recurse")
	}
	// The top marker and its immediate guard are gone. The deeper orphan
	// guards are collected on the way — a reclaim disposes its guard's
	// verified-dead rival, so an upper level may sweep the level below it —
	// and whichever survive are directly reclaimable. Either way NOTHING
	// wedges: after one pass every dead artifact of the chain is off the
	// disk, which is the contract the former depth cap could not honor.
	if _, serr := os.Stat(markerPath); serr == nil {
		t.Fatal("the dead breaker marker survived the reclaim")
	}
	if _, serr := os.Stat(markerPath + reclaimSuffix); serr == nil {
		t.Fatal("the dead reclaimer's guard survived the reclaim")
	}
	for _, deeper := range chain[2:] {
		if _, serr := os.Stat(deeper); serr != nil {
			continue // already swept as a rival guard by an upper reclaim
		}
		if _, serr := os.Stat(deeper); serr == nil && !BreakStaleLock(deeper) {
			t.Fatalf("an orphaned deep guard is not reclaimable: %s", deeper)
		}
	}
	for _, path := range chain {
		if _, serr := os.Stat(path); serr == nil {
			t.Fatalf("a dead chain marker survived every reclaim: %s", path)
		}
	}
}

// TestClaimGuardRivalDisposalIsGuardedToo pins the review-gate finding on
// card t1606 (P1): the guard claim's disposal of a DEAD rival guard must
// run through the guarded path — a bare verify-and-delete lets two
// reclaimers of the same dead guard interleave so one's late delete
// removes the other's LIVE re-acquired guard. While a rival holds the dead
// rival guard's OWN guard, this claim must refuse and touch nothing.
func TestClaimGuardRivalDisposalIsGuardedToo(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "queue.lock.breaking")
	previousBootFixture(t, markerPath) // the dead marker being reclaimed
	rivalGuard := markerPath + reclaimSuffix
	previousBootFixture(t, rivalGuard) // the dead rival guard blocking the claim

	// A rival reclaimer holds the dead rival guard's OWN guard: that
	// disposal is owned.
	release, err := ClaimSection(context.Background(), rivalGuard+reclaimSuffix, 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("the test rival could not claim the disposal: %v", err)
	}
	defer func() { _ = release() }()

	if BreakStaleLock(markerPath) {
		t.Fatal("the claim disposed a dead rival guard while a rival held ITS guard — the disposal is not serialized")
	}
	for _, p := range []string{markerPath, rivalGuard} {
		if _, serr := os.Stat(p); serr != nil {
			t.Fatalf("an owned disposal removed a marker: %s", p)
		}
	}
}

// TestBreakHonorsCallerCancellation pins the budget half of the hardening
// round: the guard reclaim must honor the CALLER's cancellation — a
// cancelled reclaim refuses before it reads anything, never entering the
// guard claim's retry loop.
func TestBreakHonorsCallerCancellation(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "queue.lock.breaking")
	previousBootFixture(t, markerPath)
	p := markerPath
	for range 5 { // a deep dead chain, deeper than any historical cap
		p += reclaimSuffix
		previousBootFixture(t, p)
	}

	// The verdict-read seam counts how far the walk got: a cancelled
	// caller's reclaim must refuse BEFORE reading anything.
	reads := 0
	prevRead := sectionRereadFn
	t.Cleanup(func() { sectionRereadFn = prevRead })
	sectionRereadFn = func(path string) ([]byte, error) {
		reads++
		return prevRead(path)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if BreakStaleLockContext(ctx, markerPath) {
		t.Fatal("a cancelled reclaim broke the lock")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("a cancelled reclaim ran %s — cancellation was not honored", elapsed)
	}
	if reads != 0 {
		t.Fatalf("a cancelled reclaim made %d verdict reads — it walked the chain", reads)
	}
}

// TestOrphanedReclaimMarkerDoesNotWedgeTheBreak: a reclaimer that died
// holding its .reclaim marker must not wedge the break — the dead
// reclaimer's marker is reclaimable through the guarded verified-dead rule,
// after which the breaker marker's own reclaim proceeds.
func TestOrphanedReclaimMarkerDoesNotWedgeTheBreak(t *testing.T) {
	dir := t.TempDir()
	markerPath := filepath.Join(dir, "queue.lock.breaking")
	previousBootFixture(t, markerPath)
	previousBootFixture(t, markerPath+reclaimSuffix) // the dead reclaimer

	if !BreakStaleLock(markerPath) {
		t.Fatal("a dead reclaimer's reclaim marker wedged the breaker marker's reclaim")
	}
	if _, serr := os.Stat(markerPath); serr == nil {
		t.Fatal("the orphaned breaker marker survived the reclaim")
	}
}
