//go:build !windows

package userassets

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// Audit iter6 S1/M2: a FIFO where the copy should be is Unknown and must not
// hang the judgment (a blocking open of a FIFO waits for a writer that never
// comes).
func TestCheckProfileReach_FifoInPlaceOfTheCopyIsUnknownAndDoesNotBlock(t *testing.T) {
	home := t.TempDir()
	d := reachProfile(t, home, "alpha", nil)
	if err := os.MkdirAll(filepath.Join(d, "skills", "moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(d, "skills", filepath.FromSlash(reachRel)), 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	done := make(chan []ProfileReach, 1)
	go func() {
		rs, _ := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
		done <- rs
	}()
	select {
	case rs := <-done:
		if got := reachStates(rs); got["alpha"] != ReachUnknown {
			t.Fatalf("a FIFO is not the copy: want Unknown, got %+v", rs)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the judgment blocked on a FIFO")
	}
}

// The open path itself, reached when the entry was a regular file at Lstat time
// and a FIFO by the time it is opened: openRegularForRead must refuse it
// without blocking.
func TestOpenRegularForRead_FifoIsRefusedWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := syscall.Mkfifo(p, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	done := make(chan error, 1)
	go func() {
		f, err := openRegularForRead(root, "f")
		if f != nil {
			_ = f.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a FIFO must be refused")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("openRegularForRead blocked on a FIFO")
	}
}
