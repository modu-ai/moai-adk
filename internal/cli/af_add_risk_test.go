package cli

// N58 (final audit of a36ea8c368, mutant N59-add-risk): `moai af add --risk` hands the ledger the risks of the card, and a
// card with a risk must be reviewed before it is offered to the merge queue and before it is done (N40). A verb that drops
// the flag passed every test: the card was an ordinary one and the review requirement vanished without a sound.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFAdd_ARiskCardNeedsItsReviewBeforeItIsOfferedAndBeforeDone(t *testing.T) {
	afSkipWithoutSh(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, c1 := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	afAddVerb(t, "add", "r1", "--class", "A", "--path", "x.txt", "--risk", "security", "--risk", "data")
	db, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := ledger.Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cards, err := st.Cards(context.Background())
	if err != nil || len(cards) != 1 || strings.Join(cards[0].Spec.Risk, ",") != "security,data" {
		t.Fatalf("the card does not carry the risks it was added with: %+v %v", cards, err)
	}

	claimed := afAddVerb(t, "claim", "--session", "s-r1", "--worktree", repo)
	e := afTrEpoch(t, claimed)
	for _, kind := range []string{"green", "gate"} {
		afAddVerb(t, "evidence", "r1", "--epoch", e, "--kind", kind, "--subject-sha", c1, "--result", "pass")
	}
	// a risk card is offered from REVIEWING, with a passing review of the commit
	if _, _, err := afRunBoth("enqueue", "r1", "--epoch", e, "--candidate", c1); err == nil || !strings.Contains(err.Error(), "reviewed") {
		t.Errorf("enqueue of a risk card that was not reviewed: %v, want the refusal that asks for the review", err)
	}
	afAddVerb(t, "move", "r1", "--epoch", e, "--stage", "REVIEWING")
	if _, _, err := afRunBoth("enqueue", "r1", "--epoch", e, "--candidate", c1); err == nil || !strings.Contains(err.Error(), "review") {
		t.Errorf("enqueue of a risk card with no passing review: %v, want a refusal that asks for it", err)
	}
	afAddVerb(t, "evidence", "r1", "--epoch", e, "--kind", "review", "--subject-sha", c1, "--result", "pass")
	afAddVerb(t, "enqueue", "r1", "--epoch", e, "--candidate", c1)
	if out := afAddVerb(t, "land", "r1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land: %s", out)
	}
	// a review that fails after the landing keeps the card from DONE
	afAddVerb(t, "evidence", "r1", "--epoch", e, "--kind", "review", "--subject-sha", c1, "--result", "fail")
	if _, _, err := afRunBoth("move", "r1", "--epoch", e, "--stage", "MERGING", "--remote-integrated"); err == nil {
		t.Error("a risk card whose review failed after the landing was integrated")
	}
	if _, _, err := afRunBoth("move", "r1", "--epoch", e, "--stage", "DONE"); err == nil {
		t.Error("a risk card whose review failed became DONE")
	}
	afAddVerb(t, "evidence", "r1", "--epoch", e, "--kind", "review", "--subject-sha", c1, "--result", "pass")
	afAddVerb(t, "move", "r1", "--epoch", e, "--stage", "MERGING", "--remote-integrated")
	afAddVerb(t, "move", "r1", "--epoch", e, "--stage", "DONE")
	cards, err = st.Cards(context.Background())
	if err != nil || cards[0].Position.Stage != core.StageDone {
		t.Fatalf("cards %+v %v, want r1 DONE after the review passed", cards, err)
	}
}

// The control: the same card added without --risk is offered with no review at all.
func TestAFAdd_ACardAddedWithoutRiskIsOfferedWithNoReview(t *testing.T) {
	afSkipWithoutSh(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	repo, c1 := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	afAddVerb(t, "add", "p1", "--class", "A", "--path", "x.txt")
	e := afTrEpoch(t, afAddVerb(t, "claim", "--session", "s-p1", "--worktree", repo))
	for _, kind := range []string{"green", "gate"} {
		afAddVerb(t, "evidence", "p1", "--epoch", e, "--kind", kind, "--subject-sha", c1, "--result", "pass")
	}
	afAddVerb(t, "enqueue", "p1", "--epoch", e, "--candidate", c1)
}
