package cli

// factory_complete_s03_test.go — card t2009 (S03): a post-merge transition
// conflict whose policy-hold write ALSO fails must keep the caller's own
// integration window. The window is released only after the hold is written
// (completePostMergeConflict's own comment; internal/factory/
// integration_merge_step.go holdThenReleaseCauseSeven is the pattern: a failed
// hold write returns before any release). Before the repair the CLI logged the
// hold failure, released the window anyway, and still told the operator that
// the window policy holds.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestS03_HoldWriteFailureKeepsOwnWindow(t *testing.T) {
	root, integ, _ := mwq19Fixture(t)
	db := fcOpen(t, root)

	policyPath := filepath.Join(root, ".moai", "state", factory.IntegrationWindowPolicyFileName)
	factoryCompleteTransitionHook = func() {
		// Transition conflict: the card's version moves after the merge.
		if _, err := db.DB.Exec("UPDATE cards SET version = version + 1 WHERE card_id = 't1' AND run_id = ?", fcRun); err != nil {
			t.Errorf("hook bump: %v", err)
		}
		// Hold-write failure: a non-empty directory sits where the policy
		// record must be renamed into place, so the atomic write cannot land.
		if err := os.MkdirAll(filepath.Join(policyPath, "blocker"), 0o755); err != nil {
			t.Errorf("hook policy blocker: %v", err)
		}
	}
	t.Cleanup(func() { factoryCompleteTransitionHook = nil })

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatalf("a post-merge transition conflict must fail the run")
	}

	// The merge commit stays on the integration branch (unchanged behavior).
	head := fcGit(t, integ, "rev-parse", "HEAD")
	parents := strings.Fields(fcGit(t, integ, "rev-list", "--parents", "-n", "1", head))
	if len(parents) != 3 {
		t.Fatalf("the merge commit must stay on the integration branch: %v", parents)
	}
	if !strings.Contains(err.Error(), head[:12]) {
		t.Fatalf("the failure must name the merge SHA %s: %v", head[:12], err)
	}

	// The card state is unchanged.
	if card := fcCard(t, root, "t1"); card.State != homestate.CardMergeReady {
		t.Fatalf("the card state must not change: %q", card.State)
	}

	// S03: the hold was NOT written, so the caller's window must still be held
	// by the caller — releasing it would let the next queued lane be promoted
	// onto the conflicted merge.
	lock := sdWindow(t, root)
	if !lock.Held() || lock.SessionID != "sess-lane-1" {
		t.Fatalf("S03_WINDOW_RELEASED_WITHOUT_HOLD: the window must stay held by the caller when the hold write failed: %+v", lock)
	}

	// And the message must not claim a hold that was never written.
	if strings.Contains(err.Error(), "the window policy holds with cause") {
		t.Fatalf("S03_FALSE_HOLD_CLAIM: the error claims the policy holds although the hold write failed: %v", err)
	}
	if !strings.Contains(err.Error(), "hold") {
		t.Fatalf("the error must say the hold was not written: %v", err)
	}
}
