package cli

// What is left of the transition group of the repair after the final audit of a36ea8c368: the E6 test (a candidate that
// is the trunk's own commit lands as already in the trunk) and the fixture the landing-window tests share. The tests begin
// with TestAFTransition_ so that one anchored run picks them, and they go through afRunBoth.

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// afTrHolder is what became of the card's holder.
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

// afTrRun runs a verb that has no positional argument, with the ledger at the end.
func afTrRun(a afN3Card, args ...string) (string, error) {
	a.t.Helper()
	out, errOut, err := afRunBoth(append(append([]string{}, args...), "--db", a.db)...)
	if err != nil {
		return out + errOut, err
	}
	return out, nil
}

// ---- N63 ----------------------------------------------------------------------------------------------------------------

// ---- N66 ----------------------------------------------------------------------------------------------------------------

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
