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

// TestOrphanedReclaimMarkerDoesNotWedgeTheBreak: a reclaimer that died
// holding its .reclaim marker must not wedge the break — the dead
// reclaimer's marker is reclaimable through the bare verified-dead rule,
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
