// backlog_migrate.go — record I/O over the SQLite engine, plus the lazy
// one-time migration that carries a legacy backlog.json into it
// (SPEC-TODO-SQLITE-001 REQ-TOSQ-004/005/011..014, M2).
//
// The store's in-memory shape is unchanged: BacklogRecord is still what every
// caller sees, and every mutation is still a whole-record read-modify-write
// under the sibling advisory lock. Only where the bytes rest changed. A write
// is therefore one transaction that replaces the items and findings tables
// wholesale — at the queue's scale (hundreds of rows) that costs microseconds,
// and it keeps the ordering contract trivially exact: array position IS the
// stored order, so a `todo move` that reorders the slice round-trips without
// any reconciliation logic.
//
// Why `seq` is a POSITION, not the id's number (deviation from design.md §2's
// inline comment, which reads "seq mirrors t<N>"): `todo move`
// (internal/cli/todo_edit_move.go) reorders rec.Items, after which array order
// and id order differ. REQ-TOSQ-004 binds array order — "reproducing exactly
// the array order the legacy JSON file preserved" — so seq carries position.
// Id integrity is unaffected: it rests on UNIQUE(id) plus meta.last_seq
// (REQ-TOSQ-005), neither of which reads seq.
//
// No SQL statement interpolates user text; every value travels through a
// placeholder (acceptance.md D.6 Secured gate).
package kanban

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// backlogMigratedSuffix names the quarantine the legacy file is RENAMED to on
// a successful migration (REQ-TOSQ-014). Never deleted, never truncated.
const backlogMigratedSuffix = ".migrated"

// backlogOpTimeout bounds a single record read or write.
const backlogOpTimeout = 30 * time.Second

// readRecord loads the whole queue record. Items come back in stored order
// (ORDER BY seq — REQ-TOSQ-004); findings in insertion order, which is the
// index `todo unrelate` addresses.
func (e *backlogEngine) readRecord(ctx context.Context) (*BacklogRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpTimeout)
	defer cancel()

	rec := &BacklogRecord{
		Version:  backlogVersion,
		Items:    []BacklogItem{},
		Findings: []BacklogFinding{},
	}

	itemRows, err := e.db.QueryContext(ctx,
		`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq`)
	if err != nil {
		return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
	}
	defer func() { _ = itemRows.Close() }()
	for itemRows.Next() {
		var it BacklogItem
		var specID sql.NullString
		var state string
		if err := itemRows.Scan(&it.ID, &it.Text, &it.AddedAt, &specID, &state); err != nil {
			return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
		}
		if specID.Valid {
			// The pointer shape is the frozen per-item contract: an absent
			// spec id round-trips as JSON null, never as an omitted key.
			v := specID.String
			it.SpecID = &v
		}
		it.State = BacklogState(state)
		rec.Items = append(rec.Items, it)
	}
	if err := itemRows.Err(); err != nil {
		return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
	}

	findingRows, err := e.db.QueryContext(ctx,
		`SELECT subject_id, related_id, relation, source, score, note, at FROM findings ORDER BY rowid`)
	if err != nil {
		return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
	}
	defer func() { _ = findingRows.Close() }()
	for findingRows.Next() {
		var f BacklogFinding
		if err := findingRows.Scan(&f.SubjectID, &f.RelatedID, &f.Relation, &f.Source, &f.Score, &f.Note, &f.At); err != nil {
			return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
		}
		rec.Findings = append(rec.Findings, f)
	}
	if err := findingRows.Err(); err != nil {
		return nil, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
	}

	lastSeq, err := e.readLastSeq(ctx)
	if err != nil {
		return nil, err
	}
	rec.LastSeq = lastSeq

	normalizeBacklogRecord(rec)
	return rec, nil
}

// readLastSeq reads the persisted high-water mark; an unstamped database
// reads as 0, which normalizeBacklogRecord then lifts to max-present.
func (e *backlogEngine) readLastSeq(ctx context.Context) (int, error) {
	var raw string
	err := e.db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = ?`, backlogMetaKeyLastSeq).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, mapBacklogEngineError(fmt.Sprintf("load backlog %s", e.dbPath), err)
	}
	n, convErr := strconv.Atoi(raw)
	if convErr != nil {
		// A meta row that is not an integer is a corrupt image, not a value to
		// guess at. Refuse; never repair.
		return 0, fmt.Errorf("load backlog %s: last_seq %q is not an integer: %w",
			e.dbPath, raw, ErrBacklogCorrupt)
	}
	return n, nil
}

// @MX:ANCHOR: [AUTO] writeRecord — the sole write path onto the queue database
// @MX:REASON: expected fan_in >= 3 (Mutate, migration cutover, export round-trip); id integrity rests on this one transaction
//
// writeRecord replaces the stored record with rec inside ONE transaction
// (REQ-TOSQ-005): a process that dies mid-write leaves the prior state intact,
// and last_seq can never advance apart from the items it issued. A duplicate
// id violates UNIQUE(id), which aborts the transaction and surfaces as the
// named id-conflict error rather than a half-written queue.
func (e *backlogEngine) writeRecord(ctx context.Context, rec *BacklogRecord) (err error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpTimeout)
	defer cancel()

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return mapBacklogEngineError(fmt.Sprintf("write backlog %s", e.dbPath), err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM items`); err != nil {
		return mapBacklogEngineError(fmt.Sprintf("write backlog %s", e.dbPath), err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM findings`); err != nil {
		return mapBacklogEngineError(fmt.Sprintf("write backlog %s", e.dbPath), err)
	}

	for i, it := range rec.Items {
		var specID any
		if it.SpecID != nil {
			specID = *it.SpecID
		}
		// seq = 1-based array position: the ordering contract (REQ-TOSQ-004).
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, ?, ?)`,
			i+1, it.ID, it.Text, it.AddedAt, specID, string(it.State)); err != nil {
			err = mapBacklogWriteError(e.dbPath, it.ID, err)
			return err
		}
	}

	for _, f := range rec.Findings {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO findings(subject_id, related_id, relation, source, score, note, at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			f.SubjectID, f.RelatedID, f.Relation, f.Source, f.Score, f.Note, f.At); err != nil {
			return mapBacklogEngineError(fmt.Sprintf("write backlog %s", e.dbPath), err)
		}
	}

	if err = upsertMeta(ctx, tx, backlogMetaKeyLastSeq, strconv.Itoa(rec.LastSeq)); err != nil {
		return fmt.Errorf("write backlog %s: %w", e.dbPath, err)
	}
	if err = upsertMeta(ctx, tx, backlogMetaKeySchemaVersion, backlogSchemaVersion); err != nil {
		return fmt.Errorf("write backlog %s: %w", e.dbPath, err)
	}

	if err = tx.Commit(); err != nil {
		return mapBacklogEngineError(fmt.Sprintf("write backlog %s", e.dbPath), err)
	}
	return nil
}

// mapBacklogWriteError classifies an insert failure, naming the offending id
// when the engine reports a constraint violation.
func mapBacklogWriteError(dbPath, id string, err error) error {
	var coded interface{ Code() int }
	if errors.As(err, &coded) && coded.Code()&0xff == sqliteConstraintCode {
		return fmt.Errorf("write backlog %s: item %s: %w: %w", dbPath, id, ErrBacklogIDConflict, err)
	}
	return mapBacklogEngineError(fmt.Sprintf("write backlog %s", dbPath), err)
}

// sqliteConstraintCode is SQLITE_CONSTRAINT (19).
const sqliteConstraintCode = 19

// upsertMeta writes one meta key.
func upsertMeta(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO meta(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value)
	if err != nil {
		return mapBacklogEngineError("meta "+key, err)
	}
	return nil
}

// countQueued returns the queued-state item count through a single aggregate —
// the constant-cost path the statusline reads per render (REQ-TOSQ-009).
func (e *backlogEngine) countQueued(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpTimeout)
	defer cancel()
	var n int
	if err := e.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM items WHERE state = ?`, string(BacklogStateQueued)).Scan(&n); err != nil {
		return 0, mapBacklogEngineError(fmt.Sprintf("count backlog %s", e.dbPath), err)
	}
	return n, nil
}

// countByState returns the picked and queued counts through ONE grouped
// aggregate over the state index — the statusline's constant-cost path.
func (e *backlogEngine) countByState(ctx context.Context) (picked, queued int, err error) {
	ctx, cancel := context.WithTimeout(ctx, backlogOpTimeout)
	defer cancel()
	rows, err := e.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM items GROUP BY state`)
	if err != nil {
		return 0, 0, mapBacklogEngineError(fmt.Sprintf("count backlog %s", e.dbPath), err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var state string
		var n int
		if scanErr := rows.Scan(&state, &n); scanErr != nil {
			return 0, 0, mapBacklogEngineError(fmt.Sprintf("count backlog %s", e.dbPath), scanErr)
		}
		switch BacklogState(state) {
		case BacklogStatePicked:
			picked = n
		case BacklogStateQueued:
			queued = n
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, mapBacklogEngineError(fmt.Sprintf("count backlog %s", e.dbPath), err)
	}
	return picked, queued, nil
}

// ---------------------------------------------------------------------------
// Migration (design.md §4 state machine)
// ---------------------------------------------------------------------------

// backlogLayout reports which artifacts exist under a queue root — the state
// the migration machine dispatches on:
//
//	A {no db, no json}  → open empty db (first run of the new binary)
//	B {no db, json}     → MIGRATE
//	C {db, no json}     → steady state
//	D {db, json}        → db authoritative; complete the quarantine best-effort
type backlogLayout struct {
	dbExists   bool
	jsonExists bool
}

// inspectBacklogLayout reads the layout without changing anything.
func inspectBacklogLayout(queuePath string) backlogLayout {
	return backlogLayout{
		dbExists:   fileExists(backlogSQLitePath(queuePath)),
		jsonExists: fileExists(queuePath),
	}
}

// fileExists reports whether path names an existing file. An unreadable path
// counts as absent — the caller's next step surfaces the real error with
// context rather than this predicate guessing at one.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// loadLegacyBacklogJSON reads a legacy backlog.json. The malformed contract is
// inherited verbatim from the pre-SQLite store: a parse failure surfaces with
// the file untouched — there is no repair-on-load path, because the operator's
// queued intent is the one thing that cannot be regenerated (C-6).
func loadLegacyBacklogJSON(path string) (*BacklogRecord, error) {
	raw, err := atomicfile.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load backlog %s: %w", path, err)
	}
	var rec BacklogRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("load backlog %s: parsing: %w", path, err)
	}
	normalizeBacklogRecord(&rec)
	return &rec, nil
}

// migrateLegacyBacklog carries the legacy JSON at queuePath into a fresh
// database beside it. The caller MUST already hold the queue lock: this is the
// single-flight window design.md §4 step 1 describes, and the double-check
// below is what makes a second process that waited on the lock observe state C
// and skip rather than migrate twice.
//
// Parity is verified BEFORE the legacy file stops being authoritative
// (REQ-TOSQ-011). On ANY failure the partial database and its -wal/-shm
// siblings are removed and the legacy file is left exactly as it was
// (REQ-TOSQ-012) — the caller keeps serving the queue it had.
func migrateLegacyBacklog(queuePath string) error {
	dbPath := backlogSQLitePath(queuePath)
	if fileExists(dbPath) {
		// Another process won the lock and migrated first: state C.
		return nil
	}

	source, err := loadLegacyBacklogJSON(queuePath)
	if err != nil {
		return err
	}

	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		removeBacklogDBArtifacts(dbPath)
		return err
	}

	ctx := context.Background()
	if err := eng.writeRecord(ctx, source); err != nil {
		_ = eng.close()
		removeBacklogDBArtifacts(dbPath)
		return fmt.Errorf("migrate backlog %s -> %s: %w", queuePath, dbPath, err)
	}

	migrated, err := eng.readRecord(ctx)
	if err != nil {
		_ = eng.close()
		removeBacklogDBArtifacts(dbPath)
		return fmt.Errorf("migrate backlog %s -> %s: re-reading for parity: %w", queuePath, dbPath, err)
	}
	if err := assertBacklogParity(source, migrated); err != nil {
		_ = eng.close()
		removeBacklogDBArtifacts(dbPath)
		return fmt.Errorf("migrate backlog %s -> %s: parity check failed, legacy file left authoritative: %w",
			queuePath, dbPath, err)
	}
	if err := eng.close(); err != nil {
		removeBacklogDBArtifacts(dbPath)
		return fmt.Errorf("migrate backlog %s -> %s: closing: %w", queuePath, dbPath, err)
	}

	// Authority flips here and not before.
	quarantineLegacyBacklog(queuePath)
	return nil
}

// quarantineLegacyBacklog RENAMES the legacy file to its .migrated sibling
// (REQ-TOSQ-014) — byte-preserved, never deleted. Best-effort by contract
// (REQ-TOSQ-013): the database is already authoritative, so a failed rename
// leaves a visible state-D pair rather than failing the caller's verb, and the
// next open retries.
//
// An existing .migrated is NOT overwritten. A rename onto it would destroy the
// bytes the first migration quarantined, which is the one thing no path in
// this SPEC may do. Leaving the pair keeps the divergence visible instead of
// silently eating it (design.md R4).
func quarantineLegacyBacklog(queuePath string) {
	if !fileExists(queuePath) {
		return
	}
	target := queuePath + backlogMigratedSuffix
	if fileExists(target) {
		return
	}
	_ = os.Rename(queuePath, target)
}

// removeBacklogDBArtifacts deletes a partially written database and its WAL
// siblings. It touches ONLY the .db/-wal/-shm triple this migration just
// created — never the legacy file, never a quarantine.
func removeBacklogDBArtifacts(dbPath string) {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		_ = os.Remove(dbPath + suffix)
	}
}

// assertBacklogParity compares the migrated record field-for-field against the
// in-memory source (REQ-TOSQ-011). It is deliberately exhaustive rather than
// count-based: a migration that moved the right NUMBER of rows while dropping
// a spec id or reordering findings would pass a count check and lose operator
// data silently.
func assertBacklogParity(source, migrated *BacklogRecord) error {
	if len(source.Items) != len(migrated.Items) {
		return fmt.Errorf("item count %d != %d", len(source.Items), len(migrated.Items))
	}
	for i := range source.Items {
		want, got := source.Items[i], migrated.Items[i]
		if want.ID != got.ID || want.Text != got.Text || want.AddedAt != got.AddedAt || want.State != got.State {
			return fmt.Errorf("item %d: %+v != %+v", i, want, got)
		}
		if !equalSpecID(want.SpecID, got.SpecID) {
			return fmt.Errorf("item %d (%s): spec_id %v != %v", i, want.ID, derefSpecID(want.SpecID), derefSpecID(got.SpecID))
		}
	}
	if len(source.Findings) != len(migrated.Findings) {
		return fmt.Errorf("finding count %d != %d", len(source.Findings), len(migrated.Findings))
	}
	for i := range source.Findings {
		if source.Findings[i] != migrated.Findings[i] {
			return fmt.Errorf("finding %d: %+v != %+v", i, source.Findings[i], migrated.Findings[i])
		}
	}
	if source.LastSeq != migrated.LastSeq {
		return fmt.Errorf("last_seq %d != %d", source.LastSeq, migrated.LastSeq)
	}
	if source.Version != migrated.Version {
		return fmt.Errorf("version %d != %d", source.Version, migrated.Version)
	}
	return nil
}

// equalSpecID compares two spec-id pointers by null-shape AND value: nil and a
// pointer to "" are different records, and conflating them would let a
// migration quietly invent a picked card.
func equalSpecID(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// derefSpecID renders a spec-id pointer for an error message.
func derefSpecID(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

// backlogDirName is the directory the queue artifacts share.
func backlogDirName(queuePath string) string { return filepath.Dir(queuePath) }
