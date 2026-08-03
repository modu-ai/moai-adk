package verify

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestRecordCheckConcurrentSameKeyNoDimensionDrop verifies AC-AUDIT-SNAPSHOT-004
// (A4, REQ-004 ¶4 — the D3 claim/lock fix): concurrent RecordCheck invocations
// on the SAME HEAD-SHA key with DIFFERENT command dimensions MUST serialize so
// that every dimension lands in the recorded snapshot. Without the claim/lock,
// last-writer-wins would silently drop the other consumers' dimensions.
//
// This is the parallel variant acceptance.md mandates: "two consumers
// concurrently invoking RecordCheck on the same SHA with different command
// dimensions MUST NOT race last-writer-wins (one claim wins, the other reads);
// assert both dimensions end up recorded rather than one silently dropping the
// other."
func TestRecordCheckConcurrentSameKeyNoDimensionDrop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	key := "concurrenthead123:digest0001"
	t0 := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	// Four distinct diagnostic dimensions a sync cycle would record.
	entries := []CheckEntry{
		{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: t0, DurationMS: 1000},
		{CheckID: "lint", Command: "golangci-lint run", ExitCode: 0, RecordedAt: t0, DurationMS: 2000},
		{CheckID: "vet", Command: "go vet ./...", ExitCode: 0, RecordedAt: t0, DurationMS: 500},
		{CheckID: "cover", Command: "go test -cover ./...", ExitCode: 0, RecordedAt: t0, DurationMS: 1500},
	}

	const goroutines = 8 // >1 goroutine per dimension to amplify the race
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		for _, e := range entries {
			e := e
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := RecordCheck(dir, key, e); err != nil {
					t.Errorf("RecordCheck concurrent: %v", err)
				}
			}()
		}
	}
	wg.Wait()

	snap, err := Load(dir, key)
	if err != nil || snap == nil {
		t.Fatalf("Load after concurrent RecordCheck: snap=%v err=%v", snap, err)
	}

	// Every distinct command MUST be present — last-writer-wins would have
	// dropped some. FindCommand is exact-byte match.
	for _, e := range entries {
		if found := snap.FindCommand(e.Command); found == nil {
			t.Errorf("dimension %q was DROPPED by a concurrent writer — claim/lock failed to serialize RecordCheck; snapshot checks: %+v",
				e.Command, snap.Checks)
		} else if found.ExitCode != e.ExitCode {
			t.Errorf("dimension %q exit code mismatch: got %d, want %d", e.Command, found.ExitCode, e.ExitCode)
		}
	}
}

// TestRecordCheckConcurrentReplaceStillConsistent verifies the claim/lock does
// not break the replace-on-same-command contract: concurrent writers recording
// the SAME command (e.g. a re-run) still produce exactly one entry for it, not
// duplicates, and the entry is internally consistent.
func TestRecordCheckConcurrentReplaceStillConsistent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	key := "concurrentreplace:digest0002"
	t0 := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	const writers = 12
	cmd := "go test ./internal/verify/..."
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			exitCode := i % 2 // some pass, some fail — final state is whichever landed last
			if _, err := RecordCheck(dir, key, CheckEntry{
				CheckID:    "test",
				Command:    cmd,
				ExitCode:   exitCode,
				RecordedAt: t0.Add(time.Duration(i) * time.Second),
				DurationMS: int64(i * 100),
			}); err != nil {
				t.Errorf("RecordCheck concurrent same-command: %v", err)
			}
		}()
	}
	wg.Wait()

	snap, err := Load(dir, key)
	if err != nil || snap == nil {
		t.Fatalf("Load: snap=%v err=%v", snap, err)
	}
	// Exactly ONE entry for the command (replace-on-match, never duplicate).
	count := 0
	for _, c := range snap.Checks {
		if c.Command == cmd {
			count++
		}
	}
	if count != 1 {
		t.Errorf("same-command concurrent writers must yield exactly 1 entry (replace), got %d", count)
	}
}

// TestSnapshotAbsentForNewSHA verifies AC-AUDIT-SNAPSHOT-004b (A4 integrity):
// when HEAD advances from S1 to S2, a consumer requesting the snapshot for S2
// does NOT receive the S1-recorded result. Load returns (nil, nil) — absence
// is the explicit "no stale service" signal; the consumer then re-records or
// surfaces the absence per its documented contract.
func TestSnapshotAbsentForNewSHA(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s1Key := "sha1aaaaaa:digestS1____"
	s2Key := "sha2bbbbbb:digestS2____"

	// Record a snapshot at S1.
	if _, err := RecordCheck(dir, s1Key, CheckEntry{
		CheckID: "test", Command: "go test ./...", ExitCode: 0,
		RecordedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC), DurationMS: 1000,
	}); err != nil {
		t.Fatalf("RecordCheck S1: %v", err)
	}

	// A consumer requesting the S2 key MUST NOT receive the S1 result.
	s2Snap, err := Load(dir, s2Key)
	if err != nil {
		t.Fatalf("Load S2 (absent) must not error: %v", err)
	}
	if s2Snap != nil {
		t.Fatalf("S2 request must return nil (no stale S1 service), got snapshot with %d checks", len(s2Snap.Checks))
	}

	// The S1 snapshot is intact for S1 consumers (the new SHA invalidates only
	// for S2, it does not delete S1's record).
	s1Snap, err := Load(dir, s1Key)
	if err != nil || s1Snap == nil {
		t.Fatalf("S1 snapshot must still be loadable for S1 consumers: snap=%v err=%v", s1Snap, err)
	}
	// Sanity: the two keys resolve to distinct on-disk paths (no collision).
	if filepath.Clean(SnapshotPath(dir, s1Key)) == filepath.Clean(SnapshotPath(dir, s2Key)) {
		t.Error("S1 and S2 keys must resolve to distinct snapshot paths")
	}
}
