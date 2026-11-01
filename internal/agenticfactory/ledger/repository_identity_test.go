package ledger

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// issue-1948, audit 46dbcce0d4 N38 (ledger side). One ledger file belongs to one repository, but nothing
// in it said which: a verb run from a second clone of the project, or with the environment naming
// another repository, wrote the ledger of one project from the repository of another. The ledger
// now holds the repository identity (the git common directory, as a plain string the caller passes
// in; the ledger runs no git). It is written when first asked and never overwritten: the same value
// again is a no-op, a different one is refused. The verbs compare the repository they are about to
// use with it (lane-2).

const repoA = "/work/project/.git"
const repoB = "/work/other-clone/.git"

func TestRepository_AnUnboundLedgerReadsEmpty(t *testing.T) {
	s := openTemp(t)
	got, err := s.Repository(bg)
	if err != nil || got != "" {
		t.Fatalf("an unbound ledger: %q %v", got, err)
	}
}

func TestRepository_TheFirstBindingIsKeptAndSurvivesAReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatalf("the first binding: %v", err)
	}
	if got, err := s.Repository(bg); err != nil || got != repoA {
		t.Fatalf("read back: %q %v", got, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if got, err := s.Repository(bg); err != nil || got != repoA {
		t.Fatalf("after a reopen: %q %v", got, err)
	}
}

func TestRepository_TheSameIdentityAgainIsANoOp(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	var at1 string
	if err := s.db.QueryRowContext(bg, `SELECT bound_at FROM repository`).Scan(&at1); err != nil {
		t.Fatal(err)
	}
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatalf("the same identity again: %v", err)
	}
	var at2 string
	if err := s.db.QueryRowContext(bg, `SELECT bound_at FROM repository`).Scan(&at2); err != nil || at2 != at1 {
		t.Fatalf("the binding moved: %q -> %q %v", at1, at2, err)
	}
	if n := count(t, s, `SELECT count(*) FROM repository`); n != 1 {
		t.Fatalf("%d rows", n)
	}
}

func TestRepository_ADifferentIdentityIsRefusedAndNothingChanges(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	err := s.BindRepository(bg, repoB)
	if !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("a second, different identity: %v", err)
	}
	if msg := err.Error(); !strings.HasPrefix(msg, "ledger: ") || !strings.Contains(msg, repoA) || !strings.Contains(msg, repoB) {
		t.Errorf("the refusal must be a ledger error naming both repositories: %q", msg)
	}
	if got, err := s.Repository(bg); err != nil || got != repoA {
		t.Fatalf("the refused binding changed the ledger: %q %v", got, err)
	}
}

func TestRepository_ABlankIdentityIsInvalid(t *testing.T) {
	s := openTemp(t)
	for _, v := range []string{"", "  ", "\t\n"} {
		if err := s.BindRepository(bg, v); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q: %v", v, err)
		}
	}
	if got, _ := s.Repository(bg); got != "" {
		t.Errorf("a blank identity was bound: %q", got)
	}
}

func TestRepository_TwoFirstWritesLeaveExactlyOneWinner(t *testing.T) {
	s := openTemp(t)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, v := range []string{repoA, repoB} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = s.BindRepository(bg, v)
		}()
	}
	wg.Wait()
	wins := 0
	for _, err := range errs {
		switch {
		case err == nil:
			wins++
		case !errors.Is(err, ErrWrongRepository):
			t.Fatalf("a loser failed for another reason: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("%d winners: %v", wins, errs)
	}
	if got, err := s.Repository(bg); err != nil || (got != repoA && got != repoB) {
		t.Fatalf("the winner: %q %v", got, err)
	}
}

// I7 again: the binding is not rewritten or removed by anyone holding the file.
func TestRepository_TheBindingCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile(t *testing.T) {
	s := openTemp(t)
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE repository SET identity=?`, repoB); err == nil {
		t.Error("the binding was updated")
	}
	if _, err := s.db.ExecContext(bg, `DELETE FROM repository`); err == nil {
		t.Error("the binding was deleted")
	}
	if _, err := s.db.ExecContext(bg, `INSERT OR REPLACE INTO repository(id,identity,bound_at) VALUES(1,?,'x')`, repoB); err == nil {
		t.Error("the binding was replaced")
	}
	if got, err := s.Repository(bg); err != nil || got != repoA {
		t.Fatalf("%q %v", got, err)
	}
}

// The schema does not move: a file of version 1, and one of version 2 written before this table
// existed, both open into the schema that has it, unbound, and take a binding. A file of this
// code opens in an older binary exactly as before, because the version number is the same.
func TestRepository_FilesOfEarlierSchemasOpenIntoTheNewOne(t *testing.T) {
	if SchemaVersion != 2 {
		t.Fatalf("the schema version moved to %d; the repository table needs no version bump", SchemaVersion)
	}
	t.Run("version 1", func(t *testing.T) {
		path := v1Millisecond(t, false, nil)
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		if got, err := s.Repository(bg); err != nil || got != "" {
			t.Fatalf("%q %v", got, err)
		}
		if err := s.BindRepository(bg, repoA); err != nil {
			t.Fatal(err)
		}
		if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
			t.Errorf("user_version %d", n)
		}
	})
	t.Run("version 2 without the table", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "af", "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		damage(t, s, `DROP TABLE repository`)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		if err := s.BindRepository(bg, repoA); err != nil {
			t.Fatal(err)
		}
		if got, err := s.Repository(bg); err != nil || got != repoA {
			t.Fatalf("%q %v", got, err)
		}
		if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
			t.Errorf("user_version %d", n)
		}
	})
	t.Run("a binding of a file survives the migration of its other tables", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "af", "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.BindRepository(bg, repoA); err != nil {
			t.Fatal(err)
		}
		// Back to a version 1 file, as if an older build had opened it: the column goes, the version
		// number goes. The binding is a table of its own and stays.
		damage(t, s, `ALTER TABLE transitions DROP COLUMN evidence_mark`, `PRAGMA user_version=1`)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		if got, err := s.Repository(bg); err != nil || got != repoA {
			t.Fatalf("after the migration: %q %v", got, err)
		}
	})
}

func TestRepository_ABrokenFileReturnsALedgerError(t *testing.T) {
	s := openTemp(t)
	damage(t, s, `DROP TABLE repository`)
	_, err := s.Repository(bg)
	wantLedgerError(t, err, "repository")
	wantLedgerError(t, s.BindRepository(bg, repoA), "repository")
}
