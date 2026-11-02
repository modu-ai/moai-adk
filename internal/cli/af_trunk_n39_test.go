package cli

// N39 (audit of 46dbcce0d4, and the native review of the same commit) through the real `moai af` verbs, a real ledger
// and real git. A landed queue row says "merged" without asking which trunk: a ledger written before the verbs refused
// another trunk, or a trunk ref moved back after the landing, leaves a row whose commit is not in refs/heads/af/trunk.
// The new holder of such a card must not finish it on that row. `af enqueue` asks the repository whether the candidate
// is in the dedicated trunk and hands the answer to the ledger (a take-over only when it is; otherwise the row starts
// again as waiting, to be landed for real), and the move to DONE asks the same of a landed row. When git cannot answer
// the verb refuses: it does not guess.

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

// row returns the card's only queue row.
func (a afN3Card) row() ledger.QueueItem {
	a.t.Helper()
	rows, err := a.st.Queue(context.Background())
	if err != nil || len(rows) != 1 {
		a.t.Fatalf("queue: %+v %v", rows, err)
	}
	return rows[0]
}

// The control: a landed row whose commit is in the dedicated trunk is still taken over (the N3 behaviour), and the
// row stays landed.
func TestAFVerbs_N39ALandedRowWhoseCommitIsInTheDedicatedTrunkIsTakenOver(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN3Reclaimed(t, false)
	a.measure(a.c1)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	if got := a.row().State; got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed (taken over)", got)
	}
	a.finish()
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
}

// The move to DONE asks the same of a landed row: the commit that landed left the dedicated trunk, the card is
// refused, the ledger unchanged, and the card can go back to be offered and landed again.
func TestAFVerbs_N39TheMoveToDoneAsksWhetherTheLandedCommitIsInTheTrunk(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN28Landed(t)
	afGit(t, a.repo, "update-ref", mergeq.DefaultTrunk, afGit(t, a.repo, "rev-parse", mergeq.DefaultTrunk+"~1"))
	a.verb("move", "t1", "--epoch", a.e2, "--stage", "MERGING", "--remote-integrated")
	before := afSnapshot(t, a.st)
	out, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "DONE")
	if err == nil {
		t.Fatalf("DONE was accepted with the landed commit outside the dedicated trunk: %.200s", out)
	}
	if !strings.Contains(err.Error(), "af/trunk") || !strings.Contains(err.Error(), a.c1) {
		t.Errorf("the refusal does not name the trunk and the commit: %v", err)
	}
	if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
		t.Errorf("the refused move changed the ledger:\nbefore %+v\nafter  %+v", before, after)
	}
	// the control: with the commit back in the trunk the same move is accepted
	afGit(t, a.repo, "update-ref", mergeq.DefaultTrunk, afGit(t, a.repo, "rev-parse", "refs/heads/c1"))
	if !a.inTrunk(a.c1) {
		t.Fatal("setup: c1 is not in the trunk after the ref was set to it")
	}
	if err := a.tryDone(); err != nil {
		t.Fatalf("DONE refused with the commit in the trunk: %v", err)
	}
}
