package cli

// factory_wait_ack_target_t2013_test.go — sync audit 6 finding C6: --ack must
// record the wake of the wait that was actually delivered and is still
// unacknowledged (the oldest such wait), not blindly the card's latest wait.
// Two waits on one card, only the earlier one delivered. Reuses the real-verb
// fixture of factory_wait_ack_t2013_test.go.

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryWaitAckTargetsTheDeliveredWait(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	for _, pred := range []string{"t2010", "t2011"} {
		if out, err := runFactoryWaitCmd(t, root, "t2020", "--on", pred, "--waker", "lane-3", "--reason", "needs "+pred); err != nil {
			t.Fatalf("record wait on %s: %v\n%s", pred, err, out)
		}
	}
	db := openWaitRoot(t, root)
	waits, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(waits) != 2 {
		t.Fatalf("want two waits, got %+v err=%v", waits, err)
	}
	first, second := waits[0], waits[1]
	landed := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	state := func(seq int64) homestate.WakeState {
		t.Helper()
		st, err := db.LoadWakeState(ctx, seq)
		if err != nil {
			t.Fatal(err)
		}
		return st
	}

	// Only t2010 lands: the FIRST wait is delivered, the second is not.
	if settled, err := db.SettleLandedWaits(ctx, "t2010", landed); err != nil || len(settled) != 1 {
		t.Fatalf("settle t2010 = %+v, %v", settled, err)
	}
	out, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "received")
	if err != nil {
		t.Fatalf("--ack received with the earlier wait delivered: %v\n%s", err, out)
	}
	if want := fmt.Sprintf("wait=%d", first.Seq); !strings.Contains(out, want) {
		t.Errorf("ack output %q does not name %s", out, want)
	}
	if st := state(first.Seq); !st.Received {
		t.Errorf("first wait %d state = %+v, want received recorded", first.Seq, st)
	}
	if st := state(second.Seq); st.Received || st.Delivered {
		t.Errorf("second (undelivered) wait %d state = %+v, want nothing recorded", second.Seq, st)
	}

	// started follows on the same wait.
	if out, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "started"); err != nil {
		t.Fatalf("--ack started: %v\n%s", err, out)
	}
	if st := state(first.Seq); !st.Started {
		t.Errorf("first wait %d state = %+v, want started recorded", first.Seq, st)
	}

	// t2011 lands too: the next received ack goes to the second wait.
	if settled, err := db.SettleLandedWaits(ctx, "t2011", landed); err != nil || len(settled) != 1 {
		t.Fatalf("settle t2011 = %+v, %v", settled, err)
	}
	if out, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "received"); err != nil {
		t.Fatalf("--ack received for the second wait: %v\n%s", err, out)
	}
	if st := state(second.Seq); !st.Received {
		t.Errorf("second wait %d state = %+v, want received recorded", second.Seq, st)
	}
}
