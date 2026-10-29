package e2e

import (
	"context"
	"errors"
	"testing"
)

const issue2 = "repo#1948"

func claimed(t *testing.T, d *ModelDriver) ClaimResult {
	t.Helper()
	r, err := d.Claim(context.Background(), issue2, "w1")
	if err != nil || !r.Won {
		t.Fatalf("setup claim: %+v %v", r, err)
	}
	return r
}

// Scenario 2: the same command id dispatched twice yields one send.
func TestScenario_DuplicateDispatch_SingleSend(t *testing.T) {
	d := NewModelDriver()
	claimed(t, d)
	for range 3 {
		if _, err := d.Dispatch(context.Background(), issue2, "cmd-1"); err != nil {
			t.Fatal(err)
		}
	}
	if n := d.SendCount(issue2, "cmd-1"); n != 1 {
		t.Errorf("send count = %d, want 1", n)
	}
}

// Scenario 3: ACK lost. Receiver already executed -> no resend.
// Receiver unqueryable -> blocked, still no resend.
func TestScenario_ACKLoss(t *testing.T) {
	cases := []struct {
		name       string
		probe      Probe
		wantResend bool
		wantBlock  string
		wantState  Delivery
	}{
		{"received", func(string) (Delivery, error) { return DeliveryReceived, nil }, false, "", DeliveryReceived},
		{"executed", func(string) (Delivery, error) { return DeliveryExecuted, nil }, false, "", DeliveryExecuted},
		{"not-received", func(string) (Delivery, error) { return DeliverySent, nil }, true, "", DeliverySent},
		{"unqueryable", func(string) (Delivery, error) { return "", errors.New("down") }, false, "connection-failed", DeliverySent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewModelDriver()
			claimed(t, d)
			if _, err := d.Dispatch(context.Background(), issue2, "cmd-1"); err != nil {
				t.Fatal(err)
			}
			out := d.OnAckTimeout(issue2, "cmd-1", tc.probe)
			if out.Resent != tc.wantResend {
				t.Errorf("resent = %v, want %v", out.Resent, tc.wantResend)
			}
			if out.BlockReason != tc.wantBlock {
				t.Errorf("block = %q, want %q", out.BlockReason, tc.wantBlock)
			}
			if out.State != tc.wantState {
				t.Errorf("state = %q, want %q", out.State, tc.wantState)
			}
			wantSends := 1
			if tc.wantResend {
				wantSends = 2
			}
			if n := d.SendCount(issue2, "cmd-1"); n != wantSends {
				t.Errorf("sends = %d, want %d", n, wantSends)
			}
		})
	}
}

// Scenario 4: interruption. A live or indeterminate owner is never replaced;
// a dead owner is replaced under the same claim with epoch+1.
func TestScenario_InterruptionRecovery(t *testing.T) {
	d := NewModelDriver()
	first := claimed(t, d)

	for _, liveness := range []Liveness{OwnerLive, OwnerIndeterminate} {
		r, err := d.Resume(context.Background(), issue2, liveness)
		if err != nil {
			t.Fatal(err)
		}
		if r.NewAttempt {
			t.Errorf("%v owner got a new attempt", liveness)
		}
	}
	r, err := d.Resume(context.Background(), issue2, OwnerDead)
	if err != nil || !r.NewAttempt {
		t.Fatalf("dead owner: %+v %v", r, err)
	}
	if r.AttemptID == first.AttemptID {
		t.Error("new attempt reused old id")
	}
	if r.Epoch != 2 {
		t.Errorf("epoch = %d, want 2", r.Epoch)
	}
	if n := d.LiveAttempts(issue2); n != 1 {
		t.Errorf("live attempts = %d, want 1", n)
	}
	if got := d.AttemptState(issue2, first.AttemptID); got != "superseded" {
		t.Errorf("old attempt = %q, want superseded", got)
	}
}

// Scenario 5: audit rounds past the cap block the Issue; ready is refused.
func TestScenario_AuditBudgetExhausted(t *testing.T) {
	d := NewModelDriver()
	claimed(t, d)
	for i := range 3 {
		if st := d.RecordAuditRound(issue2, 3, "FAIL"); i < 2 && st == "blocked" {
			t.Fatalf("blocked early at round %d", i+1)
		}
	}
	if st, reason := d.State(issue2); st != "blocked" || reason != "budget-exhausted" {
		t.Errorf("state = %s/%s, want blocked/budget-exhausted", st, reason)
	}
	if err := d.MarkReady(issue2); err == nil {
		t.Error("ready accepted while budget-exhausted")
	}
}

// Scenario 6: merge outcome unknown stays in landing and never reaches done;
// a failed remote lookup blocks.
func TestScenario_UnknownMergeResult(t *testing.T) {
	d := NewModelDriver()
	claimed(t, d)
	d.ForceState(issue2, "landing")

	if st := d.ObserveRemote(issue2, RemoteUnknown); st != "landing" {
		t.Errorf("unknown merge -> %s, want landing", st)
	}
	if st, _ := d.State(issue2); st == "done" {
		t.Error("done without remote evidence")
	}
	d.ObserveRemote(issue2, RemoteLookupFailed)
	if st, reason := d.State(issue2); st != "blocked" || reason != "connection-failed" {
		t.Errorf("lookup failure -> %s/%s", st, reason)
	}
}

// Scenario 7: done needs remote main + local sync + deploy record; the Issue
// closes only after done; the next dispatch then succeeds. (Whether the next
// dispatch may start earlier is not fixed by the contract, so it is not asserted.)
func TestScenario_ClosureThenNextDispatch(t *testing.T) {
	d := NewModelDriver()
	claimed(t, d)
	d.ForceState(issue2, "landing")

	if err := d.Complete(issue2, Completion{RemoteMain: true}); err == nil {
		t.Fatal("done accepted with remote only")
	}
	if d.IssueClosed(issue2) {
		t.Fatal("issue closed before done")
	}
	if err := d.Complete(issue2, Completion{RemoteMain: true, LocalSynced: true, DeployNA: "no deploy surface"}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if !d.IssueClosed(issue2) {
		t.Error("issue not closed after done")
	}
	if r, err := d.NextDispatch(context.Background(), "repo#1949", "w2"); err != nil || !r.Won {
		t.Errorf("next dispatch after done: %+v %v", r, err)
	}
}
