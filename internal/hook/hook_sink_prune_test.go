// hook_sink_prune_test.go — guard 4 of SPEC-HOOK-DIAG-SINK-001 (AC-HDS-010).
//
// The hook path's log sink (.moai/logs/hook-runtime.log, written by
// internal/cli/hook_sink.go) is an append-only file with no age-out of its own.
// M4 folds it into the retention machinery that already prunes the other
// observation logs at SessionEnd, rather than inventing a second mechanism.
//
// Both arms matter. A guard that only proves removal cannot separate "prunes
// correctly" from "prunes everything", so the fresh-file arm is not decoration:
// it is what makes the aged-file arm mean anything.
//
// All tests operate exclusively under t.TempDir() — the LIVE .moai/logs/ dir is
// never touched by test code.
package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeSinkFile creates a hook-runtime.log under logsDir with the given content
// and mod time. Returns the full path.
func writeSinkFile(t *testing.T, logsDir, content string, modTime time.Time) string {
	t.Helper()
	path := filepath.Join(logsDir, hookRuntimeLogFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	return path
}

// AC-HDS-010 — the sink file is a SessionEnd pruning candidate: an aged sink is
// removed and reported through PruneStats, while one inside the retention
// threshold is kept.
func TestHookSinkIsPrunedAtSessionEnd(t *testing.T) {
	t.Parallel()

	t.Run("aged sink is removed and reported", func(t *testing.T) {
		t.Parallel()
		logsDir := t.TempDir()
		now := time.Now()
		path := writeSinkFile(t, logsDir, `{"level":"WARN","msg":"x"}`+"\n",
			now.Add(-45*24*time.Hour))

		stats := PruneObservationLogs(logsDir, "current-session", 30, now)

		if stats.HookRuntimeLogAged != 1 {
			t.Errorf("HookRuntimeLogAged = %d, want 1", stats.HookRuntimeLogAged)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("aged sink still present: stat err = %v, want IsNotExist", err)
		}
	})

	t.Run("sink inside the threshold is kept", func(t *testing.T) {
		t.Parallel()
		logsDir := t.TempDir()
		now := time.Now()
		path := writeSinkFile(t, logsDir, `{"level":"WARN","msg":"y"}`+"\n",
			now.Add(-5*24*time.Hour))

		stats := PruneObservationLogs(logsDir, "current-session", 30, now)

		if stats.HookRuntimeLogAged != 0 {
			t.Errorf("HookRuntimeLogAged = %d, want 0 (file is inside the threshold)",
				stats.HookRuntimeLogAged)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("sink inside the threshold was removed: stat err = %v", err)
		}
	})
}
