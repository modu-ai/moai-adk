package cli

// N3 (audits a1ec370195 and 22a0afd247, reopened as a blocker by the operator) through the real `moai af` verbs, a real
// ledger and real git: c1 lands at epoch 1, the holder takes the card out of MERGING (back to REVIEWING, or on through
// BUILDING), lets go, and a new session claims the card at epoch 2. A landing is a fact about the trunk, not about
// the generation of the card that caused it, so the new holder can finish the card on it: by measuring c1 again in its
// own epoch (the landing is taken over) or by measuring a new commit c2 in its own epoch (the row is replaced and c2
// lands). Without a measurement in the new epoch the offer is refused and nothing changes. The trunk is the dedicated
// ref and the project directory is the repository, as the verbs require.

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

// afN3Card is a card whose c1 landed at epoch 1 through the verbs, was withdrawn (to REVIEWING, and on to BUILDING
// when rework is set), released, and claimed again by another session at epoch 2.
type afN3Card struct {
	t      *testing.T
	db     string
	repo   string
	c1     string
	st     *ledger.Store
	e2     string
	rework bool
}

func (a afN3Card) verb(args ...string) string {
	a.t.Helper()
	out, errOut, err := afRunBoth(append(args[:1:1], append([]string{args[1], "--db", a.db}, args[2:]...)...)...)
	if err != nil {
		a.t.Fatalf("moai af %s: %v\nstdout %.300s\nstderr %.300s", strings.Join(args, " "), err, out, errOut)
	}
	return out
}

func (a afN3Card) try(args ...string) (string, error) {
	a.t.Helper()
	out, errOut, err := afRunBoth(append(args[:1:1], append([]string{args[1], "--db", a.db}, args[2:]...)...)...)
	return out + errOut, err
}

func (a afN3Card) measure(sha string) {
	a.t.Helper()
	for _, kind := range []string{"green", "gate"} {
		a.verb("evidence", "t1", "--epoch", a.e2, "--kind", kind, "--subject-sha", sha, "--result", "pass", "--repo", a.repo)
	}
}

func (a afN3Card) newCommit() string {
	a.t.Helper()
	afGit(a.t, a.repo, "checkout", "-q", "-b", "c2", a.c1)
	if err := os.WriteFile(a.repo+"/fix.txt", []byte("fix\n"), 0o644); err != nil {
		a.t.Fatal(err)
	}
	afGit(a.t, a.repo, "add", "fix.txt")
	afGit(a.t, a.repo, "commit", "-q", "-m", "card t1: the fix")
	c2 := afGit(a.t, a.repo, "rev-parse", "HEAD")
	afGit(a.t, a.repo, "checkout", "-q", "work")
	return c2
}

// handOff takes a card that was taken back through BUILDING to REVIEWING, where it can be offered.
func (a afN3Card) handOff() {
	a.t.Helper()
	if a.rework {
		a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
	}
}

func (a afN3Card) finish() {
	a.t.Helper()
	a.verb("move", "t1", "--epoch", a.e2, "--stage", "MERGING", "--remote-integrated")
	a.verb("move", "t1", "--epoch", a.e2, "--stage", "DONE")
}

func (a afN3Card) stage() core.Stage {
	a.t.Helper()
	cards, err := a.st.Cards(context.Background())
	if err != nil || len(cards) != 1 {
		a.t.Fatalf("cards %+v %v", cards, err)
	}
	return cards[0].Position.Stage
}

func (a afN3Card) inTrunk(sha string) bool {
	return exec.Command("git", "-C", a.repo, "merge-base", "--is-ancestor", sha, mergeq.DefaultTrunk).Run() == nil
}

func afN3Reclaimed(t *testing.T, rework bool) afN3Card {
	t.Helper()
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	afBindFixture(t, st, repo)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, rework: rework}
	e1 := strconv.FormatInt(epoch, 10)
	for _, kind := range []string{"green", "gate"} {
		a.verb("evidence", "t1", "--epoch", e1, "--kind", kind, "--subject-sha", c1, "--result", "pass", "--repo", repo)
	}
	a.verb("enqueue", "t1", "--epoch", e1, "--candidate", c1, "--repo", repo)
	if out := a.verb("land", "t1", "--repo", repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("c1 did not land: %s", out)
	}
	if !a.inTrunk(c1) {
		t.Fatal("c1 landed but is not in the trunk")
	}
	a.verb("move", "t1", "--epoch", e1, "--stage", "REVIEWING")
	if rework {
		a.verb("move", "t1", "--epoch", e1, "--stage", "BUILDING", "--sub", "IMPLEMENT")
	}
	a.verb("release", "t1", "--epoch", e1)
	c, err := st.Claim(context.Background(), ledger.ClaimRequest{SessionID: "next", Worktree: repo, PID: os.Getpid(), PIDStart: "start-next"})
	if err != nil || c.Epoch != epoch+1 {
		t.Fatalf("claimed again: %+v %v", c, err)
	}
	a.e2 = strconv.FormatInt(c.Epoch, 10)
	return a
}

func afN3Name(rework bool) string {
	if rework {
		return "taken back through BUILDING"
	}
	return "withdrawn to REVIEWING"
}

func TestAFVerbs_N3AReclaimedCardFinishesOnTheLandedCommitOnceItMeasuredItAgain(t *testing.T) {
	afSkipWithoutSh(t)
	for _, rework := range []bool{false, true} {
		t.Run(afN3Name(rework), func(t *testing.T) {
			a := afN3Reclaimed(t, rework)
			a.measure(a.c1)
			a.handOff()
			a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
			a.finish()
			if got := a.stage(); got != core.StageDone {
				t.Errorf("stage %s, want DONE", got)
			}
			if !a.inTrunk(a.c1) {
				t.Error("the card is DONE and c1 is not in the trunk")
			}
		})
	}
}

func TestAFVerbs_N3AReclaimedCardFinishesOnANewCommitOnceItMeasuredItAndLandedIt(t *testing.T) {
	afSkipWithoutSh(t)
	for _, rework := range []bool{false, true} {
		t.Run(afN3Name(rework), func(t *testing.T) {
			a := afN3Reclaimed(t, rework)
			c2 := a.newCommit()
			a.measure(c2)
			a.handOff()
			a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", c2, "--repo", a.repo)
			if out := a.verb("land", "t1", "--repo", a.repo, "--gate", "true"); !strings.Contains(out, `"outcome":"landed"`) {
				t.Fatalf("c2 did not land: %s", out)
			}
			a.finish()
			if got := a.stage(); got != core.StageDone {
				t.Errorf("stage %s, want DONE", got)
			}
			if !a.inTrunk(a.c1) || !a.inTrunk(c2) {
				t.Errorf("the card is DONE and the trunk holds c1=%v c2=%v", a.inTrunk(a.c1), a.inTrunk(c2))
			}
		})
	}
}

// Nothing measured in the new epoch: the evidence of epoch 1 does not count, the offer is refused, and the row, the
// stage and everything else the ledger holds stay as they were.
func TestAFVerbs_N3AReclaimedCardThatMeasuredNothingInItsOwnEpochCannotOfferAnything(t *testing.T) {
	afSkipWithoutSh(t)
	for _, rework := range []bool{false, true} {
		t.Run(afN3Name(rework), func(t *testing.T) {
			a := afN3Reclaimed(t, rework)
			c2 := a.newCommit()
			if rework {
				// A card without a receipt of its own may not hand off (the ledger says so); it is offered from where it is.
				_, _ = a.try("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
			}
			before := afSnapshot(t, a.st)
			stageBefore := a.stage()
			for _, sha := range []string{a.c1, c2} {
				out, err := a.try("enqueue", "t1", "--epoch", a.e2, "--candidate", sha, "--repo", a.repo)
				if err == nil {
					t.Errorf("enqueue of %.8s with nothing measured in the new epoch was accepted: %.200s", sha, out)
				}
				if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
					t.Fatalf("the refused enqueue of %.8s changed the ledger:\nbefore %+v\nafter  %+v", sha, before, after)
				}
			}
			if got := a.stage(); got != stageBefore {
				t.Errorf("stage %s -> %s", stageBefore, got)
			}
			// the card cannot be finished without the measurement either
			if _, err := a.try("move", "t1", "--epoch", a.e2, "--stage", "DONE"); err == nil {
				t.Error("the card moved to DONE with nothing measured in its own epoch")
			}
		})
	}
}
