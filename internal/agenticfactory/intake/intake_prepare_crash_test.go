package intake

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
)

// t2080 (follow-up of t2078): does prepare share the "effect sent, nothing recorded"
// gap of publish and register? Not in intake's design: intake treats preparing as
// repeatable (see the Preparer contract), and the slot is one the request already
// holds exclusively. A process that dies mid-prepare therefore restarts by
// preparing again, in the same slot, without the operator confirming anything.
// This pins intake's side of that contract; an intent-first marker like the one
// publish and register use would turn a repeatable step into one that needs
// ConfirmNotCreated. It does NOT show that a real preparer is repeatable: none
// exists in this repository yet, the fake here cannot tell a repeatable preparer
// from one that fails on an existing worktree, and that obligation sits on the
// implementer.
func TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot(t *testing.T) {
	r := newRig(t, Config{})
	reserved, _ := advanceTo(t, r, "crash-prep", claim.StageAgentReserved)

	r.pr.crash = true
	func() {
		defer func() { _ = recover() }() // the process died here
		_, _ = r.in.Advance(ctx(), "crash-prep")
	}()
	r.pr.crash = false

	got, err := r.in.Advance(ctx(), "crash-prep")
	if err != nil {
		t.Fatalf("restart after a crash mid-prepare needs no operator step: %v", err)
	}
	if got.Stage != claim.StageWTPrepared || got.Blocked != nil {
		t.Errorf("stage=%s blocked=%v, want %s with no block", got.Stage, got.Blocked, claim.StageWTPrepared)
	}
	if got.Ready == nil || got.Ready.Slot != reserved.Slot {
		t.Errorf("the restart must prepare the slot the request holds (%q): %+v", reserved.Slot, got.Ready)
	}
	if n := r.pr.calls.Load(); n != 2 {
		t.Errorf("PrepareWorktree calls = %d, want 2 (the crashed attempt and the restart)", n)
	}
	for _, s := range r.pr.slots {
		if s != reserved.Slot {
			t.Errorf("a prepare touched slot %q, the request holds %q", s, reserved.Slot)
		}
	}
}
