package bugreport

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestCaptureReturnsWhenConsentPathBlocks is review-gate finding #3's pin:
// the 50 ms fail-deadline must cover the WHOLE capture path, including the
// consent read that opens this test's FIFO at the consent file's path. The
// fixture is the established environment-scrub pattern: a real FIFO under a
// temporary MOAI_HOME, a writer that never writes. Capture must abandon
// within the box; before the fix the consent read sat OUTSIDE the deadline
// and Capture blocked indefinitely — which, on the hook and panic paths, is
// the host stall the fail-open contract exists to prevent.
func TestCaptureReturnsWhenConsentPathBlocks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")

	consentDir := filepath.Join(home, "config")
	if err := os.MkdirAll(consentDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	fifoPath := filepath.Join(consentDir, "participation.yaml")
	if err := syscall.Mkfifo(fifoPath, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	// FIFO open ordering: a blocking O_WRONLY open parks until a reader
	// exists, and the test has no reader of its own — so the TEST opens the
	// read end non-blocking FIRST (immediate), lets the writer's blocking
	// open succeed, then closes its own descriptor. From there the writer
	// holds the FIFO's write end open without ever writing, and Capture's
	// consent read opens fine and blocks on the read — the exact stall the
	// box must abandon.
	r, err := os.OpenFile(fifoPath, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("open fifo read end: %v", err)
	}

	// Hold a writer open (so the FIFO's write end stays held) but never
	// write — the read inside Capture blocks until the deadline.
	blocked := make(chan struct{})
	go func() {
		f, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
		if err != nil {
			close(blocked)
			return
		}
		defer func() { _ = f.Close() }()
		close(blocked)
		time.Sleep(10 * time.Second)
	}()
	// Drain the fixture goroutine when the test ends; a writer still
	// holding the FIFO at process exit is fine (temp dir dies with it).
	defer func() {
		select {
		case <-blocked:
		case <-time.After(100 * time.Millisecond):
		}
	}()
	select {
	case <-blocked:
	case <-time.After(2 * time.Second):
		t.Skip("writer could not open the fifo")
	}
	_ = r.Close()

	start := time.Now()
	Capture(KindHookTimeout, errStringForDeadline("deadline"), "", nil)
	elapsed := time.Since(start)
	if elapsed > 250*time.Millisecond {
		t.Fatalf("capture blocked %s on the consent read, want it to abandon within the box", elapsed)
	}
}

// errStringForDeadline is this file's error carrier (the shaped-error
// helpers live in the sibling test files).
type errStringForDeadline string

func (e errStringForDeadline) Error() string { return string(e) }
