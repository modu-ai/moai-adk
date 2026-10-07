//go:build unix

// The consent reader's blocking-file fixture is unix-only (FIFO semantics;
// the B1 split rule — a build tag, not a runtime skip). The capture path
// already covers the consent read with its own time box; the DRAIN and the
// SENDER call ReadUserParticipation directly, so the reader itself must not
// block past any deadline (review gate finding, P2): a FIFO at the consent
// path used to park a plain os.ReadFile — open, then read — indefinitely.

package config

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// fifoConsentFixture creates the consent file's path as a FIFO under a
// temporary MOAI_HOME and returns the path.
func fifoConsentFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	consentDir := filepath.Join(home, "config")
	if err := os.MkdirAll(consentDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	fifoPath := filepath.Join(consentDir, "participation.yaml")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}
	return fifoPath
}

// holdFifoWriter holds a writer on the FIFO open without writing — the
// reader's open succeeds but its read blocks forever. FIFO open ordering:
// a blocking O_WRONLY open parks until a reader exists, so the fixture
// opens the read end NON-BLOCKING first, lets the writer's open succeed,
// then closes its own descriptor. Cleanup-guaranteed: the writer is closed
// when the test ends so the leaked reader goroutine (on the pre-fix tree)
// can exit.
func holdFifoWriter(t *testing.T, fifoPath string) {
	t.Helper()
	r, err := os.OpenFile(fifoPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}
	opened := make(chan struct{})
	go func() {
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
		if err != nil {
			close(opened)
			return
		}
		defer func() { _ = f.Close() }()
		close(opened)
		time.Sleep(10 * time.Second)
	}()
	t.Cleanup(func() {
		select {
		case <-opened:
		case <-time.After(100 * time.Millisecond):
		}
	})
	select {
	case <-opened:
	case <-time.After(2 * time.Second):
		t.Skip("writer could not open the fifo")
	}
	_ = r.Close()
}

// TestConsentReaderRejectsFifoWithoutBlocking: the reader must refuse a
// non-regular consent file WITHOUT opening it — a plain read parked on the
// FIFO until an outside writer released it, stalling the drain and the
// sender past their own time boxes.
func TestConsentReaderRejectsFifoWithoutBlocking(t *testing.T) {
	fifoPath := fifoConsentFixture(t)
	holdFifoWriter(t, fifoPath)

	start := time.Now()
	up := ReadUserParticipation()
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Fatalf("the consent reader blocked %s on a FIFO consent file — a non-regular file must be refused without opening it", elapsed)
	}
	if up.Enabled || up.Asked || up.Repository != "" {
		t.Fatalf("a FIFO consent file read as %+v — fail-closed means the zero consent", up)
	}
}
