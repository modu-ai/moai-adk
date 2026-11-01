package cli

// factory_complete_guidance_t2013_test.go — card t2013 (decisions row R22):
// the completion guidance `moai factory complete` prints on a genuinely
// completed edge, and the edges that must NOT print it.
//
// The guidance is TEXT addressed to the operator. Nothing here asserts that
// ExitWorktree, the default-branch check, or /clear were executed: the tests
// read the printed order and the preserve-first wording, and nothing more.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t2013WantCompletionGuidance asserts the printed guidance names the four steps
// in the official order (ExitWorktree, default-branch check, /clear, next-card
// request), puts the preserve-first checklist before the first step, says the
// session cannot run /clear itself, and never suggests a force-clear.
func t2013WantCompletionGuidance(t *testing.T, out string) {
	t.Helper()
	markers := []string{"ExitWorktree", "default branch", "/clear", "next card"}
	last := -1
	for _, m := range markers {
		i := strings.Index(out, m)
		if i < 0 {
			t.Fatalf("guidance does not name %q:\n%s", m, out)
		}
		if i <= last {
			t.Fatalf("guidance names %q out of order (offset %d, previous marker at %d):\n%s", m, i, last, out)
		}
		last = i
	}
	pre := strings.Index(out, "ExitWorktree")
	for _, m := range []string{"completion evidence", "dirty changes", "unpushed commits", "active checks", "handoff", "unsubmitted input"} {
		i := strings.Index(out, m)
		if i < 0 || i > pre {
			t.Errorf("preserve-first item %q must appear before the ExitWorktree step (offset %d vs %d):\n%s", m, i, pre, out)
		}
	}
	for _, want := range []string{"operator", "cannot run /clear", "guidance only", "do not clear"} {
		if !strings.Contains(out, want) {
			t.Errorf("guidance does not say %q:\n%s", want, out)
		}
	}
	// A force-clear is never suggested: the only mention is the prohibition.
	if strings.Contains(strings.ReplaceAll(out, "never force-clear", ""), "force") {
		t.Errorf("guidance suggests a forced clear:\n%s", out)
	}
	// The line never claims a step happened.
	for _, claim := range []string{"was cleared", "has been cleared", "already exited", "exited the worktree"} {
		if strings.Contains(out, claim) {
			t.Errorf("guidance claims an executed step (%q):\n%s", claim, out)
		}
	}
}

// t2013WantNoCompletionGuidance asserts the output carries no completion
// guidance at all: not a step, not the word that marks a completed card.
func t2013WantNoCompletionGuidance(t *testing.T, what, out string) {
	t.Helper()
	for _, m := range []string{"card complete", "/clear", "ExitWorktree", "next card"} {
		if strings.Contains(out, m) {
			t.Errorf("%s: output carries completion guidance (%q):\n%s", what, m, out)
		}
	}
}

func TestT2013_CompletionGuidanceWording(t *testing.T) {
	t.Run("auto dispatch prints the ordered guidance", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		t.Setenv(config.EnvFactoryAutoDispatch, "")
		out := sdClearPolicyOutput(t, root)
		t2013WantCompletionGuidance(t, out)
		if !strings.Contains(out, "moai factory next") {
			t.Errorf("auto dispatch must name the lease verb for the next-card request:\n%s", out)
		}
	})
	t.Run("manual dispatch asks the leader for the next card, leases nothing", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchManual)
		out := sdClearPolicyOutput(t, root)
		t2013WantCompletionGuidance(t, out)
		if strings.Contains(out, "moai factory next") {
			t.Errorf("manual dispatch must not tell the lane to lease a card:\n%s", out)
		}
		if !strings.Contains(out, "leader") {
			t.Errorf("manual dispatch must route the next-card request to the leader:\n%s", out)
		}
	})
	t.Run("the guidance is one line", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		t.Setenv(config.EnvFactoryAutoDispatch, "")
		out := strings.TrimRight(sdClearPolicyOutput(t, root), "\n")
		if out == "" || strings.Contains(out, "\n") {
			t.Fatalf("want exactly one line, got %q", out)
		}
	})
}

// The edges. Only a genuinely completed edge (merged-local, merged-pr) prints
// the guidance; a PR that is opened but not merged, a still-open PR, a closed
// PR, a refused delivery (audit-pending, no origin, no re-measure record) print
// none.
func TestT2013_CompletionGuidanceOnlyOnCompletedEdges(t *testing.T) {
	t.Run("pr-open prints no completion guidance", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := agenticTrunkPrimaryFixture(t)
		_ = newGHDouble(t, f)
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
			t.Fatalf("card = %s, want pr-open", c.State)
		}
		t2013WantNoCompletionGuidance(t, "pr-open", out)
	})

	t.Run("pr still open on the re-run prints no completion guidance", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := agenticTrunkPrimaryFixture(t)
		_ = newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("second complete: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
			t.Fatalf("card = %s, want pr-open", c.State)
		}
		if !strings.Contains(out, "still open") {
			t.Fatalf("expected the still-open note, got:\n%s", out)
		}
		t2013WantNoCompletionGuidance(t, "pr still open", out)
	})

	t.Run("pr closed without merging prints no completion guidance", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := agenticTrunkPrimaryFixture(t)
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		d.state = "CLOSED"
		out, err := ghfComplete(t)
		if err == nil {
			t.Fatalf("a closed PR completed the card; out:\n%s", out)
		}
		t2013WantNoCompletionGuidance(t, "pr closed", out+"\n"+err.Error())
		if c := fcCard(t, f.root, "t1"); c.State == homestate.CardMergedPR {
			t.Fatalf("card = %s after a closed PR", c.State)
		}
	})

	t.Run("pr merged (merged-pr) prints the ordered guidance", func(t *testing.T) {
		sdClearLaneEnv(t)
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		f := agenticTrunkPrimaryFixture(t)
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		d.state, d.mergeOid = "MERGED", f.squashOnMain(t)
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete after the merge: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergedPR {
			t.Fatalf("card = %s, want merged-pr", c.State)
		}
		t2013WantCompletionGuidance(t, out)
	})

	for _, tc := range []struct {
		name string
		o    ghfOpts
	}{
		{"audit-pending (no sync-audit PASS record)", ghfOpts{syncStatus: ""}},
		{"audit-pending (sync_status complete-pending-audit)", ghfOpts{syncStatus: "complete-pending-audit"}},
		{"unpushed (no origin remote)", ghfOpts{syncStatus: "complete", noOrigin: true}},
		{"conflicting card", ghfOpts{syncStatus: "complete", mainMoves: "conflict"}},
	} {
		t.Run("refused complete prints no completion guidance: "+tc.name, func(t *testing.T) {
			sdClearLaneEnv(t)
			f := ghfNew(t, tc.o)
			_ = newGHDouble(t, f)
			out, err := ghfComplete(t)
			if err == nil {
				t.Fatalf("complete passed a card that must be refused; out:\n%s", out)
			}
			t2013WantNoCompletionGuidance(t, tc.name, out+"\n"+err.Error())
			if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady {
				t.Errorf("a refused complete moved the card to %s", c.State)
			}
		})
	}

	t.Run("merged-local refused for a missing re-measure record prints no guidance", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integWT, "t1")
		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		t.Setenv(config.EnvFactoryClearPolicy, config.FactoryClearPolicyEach)
		t.Chdir(root)
		out, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil {
			t.Fatalf("complete without a re-measure record succeeded; out:\n%s", out)
		}
		t2013WantNoCompletionGuidance(t, "merged-local refusal", out+"\n"+err.Error())
	})
}
