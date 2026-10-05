package hygiene

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// lockResult is the outcome of a cross-process pass-lock attempt.
type lockResult int

const (
	lockAcquired lockResult = iota
	lockHeld
	lockUnverifiable
)

// Rotator holds audit-log sinks to a size bound with a keep-1 chunk cap
// (REQ-HYG-001) through a staged, crash-recoverable displacement executed
// under a cross-process pass lock (REQ-HYG-002). The unit fails
// independently of the GC (REQ-HYG-012).
type Rotator struct {
	// LogDir is the absolute path to <projectRoot>/.moai/logs.
	LogDir string
	// MaxBytes is the rotation threshold (10 MiB in production, from
	// internal/config defaults — never a call-site literal).
	MaxBytes int64
	// KeptRotations is pinned to 1 (D30); the defensive check here backs
	// the config-layer validation.
	KeptRotations int

	// now is overridable for tests.
	now func() time.Time
	// statFn, when non-nil, replaces os.Stat — the stale-decision seam
	// (AC-HYG-004 arm a). The first call models the stale pre-lock
	// observation; the under-lock re-stat is a later call.
	statFn func(string) (fs.FileInfo, error)
	// lockProbe, when non-nil, replaces the real pass lock — the sidecar
	// seam modeling held / unverifiable exclusion (AC-HYG-004 arm b).
	lockProbe func(string) lockResult
}

// pnow returns the rotator's clock.
func (r *Rotator) pnow() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

// statPath stats through the seam when installed.
func (r *Rotator) statPath(path string) (fs.FileInfo, error) {
	if r.statFn != nil {
		return r.statFn(path)
	}
	return os.Stat(path)
}

// Run executes one rotator pass in the given mode. Report mode rotates
// nothing, creates no lockfile, and appends exactly one summary row;
// apply mode executes the staged rotation under the pass lock and appends
// one row per action (REQ-HYG-002, REQ-HYG-004, REQ-HYG-010).
//
// @MX:ANCHOR: [AUTO] Rotator.Run — the apply/report mode gate for every rotation
// @MX:REASON: every mutation of a registered sink passes through here; the
// report branch must stay non-mutating and lockfile-free or the shipped
// default would grow sinks and write lock files on session start (REQ-HYG-010).
func (r *Rotator) Run(mode Mode) ([]AuditRow, error) {
	if err := validateRoot(r.LogDir); err != nil {
		return nil, err
	}
	if r.MaxBytes <= 0 {
		return nil, fmt.Errorf("rotator: non-positive threshold %d (config-invalid)", r.MaxBytes)
	}
	if r.KeptRotations != 1 {
		return nil, fmt.Errorf("rotator: kept_rotations pinned to 1, got %d (config-invalid)", r.KeptRotations)
	}

	if mode == ModeReport {
		return r.runReport()
	}
	return r.runApply()
}

// runReport counts decisions and appends one summary row (REQ-HYG-004,
// REQ-HYG-010). The lockfile is never created on this path.
func (r *Rotator) runReport() ([]AuditRow, error) {
	counts := map[string]int{}
	for _, name := range SinkRegistry() {
		path := filepath.Join(r.LogDir, name)
		info, err := r.statPath(path)
		switch {
		case err != nil && !errors.Is(err, fs.ErrNotExist):
			counts[string(OutcomeError)]++
		case err != nil:
			counts[string(OutcomeSkippedAbsent)]++
		case info.Size() >= r.MaxBytes:
			// The hygiene sink's own over-threshold skip is recorded with
			// the same outcome as any other sink (REQ-HYG-004 — no
			// exemption).
			counts[string(OutcomeSkippedReportMode)]++
		default:
			counts[string(OutcomeSkippedUnderThresh)]++
		}
	}
	row := AuditRow{
		TS:      r.pnow().UTC().Format(time.RFC3339),
		Unit:    "rotator",
		Mode:    string(ModeReport),
		Outcome: OutcomeSummary,
		Counts:  counts,
	}
	if err := appendAuditRows(r.LogDir, []AuditRow{row}); err != nil {
		return nil, err
	}
	return []AuditRow{row}, nil
}

// runApply executes the mutating pass under the cross-process lock.
func (r *Rotator) runApply() ([]AuditRow, error) {
	var rows []AuditRow

	release, res := r.acquirePassLock()
	switch res {
	case lockHeld:
		for _, name := range SinkRegistry() {
			rows = append(rows, r.skipRow(name, OutcomeSkippedLocked,
				"cross-process exclusion held by another rotator"))
		}
		if err := appendAuditRows(r.LogDir, rows); err != nil {
			return rows, err
		}
		return rows, nil
	case lockUnverifiable:
		for _, name := range SinkRegistry() {
			rows = append(rows, r.skipRow(name, OutcomeSkippedPlatform,
				"cross-process exclusion cannot be established on this platform"))
		}
		if err := appendAuditRows(r.LogDir, rows); err != nil {
			return rows, err
		}
		return rows, nil
	}
	defer release()

	// Under the lock: recover orphan stagings, then decide per sink from a
	// fresh re-stat — never from the pre-lock observation (REQ-HYG-002).
	for _, name := range SinkRegistry() {
		primary := filepath.Join(r.LogDir, name)
		pre, preErr := r.statPath(primary)
		outcome, path, reason := r.rotateSink(primary, pre, preErr)
		rows = append(rows, AuditRow{
			TS:      r.pnow().UTC().Format(time.RFC3339),
			Unit:    "rotator",
			Mode:    string(ModeApply),
			Outcome: outcome,
			Path:    path,
			Reason:  reason,
		})
	}
	if err := appendAuditRows(r.LogDir, rows); err != nil {
		return rows, err
	}
	return rows, nil
}

// skipRow builds a skip row for one sink.
func (r *Rotator) skipRow(name string, outcome Outcome, reason string) AuditRow {
	return AuditRow{
		TS:      r.pnow().UTC().Format(time.RFC3339),
		Unit:    "rotator",
		Mode:    string(ModeApply),
		Outcome: outcome,
		Path:    filepath.Join(r.LogDir, name),
		Reason:  reason,
	}
}

// rotateSink runs the staged sequence for one sink primary and returns the
// outcome, the affected path, and a reason for skips. The decision is made
// from the under-lock re-stat; a pre-lock observation that would have
// rotated but no longer holds records skipped-stale (REQ-HYG-002).
func (r *Rotator) rotateSink(primary string, pre fs.FileInfo, preErr error) (Outcome, string, string) {
	chunk, staging := chunkNames(primary)

	// Pass-start recovery: an orphan staging is the displaced primary of a
	// crashed pass — complete the interrupted placement; a staged chunk is
	// never deleted and never overwritten (D16).
	if fileExists(staging) {
		if fileExists(chunk) {
			if err := os.Remove(chunk); err != nil {
				return OutcomeError, chunk, fmt.Sprintf("staging recovery: remove stale chunk: %v", err)
			}
		}
		if err := os.Rename(staging, chunk); err != nil {
			return OutcomeError, staging, fmt.Sprintf("staging recovery: promote: %v", err)
		}
		if !fileExists(primary) {
			if err := recreateEmptyPrimary(primary); err != nil {
				return OutcomeError, primary, fmt.Sprintf("staging recovery: recreate primary: %v", err)
			}
		}
	}

	info, err := r.statPath(primary)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return OutcomeSkippedAbsent, primary, "sink absent"
		}
		return OutcomeError, primary, fmt.Sprintf("re-stat: %v", err)
	}
	if !info.Mode().IsRegular() {
		return OutcomeSkippedAbsent, primary, "sink is not a regular file"
	}
	if info.Size() < r.MaxBytes {
		if preErr == nil && pre != nil && pre.Size() >= r.MaxBytes {
			return OutcomeSkippedStale, primary,
				"pre-lock observation was over threshold; the under-lock re-stat is not"
		}
		return OutcomeSkippedUnderThresh, primary, "sink below threshold"
	}

	// Staged displacement: rename primary to staging, remove the existing
	// chunk only while a staged replacement exists, promote staging, then
	// recreate the empty primary (D4+D16+D33).
	if err := os.Rename(primary, staging); err != nil {
		return OutcomeError, primary, fmt.Sprintf("stage primary: %v", err)
	}
	if fileExists(chunk) {
		if err := os.Remove(chunk); err != nil {
			return OutcomeError, chunk, fmt.Sprintf("remove displaced chunk: %v", err)
		}
	}
	if err := os.Rename(staging, chunk); err != nil {
		return OutcomeError, staging, fmt.Sprintf("promote staging: %v", err)
	}
	if err := recreateEmptyPrimary(primary); err != nil {
		return OutcomeError, primary, fmt.Sprintf("recreate primary: %v", err)
	}
	return OutcomeRotated, primary, ""
}

// recreateEmptyPrimary recreates a rotated sink's primary as an empty file
// so the primary path exists continuously for its readers (D33).
func recreateEmptyPrimary(primary string) error {
	f, err := os.OpenFile(primary, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}
