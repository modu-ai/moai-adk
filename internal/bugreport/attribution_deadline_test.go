package bugreport

import (
	"context"
	"fmt"
	"testing"
)

// TestAttributeHookTimeoutStaysAmbiguous pins the hook-timeout routing
// (review-gate P2): the register's verdict for hook_timeout is
// unconditional-ambiguous — "load and a stuck handler look alike without
// more data" — so the kind routes BEFORE the generic error-chain rows.
// context.DeadlineExceeded implements net.Error, so the A2 network row
// fired first and classified environment, which Capture then discards
// WITHOUT recording — the locally-retained entry the SPEC defines for
// hook_timeout (DEC-7) never existed.
func TestAttributeHookTimeoutStaysAmbiguous(t *testing.T) {
	verdict, _ := Attribute(KindHookTimeout, context.DeadlineExceeded, "")
	if verdict != VerdictAmbiguous {
		t.Fatalf("Attribute(hook_timeout, DeadlineExceeded) = %s, want ambiguous (the A2 net.Error row must not outrank the hook_timeout kind)", verdict)
	}

	// The wrapped shape the registry actually captures: the handler's
	// timeout-shaped error, whatever its wrapping.
	wrapped := fmt.Errorf("handler 2 for event session-start: %w", context.DeadlineExceeded)
	verdict, _ = Attribute(KindHookTimeout, wrapped, "")
	if verdict != VerdictAmbiguous {
		t.Fatalf("Attribute(hook_timeout, wrapped DeadlineExceeded) = %s, want ambiguous", verdict)
	}
}
