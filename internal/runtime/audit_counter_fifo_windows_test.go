//go:build windows

package runtime

import "testing"

// TestCountAuditRoundsDoesNotBlockOnFifoEvidence is the windows twin of the
// POSIX regression guard (card t1563, REQ-CIS-002): FIFOs are POSIX-only, so
// the body of the guard cannot run here — but the test name stays in the
// swept set on every platform, so a reader of the windows run sees a SKIP
// rather than an absent test. The regular-file guard itself is enforced at
// the source level (homestate readBoundedFile) and is platform-independent.
func TestCountAuditRoundsDoesNotBlockOnFifoEvidence(t *testing.T) {
	t.Skip("FIFOs are POSIX-only; the regular-file guard is source-level there (see audit_counter_fifo_test.go on !windows)")
}
