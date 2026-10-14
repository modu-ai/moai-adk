//go:build unix

package backup

// run_marker_fifo_unix_test.go — card t1619, codex review P2 (update backup run markers): a
// FIFO named .moai-run inside a backup directory must not hang the update's run collection.
// The marker read is non-blocking and bound to a regular file.

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestReadRunIDSkipsFIFOMarker(t *testing.T) {
	dir := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(dir, RunMarkerName), 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan bool, 1)
	go func() {
		_, ok := readRunID(dir)
		done <- ok
	}()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("a FIFO run marker must read as no marker, never as a run id")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("readRunID blocked on a FIFO run marker: the run collection must read non-blocking")
	}
}
