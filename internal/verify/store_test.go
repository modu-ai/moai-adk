package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestStoreLoadMissing asserts a missing snapshot returns (nil, nil) — absence
// is the plain re-execution path, never an error (strictly additive contract).
func TestStoreLoadMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	snap, err := Load(dir, "absent-head:absentdigest0000")
	if err != nil {
		t.Fatalf("missing snapshot must not error: %v", err)
	}
	if snap != nil {
		t.Fatalf("missing snapshot must be nil, got %+v", snap)
	}
}

// TestStoreSaveLoadRoundTrip asserts atomic save + load round-trips the
// snapshot document under .moai/state/verify/snapshots/.
func TestStoreSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "abcdef1234567890abcdef1234567890abcdef12:0123456789abcdef"
	in := &Snapshot{
		Key:        key,
		RecordedAt: time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC),
		Checks: []CheckEntry{
			{CheckID: "lint", Command: "golangci-lint run", ExitCode: 0,
				RecordedAt: time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC), DurationMS: 900},
		},
	}
	if err := Save(dir, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// The artifact lives under the gitignored runtime-state namespace.
	path := SnapshotPath(dir, key)
	if !strings.Contains(filepath.ToSlash(path), ".moai/state/verify/snapshots/") {
		t.Errorf("snapshot path must be under .moai/state/verify/snapshots/: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file must exist: %v", err)
	}
	out, err := Load(dir, key)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out == nil || out.Key != key || len(out.Checks) != 1 || out.Checks[0].Command != "golangci-lint run" {
		t.Fatalf("round-trip mismatch: %+v", out)
	}
}

// TestStoreRecordCheckReplaceOrAppend asserts RecordCheck replaces an entry
// with the identical command (fresher result for the same tree) and appends a
// distinct command.
func TestStoreRecordCheckReplaceOrAppend(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "head:digest"
	t0 := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)

	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 1, RecordedAt: t0}); err != nil {
		t.Fatalf("RecordCheck #1: %v", err)
	}
	// Same command → replace (exit updated 1 → 0).
	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "test", Command: "go test ./...", ExitCode: 0, RecordedAt: t0.Add(time.Minute)}); err != nil {
		t.Fatalf("RecordCheck #2: %v", err)
	}
	// Distinct command → append.
	if _, err := RecordCheck(dir, key, CheckEntry{CheckID: "lint", Command: "golangci-lint run", ExitCode: 0, RecordedAt: t0.Add(2 * time.Minute)}); err != nil {
		t.Fatalf("RecordCheck #3: %v", err)
	}

	snap, err := Load(dir, key)
	if err != nil || snap == nil {
		t.Fatalf("Load: snap=%v err=%v", snap, err)
	}
	if len(snap.Checks) != 2 {
		t.Fatalf("want 2 checks (replace + append), got %d: %+v", len(snap.Checks), snap.Checks)
	}
	e := snap.FindCommand("go test ./...")
	if e == nil || e.ExitCode != 0 {
		t.Fatalf("same-command re-record must replace the entry: %+v", e)
	}
	// Snapshot-level RecordedAt advances to the latest write.
	if !snap.RecordedAt.Equal(t0.Add(2 * time.Minute)) {
		t.Errorf("snapshot RecordedAt must track the latest write: %s", snap.RecordedAt)
	}
}

// TestStoreKeyMismatchGuard asserts Load rejects a snapshot file whose stored
// key differs from the requested key (filename-collision defense).
func TestStoreKeyMismatchGuard(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	key := "realhead:realdigest00000"
	if err := Save(dir, &Snapshot{Key: "other:key", RecordedAt: time.Now()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Manually place the other-key snapshot at this key's path.
	src := SnapshotPath(dir, "other:key")
	dst := SnapshotPath(dir, key)
	if err := os.Rename(src, dst); err != nil {
		t.Fatalf("rename: %v", err)
	}
	snap, err := Load(dir, key)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if snap != nil {
		t.Fatalf("key-mismatched snapshot must be treated as absent, got %+v", snap)
	}
}

// TestSnapshotPathWindowsSafe asserts the on-disk filename never carries the
// raw ':' key separator (Windows-invalid filename character).
func TestSnapshotPathWindowsSafe(t *testing.T) {
	t.Parallel()
	p := SnapshotPath(t.TempDir(), "abcdef1234567890abcdef1234567890abcdef12:0123456789abcdef")
	if strings.ContainsRune(filepath.Base(p), ':') {
		t.Errorf("snapshot filename must not contain ':' : %s", p)
	}
}
