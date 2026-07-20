package verify

import (
	"context"
	"testing"
	"time"
)

// TestFreshnessPredicate covers the binary two-leg predicate: key equality AND
// wall-clock TTL — either leg failing ⇒ stale. Default 10-minute TTL applies
// when ttl <= 0; a configured value is honored.
func TestFreshnessPredicate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		storedKey  string
		currentKey string
		recordedAt time.Time
		ttl        time.Duration
		want       bool
	}{
		{"key match + in default TTL", "k1", "k1", now.Add(-9 * time.Minute), 0, true},
		{"key match + exactly at default TTL boundary", "k1", "k1", now.Add(-DefaultTTL), 0, true},
		{"key match + past default TTL", "k1", "k1", now.Add(-11 * time.Minute), 0, false},
		{"key mismatch + in TTL", "k1", "k2", now.Add(-1 * time.Minute), 0, false},
		{"key mismatch + past TTL", "k1", "k2", now.Add(-11 * time.Minute), 0, false},
		{"configured TTL honored (shorter)", "k1", "k1", now.Add(-2 * time.Minute), time.Minute, false},
		{"configured TTL honored (longer)", "k1", "k1", now.Add(-30 * time.Minute), time.Hour, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Fresh(tt.storedKey, tt.currentKey, tt.recordedAt, now, tt.ttl)
			if got != tt.want {
				t.Errorf("Fresh(%q,%q,rec=%s,ttl=%s) = %v, want %v",
					tt.storedKey, tt.currentKey, tt.recordedAt, tt.ttl, got, tt.want)
			}
		})
	}
}

// TestFreshness is the E2E freshness cycle against a real git repo:
// record → same-tree in-TTL check accepts → (a) mutate tracked file → stale,
// (b) add untracked file → stale, (c) injected clock past TTL → stale.
func TestFreshness(t *testing.T) {
	t.Parallel()
	dir := initTestRepo(t)
	ctx := context.Background()

	key, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key: %v", err)
	}
	recordedAt := time.Now()
	if _, err := RecordCheck(dir, key, CheckEntry{
		CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: recordedAt,
	}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	// Same tree, in-TTL → fresh (reuse accepted).
	snap, err := Load(dir, key)
	if err != nil || snap == nil {
		t.Fatalf("Load: snap=%v err=%v", snap, err)
	}
	curKey, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key(recheck): %v", err)
	}
	if !Fresh(snap.Key, curKey, snap.RecordedAt, time.Now(), 0) {
		t.Fatal("same-tree in-TTL snapshot must be fresh")
	}

	// (a) Mutate a tracked file → key changes → stale.
	writeFile(t, dir, "tracked.txt", "mutated\n")
	mutatedKey, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key(mutated): %v", err)
	}
	if Fresh(snap.Key, mutatedKey, snap.RecordedAt, time.Now(), 0) {
		t.Fatal("tracked-file mutation must make the snapshot stale")
	}
	// Restore the tree so the next leg isolates the untracked-file change.
	gitRun(t, dir, "checkout", "--", "tracked.txt")

	// (b) Add an untracked file → key changes → stale.
	writeFile(t, dir, "untracked.txt", "u\n")
	untrackedKey, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key(untracked): %v", err)
	}
	if Fresh(snap.Key, untrackedKey, snap.RecordedAt, time.Now(), 0) {
		t.Fatal("untracked-file add must make the snapshot stale")
	}

	// (c) Injected clock past the TTL (key unchanged) → stale.
	future := time.Now().Add(DefaultTTL + time.Minute)
	if Fresh(snap.Key, snap.Key, snap.RecordedAt, future, 0) {
		t.Fatal("TTL expiry must make the snapshot stale even on key equality")
	}
	// Configurable TTL honored on the same data.
	if !Fresh(snap.Key, snap.Key, snap.RecordedAt, future, time.Hour) {
		t.Fatal("configured longer TTL must be honored")
	}
}

// TestFreshnessStaleNeverReusable asserts the stale outcome is binary and
// non-reusable: a stale check answer never reports fresh on partial key match
// or slightly-past TTL (REQ-SNAP-011 — a stale snapshot is never citable
// evidence).
func TestFreshnessStaleNeverReusable(t *testing.T) {
	t.Parallel()
	now := time.Now()

	// Partial key match ("HEAD same, tree probably same") is NOT a match.
	if Fresh("head1:digestA", "head1:digestB", now, now, 0) {
		t.Fatal("partial key match (same HEAD, different tree digest) must be stale")
	}
	// Slightly past the TTL is stale — no grace beyond the boundary.
	past := now.Add(-DefaultTTL - time.Second)
	if Fresh("k", "k", past, now, 0) {
		t.Fatal("past-TTL snapshot must be stale (\"only 12 minutes old\" is stale)")
	}
}
