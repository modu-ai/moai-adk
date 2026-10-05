//go:build unix

package cli

// memory_fold_wiring_fifo_unix_test.go — the blocked-read fixture of the
// AC-MFB-008 FIFO cells (vii), (viii), (ix) (plan M4, card t1502). Unix
// only; the windows twin keeps the package compiling (plan B1).

import (
	"os"
	"testing"
	"time"
)

// wireFIFO is one blocking-read fixture: the FIFO plus its silent writer.
type wireFIFO struct {
	writer *os.File
}

// release closes the silent writer, so a reader parked in a blocking read
// returns and the abandoned step finishes.
func (f *wireFIFO) release() {
	_ = f.writer.Close()
}

// blockOnRead replaces path with a named pipe and holds ONE writer open
// without ever writing, so a later reader parks in a blocking read until
// the writer closes. Neither open blocks the test: the write end is opened
// in a goroutine (its open stays pending), the read end is opened in the
// test goroutine, and the two pending opens complete each other's
// rendezvous. The read handle is closed again immediately, so the fold's
// own open later succeeds (a writer is present) and its READ is what
// blocks — a writerless FIFO read would return EOF at once under Go's
// poller on darwin.
func blockOnRead(t *testing.T, path string) *wireFIFO {
	t.Helper()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Fatalf("clear %s: %v", path, err)
	}
	if err := mkfifoForTest(path); err != nil {
		t.Fatalf("mkfifo %s: %v", path, err)
	}
	writerCh := make(chan *os.File, 1)
	errCh := make(chan error, 1)
	go func() {
		w, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			errCh <- err
			return
		}
		writerCh <- w
	}()
	r, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open fifo read end %s: %v", path, err)
	}
	var w *os.File
	select {
	case w = <-writerCh:
	case err := <-errCh:
		_ = r.Close()
		t.Fatalf("hold fifo writer open %s: %v", path, err)
	case <-time.After(2 * time.Second):
		_ = r.Close()
		t.Fatalf("hold fifo writer open %s: timed out waiting for the rendezvous", path)
	}
	_ = r.Close() // the fold's own open must be the blocking read
	fx := &wireFIFO{writer: w}
	t.Cleanup(fx.release)
	return fx
}
