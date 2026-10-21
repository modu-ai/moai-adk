package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationJournalReplacementPreservesPriorSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")
	prior := []byte(`{"steps":[{"outcome":"pending"}]}`)
	next := []byte(`{"steps":[{"outcome":"moved"}]}`)
	if err := journalWrite(path, prior); err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join(dir, "reader-snapshot.json")
	if err := os.Link(path, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := journalWrite(path, next); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(snapshot)
	if err != nil || string(got) != string(prior) {
		t.Fatalf("journal update changed the prior inode instead of atomically replacing it: %q, %v", got, err)
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != string(next) {
		t.Fatalf("replacement journal contents: %q, %v", got, err)
	}
}
