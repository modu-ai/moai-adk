package cli

// local_main_verb_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for the landing verb and factory complete on the primary surface
// (REQ-LMF-002, REQ-LMF-003; plan §B2, §B3). Each test runs the production
// entry point in process: the verb through runIntegrationMerge, complete
// through runFactory.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestIntegrationMergeWorktreeAcceptsPrimaryWhenEnabled(t *testing.T) {
	// With the gate on, the verb's surface check admits the primary and the
	// landing reaches the holder decision. A window held by another session is
	// then refused as NotHolder, not as a primary-checkout refusal.
	lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}, holder: "sess-foreign"})
	err := lmfVerb(t)
	if err == nil {
		t.Fatalf("a window held by another session must refuse the verb")
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitNotHolder {
		t.Fatalf("with the gate on, the verb must reach the holder decision (MergeExitNotHolder %d), got code %d (ok=%v): %v", factory.MergeExitNotHolder, code, ok, err)
	}
	if !strings.Contains(err.Error(), "sess-foreign") {
		t.Fatalf("the holder refusal must name the holder: %v", err)
	}
}

func TestFactoryCompletePrimaryTreeGateEnabled(t *testing.T) {
	// With the gate on, complete lands the card through the primary's local main.
	root, _ := lmfMergeFixture(t, true, lmfSpec{cardPaths: []string{"alpha.txt"}})
	before := lmfHead(t, root)
	if _, _, err := runFactory(t, "complete", lmfCard, "--run", fcRun); err != nil {
		t.Fatalf("with the gate on, complete must land the card through the primary's local main: %v", err)
	}
	card := fcCard(t, root, lmfCard)
	if card.State != homestate.CardMergedLocal {
		t.Fatalf("the card must land merged-local, got %q", card.State)
	}
	if card.MergeSHA == "" {
		t.Fatalf("the merged-local record must carry the merge SHA")
	}
	if after := lmfHead(t, root); after == before {
		t.Fatalf("the landing must advance the primary's local main from %s", before)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("complete must release the window after its transitions: %+v", lock)
	}
}

func TestFactoryCompletePrimaryTreeGateDisabled(t *testing.T) {
	// With the gate off, the primary holding main keeps refusing complete, and
	// the refusal changes nothing.
	root, _ := lmfMergeFixture(t, false, lmfSpec{cardPaths: []string{"alpha.txt"}})
	before := fcCard(t, root, lmfCard)
	beforeHead := lmfHead(t, root)
	_, _, err := runFactory(t, "complete", lmfCard, "--run", fcRun)
	if err == nil {
		t.Fatalf("with the gate off, the primary holding %s must keep refusing complete", lmfBranch)
	}
	if !strings.Contains(err.Error(), "parent checkout") {
		t.Fatalf("the refusal must name the parent checkout as the holder: %v", err)
	}
	sdCardUnchanged(t, "gate off", root, lmfCard, before)
	if head := lmfHead(t, root); head != beforeHead {
		t.Fatalf("a refused complete must move no HEAD: %s -> %s", beforeHead, head)
	}
}

func TestFactoryCompleteNoIntegrationTreeRefused(t *testing.T) {
	// No tree holds main: the primary checks out feature. Complete refuses in
	// both gate states, switches no branch, and leaves the card unchanged.
	for _, gateOn := range []bool{false, true} {
		gateOn := gateOn
		name := "gate-off"
		if gateOn {
			name = "gate-on"
		}
		t.Run(name, func(t *testing.T) {
			root, _ := lmfMergeFixture(t, gateOn, lmfSpec{cardPaths: []string{"alpha.txt"}})
			fcGit(t, root, "branch", "feature")
			fcGit(t, root, "symbolic-ref", "HEAD", "refs/heads/feature")
			before := fcCard(t, root, lmfCard)
			_, _, err := runFactory(t, "complete", lmfCard, "--run", fcRun)
			if err == nil {
				t.Fatalf("with no tree holding %s, complete must refuse", lmfBranch)
			}
			if !strings.Contains(err.Error(), "is not provisioned") {
				t.Fatalf("the refusal must name the unprovisioned integration worktree: %v", err)
			}
			sdCardUnchanged(t, name, root, lmfCard, before)
			if head := fcGit(t, root, "rev-parse", "--symbolic-full-name", "HEAD"); head != "refs/heads/feature" {
				t.Fatalf("the refusal must switch no branch: HEAD now names %q", head)
			}
		})
	}
}
