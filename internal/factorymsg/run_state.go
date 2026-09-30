package factorymsg

// run_state.go — the shared run-state measurement behind the hook
// prescription gate (SPEC-STALE-RUN-LABEL-001 REQ-SRL-003).
//
// ValidateActiveRun's error channel conflates two things the gate must keep
// apart: a MEASURED not-active verdict (the runs row says retired, the run
// has no row, or this project never had a factory DB) and a measurement
// FAILURE (busy database, spent budget, I/O error). The gate acts on the
// first and fails open on the second, so it reads through this tri-state
// accessor instead. The same store the CLI retire path writes
// (homestate.FactoryDBPath, runs.status) is the only measurement source —
// broker-file absence is never consulted here.

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// RunState is the tri-state verdict of the shared run-state measurement.
//
// @MX:NOTE: [AUTO] The prescription gate (internal/hook staleRunPrescriptionGate)
// acts on Active/NotActive and fails open on Unavailable — the three states
// are load-bearing, do not collapse them back into an error channel.
type RunState uint8

const (
	// RunStateActive — the runs row measures status='active'.
	RunStateActive RunState = iota
	// RunStateNotActive — a measured verdict: the run's status is not
	// active (retired or otherwise), the run has no row, or this project
	// has no factory DB at all (never a factory project).
	RunStateNotActive
	// RunStateUnavailable — the measurement did not complete (busy DB,
	// spent budget, I/O failure). Callers fail open: degraded answer,
	// never a prescription.
	RunStateUnavailable
)

// String renders the state for logs and notices.
func (s RunState) String() string {
	switch s {
	case RunStateActive:
		return "active"
	case RunStateNotActive:
		return "not-active"
	default:
		return "unavailable"
	}
}

// ProbeRunState measures the named run's status through the shared store.
// detail carries the measured runs.status value verbatim, "absent" when the
// run has no row (or the id is malformed / the DB is absent), and "" when
// the state is unavailable.
//
// @MX:NOTE: [AUTO] Shared accessor per SPEC-STALE-RUN-LABEL-001 REQ-SRL-003 —
// every prescription surface must measure through here, never re-derive run
// state from broker-file absence or text patterns.
func ProbeRunState(ctx context.Context, projectRoot, runID string) (state RunState, detail string, err error) {
	if !safeID.MatchString(runID) {
		// A malformed id cannot name an active run — a measured property of
		// the label, not a measurement failure.
		return RunStateNotActive, "absent", nil
	}
	path, err := homestate.FactoryDBPath(projectRoot)
	if err != nil {
		return RunStateUnavailable, "", err
	}
	if _, err := os.Stat(path); err != nil {
		// The accessor's own DB-file probe is part of the shared measurement
		// (REQ-SRL-003): no factory DB means this project never carried the
		// run — measured not-active, never a failure.
		return RunStateNotActive, "absent", nil
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(100)")
	if err != nil {
		return RunStateUnavailable, "", err
	}
	defer func() { _ = db.Close() }() // read-only measurement; a close failure cannot change the verdict
	var status string
	scanErr := db.QueryRowContext(ctx, `SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status)
	switch {
	case errors.Is(scanErr, sql.ErrNoRows):
		return RunStateNotActive, "absent", nil
	case scanErr != nil:
		// Busy database, spent budget, schema damage — the measurement did
		// not complete, which is not a verdict about the run.
		return RunStateUnavailable, "", scanErr
	case status != "active":
		return RunStateNotActive, status, nil
	}
	return RunStateActive, status, nil
}

// ActiveRunExists reports whether any run in the project root measures
// status='active' — the re-bind-line condition of the unbind notice
// (REQ-SRL-006). A project without a factory DB measures false; a failed
// measurement returns the error and the caller omits the line (fail-open).
//
// @MX:NOTE: [AUTO] Read-side sibling of ProbeRunState for the unbind notice's
// re-bind line; same shared-store rule (never the broker files).
func ActiveRunExists(ctx context.Context, projectRoot string) (active bool, err error) {
	path, err := homestate.FactoryDBPath(projectRoot)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(path); err != nil {
		return false, nil
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(100)")
	if err != nil {
		return false, err
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM runs WHERE status='active')`).Scan(&n); err != nil {
		return false, err
	}
	return n == 1, nil
}
