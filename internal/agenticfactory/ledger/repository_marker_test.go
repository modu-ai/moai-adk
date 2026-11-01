package ledger

// Audit of a36ea8c368, N61 N65 N67 N68: the repository identity the ledger is bound to is a path, and a path is not a
// repository (a clone, a mirror or a re-created repository stands at it, or at the path a move names). The caller computes
// a marker (a random token kept inside the repository's git directory) and the ledger stores and compares it as a plain
// string; the ledger runs no git (REQ-AFL-008). N68: the first recorded move raises PRAGMA user_version, so a build that
// does not know the move refuses the file instead of accepting the old path of a moved repository.

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

	// previousBuildSchema is the schema number the builds before the move record knew: a file with a higher number is
	// refused by them ("the file has schema N, this code knows 2").
	previousBuildSchema = 2
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

// MoveRepositoryMarked is the move and the marker in one transaction.
func TestRepositoryMarker_AMarkedMoveRecordsTheMarkerWithTheMove(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepositoryMarked(bg, repoA, repoB, "moved", markerA); err != nil {
		t.Fatalf("the move: %v", err)
	}
	if got, err := s.Repository(bg); err != nil || got != repoB {
		t.Errorf("identity %q %v", got, err)
	}
	if got, err := s.RepositoryMarker(bg); err != nil || got != markerA {
		t.Errorf("marker %q %v, want the marker the move was given", got, err)
	}
}

func TestRepositoryMarker_AMarkedMoveWithAnotherMarkerThanTheRecordedOneChangesNothing(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordRepositoryMarker(bg, markerA); err != nil {
		t.Fatal(err)
	}
	err := s.MoveRepositoryMarked(bg, repoA, repoB, "moved", markerB)
	if !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("error %v, want ErrWrongRepository", err)
	}
	if moves, _ := s.RepositoryMoves(bg); len(moves) != 0 {
		t.Errorf("a refused move was recorded: %+v", moves)
	}
	if got, _ := s.Repository(bg); got != repoA {
		t.Errorf("identity %q after the refusal, want %q", got, repoA)
	}
	if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
		t.Errorf("user_version %d after a refused move, want %d", n, SchemaVersion)
	}
	// the same move with the recorded marker goes through
	if err := s.MoveRepositoryMarked(bg, repoA, repoB, "moved", markerA); err != nil {
		t.Fatalf("with the recorded marker: %v", err)
	}
}

func TestRepositoryMarker_AMoveWithoutAMarkerLeavesTheRecordAlone(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordRepositoryMarker(bg, markerA); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.RepositoryMarker(bg); got != markerA {
		t.Errorf("marker %q after an unmarked move, want %q", got, markerA)
	}
}

// ---- N68: the first recorded move raises the schema number ------------------------------------------------------------

func TestSchemaVersion_ALedgerNeverMovedStaysReadableByThePreviousBuild(t *testing.T) {
	s, path := openAt(t)
	if n := count(t, s, `PRAGMA user_version`); n != previousBuildSchema {
		t.Fatalf("a fresh ledger has user_version %d, want %d", n, previousBuildSchema)
	}
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordRepositoryMarker(bg, markerA); err != nil {
		t.Fatal(err)
	}
	if n := count(t, s, `PRAGMA user_version`); n != previousBuildSchema {
		t.Errorf("a bound, marked ledger that never moved has user_version %d, want %d", n, previousBuildSchema)
	}
	// and this build opens it again
	s = reopen(t, s, path)
	if got, _ := s.Repository(bg); got != repoA {
		t.Errorf("identity %q after a reopen", got)
	}
}

func TestSchemaVersion_TheFirstRecordedMoveRaisesItSoThePreviousBuildRefusesTheFile(t *testing.T) {
	s, path := openAt(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
		t.Fatal(err)
	}
	n := count(t, s, `PRAGMA user_version`)
	if n <= previousBuildSchema {
		t.Fatalf("user_version %d after a recorded move: a build that knows schema %d would accept the file and the old path", n, previousBuildSchema)
	}
	if n != SchemaVersionMoved {
		t.Errorf("user_version %d, want SchemaVersionMoved (%d)", n, SchemaVersionMoved)
	}
	// the previous build's rule, as it was written, refuses have > 2: the n > previousBuildSchema check above is that rule firing
	// this build keeps opening the file, and a second move leaves the number where it is
	s = reopen(t, s, path)
	if got, err := s.Repository(bg); err != nil || got != repoB {
		t.Fatalf("after a reopen: %q %v", got, err)
	}
	if err := s.MoveRepository(bg, repoB, repoA, "back"); err != nil {
		t.Fatal(err)
	}
	if got := count(t, s, `PRAGMA user_version`); got != SchemaVersionMoved {
		t.Errorf("user_version %d after the second move, want %d", got, SchemaVersionMoved)
	}
}

func TestSchemaVersion_ARefusedMoveDoesNotRaiseIt(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoB, repoA, "wrong start"); err == nil {
		t.Fatal("a move from an identity that is not current was accepted")
	}
	if err := s.MoveRepository(bg, repoA, repoA, "nowhere"); err == nil {
		t.Fatal("a move to the same identity was accepted")
	}
	if n := count(t, s, `PRAGMA user_version`); n != previousBuildSchema {
		t.Errorf("user_version %d after refused moves, want %d", n, previousBuildSchema)
	}
}

func TestSchemaVersion_ThisBuildOpensBothAndRefusesWhatIsNewerThanBoth(t *testing.T) {
	for _, tc := range []struct {
		version int
		opens   bool
	}{{1, true}, {previousBuildSchema, true}, {SchemaVersionMoved, true}, {SchemaVersionMoved + 1, false}, {99, false}} {
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
			want := tc.version
			if want < previousBuildSchema {
				want = previousBuildSchema // an older file is brought to the schema of a ledger that never moved
			}
			if n := count(t, s, `PRAGMA user_version`); n != want {
				t.Errorf("schema %d: user_version %d after the open, want %d", tc.version, n, want)
			}
			_ = s.Close()
		}
	}
}
