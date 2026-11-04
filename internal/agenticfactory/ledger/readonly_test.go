package ledger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenReadOnly_MissingFileIsRefusedAndNothingCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "ledger.db")
	if st, err := OpenReadOnly(context.Background(), path); err == nil {
		_ = st.Close()
		t.Fatal("OpenReadOnly succeeded on a missing file")
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); !os.IsNotExist(err) {
		t.Errorf("OpenReadOnly created the directory (stat err: %v)", err)
	}
}

func TestOpenReadOnly_RefusesOtherSchemaVersionAndReadsCurrent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "ledger.db")
	st, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AddNote(ctx, Note{Kind: NoteDecision, Text: "n"}); err != nil {
		t.Fatal(err)
	}
	ro, err := OpenReadOnly(ctx, path)
	if err != nil {
		t.Fatalf("current schema refused: %v", err)
	}
	notes, err := ro.Notes(ctx, 0)
	if err != nil || len(notes) != 1 {
		t.Errorf("Notes = %v, %v; want one note", notes, err)
	}
	if err := ro.AddNote(ctx, Note{Kind: NoteDecision, Text: "x"}); err == nil {
		t.Error("write through a read-only store succeeded")
	}
	_ = ro.Close()
	for _, v := range []string{"1", "9"} {
		if _, err := st.db.ExecContext(ctx, "PRAGMA user_version="+v); err != nil {
			t.Fatal(err)
		}
		if r, err := OpenReadOnly(ctx, path); err == nil {
			_ = r.Close()
			t.Errorf("user_version %s was accepted", v)
		}
	}
	_ = st.Close()
}
