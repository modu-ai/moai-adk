package homestate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// factory_run_lease_rebound.go — the lease-path half of the run lifecycle
// repair (SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-019, card t1963).
//
// A lease records the card in the CURRENT run and used to leave the same
// card's non-terminal row in an earlier dead run untouched: one card, two live
// rows, and a dead run that could never be retired. RecordPickedReconciled is
// RecordPicked plus the cleanup of that earlier row, in the same transaction.

// leaseReboundBeforeTx is a test-only seam: nil in production, it runs after
// the other runs were classified and before the transaction opens, so a test
// can change an owner stamp inside that window.
var leaseReboundBeforeTx func()

// RecordPickedReconciled is T1 (RecordPicked) that also moves the card's
// non-terminal row in another run to abandoned when that row is the residue of
// a dead run: the other run's owner classified OwnerDead, the row holds no
// valid lease, and the run has no worker that is not positively dead. Each
// moved row gets one card.rebound event (from_run, to_run, the prior state,
// version, stage and evidence_sha). A row failing any condition is left exactly
// as it was — the lease then behaves as it always has (a new row, the other
// run's row untouched).
//
// The abandon runs inside the T1 transaction and after it, so a failure to
// abandon rolls the T1 record back too. The owner classification is read before
// the transaction; inside it only the classified stamp is re-checked. The
// caller reads the re-lease carry-over BEFORE calling, because abandoning
// changes the row that read looks at.
//
// afterRecord (nil = none) runs between the T1 record and the abandon; a
// non-nil return rolls the transaction back (the test seam for the one-
// transaction guarantee).
//
// @MX:WARN: [AUTO] abandons another run's card row inside the lease transaction; every skip condition fails toward leaving the row alone
// @MX:REASON: a row abandoned under a live lane splits the card record and fails that lane's next transition (REQ-FCR-019)
func (f *FactoryDB) RecordPickedReconciled(ctx context.Context, runID, cardID string, fields CardFields, actor string, now time.Time, opts ReconcileOptions, afterRecord func() error) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	// SQL: sqlTerminalStates is a compile-time constant; every value goes through a placeholder.
	var rival int
	if err := f.DB.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE card_id=? AND run_id!=? AND state NOT IN `+sqlTerminalStates, cardID, runID).Scan(&rival); err != nil {
		return Card{}, err
	}
	if rival == 0 {
		// No other run holds a live row for the card: nothing to reconcile and
		// no reason to classify any owner.
		return f.RecordPicked(ctx, runID, cardID, fields, actor, now)
	}
	if err := fields.validate(); err != nil {
		return Card{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	owners, err := f.classifyRuns(ctx, opts, "")
	if err != nil {
		return Card{}, err
	}
	byRun := make(map[string]RunOwner, len(owners))
	for _, o := range owners {
		byRun[o.RunID] = o
	}
	if leaseReboundBeforeTx != nil {
		leaseReboundBeforeTx()
	}
	// A run proved dead through the lead peer or the boot proof has no row-level
	// stamp for the transaction below to compare, so those owners are classified
	// again right before BEGIN and a run that is no longer dead keeps its row
	// (REQ-FCR-019). The window is narrowed, not closed: the peer identity lives
	// in the broker database, outside the transaction.
	if hasUnstampedDeadOwner(byRun) {
		fresh, err := f.classifyRuns(ctx, opts, "")
		if err != nil {
			return Card{}, err
		}
		for _, o := range fresh {
			if prev, ok := byRun[o.RunID]; ok && unstampedDead(prev) {
				byRun[o.RunID] = o
			}
		}
	}
	probe := opts.workerProbe()
	var result Card
	err = f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		c, err := recordPickedTx(ctx, tx, runID, cardID, fields, actor, now)
		if err != nil {
			return nil, err
		}
		result = c
		if afterRecord != nil {
			if err := afterRecord(); err != nil {
				return nil, err
			}
		}
		rows, err := rivalRows(ctx, tx, cardID, runID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !deadRunResidue(ctx, tx, row, byRun[row.RunID], now, probe) {
				continue
			}
			if err := abandonWithPointer(ctx, tx, row.RunID, runID, row, "lease", now); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// unstampedDead reports a run classified dead on a proof that carries no
// row-level stamp: the lead peer or the boot proof.
func unstampedDead(o RunOwner) bool {
	return retirable(o.Classification) && (o.Basis == BasisPeer || o.Basis == BasisBoot)
}

func hasUnstampedDeadOwner(byRun map[string]RunOwner) bool {
	for _, o := range byRun {
		if unstampedDead(o) {
			return true
		}
	}
	return false
}

// rivalRows reads the card's non-terminal rows of every run except runID.
func rivalRows(ctx context.Context, tx *sql.Tx, cardID, runID string) ([]Card, error) {
	// SQL: the concatenated fragments are compile-time constants; every value goes through a placeholder.
	rows, err := tx.QueryContext(ctx, `SELECT `+cardSelectColumns+` FROM cards WHERE card_id=? AND run_id!=? AND state NOT IN `+sqlTerminalStates+` ORDER BY run_id`, cardID, runID)
	if err != nil {
		return nil, err
	}
	var out []Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	return out, rows.Close()
}

// deadRunResidue reports whether a non-terminal row of another run is the
// residue of a dead run and may be abandoned: the run's owner classified dead
// and still carries the stamp it was classified on, the row holds no valid
// lease (the lane that leased it may outlive its leader), and the run has no
// worker that is not positively dead. Any doubt answers false — the row stays.
func deadRunResidue(ctx context.Context, tx *sql.Tx, row Card, owner RunOwner, now time.Time, probe func(int) ProcessIdentityState) bool {
	if owner.RunID != row.RunID || !retirable(owner.Classification) {
		return false
	}
	var pid int
	var start string
	if err := tx.QueryRowContext(ctx, `SELECT lead_pid, lead_process_start FROM runs WHERE run_id=?`, row.RunID).Scan(&pid, &start); err != nil {
		return false
	}
	if pid != owner.LeadPID || start != owner.LeadProcessStart {
		return false
	}
	if holdsValidLease(row, now) {
		return false
	}
	workers, err := workerPIDs(ctx, tx, row.RunID)
	if err != nil {
		return false
	}
	for _, w := range workers {
		if probe(w.pid) != ProcessIdentityDead {
			return false
		}
	}
	return true
}

// holdsValidLease reports whether the card holds a lease that has not expired.
// A lease-holding state that names a holder but carries no expiry counts as
// valid (the same open-ended reading Card.LeaseExpired gives an empty expiry),
// and an unparseable expiry counts as expired.
func holdsValidLease(c Card, now time.Time) bool {
	if !IsLeaseHoldingState(c.State) {
		return false
	}
	if c.LeaseHolder == "" && c.LeaseExpiresAt == "" {
		return false
	}
	return !c.LeaseExpired(now)
}
