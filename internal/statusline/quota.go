package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-QUOTA-AWARE-SCHEDULING-001 M2 (REQ-QAS-005..007): the quota aggregator.
//
// The Claude 5h / 7d quota is account-level, so any live Claude session's
// reading describes it. Each session's statusline writes its own telemetry
// record (context_usage.go); this file reads those records and answers one
// question per window: what is the account's reading now, and can it be
// trusted? It is a pure reader — no network, no process spawn, no write — and
// every failure maps to "unknown", which never holds a lane (fail open).

// QuotaState is how one rate-limit window reads in the aggregate.
type QuotaState string

const (
	// QuotaFresh: a record within the max age carries the window and its reset
	// time is still ahead.
	QuotaFresh QuotaState = "fresh"
	// QuotaReset: the freshest record's reset time is not after now, so the
	// window has rolled over. It contributes nothing and is never read as a
	// stale high.
	QuotaReset QuotaState = "reset"
	// QuotaUnknown: no usable record carries the window, or the directory or a
	// record could not be read.
	QuotaUnknown QuotaState = "unknown"
)

// QuotaReading is one window's aggregate reading. Only State is meaningful for
// an unknown window; a reset window carries the source record's times (so a
// reader can say when it rolled) but no percentage.
type QuotaReading struct {
	State          QuotaState
	UsedPercentage float64   // zero unless State is QuotaFresh
	ResetsAt       int64     // Unix epoch seconds, as the stdin supplied it
	CapturedAt     time.Time // the source record's capture time
	ExhaustedAt    time.Time // the source's first-observed-exhausted time; zero when none
}

// QuotaAggregate is the account-level reading of both windows.
type QuotaAggregate struct {
	FiveHour QuotaReading
	SevenDay QuotaReading
}

// AggregateQuota reads the session telemetry records under stateDir (the
// project's .moai/state directory) and reports each window per REQ-QAS-005:
//
//   - only record files whose modification time is within maxAge are parsed, so
//     the cost does not grow with the count of dead-session records;
//   - a record is fresh while its capture time is at most maxAge before now (one
//     second older contributes nothing) and no more than
//     config.QuotaClockSkewTolerance after it (a record from a skewed clock or
//     another machine must not win by being "newest");
//   - per window, the freshest such record that carries the window wins by
//     capture time — never the maximum percentage, never the minimum;
//   - a winning window whose reset time is not after now reads as reset.
//
// An absent or unreadable directory, an unreadable record, or no usable record
// yields unknown (REQ-QAS-006).
//
// It is the one-directory case of AggregateQuotaDirs (quota_dirs.go), which
// reads the same rules across several directories.
func AggregateQuota(stateDir string, now time.Time, maxAge time.Duration) QuotaAggregate {
	return AggregateQuotaDirs([]string{stateDir}, now, maxAge)
}

// scanQuotaDir offers every usable record under stateDir (the project's
// .moai/state directory) to the per-window winners, applying the pre-filters of
// AggregateQuota. An absent or unreadable directory, or an unreadable record,
// offers nothing.
func scanQuotaDir(stateDir string, now time.Time, maxAge time.Duration, fiveHour, sevenDay *winner) {
	dir := filepath.Join(stateDir, contextUsageDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) > maxAge {
			continue // unreadable, or not changed within the max age: never parsed
		}
		rec, err := ReadSessionTelemetry(filepath.Join(dir, e.Name()))
		if err != nil || rec == nil {
			continue // an unparseable record is skipped; fail open
		}
		captured, err := time.Parse(time.RFC3339Nano, rec.CapturedAt)
		if err != nil || now.Sub(captured) > maxAge || captured.Sub(now) > config.QuotaClockSkewTolerance {
			continue // stale, or dated beyond the skew tolerance: unknown, never "freshest"
		}
		fiveHour.offer(rec.FiveHour, captured)
		sevenDay.offer(rec.SevenDay, captured)
	}
}

// winner tracks the freshest record carrying one window.
type winner struct {
	win      *QuotaWindowRecord
	captured time.Time
}

// offer replaces the current winner when win is carried and captured is later.
// A tie keeps the earlier offer, so the result is deterministic over the sorted
// directory listing.
func (w *winner) offer(win *QuotaWindowRecord, captured time.Time) {
	if win == nil {
		return
	}
	if w.win == nil || captured.After(w.captured) {
		w.win, w.captured = win, captured
	}
}

// reading converts the winner into the window's aggregate reading at now.
func (w *winner) reading(now time.Time) QuotaReading {
	if w.win == nil {
		return QuotaReading{State: QuotaUnknown}
	}
	r := QuotaReading{
		State:          QuotaFresh,
		UsedPercentage: w.win.UsedPercentage,
		ResetsAt:       w.win.ResetsAt,
		CapturedAt:     w.captured,
	}
	if ex, err := time.Parse(time.RFC3339Nano, w.win.ExhaustedAt); err == nil {
		r.ExhaustedAt = ex
	}
	if !time.Unix(w.win.ResetsAt, 0).After(now) {
		// The window has rolled over: its percentage describes the previous
		// window, so it is dropped rather than left to be read as a stale high.
		r.State, r.UsedPercentage = QuotaReset, 0
	}
	return r
}
