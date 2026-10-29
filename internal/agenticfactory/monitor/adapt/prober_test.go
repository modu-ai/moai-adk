package adapt

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/reconcile"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func sources(live reconcile.Liveness) Sources {
	return Sources{
		Holder:           func(int) (Holder, error) { return Holder{PID: 10, ProcessStart: "t", Worktree: "/w"}, nil },
		Classify:         func(int, string) reconcile.Liveness { return live },
		WorktreeHead:     func(string) (string, error) { return head, nil },
		RemoteHead:       func(context.Context, int) (string, error) { return head, nil },
		Delivery:         func(int) (effect.Stage, error) { return effect.Executed, nil },
		ActiveDispatches: func(int) ([]string, error) { return nil, nil },
		PR:               func(context.Context, int) (int, bool, error) { return 0, false, nil },
	}
}

func TestProbeCollectsActualState(t *testing.T) {
	act, err := sources(reconcile.Live).Probe(t.Context(), reconcile.Recorded{Issue: 7})
	if err != nil || act.Liveness != reconcile.Live || act.WorktreeHead != head || act.Delivery != effect.Executed {
		t.Fatalf("act = %+v err = %v", act, err)
	}
}

func TestProbeErrorMeansUnconfirmedPath(t *testing.T) {
	boom := errors.New("boom")
	for name, mut := range map[string]func(*Sources){
		"holder":   func(s *Sources) { s.Holder = func(int) (Holder, error) { return Holder{}, boom } },
		"delivery": func(s *Sources) { s.Delivery = func(int) (effect.Stage, error) { return "", boom } },
		"pr":       func(s *Sources) { s.PR = func(context.Context, int) (int, bool, error) { return 0, false, boom } },
		"active":   func(s *Sources) { s.ActiveDispatches = func(int) ([]string, error) { return nil, boom } },
	} {
		s := sources(reconcile.Live)
		mut(&s)
		if _, err := s.Probe(t.Context(), reconcile.Recorded{Issue: 7}); err == nil {
			t.Errorf("%s error swallowed", name)
		}
	}
}

// A leader restart: everything in memory is gone, only the claim store, the worktree and the remote
// remain. Reconciling twice from fresh objects restores the same view and changes nothing — no
// re-dispatch, no new attempt for a live holder, no write to the store.
func TestLeaderRestartRestoresFromActualAndNeverRedispatches(t *testing.T) {
	store, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	iss, _ := core.NewIssueRef(7)
	run, _ := core.ParseRunID("run-1")
	if _, err := store.Acquire(iss, "lane-18", "cmd-1", run); err != nil {
		t.Fatal(err)
	}
	before, _ := store.Load(iss)

	pass := func() reconcile.Report {
		src := sources(reconcile.Live) // fresh, no memory of the previous pass
		src.PR = func(context.Context, int) (int, bool, error) { return 55, false, nil }
		rec, err := store.Load(iss)
		if err != nil {
			t.Fatal(err)
		}
		r := RecordedFrom(rec, true, "")
		return reconcile.Reconcile(t.Context(), src, r)
	}
	first, second := pass(), pass()

	for _, rep := range []reconcile.Report{first, second} {
		for _, d := range rep.Divergences {
			if d.Action == reconcile.ActionNewAttempt {
				t.Fatalf("restart offered a new attempt for a live holder: %+v", d)
			}
		}
	}
	if len(first.Divergences) != len(second.Divergences) {
		t.Fatalf("restore not stable: %v vs %v", first.Divergences, second.Divergences)
	}
	foundPR := false
	for _, d := range first.Divergences {
		if d.Kind == reconcile.KindPRUnrecorded {
			foundPR = true
		}
	}
	if !foundPR {
		t.Errorf("open PR on the remote was not restored into the view: %v", first.Divergences)
	}
	after, _ := store.Load(iss)
	if !reflect.DeepEqual(after, before) {
		t.Errorf("claim changed across reconcile: %+v -> %+v", before, after)
	}
}

func TestLiveHolderWithExpiredLeaseIsNotReissuedAfterRestart(t *testing.T) {
	rep := reconcile.Reconcile(t.Context(), sources(reconcile.Live), reconcile.Recorded{Issue: 7, HeadSHA: head, LeaseExpired: true, Delivery: effect.Executed})
	for _, d := range rep.Divergences {
		if d.Action == reconcile.ActionNewAttempt {
			t.Fatal("re-issue offered for a live holder")
		}
	}
	_ = claim.Record{}
}

func TestClassifierFromHomestateMapsThreeStates(t *testing.T) {
	live := ClassifierFromHomestate(func(int) (string, homestate.ProcessIdentityState) { return "t", homestate.ProcessIdentityLive })
	if got := live(10, "t"); got != reconcile.Live {
		t.Errorf("same fingerprint = %s, want live", got)
	}
	if got := live(10, "other"); got != reconcile.Dead {
		t.Errorf("reused pid with other fingerprint = %s, want dead", got)
	}
	dead := ClassifierFromHomestate(func(int) (string, homestate.ProcessIdentityState) { return "", homestate.ProcessIdentityDead })
	if got := dead(10, "t"); got != reconcile.Dead {
		t.Errorf("dead pid = %s", got)
	}
	if got := live(0, "t"); got != reconcile.Indeterminate {
		t.Errorf("no pid = %s, want indeterminate", got)
	}
}
