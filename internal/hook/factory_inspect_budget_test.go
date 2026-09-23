package hook

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestFactoryHookBatchDeadlineIsNotReportedAsUnbound reproduces the t1133
// defect: when the 200ms inspection budget is exhausted before the peer lookup
// completes — the shape a loaded host produces — factoryHookBatch reports the
// bound session as "unbound-session". That state is indistinguishable from a
// genuinely unregistered session, so the inbox reads as empty and no degraded
// signal reaches the caller.
func TestFactoryHookBatchDeadlineIsNotReportedAsUnbound(t *testing.T) {
	_, _, _, p, in := factoryHookFixture(t)

	// Sanity: with an unexhausted budget the bound session resolves.
	if _, _, state := factoryHookBatch(context.Background(), in, EventStop); state == "unbound-session" {
		t.Fatalf("fixture peer %s is not bound; got state=%s", p.SessionUUID, state)
	}

	// The budget is already spent when the hook runs — the shape contention on
	// a loaded host produces partway through the batch.
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)

	msg, cont, state := factoryHookBatch(ctx, in, EventStop)
	t.Logf("msg=%q cont=%v state=%s", msg, cont, state)

	if state == "unbound-session" {
		t.Fatalf("budget exhaustion reported as unbound-session: a bound peer reads as unregistered and the empty inbox carries no degraded signal (msg=%q cont=%v)", msg, cont)
	}
	if !strings.HasPrefix(state, "degraded:") {
		t.Fatalf("budget exhaustion must surface as a degraded state, got %q", state)
	}
}
