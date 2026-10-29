package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// S46 (E13/E14, M34): the Issue close is part of the evidence for done, and a
// failed or missing close leaves done at zero. The merge fact that was observed
// is kept, and the open obligation is named. Promoted from the RED tests once
// lane-18 implemented the order.

var fullLanding = runner.Landing{
	RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: no deploy surface",
}

func TestRealClosure_MissingClose_NeverDone_MergeFactKept_DirectPath(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	l := fullLanding
	l.IssueClosed = false
	rec, err := r.Complete(tk, l)
	if !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("complete without a close: %v, want ErrEvidenceIncomplete", err)
	}
	if rec.State == core.StateDone {
		t.Fatal("done reached without an Issue close")
	}
	got := state(t, r, tk)
	if got.State != core.StateLanding || !got.Merged || got.Pending != "issue-close" {
		t.Errorf("record = state %s merged %v pending %q; want landing, merged, pending issue-close", got.State, got.Merged, got.Pending)
	}
}

func TestRealClosure_MissingClose_NeverDone_HandedOffPath(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	l := fullLanding
	l.IssueClosed = false
	if _, err := r.CompleteHandedOff(tk.Issue, l); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("handed-off complete without a close: %v", err)
	}
	if rec, err := r.Store.Read(tk.Issue); err != nil || rec.State == core.StateDone || rec.Pending != "issue-close" {
		t.Errorf("record after refusal = %+v, %v", rec, err)
	}
	// The close is then observed: the same call completes it.
	rec, err := r.CompleteHandedOff(tk.Issue, fullLanding)
	if err != nil || rec.State != core.StateDone || !rec.Closed || rec.Pending != "" {
		t.Errorf("done after the close: %+v, %v", rec, err)
	}
}

// Nothing is recorded as merged when the merge itself was not observed.
func TestRealClosure_UnobservedMerge_RecordsNothing(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	l := fullLanding
	l.RemoteMerged = false
	if _, err := r.Complete(tk, l); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("complete: %v", err)
	}
	if got := state(t, r, tk); got.Merged || got.Pending != "" {
		t.Errorf("unobserved merge was recorded: merged %v pending %q", got.Merged, got.Pending)
	}
}
