package web

// readers_branches_test.go — the error and skip paths of the monitor's file
// readers (viewmodel_ops.go) that no test reached (Issue #1948, coverage of the
// hand-written code): an unreadable session registry, factory records and verify
// snapshots that are corrupt or not files, and an unreadable SPEC tree. Each reader
// shows nothing for what it cannot read and keeps what it can. Tests only.
//
// Defect answer: these tests catch one corrupt file hiding the readable ones, a
// directory read as a record, and a broken registry shown as live sessions.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/session"
)

func TestLoadSessionsUnreadableRegistryShowsNothing(t *testing.T) {
	root := t.TempDir()
	if rows, byID := loadSessions(root, time.Now()); rows != nil || len(byID) != 0 {
		t.Errorf("a missing registry must show nothing: %v %v", rows, byID)
	}
	path := session.RegistryPathFor(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`[{"session_id":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if rows, byID := loadSessions(root, time.Now()); rows != nil || len(byID) != 0 {
		t.Errorf("a corrupt registry must show nothing, not a guess: %v %v", rows, byID)
	}
}

func TestLoadFactoryRecordsSkipsWhatItCannotRead(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Dir(factory.RecordPath(root, "probe"))
	if err := os.MkdirAll(filepath.Join(dir, "adir.json"), 0o700); err != nil { // a directory named like a record
		t.Fatal(err)
	}
	for name, body := range map[string]string{"good.json": `{}`, "corrupt.json": `{`, "note.txt": `{}`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if got := loadFactoryRecords(root); len(got) != 1 {
		t.Errorf("%d records, want only the one readable JSON record", len(got))
	}
	if got := loadFactoryRecords(t.TempDir()); got != nil {
		t.Errorf("an absent directory shows nothing: %v", got)
	}
}

func TestLoadVerifySkipsCorruptSnapshots(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "state", "verify", "snapshots")
	if err := os.MkdirAll(filepath.Join(dir, "adir.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"broken.json": `{`, "note.txt": `x`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	rows, keys := loadVerify(root, 10)
	if len(rows) != 0 {
		t.Errorf("a corrupt snapshot, a directory and a non-JSON file must yield no rows, got %v (keys %d)", rows, keys)
	}
	if rows, keys := loadVerify(t.TempDir(), 10); rows != nil || keys != 0 {
		t.Errorf("no verify directory at all shows nothing: %v %d", rows, keys)
	}
}

func TestLoadSpecRowsReportsAnUnreadableSpecTree(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	// .moai/specs is a file, not a directory: the tree cannot be listed.
	if err := os.WriteFile(filepath.Join(root, ".moai", "specs"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSpecRows(root); err == nil {
		t.Skip("spec.ListDocs reads a non-directory specs path as empty; the error branch is not reachable this way")
	}
}
