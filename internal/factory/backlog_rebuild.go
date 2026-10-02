// backlog_rebuild.go — the v1→v2 items-table rebuild
// (SPEC-TODO-HOLD-STATE-001, REQ-THS-002/003/004).
//
// SQLite cannot ALTER a CHECK constraint, so widening the live state enum
// from three values to four is a table rebuild: create items_new with the
// current column list, copy every row, verify full-field parity, swap, and
// stamp schema_version — ALL inside one transaction. The safety contract is
// the JSON→SQLite migration discipline inherited from SPEC-TODO-SQLITE-001,
// carried as requirements rather than prose:
//
//   - full-field parity (row count plus every column tuple) is verified
//     BEFORE the switch (REQ-THS-002);
//   - any failure aborts the transaction, leaving the original table — and
//     the original database file — untouched: never delete, never overwrite
//     (REQ-THS-003);
//   - the stamp lands inside the committing transaction, so a binary that
//     knows only the previous stamp refuses the database at open instead of
//     misreading it (REQ-THS-004).
//
// The rebuild touches `items` ONLY: archived_items carries no state CHECK by
// design, and the landing column is already present on any database that
// reaches here (ensureLandingColumn runs before the version switch).
package factory

import (
	"context"
	"database/sql"
	"fmt"
)

// backlogItemRow is one items-table row at the rebuild's verification seam —
// the full physical column tuple, compared field-for-field.
type backlogItemRow struct {
	Seq     int
	ID      string
	Text    string
	AddedAt string
	SpecID  sql.NullString
	Landing sql.NullString
	State   string
}

// readItemsRows loads every row of the named table in stored order. The
// table name is interpolated but is a compile-time constant at both call
// sites ("items" / "items_new") — never a runtime value.
func readItemsRows(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, table string) ([]backlogItemRow, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT seq, id, text, added_at, spec_id, landing, state FROM `+table+` ORDER BY seq`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []backlogItemRow
	for rows.Next() {
		var r backlogItemRow
		if err := rows.Scan(&r.Seq, &r.ID, &r.Text, &r.AddedAt, &r.SpecID, &r.Landing, &r.State); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// @MX:NOTE: [AUTO] parity is deliberately exhaustive (every column tuple), not count-based — a count check would pass a rebuild that silently drops a spec id or landing evidence
// @MX:SPEC: SPEC-TODO-HOLD-STATE-001
// assertRebuildParity compares the copied rows against the source rows —
// deliberately exhaustive rather than count-based, for the same reason the
// JSON→SQLite parity is: a rebuild that moved the right NUMBER of rows while
// dropping a spec id or landing evidence would pass a count check and lose
// operator data silently (REQ-THS-002).
func assertRebuildParity(source, copied []backlogItemRow) error {
	if len(source) != len(copied) {
		return fmt.Errorf("row count %d != %d", len(source), len(copied))
	}
	for i := range source {
		want, got := source[i], copied[i]
		if want != got {
			return fmt.Errorf("row %d (%s): %+v != %+v", i, want.ID, want, got)
		}
	}
	return nil
}

// @MX:WARN: [AUTO] rebuildItemsTable drops and renames the live items table inside a migration transaction — a fault past the copy step is recoverable only by rollback, never by repair
// @MX:REASON: destructive DROP/RENAME sequence over operator data; every step after the copy must abort through the single error path so the original table and file stand untouched (REQ-THS-003)
// @MX:SPEC: SPEC-TODO-HOLD-STATE-001
// rebuildItemsTable migrates a stamped-v1 items table to the current
// four-state CHECK inside ONE transaction. On any failure the transaction
// rolls back and the original table (and file) stand exactly as they were.
// The caller has already run the idempotent DDL and ensureLandingColumn, so
// the source table carries the landing column this copy names.
func (e *backlogEngine) rebuildItemsTable(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, backlogOpTimeout)
	defer cancel()
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return mapBacklogEngineError(fmt.Sprintf("rebuild items %s: begin", e.dbPath), err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	source, err := readItemsRows(ctx, tx, "items")
	if err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: read source", e.dbPath), err)
		return err
	}

	if _, err = tx.ExecContext(ctx,
		fmt.Sprintf("CREATE TABLE items_new (%s)", backlogItemsTableColumns)); err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: create items_new", e.dbPath), err)
		return err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO items_new(seq, id, text, added_at, spec_id, landing, state)
		 SELECT seq, id, text, added_at, spec_id, landing, state FROM items`); err != nil {
		// A source row the new CHECK cannot represent lands here — abort,
		// never repair (REQ-THS-003).
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: copy rows", e.dbPath), err)
		return err
	}

	copied, err := readItemsRows(ctx, tx, "items_new")
	if err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: read copy", e.dbPath), err)
		return err
	}
	if err = assertRebuildParity(source, copied); err != nil {
		err = fmt.Errorf("rebuild items %s: parity verification failed: %w", e.dbPath, err)
		return err
	}

	// Parity holds — the switch. The DROP takes the old table and its index;
	// the rename promotes the copy; the index is rebuilt on the new table.
	if _, err = tx.ExecContext(ctx, `DROP TABLE items`); err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: drop old items", e.dbPath), err)
		return err
	}
	if _, err = tx.ExecContext(ctx, `ALTER TABLE items_new RENAME TO items`); err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: promote items_new", e.dbPath), err)
		return err
	}
	if _, err = tx.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_items_state ON items(state)`); err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: recreate state index", e.dbPath), err)
		return err
	}

	// REQ-THS-004: the stamp is part of the SAME transaction, so the commit
	// makes the new CHECK and the new stamp indivisible.
	if err = upsertMeta(ctx, tx, backlogMetaKeySchemaVersion, backlogSchemaVersion); err != nil {
		err = fmt.Errorf("rebuild items %s: %w", e.dbPath, err)
		return err
	}

	if err = tx.Commit(); err != nil {
		err = mapBacklogEngineError(fmt.Sprintf("rebuild items %s: commit", e.dbPath), err)
		return err
	}
	return nil
}
