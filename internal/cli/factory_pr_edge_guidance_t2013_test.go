package cli

// factory_pr_edge_guidance_t2013_test.go — card t2013: the pull-request edge of
// `moai factory complete` prints the review-start and ready-conversion guidance
// (a draft does not start the native review; a failing non-required check does
// not block ready), and no other edge prints it.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

var t2013PREdgeMarkers = []string{
	"does not start the native review",
	"leaving the PR in draft while waiting is not a wait reason",
	"outside the protected required set",
	"does not block the ready conversion",
	"record the wait with `moai factory wait` and end the turn",
	"repaired by this same lane in this worktree",
}

func TestT2013_PREdgeGuidanceDraftAndNonRequiredChecks(t *testing.T) {
	t.Run("pr-open edge prints both sentences", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		_ = newGHDouble(t, f)
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete: %v", err)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
			t.Fatalf("card = %s, want pr-open", c.State)
		}
		for _, m := range t2013PREdgeMarkers {
			if !strings.Contains(out, m) {
				t.Errorf("pr-open output lacks %q:\n%s", m, out)
			}
		}
	})

	t.Run("pr still open on the re-run does not repeat it", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		_ = newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("second complete: %v", err)
		}
		for _, m := range t2013PREdgeMarkers {
			if strings.Contains(out, m) {
				t.Errorf("still-open re-run carries PR-edge guidance %q:\n%s", m, out)
			}
		}
	})

	t.Run("merged-pr edge does not carry it", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("first complete: %v", err)
		}
		d.state, d.mergeOid = "MERGED", f.squashOnMain(t)
		out, err := ghfComplete(t)
		if err != nil {
			t.Fatalf("complete after the merge: %v", err)
		}
		for _, m := range t2013PREdgeMarkers {
			if strings.Contains(out, m) {
				t.Errorf("merged-pr output carries PR-edge guidance %q:\n%s", m, out)
			}
		}
	})

	t.Run("refused delivery does not carry it", func(t *testing.T) {
		sdClearLaneEnv(t)
		f := ghfNew(t, ghfOpts{syncStatus: "complete-pending-audit"})
		_ = newGHDouble(t, f)
		out, err := ghfComplete(t)
		if err == nil {
			t.Fatalf("complete passed a card that must be refused; out:\n%s", out)
		}
		for _, m := range t2013PREdgeMarkers {
			if strings.Contains(out+"\n"+err.Error(), m) {
				t.Errorf("refused output carries PR-edge guidance %q", m)
			}
		}
	})
}
