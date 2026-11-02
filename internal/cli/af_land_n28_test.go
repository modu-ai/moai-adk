package cli

// N28 (audit db0a69698a, reopened as a blocker by the operator) through the real `moai af` verbs, a real ledger and
// real git. A card must not finish on a landed candidate that no longer stands:
//  1. the newest green run and the newest gate run about the landed commit must not be failures, whether or not the
//     card was ever taken back;
//  2. when the card left MERGING after the candidate was offered and has since recorded evidence about another
//     commit, the landed commit counts only if it was measured again after that evidence, with a passing green and a
//     passing gate (a gate of the landed commit alone does not say that the work in between was given up).
//
// Each refusal has the way out the rules ask for: measuring the landed commit again.

import (
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// afN28Landed is a card whose c1 was measured, queued and landed at its only epoch, still in MERGING.
func afN28Landed(t *testing.T) afN3Card {
	t.Helper()
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: strconv.FormatInt(epoch, 10)}
	a.measure(c1)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", c1)
	if out := a.verb("land", "t1", "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("c1 did not land: %s", out)
	}
	return a
}

// record writes one run of a kind about sha with the result, at the card's epoch.
func (a afN3Card) record(kind, sha, result string) {
	a.t.Helper()
	a.verb("evidence", "t1", "--epoch", a.e2, "--kind", kind, "--subject-sha", sha, "--result", result)
}

// tryFinish makes the two moves that finish a card that is in MERGING and reports the first that is refused.
func (a afN3Card) tryFinish() error {
	a.t.Helper()
	if _, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "MERGING", "--remote-integrated"); err != nil {
		return err
	}
	_, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "DONE")
	return err
}

// tryDone moves the card to DONE from where it is (REVIEWING or BUILDING) and reports a refusal.
func (a afN3Card) tryDone() error {
	a.t.Helper()
	_, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "DONE")
	return err
}

func (a afN3Card) wantNotDone(what string) {
	a.t.Helper()
	if got := a.stage(); got == core.StageDone {
		a.t.Errorf("%s: the card is DONE", what)
	}
}

// A failure recorded after the landing refuses DONE, with the card still in MERGING and after a withdrawal to
// REVIEWING alike; measured again and passing, the card finishes.
func TestAFVerbs_N28AFailureRecordedAfterTheLandingRefusesDoneUntilTheCommitIsMeasuredAgain(t *testing.T) {
	afSkipWithoutSh(t)
	for _, tc := range []struct {
		name string
		do   func(a afN3Card)
	}{
		{"the green and the gate failed", func(a afN3Card) { a.record("green", a.c1, "fail"); a.record("gate", a.c1, "fail") }},
		{"the green failed", func(a afN3Card) { a.record("green", a.c1, "fail") }},
		{"the gate failed", func(a afN3Card) { a.record("gate", a.c1, "fail") }},
	} {
		t.Run(tc.name+", still in MERGING", func(t *testing.T) {
			a := afN28Landed(t)
			tc.do(a)
			if err := a.tryFinish(); err == nil {
				t.Fatal("the card finished with a failing newest run of its landed commit")
			}
			a.wantNotDone("a failing newest run")
			a.measure(a.c1)
			if err := a.tryFinish(); err != nil {
				t.Fatalf("c1 passes again and the card is refused: %v", err)
			}
			if got := a.stage(); got != core.StageDone {
				t.Errorf("stage %s, want DONE", got)
			}
		})
		t.Run(tc.name+", withdrawn to REVIEWING", func(t *testing.T) {
			a := afN28Landed(t)
			a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
			tc.do(a)
			if err := a.tryDone(); err == nil {
				t.Fatal("the card finished with a failing newest run of its landed commit")
			}
			a.wantNotDone("a failing newest run after a withdrawal")
			a.measure(a.c1)
			if err := a.tryDone(); err != nil {
				t.Fatalf("c1 passes again and the card is refused: %v", err)
			}
		})
	}
}

// After a withdrawal to REVIEWING, and after a move back into BUILDING, a gate-only new commit c2 that is never
// offered does not let the card finish on c1: a gate of c1 alone is not enough either, c1 must be measured again in
// full (green and gate) after the evidence about c2.
func TestAFVerbs_N28AGateOnlyNewCommitAfterAWithdrawalOrReworkRefusesDoneUntilC1IsMeasuredAgainInFull(t *testing.T) {
	afSkipWithoutSh(t)
	for _, rework := range []bool{false, true} {
		t.Run(afN3Name(rework), func(t *testing.T) {
			a := afN28Landed(t)
			c2 := a.newCommit()
			a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
			if rework {
				a.verb("move", "t1", "--epoch", a.e2, "--stage", "BUILDING", "--sub", "IMPLEMENT")
			}
			a.record("gate", c2, "pass")
			if err := a.tryDone(); err == nil {
				t.Fatal("DONE accepted with c2 gated after the withdrawal and c1 not measured again")
			}
			a.record("gate", a.c1, "pass")
			if err := a.tryDone(); err == nil {
				t.Fatal("DONE accepted on a gate of c1 alone after the evidence about c2")
			}
			a.wantNotDone("c2 gated, c1 gated alone")
			a.record("green", a.c1, "pass")
			a.record("gate", a.c1, "pass")
			if err := a.tryDone(); err != nil {
				t.Fatalf("c1 measured again in full after c2, and the card is refused: %v", err)
			}
			if got := a.stage(); got != core.StageDone {
				t.Errorf("stage %s, want DONE", got)
			}
		})
	}
}

// The control: a gate of another commit while the card never left MERGING does not block the card.
func TestAFVerbs_N28AGateOfAnotherCommitWhileTheCardStaysInMergingDoesNotBlockIt(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	c2 := a.newCommit()
	a.record("gate", c2, "pass")
	if err := a.tryFinish(); err != nil {
		t.Fatalf("a second commit was gated and the card never left MERGING, and it is refused: %v", err)
	}
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
}
