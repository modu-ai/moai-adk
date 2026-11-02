package cli

// The transition group of the repair after the final audit of a36ea8c368 (N63, N66, N69 first bullet, N58 for the two),
// through the real `moai af` verbs, a real ledger and real git. The tests begin with TestAFTransition_ so that one anchored
// run picks them, and they go through afRunBoth, so with AF_TEST_BIN set the same tests drive a built binary (every verb
// then a separate OS process, which is what a restart is).
//
//   - N63: after the N54 refusal a card whose holder left (released it, or was reaped after its process went) had no way
//     out: land refused, evidence, enqueue and move needed the lease, resume and reconcile and reap did nothing, and a
//     MERGING card is not claimable. The way out is `move --stage REVIEWING` at the card's own epoch; the ledger accepts
//     it for a waiting row whose candidate cannot land, and the new holder claims the card and lands it with a real pass.
//   - N66: the failing-run check ran once, before the gate; a run that failed while the gate ran was landed all the same.
//     It is asked again under the landing lock, immediately before the swap of the trunk ref.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// afTrHolder is what became of the card's holder when the failing run was on the row.
type afTrHolder string

const (
	afTrAlive    afTrHolder = "the holder is alive"
	afTrReleased afTrHolder = "the holder released the card"
	afTrGone     afTrHolder = "the holder process is gone, then reap"
	afTrExpired  afTrHolder = "the lease ran out and the holder process is gone, then reap"
)

// afTrGoneProcess is the pid of a process that has run and been waited for.
func afTrGoneProcess(t *testing.T) int {
	t.Helper()
	c := exec.Command("true")
	if err := c.Run(); err != nil {
		t.Fatal(err)
	}
	return c.ProcessState.Pid()
}

// afTrMeasured is a card whose c1 is measured (green and gate pass) and enqueued: the row waits, the card is in MERGING,
// and it is held as the variant says.
func afTrMeasured(t *testing.T, how afTrHolder) afN3Card {
	t.Helper()
	claim := ledger.ClaimRequest{SessionID: "s-t1"}
	switch how {
	case afTrAlive, afTrReleased:
		start, _ := homestate.ProbeProcessIdentity(os.Getpid())
		claim.PID, claim.PIDStart = os.Getpid(), start
	default:
		claim.PID, claim.PIDStart = afTrGoneProcess(t), "gone"
	}
	if how == afTrExpired {
		claim.Lease = time.Millisecond
	}
	db, repo, c1, st, epoch := afUnqueuedCardAs(t, claim)
	afBindFixture(t, st, repo)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: fmt.Sprint(epoch)}
	a.measure(c1)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", c1, "--repo", repo)
	if got := a.row().State; got != ledger.QueueWaiting {
		t.Fatalf("setup: the row is %s, want waiting", got)
	}
	return a
}

// afTrLeave is the holder leaving the card: it releases it, or the process is gone and a reap takes the lease back.
func afTrLeave(t *testing.T, a afN3Card, how afTrHolder) {
	t.Helper()
	switch how {
	case afTrReleased:
		a.verb("release", "t1", "--epoch", a.e2)
	case afTrGone, afTrExpired:
		time.Sleep(5 * time.Millisecond)
		if out, err := afTrRun(a, "reap"); err != nil || !strings.Contains(out, `"t1"`) {
			t.Fatalf("reap did not take the card back from the holder that is gone: %q %v", out, err)
		}
	}
}

// afTrRun runs a verb that has no positional argument, with the ledger at the end.
func afTrRun(a afN3Card, args ...string) (string, error) {
	a.t.Helper()
	out, errOut, err := afRunBoth(append(append([]string{}, args...), "--db", a.db)...)
	if err != nil {
		return out + errOut, err
	}
	return out, nil
}

func afTrEpoch(t *testing.T, claimOut string) string {
	t.Helper()
	var c struct {
		CardID string `json:"card_id"`
		Epoch  int64  `json:"epoch"`
	}
	if err := json.Unmarshal([]byte(claimOut), &c); err != nil || c.CardID == "" || c.Epoch == 0 {
		t.Fatalf("claim printed %q: %+v %v", claimOut, c, err)
	}
	return fmt.Sprint(c.Epoch)
}

// afTrRestart is a restart of the process that set the card up: its connection to the ledger is closed and a new one is
// opened. Every verb after it opens the ledger afresh (and with AF_TEST_BIN it is a process of its own).
func afTrRestart(t *testing.T, a afN3Card) afN3Card {
	t.Helper()
	if err := a.st.Close(); err != nil {
		t.Fatal(err)
	}
	st, err := ledger.Open(t.Context(), a.db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	a.st = st
	return a
}

// afTrWayOut is `moai af move <card> --epoch <n> --stage REVIEWING`, the way out the refusal names.
func afTrWayOut(card, epoch string) string {
	return fmt.Sprintf("moai af move %s --epoch %s --stage REVIEWING", card, epoch)
}

// afTrCarryOn is what follows the way out: a new session claims the card, measures the commit again with real passes,
// offers it, lands it, and finishes the card.
func afTrCarryOn(t *testing.T, a afN3Card, alsoPassing ...string) {
	t.Helper()
	if out, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING", "--reason", "the holder is gone"); err != nil {
		t.Fatalf("the way out (%s) is refused: %v (%.200s)", afTrWayOut("t1", a.e2), err, out)
	}
	if got := a.stage(); got != core.StageReviewing {
		t.Fatalf("stage %s after the way out, want REVIEWING", got)
	}
	out, err := afTrRun(a, "claim", "--session", "s-next", "--worktree", a.repo)
	if err != nil {
		t.Fatalf("the card cannot be claimed after the way out: %v", err)
	}
	a.e2 = afTrEpoch(t, out)
	if o, err := a.try("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo); err == nil {
		t.Errorf("enqueue accepted a commit whose newest run failed and that was not measured again: %.200s", o)
	}
	a.measure(a.c1)
	for _, kind := range alsoPassing {
		a.record(kind, a.c1, "pass")
	}
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	if landed := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(landed, `"outcome":"landed"`) {
		t.Fatalf("the re-landing: %s", landed)
	}
	if !a.inTrunk(a.c1) {
		t.Error("c1 landed and is not in the trunk")
	}
	if err := a.tryFinish(); err != nil {
		t.Fatalf("the card is refused after the re-landing: %v", err)
	}
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
}

// ---- N63 ----------------------------------------------------------------------------------------------------------------

// The audit's wedge for every way a holder can leave: land refuses, and every other verb is tried in turn; none of them
// moves the card, and the documented move does.
func TestAFTransition_N63AHolderThatLeftARefusedRowLeavesAWayOutThatEndsInDone(t *testing.T) {
	afSkipWithoutSh(t)
	for _, how := range []afTrHolder{afTrReleased, afTrGone, afTrExpired} {
		t.Run(string(how), func(t *testing.T) {
			a := afTrMeasured(t, how)
			a.record("gate", a.c1, "fail")
			afTrLeave(t, a, how)
			a = afTrRestart(t, a)
			before := afSnapshot(t, a.st)

			if out, err := a.try("land", "t1", "--gate", "true"); err == nil {
				t.Fatalf("land landed a commit whose newest gate run failed: %.200s", out)
			}
			for name, args := range map[string][]string{
				"evidence": {"evidence", "t1", "--epoch", a.e2, "--kind", "gate", "--subject-sha", a.c1, "--result", "pass"},
				"enqueue":  {"enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1},
				"resume":   {"resume", "t1", "--epoch", a.e2, "--actor", "operator", "--reason", "way out"},
			} {
				if out, err := a.try(args...); err == nil {
					t.Errorf("%s changed a card that holds no lease: %.200s", name, out)
				}
			}
			if out, err := afTrRun(a, "claim", "--session", "s9", "--worktree", a.repo); err == nil {
				t.Errorf("a MERGING card was claimed: %.200s", out)
			}
			for _, verb := range []string{"reconcile", "reap"} {
				if _, err := afTrRun(a, verb); err != nil {
					t.Errorf("%s: %v", verb, err)
				}
			}
			if got := a.stage(); got != core.StageMerging {
				t.Fatalf("stage %s after the verbs that must not move the card, want MERGING", got)
			}
			if got := a.row().State; got != ledger.QueueWaiting {
				t.Fatalf("the row is %s, want waiting", got)
			}
			after := afSnapshot(t, a.st)
			if len(after.Evidence) != len(before.Evidence) || len(after.Queue) != len(before.Queue) {
				t.Errorf("the refused verbs changed the evidence or the queue: %+v -> %+v", before, after)
			}

			afTrCarryOn(t, a)
		})
	}
}

// The ways out are for the case that needs them. Beside them the ordinary paths hold: a landing that is not refused lets
// the card out as it did before N54, a live holder re-measures and lands, and a row that can land is not let go without
// its holder.
func TestAFTransition_N63TheOrdinaryPathsHoldBesideTheWayOut(t *testing.T) {
	afSkipWithoutSh(t)
	t.Run("no failing run: the holder releases in MERGING, land lands, the card goes back and is claimed (as before N54)", func(t *testing.T) {
		a := afTrMeasured(t, afTrReleased)
		a.verb("release", "t1", "--epoch", a.e2)
		if out, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING"); err == nil {
			t.Fatalf("a row that can land let the card out without a landing: %.200s", out)
		}
		if out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
			t.Fatalf("land: %s", out)
		}
		a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
		if out, err := afTrRun(a, "claim", "--session", "s-next", "--worktree", a.repo); err != nil || !strings.Contains(out, `"card_id":"t1"`) {
			t.Fatalf("claim after the landing: %q %v", out, err)
		}
	})
	t.Run("a failing run and a live holder: measure again with a pass, land, finish", func(t *testing.T) {
		a := afTrMeasured(t, afTrAlive)
		a.record("gate", a.c1, "fail")
		if out, err := a.try("land", "t1", "--gate", "true"); err == nil {
			t.Fatalf("landed with a failing newest gate run: %.200s", out)
		}
		a.record("gate", a.c1, "pass")
		if out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
			t.Fatalf("land after measuring again: %s", out)
		}
		if err := a.tryFinish(); err != nil {
			t.Fatal(err)
		}
		if got := a.stage(); got != core.StageDone {
			t.Errorf("stage %s, want DONE", got)
		}
	})
	t.Run("a failing run and a live holder: the holder takes the card back with its lease", func(t *testing.T) {
		a := afTrMeasured(t, afTrAlive)
		a.record("gate", a.c1, "fail")
		a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
		if got := a.stage(); got != core.StageReviewing {
			t.Errorf("stage %s, want REVIEWING", got)
		}
	})
	t.Run("a failure about another commit does not open the way out", func(t *testing.T) {
		a := afTrMeasured(t, afTrReleased)
		c2 := a.newCommit()
		a.record("green", c2, "fail")
		a.verb("release", "t1", "--epoch", a.e2)
		if out, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING"); err == nil {
			t.Fatalf("a failure about c2 let a card out whose row (c1) can land: %.200s", out)
		}
		if out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
			t.Fatalf("c1 does not land although only c2 failed: %s", out)
		}
	})
}

// ---- N66 ----------------------------------------------------------------------------------------------------------------

// afTrLandWhile runs `af land` with a gate that waits, calls during once the gate is running, lets the gate go, and returns
// what land answered.
func afTrLandWhile(t *testing.T, a afN3Card, during func()) (string, error) {
	t.Helper()
	root := t.TempDir()
	marker, release := filepath.Join(root, "gate-running"), filepath.Join(root, "gate-go")
	gate := fmt.Sprintf("touch %q; i=0; while [ ! -e %q ] && [ $i -lt 200 ]; do sleep 0.1; i=$((i+1)); done; true", marker, release)
	type result struct {
		out string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		out, errOut, err := afRunBoth("land", "t1", "--db", a.db, "--repo", a.repo, "--gate", gate)
		ch <- result{out + errOut, err}
	}()
	deadline := time.Now().Add(60 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = os.WriteFile(release, nil, 0o644)
			t.Fatal("the gate of land never started")
		}
		time.Sleep(20 * time.Millisecond)
	}
	during()
	if err := os.WriteFile(release, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	r := <-ch
	return r.out, r.err
}

func TestAFTransition_N66AFailureRecordedWhileTheGateRunsIsNotLanded(t *testing.T) {
	afSkipWithoutSh(t)
	for _, how := range []afTrHolder{afTrAlive, afTrReleased, afTrGone} {
		t.Run(string(how), func(t *testing.T) {
			a := afTrMeasured(t, how)
			trunk := afGit(t, a.repo, "rev-parse", mergeq.DefaultTrunk)
			out, err := afTrLandWhile(t, a, func() { a.record("review", a.c1, "fail") })
			if err == nil {
				t.Fatalf("land landed a commit whose review failed while the gate ran: %.300s", out)
			}
			for _, want := range []string{"newest review run", "measure", "again"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
			if got := afGit(t, a.repo, "rev-parse", mergeq.DefaultTrunk); got != trunk {
				t.Errorf("the trunk moved from %s to %s", trunk, got)
			}
			if a.inTrunk(a.c1) {
				t.Error("c1 is in the trunk")
			}
			if got := a.row().State; got != ledger.QueueWaiting {
				t.Fatalf("the row is %s, want waiting so that the card is not stuck", got)
			}
			if got := a.stage(); got != core.StageMerging {
				t.Fatalf("stage %s, want MERGING", got)
			}
			switch how {
			case afTrAlive:
				// the holder measures the review again and lands
				a.record("review", a.c1, "pass")
				if out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
					t.Fatalf("land after the review passed: %s", out)
				}
				if err := a.tryFinish(); err != nil {
					t.Fatal(err)
				}
			default:
				// the holder leaves (it went, or released the card) and a successor takes the way out
				afTrLeave(t, a, how)
				a = afTrRestart(t, a)
				afTrCarryOn(t, a, "review")
			}
		})
	}
}

// The pair: nothing recorded while the gate runs, and the same land lands.
func TestAFTransition_N66WithNothingRecordedWhileTheGateRunsTheCardLands(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	out, err := afTrLandWhile(t, a, func() {})
	if err != nil || !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land: %q %v", out, err)
	}
	if !a.inTrunk(a.c1) {
		t.Error("c1 landed and is not in the trunk")
	}
}

// A failure about ANOTHER commit recorded while the gate runs is not a failure of the candidate (D5, N74): the landing
// stands, as it does when it is recorded before the call.
func TestAFTransition_N66AFailureAboutAnotherCommitWhileTheGateRunsDoesNotStopTheLanding(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	c2 := a.newCommit()
	out, err := afTrLandWhile(t, a, func() { a.record("green", c2, "fail") })
	if err != nil || !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("land: %q %v", out, err)
	}
}

// E6 (decided by the operator, unchanged by the transition group): a candidate that is the trunk's own commit lands as
// "already in the trunk" and the card reaches DONE. The new last check sits on the path that moves the trunk, which this
// landing does not take.
func TestAFTransition_E6ACandidateThatIsTheTrunksOwnCommitStillLandsAsAlreadyInTheTrunk(t *testing.T) {
	afSkipWithoutSh(t)
	a := afTrMeasured(t, afTrAlive)
	// the card is taken back, and offered again with the commit the trunk already is
	a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
	base := afGit(t, a.repo, "rev-parse", mergeq.DefaultTrunk)
	a.measure(base)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", base, "--repo", a.repo)
	out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true")
	if !strings.Contains(out, `"outcome":"landed"`) || !strings.Contains(out, "already in the trunk") {
		t.Fatalf("land of the trunk's own commit: %s", out)
	}
	if got := afGit(t, a.repo, "rev-parse", mergeq.DefaultTrunk); got != base {
		t.Errorf("the trunk moved from %s to %s", base, got)
	}
	if err := a.tryFinish(); err != nil {
		t.Fatalf("the card is refused DONE: %v", err)
	}
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
}

// ---- N69, first bullet: the refusal names commands that run as printed ----------------------------------------------------

var (
	afTrEvidenceCmd = regexp.MustCompile("`(moai af evidence [^`]+)`")
	afTrMoveCmd     = regexp.MustCompile("`(moai af move [^`]+)`")
)

// afTrAsPrinted turns a command printed in a message into the arguments of the verb, with the result the measurement
// gave filled in.
func afTrAsPrinted(t *testing.T, a afN3Card, printed, result string) []string {
	t.Helper()
	printed = strings.ReplaceAll(printed, "<pass|fail>", result)
	f := strings.Fields(printed)
	if len(f) < 4 || f[0] != "moai" || f[1] != "af" {
		t.Fatalf("not a moai af command: %q", printed)
	}
	return append(f[2:], "--db", a.db)
}

// TestAFMessage_ reads the text of the in-process command: a built binary frames the message and cuts a long word.
func TestAFMessage_TheRefusalNamesCommandsThatRunAsPrinted(t *testing.T) {
	afSkipWithoutSh(t)
	if os.Getenv("AF_TEST_BIN") != "" {
		t.Skip("reads the message of the in-process command; the binary frames it and may cut a word")
	}
	for _, result := range []string{"pass", "fail"} {
		t.Run("the measurement gave "+result, func(t *testing.T) {
			a := afTrMeasured(t, afTrAlive)
			a.record("gate", a.c1, "fail")
			_, err := a.try("land", "t1", "--gate", "true")
			if err == nil {
				t.Fatal("landed with a failing newest gate run")
			}
			msg := err.Error()
			if strings.Contains(msg, "--result pass") {
				t.Errorf("the message tells the reader to type --result pass: %s", msg)
			}
			for _, want := range []string{"record what the measurement gives", "<pass|fail>", a.c1, "--epoch " + a.e2} {
				if !strings.Contains(msg, want) {
					t.Errorf("the message does not say %q: %s", want, msg)
				}
			}
			m := afTrEvidenceCmd.FindStringSubmatch(msg)
			if m == nil {
				t.Fatalf("the message prints no `moai af evidence ...` command: %s", msg)
			}
			args := afTrAsPrinted(t, a, m[1], result)
			if out, errOut, err := afRunBoth(args...); err != nil {
				t.Fatalf("the printed command %v is refused: %v (%.200s %.200s)", args, err, out, errOut)
			} else if strings.Contains(errOut, "required flag") {
				t.Errorf("the printed command misses a flag: %s", errOut)
			}
			ev, err := a.st.EvidenceFor(t.Context(), "t1")
			if err != nil || len(ev) == 0 || string(ev[len(ev)-1].Result) != result || ev[len(ev)-1].Kind != "gate" {
				t.Fatalf("the printed command recorded %+v %v, want a gate run with the result %s", ev, err, result)
			}
		})
	}
	t.Run("the way out for a holder that is gone runs as printed", func(t *testing.T) {
		a := afTrMeasured(t, afTrReleased)
		a.record("gate", a.c1, "fail")
		afTrLeave(t, a, afTrReleased)
		_, err := a.try("land", "t1", "--gate", "true")
		if err == nil {
			t.Fatal("landed with a failing newest gate run")
		}
		m := afTrMoveCmd.FindStringSubmatch(err.Error())
		if m == nil || m[1] != afTrWayOut("t1", a.e2) {
			t.Fatalf("the message prints %v, want `%s`: %v", m, afTrWayOut("t1", a.e2), err)
		}
		args := afTrAsPrinted(t, a, m[1], "")
		if out, errOut, err := afRunBoth(args...); err != nil {
			t.Fatalf("the printed way out %v is refused: %v (%.200s %.200s)", args, err, out, errOut)
		}
		if got := a.stage(); got != core.StageReviewing {
			t.Errorf("stage %s, want REVIEWING", got)
		}
	})
}
