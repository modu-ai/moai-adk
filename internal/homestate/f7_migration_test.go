package homestate

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// frSeedV1 builds a raw schema-version-1 factory database whose
// resume_handoffs table predates the claim columns, with three rows: a
// claimed handoff with a valid claimed_at, one with an unparseable
// claimed_at, and one that was never claimed.
func frSeedV1(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "factory.db")
	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("open raw seed: %v", err)
	}
	defer func() { _ = raw.Close() }()
	stmts := []string{
		`CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`INSERT INTO meta(key,value) VALUES('schema_version','1')`,
		`CREATE TABLE resume_handoffs (id INTEGER PRIMARY KEY, status TEXT NOT NULL, claimed_at TEXT NOT NULL DEFAULT '')`,
		`INSERT INTO resume_handoffs(id,status,claimed_at) VALUES (1,'claimed','2026-01-02T03:04:05Z')`,
		`INSERT INTO resume_handoffs(id,status,claimed_at) VALUES (2,'claimed','not-a-timestamp')`,
		`INSERT INTO resume_handoffs(id,status,claimed_at) VALUES (3,'pending','')`,
	}
	for _, s := range stmts {
		if _, err := raw.Exec(s); err != nil {
			t.Fatalf("seed %q: %v", s, err)
		}
	}
	return path
}

// Opening a v1 database migrates it all the way to v4 and backfills the v1
// claim bookkeeping: a parseable claimed_at becomes a claim expiry five
// minutes later; an unparseable one flags the row for legacy recovery; rows
// the backfill does not touch keep a NULL claim expiry (the v1→v2 ALTER adds
// the column without a default).
func TestOpenFactoryV1ClaimedHandoffBackfill(t *testing.T) {
	path := frSeedV1(t)
	db, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatalf("open migrated v1 factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if got := frSchemaVersion(t, db.DB); got != "4" {
		t.Fatalf("schema_version = %q, want \"4\"", got)
	}
	rows, err := db.DB.Query(`SELECT id, ifnull(claim_expires_at,''), legacy_recovery, legacy_recovery_reason FROM resume_handoffs ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	got := map[int][3]string{}
	for rows.Next() {
		var id int
		var exp, flag, reason string
		if err := rows.Scan(&id, &exp, &flag, &reason); err != nil {
			t.Fatal(err)
		}
		got[id] = [3]string{exp, flag, reason}
	}
	if r := got[1]; r[0] != "2026-01-02T03:09:05Z" || r[1] != "0" || r[2] != "" {
		t.Fatalf("claimed row 1 = %v, want backfilled 03:09:05Z expiry, no flag", r)
	}
	if r := got[2]; r[0] != "" || r[1] != "1" || r[2] == "" {
		t.Fatalf("claimed row 2 = %v, want no expiry, legacy-recovery flag with reason", r)
	}
	if r := got[3]; r[0] != "" || r[1] != "0" || r[2] != "" {
		t.Fatalf("unclaimed row 3 = %v, want untouched", r)
	}
}

// A schema version the binary does not know is refused, not guessed at.
func TestOpenFactoryUnsupportedSchemaVersion(t *testing.T) {
	path := frSeedV1(t)
	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`UPDATE meta SET value='9' WHERE key='schema_version'`); err != nil {
		t.Fatal(err)
	}
	_ = raw.Close()
	_, err = OpenFactoryPath(path)
	if err == nil || !strings.Contains(err.Error(), `unsupported factory schema version "9"`) {
		t.Fatalf("err = %v, want unsupported factory schema version \"9\"", err)
	}
}

// A file that is not a database fails at DDL time, not at first use.
func TestOpenFactoryCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "factory.db")
	if err := os.WriteFile(path, []byte("this is not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenFactoryPath(path); err == nil || !strings.Contains(err.Error(), "initialize factory database") {
		t.Fatalf("err = %v, want initialize factory database failure", err)
	}
}

// A database path whose parent cannot be created as a directory fails
// cleanly.
func TestOpenFactoryParentIsFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenFactoryPath(filepath.Join(blocker, "factory.db")); err == nil || !strings.Contains(err.Error(), "mkdir") {
		t.Fatalf("open under a file parent: err = %v, want MkdirAll (mkdir) failure", err)
	}
}
