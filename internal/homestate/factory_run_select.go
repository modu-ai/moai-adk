package homestate

import (
	"context"
	"time"
)

// factory_run_select.go — the read path behind the lane-join run chooser
// (SPEC-LANE-JOIN-RUN-SELECT-001 REQ-LJR-001): one fact set per active run,
// gathered without writing a runs or events row and without running the
// reconciler.

// RunFacts is one active run's owner classification together with the facts a
// lane-join chooser shows beside it.
type RunFacts struct {
	RunOwner
	// LastActivity is the stored string of the latest lane activity time
	// (worker heartbeat, card heartbeat, card update), empty when none parses.
	LastActivity string
	// InProgress counts the run's cards outside the shared terminal set.
	InProgress int
}

// laneActivityQuery lists every time a lane leaves on a run's records.
const laneActivityQuery = `
SELECT heartbeat_at FROM workers WHERE run_id=?1
UNION ALL SELECT heartbeat_at FROM cards WHERE run_id=?1
UNION ALL SELECT updated_at FROM cards WHERE run_id=?1`

// ActiveRunFacts reports every active run with its classification, its count
// of non-terminal cards and its last lane activity. It is read-only: it calls
// classifyRuns (SELECTs, identity probes, boot-proof queries) and never the
// reconciler, so a dead zero-card run is listed rather than retired.
//
// Every statement is drained before the next starts — the pool holds a single
// connection, the hazard classifyRuns documents.
func (f *FactoryDB) ActiveRunFacts(ctx context.Context, opts ReconcileOptions) ([]RunFacts, error) {
	owners, err := f.classifyRuns(ctx, opts, "active")
	if err != nil {
		return nil, err
	}
	facts := make([]RunFacts, 0, len(owners))
	for _, o := range owners {
		rf := RunFacts{RunOwner: o}
		if err := f.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM cards WHERE run_id=? AND state NOT IN `+sqlTerminalStates, o.RunID).Scan(&rf.InProgress); err != nil {
			return nil, err
		}
		if rf.LastActivity, err = f.lastLaneActivity(ctx, o.RunID); err != nil {
			return nil, err
		}
		facts = append(facts, rf)
	}
	return facts, nil
}

// lastLaneActivity returns the original stored string of the latest lane time
// by PARSED value. Stored times are RFC 3339 with a variable-length fraction
// (trailing zeros are trimmed), so neither SQL MAX nor text order is safe;
// empty and unparseable values are skipped.
func (f *FactoryDB) lastLaneActivity(ctx context.Context, runID string) (_ string, err error) {
	rows, err := f.DB.QueryContext(ctx, laneActivityQuery, runID)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var (
		best    time.Time
		bestRaw string
	)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return "", err
		}
		at, perr := time.Parse(time.RFC3339Nano, raw)
		if perr != nil {
			continue
		}
		if bestRaw == "" || at.After(best) {
			best, bestRaw = at, raw
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return bestRaw, nil
}
