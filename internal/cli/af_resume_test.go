package cli

// N42 (audit 46dbcce0d4) through the real `moai af` verbs, a real ledger and real git: the third reclaim of a card
// parks it for the operator (reason reclaim-cap), the ledger resolves that park only by an in-place move that carries
// the evidence "resume-guard:reclaim-cap", and no verb passed it. `moai af resume <card> --epoch N --guard
// reclaim-cap --reason <text>` is that verb. It adds no rule: it reads the card, and hands the ledger the in-place
// move with the guard. Without it the existing refusals stay; with it the card can be claimed again, measured in its
// own epoch, taken over and finished on the commit that already landed.

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afN42Parked is a card whose c1 landed at epoch 1, whose holders then died ledger.ReclaimCap times (the reap after
// the last of them parks it): REVIEWING, parked with reclaim-cap, no lease, e2 the epoch of the park.
func afN42Parked(t *testing.T) afN3Card {
	t.Helper()
	a := afN28Landed(t)
	a.verb("move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
	for i := 1; i <= ledger.ReclaimCap; i++ {
		out, errOut, err := afRunBoth("reap", "--db", a.db)
		if err != nil {
			t.Fatalf("reap %d: %v %s %s", i, err, out, errOut)
		}
		var got struct {
			Reclaimed []struct{ Parked bool } `json:"reclaimed"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil || len(got.Reclaimed) != 1 {
			t.Fatalf("reap %d: %q %v", i, out, err)
		}
		if parked := got.Reclaimed[0].Parked; parked != (i == ledger.ReclaimCap) {
			t.Fatalf("reap %d: parked=%v", i, parked)
		}
		if i < ledger.ReclaimCap {
			a = a.claimAgain("holder-" + strconv.Itoa(i))
		}
	}
	return a
}

// claimAgain is a claim by another session, in a card with the epoch that claim returned.
func (a afN3Card) claimAgain(session string) afN3Card {
	a.t.Helper()
	c, err := a.st.Claim(context.Background(), ledger.ClaimRequest{SessionID: session, Worktree: a.repo, PID: os.Getpid(), PIDStart: "start-" + session})
	if err != nil {
		a.t.Fatalf("claim again as %s: %v", session, err)
	}
	a.e2 = strconv.FormatInt(c.Epoch, 10)
	return a
}

func (a afN3Card) card() ledger.Card {
	a.t.Helper()
	cards, err := a.st.Cards(context.Background())
	if err != nil || len(cards) != 1 {
		a.t.Fatalf("cards %+v %v", cards, err)
	}
	return cards[0]
}

func TestAFVerbs_N42TheRefusalsOfAReclaimCapParkStandWithoutTheResumeVerb(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN42Parked(t)
	if c := a.card(); c.Position.Park == nil || c.Position.Park.Reason != core.ReasonReclaimCap {
		t.Fatalf("setup: the card is not parked for reclaim-cap: %+v", c.Position)
	}
	if _, err := a.st.Claim(context.Background(), ledger.ClaimRequest{SessionID: "x", Worktree: a.repo, PID: os.Getpid(), PIDStart: "x"}); err == nil {
		t.Error("a parked card was claimed")
	}
	before := afSnapshot(t, a.st)
	for _, args := range [][]string{
		{"move", "t1", "--epoch", a.e2, "--stage", "REVIEWING"},
		{"move", "t1", "--epoch", a.e2, "--stage", "CANCELLED"},
		{"move", "t1", "--epoch", a.e2, "--stage", "DONE"},
	} {
		if out, err := a.try(args...); err == nil {
			t.Errorf("`af %s` was accepted for a card parked at reclaim-cap: %.200s", strings.Join(args, " "), out)
		}
	}
	if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
		t.Errorf("the refused moves changed the ledger:\nbefore %+v\nafter  %+v", before, after)
	}
}

// Every refused resume leaves the ledger as it was: an empty or missing reason, another guard, a missing guard, a
// stale epoch, and a card that is not parked at all.
func TestAFVerbs_N42TheResumeVerbRefusesWithoutAReasonOrWithAnotherGuardOrEpochAndChangesNothing(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN42Parked(t)
	before := afSnapshot(t, a.st)
	epochPlus := strconv.FormatInt(a.card().Epoch+1, 10)
	epochMinus := strconv.FormatInt(a.card().Epoch-1, 10)
	for _, tc := range []struct {
		name string
		args []string
		want string // what the refusal says: a refusal for another cause (a slip in the fixture) does not count
	}{
		{"no reason", []string{"resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap"}, `"reason" not set`},
		{"an empty reason", []string{"resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", ""}, "--reason must say why"},
		{"a reason of blanks", []string{"resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", "  \t "}, "--reason must say why"},
		{"the guard of another park", []string{"resume", "t1", "--epoch", a.e2, "--guard", "operator-hold", "--reason", "r"}, `--guard "operator-hold" refused`},
		{"the guard fix-limit", []string{"resume", "t1", "--epoch", a.e2, "--guard", "fix-limit", "--reason", "r"}, `--guard "fix-limit" refused`},
		{"the guard spelled as the ledger's key", []string{"resume", "t1", "--epoch", a.e2, "--guard", "resume-guard:reclaim-cap", "--reason", "r"}, `--guard "resume-guard:reclaim-cap" refused`},
		{"an empty guard", []string{"resume", "t1", "--epoch", a.e2, "--guard", "", "--reason", "r"}, `--guard "" refused`},
		{"no guard", []string{"resume", "t1", "--epoch", a.e2, "--reason", "r"}, `"guard" not set`},
		{"a newer epoch", []string{"resume", "t1", "--epoch", epochPlus, "--guard", "reclaim-cap", "--reason", "r"}, "stale epoch"},
		{"an older epoch", []string{"resume", "t1", "--epoch", epochMinus, "--guard", "reclaim-cap", "--reason", "r"}, "stale epoch"},
		{"another card", []string{"resume", "nope", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", "r"}, "does not exist"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := a.try(tc.args...)
			if err == nil {
				t.Fatalf("accepted: %.200s", out)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal %q does not say %q", err, tc.want)
			}
			if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
				t.Errorf("the refusal changed the ledger:\nbefore %+v\nafter  %+v", before, after)
			}
			if c := a.card(); c.Position.Park == nil {
				t.Error("the park is gone after a refusal")
			}
		})
	}
}

// A card that is not parked has nothing to resume, whatever it is told.
func TestAFVerbs_N42TheResumeVerbRefusesACardThatIsNotParkedAndChangesNothing(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, c1, st, epoch := afUnqueuedCard(t)
	afBindFixture(t, st, repo)
	a := afN3Card{t: t, db: db, repo: repo, c1: c1, st: st, e2: strconv.FormatInt(epoch, 10)}
	before := afSnapshot(t, a.st)
	out, err := a.try("resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", "r")
	if err == nil {
		t.Fatalf("a card that is not parked was resumed: %.200s", out)
	}
	if !strings.Contains(err.Error(), "is not parked at reclaim-cap") {
		t.Errorf("the refusal %q does not say that the card is not parked", err)
	}
	if after := afSnapshot(t, a.st); !reflect.DeepEqual(before, after) {
		t.Errorf("the refusal changed the ledger:\nbefore %+v\nafter  %+v", before, after)
	}
}

// The acceptance of N42: from the parked state, the operator's resume, then a claim, a measurement in the new epoch,
// the takeover of the landed commit and DONE, with the commit that landed in the trunk.
func TestAFVerbs_N42TheResumeVerbLetsAReclaimCapParkedCardFinishOnItsLandedCommit(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN42Parked(t)
	parkedEpoch := a.card().Epoch
	const reason = "operator looked: the holders died of the machine rebooting, not of the card"
	out := a.verb("resume", "t1", "--epoch", a.e2, "--guard", "reclaim-cap", "--reason", reason)
	if !strings.Contains(out, `"card_id":"t1"`) {
		t.Errorf("the verb printed %q", out)
	}
	c := a.card()
	if c.Position.Park != nil || c.Position.Stage != core.StageReviewing {
		t.Fatalf("after the resume: %+v", c.Position)
	}
	if c.ReclaimCount != 0 || c.Epoch != parkedEpoch {
		t.Errorf("after the resume the reclaim count is %d and the epoch %d (was %d)", c.ReclaimCount, c.Epoch, parkedEpoch)
	}
	if got, want := afLastTransition(t, a.db), (afTransitionRow{Actor: "operator", Reason: reason, To: "REVIEWING"}); got != want {
		t.Errorf("the resume recorded %+v, want %+v", got, want)
	}
	// it is claimable now, and finishes on the landed commit once it is measured again in its own epoch
	a = a.claimAgain("after-the-resume")
	if a.e2 == strconv.FormatInt(parkedEpoch, 10) {
		t.Fatal("the claim after the resume kept the parked epoch")
	}
	if a.tryDone() == nil {
		t.Error("DONE accepted before the landed commit was measured in the new epoch")
	}
	a.measure(a.c1)
	a.verb("enqueue", "t1", "--epoch", a.e2, "--candidate", a.c1, "--repo", a.repo)
	a.finish()
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
	if !a.inTrunk(a.c1) {
		t.Error("the card is DONE and c1 is not in the trunk")
	}
}

// N45 (part): the help says what the verbs do. `resume` names the park it resolves and the two flags it insists on,
// and `move`, which cannot carry a guard, points at `resume` for the one resume it used to send to the ledger API.
func TestAFVerbs_N45TheHelpOfResumeAndMoveMatchesWhatTheVerbsDo(t *testing.T) {
	af := newAFCmd()
	resume := afVerb(af, "resume")
	if resume == nil {
		t.Fatal("no resume verb")
	}
	for _, want := range []string{"reclaim-cap", "--guard", "--reason", "epoch", "recorded as given"} {
		if !strings.Contains(resume.Long, want) {
			t.Errorf("the help of resume does not say %q", want)
		}
	}
	for _, name := range []string{"guard", "reason", "epoch"} {
		f := resume.Flags().Lookup(name)
		if f == nil {
			t.Errorf("resume has no --%s", name)
			continue
		}
		if _, required := f.Annotations[cobra.BashCompOneRequiredFlag]; !required {
			t.Errorf("--%s is not marked required, and the help says it is", name)
		}
	}
	move := afVerb(af, "move")
	if !strings.Contains(move.Long, `"resume" verb`) {
		t.Errorf("the help of move still sends the reclaim-cap resume to the ledger API: %s", move.Long)
	}
}

type afTransitionRow struct{ Actor, Reason, To string }

// afLastTransition is the newest transition of the ledger at db, read straight from the file.
func afLastTransition(t *testing.T, db string) afTransitionRow {
	t.Helper()
	raw, err := sql.Open("sqlite", "file:"+db)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	var r afTransitionRow
	var reason sql.NullString
	if err := raw.QueryRow(`SELECT actor,reason,to_stage FROM transitions WHERE card_id='t1' ORDER BY seq DESC LIMIT 1`).Scan(&r.Actor, &reason, &r.To); err != nil {
		t.Fatal(err)
	}
	r.Reason = reason.String
	return r
}
