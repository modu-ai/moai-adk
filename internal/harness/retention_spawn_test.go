// Package harness — retention spawn gate tests (SPEC-HARNESS-DETACHED-PRUNE-001).
// REQ-DP-002/004/007: the gate reads the stamp lock-free once, spawns only on a
// stale-or-absent stamp, suppresses on a fresh stamp, and returns the seam's
// error for the caller to log fail-open. Tests inject recording fakes as
// parameters — no test spawns a real detached child (REQ-DP-007).
package harness

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// writeSpawnStamp writes a prune stamp carrying the given wall-clock time next
// to the log path, so a test controls the gate's fresh/stale view through disk
// state alone (the gate reads time.Now internally; no clock injection needed).
func writeSpawnStamp(t *testing.T, logPath string, writtenAt time.Time) {
	t.Helper()
	if err := os.WriteFile(logPath+pruneStateSuffix, []byte(writtenAt.UTC().Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp file: %v", err)
	}
}

// recordingSpawn is the gate-test fake for the spawn seam: it counts invocations
// and records the (executable, args) pair it was handed.
type recordingSpawn struct {
	calls int
	exe   string
	args  []string
	err   error // returned verbatim when non-nil
}

func (r *recordingSpawn) spawn(executable string, args []string) error {
	r.calls++
	r.exe, r.args = executable, args
	return r.err
}

// TestMaybeSpawnRetentionPruner verifies the gate's stale-or-absent arm: the
// seam is invoked exactly once with the detached child's argv (REQ-DP-002).
func TestMaybeSpawnRetentionPruner(t *testing.T) {
	t.Run("stale_stamp_spawns_once_with_child_argv", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "usage-log.jsonl")
		writeSpawnStamp(t, logPath, time.Now().Add(-2*time.Hour))

		rec := &recordingSpawn{}
		if err := MaybeSpawnRetentionPruner(logPath, rec.spawn); err != nil {
			t.Fatalf("MaybeSpawnRetentionPruner returned error: %v", err)
		}
		if rec.calls != 1 {
			t.Fatalf("spawn calls: got=%d, want=1", rec.calls)
		}
		if rec.exe == "" {
			t.Error("executable handed to the spawn seam is empty")
		}
		wantArgs := []string{
			"hook", "retention-prune",
			"--log", logPath,
			"--archive", filepath.Join(dir, "learning-history", "archive"),
			"--days", strconv.Itoa(DefaultRetentionDays),
		}
		if !reflect.DeepEqual(rec.args, wantArgs) {
			t.Errorf("child argv mismatch:\ngot:  %v\nwant: %v", rec.args, wantArgs)
		}
	})

	t.Run("absent_stamp_spawns", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "usage-log.jsonl")

		rec := &recordingSpawn{}
		if err := MaybeSpawnRetentionPruner(logPath, rec.spawn); err != nil {
			t.Fatalf("MaybeSpawnRetentionPruner returned error: %v", err)
		}
		if rec.calls != 1 {
			t.Fatalf("spawn calls: got=%d, want=1", rec.calls)
		}
	})
}

// TestSpawnGateSuppressesOnFreshStamp verifies the once-per-interval gate: a
// fresh stamp spawns nothing and the gate's added synchronous work is that one
// lock-free stamp read (REQ-DP-002).
func TestSpawnGateSuppressesOnFreshStamp(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	writeSpawnStamp(t, logPath, time.Now().Add(-5*time.Minute))

	rec := &recordingSpawn{}
	if err := MaybeSpawnRetentionPruner(logPath, rec.spawn); err != nil {
		t.Fatalf("MaybeSpawnRetentionPruner returned error: %v", err)
	}
	if rec.calls != 0 {
		t.Errorf("fresh stamp spawned: calls=%d, want=0", rec.calls)
	}
}

// TestSpawnFailureFailOpen verifies the fail-open contract: a seam failure is
// returned to the caller verbatim — the wrapper logs it at exit 0 and the event
// stays on disk (REQ-DP-004).
func TestSpawnFailureFailOpen(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	writeSpawnStamp(t, logPath, time.Now().Add(-2*time.Hour))

	sentinel := errors.New("spawn boom")
	rec := &recordingSpawn{err: sentinel}
	err := MaybeSpawnRetentionPruner(logPath, rec.spawn)
	if !errors.Is(err, sentinel) {
		t.Fatalf("gate must return the seam's error verbatim: got=%v", err)
	}
	if rec.calls != 1 {
		t.Errorf("spawn calls: got=%d, want=1", rec.calls)
	}
}
