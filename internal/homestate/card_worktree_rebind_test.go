package homestate

// card_worktree_rebind_test.go — SPEC-WORKTREE-HOME-ROOT-001 M5 RED (card
// t1640): AC-WHR-011, the card record rebind.
//
// Written before the rebind operation exists — RebindCardWorktree and the
// registered-target probe are undefined here, so this package fails to
// compile and that compile failure is the RED (plan §C test-compilation
// order; research §R13's M5 row: the tests name M5 symbols plus the
// existing store surface).
//
// The pins: a rebind (card, old, new) updates EVERY run row of the card
// carrying the old path, appends one card.fields event recording both
// paths, leaves PreviousCardWorktree answering the new path; a stale old
// path is refused; an unregistered target is refused through the
// registered-target probe (REQ-WHR-011(b)); the reverse rebind after a
// reverse move succeeds symmetrically; and RecordCardWorktree's
// never-moves-a-card refusal is unchanged.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// rebindNow is the fixed clock every rebind fixture writes with.
var rebindNow = time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)

// rebindPick records cardID at picked for runID carrying worktree.
func rebindPick(t *testing.T, db *FactoryDB, runID, cardID, worktree string) {
	t.Helper()
	fields := CardFields{}
	if worktree != "" {
		fields.WorktreePath = &worktree
	}
	if _, err := db.RecordPicked(context.Background(), runID, cardID, fields, "rebind-test", rebindNow); err != nil {
		t.Fatalf("record picked %s/%s: %v", runID, cardID, err)
	}
}

// rebindOpen opens a factory database in a temp directory.
func rebindOpen(t *testing.T) *FactoryDB {
	t.Helper()
	db, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatalf("open factory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestRebindCardWorktree_RebindsEveryRunRowOfTheCard pins AC-WHR-011's
// positive shape: every run row of the card carrying P_old carries P_new
// afterwards, one card.fields event per rebound row records both paths, and
// PreviousCardWorktree answers P_new for the other run.
func TestRebindCardWorktree_RebindsEveryRunRowOfTheCard(t *testing.T) {
	db := rebindOpen(t)
	pOld := filepath.Join(t.TempDir(), "old-moai", "t7")
	pNew := filepath.Join(t.TempDir(), "new-home", "t7")
	rebindPick(t, db, "run-a", "t7", pOld)
	rebindPick(t, db, "run-b", "t7", pOld)
	// The migration moved the tree to P_new before rebinding: the target is
	// a registered worktree (the probe's default mechanical check).
	if err := os.MkdirAll(filepath.Join(pNew, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir moved-tree .git: %v", err)
	}

	rows, err := db.RebindCardWorktree(context.Background(), "t7", pOld, pNew, "rebind-test", rebindNow)
	if err != nil {
		t.Fatalf("RebindCardWorktree: %v", err)
	}
	if rows < 2 {
		t.Errorf("rows rebound = %d, want at least the two seeded rows", rows)
	}
	for _, run := range []string{"run-a", "run-b"} {
		c := rebindCard(t, db, run, "t7")
		if c.WorktreePath != pNew {
			t.Errorf("%s worktree path = %q, want %q", run, c.WorktreePath, pNew)
		}
	}
	if prev, ok, err := db.PreviousCardWorktree(context.Background(), "t7", "run-b"); err != nil || !ok || prev != pNew {
		t.Errorf("PreviousCardWorktree = (%q, %v, %v), want (%q, true, nil)", prev, ok, err, pNew)
	}
	// One card.fields event per rebound run records both paths.
	for _, run := range []string{"run-a", "run-b"} {
		var payload string
		err := db.DB.QueryRowContext(context.Background(),
			`SELECT payload_json FROM events WHERE run_id = ? AND kind = 'card.fields' ORDER BY created_at DESC LIMIT 1`,
			run).Scan(&payload)
		if err != nil {
			t.Errorf("%s: card.fields event: %v", run, err)
			continue
		}
		if !strings.Contains(payload, pOld) || !strings.Contains(payload, pNew) {
			t.Errorf("%s: card.fields payload %q does not record both paths", run, payload)
		}
	}
}

// rebindCard loads one card row of one run.
func rebindCard(t *testing.T, db *FactoryDB, runID, cardID string) Card {
	t.Helper()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s/%s: %v", runID, cardID, err)
	}
	return c
}

// TestRebindCardWorktree_StaleOldPathIsRefused pins the negative case: a
// rebind whose old path does not match the recorded value is refused and
// changes nothing.
func TestRebindCardWorktree_StaleOldPathIsRefused(t *testing.T) {
	db := rebindOpen(t)
	pOld := filepath.Join(t.TempDir(), "old-moai", "t7")
	pNew := filepath.Join(t.TempDir(), "new-home", "t7")
	rebindPick(t, db, "run-a", "t7", pOld)

	if _, err := db.RebindCardWorktree(context.Background(), "t7", "/some/other/old", pNew, "rebind-test", rebindNow); err == nil {
		t.Fatal("a rebind with a stale old path was accepted")
	}
	if c := rebindCard(t, db, "run-a", "t7"); c.WorktreePath != pOld {
		t.Errorf("a refused rebind moved the row: %q", c.WorktreePath)
	}
}

// TestRebindCardWorktree_TargetMustBeRegistered pins REQ-WHR-011(b): the
// registered-target check runs when the operation is called — a target that
// is not a registered worktree is refused; a registered target passes.
func TestRebindCardWorktree_TargetMustBeRegistered(t *testing.T) {
	db := rebindOpen(t)
	pOld := filepath.Join(t.TempDir(), "old-moai", "t7")
	pRegistered := "/repo/.moai/worktrees/t7-moved"
	pUnregistered := "/nowhere/t7-moved"
	rebindPick(t, db, "run-a", "t7", pOld)

	orig := cardWorktreeTargetProbe
	t.Cleanup(func() { cardWorktreeTargetProbe = orig })
	cardWorktreeTargetProbe = func(path string) (bool, error) {
		return path == pRegistered, nil
	}

	if _, err := db.RebindCardWorktree(context.Background(), "t7", pOld, pUnregistered, "rebind-test", rebindNow); err == nil {
		t.Error("a rebind to an unregistered path was accepted")
	}
	if _, err := db.RebindCardWorktree(context.Background(), "t7", pOld, pRegistered, "rebind-test", rebindNow); err != nil {
		t.Errorf("a rebind to the registered target was refused: %v", err)
	}
}

// TestRebindCardWorktree_ReverseRebindAfterReverseMove pins the symmetric
// shape the migration's reversal consumes: after old→new, a reverse rebind
// (old=new, new=old) succeeds when the reverse move re-registered the old
// path — the probe accepts both leaves.
func TestRebindCardWorktree_ReverseRebindAfterReverseMove(t *testing.T) {
	db := rebindOpen(t)
	pOld := filepath.Join(t.TempDir(), "old-moai", "t7")
	pNew := filepath.Join(t.TempDir(), "new-home", "t7")
	rebindPick(t, db, "run-a", "t7", pOld)

	orig := cardWorktreeTargetProbe
	t.Cleanup(func() { cardWorktreeTargetProbe = orig })
	cardWorktreeTargetProbe = func(path string) (bool, error) {
		return path == pOld || path == pNew, nil
	}

	if _, err := db.RebindCardWorktree(context.Background(), "t7", pOld, pNew, "rebind-test", rebindNow); err != nil {
		t.Fatalf("forward rebind: %v", err)
	}
	if _, err := db.RebindCardWorktree(context.Background(), "t7", pNew, pOld, "rebind-test", rebindNow); err != nil {
		t.Fatalf("reverse rebind after the reverse move: %v", err)
	}
	if c := rebindCard(t, db, "run-a", "t7"); c.WorktreePath != pOld {
		t.Errorf("after the reverse rebind the row carries %q, want %q", c.WorktreePath, pOld)
	}
}

// TestRecordCardWorktree_StillRefusesPathChange pins that the existing
// never-moves-a-card refusal is unchanged by the rebind operation's
// arrival (AC-WHR-011's last clause).
func TestRecordCardWorktree_StillRefusesPathChange(t *testing.T) {
	db := rebindOpen(t)
	pOld := filepath.Join(t.TempDir(), "old-moai", "t7")
	rebindPick(t, db, "run-a", "t7", pOld)

	if _, err := db.RecordCardWorktree(context.Background(), "run-a", "t7", "/elsewhere/t7", "rebind-test", rebindNow); err == nil {
		t.Error("RecordCardWorktree moved a card onto a different tree; the refusal must stand")
	}
}

// TestCardWorktreeTargetProbe_DefaultChecksGitEntry pins the default
// registered-target probe: a directory carrying a .git entry is registered;
// anything else is not.
func TestCardWorktreeTargetProbe_DefaultChecksGitEntry(t *testing.T) {
	registered := t.TempDir()
	if err := os.MkdirAll(filepath.Join(registered, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	plain := t.TempDir()

	if ok, err := defaultCardWorktreeTargetProbe(registered); err != nil || !ok {
		t.Errorf("default probe(%s) = (%v, %v), want registered", registered, ok, err)
	}
	if ok, err := defaultCardWorktreeTargetProbe(plain); err != nil || ok {
		t.Errorf("default probe(%s) = (%v, %v), want unregistered", plain, ok, err)
	}
}
