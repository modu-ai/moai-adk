package intake

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
)

// t2080 (follow-up of t2078): does prepare share the "effect sent, nothing recorded"
// gap of publish and register? It does not, by design: PrepareWorktree builds a
// LOCAL worktree in a slot the request already holds exclusively, and preparing
// again keeps what the first attempt made. A process that dies mid-prepare must
// therefore restart by preparing again, in the same slot, without the operator
// having to confirm anything. This pins that behavior; an intent-first marker like
// publish and register use would be wrong here, because it would turn a repeatable
// step into one that needs ConfirmNotCreated.
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
