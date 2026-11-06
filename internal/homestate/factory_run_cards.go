package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// factory_run_cards.go — the read side of the operator retirement guard
// (SPEC-FACTORY-RETIRE-GUARD-001, card t1852): the non-terminal card listing,
// and the typed error a retirement returns when the run still holds cards.

// InProgressCard is one non-terminal card of a run: a card whose state lies
// outside the shared terminalStates set (done, failed, abandoned).
type InProgressCard struct {
	CardID     string
	State      string
	OwnerLabel string
}

// ErrRunHasInProgressCards reports a refused operator retirement: the run's
// owner classified dead, but the run still holds non-terminal cards, and
// retiring it would strand them inside a retired run. It is distinct from
// ErrRunOwnerNotDead; the two refusals mean different things.
var ErrRunHasInProgressCards = errors.New("factory run holds in-progress cards")

// RunInProgressCardsError is the typed form of ErrRunHasInProgressCards. It
// carries the run id and every non-terminal card so the caller can name them.
// errors.Is(err, ErrRunHasInProgressCards) and errors.As both match it.
type RunInProgressCardsError struct {
	RunID string
	Cards []InProgressCard
}

func (e *RunInProgressCardsError) Error() string {
	parts := make([]string, len(e.Cards))
	for i, c := range e.Cards {
		owner := c.OwnerLabel
		if owner == "" {
			owner = "-"
		}
		parts[i] = fmt.Sprintf("%s (%s, %s)", c.CardID, c.State, owner)
	}
	return fmt.Sprintf("%v: run %s holds %d: %s", ErrRunHasInProgressCards, e.RunID, len(e.Cards), strings.Join(parts, ", "))
}

// Is lets errors.Is match the sentinel through the typed value.
func (e *RunInProgressCardsError) Is(target error) bool { return target == ErrRunHasInProgressCards }

// rowsQueryer is the multi-row sibling of queryRower; both *sql.DB and *sql.Tx
// satisfy it, so the listing can run on a caller's transaction.
type rowsQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// retireAfterCardRead is a test-only seam: nil in production, it is called as
// the last statement of inProgressCards, after the rows are closed. A test uses
// it to attempt a competing write at the exact moment the card read finished,
// so a refactor that moves the read out of the write-locked transaction moves
// the seam with it and the test then sees the competing write commit.
//
// @MX:NOTE: [AUTO] test-only seam; production code never sets it and the nil check is the only production cost
// @MX:REASON: AC-RRG-002 needs a probe that cannot be left behind when the card read is moved out of the retiring transaction
var retireAfterCardRead func()

// inProgressCards lists the run's non-terminal cards ordered by card id. The
// single-connection pool deadlocks when a transaction is open and the pool
// handle is used, so a caller inside a transaction passes the transaction.
func inProgressCards(ctx context.Context, q rowsQueryer, runID string) (_ []InProgressCard, err error) {
	// SQL: sqlTerminalStates is a compile-time constant; the run id goes through a placeholder.
	rows, err := q.QueryContext(ctx, `SELECT card_id,state,owner_label FROM cards WHERE run_id=? AND state NOT IN `+sqlTerminalStates+` ORDER BY card_id`, runID)
	if err != nil {
		return nil, err
	}
	var cards []InProgressCard
	for rows.Next() {
		var c InProgressCard
		if err = rows.Scan(&c.CardID, &c.State, &c.OwnerLabel); err != nil {
			_ = rows.Close()
			return nil, err
		}
		cards = append(cards, c)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if retireAfterCardRead != nil {
		retireAfterCardRead()
	}
	return cards, nil
}

// InProgressCards lists the run's non-terminal cards through the pool handle,
// for read-only callers outside any transaction.
func (f *FactoryDB) InProgressCards(ctx context.Context, runID string) ([]InProgressCard, error) {
	return inProgressCards(ctx, f.DB, runID)
}
