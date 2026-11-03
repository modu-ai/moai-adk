package claim

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func leftovers(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.Contains(d.Name(), ".tmp-") {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// A document write is acknowledged only after the file and its directory are
// synced; a failed sync is a failed PutDoc with nothing renamed into place.
func TestPutDocFailsWhenTheFileSyncFailsAndRenamesNothing(t *testing.T) {
	s := open(t)
	boom := errors.New("injected fsync failure")
	s.sync.file = func(*os.File) error { return boom }
	if _, err := s.PutDoc("intent", "abc", 0, []byte(`{"key":"abc"}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("PutDoc with a failing fsync: %v", err)
	}
	if _, err := os.Stat(s.docPath("intent", "abc")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the document must not exist after a failed sync: %v", err)
	}
	if l := leftovers(t, s.dir); len(l) != 0 {
		t.Errorf("temp files left: %v", l)
	}
	if _, err := s.GetDoc("intent", "abc"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("a reader sees a document that was never acknowledged: %v", err)
	}
}

func TestFailedSyncKeepsTheOlderDocumentIntact(t *testing.T) {
	s := open(t)
	if _, err := s.PutDoc("intent", "abc", 0, []byte(`{"v":1}`)); err != nil {
		t.Fatal(err)
	}
	s.sync.file = func(*os.File) error { return errors.New("injected") }
	if _, err := s.PutDoc("intent", "abc", 1, []byte(`{"v":2}`)); err == nil {
		t.Fatal("want an error")
	}
	s.sync = syncer{}
	d, err := s.GetDoc("intent", "abc")
	if err != nil || string(d.Data) != `{"v":1}` || d.Rev != 1 {
		t.Errorf("the acknowledged document changed: %+v %v", d, err)
	}
}

func TestPutDocFailsWhenTheDirectorySyncFails(t *testing.T) {
	s := open(t)
	var order []string
	s.sync.file = func(f *os.File) error { order = append(order, "file"); return f.Sync() }
	s.sync.dir = func(string) error { order = append(order, "dir"); return errors.New("injected dir fsync failure") }
	if _, err := s.PutDoc("intent", "abc", 0, []byte(`{}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("a failed directory sync must fail the write: %v", err)
	}
	if strings.Join(order, ",") != "file,dir" {
		t.Errorf("sync order = %v, want file then directory", order)
	}
}

func TestEveryRecordWriteSyncsFileThenDirectory(t *testing.T) {
	s := open(t)
	var order []string
	s.sync.file = func(f *os.File) error { order = append(order, "file"); return f.Sync() }
	s.sync.dir = func(p string) error { order = append(order, "dir"); return syncer{}.syncDir(p) }
	if _, err := s.Acquire(ref(t, 1), "w", "d", "run-1"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(order, ",") != "file,dir" {
		t.Errorf("record write sync order = %v", order)
	}
	s.sync.file = func(*os.File) error { return errors.New("injected") }
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("a record whose sync failed was reported as written: %v", err)
	}
	if _, err := os.Stat(s.path(ref(t, 2))); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("record file renamed despite the failed sync: %v", err)
	}
}
