// backlog_downgrade_test.go — the old-binary compat gate, REPOINTED by
// SPEC-TODO-HOLD-STATE-001.
//
// Under SPEC-TODO-LANDING-EVIDENCE-001 (card t359) this test proved that a
// pre-change binary STILL SERVES a database carrying the new column: the
// landing column was additive, so serving on was safe. The hold state is NOT
// additive — it is a table rebuild behind a schema_version bump — and the
// SPEC fixes the doctrine in the opposite direction (REQ-THS-005 /
// AC-THS-005): a binary that knows only the previous stamp REFUSES the
// database at open, refuse-to-operate, never repair-by-delete. The clauses
// below keep the frozen-replica machinery and assert the new direction:
//
//	(a) the pre-"2" STATEMENTS, verbatim, still run against the v2 database —
//	    what changed is the version gate, not statement shape.
//	(b) the reconstructed pre-change OPEN path — DDL exec, schema_version
//	    read, version switch — refuses the stamped-"2" database with
//	    ErrBacklogCorrupt naming unsupported schema_version, and (positive
//	    control) accepts a stamped-"1" database. This is AC-THS-005 enforced
//	    without any test seam: the frozen switch's accepted set IS the old
//	    binary's.
//	(c) the frozen replica has not drifted: its items CHECK is the
//	    THREE-value tuple the v1 stamp marks, while the live backlogDDL
//	    carries the FOUR-value tuple — both halves pinned, so an accidental
//	    edit to either trips.
//
// The clause (c) withdrawn at v0.3.1 (comparing accepted version SETS against
// the live switch) stays withdrawn — control flow is still not a structure a
// test can extract, and clause (b) now reds on exactly the drift that
// comparison existed to catch: the live gate's behavior versus a frozen
// prior-version gate, exercised on real databases.
package kanban

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// The frozen pre-"2" constants. They are LITERALS, deliberately not
// references to the live consts: a reference would track a bump instead of
// refusing it, which is the whole failure this criterion exists to detect.
const (
	frozenSchemaVersion    = "1"
	frozenMetaKeySchemaVer = "schema_version"
)

// frozenPreChangeDDL is a verbatim copy of the live backlogDDL as it stood
// before the hold state: the three-state items CHECK, no landing column in
// the CREATE (t359's ALTER appended it to databases in the field). Clause
// (c) pins it as the layout the "1" stamp marks.
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

// The old-binary compat gate, all three clauses.
func TestBacklogDowngrade_OldBinaryRefusesTheNewStamp(t *testing.T) {
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
		// names no landing column and writes a three-legal state, so it
		// succeeds against the v2 schema: statement shape is compatible even
		// where the version gate is not.
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
		if version != "2" {
			t.Errorf("schema_version = %q, want %q — the fixture must be the current layout", version, "2")
		}
	})

	t.Run("(b) the reconstructed pre-change open path refuses the v2 stamp", func(t *testing.T) {
		err := frozenPreChangeOpen(t, dbPath)
		if !errors.Is(err, ErrBacklogCorrupt) {
			t.Fatalf("pre-change open over a stamped-v2 database: err = %v, want ErrBacklogCorrupt (REQ-THS-005: refuse-to-operate)", err)
		}
		if !strings.Contains(err.Error(), "unsupported schema_version") {
			t.Errorf("refusal error = %v, want it to name unsupported schema_version", err)
		}

		// Positive control: the SAME frozen path accepts the stamp it knows —
		// the refusal is about the version, not the replica.
		_, v1Path := holdFixtureRoot(t)
		seedV1Database(t, v1Path)
		if err := frozenPreChangeOpen(t, v1Path); err != nil {
			t.Fatalf("pre-change open over a stamped-v1 database must succeed: %v", err)
		}
	})

	t.Run("(c) the frozen replica has not drifted from the v1 layout", func(t *testing.T) {
		// The frozen DDL is the layout the "1" stamp marks: its items CHECK
		// is the THREE-value tuple.
		const frozenCheck = "CHECK (state IN ('queued','picked','dropped'))"
		if !strings.Contains(frozenPreChangeDDL, frozenCheck) {
			t.Error("frozen v1 DDL replica lost the three-state items CHECK")
		}
		// The live DDL is the layout the "2" stamp marks: its items CHECK is
		// the FOUR-value tuple. Together the two pins mean no state can be
		// added to either layout without tripping this test.
		const liveCheck = "CHECK (state IN ('queued','picked','dropped','hold'))"
		if !strings.Contains(backlogDDL, liveCheck) {
			t.Errorf("live backlogDDL lost the four-state items CHECK.\nlive:\n%s", backlogDDL)
		}
	})
}

// frozenPreChangeOpen reconstructs the pre-change open path: DDL exec, then
// the schema_version read, then the version switch. It references the FROZEN
// constants throughout, so it behaves exactly as a binary that knows only
// stamp "1" behaves — which is what makes clause (b) an enforcement of
// REQ-THS-005 rather than a rehearsal.
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
