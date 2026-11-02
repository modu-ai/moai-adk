package cli

// The landing-window repair (follow-up of the final code audit of a36ea8c368, the window left open by N66), through the
// real `moai af` verbs, a real ledger and real git. The tests begin with TestAFLandingWindow_ so that one anchored run
// picks them, and they go through afRunBoth, so with AF_TEST_BIN set the same tests drive a built binary.
//
// The ledger refuses a failing green, gate or review run about the candidate of a row that is in landing, so the instant
// between a landing's last look at the newest runs and the swap of the trunk ref is closed from the evidence side. What a
// command line can set up deterministically is the state that instant leaves behind when the lander dies: a row left in
// landing. (The instant itself is hooked in the mergeq tests, on Merger.LastCheck.)
//
//   - the verb prints the refusal, says why and what to do, exits non-zero and changes nothing;
//   - `moai af reconcile` settles the row (back to waiting, or landed when the candidate is in the trunk) and the same
//     command is then accepted;
//   - the retry after the landing is recorded, and the rule after landing holds the card (E12).

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afLWLeftLanding takes the card's waiting row to landing the way a lander that then died leaves it: the lock is given back,
// the row says landing.
func afLWLeftLanding(t *testing.T, a afN3Card) {
	t.Helper()
	ctx := context.Background()
	row := a.row()
	lock, err := a.st.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead", PID: afTrGoneProcess(t), PIDStart: "gone"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.st.QueueUpdate(ctx, "t1", ledger.CardEpoch(row.CardEpoch, row.CandidateSHA), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := a.st.QueueUpdate(ctx, "t1", ledger.LockEpoch(lock, row.CandidateSHA), ledger.QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := a.st.ReleaseLock(ctx, ledger.LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	if got := a.row().State; got != ledger.QueueLanding {
		t.Fatalf("setup: the row is %s, want landing", got)
	}
}

func afLWEvidence(a afN3Card, kind, sha, result string) (string, error) {
	a.t.Helper()
	return a.try("evidence", "t1", "--epoch", a.e2, "--kind", kind, "--subject-sha", sha, "--result", result)
}

func afLWEvidenceCount(a afN3Card) int {
	a.t.Helper()
	return len(afSnapshot(a.t, a.st).Evidence)
}

// The verb refuses a failing green, gate or review run about the candidate of a landing row: non-zero exit, the reason,
// the way on, and nothing recorded. Beside it, what stays accepted.
func TestAFLandingWindow_TheVerbRefusesAFailingRunAboutTheCandidateOfALandingRow(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	afLWLeftLanding(t, a)
	c2 := a.newCommit()

	before := afLWEvidenceCount(a)
	for _, kind := range []string{"green", "gate", "review"} {
		out, err := afLWEvidence(a, kind, a.c1, "fail")
		if err == nil {
			t.Fatalf("a failing %s run about the candidate of a landing row was recorded: %.200s", kind, out)
		}
		for _, want := range []string{"being landed", "moai af reconcile", "moai af land"} {
			if !afIDSays(err.Error(), want) {
				t.Errorf("the refusal of the %s run does not say %q: %v", kind, want, err)
			}
		}
	}
	if got := afLWEvidenceCount(a); got != before {
		t.Errorf("the refused runs changed the evidence: %d rows, was %d", got, before)
	}
	if got := a.row().State; got != ledger.QueueLanding {
		t.Errorf("the refused runs moved the row to %s", got)
	}

	// controls: a passing run, a failing run about another commit, the kinds no landing reads
	for _, tc := range []struct{ kind, sha, result string }{
		{"gate", a.c1, "pass"}, {"review", a.c1, "pass"}, {"green", c2, "fail"}, {"gate", c2, "fail"}, {"red", a.c1, "fail"}, {"merge", a.c1, "fail"},
	} {
		if out, err := afLWEvidence(a, tc.kind, tc.sha, tc.result); err != nil {
			t.Errorf("a %s %s run about %.8s was refused: %v (%.200s)", tc.result, tc.kind, tc.sha, err, out)
		}
	}
}

// The way on the refusal names: reconcile settles the row (the candidate is not in the trunk, so it goes back to
// waiting) and the same command is then accepted.
func TestAFLandingWindow_ReconcileFreesTheRowAndTheRunIsThenRecorded(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	afLWLeftLanding(t, a)
	if _, err := afLWEvidence(a, "gate", a.c1, "fail"); err == nil {
		t.Fatal("a failing run was recorded against a landing row")
	}
	if out, err := afTrRun(a, "reconcile"); err != nil {
		t.Fatalf("reconcile: %v %.200s", err, out)
	}
	if got := a.row().State; got == ledger.QueueLanding {
		t.Fatalf("the row is still landing after the reconcile")
	}
	if out, err := afLWEvidence(a, "gate", a.c1, "fail"); err != nil {
		t.Fatalf("still refused after the reconcile: %v %.200s", err, out)
	}
}

// The retry path: a lander that moved the trunk and died before it wrote the row. Reconcile settles the row as landed,
// the failing run is then recorded (the retry), the post-landing rule refuses DONE, and measured again the card finishes.
func TestAFLandingWindow_TheRetryAfterTheLandingIsRecordedAndHoldsTheCard(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	afLWLeftLanding(t, a)
	if _, err := afLWEvidence(a, "gate", a.c1, "fail"); err == nil {
		t.Fatal("a failing run was recorded against a landing row")
	}
	// the trunk moved (the candidate is a fast-forward of it) and the lander died before it wrote the row
	afGit(t, a.repo, "update-ref", "refs/heads/af/trunk", a.c1)
	if out, err := afTrRun(a, "reconcile"); err != nil {
		t.Fatalf("reconcile: %v %.200s", err, out)
	}
	if got := a.row().State; got != ledger.QueueLanded {
		t.Fatalf("the row is %s after the reconcile of a candidate that is in the trunk, want landed", got)
	}
	if out, err := afLWEvidence(a, "gate", a.c1, "fail"); err != nil {
		t.Fatalf("the retry after the landing is refused: %v %.200s", err, out)
	}
	if err := a.tryFinish(); err == nil {
		t.Fatal("a failure recorded after the landing let the card finish")
	}
	if got := a.stage(); got == core.StageDone {
		t.Fatal("the card is DONE")
	}
	a.measure(a.c1)
	if err := a.tryFinish(); err != nil {
		t.Fatalf("measured again and passing, the card is refused: %v", err)
	}
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
}
