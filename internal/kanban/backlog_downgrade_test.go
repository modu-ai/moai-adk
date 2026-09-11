// backlog_downgrade_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M5:
// AC-TLE-018 — a binary built before the landing column still serves a
// database that carries it.
//
// The criterion has three clauses and they are NOT interchangeable:
//
//	(a) the pre-change production statements, verbatim, still run.
//	(b) a reconstruction of the pre-change OPEN path — DDL exec, schema_version
//	    read, version switch — still opens the database without classifying it
//	    as corrupt. This is the path clause (a) never reaches.
//	(c) the frozen replica has not drifted from the live source.
//
// Clause (c) is what makes (b) a detector rather than a coverage assertion.
// Without it the replica goes stale silently while the test keeps reporting
// that it exercises the pre-change open path. Its RED is: edit the live
// backlogDDL and leave the frozen copy untouched — (a) and (b) both stay
// green and only (c) fails.
//
// One conjunct of (c) was WITHDRAWN as unrealisable at v0.3.1 and is
// deliberately NOT reconstructed here: comparing the frozen switch's accepted
// version set against the live one. The live set is control flow, not a
// structure a test can extract, and its cheapest runnable reading (comparing
// the version const) leaves the named drift undetected, because adding a case
// to the live switch changes no const.
//
// What this does NOT demonstrate: the replica is compiled from TODAY's source.
// A divergence between it and a genuinely older released binary is invisible
// to all three clauses.
package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

// The frozen pre-change constants. They are LITERALS, deliberately not
// references to the live consts: a reference would track a bump instead of
// refusing it, which is the whole failure this criterion exists to detect.
const (
	frozenSchemaVersion    = "1"
	frozenMetaKeySchemaVer = "schema_version"
)

// frozenPreChangeDDL is a verbatim copy of the live backlogDDL as it stood
// before the landing column was introduced. Clause (c) asserts it still
// equals the live const byte for byte.
const frozenPreChangeDDL = `
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
);
CREATE TABLE IF NOT EXISTS findings (
  subject_id TEXT  NOT NULL,
  related_id TEXT  NOT NULL,
  relation   TEXT  NOT NULL,
  source     TEXT  NOT NULL,
  score      REAL  NOT NULL,
  note       TEXT  NOT NULL DEFAULT '',
  at         TEXT  NOT NULL
);
CREATE TABLE IF NOT EXISTS archived_items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL,
  position INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS archived_findings (
  archive_seq INTEGER NOT NULL,
  position    INTEGER NOT NULL,
  subject_id  TEXT  NOT NULL,
  related_id  TEXT  NOT NULL,
  relation    TEXT  NOT NULL,
  source      TEXT  NOT NULL,
  score       REAL  NOT NULL,
  note        TEXT  NOT NULL DEFAULT '',
  at          TEXT  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_items_state ON items(state);
`

// AC-TLE-018 — all three clauses, against one post-change database.
func TestBacklogDowngrade_PreChangeBinaryStillServes(t *testing.T) {
	dbPath := postChangeDatabaseWithEvidence(t)

	t.Run("(a) pre-change statements run verbatim", func(t *testing.T) {
		db := openRawBacklogDB(t, dbPath)

		// The SELECT a pre-change binary issued, character for character.
		rows, err := db.Query(`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq`)
		if err != nil {
			t.Fatalf("pre-change SELECT failed: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var got string
		for rows.Next() {
			var id, text, addedAt, state string
			var specID sql.NullString
			if err := rows.Scan(&id, &text, &addedAt, &specID, &state); err != nil {
				t.Fatalf("pre-change scan failed: %v", err)
			}
			if got != "" {
				got += ";"
			}
			got += fmt.Sprintf("%s|%s|%s|%s|%s", id, text, addedAt, specID.String, state)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("pre-change iterate failed: %v", err)
		}
		const want = "t1|alpha work|2026-01-01T00:00:00Z||queued"
		if got != want {
			t.Errorf("pre-change SELECT returned\n %s\nwant\n %s", got, want)
		}

		// The INSERT a pre-change binary issued, character for character. It
		// names no landing column, so it can only succeed while that column
		// stays nullable and defaulted.
		if _, err := db.Exec(
			`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, ?, ?)`,
			99, "t99", "written by a pre-change binary", "2026-01-05T00:00:00Z", nil, "queued"); err != nil {
			t.Fatalf("pre-change INSERT failed: %v", err)
		}

		var version string
		if err := db.QueryRow(
			`SELECT value FROM meta WHERE key = ?`, frozenMetaKeySchemaVer).Scan(&version); err != nil {
			t.Fatalf("read schema_version: %v", err)
		}
		if version != frozenSchemaVersion {
			t.Errorf("schema_version = %q, want %q — a bump breaks every older binary", version, frozenSchemaVersion)
		}
	})

	t.Run("(b) the reconstructed pre-change open path accepts the database", func(t *testing.T) {
		if err := frozenPreChangeOpen(t, dbPath); err != nil {
			t.Fatalf("reconstructed pre-change open failed: %v", err)
		}
		// Named separately from the error check: the criterion asks not merely
		// that the open succeeded but that the database was not classified
		// corrupt, and a future refactor could return a non-nil non-corrupt
		// error without that distinction being visible.
		if err := frozenPreChangeOpen(t, dbPath); errors.Is(err, ErrBacklogCorrupt) {
			t.Errorf("reconstructed open classified the database as corrupt: %v", err)
		}
	})

	t.Run("(c) the frozen replica has not drifted from the live source", func(t *testing.T) {
		if frozenPreChangeDDL != backlogDDL {
			t.Errorf("frozen DDL replica has drifted from the live backlogDDL const.\nfrozen:\n%s\nlive:\n%s",
				frozenPreChangeDDL, backlogDDL)
		}
	})
}

// frozenPreChangeOpen reconstructs the pre-change open path: DDL exec, then
// the schema_version read, then the version switch — the sequence the live
// ensureSchema still runs, minus the landing-column step that did not exist.
//
// It references the FROZEN constants throughout, so a bump of the live
// schemaVersion reds this path exactly as it would reject a real older
// binary's open.
func frozenPreChangeOpen(t *testing.T, dbPath string) error {
	t.Helper()
	db := openRawBacklogDB(t, dbPath)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, frozenPreChangeDDL); err != nil {
		return fmt.Errorf("schema %s: %w", dbPath, err)
	}

	var version string
	err := db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = ?`, frozenMetaKeySchemaVer).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		version = ""
	} else if err != nil {
		return fmt.Errorf("read schema_version %s: %w", dbPath, err)
	}

	switch version {
	case "":
		if _, err := db.ExecContext(ctx,
			`INSERT INTO meta(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
			frozenMetaKeySchemaVer, frozenSchemaVersion); err != nil {
			return fmt.Errorf("stamp schema_version %s: %w", dbPath, err)
		}
	case frozenSchemaVersion:
		// current layout
	default:
		return fmt.Errorf("schema %s: unsupported schema_version %q (want %q): %w",
			dbPath, version, frozenSchemaVersion, ErrBacklogCorrupt)
	}
	return nil
}

// postChangeDatabaseWithEvidence builds a database through the LIVE open path
// and records landing evidence on its one card, so the pre-change statements
// above are exercised against a row that actually carries a value in the new
// column rather than against a NULL that would prove nothing.
func postChangeDatabaseWithEvidence(t *testing.T) string {
	t.Helper()
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Items[0].AddedAt = "2026-01-01T00:00:00Z"
		rec.Items[0].Landing = &LandingEvidence{
			Ref:        "origin/main",
			RefHead:    "7777777777777777777777777777777777777777",
			ObservedAt: "2026-01-06T00:00:00Z",
		}
		return nil
	}); err != nil {
		t.Fatalf("record landing: %v", err)
	}
	return store.EnginePath()
}

// openRawBacklogDB opens the database file directly, bypassing the engine, so
// a test can issue statements the engine would never issue itself.
func openRawBacklogDB(t *testing.T, dbPath string) *sql.DB {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, backlogDSN(dbPath))
	if err != nil {
		t.Fatalf("open %s: %v", dbPath, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
