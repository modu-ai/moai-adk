//go:build afred

package e2e

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// RED on purpose (Issue #1948 S46, E13/E14, M34): the Issue close comes BEFORE
// done, and a failed close leaves done at zero. lane-18 still enforces the
// opposite order (MarkClosed requires done). These pass once lane-18 flips it.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_

func TestRED_CloseBeforeDone_HandedOff(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	land := runner.Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: no deploy surface"}
	if _, err := r.CompleteHandedOff(tk.Issue, land); err == nil {
		t.Error("done accepted before the Issue was closed")
	}
	if rec, err := r.MarkClosed(tk.Issue); err != nil || !rec.Closed {
		t.Fatalf("close before done refused: %+v, %v", rec, err)
	}
	if rec, err := r.CompleteHandedOff(tk.Issue, land); err != nil || rec.State != core.StateDone {
		t.Errorf("done after close: %+v, %v", rec, err)
	}
}

func TestRED_CloseBeforeDone_DirectPath(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	land := runner.Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: no deploy surface"}
	if _, err := r.Complete(tk, land); err == nil {
		t.Error("done accepted before the Issue was closed")
	}
	if got := state(t, r, tk).State; got == core.StateDone {
		t.Error("state is done without a close")
	}
}
