package ledger

// Audit of a36ea8c368, N61 N65 N67 N68: the repository identity the ledger is bound to is a path, and a path is not a
// repository (a clone, a mirror or a re-created repository stands at it, or at the path a move names). The caller computes
// a marker (a random token kept inside the repository's git directory) and the ledger stores and compares it as a plain
// string; the ledger runs no git (REQ-AFL-008).

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const (
	markerA = "0123456789abcdef0123456789abcdef"
	markerB = "fedcba9876543210fedcba9876543210"
)

func TestRepositoryMarker_NoneUntilRecorded(t *testing.T) {
	s := openTemp(t)
	if got, err := s.RepositoryMarker(bg); err != nil || got != "" {
		t.Fatalf("an unmarked ledger answers %q %v, want an empty marker", got, err)
	}
}

func TestRepositoryMarker_TheFirstRecordWinsAndTheRecordIsKept(t *testing.T) {
	s, path := openAt(t)
	got, err := s.RecordRepositoryMarker(bg, markerA)
	if err != nil || got != markerA {
		t.Fatalf("first record: %q %v", got, err)
	}
	// a second caller with another marker learns the one on record and changes nothing
	got, err = s.RecordRepositoryMarker(bg, markerB)
	if err != nil || got != markerA {
		t.Fatalf("second record: %q %v, want the first marker back", got, err)
	}
	// the same marker again is the same record
	if got, err = s.RecordRepositoryMarker(bg, markerA); err != nil || got != markerA {
		t.Fatalf("same marker again: %q %v", got, err)
	}
	s = reopen(t, s, path)
	if got, err := s.RepositoryMarker(bg); err != nil || got != markerA {
		t.Fatalf("after a reopen: %q %v", got, err)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_marker`); n != 1 {
		t.Errorf("%d marker rows, want exactly one", n)
	}
}

func TestRepositoryMarker_ABlankMarkerIsRefused(t *testing.T) {
	s := openTemp(t)
	for _, m := range []string{"", "  ", "\t\n"} {
		if _, err := s.RecordRepositoryMarker(bg, m); !errors.Is(err, ErrInvalid) {
			t.Errorf("marker %q: %v, want ErrInvalid", m, err)
		}
	}
	if got, _ := s.RepositoryMarker(bg); got != "" {
		t.Errorf("a refused record left the marker %q", got)
	}
}

func TestRepositoryMarker_TheRecordIsNeverRewrittenOrDeleted(t *testing.T) {
	s := openTemp(t)
	if _, err := s.RecordRepositoryMarker(bg, markerA); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`UPDATE repository_marker SET marker='x'`,
		`DELETE FROM repository_marker`,
		`INSERT OR REPLACE INTO repository_marker(id,marker,recorded_at) VALUES(1,'x','now')`,
	} {
		if _, err := s.db.Exec(q); err == nil {
			t.Errorf("%s was accepted", q)
		}
	}
	if got, _ := s.RepositoryMarker(bg); got != markerA {
		t.Errorf("marker %q after the attempts, want %q", got, markerA)
	}
}

// Concurrent creators converge: whatever they each bring, every one is told the single value on record.
func TestRepositoryMarker_ConcurrentRecordsConverge(t *testing.T) {
	s := openTemp(t)
	const n = 12
	got := make([]string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m := strings.Repeat(string(rune('a'+i)), 32)
			got[i], _ = s.RecordRepositoryMarker(bg, m)
		}(i)
	}
	wg.Wait()
	for i := range got {
		if got[i] == "" || got[i] != got[0] {
			t.Fatalf("callers were told %q, want one value", got)
		}
	}
	if n := count(t, s, `SELECT count(*) FROM repository_marker`); n != 1 {
		t.Errorf("%d marker rows", n)
	}
}

// The marker table is additive: a file of an earlier build opens into the schema that has it.
func TestRepositoryMarker_AFileWithoutTheTableOpensIntoOneThatHasIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	damage(t, s, `DROP TABLE repository_marker`)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.RecordRepositoryMarker(bg, markerA); err != nil {
		t.Fatal(err)
	}
	if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
		t.Errorf("user_version %d after recording a marker, want %d: only a recorded move raises it", n, SchemaVersion)
	}
}

// A ledger file of schema 1 or 2 opens into the current schema; a file written by newer code is refused.
func TestSchemaVersion_ThisBuildOpensItsOwnAndOlderFilesAndRefusesNewerOnes(t *testing.T) {
	for _, tc := range []struct {
		version int
		opens   bool
	}{{1, true}, {SchemaVersion, true}, {SchemaVersion + 1, false}, {99, false}} {
		path := filepath.Join(t.TempDir(), "af", "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		if tc.version == 1 {
			damage(t, s, `ALTER TABLE transitions DROP COLUMN evidence_mark`)
		}
		damage(t, s, "PRAGMA user_version="+strconv.Itoa(tc.version))
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(bg, path)
		switch {
		case tc.opens && err != nil:
			t.Errorf("schema %d: %v, want it to open", tc.version, err)
		case !tc.opens && err == nil:
			t.Errorf("schema %d opened", tc.version)
		case !tc.opens:
			wantLedgerError(t, err, "schema "+strconv.Itoa(tc.version))
		}
		if s != nil {
			if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
				t.Errorf("schema %d: user_version %d after the open, want %d", tc.version, n, SchemaVersion)
			}
			_ = s.Close()
		}
	}
}
