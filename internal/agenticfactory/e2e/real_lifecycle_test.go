package e2e

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// The three lifecycle scenarios the operator fixed for the Agentic Factory
// (Issue #1948), on the real claim store and runner, with a fake
// host (no real GitHub, no network):
//
//	(a) a failing CI, the SAME agent repairs it, CI runs again, a review asks for
//	    changes, the same agent answers, the PR merges, the Issue is done and closed;
//	(b) the owner crashes: what it left is preserved and its writer is stopped
//	    before anyone else may take the Issue over, and the old owner stops writing;
//	(c) the reply-lost reminder scenario retired with the notify service (local-only redesign).
//
// They were written RED first and fixed behind the afred tag until lane-18 and
// lane-20 delivered ReopenOwn, SetWait/RetryWait, TakeOver, MarkRepairNeeded and
// the notify service; those are merged, so these run in the default suite.

// t0 is a fixed test instant.
var t0 = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

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
		// An invented stop reference stands in for nothing (behavior change: the reference is checked
		// against the recorded stop records).
		invented := runner.Takeover{SafeStop: "stop-record-1", WriterStopped: true, UncommittedPreserved: true, PreservedRef: full.PreservedRef}
		if _, err := r.TakeOver(tk.Issue, tk, "w2", "d-w2", run, invented); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("an invented stop reference: %v, want ErrEvidenceIncomplete", err)
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
			if ev.SafeStop != "" {
				// A stop record must exist and speak for this Issue, holder and epoch (lane-18 F5): the
				// owner leaves one on purpose, so the test has the owner leave it.
				ref, serr := r.Store.RecordSafeStop(tk.Issue, tk.Holder, tk.Epoch, "stopped on purpose")
				if serr != nil {
					t.Fatal(serr)
				}
				ev.SafeStop = ref
			}
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
			// Since lane-18 52725aab26 (TakeOver TOCTOU) the ticket's ownership is checked before the safe-stop
			// record's freshness, so a replay is refused as an unverified owner whether or not it rests on a stop
			// record. (Between 4381200572 and that fix the stop-record case ended as ErrEvidenceIncomplete.)
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
