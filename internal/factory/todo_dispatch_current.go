package factory

import (
	"context"
	"errors"
	"strings"
)

// RefreshDispatchCurrent keeps the queue's current-dispatch record in step
// with a binding move, while the caller holds the queue lock (review round-24
// P1-3 / relay #2-#3, card t1538): every path that moves a card's factory
// binding onto a run — direct assign, lease, audit approval, selection —
// refreshes the record the dispatch hook wrote (the dispatch mirror needs no
// refresh of its own: the hook already named its run), so a retried older
// operation reads as superseded instead of dragging the binding back. A caller
// whose binding move is a transition passes this to TransitionRequest's
// BeforeCommit, so the record is written only for a transition that lands.
//
// It only REFRESHES: a card the dispatch hook never recorded has nothing to
// keep in step (no operation can be in flight for it), and a queue that never
// carried the record is left exactly as it was — no schema, no migration, no
// file — because the factory record's verbs leave the queue alone (AC-FR-021).
func (l *LockedBacklog) RefreshDispatchCurrent(cardID, runID, owner string) error {
	return l.s.RefreshDispatchCurrentLockHeld(cardID, runID, owner)
}

// RefreshDispatchCurrentLockHeld is RefreshDispatchCurrent for a caller that
// holds the queue lock without a LockedBacklog handle (the lease claim path,
// which runs inside the lease section on this same store). The caller MUST
// hold the queue lock; the write takes none of its own.
func (s *BacklogStore) RefreshDispatchCurrentLockHeld(cardID, runID, owner string) error {
	if strings.TrimSpace(cardID) == "" || strings.TrimSpace(runID) == "" {
		return errors.New("dispatch current: card id and run id are required")
	}
	// Look before opening a writer (turn-end gate, card t1538): LoadPure serves
	// whichever layout it finds and adopts nothing, whereas the writable open
	// would migrate a legacy JSON queue into SQLite, create an absent database,
	// and run the schema DDL — all just to learn there is no row to refresh.
	record, err := s.LoadPure()
	if err != nil {
		return err
	}
	recorded := false
	for _, cur := range record.Runtime.DispatchCurrent {
		if cur.CardID == cardID {
			recorded = true
			break
		}
	}
	if !recorded {
		return nil
	}
	eng, err := openBacklogEngine(backlogSQLitePath(s.path))
	if err != nil {
		return err
	}
	defer func() { _ = eng.close() }()
	ctx, cancel := context.WithTimeout(context.Background(), backlogOpTimeout)
	defer cancel()
	// owner "" claims no owner (a selection that re-points the binding knows
	// only the run): the record keeps the owner it already holds for the SAME
	// run, and carries none onto a different one. SQLite evaluates every SET
	// expression against the pre-update row, so run_id in the CASE is the
	// record's current run.
	_, err = eng.db.ExecContext(ctx, `UPDATE todo_dispatch_current SET
 owner_label=CASE WHEN ?='' AND run_id=? THEN owner_label ELSE ? END,
 run_id=?
WHERE card_id=?`, canonicalOwnerLabel(owner), runID, canonicalOwnerLabel(owner), runID, cardID)
	return err
}
