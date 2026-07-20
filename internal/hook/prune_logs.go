// prune_logs.go — SessionEnd observation-log pruning.
//
// SPEC-OBSERVE-HYGIENE-001 M2 (REQ-OBH-002, AC-OBH-002).
//
// The .moai/logs/ directory accumulates per-session trace-*.jsonl files (written
// by internal/hook/trace/writer.go) with no age-based eviction — only size-based
// rotation. Hundreds of zero-byte traces and months-old non-empty traces pile up.
// This step, invoked from the SessionEnd handler, applies a documented retention
// policy: zero-byte traces are removed unconditionally; non-empty traces older
// than DefaultTraceRetentionDays are removed; the current session's active trace
// is always preserved (EC-3); an absent logs dir is a silent no-op (EC-2).
//
// task-metrics.jsonl is included in the age-out (documented write-only
// disposition per SPEC-OBSERVE-HYGIENE-001 D2): the Agent tool (renamed from
// Task in Claude Code v2.1.63) response no longer carries a `metrics` field, so
// logTaskMetrics returns early at the `resp.Metrics == nil` guard and the file
// is dormant. See the finding note on logTaskMetrics in post_tool_metrics.go.
//
// All operations are best-effort: errors are logged via slog.Warn, never
// returned. The pruning never escapes logsDir (only trace-*.jsonl and
// task-metrics.jsonl are candidates).
package hook

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// PruneStats summarizes a single SessionEnd pruning run for observability.
type PruneStats struct {
	// TraceZeroBytePruned counts zero-byte trace-*.jsonl files removed.
	TraceZeroBytePruned int
	// TraceAgedPruned counts non-empty trace-*.jsonl files removed for exceeding
	// the retention threshold.
	TraceAgedPruned int
	// TaskMetricsAged is 1 when the stale task-metrics.jsonl was removed (0
	// otherwise — the file is either under threshold, absent, or the current
	// write target).
	TaskMetricsAged int
	// Skipped counts candidates that were intentionally preserved (the current
	// session's active trace) or skipped on a per-file error.
	Skipped int
}

// PruneObservationLogs is the SessionEnd pruning step (REQ-OBH-002).
//
// Behavior:
//   - zero-byte trace-*.jsonl → removed unconditionally (stale empty sessions).
//   - non-empty trace-*.jsonl older than retentionDays → removed.
//   - the current session's trace (name contains currentSessionID) → ALWAYS
//     preserved, regardless of age or size (EC-3 — never prune the active trace).
//   - task-metrics.jsonl older than retentionDays → removed (documented
//     write-only disposition; the writer is dormant).
//   - all other files under logsDir → untouched (scope is trace + task-metrics).
//   - logsDir absent or unreadable → silent no-op, zero stats (EC-2).
//
// now is injected so tests are deterministic; production passes time.Now().
func PruneObservationLogs(logsDir, currentSessionID string, retentionDays int, now time.Time) PruneStats {
	var stats PruneStats
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		// EC-2: absent / unreadable logs dir → silent no-op.
		return stats
	}
	if retentionDays < 1 {
		retentionDays = config.DefaultTraceRetentionDays
	}
	cutoff := now.AddDate(0, 0, -retentionDays)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			stats.Skipped++
			continue
		}

		if isTraceFile(name) {
			// EC-3 — never prune the current session's active trace (or its
			// rotation backup trace-<id>.1.jsonl).
			if currentSessionID != "" && strings.Contains(name, currentSessionID) {
				stats.Skipped++
				continue
			}
			path := filepath.Join(logsDir, name)
			if info.Size() == 0 {
				if rmErr := os.Remove(path); rmErr != nil {
					slog.Warn("prune_logs: failed to remove zero-byte trace",
						"path", path, "error", rmErr)
					stats.Skipped++
				} else {
					stats.TraceZeroBytePruned++
				}
				continue
			}
			if info.ModTime().Before(cutoff) {
				if rmErr := os.Remove(path); rmErr != nil {
					slog.Warn("prune_logs: failed to remove aged trace",
						"path", path, "error", rmErr)
					stats.Skipped++
				} else {
					stats.TraceAgedPruned++
				}
			}
			continue
		}

		if name == "task-metrics.jsonl" {
			path := filepath.Join(logsDir, name)
			if info.ModTime().Before(cutoff) {
				if rmErr := os.Remove(path); rmErr != nil {
					slog.Warn("prune_logs: failed to remove aged task-metrics.jsonl",
						"path", path, "error", rmErr)
					stats.Skipped++
				} else {
					stats.TaskMetricsAged++
				}
			}
		}
	}
	if stats.TraceZeroBytePruned+stats.TraceAgedPruned+stats.TaskMetricsAged > 0 {
		slog.Info("prune_logs: SessionEnd observation-log pruning complete",
			"trace_zero_byte_pruned", stats.TraceZeroBytePruned,
			"trace_aged_pruned", stats.TraceAgedPruned,
			"task_metrics_aged", stats.TaskMetricsAged,
			"skipped", stats.Skipped,
		)
	}
	return stats
}

// isTraceFile reports whether name matches the per-session trace file pattern
// trace-*.jsonl (including rotated backups trace-*.1.jsonl). It does not open
// the file — a name-shape check only.
func isTraceFile(name string) bool {
	return strings.HasPrefix(name, "trace-") && strings.HasSuffix(name, ".jsonl")
}
