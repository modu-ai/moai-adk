// prune_logs_test.go — TDD coverage for SessionEnd observation-log pruning
// (SPEC-OBSERVE-HYGIENE-001 M2, REQ-OBH-002, AC-OBH-002).
//
// All tests operate exclusively under t.TempDir() — the LIVE .moai/logs/ dir is
// never touched by test code (B8 working-tree hygiene).
package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTrace creates a trace file under logsDir with the given session-id
// suffix, content, and mod time. Returns the full path.
func writeTrace(t *testing.T, logsDir, sessionID, content string, modTime time.Time) string {
	t.Helper()
	name := "trace-" + sessionID + ".jsonl"
	if content == "" {
		// zero-byte file
		path := filepath.Join(logsDir, name)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatal(err)
		}
		return path
	}
	path := filepath.Join(logsDir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	return path
}

// AC-OBH-002 — zero-byte trace files are pruned unconditionally at SessionEnd.
func TestPruneObservationLogs_ZeroBytePruned(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	// zero-byte stale trace
	writeTrace(t, logsDir, "sess-dead-001", "", now.Add(-2*time.Hour))
	// zero-byte fresh trace
	writeTrace(t, logsDir, "sess-dead-002", "", now)

	stats := PruneObservationLogs(logsDir, "current-session", 30, now)
	if stats.TraceZeroBytePruned != 2 {
		t.Errorf("TraceZeroBytePruned = %d, want 2", stats.TraceZeroBytePruned)
	}
	// Both zero-byte files removed.
	entries, _ := os.ReadDir(logsDir)
	if len(entries) != 0 {
		t.Errorf("expected 0 remaining files, got %d", len(entries))
	}
}

// AC-OBH-002 — non-empty traces older than retention are pruned; fresh ones kept.
func TestPruneObservationLogs_AgedPruned(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	// over-threshold (45 days old)
	writeTrace(t, logsDir, "sess-old-001", `{"event":"x"}`+"\n", now.Add(-45*24*time.Hour))
	// under-threshold (5 days old)
	writeTrace(t, logsDir, "sess-fresh-001", `{"event":"y"}`+"\n", now.Add(-5*24*time.Hour))

	stats := PruneObservationLogs(logsDir, "current-session", 30, now)
	if stats.TraceAgedPruned != 1 {
		t.Errorf("TraceAgedPruned = %d, want 1", stats.TraceAgedPruned)
	}
	// The fresh trace survives; the old one is gone.
	if _, err := os.Stat(filepath.Join(logsDir, "trace-sess-fresh-001.jsonl")); err != nil {
		t.Errorf("fresh trace should survive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(logsDir, "trace-sess-old-001.jsonl")); err == nil {
		t.Errorf("aged trace should be pruned")
	}
}

// EC-3 — the current session's active trace is NEVER pruned (even if zero-byte
// or over-threshold).
func TestPruneObservationLogs_CurrentSessionTracePreserved(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	currentSession := "current-live-session"
	// current session's trace, over-threshold AND non-empty — must survive
	writeTrace(t, logsDir, currentSession, `{"event":"live"}`+"\n", now.Add(-90*24*time.Hour))

	stats := PruneObservationLogs(logsDir, currentSession, 30, now)
	if stats.TraceAgedPruned != 0 {
		t.Errorf("current session trace must not be age-pruned; got %d", stats.TraceAgedPruned)
	}
	if stats.Skipped < 1 {
		t.Errorf("expected the current-session trace to be counted as Skipped, got %d", stats.Skipped)
	}
	if _, err := os.Stat(filepath.Join(logsDir, "trace-"+currentSession+".jsonl")); err != nil {
		t.Errorf("current session trace must survive: %v", err)
	}
}

// EC-2 — absent logs dir → silent no-op, zero stats, no panic.
func TestPruneObservationLogs_AbsentLogsDir(t *testing.T) {
	t.Parallel()
	missingDir := filepath.Join(t.TempDir(), "does", "not", "exist")
	stats := PruneObservationLogs(missingDir, "sess", 30, time.Now())
	// zero stats, no panic
	if stats.TraceZeroBytePruned != 0 || stats.TraceAgedPruned != 0 || stats.TaskMetricsAged != 0 {
		t.Errorf("absent dir should yield zero stats, got %+v", stats)
	}
}

// task-metrics.jsonl older than retention is pruned (documented write-only
// disposition per D2 default — the Agent tool response no longer carries a
// `metrics` field so the writer is dormant).
func TestPruneObservationLogs_TaskMetricsAgedOut(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	// stale task-metrics (45 days old)
	tmPath := filepath.Join(logsDir, "task-metrics.jsonl")
	if err := os.WriteFile(tmPath, []byte(`{"x":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(tmPath, now.Add(-45*24*time.Hour), now.Add(-45*24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	stats := PruneObservationLogs(logsDir, "sess", 30, now)
	if stats.TaskMetricsAged != 1 {
		t.Errorf("TaskMetricsAged = %d, want 1", stats.TaskMetricsAged)
	}
	if _, err := os.Stat(tmPath); err == nil {
		t.Errorf("stale task-metrics.jsonl should be pruned")
	}
}

// task-metrics.jsonl under threshold is kept (writer may resume).
func TestPruneObservationLogs_TaskMetricsFreshKept(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	tmPath := filepath.Join(logsDir, "task-metrics.jsonl")
	if err := os.WriteFile(tmPath, []byte(`{"x":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(tmPath, now.Add(-5*24*time.Hour), now.Add(-5*24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	stats := PruneObservationLogs(logsDir, "sess", 30, now)
	if stats.TaskMetricsAged != 0 {
		t.Errorf("fresh task-metrics should not be pruned, got %d", stats.TaskMetricsAged)
	}
	if _, err := os.Stat(tmPath); err != nil {
		t.Errorf("fresh task-metrics.jsonl should survive: %v", err)
	}
}

// Non-trace files (other logs) are never touched by the pruning.
func TestPruneObservationLogs_NonTraceFilesUntouched(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	for _, name := range []string{"hook-skip.log", "sync-quality-gate.log", "status-transition-audit.log"} {
		p := filepath.Join(logsDir, name)
		if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, now.Add(-90*24*time.Hour), now.Add(-90*24*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	stats := PruneObservationLogs(logsDir, "sess", 30, now)
	if stats.TraceZeroBytePruned != 0 || stats.TraceAgedPruned != 0 || stats.TaskMetricsAged != 0 {
		t.Errorf("non-trace logs must not be pruned, got %+v", stats)
	}
	for _, name := range []string{"hook-skip.log", "sync-quality-gate.log", "status-transition-audit.log"} {
		if _, err := os.Stat(filepath.Join(logsDir, name)); err != nil {
			t.Errorf("non-trace log %s must survive: %v", name, err)
		}
	}
}

// Rotated trace backups (trace-<id>.1.jsonl) are subject to the same policy.
func TestPruneObservationLogs_RotatedTracePruned(t *testing.T) {
	t.Parallel()
	logsDir := t.TempDir()
	now := time.Now()
	// rotated backup, over-threshold
	rotPath := filepath.Join(logsDir, "trace-sess-x.1.jsonl")
	if err := os.WriteFile(rotPath, []byte(`{"old"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(rotPath, now.Add(-60*24*time.Hour), now.Add(-60*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	stats := PruneObservationLogs(logsDir, "current-sess", 30, now)
	if stats.TraceAgedPruned != 1 {
		t.Errorf("rotated over-threshold trace should be pruned, got %d", stats.TraceAgedPruned)
	}
}
