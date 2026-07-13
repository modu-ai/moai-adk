package verify

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestSourceLookupHit asserts an exact byte-string command match against a
// fresh snapshot returns the recorded exit code + a citable attribution
// (snapshot path + key).
func TestSourceLookupHit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "head:digest"
	now := time.Now()
	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: now}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	src := &Source{
		ProjectRoot: dir,
		KeyFunc:     func(context.Context, string) (string, error) { return key, nil },
		Now:         func() time.Time { return now },
	}
	exit, attr, ok := src.Lookup(context.Background(), "go test ./...")
	if !ok {
		t.Fatal("exact-match fresh entry must hit")
	}
	if exit != 0 {
		t.Errorf("exit: want 0, got %d", exit)
	}
	for _, tok := range []string{SnapshotPath(dir, key), key, "go test ./..."} {
		if !strings.Contains(attr, tok) {
			t.Errorf("attribution must cite %q: %s", tok, attr)
		}
	}
}

// TestSourceLookupMissAndNearMiss asserts a missing command and a near-miss
// variant (added flag) both miss — the caller re-executes (no normalization
// in the exact-match contract).
func TestSourceLookupMissAndNearMiss(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "head:digest"
	now := time.Now()
	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: now}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	src := &Source{
		ProjectRoot: dir,
		KeyFunc:     func(context.Context, string) (string, error) { return key, nil },
		Now:         func() time.Time { return now },
	}
	if _, _, ok := src.Lookup(context.Background(), "go vet ./..."); ok {
		t.Error("unrecorded command must miss")
	}
	if _, _, ok := src.Lookup(context.Background(), "go test -count=1 ./..."); ok {
		t.Error("near-miss command variant must miss (exact byte-string only)")
	}
}

// TestSourceLookupStale asserts a TTL-expired entry and a key-mismatched tree
// both miss (stale is never reusable).
func TestSourceLookupStale(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "head:digest"
	recorded := time.Now()

	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: recorded}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	// TTL-expired (clock injected past the default TTL).
	expired := &Source{
		ProjectRoot: dir,
		KeyFunc:     func(context.Context, string) (string, error) { return key, nil },
		Now:         func() time.Time { return recorded.Add(DefaultTTL + time.Minute) },
	}
	if _, _, ok := expired.Lookup(context.Background(), "go test ./..."); ok {
		t.Error("TTL-expired entry must miss")
	}

	// Key mismatch (tree changed since recording).
	mismatched := &Source{
		ProjectRoot: dir,
		KeyFunc:     func(context.Context, string) (string, error) { return "head:otherdigest", nil },
		Now:         func() time.Time { return recorded },
	}
	if _, _, ok := mismatched.Lookup(context.Background(), "go test ./..."); ok {
		t.Error("key-mismatched tree must miss")
	}
}

// TestSourceConstantCost is the Advisory-Check regression guard: key
// computation + snapshot load run at most ONCE per Source instance regardless
// of how many conditions are looked up — the per-turn cost is constant w.r.t.
// condition count (and w.r.t. corpus size: one key computation, one file read).
func TestSourceConstantCost(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "head:digest"
	now := time.Now()
	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: now}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	var keyCalls atomic.Int32
	src := &Source{
		ProjectRoot: dir,
		KeyFunc: func(context.Context, string) (string, error) {
			keyCalls.Add(1)
			return key, nil
		},
		Now: func() time.Time { return now },
	}
	for i := 0; i < 20; i++ {
		src.Lookup(context.Background(), "go test ./...")
		src.Lookup(context.Background(), "go vet ./...")
	}
	if got := keyCalls.Load(); got != 1 {
		t.Fatalf("key computation must run exactly once per Source (constant cost), ran %d times", got)
	}
}

// TestSourceDeadlineFallback asserts a key computation exceeding the time-box
// degrades to a miss (ok=false) — the caller re-executes; the turn is never
// blocked on the optimization (Advisory-Check Discipline).
func TestSourceDeadlineFallback(t *testing.T) {
	t.Parallel()
	src := &Source{
		ProjectRoot: t.TempDir(),
		Timeout:     10 * time.Millisecond,
		KeyFunc: func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done() // simulate a key computation that outlives the deadline
			return "", ctx.Err()
		},
	}
	start := time.Now()
	_, _, ok := src.Lookup(context.Background(), "go test ./...")
	if ok {
		t.Fatal("deadline-exceeded key computation must degrade to a miss")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("lookup must respect the time-box, took %s", elapsed)
	}
}

// TestSourceKeyErrorFallback asserts any key-computation error (e.g. non-git
// dir) degrades to a miss — fail-open, never fabricate.
func TestSourceKeyErrorFallback(t *testing.T) {
	t.Parallel()
	src := &Source{
		ProjectRoot: t.TempDir(),
		KeyFunc: func(context.Context, string) (string, error) {
			return "", errors.New("not a git repository")
		},
	}
	if _, _, ok := src.Lookup(context.Background(), "go test ./..."); ok {
		t.Fatal("key-computation error must degrade to a miss")
	}
}
