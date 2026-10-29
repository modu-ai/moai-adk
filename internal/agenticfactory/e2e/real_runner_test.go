package e2e

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// Scenarios 5-7 (real, lane-18 runner) on the minimum claimed -> done path.

const (
	headA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	headB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

// passAll is a complete, all-PASS evidence set for head. The writer handoff
// carries the holder and epoch the attempt was issued to (as the registering
// service observes them); the other producer-checked kinds are attributed by
// the Attributor the runner is wired with.
func passAll(tk runner.Ticket, head string) runner.Verification {
	v := runner.Verification{Head: head}
	for _, k := range []runner.Kind{
		runner.KindCheck, runner.KindReview, runner.KindAudit,
		runner.KindHandoffWriter, runner.KindHandoffCoordinator,
	} {
		res := runner.Result{Kind: k, Name: string(k), Head: head, Verdict: runner.VerdictPass}
		if k == runner.KindHandoffWriter {
			res.Session, res.Epoch = tk.Holder, tk.Epoch
		}
		v.Results = append(v.Results, res)
	}
	return v
}

// roleAttributor is a stand-in for the identity SPEC's attribution lookup that
// attributes every producer-checked result to exactly the role its kind needs.
// It exists to reach the paths behind the lock; the lock itself is tested
// with no attributor at all (real_role_attribution_test.go).
type roleAttributor struct{}

func (roleAttributor) Attribute(res runner.Result) runner.Attribution {
	role := map[runner.Kind]string{
		runner.KindReview: core.RoleReviewer, runner.KindAudit: core.RoleAuditor,
		runner.KindQA: core.RoleQA, runner.KindHandoffCoordinator: core.RoleCoordinator,
	}[res.Kind]
	return runner.Attribution{Attributed: true, Role: role, Session: "producer-" + string(res.Kind)}
}

// fakeFacts is a fake host for the permission decision: per-Issue facts, with an
// optional fallback for every Issue not listed (nil = the lookup fails).
type fakeFacts struct {
	byIssue  map[int]runner.IssueFacts
	fallback *runner.IssueFacts
	err      error
	mu       sync.Mutex
	calls    int
}

func (f *fakeFacts) IssueFacts(i core.IssueRef) (runner.IssueFacts, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.err != nil {
		return runner.IssueFacts{}, f.err
	}
	if v, ok := f.byIssue[i.Number]; ok {
		return v, nil
	}
	if f.fallback != nil {
		return *f.fallback, nil
	}
	return runner.IssueFacts{}, errors.New("issue lookup failed")
}

// ownIssueFacts: an Issue I opened, in the repository asked about, unowned.
func ownIssueFacts() runner.IssueFacts {
	return runner.IssueFacts{RequestedRepo: repoA, ActualRepo: repoA, UserID: me, AuthorID: me, Owner: core.OwnerNone}
}

// foreignIssueFacts: an Issue another author opened.
func foreignIssueFacts(owner core.OwnerState) runner.IssueFacts {
	return runner.IssueFacts{RequestedRepo: repoA, ActualRepo: repoA, UserID: me, AuthorID: them, Owner: owner}
}

// myIssuesFacts treats every Issue as mine, which is what the minimum path needs.
func myIssuesFacts() *fakeFacts { f := ownIssueFacts(); return &fakeFacts{fallback: &f} }

// wired is a Runner over st with the stand-in attributor and a host in which
// every Issue is mine.
func wired(st *claim.Store) *runner.Runner {
	r := runner.New(st)
	r.Attributor = roleAttributor{}
	r.Facts = myIssuesFacts()
	return r
}

func newRunner(t *testing.T) (*runner.Runner, runner.Ticket) {
	t.Helper()
	st, ref, run := realStore(t)
	r := wired(st)
	tk, err := r.Claim(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	return r, tk
}

func upTo(t *testing.T, r *runner.Runner, tk runner.Ticket, stage core.State) {
	t.Helper()
	steps := []func() error{
		func() error { _, err := r.StartAttempt(tk, "att-1", core.DeliveryExecuted); return err },
		func() error { _, err := r.Submit(tk, headA, false); return err },
		func() error { _, err := r.Verify(tk, passAll(tk, headA)); return err },
		func() error {
			// The PR text is checked for closing keywords before the PR exists.
			if _, err := r.PreflightPR(tk, headA, repoA, "Refs #"+issueNumber(tk)); err != nil {
				return err
			}
			_, err := r.Land(tk, headA)
			return err
		},
	}
	want := []core.State{core.StateImplementing, core.StateVerifying, core.StateReady, core.StateLanding}
	for i, s := range steps {
		if err := s(); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
		if want[i] == stage {
			return
		}
	}
	t.Fatalf("unknown stage %s", stage)
}

func issueNumber(tk runner.Ticket) string { return strconv.Itoa(tk.Issue.Number) }

func state(t *testing.T, r *runner.Runner, tk runner.Ticket) claim.Record {
	t.Helper()
	rec, err := r.Store.Load(tk.Issue)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// Scenario 5: audit budget exhausted -> blocked(budget-exhausted); the Issue
// cannot advance, and resumes only to the stage it left, once the condition is met.
func TestReal_AuditBudgetExhausted_BlocksAndResumesToReturnStage(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateVerifying)

	rec, err := r.Pause(tk, core.StateBlocked, core.ReasonBudgetExhausted, "operator approves a new audit round")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != core.StateBlocked || rec.Reason != core.ReasonBudgetExhausted || rec.Return != core.StateVerifying {
		t.Fatalf("pause record = %+v", rec)
	}
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("land while blocked: %v, want ErrTransitionInvalid", err)
	}
	if _, err := r.Verify(tk, passAll(tk, headA)); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("verify while blocked: %v, want ErrTransitionInvalid", err)
	}
	if _, err := r.Resume(tk, false); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("resume without condition: %v", err)
	}
	rec, err = r.Resume(tk, true)
	if err != nil || rec.State != core.StateVerifying {
		t.Errorf("resume = %+v, %v; want verifying", rec, err)
	}
}

// Same-HEAD binding: evidence for another HEAD never makes the candidate ready.
func TestReal_StaleHeadEvidence_ChangesNothing(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateVerifying)
	before := state(t, r, tk)
	if _, err := r.Verify(tk, passAll(tk, headB)); !errors.Is(err, core.ErrStaleHead) {
		t.Fatalf("stale evidence: %v, want ErrStaleHead", err)
	}
	if after := state(t, r, tk); !reflect.DeepEqual(after, before) {
		t.Errorf("record changed on refused verify: %+v -> %+v", before, after)
	}
}

// Scenario 6: merge result unknown -> Complete refused, state stays landing.
// A remote lookup that fails outright pauses as blocked(connection-failed).
func TestReal_UnknownMerge_StaysInLandingNeverDone(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)

	_, err := r.Complete(tk, runner.Landing{RemoteMerged: false, LocalSyncRecorded: true, Deploy: "N/A: no deploy surface"})
	if !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("complete with unknown merge: %v", err)
	}
	if got := state(t, r, tk).State; got != core.StateLanding {
		t.Fatalf("state = %s, want landing", got)
	}
	rec, err := r.Pause(tk, core.StateBlocked, core.ReasonConnectionFailed, "remote lookup succeeds")
	if err != nil || rec.Return != core.StateLanding {
		t.Errorf("pause on failed lookup = %+v, %v", rec, err)
	}
}

// Scenario 7: done needs remote merge + local sync + deploy record or a
// reasoned N/A; nothing after done may pause it.
func TestReal_Done_RequiresAllThreeRecords(t *testing.T) {
	cases := map[string]runner.Landing{
		"no-local-sync": {RemoteMerged: true, IssueClosed: true, Deploy: "N/A: none"},
		"no-deploy":     {RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true},
		"na-no-reason":  {RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A:  "},
		"bare-na":       {RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A"},
		"no-close":      {RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: none"},
	}
	for name, l := range cases {
		t.Run(name, func(t *testing.T) {
			r, tk := newRunner(t)
			upTo(t, r, tk, core.StateLanding)
			if _, err := r.Complete(tk, l); !errors.Is(err, core.ErrEvidenceIncomplete) {
				t.Errorf("complete(%+v) = %v, want ErrEvidenceIncomplete", l, err)
			}
			if got := state(t, r, tk).State; got != core.StateLanding {
				t.Errorf("state = %s, want landing", got)
			}
		})
	}

	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	rec, err := r.Complete(tk, fullLanding)
	if err != nil || rec.State != core.StateDone {
		t.Fatalf("complete = %+v, %v", rec, err)
	}
	if _, err := r.Pause(tk, core.StateOnHold, core.ReasonOperatorHold, "late hold"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("pause after done: %v, want ErrTransitionInvalid", err)
	}
}
