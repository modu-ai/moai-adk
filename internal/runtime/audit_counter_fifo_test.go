package runtime

import (
	"path/filepath"
	goruntime "runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestCountAuditRoundsDoesNotBlockOnFifoEvidence is the card-review round-2
// P1 regression guard (card t1500): an evidence file that is a FIFO — or any
// non-regular file — must never block the counter. os.ReadFile on a FIFO
// blocks until a writer appears, which hung kickoff and card transitions.
// The counter reads evidence through the homestate readBoundedFile
// discipline instead: a regular-file check plus a size bound, refusing
// non-conforming files without opening them.
func TestCountAuditRoundsDoesNotBlockOnFifoEvidence(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("FIFOs are POSIX-only; the regular-file guard is source-level there")
	}
	cases := []struct {
		name     string
		filename string
	}{
		// The convention family reaches the body read inside the scan loop.
		{name: "convention family", filename: "plan-audit-iter2.md"},
		// A legacy-family FIFO is never body-read in the loop, but it is
		// recorded as LatestPath and parsed after the scan — the same hang,
		// reached through the back door.
		{name: "legacy family latest", filename: "SPEC-X-review-2.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fifo := filepath.Join(dir, tc.filename)
			if err := syscall.Mkfifo(fifo, 0o600); err != nil {
				t.Fatalf("mkfifo: %v", err)
			}
			done := make(chan error, 1)
			go func() {
				_, err := CountAuditRounds("SPEC-X", []string{dir})
				done <- err
			}()
			select {
			case err := <-done:
				if err == nil {
					t.Fatalf("CountAuditRounds accepted FIFO evidence %s without error", fifo)
				}
				if !strings.Contains(err.Error(), "regular file") {
					t.Fatalf("refusal does not name the regular-file rule: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("CountAuditRounds blocked on FIFO evidence for 5s (non-regular file opened, not rejected): %s", fifo)
			}
		})
	}
}
