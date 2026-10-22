package factory

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

func TestRestoreGTDSnapshotIsPrivateBeforeSQLiteReopens(t *testing.T) {
	dir := privateBacklogFixtureDir(t)
	destination := filepath.Join(dir, "backlog.db")
	wholeFile := []byte("synthetic full-file snapshot\x00\x01\xff")
	if err := restoreGTDSnapshot(t.Context(), bytes.NewReader(wholeFile), destination); err != nil {
		t.Fatal(err)
	}
	// Inspect publication before openGTDDB can validate or narrow an existing file.
	// Windows uses the actual SID/DACL; Unix retains the exact 0600 oracle.
	if err := privatefile.Check(destination); err != nil {
		t.Fatalf("published snapshot native policy: %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil || !bytes.Equal(got, wholeFile) {
		t.Fatalf("whole-file bytes not preserved: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "backlog.db" {
		t.Fatalf("publication left staging artifacts: %v", err)
	}
}
