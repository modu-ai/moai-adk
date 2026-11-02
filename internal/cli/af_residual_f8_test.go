package cli

// F8 (audit a1ec370195 onwards, residual classification): `moai af land --gate-timeout 0` (or a negative value) is not refused.
// The context of the gate has already ended when it starts, so the gate "fails" with "the gate ran past 0s" without having run,
// and the failure is recorded against the card's candidate. The expectation: a timeout that cannot let any gate run is refused
// up front, before the gate and before any write.

import (
	"context"
	"strings"
	"testing"
)

func TestAFResidual_F8AGateTimeoutThatLetsNoGateRunIsRefusedBeforeAnyWrite(t *testing.T) {
	afSkipWithoutSh(t)
	for _, timeout := range []string{"0", "-1s"} {
		t.Run(timeout, func(t *testing.T) {
			db, _, _, st, _ := afQueuedCard(t)
			out, errOut, err := afRunBoth("land", "t1", "--db", db, "--gate", "true", "--gate-timeout", timeout)
			rows, qerr := st.Queue(context.Background())
			if qerr != nil || len(rows) != 1 {
				t.Fatalf("queue: %+v %v", rows, qerr)
			}
			t.Logf("land --gate-timeout %s: err=%v stdout=%.240q row=%s", timeout, err, out+errOut, rows[0].State)
			if err == nil && strings.Contains(out, `"outcome":"failed"`) {
				t.Errorf("a gate that is `true` failed because its timeout was %s: %.300s", timeout, out)
			}
			if rows[0].State != "waiting" {
				t.Errorf("the row is %s after a landing with --gate-timeout %s, want waiting (refused, nothing changed)", rows[0].State, timeout)
			}
		})
	}
}
