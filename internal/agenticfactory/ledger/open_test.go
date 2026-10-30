package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var bg = context.Background()

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(bg, filepath.Join(t.TempDir(), "db", "proj", "af", "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpenCreatesTheDirectoryTheSchemaAndTheSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db", "proj", "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the file: %v", err)
	}
	for _, table := range []string{"cards", "transitions", "leases", "lease_keys", "evidence", "merge_queue", "locks", "notes"} {
		var n int
		if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&n); err != nil || n != 1 {
			t.Errorf("table %s: %d %v", table, n, err)
		}
	}
	var mode string
	var busy, ver int
	if err := s.db.QueryRowContext(bg, `PRAGMA journal_mode`).Scan(&mode); err != nil || mode != "wal" {
		t.Errorf("journal_mode %q %v", mode, err)
	}
	if err := s.db.QueryRowContext(bg, `PRAGMA busy_timeout`).Scan(&busy); err != nil || busy != int(BusyTimeout.Milliseconds()) {
		t.Errorf("busy_timeout %d %v", busy, err)
	}
	if err := s.db.QueryRowContext(bg, `PRAGMA user_version`).Scan(&ver); err != nil || ver != SchemaVersion {
		t.Errorf("user_version %d %v", ver, err)
	}
}

func TestOpenTwiceIsTheSameLedger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.db")
	a, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.db.ExecContext(bg, `INSERT INTO notes(kind,text,at) VALUES('decision','kept','2026-10-30T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	_ = a.Close()
	b, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()
	var n int
	if err := b.db.QueryRowContext(bg, `SELECT count(*) FROM notes`).Scan(&n); err != nil || n != 1 {
		t.Errorf("reopened ledger lost its rows: %d %v", n, err)
	}
}

func TestTheSchemaRefusesWhatTheContractClosesOff(t *testing.T) {
	s := openTemp(t)
	if _, err := s.db.ExecContext(bg, `INSERT INTO notes(kind,text,at) VALUES('gossip','x','t')`); err == nil {
		t.Error("a note kind outside the closed set was accepted")
	}
	if _, err := s.db.ExecContext(bg, `INSERT INTO merge_queue(card_id,candidate_sha,state,enqueued_at,updated_at) VALUES('c','s','paused','t','t')`); err == nil {
		t.Error("a queue state outside the closed set was accepted")
	}
}

func TestTheEngineVersionIsChecked(t *testing.T) {
	for v, ok := range map[string]bool{"3.35.0": true, "3.35.5": true, "3.40.1": true, "3.46.0": true, "4.0.0": true, "3.34.9": false, "3.9.0": false, "2.9.9": false, "nonsense": false, "": false} {
		if err := checkVersion(v); (err == nil) != ok {
			t.Errorf("%q: %v", v, err)
		}
	}
}

// A power loss must not lose an epoch bump, or two claims could share one epoch.
func TestOpenSetsSynchronousFull(t *testing.T) {
	s := openTemp(t)
	var sync int
	if err := s.db.QueryRowContext(bg, `PRAGMA synchronous`).Scan(&sync); err != nil || sync != 2 {
		t.Errorf("synchronous = %d (2 is FULL) %v", sync, err)
	}
}

func TestAPathWithASpaceOrOtherURLCharactersOpens(t *testing.T) {
	for _, dir := range []string{"my dir", "a#b", "a%20b", "q?x", "ünï"} {
		path := filepath.Join(t.TempDir(), dir, "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Errorf("%q: %v", dir, err)
			continue
		}
		if err := s.AddCard(bg, Card{ID: "c", Title: "x", Position: coreReady()}); err != nil {
			t.Errorf("%q: %v", dir, err)
		}
		_ = s.Close()
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%q: the file is not where the path says: %v", dir, err)
		}
	}
}

// I7 at the database: the four history tables refuse a delete, whoever asks.
func TestRowsOfTheHistoryTablesCannotBeDeleted(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	if _, err := s.Claim(bg, req("s", 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `INSERT INTO notes(kind,card_id,text,at) VALUES('decision','a','x','2026-10-30T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `INSERT INTO evidence(card_id,kind,tree_sha,result,epoch,at) VALUES('a','red','t','fail',1,'2026-10-30T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"cards", "transitions", "evidence", "notes"} {
		_, err := s.db.ExecContext(bg, `DELETE FROM `+table)
		if err == nil || !strings.Contains(err.Error(), "never deleted") {
			t.Errorf("DELETE FROM %s: %v", table, err)
		}
		var n int
		if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM `+table).Scan(&n); err != nil || n == 0 {
			t.Errorf("%s lost its rows: %d %v", table, n, err)
		}
	}
	// the lease rows are the ones that go
	if _, err := s.db.ExecContext(bg, `DELETE FROM lease_keys`); err != nil {
		t.Errorf("lease_keys: %v", err)
	}
	if _, err := s.db.ExecContext(bg, `DELETE FROM leases`); err != nil {
		t.Errorf("leases: %v", err)
	}
}
