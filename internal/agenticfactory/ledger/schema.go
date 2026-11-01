package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // the pure Go driver: no cgo, one binary
)

// SchemaVersion is PRAGMA user_version of a ledger this code wrote. Version 2 adds
// transitions.evidence_mark (see migrateEvidenceMark).
const SchemaVersion = 2

// timeLayout is how the database clock writes a time (I4).
const timeLayout = "2006-01-02T15:04:05.000Z"

// nowSQL is the database clock as an SQL expression; offsetSQL(d) is d from now.
const nowSQL = `strftime('%Y-%m-%dT%H:%M:%fZ','now')`

func offsetSQL(d time.Duration) string {
	return fmt.Sprintf(`strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ','now','%+.3f seconds')`, d.Seconds())
}

// The schema of contract §2, plus two additive columns on cards (park_evidence,
// park_resume) that a core.Park needs.
const schemaDDL = `
CREATE TABLE IF NOT EXISTS cards(
  id TEXT PRIMARY KEY, title TEXT NOT NULL, stage TEXT NOT NULL, sub_stage TEXT,
  parked INTEGER NOT NULL DEFAULT 0, park_reason TEXT, park_evidence TEXT, park_resume TEXT,
  priority INTEGER NOT NULL DEFAULT 0, spec_json TEXT NOT NULL, epoch INTEGER NOT NULL DEFAULT 0,
  integrated INTEGER NOT NULL DEFAULT 0, -- a MERGING card whose remote merge was seen
  reclaim_count INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS cards_pick ON cards(stage, parked, priority DESC);
-- transitions: one row per change of position. The first row of a card is its adding
-- (from_stage NULL). A row whose from and to are the same position is not an edge: it
-- records a claim again, a reclaim, or a park in place. evidence_mark is the id of the card's
-- newest evidence row when the move was made: evidence and transitions count in separate
-- sequences, so this is what says which came first.
CREATE TABLE IF NOT EXISTS transitions(
  seq INTEGER PRIMARY KEY AUTOINCREMENT, card_id TEXT NOT NULL, from_stage TEXT, from_sub TEXT,
  to_stage TEXT NOT NULL, to_sub TEXT, epoch INTEGER NOT NULL, actor TEXT NOT NULL, reason TEXT, at TEXT NOT NULL,
  evidence_mark INTEGER);
CREATE INDEX IF NOT EXISTS transitions_card ON transitions(card_id);
CREATE TABLE IF NOT EXISTS leases(
  card_id TEXT PRIMARY KEY, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
  worktree TEXT NOT NULL, heartbeat_at TEXT NOT NULL, next_wake_at TEXT, expires_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lease_keys(key TEXT PRIMARY KEY, card_id TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS lease_keys_card ON lease_keys(card_id);
CREATE TABLE IF NOT EXISTS evidence(
  id INTEGER PRIMARY KEY AUTOINCREMENT, card_id TEXT NOT NULL, kind TEXT NOT NULL, tree_sha TEXT NOT NULL,
  subject_sha TEXT, base_sha TEXT, command TEXT, result TEXT NOT NULL, path TEXT, epoch INTEGER NOT NULL, at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS evidence_card ON evidence(card_id);
CREATE TABLE IF NOT EXISTS merge_queue(
  card_id TEXT PRIMARY KEY, candidate_sha TEXT NOT NULL, trunk_before TEXT,
  state TEXT NOT NULL CHECK(state IN ('waiting','verifying','landing','landed','failed')),
  fail_reason TEXT, card_epoch INTEGER NOT NULL DEFAULT 0, enqueued_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS locks(
  name TEXT PRIMARY KEY, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
  expires_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS notes(
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL CHECK(kind IN ('landed','decision','gate_wait','stalled')),
  card_id TEXT, text TEXT NOT NULL, at TEXT NOT NULL);
-- I7: the history is never deleted, by anyone holding the file
CREATE TRIGGER IF NOT EXISTS cards_no_delete BEFORE DELETE ON cards BEGIN SELECT RAISE(ABORT,'ledger rows are never deleted'); END;
CREATE TRIGGER IF NOT EXISTS transitions_no_delete BEFORE DELETE ON transitions BEGIN SELECT RAISE(ABORT,'ledger rows are never deleted'); END;
CREATE TRIGGER IF NOT EXISTS evidence_no_delete BEFORE DELETE ON evidence BEGIN SELECT RAISE(ABORT,'ledger rows are never deleted'); END;
CREATE TRIGGER IF NOT EXISTS notes_no_delete BEFORE DELETE ON notes BEGIN SELECT RAISE(ABORT,'ledger rows are never deleted'); END;
`

// Open opens or creates the ledger at path: its directory (0700), WAL mode, the
// busy timeout, immediate write transactions (I3), the schema, and the engine version
// check (ErrSQLiteTooOld).
func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	q := url.Values{}
	q.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", BusyTimeout.Milliseconds()))
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(FULL)")
	q.Add("_pragma", "recursive_triggers(1)") // so INSERT OR REPLACE fires the delete triggers
	q.Set("_txlock", "immediate")
	u := url.URL{Scheme: "file", Path: path, RawQuery: strings.ReplaceAll(q.Encode(), "%28", "(")}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, fmt.Errorf("ledger: %w", err)
	}
	s := &Store{db: db}
	if err := s.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init(ctx context.Context) error {
	var v string
	if err := s.db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&v); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if err := checkVersion(v); err != nil {
		return err
	}
	var have int
	if err := s.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&have); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if have > SchemaVersion {
		return fmt.Errorf("ledger: the file has schema %d, this code knows %d", have, SchemaVersion)
	}
	if _, err := s.db.ExecContext(ctx, schemaDDL); err != nil {
		return fmt.Errorf("ledger: schema: %w", err)
	}
	if have < SchemaVersion {
		return s.migrate(ctx)
	}
	return nil
}

// migrate brings a ledger written by older code to SchemaVersion in one write transaction, so a
// second process opening the same file at the same moment waits and then finds it done.
func (s *Store) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // nothing to undo after a commit
	if err := migrateEvidenceMark(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version=%d`, SchemaVersion)); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}

// migrateEvidenceMark adds transitions.evidence_mark to a version 1 ledger. Its rows are given the
// mark the old data supports: the newest evidence the card had at or before the moment of the
// move. The two clocks only share the database's own millisecond timestamps, so evidence written
// in the very same millisecond is ranked by what the move is used for. For a move back into
// BUILDING it counts as before the move, the side that withholds a DONE. For the move that offers
// a candidate to the queue (into MERGING) it counts as after: the mark there says which
// green is newer work than the offer, and a tie must not make newer work look older. A fresh
// ledger already has the column and no rows to mark.
func migrateEvidenceMark(ctx context.Context, q querier) error {
	var n int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM pragma_table_info('transitions') WHERE name='evidence_mark'`).Scan(&n); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if n == 0 {
		if _, err := q.ExecContext(ctx, `ALTER TABLE transitions ADD COLUMN evidence_mark INTEGER`); err != nil {
			return fmt.Errorf("ledger: migrate: %w", err)
		}
	}
	if _, err := q.ExecContext(ctx, `UPDATE transitions SET evidence_mark=COALESCE((SELECT MAX(e.id) FROM evidence e WHERE e.card_id=transitions.card_id AND (e.at<transitions.at OR (e.at=transitions.at AND transitions.to_stage<>'MERGING'))),0) WHERE evidence_mark IS NULL`); err != nil {
		return fmt.Errorf("ledger: migrate: %w", err)
	}
	return nil
}

// checkVersion refuses an engine older than MinSQLite.
func checkVersion(v string) error {
	min := strings.Split(MinSQLite, ".")
	got := strings.Split(strings.TrimSpace(v), ".")
	for i := range min {
		if i >= len(got) {
			return fmt.Errorf("%w (found %q)", ErrSQLiteTooOld, v)
		}
		g, err := strconv.Atoi(got[i])
		if err != nil {
			return fmt.Errorf("%w (found %q)", ErrSQLiteTooOld, v)
		}
		m, _ := strconv.Atoi(min[i])
		switch {
		case g > m:
			return nil
		case g < m:
			return fmt.Errorf("%w (found %q)", ErrSQLiteTooOld, v)
		}
	}
	return nil
}

// Close closes the ledger.
func (s *Store) Close() error { return s.db.Close() }

func parseTime(s string) (time.Time, error) { return time.Parse(timeLayout, s) }
