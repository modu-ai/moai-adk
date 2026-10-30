package ledger

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// A store whose database is gone refuses every operation with an error; none panics, none
// reports success, and none leaves the caller guessing which call failed.
func TestAClosedStoreRefusesEveryOperation(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	ev := Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "c1", Result: ResultPass}
	call := map[string]func() error{
		"AddCard": func() error {
			return s.AddCard(bg, Card{ID: "z", Title: "z", Position: core.Position{Stage: core.StageReady}, Spec: Spec{Class: "A", Paths: []string{"internal/z/x.go"}}})
		},
		"Cards":     func() error { _, err := s.Cards(bg); return err },
		"Leases":    func() error { _, err := s.Leases(bg); return err },
		"Claim":     func() error { _, err := s.Claim(bg, req("s", 1)); return err },
		"Reap":      func() error { _, err := s.Reap(bg, alwaysDead); return err },
		"Heartbeat": func() error { return s.Heartbeat(bg, "a", c.Epoch, nil) },
		"Release":   func() error { return s.Release(bg, "a", c.Epoch) },
		"ExpandKeys": func() error {
			return s.ExpandKeys(bg, "a", c.Epoch, []string{"pkg:internal/z"})
		},
		"AddEvidence": func() error { return s.AddEvidence(bg, "a", c.Epoch, ev) },
		"EvidenceFor": func() error { _, err := s.EvidenceFor(bg, "a"); return err },
		"AddNote":     func() error { return s.AddNote(bg, Note{Kind: NoteDecision, Text: "x"}) },
		"NotesSince":  func() error { _, err := s.NotesSince(bg, 0); return err },
		"Notes":       func() error { _, err := s.Notes(bg, 0); return err },
		"Move":        func() error { return s.Move(bg, "a", c.Epoch, posReview, "lane", "r", ev) },
		"Enqueue":     func() error { return s.Enqueue(bg, "a", c.Epoch, "c1") },
		"QueueUpdate": func() error {
			return s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", "")
		},
		"Queue": func() error { _, err := s.Queue(bg); return err },
		"AcquireLock": func() error {
			_, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "x"}, time.Minute)
			return err
		},
		"ReleaseLock": func() error { return s.ReleaseLock(bg, LockMerge, 1) },
		"RenewLock":   func() error { return s.RenewLock(bg, LockMerge, 1, time.Minute) },
		"Snapshot":    func() error { return s.Snapshot(bg, filepath.Join(t.TempDir(), "snap.db")) },
	}
	for name, fn := range call {
		if err := fn(); err == nil {
			t.Errorf("%s on a closed store reported success", name)
		}
	}
}

// A row the database cannot hand back is reported with what it was: a card whose spec is not
// JSON names the card and its epoch; a row of the wrong shape is the database's fault.
func TestAnUnreadableCardRowIsReportedByCardOrAsADatabaseFault(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET spec_json='{not json' WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	_, err := s.Cards(bg)
	var ce *cardError
	if !errors.As(err, &ce) || ce.id != "a" {
		t.Fatalf("a spec that is not JSON must come back as a card error naming the card: %v", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "card a epoch") || !strings.Contains(msg, "unreadable spec") {
		t.Errorf("message: %q", msg)
	}
	if errors.Unwrap(err) == nil {
		t.Error("the card error must wrap its cause")
	}

	if _, err := s.db.ExecContext(bg, `UPDATE cards SET spec_json='{}', epoch='not a number' WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	_, err = s.Cards(bg)
	var se *scanError
	if !errors.As(err, &se) {
		t.Fatalf("a row that does not scan is a database fault, not a card error: %v", err)
	}
	if !strings.HasPrefix(err.Error(), "ledger: ") || errors.Unwrap(err) == nil {
		t.Errorf("message %q, unwrap %v", err.Error(), errors.Unwrap(err))
	}
}

// A snapshot that cannot be written fails, leaves nothing half-made beside the target, and
// does not touch what was there.
func TestSnapshotRefusalsLeaveNothingBehind(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	dir := t.TempDir()

	file := filepath.Join(dir, "plain-file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Snapshot(bg, filepath.Join(file, "sub", "snap.db")); err == nil {
		t.Error("a snapshot below a plain file must fail")
	}

	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "keep"), []byte("kept"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Snapshot(bg, target); err == nil {
		t.Error("a snapshot onto a non-empty directory must fail")
	}
	if b, err := os.ReadFile(filepath.Join(target, "keep")); err != nil || string(b) != "kept" {
		t.Errorf("the target directory was disturbed: %q %v", b, err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("a temporary snapshot file was left behind: %s", e.Name())
		}
	}
}

// Open refuses a path it cannot use and a file that is not a ledger.
func TestOpenRefusals(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "plain-file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if s, err := Open(bg, filepath.Join(file, "sub", "ledger.db")); err == nil {
		_ = s.Close()
		t.Error("a ledger below a plain file must not open")
	}
	junk := filepath.Join(dir, "junk.db")
	if err := os.WriteFile(junk, []byte(strings.Repeat("this is not a sqlite database ", 200)), 0o600); err != nil {
		t.Fatal(err)
	}
	if s, err := Open(bg, junk); err == nil {
		_ = s.Close()
		t.Error("a file that is not a database must not open as a ledger")
	}
}
