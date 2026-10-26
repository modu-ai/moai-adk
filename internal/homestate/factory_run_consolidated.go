package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// factory_run_consolidated.go — the read side of REQ-FCR-024 (card t1963): a
// dispatch repair or retry aimed at a run that a consolidation retired must be
// refused. The consolidation leaves two marks in one transaction
// (factory_run_rebind.go): the source's status becomes 'retired', and its
// run.retired event carries rebound_to. An ordinary retirement writes the same
// event kind WITHOUT rebound_to, so the pair — not either alone — identifies a
// consolidation.

// ErrRunRetiredByConsolidation marks an operation refused because the run it
// names was retired by a consolidation: its cards live in the surviving run
// now, and recreating work in the retired one would leave a non-terminal row in
// a run that can never be retired and a binding to a run nobody owns.
var ErrRunRetiredByConsolidation = errors.New("factory run was retired by a consolidation")

// RunRetiredByConsolidation reports whether runID was retired by a
// consolidation and, if so, the run its cards went to. A run that is active,
// retired the ordinary way, or absent reads false with a nil error; only a
// failing store is an error. It reads outside any transaction; an operation
// that must not race a consolidation uses the transaction-scoped form through
// RepairDispatchAtomic instead.
func (f *FactoryDB) RunRetiredByConsolidation(ctx context.Context, runID string) (into string, retired bool, err error) {
	return runRetiredByConsolidation(ctx, f.DB, runID)
}

// runRetiredByConsolidation is the marker read on the caller's queryer — the
// pool handle, or a write transaction that already holds the lock a
// consolidation needs.
func runRetiredByConsolidation(ctx context.Context, q rowsRowQueryer, runID string) (into string, retired bool, err error) {
	var status string
	switch err := q.QueryRowContext(ctx, `SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); {
	case errors.Is(err, sql.ErrNoRows):
		return "", false, nil
	case err != nil:
		return "", false, err
	}
	if status != "retired" {
		return "", false, nil
	}
	rows, err := q.QueryContext(ctx, `SELECT payload_json FROM events WHERE run_id=? AND kind='run.retired' ORDER BY seq DESC`, runID)
	if err != nil {
		return "", false, err
	}
	var payloads []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			_ = rows.Close()
			return "", false, err
		}
		payloads = append(payloads, p)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return "", false, err
	}
	if err := rows.Close(); err != nil {
		return "", false, err
	}
	for _, p := range payloads {
		var body struct {
			ReboundTo string `json:"rebound_to"`
		}
		if json.Unmarshal([]byte(p), &body) == nil && body.ReboundTo != "" {
			return body.ReboundTo, true, nil
		}
	}
	return "", false, nil
}

// refuseIfRetiredByConsolidation is the marker read as a guard: an error
// wrapping ErrRunRetiredByConsolidation, naming the run and the run its cards
// went to, when the run was consolidated away.
func refuseIfRetiredByConsolidation(ctx context.Context, q rowsRowQueryer, runID string) error {
	into, retired, err := runRetiredByConsolidation(ctx, q, runID)
	if err != nil {
		return err
	}
	if retired {
		return fmt.Errorf("%w: run %s was retired by a consolidation into %s — its work lives in %s now (if that run was retired in turn, follow its own consolidation record)", ErrRunRetiredByConsolidation, runID, into, into)
	}
	return nil
}
