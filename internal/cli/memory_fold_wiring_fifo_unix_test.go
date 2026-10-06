//go:build unix

package cli

// memory_fold_wiring_fifo_unix_test.go — the blocked-read fixture of the
// AC-MFB-008 FIFO cells (vii), (viii), (ix) and the effective-scan windows
// of the gate/card-review regressions (plan M4, card t1502). Unix only; the
// windows twin keeps the package compiling (plan B1).

import (
	"errors"
	"os"
	"testing"
	"time"
)

// wireFIFO is one blocking-read fixture: the FIFO plus its silent writer.
type wireFIFO struct {
	writer *os.File
}

// release wakes a reader parked in a blocking read: one byte is written
// (darwin's poller does not reliably wake on a mere close) and the writer
// closes, so the reader sees its data and then EOF. Safe to call twice.
func (f *wireFIFO) release() {
	if f.writer == nil {
		return
	}
	_, _ = f.writer.Write([]byte("x"))
	_ = f.writer.Close()
	f.writer = nil
}

// blockOnRead replaces path with a named pipe and holds ONE writer open
// without ever writing, so a later reader parks in a blocking read until
// the writer releases. The t-taking wrapper registers the release as
// cleanup for the AC-MFB-008 cells.
func blockOnRead(t *testing.T, path string) *wireFIFO {
	t.Helper()
	fx, err := blockOnReadRaw(path)
	if err != nil {
		t.Fatalf("block-on-read fixture %s: %v", path, err)
	}
	t.Cleanup(fx.release)
	return fx
}

// blockOnReadRaw is the t-free form, callable from worker-goroutine seams
// (an orderProbe runs on the fold's worker, where testing.T is not
// available). The rendezvous: the write end is opened in a goroutine (its
// open stays pending), the read end is opened here, and the two pending
// opens complete each other. The read handle closes again, so the fold's
// own open later succeeds (a writer is present) and its READ is what
// blocks — a writerless FIFO read returns EOF at once under Go's poller
// on darwin.
func blockOnReadRaw(path string) (*wireFIFO, error) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := mkfifoForTest(path); err != nil {
		return nil, err
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
		return nil, err
	}
	var w *os.File
	select {
	case w = <-writerCh:
	case err := <-errCh:
		_ = r.Close()
		return nil, err
	case <-time.After(5 * time.Second):
		_ = r.Close()
		return nil, errors.New("timed out waiting for the fifo rendezvous")
	}
	_ = r.Close() // the fold's own open must be the blocking read
	return &wireFIFO{writer: w}, nil
}
