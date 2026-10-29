package e2e

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/intake"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/adapt"
)

// AC19 on lane-20's S51 read path over lane-18's real intake requests: the
// observation view shows each readiness stage and each non-stage state as its own
// state, never as running or done, and reading changes nothing.

var observationCfg = intake.Config{MaxInFlight: 8, MaxParked: 1, Slots: []string{"slot-1", "slot-2", "slot-3"}, StepTimeout: 2 * time.Second}

// observationRig builds nine requests, one per displayed state:
// ACCEPTED, ISSUE_CONFIRMED, BRANCH_REGISTERED, AGENT_RESERVED, WT_PREPARED,
// READY_ACKED, BLOCKED, EFFECT_UNKNOWN and PARKED.
func observationRig(t *testing.T) (*intakeRig, map[string]string) {
	t.Helper()
	r := newRig(t, observationCfg)
	ctx := context.Background()
	want := map[string]string{
		"s1-accepted": "ACCEPTED", "s2-issue": "ISSUE_CONFIRMED", "s3-branch": "BRANCH_REGISTERED",
		"s4-reserved": "AGENT_RESERVED", "s5-prepared": "WT_PREPARED", "s6-ready": "READY_ACKED",
		"s7-blocked": "BLOCKED", "s8-unknown": "EFFECT_UNKNOWN", "s9-parked": "PARKED",
	}
	order := []string{"s1-accepted", "s2-issue", "s3-branch", "s4-reserved", "s5-prepared", "s6-ready", "s7-blocked", "s8-unknown", "s9-parked"}
	for _, key := range order {
		if _, err := r.in.Accept(key, []byte(key), 1); err != nil {
			t.Fatalf("accept %s: %v", key, err)
		}
	}
	adv := func(key string, n int) {
		for range n {
			if _, err := r.in.Advance(ctx, key); err != nil {
				t.Fatalf("advance %s: %v", key, err)
			}
		}
	}
	adv("s2-issue", 1)
	adv("s3-branch", 2)
	for i, key := range []string{"s4-reserved", "s5-prepared", "s6-ready"} {
		adv(key, 2)
		agent := registerAgent(t, r.st, i+1)
		if _, err := r.in.Reserve(key, agent); err != nil {
			t.Fatalf("reserve %s: %v", key, err)
		}
		if key != "s4-reserved" {
			adv(key, 1) // prepare
		}
		if key == "s6-ready" {
			if _, err := r.in.AckReady(key, agent); err != nil {
				t.Fatal(err)
			}
		}
	}
	// BLOCKED: the request provably never left (retryable). EFFECT_UNKNOWN: it may have.
	r.pub.publishErr = errNotSent()
	_, _ = r.in.Advance(ctx, "s7-blocked")
	r.pub.publishErr = errors.New("connection reset after sending")
	_, _ = r.in.Advance(ctx, "s8-unknown")
	r.pub.publishErr = nil
	return r, want
}

func errNotSent() error { return errors.Join(intake.ErrNotSent, errors.New("refused before any call")) }

func readItems(t *testing.T, st *claim.Store) map[string]monitor.IntakeItem {
	t.Helper()
	items, err := adapt.RequestSource{Store: st}.ReadIntake(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]monitor.IntakeItem{}
	for _, it := range items {
		out[it.Key] = it
	}
	return out
}

// Each of the nine states is shown as its own, canonical name; nothing is shown
// as running or done; none has started.
func TestObservation_NineStatesAreDistinct_AndNeverRunningOrDone(t *testing.T) {
	r, want := observationRig(t)
	got := readItems(t, r.st)
	seen := map[string]string{}
	for key, wantState := range want {
		it, ok := got[key]
		if !ok {
			t.Errorf("%s is missing from the view", key)
			continue
		}
		if it.State != wantState {
			t.Errorf("%s shows %q, want %q", key, it.State, wantState)
		}
		if prev, dup := seen[it.State]; dup {
			t.Errorf("%s and %s are both shown as %q", prev, key, it.State)
		}
		seen[it.State] = key
		for _, bad := range []string{"BUILDING", "RUNNING", "DONE", "COMPLETE", "STARTED"} {
			if strings.EqualFold(it.State, bad) {
				t.Errorf("%s is shown as %s", key, it.State)
			}
		}
		if it.Started {
			t.Errorf("%s shows started with no attempt on record", key)
		}
	}
	if len(seen) != len(want) {
		t.Errorf("distinct states shown = %d, want %d", len(seen), len(want))
	}
}

// A saturated intake shows the waiting request as PARKED with its reason, not as accepted work.
func TestObservation_Saturation_ShowsParkedWithReason(t *testing.T) {
	r, _ := observationRig(t)
	it := readItems(t, r.st)["s9-parked"]
	if it.State != monitor.StateParked || it.Parked == "" {
		t.Errorf("parked request = state %q parked %q; want PARKED with a reason", it.State, it.Parked)
	}
	if it.State == string(monitor.StageAccepted) {
		t.Error("a request waiting for capacity is shown as ACCEPTED")
	}
}

// A blocked request carries its step, its detail and whether it can be retried;
// an unknown effect carries how to resume; neither looks like success.
func TestObservation_BlockedAndUnknown_ShowStepDetailAndResume(t *testing.T) {
	r, _ := observationRig(t)
	got := readItems(t, r.st)
	b := got["s7-blocked"]
	if b.Block == nil || b.Block.Step != monitor.StageAccepted || b.Block.Detail == "" || !b.Block.Retryable || b.Block.Unknown {
		t.Errorf("blocked row = %+v", b.Block)
	}
	u := got["s8-unknown"]
	if u.Block == nil || !u.Block.Unknown || u.Block.Resume == "" || u.Block.Retryable {
		t.Errorf("unknown-effect row = %+v", u.Block)
	}
	for _, it := range []monitor.IntakeItem{b, u} {
		if it.Stage != monitor.StageAccepted {
			t.Errorf("%s stage = %s; the failed step must stay visible as ACCEPTED", it.Key, it.Stage)
		}
	}
}

// Reading changes nothing: the stored requests are identical before and after.
func TestObservation_Reading_ChangesNothing(t *testing.T) {
	r, _ := observationRig(t)
	before, err := r.st.ListRequests()
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		_ = readItems(t, r.st)
	}
	after, err := r.st.ListRequests()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Error("reading the view changed the stored requests")
	}
}

// "Started" is read from the claim: once the reserved, acknowledging agent has an
// attempt, the view shows started while the stored stage stays READY_ACKED.
func TestObservation_Started_IsDerivedFromTheClaim_NotStored(t *testing.T) {
	r, _ := observationRig(t)
	view, err := r.in.View("s6-ready")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewIssueRef(view.Issue)
	if err != nil {
		t.Fatal(err)
	}
	agent := registerAgent(t, r.st, 3) // the agent s6-ready was reserved for (third registered)
	run, _ := core.ParseRunID("run-1")
	rn := wired(r.st)
	tk, err := rn.ClaimAs(ref, agent, "d-s6", run)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if _, err := rn.StartAttempt(tk, "att-1", core.DeliveryExecuted); err != nil {
		t.Fatalf("start: %v", err)
	}
	it := readItems(t, r.st)["s6-ready"]
	if !it.Started {
		t.Error("the attempt exists but the view does not show it as started")
	}
	if it.Stage != monitor.StageReadyAcked {
		t.Errorf("stored stage = %s, want READY_ACKED (started is not a stage)", it.Stage)
	}
	stored, _ := r.st.GetRequest("s6-ready")
	if stored.Started {
		t.Error("started was written into the stored request")
	}
}
