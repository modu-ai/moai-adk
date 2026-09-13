//go:build !windows

package codexapp

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestProfileLeaseIsCloseOnExec(t *testing.T) {
	file, err := profileLease(filepath.Join(t.TempDir(), "lease"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close lease: %v", err)
		}
	}()
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, file.Fd(), syscall.F_GETFD, 0)
	if errno != 0 {
		t.Fatal(errno)
	}
	if flags&syscall.FD_CLOEXEC == 0 {
		t.Fatal("profile lease descriptor leaks across exec")
	}
}
