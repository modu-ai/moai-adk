// m1_journal_gate_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M1: the
// journal.go critical-path gate (AC-023: journal.go @90%, unix execution).
// The M0 battery drove the happy paths and the refused-schema arm; these
// cases pin the ERROR arms the recovery contract depends on — a failed
// persist must surface as an error (the caller records a failure and the
// journal's on-disk state stays the last good write), never as silence.
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadJournalUnreadableTargetIsAnError — a journal path that exists but
// cannot be read as a file (a directory) must error, not read as absent:
// conflating it with absence would let a run proceed as if no recovery data
// existed.
func TestLoadJournalUnreadableTargetIsAnError(t *testing.T) {
	home := t.TempDir()
	jp := JournalPath(home)
	if err := os.MkdirAll(jp, 0o755); err != nil {
		t.Fatal(err)
	}
	j, err := LoadJournal(jp)
	if err == nil {
		t.Fatalf("a directory at the journal path loaded as %+v (want an error) — the read arm conflates unreadable with absent", j)
	}
	if j != nil {
		t.Fatalf("an error return carried a journal: %+v", j)
	}
	if strings.Contains(err.Error(), "absent") {
		t.Errorf("the error mislabels the failure as absence: %v", err)
	}
}

// TestClearJournalNonRemovableTargetIsAnError — ClearJournal must surface a
// failed removal (a non-empty directory occupies the path) instead of
// silently reporting success: the manifest-save atomicity contract clears
// the journal ONLY on the success path, and a silent no-op here would leave
// recovery data that reads as cleared.
func TestClearJournalNonRemovableTargetIsAnError(t *testing.T) {
	home := t.TempDir()
	jp := JournalPath(home)
	if err := os.MkdirAll(filepath.Join(jp, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ClearJournal(jp); err == nil {
		t.Fatal("ClearJournal reported success over a non-removable target")
	}
	// The absent path stays a clean no-op (the atomic-with-save contract).
	if err := ClearJournal(filepath.Join(home, "absent-journal.json")); err != nil {
		t.Fatalf("ClearJournal on an absent path errored: %v", err)
	}
}

// TestWriteJournalUnwritableHomeIsAnError — a journal persist into an
// unwritable home must error: the per-file flag persistence (M1,
// REQ-JRN-003) records a per-file failure from this error, and a silent
// success would leave the interruption contract claiming flags it never
// wrote.
func TestWriteJournalUnwritableHomeIsAnError(t *testing.T) {
	home := t.TempDir()
	moai := filepath.Join(home, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(moai, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(moai, 0o755) })

	err := WriteJournal(JournalPath(home), &PendingJournal{SchemaVersion: SchemaVersion})
	if err == nil {
		t.Fatal("WriteJournal reported success into an unwritable home")
	}
	if !strings.Contains(err.Error(), "journal") {
		t.Errorf("the error does not name the journal operation: %v", err)
	}
}

// TestWriteJournalZeroSchemaStamped — a zero-schema journal is stamped with
// the current SchemaVersion at write time (the writer never persists a 0).
func TestWriteJournalZeroSchemaStamped(t *testing.T) {
	home := t.TempDir()
	j := &PendingJournal{}
	if err := WriteJournal(JournalPath(home), j); err != nil {
		t.Fatalf("WriteJournal: %v", err)
	}
	loaded, err := LoadJournal(JournalPath(home))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if loaded.SchemaVersion != SchemaVersion {
		t.Fatalf("written schema_version = %d, want %d — the zero-schema stamp is broken", loaded.SchemaVersion, SchemaVersion)
	}
}

// TestWriteJournalFileAsHomeIsAnError — the journal home being a regular
// FILE makes the mkdir arm fail (ENOTDIR): the error must surface, never a
// silent success.
func TestWriteJournalFileAsHomeIsAnError(t *testing.T) {
	home := t.TempDir()
	fileHome := filepath.Join(home, "not-a-dir")
	if err := os.WriteFile(fileHome, []byte("a regular file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	jp := filepath.Join(fileHome, "journal.json")
	if err := WriteJournal(jp, &PendingJournal{SchemaVersion: SchemaVersion}); err == nil {
		t.Fatal("WriteJournal reported success with a file occupying the home path")
	}
}

// TestDirOfBareNameFallsBackToDot — dirOf on a separator-less path returns
// "." (the caller's MkdirAll then targets the working directory).
func TestDirOfBareNameFallsBackToDot(t *testing.T) {
	if got := dirOf("bare-journal.json"); got != "." {
		t.Fatalf("dirOf(bare name) = %q, want %q", got, ".")
	}
}
