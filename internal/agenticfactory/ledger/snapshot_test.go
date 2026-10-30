//go:build afred

package ledger

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSnapshotWritesAConsistentCopyAtomically(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	dst := filepath.Join(t.TempDir(), "backup", "nested", "ledger.snap")
	if err := s.Snapshot(bg, dst); err != nil {
		t.Fatal(err)
	}
	snap, err := Open(bg, dst)
	if err != nil {
		t.Fatalf("the snapshot opens as a ledger: %v", err)
	}
	cs, _ := snap.Cards(bg)
	_ = snap.Close()
	if len(cs) != 1 || cs[0].ID != "a" {
		t.Errorf("snapshot cards: %+v", cs)
	}
	if st, err := os.Stat(dst); err != nil || st.Mode().Perm()&0o077 != 0 {
		t.Errorf("a snapshot is private to its owner: %v %v", st, err)
	}
	// a second snapshot replaces the first and leaves no temporary file
	addCard(t, s, "b", 1, "A", nil, "internal/b/x.go")
	if err := s.Snapshot(bg, dst); err != nil {
		t.Fatal(err)
	}
	snap, _ = Open(bg, dst)
	cs, _ = snap.Cards(bg)
	_ = snap.Close()
	if len(cs) != 2 {
		t.Errorf("the second snapshot: %+v", cs)
	}
	left, _ := filepath.Glob(filepath.Join(filepath.Dir(dst), "*"))
	for _, f := range left {
		if f != dst && f != dst+"-wal" && f != dst+"-shm" {
			t.Errorf("a leftover: %s", f)
		}
	}
}

func TestSnapshotWhileWritersWorkStaysIntact(t *testing.T) {
	s := openTemp(t)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = s.AddNote(bg, Note{Kind: NoteDecision, Text: "x"})
			}
		}
	}()
	dst := filepath.Join(t.TempDir(), "snap.db")
	for i := 0; i < 3; i++ {
		if err := s.Snapshot(bg, dst); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
	snap, err := Open(bg, dst)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Close()
	var ok string
	if err := snap.db.QueryRowContext(bg, `PRAGMA integrity_check`).Scan(&ok); err != nil || ok != "ok" {
		t.Errorf("integrity_check: %q %v", ok, err)
	}
}
