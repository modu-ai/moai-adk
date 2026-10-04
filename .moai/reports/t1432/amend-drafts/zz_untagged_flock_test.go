package harness

// DRAFT control (t1432 amendment 0.4.0): a test file that calls the POSIX-only syscall.Flock with NO
// //go:build !windows constraint. The Windows vet must reject it; this is the defect the heal-lock
// files must not have (REQ-HRH-012).

import (
	"os"
	"syscall"
	"testing"
)

func TestZZUntaggedFlock(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "x")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}
