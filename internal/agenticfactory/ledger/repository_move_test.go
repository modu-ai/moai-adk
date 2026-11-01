package ledger

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// issue-1948, audit 1c05083f71 N52 (ledger side). The ledger holds one repository identity, bound by
// the first caller and never overwritten (N38). When the repository directory is moved on disk, every
// verb is refused and nothing could record the move, so a card that had landed on the trunk could
// never finish. The ledger now takes a recorded move: an append-only row naming the identity the
// caller believes is current, the identity it moved to, a reason, and the database clock's time.
// The current identity is the newest move's new identity, or the first binding when there is none.
// The first-binding row and its UPDATE and DELETE triggers are untouched. The ledger runs no git
// (REQ-AFL-008); the identity strings are the caller's. The verb that calls this is lane-2's.

const repoC = "/elsewhere/project/.git"

func boundTo(t *testing.T, identity string) *Store {
	t.Helper()
	s := openTemp(t)
	if err := s.BindRepository(bg, identity); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRepositoryMove_RecordsTheMoveAndTheCurrentIdentityFollowsIt(t *testing.T) {
	s := boundTo(t, repoA)
	before := time.Now().UTC().Add(-time.Minute)
	if err := s.MoveRepository(bg, repoA, repoB, "the directory was renamed"); err != nil {
		t.Fatalf("the move: %v", err)
	}
	if got, err := s.Repository(bg); err != nil || got != repoB {
		t.Fatalf("the current identity after the move: %q %v", got, err)
	}
	moves, err := s.RepositoryMoves(bg)
	if err != nil || len(moves) != 1 {
		t.Fatalf("the moves: %v %v", moves, err)
	}
	m := moves[0]
	if m.From != repoA || m.To != repoB || m.Reason != "the directory was renamed" || m.Seq < 1 {
		t.Errorf("the recorded move: %+v", m)
	}
	if m.At.Before(before) || m.At.After(time.Now().UTC().Add(time.Minute)) {
		t.Errorf("the move was stamped %v, not by the database clock now", m.At)
	}
}

func TestRepositoryMove_AnUnmovedLedgerHasNoMoves(t *testing.T) {
	s := boundTo(t, repoA)
	moves, err := s.RepositoryMoves(bg)
	if err != nil || len(moves) != 0 {
		t.Fatalf("%v %v", moves, err)
	}
	if got, _ := s.Repository(bg); got != repoA {
		t.Fatalf("%q", got)
	}
}

func TestRepositoryMove_TheMoveSurvivesAReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.BindRepository(bg, repoA); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if got, err := s.Repository(bg); err != nil || got != repoB {
		t.Fatalf("after a reopen: %q %v", got, err)
	}
	if moves, err := s.RepositoryMoves(bg); err != nil || len(moves) != 1 {
		t.Fatalf("after a reopen: %v %v", moves, err)
	}
}

func TestRepositoryMove_SeveralMovesChainInOrderAndTheNewestIsCurrent(t *testing.T) {
	s := boundTo(t, repoA)
	if err := s.MoveRepository(bg, repoA, repoB, "first"); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoB, repoC, "second"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Repository(bg); err != nil || got != repoC {
		t.Fatalf("%q %v", got, err)
	}
	moves, err := s.RepositoryMoves(bg)
	if err != nil || len(moves) != 2 {
		t.Fatalf("%v %v", moves, err)
	}
	if moves[0].From != repoA || moves[0].To != repoB || moves[0].Reason != "first" ||
		moves[1].From != repoB || moves[1].To != repoC || moves[1].Reason != "second" || moves[0].Seq >= moves[1].Seq {
		t.Errorf("the history must read oldest first and chain: %+v", moves)
	}
}

func TestRepositoryMove_MovingBackToAnEarlierIdentityIsAMoveLikeAnyOther(t *testing.T) {
	s := boundTo(t, repoA)
	if err := s.MoveRepository(bg, repoA, repoB, "away"); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoB, repoA, "back"); err != nil {
		t.Fatalf("back to the first path: %v", err)
	}
	if got, err := s.Repository(bg); err != nil || got != repoA {
		t.Fatalf("%q %v", got, err)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 2 {
		t.Fatalf("%d moves", n)
	}
}

// BindRepository keeps its meaning on top of the moves: the current identity binds as a no-op, an
// older one is refused, and the refusal names both.
func TestRepositoryMove_BindRepositoryFollowsTheCurrentIdentity(t *testing.T) {
	s := boundTo(t, repoA)
	if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
		t.Fatal(err)
	}
	if err := s.BindRepository(bg, repoB); err != nil {
		t.Fatalf("binding the current identity: %v", err)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 1 {
		t.Errorf("binding the current identity wrote a move: %d rows", n)
	}
	if n := count(t, s, `SELECT count(*) FROM repository`); n != 1 {
		t.Errorf("binding the current identity wrote a binding: %d rows", n)
	}
	err := s.BindRepository(bg, repoA)
	if !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("binding the identity from before the move: %v", err)
	}
	if msg := err.Error(); !strings.HasPrefix(msg, "ledger: ") || !strings.Contains(msg, repoA) || !strings.Contains(msg, repoB) {
		t.Errorf("the refusal must name the current and the offered identity: %q", msg)
	}
	if err := s.BindRepository(bg, repoC); !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("binding an unrelated identity: %v", err)
	}
	if got, _ := s.Repository(bg); got != repoB {
		t.Fatalf("a refused binding changed the current identity: %q", got)
	}
}

func TestRepositoryMove_AMoveMustStartFromTheCurrentIdentity(t *testing.T) {
	s := boundTo(t, repoA)
	err := s.MoveRepository(bg, repoB, repoC, "from a path the ledger is not at")
	if !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("a move from an identity that is not current: %v", err)
	}
	if msg := err.Error(); !strings.HasPrefix(msg, "ledger: ") || !strings.Contains(msg, repoA) || !strings.Contains(msg, repoB) {
		t.Errorf("the refusal must be a ledger error naming the current and the offered identity: %q", msg)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
		t.Errorf("a refused move was recorded: %d rows", n)
	}
	if got, _ := s.Repository(bg); got != repoA {
		t.Errorf("%q", got)
	}
}

// A second caller that read the identity before the first move holds a stale from: its move is
// refused, so two callers cannot fork the history.
func TestRepositoryMove_AStaleMoveIsRefusedAndTheHistoryDoesNotFork(t *testing.T) {
	s := boundTo(t, repoA)
	if err := s.MoveRepository(bg, repoA, repoB, "first caller"); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveRepository(bg, repoA, repoC, "second caller, stale"); !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("the stale move: %v", err)
	}
	if got, _ := s.Repository(bg); got != repoB {
		t.Errorf("the stale move changed the current identity: %q", got)
	}
	if moves, _ := s.RepositoryMoves(bg); len(moves) != 1 || moves[0].To != repoB {
		t.Errorf("%+v", moves)
	}
}

func TestRepositoryMove_AnUnboundLedgerHasNothingToMoveFrom(t *testing.T) {
	s := openTemp(t)
	err := s.MoveRepository(bg, repoA, repoB, "nothing bound")
	if !errors.Is(err, ErrWrongRepository) {
		t.Fatalf("a move on an unbound ledger: %v", err)
	}
	wantLedgerError(t, err, "not bound")
	if got, _ := s.Repository(bg); got != "" {
		t.Errorf("the refused move bound the ledger: %q", got)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
		t.Errorf("%d rows", n)
	}
}

// Moving to the identity it already has is refused (not a silent no-op): the caller's belief that
// the repository moved is wrong, and the history should not carry a move that moved nothing.
func TestRepositoryMove_MovingToTheSameIdentityIsRefused(t *testing.T) {
	s := boundTo(t, repoA)
	err := s.MoveRepository(bg, repoA, repoA, "no change")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("a move to the same identity: %v", err)
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
		t.Errorf("%d rows", n)
	}
}

func TestRepositoryMove_BlankValuesAreInvalid(t *testing.T) {
	s := boundTo(t, repoA)
	cases := map[string][3]string{
		"blank from":   {"", repoB, "why"},
		"spaces from":  {"  ", repoB, "why"},
		"blank to":     {repoA, "", "why"},
		"spaces to":    {repoA, "\t\n", "why"},
		"blank reason": {repoA, repoB, ""},
		"space reason": {repoA, repoB, "   "},
	}
	for name, c := range cases {
		if err := s.MoveRepository(bg, c[0], c[1], c[2]); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
		t.Errorf("%d rows", n)
	}
	if got, _ := s.Repository(bg); got != repoA {
		t.Errorf("%q", got)
	}
}

func TestRepositoryMove_TwoMovesFromTheSameIdentityLeaveExactlyOneWinner(t *testing.T) {
	s := boundTo(t, repoA)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, to := range []string{repoB, repoC} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = s.MoveRepository(bg, repoA, to, "racing")
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
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 1 {
		t.Fatalf("%d moves", n)
	}
}

// I7 once more: the moves are an append-only record, and the first binding stays as it was.
func TestRepositoryMove_TheRecordCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile(t *testing.T) {
	s := boundTo(t, repoA)
	if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE repository_move SET new_identity=?`, repoC); err == nil {
		t.Error("a move was updated")
	}
	if _, err := s.db.ExecContext(bg, `DELETE FROM repository_move`); err == nil {
		t.Error("a move was deleted")
	}
	if _, err := s.db.ExecContext(bg, `INSERT OR REPLACE INTO repository_move(seq,old_identity,new_identity,reason,moved_at) VALUES(1,?,?,'x','2026-11-01T00:00:00.000Z')`, repoA, repoC); err == nil {
		t.Error("a move was replaced")
	}
	if got, err := s.Repository(bg); err != nil || got != repoB {
		t.Fatalf("%q %v", got, err)
	}
	var first string
	if err := s.db.QueryRowContext(bg, `SELECT identity FROM repository WHERE id=1`).Scan(&first); err != nil || first != repoA {
		t.Fatalf("the first binding changed: %q %v", first, err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE repository SET identity=?`, repoB); err == nil {
		t.Error("the first binding was updated after a move")
	}
	if _, err := s.db.ExecContext(bg, `DELETE FROM repository`); err == nil {
		t.Error("the first binding was deleted after a move")
	}
}

// The table refuses, by itself, a row the API would not write.
func TestRepositoryMove_TheTableRefusesBlankOrUnchangedRows(t *testing.T) {
	s := boundTo(t, repoA)
	for name, args := range map[string][]any{
		"blank old":    {"", repoB, "why"},
		"blank new":    {repoA, "", "why"},
		"blank reason": {repoA, repoB, ""},
		"unchanged":    {repoA, repoA, "why"},
	} {
		if _, err := s.db.ExecContext(bg, `INSERT INTO repository_move(old_identity,new_identity,reason,moved_at) VALUES(?,?,?,'2026-11-01T00:00:00.000Z')`, args...); err == nil {
			t.Errorf("%s: the row was accepted", name)
		}
	}
	if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
		t.Errorf("%d rows", n)
	}
}

// The table is additive, like repository: a file of version 1 and one of version 2 written before
// this table existed both open into the schema that has it, and take a binding and then a move.
func TestRepositoryMove_FilesOfEarlierSchemasOpenIntoTheNewOne(t *testing.T) {
	if SchemaVersion != 2 {
		t.Fatalf("the schema version moved to %d; the move table needs no version bump", SchemaVersion)
	}
	moveOnce := func(t *testing.T, s *Store) {
		t.Helper()
		if moves, err := s.RepositoryMoves(bg); err != nil || len(moves) != 0 {
			t.Fatalf("a file without the table: %v %v", moves, err)
		}
		if err := s.BindRepository(bg, repoA); err != nil {
			t.Fatal(err)
		}
		if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
			t.Fatalf("the move: %v", err)
		}
		if got, err := s.Repository(bg); err != nil || got != repoB {
			t.Fatalf("%q %v", got, err)
		}
		if n := count(t, s, `PRAGMA user_version`); n != SchemaVersion {
			t.Errorf("user_version %d", n)
		}
	}
	t.Run("version 1", func(t *testing.T) {
		s, err := Open(bg, v1Millisecond(t, false, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		moveOnce(t, s)
	})
	t.Run("version 2 without the table", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "af", "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		damage(t, s, `DROP TABLE repository_move`)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		moveOnce(t, s)
	})
	t.Run("a recorded move survives the migration of the other tables", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "af", "ledger.db")
		s, err := Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.BindRepository(bg, repoA); err != nil {
			t.Fatal(err)
		}
		if err := s.MoveRepository(bg, repoA, repoB, "moved"); err != nil {
			t.Fatal(err)
		}
		damage(t, s, `ALTER TABLE transitions DROP COLUMN evidence_mark`, `PRAGMA user_version=1`)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(bg, path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		if got, err := s.Repository(bg); err != nil || got != repoB {
			t.Fatalf("after the migration: %q %v", got, err)
		}
		if moves, err := s.RepositoryMoves(bg); err != nil || len(moves) != 1 {
			t.Fatalf("after the migration: %v %v", moves, err)
		}
	})
}

// A file written by newer code is refused, as it was before the move table existed.
func TestRepositoryMove_AFileOfANewerSchemaIsStillRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	damage(t, s, `PRAGMA user_version=3`)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(bg, path)
	if err == nil {
		_ = s.Close()
		t.Fatal("a file of schema 3 opened")
	}
	wantLedgerError(t, err, "schema 3")
}

// Failure paths: every error branch of the new calls comes back as a "ledger: ..." error and leaves
// the rows as they were.
func TestRepositoryMove_FailurePathsAreLedgerErrors(t *testing.T) {
	t.Run("the move table is gone", func(t *testing.T) {
		s := boundTo(t, repoA)
		damage(t, s, `DROP TABLE repository_move`)
		_, err := s.Repository(bg)
		wantLedgerError(t, err, "repository_move")
		_, err = s.RepositoryMoves(bg)
		wantLedgerError(t, err, "repository_move")
		wantLedgerError(t, s.MoveRepository(bg, repoA, repoB, "x"), "repository_move")
		wantLedgerError(t, s.BindRepository(bg, repoA), "repository_move")
	})
	t.Run("the binding table is gone", func(t *testing.T) {
		s := boundTo(t, repoA)
		damage(t, s, `DROP TABLE repository`)
		_, err := s.Repository(bg)
		wantLedgerError(t, err, "repository")
		wantLedgerError(t, s.MoveRepository(bg, repoA, repoB, "x"), "repository")
	})
	t.Run("a column is renamed", func(t *testing.T) {
		s := boundTo(t, repoA)
		damage(t, s, `ALTER TABLE repository_move RENAME COLUMN new_identity TO new_identity_gone`)
		_, err := s.Repository(bg)
		wantLedgerError(t, err, "no such column")
		wantLedgerError(t, s.MoveRepository(bg, repoA, repoB, "x"), "no such column")
		_, err = s.RepositoryMoves(bg)
		wantLedgerError(t, err, "no such column")
	})
	t.Run("the engine refuses the write", func(t *testing.T) {
		s := boundTo(t, repoA)
		damage(t, s, `CREATE TRIGGER move_refused BEFORE INSERT ON repository_move BEGIN SELECT RAISE(ABORT,'move refused'); END`)
		wantLedgerError(t, s.MoveRepository(bg, repoA, repoB, "x"), "move refused")
		if got, _ := s.Repository(bg); got != repoA {
			t.Errorf("a refused write changed the current identity: %q", got)
		}
		if n := count(t, s, `SELECT count(*) FROM repository_move`); n != 0 {
			t.Errorf("%d rows", n)
		}
	})
	t.Run("a row whose time cannot be read", func(t *testing.T) {
		s := boundTo(t, repoA)
		damage(t, s, `INSERT INTO repository_move(old_identity,new_identity,reason,moved_at) VALUES('a','b','x','not a time')`)
		_, err := s.RepositoryMoves(bg)
		wantLedgerError(t, err, "parsing time")
	})
	t.Run("a closed ledger", func(t *testing.T) {
		s := boundTo(t, repoA)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		wantLedgerError(t, s.MoveRepository(bg, repoA, repoB, "x"), "closed")
		_, err := s.RepositoryMoves(bg)
		wantLedgerError(t, err, "closed")
		_, err = s.Repository(bg)
		wantLedgerError(t, err, "closed")
	})
}
