package claim

import (
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// t2072 (RC40 carried-over review finding, claim/request.go:193): the request
// file was renamed into place without an fsync of the temp file or the
// directory, while every other record goes through writeAtomic. A crash could
// leave an acknowledged request as an empty file. The injectable syncer makes the
// missing steps observable.
func TestWriteRequestLocalSyncsFileAndDirectoryAndRefusesOnFailure(t *testing.T) {
	s := open(t)
	if err := os.MkdirAll(s.requestsDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	var files, dirs int
	s.sync = syncer{
		file: func(f *os.File) error { files++; return f.Sync() },
		dir:  func(string) error { dirs++; return nil },
	}
	if err := s.writeRequestLocal(Request{Key: "k-durable"}); err != nil {
		t.Fatal(err)
	}
	if files != 1 || dirs != 1 {
		t.Errorf("fsync of file=%d dir=%d, want 1 and 1", files, dirs)
	}

	// a failed fsync must not be acknowledged, and must leave no temp file behind
	s.sync = syncer{file: func(*os.File) error { return errors.New("disk full") }}
	err := s.writeRequestLocal(Request{Key: "k-failed"})
	if !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("a failed fsync was acknowledged: %v", err)
	}
	if _, serr := os.Stat(s.requestPath("k-failed")); !errors.Is(serr, os.ErrNotExist) {
		t.Errorf("a request whose fsync failed is on disk: %v", serr)
	}
	ents, _ := os.ReadDir(s.requestsDir())
	for _, e := range ents {
		if e.Name() != "k-durable.json" {
			t.Errorf("leftover file after a failed write: %s", e.Name())
		}
	}
}
