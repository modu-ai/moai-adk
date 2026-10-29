package e2e

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// The three lifecycle scenarios the operator fixed for the Agentic Factory
// (Issue #1948), on the real claim store, runner and notify service, with a fake
// host (no real GitHub, no network):
//
//	(a) a failing CI, the SAME agent repairs it, CI runs again, a review asks for
//	    changes, the same agent answers, the PR merges, the Issue is done and closed;
//	(b) the owner crashes: what it left is preserved and its writer is stopped
//	    before anyone else may take the Issue over, and the old owner stops writing;
//	(c) the owner's reply is lost: reminders go to the holder, and when the retries
//	    run out the escalation target is told once.
//
// They were written RED first and fixed behind the afred tag until lane-18 and
// lane-20 delivered ReopenOwn, SetWait/RetryWait, TakeOver, MarkRepairNeeded and
// the notify service; those are merged, so these run in the default suite.

const headC = "cccccccccccccccccccccccccccccccccccccccc"

// failingResults returns the evidence for head with one result of kind failed.
func failingResults(tk runner.Ticket, head string, kind runner.Kind) runner.Verification {
	v := passAll(tk, head)
	for i := range v.Results {
		if v.Results[i].Kind == kind {
			v.Results[i].Verdict = runner.VerdictFail
		}
	}
	return v
}

// sameOwner asserts the Issue is still held by tk's agent at tk's epoch.
func sameOwner(t *testing.T, rec claim.Record, tk runner.Ticket, step string) {
	t.Helper()
	if rec.Holder != tk.Holder || rec.Epoch != tk.Epoch {
		t.Fatalf("%s: holder %q epoch %d, want %q epoch %d (the agent must keep the Issue)", step, rec.Holder, rec.Epoch, tk.Holder, tk.Epoch)
	}
}

// repairCycle is one round of "the host reports a problem, the owner repairs it":
// the monitor records the signal, the owner reopens its own PR, the failing
// evidence of kind sends it back to implementing, it submits the new head, the
// evidence passes, and it publishes the new head with a fresh receipt.
func repairCycle(t *testing.T, r *runner.Runner, tk runner.Ticket, h runner.Handoff, from, to, why, receiptID string, failed runner.Kind) claim.Record {
	t.Helper()
	sig, err := r.Store.MarkRepairNeeded(tk.Issue, why)
	if err != nil {
		t.Fatalf("%s: signal: %v", why, err)
	}
	sameOwner(t, sig, tk, "after the repair signal")
	if sig.State != core.StateLanding || sig.RepairNeeded == "" {
		t.Fatalf("the signal must only mark the record: state %s, repair_needed %q", sig.State, sig.RepairNeeded)
	}
	re, err := r.ReopenOwn(tk, why, runner.RepairEvidence{ScopeReserved: true, OldWriterEnded: true})
	if err != nil {
		t.Fatalf("%s: reopen: %v", why, err)
	}
	sameOwner(t, re, tk, "after the owner reopened")
	if re.State != core.StateVerifying || re.RepairNeeded != "" || re.Repair == nil || re.Repair.Owner != tk.Holder {
		t.Fatalf("after reopen: state %s, repair_needed %q, repair %+v", re.State, re.RepairNeeded, re.Repair)
	}
	// Nobody else can take the repair while the owner holds the Issue.
	run, _ := core.ParseRunID("run-1")
	if _, err := r.ClaimRepair(tk.Issue, "intruder", "d-intruder-"+receiptID, run, h.Branch); !errors.Is(err, core.ErrHeld) {
		t.Fatalf("a second session's repair claim: %v, want ErrHeld", err)
	}
	if rec, err := r.Verify(tk, failingResults(tk, from, failed)); err != nil || rec.State != core.StateImplementing {
		t.Fatalf("the failing %s on %.12s: %+v, %v; want implementing", failed, from, rec, err)
	}
	if rec, err := r.Submit(tk, to, false); err != nil || rec.State != core.StateVerifying || string(rec.Head) != to {
		t.Fatalf("submit %.12s: %+v, %v", to, rec, err)
	}
	if rec, err := r.Verify(tk, passAll(tk, to)); err != nil || rec.State != core.StateReady {
		t.Fatalf("verify %.12s: %+v, %v; want ready", to, rec, err)
	}
	if _, err := r.PreflightPR(tk, to, repoA, "Refs #"+issueNumber(tk)); err != nil {
		t.Fatal(err)
	}
	if rec, err := r.Land(tk, to); err != nil || rec.State != core.StateLanding {
		t.Fatalf("land %.12s: %+v, %v", to, rec, err)
	}
	next := h
	next.HandoffID, next.Head, next.BoundaryHead = receiptID, to, to
	rec, err := r.HandOff(tk, next)
	if err != nil {
		t.Fatalf("receipt %s: %v", receiptID, err)
	}
	sameOwner(t, rec, tk, "after the new receipt")
	if rec.Repair != nil || rec.Receipt == nil || string(rec.Receipt.Head) != to {
		t.Fatalf("a new receipt must end the repair and carry the new head: repair %+v, receipt %+v", rec.Repair, rec.Receipt)
	}
	return rec
}

func countEvents(rec claim.Record, kind string) int {
	n := 0
	for _, e := range rec.History {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

// (a) A failing CI, the same agent repairs, CI runs again, a review asks for
// changes, the same agent answers, the PR merges, the Issue is done and closed.
func TestLifecycle_FailingCI_SameAgentRepairs_ReviewAnswered_MergedAndDone(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	h := goodHandoff(t)
	first, err := r.HandOff(tk, h)
	if err != nil {
		t.Fatal(err)
	}
	sameOwner(t, first, tk, "after the first receipt")

	// CI fails on the published head: the agent repairs it, CI passes on the new head.
	afterCI := repairCycle(t, r, tk, h, headA, headB, "required check failed on "+headA[:12], "handoff-2", runner.KindCheck)
	// A review asks for changes on that head: the same agent answers it.
	afterReview := repairCycle(t, r, tk, h, headB, headC, "changes were requested on PR #7", "handoff-3", runner.KindReview)

	// Three attempts, three receipts, two reopens: the history says so.
	if afterReview.AttemptSeq <= afterCI.AttemptSeq || afterCI.AttemptSeq <= first.AttemptSeq {
		t.Errorf("attempt numbers must only grow: %d -> %d -> %d", first.AttemptSeq, afterCI.AttemptSeq, afterReview.AttemptSeq)
	}
	if first.Attempt == afterCI.Attempt || afterCI.Attempt == afterReview.Attempt {
		t.Errorf("each repair is a new attempt id: %q %q %q", first.Attempt, afterCI.Attempt, afterReview.Attempt)
	}
	if n := countEvents(afterReview, "reopen"); n != 2 {
		t.Errorf("reopen events = %d, want 2", n)
	}
	if n := countEvents(afterReview, "handoff"); n != 3 {
		t.Errorf("handoff events = %d, want 3", n)
	}

	// The PR merges and the Issue is closed: done, with the close as part of the evidence.
	done, err := r.Complete(tk, fullLanding)
	if err != nil {
		t.Fatal(err)
	}
	if done.State != core.StateDone || !done.Merged || !done.Closed || done.Pending != "" {
		t.Errorf("final record = state %s merged %v closed %v pending %q; want done, merged, closed, nothing pending", done.State, done.Merged, done.Closed, done.Pending)
	}
	if done.Epoch != tk.Epoch {
		t.Errorf("epoch moved from %d to %d during a lifecycle that never changed owner", tk.Epoch, done.Epoch)
	}
}

// (b) The owner crashes. Nothing moves until the owner is known to be gone (or has
// left a stop record), its writer has stopped and what it left uncommitted is
// preserved; then the Issue changes hands once, and the old owner can no longer write.
func TestLifecycle_OwnerCrash_PreservedAndStopped_ThenSafeTakeover(t *testing.T) {
	setup := func(t *testing.T) (*runner.Runner, runner.Ticket, core.RunID) {
		t.Helper()
		r, tk := newRunner(t)
		upTo(t, r, tk, core.StateLanding)
		if _, err := r.HandOff(tk, goodHandoff(t)); err != nil {
			t.Fatal(err)
		}
		run, _ := core.ParseRunID("run-1")
		return r, tk, run
	}
	full := runner.Takeover{OwnerGone: true, WriterStopped: true, UncommittedPreserved: true, PreservedRef: "refs/heads/" + branch + "@preserve-1"}

	t.Run("short_of_any_evidence_nothing_moves", func(t *testing.T) {
		r, tk, run := setup(t)
		before := state(t, r, tk)
		for name, mod := range map[string]func(*runner.Takeover){
			"owner not gone and no stop record": func(e *runner.Takeover) { e.OwnerGone, e.SafeStop = false, "" },
			"writer not stopped":                func(e *runner.Takeover) { e.WriterStopped = false },
			"uncommitted not preserved":         func(e *runner.Takeover) { e.UncommittedPreserved = false },
			"no preserved reference":            func(e *runner.Takeover) { e.PreservedRef = "  " },
		} {
			ev := full
			mod(&ev)
			if _, err := r.TakeOver(tk.Issue, tk, "w2", "d-w2", run, ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
				t.Errorf("%s: %v, want ErrEvidenceIncomplete", name, err)
			}
		}
		// A decision taken about an owner that is no longer the recorded one is refused.
		stale := tk
		stale.Epoch += 7
		if _, err := r.TakeOver(tk.Issue, stale, "w2", "d-w2", run, full); !errors.Is(err, core.ErrOwnershipUnverified) {
			t.Errorf("stale epoch: %v, want ErrOwnershipUnverified", err)
		}
		if _, err := r.TakeOver(tk.Issue, tk, "w2", "d-1", run, full); !errors.Is(err, core.ErrDuplicateDispatch) {
			t.Errorf("the dispatch the owner already has: %v, want ErrDuplicateDispatch", err)
		}
		if after := state(t, r, tk); !reflect.DeepEqual(after, before) {
			t.Errorf("a refused takeover changed the record:\nbefore %+v\nafter  %+v", before, after)
		}
	})

	for name, ev := range map[string]runner.Takeover{
		"owner_observed_gone":                     full,
		"stop_record_stands_in_for_the_departure": {SafeStop: "stop-record-1", WriterStopped: true, UncommittedPreserved: true, PreservedRef: full.PreservedRef},
	} {
		t.Run(name, func(t *testing.T) {
			r, tk, run := setup(t)
			before := state(t, r, tk)
			rec, err := r.TakeOver(tk.Issue, tk, "w2", "d-w2", run, ev)
			if err != nil {
				t.Fatal(err)
			}
			if rec.Holder != "w2" || rec.Epoch != tk.Epoch+1 {
				t.Fatalf("holder %q epoch %d, want w2 at epoch %d", rec.Holder, rec.Epoch, tk.Epoch+1)
			}
			if rec.State != core.StateLanding || rec.Receipt == nil || rec.Receipt.Session != "w2" {
				t.Errorf("the PR and its receipt stay and the new holder becomes responsible for them: state %s, receipt %+v", rec.State, rec.Receipt)
			}
			if !strings.HasPrefix(string(rec.Attempt), "takeover-") || rec.AttemptSeq != before.AttemptSeq+1 {
				t.Errorf("a takeover is a new attempt: %q seq %d (was seq %d)", rec.Attempt, rec.AttemptSeq, before.AttemptSeq)
			}
			last := rec.History[len(rec.History)-1]
			if last.Kind != "takeover" || last.Session != "w2" || !strings.Contains(last.Note, "from w1") || !strings.Contains(last.Note, ev.PreservedRef) {
				t.Errorf("the takeover must be on the record with the preserved reference: %+v", last)
			}

			// The old owner's ticket no longer writes anything.
			beforeWrites := state(t, r, tk)
			if _, err := r.Complete(tk, fullLanding); !errors.Is(err, core.ErrOwnershipUnverified) {
				t.Errorf("the old owner completing: %v, want ErrOwnershipUnverified", err)
			}
			if _, err := r.SetWait(tk, waitOf(tk)); !errors.Is(err, core.ErrOwnershipUnverified) {
				t.Errorf("the old owner recording a wait: %v, want ErrOwnershipUnverified", err)
			}
			if _, err := r.ReopenOwn(tk, "late repair", runner.RepairEvidence{ScopeReserved: true, OldWriterEnded: true}); !errors.Is(err, core.ErrOwnershipUnverified) {
				t.Errorf("the old owner reopening: %v, want ErrOwnershipUnverified", err)
			}
			// A second takeover decided about the old owner cannot be applied again.
			if _, err := r.TakeOver(tk.Issue, tk, "w3", "d-w3", run, ev); !errors.Is(err, core.ErrOwnershipUnverified) {
				t.Errorf("a replayed takeover: %v, want ErrOwnershipUnverified", err)
			}
			if got := state(t, r, runner.Ticket{Issue: tk.Issue, Holder: "w2", Epoch: rec.Epoch}); !reflect.DeepEqual(got, beforeWrites) {
				t.Errorf("refused writes changed the record:\nbefore %+v\nafter  %+v", beforeWrites, got)
			}

			// The new owner finishes the merge.
			done, err := r.Complete(runner.Ticket{Issue: tk.Issue, Holder: "w2", Epoch: rec.Epoch}, fullLanding)
			if err != nil || done.State != core.StateDone || !done.Closed {
				t.Errorf("the new owner completing: %+v, %v", done, err)
			}
		})
	}
}

// recordingNotifier keeps every notice; fail makes it report non-delivery.
type recordingNotifier struct {
	mu      sync.Mutex
	notices []notify.Notice
	fail    bool
}

func (n *recordingNotifier) Notify(_ context.Context, x notify.Notice) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.notices = append(n.notices, x)
	if n.fail {
		return errors.New("channel down")
	}
	return nil
}

// recordingEscalator keeps every escalation; failFirst makes the first one fail.
type recordingEscalator struct {
	mu        sync.Mutex
	told      []notify.Escalation
	failFirst bool
	attempts  int
}

func (e *recordingEscalator) Escalate(_ context.Context, x notify.Escalation) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.attempts++
	if e.failFirst && e.attempts == 1 {
		return errors.New("escalation channel down")
	}
	e.told = append(e.told, x)
	return nil
}

// (c) The owner's reply is lost. The holder (and only the holder) is reminded; each
// reminder spends a retry; when the retries run out the escalation target is told
// exactly once and nobody retries past the limit.
func TestLifecycle_ReplyLost_Reminders_ThenEscalationOnce(t *testing.T) {
	const gap = 30 * time.Minute
	run := func(t *testing.T, n *recordingNotifier, esc *recordingEscalator) (*runner.Runner, runner.Ticket, *notify.Service, *time.Time) {
		t.Helper()
		r, tk := newRunner(t)
		upTo(t, r, tk, core.StateLanding)
		if _, err := r.HandOff(tk, goodHandoff(t)); err != nil {
			t.Fatal(err)
		}
		w := waitOf(tk)
		w.RecheckAt, w.MaxRetries, w.Escalate = t0.Add(10*time.Minute), 2, "leader"
		if _, err := r.SetWait(tk, w); err != nil {
			t.Fatal(err)
		}
		now := t0
		svc := &notify.Service{
			Records: r.Store, Waits: r, Notifier: n, Escalator: esc,
			State: docstore.CollectorCursors(r.Store), Now: func() time.Time { return now }, ReNotifyAfter: gap,
		}
		return r, tk, svc, &now
	}
	tick := func(t *testing.T, svc *notify.Service) []notify.Action {
		t.Helper()
		acts, err := svc.Tick(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return acts
	}
	count := func(acts []notify.Action, what string) int {
		n := 0
		for _, a := range acts {
			if a.What == what {
				n++
			}
		}
		return n
	}

	t.Run("lost_reply", func(t *testing.T) {
		n, esc := &recordingNotifier{}, &recordingEscalator{}
		r, tk, svc, now := run(t, n, esc)

		if acts := tick(t, svc); count(acts, "reminder") != 0 {
			t.Fatalf("before the recheck time nobody is reminded: %+v", acts)
		}
		*now = t0.Add(10 * time.Minute)
		if acts := tick(t, svc); count(acts, "reminder") != 1 || count(acts, "escalation") != 0 {
			t.Fatalf("first overdue pass: %+v, want one reminder and no escalation", acts)
		}
		if acts := tick(t, svc); len(acts) != 0 {
			t.Fatalf("a second pass at the same moment must do nothing: %+v", acts)
		}
		*now = now.Add(gap)
		acts := tick(t, svc)
		if count(acts, "reminder") != 1 || count(acts, "escalation") != 1 {
			t.Fatalf("the reminder that spends the last retry also escalates: %+v", acts)
		}
		*now = now.Add(gap)
		if acts := tick(t, svc); len(acts) != 0 {
			t.Fatalf("after the escalation nothing is sent again: %+v", acts)
		}

		if len(n.notices) != 2 {
			t.Fatalf("notices = %d, want 2 (one per retry)", len(n.notices))
		}
		for i, x := range n.notices {
			if x.Holder != tk.Holder || x.Kind != notify.WaitOverdue || x.Attempt != i+1 || x.Issue != tk.Issue.Number {
				t.Errorf("notice %d = %+v; want the holder %q, wait-overdue, attempt %d", i, x, tk.Holder, i+1)
			}
		}
		if len(esc.told) != 1 {
			t.Fatalf("escalations = %d, want exactly 1", len(esc.told))
		}
		if e := esc.told[0]; e.To != "leader" || e.Holder != tk.Holder || e.Retries != 2 || e.MaxRetries != 2 || e.Issue != tk.Issue.Number {
			t.Errorf("escalation = %+v", e)
		}
		rec := state(t, r, tk)
		sameOwner(t, rec, tk, "after the escalation")
		if rec.Wait == nil || !rec.Wait.Escalated || rec.Wait.Retries != 2 {
			t.Errorf("the wait must be recorded as escalated at 2 of 2: %+v", rec.Wait)
		}
		if rec.State != core.StateLanding {
			t.Errorf("the monitor must not move the Issue: state %s", rec.State)
		}
		// Nobody retries past the limit, the owner included: the wait, the owner and
		// the state stay as they were. (The record's revision number still moves on
		// such a call; that is lane-18's, pinned RED in real_lifecycle_afred_test.go.)
		before := state(t, r, tk)
		if _, err := r.RetryWait(tk, "again", "again", t0.Add(24*time.Hour)); !errors.Is(err, core.ErrRetryLimit) {
			t.Errorf("a retry past the limit: %v, want ErrRetryLimit", err)
		}
		after := state(t, r, tk)
		sameOwner(t, after, tk, "after a retry past the limit")
		if after.State != before.State || !reflect.DeepEqual(after.Wait, before.Wait) {
			t.Errorf("a retry past the limit changed the wait or the state:\nbefore %+v %+v\nafter  %+v %+v", before.State, before.Wait, after.State, after.Wait)
		}
	})

	// A channel known to be down still ends in an escalation instead of looping.
	t.Run("broken_channel_still_escalates", func(t *testing.T) {
		n, esc := &recordingNotifier{fail: true}, &recordingEscalator{}
		_, _, svc, now := run(t, n, esc)
		*now = t0.Add(10 * time.Minute)
		acts := tick(t, svc)
		*now = now.Add(gap)
		acts = append(acts, tick(t, svc)...)
		failed := 0
		for _, a := range acts {
			if a.What == "reminder" && a.Err != nil {
				failed++
			}
		}
		if failed != 2 || len(esc.told) != 1 {
			t.Errorf("reminders reported undelivered = %d (want 2), escalations = %d (want 1); actions %+v", failed, len(esc.told), acts)
		}
	})

	// An escalation that could not be delivered is tried again, then sent once.
	t.Run("escalation_retried_until_delivered_then_once", func(t *testing.T) {
		n, esc := &recordingNotifier{}, &recordingEscalator{failFirst: true}
		_, _, svc, now := run(t, n, esc)
		*now = t0.Add(10 * time.Minute)
		tick(t, svc)
		*now = now.Add(gap)
		acts := tick(t, svc)
		if count(acts, "escalation") != 1 || len(esc.told) != 0 {
			t.Fatalf("the first escalation attempt must fail and be reported: %+v (told %d)", acts, len(esc.told))
		}
		tick(t, svc)
		tick(t, svc)
		if len(esc.told) != 1 {
			t.Errorf("escalations delivered = %d, want exactly 1 (attempts %d)", len(esc.told), esc.attempts)
		}
	})
}
