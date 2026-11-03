package cli

// factory_wait_ack_t2013_test.go — card t2013: `moai factory wait <card> --ack
// received|started` records the waker's acknowledgement and start of a wake the
// landing edge delivered. Each event is recorded once; an out-of-order event is
// refused with a message and writes nothing. The fixture is a real temp factory
// record; nothing outside it is touched.

import (
	"context"
	"strings"
	"testing"
	"time"
)

func ackFixture(t *testing.T, deliver bool) string {
	t.Helper()
	root := t.TempDir()
	if out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010"); err != nil {
		t.Fatalf("record wait: %v\n%s", err, out)
	}
	if deliver {
		settled, err := openWaitRoot(t, root).SettleLandedWaits(context.Background(), "t2010", time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC))
		if err != nil || len(settled) != 1 {
			t.Fatalf("settle = %+v, %v", settled, err)
		}
	}
	return root
}

func ackEventCount(t *testing.T, root string) int {
	t.Helper()
	var n int
	if err := openWaitRoot(t, root).DB.QueryRow(`SELECT COUNT(*) FROM events WHERE kind LIKE 'wake.%'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestFactoryWaitAckRecordsReceivedThenStartedOnce(t *testing.T) {
	root := ackFixture(t, true)

	out, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "received")
	if err != nil || !strings.Contains(out, "received recorded") {
		t.Fatalf("--ack received = %v / %q, want a recorded line", err, out)
	}
	st, err := openWaitRoot(t, root).LoadWakeState(context.Background(), 1)
	if err != nil || !st.Delivered || !st.Received || st.Started {
		t.Fatalf("state after the acknowledgement = %+v (%v), want delivered+received and NOT started", st, err)
	}
	out, err = runFactoryWaitCmd(t, root, "t2020", "--ack", "received")
	if err != nil || !strings.Contains(out, "already recorded") {
		t.Fatalf("duplicate --ack received = %v / %q, want already recorded", err, out)
	}
	if out, err = runFactoryWaitCmd(t, root, "t2020", "--ack", "started"); err != nil || !strings.Contains(out, "started recorded") {
		t.Fatalf("--ack started = %v / %q", err, out)
	}
	if n := ackEventCount(t, root); n != 3 {
		t.Fatalf("wake events = %d, want 3 (delivered, received, started once each)", n)
	}
}

func TestFactoryWaitAckRefusesOutOfOrderAndWritesNothing(t *testing.T) {
	root := ackFixture(t, true)
	out, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "started")
	if err == nil || !strings.Contains(err.Error()+out, "received") {
		t.Fatalf("--ack started before received = %v / %q, want a refusal naming received", err, out)
	}
	if n := ackEventCount(t, root); n != 1 {
		t.Fatalf("wake events = %d, want 1 (delivered only)", n)
	}

	undelivered := ackFixture(t, false)
	if out, err := runFactoryWaitCmd(t, undelivered, "t2020", "--ack", "received"); err == nil || !strings.Contains(err.Error()+out, "delivered") {
		t.Fatalf("--ack received before delivered = %v / %q, want a refusal naming delivered", err, out)
	}
	if n := ackEventCount(t, undelivered); n != 0 {
		t.Fatalf("wake events = %d, want 0", n)
	}
}

func TestFactoryWaitAckRefusesBadValueAndMixedFlags(t *testing.T) {
	root := ackFixture(t, true)
	if _, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "delivered"); err == nil {
		t.Errorf("--ack delivered must be refused (delivery is written by the landing edge)")
	}
	if _, err := runFactoryWaitCmd(t, root, "t2020", "--ack", "received", "--on", "t2010"); err == nil {
		t.Errorf("--ack with --on must be refused")
	}
	if n := ackEventCount(t, root); n != 1 {
		t.Fatalf("a refused call wrote: wake events = %d, want 1", n)
	}
}
