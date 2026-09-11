package kanban

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPureBacklogReaderPreservesSidecarInventory(t *testing.T) {
	dir := t.TempDir()
	s := NewBacklogStore(filepath.Join(dir, backlogFileName))
	if _, _, err := s.Add("card"); err != nil {
		t.Fatal(err)
	}
	before := dirCensus(t, dir)
	beforeDB, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadPure(); err != nil {
		t.Fatal(err)
	}
	afterDB, err := os.ReadFile(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeDB, afterDB) {
		t.Fatal("pure read changed database bytes")
	}
	after := dirCensus(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("pure reader sidecars changed: before=%v after=%v", before, after)
	}
}

func TestPureBacklogReaderWorksWithReadOnlyDatabase(t *testing.T) {
	s := NewBacklogStore(filepath.Join(t.TempDir(), backlogFileName))
	if _, _, err := s.Add("read only"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.EnginePath(), 0400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(s.EnginePath(), 0600) })
	r, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Items) != 1 {
		t.Fatalf("read-only card count=%d", len(r.Items))
	}
}

func TestPureBacklogReaderRejectsSQLWrites(t *testing.T) {
	s := NewBacklogStore(filepath.Join(t.TempDir(), backlogFileName))
	if _, _, err := s.Add("retained"); err != nil {
		t.Fatal(err)
	}
	reader, err := openBacklogReader(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer reader.close()
	if _, err := reader.db.Exec("DELETE FROM items"); err == nil {
		t.Fatal("query-only connection accepted SQL write")
	}
}

func TestPureBacklogReaderNeverCreatesMissingDB(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.db")
	if reader, err := openBacklogReader(p); err == nil {
		reader.close()
		t.Fatal("missing DB opened successfully")
	}
	if archiveTablesPresent(p) {
		t.Fatal("missing DB unexpectedly has archive")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("reader recreated missing database: %v", err)
	}
}

func TestPureBacklogReaderSeesCommittedWAL(t *testing.T) {
	s := NewBacklogStore(filepath.Join(t.TempDir(), backlogFileName))
	if _, _, err := s.Add("first"); err != nil {
		t.Fatal(err)
	}
	writer, err := openBacklogEngine(s.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.close()
	if _, err := writer.db.Exec("PRAGMA wal_autocheckpoint=0"); err != nil {
		t.Fatal(err)
	}
	rec, err := writer.readRecord(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rec.Items = append(rec.Items, BacklogItem{ID: "t2", Text: "committed WAL", State: BacklogStateQueued})
	rec.LastSeq = 2
	if err := writer.writeRecord(context.Background(), rec); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(s.EnginePath() + "-wal"); err != nil || info.Size() == 0 {
		t.Fatalf("missing nonempty WAL: %v", err)
	}
	got, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[1].Text != "committed WAL" {
		t.Fatalf("reader missed committed WAL: %+v", got)
	}
}
