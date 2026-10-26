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
// failing store is an error.
func (f *FactoryDB) RunRetiredByConsolidation(ctx context.Context, runID string) (into string, retired bool, err error) {
	var status string
	switch err := f.DB.QueryRowContext(ctx, `SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); {
	case errors.Is(err, sql.ErrNoRows):
		return "", false, nil
	case err != nil:
		return "", false, err
	}
	if status != "retired" {
		return "", false, nil
	}
	rows, err := f.DB.QueryContext(ctx, `SELECT payload_json FROM events WHERE run_id=? AND kind='run.retired' ORDER BY seq DESC`, runID)
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

// RefuseIfRetiredByConsolidation is RunRetiredByConsolidation as a guard: it
// returns an error wrapping ErrRunRetiredByConsolidation, naming the run and
// the run its cards went to, when the run was consolidated away.
func (f *FactoryDB) RefuseIfRetiredByConsolidation(ctx context.Context, runID string) error {
	into, retired, err := f.RunRetiredByConsolidation(ctx, runID)
	if err != nil {
		return err
	}
	if retired {
		return fmt.Errorf("%w: run %s was consolidated into %s — re-dispatch the card against %s", ErrRunRetiredByConsolidation, runID, into, into)
	}
	return nil
}
