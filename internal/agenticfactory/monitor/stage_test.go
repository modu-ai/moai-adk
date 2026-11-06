package monitor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func stageRow(cr ClaimRow) Row {
	cr.Issue, cr.HeadSHA = 7, head
	c := collector(fakeClaims{rows: map[int]ClaimRow{7: cr}}, fakeRemote{}, fakeLedger{})
	return c.Collect(context.Background(), []int{7})[0]
}

func TestTheRowShowsTheS46StageAndSubStageFromTheClaim(t *testing.T) {
	r := stageRow(ClaimRow{State: "implementing", Stage: string(core.StageBuilding), SubStage: string(core.SubFix)})
	if r.Stage.Value != "BUILDING" || r.SubStage.Value != "FIX" || r.State.Value != "implementing" {
		t.Errorf("stage=%q sub=%q state=%q", r.Stage.Value, r.SubStage.Value, r.State.Value)
	}
	if !r.Stage.Confirmed || !r.SubStage.Confirmed {
		t.Error("fields read from the claim must be confirmed")
	}
}

// PARKED is an overlay, not a stage: the stage stays and the park is shown beside it.
func TestAParkedPositionKeepsItsStageAndShowsTheOverlay(t *testing.T) {
	r := stageRow(ClaimRow{
		Stage: string(core.StageReviewing),
		Park:  &ParkInfo{Reason: "rate-limited", Evidence: "429 with retry-after 30", Resume: "after the reset"},
	})
	if r.Stage.Value != "REVIEWING" {
		t.Errorf("stage = %q: a park must not replace the stage", r.Stage.Value)
	}
	p := r.Park.Value
	if p == nil || p.Reason != "rate-limited" || p.Evidence == "" || p.Resume != "after the reset" {
		t.Errorf("park = %+v", p)
	}
	if clear := stageRow(ClaimRow{Stage: string(core.StageReviewing)}); clear.Park.Value != nil {
		t.Errorf("a record with no park shows one: %+v", clear.Park.Value)
	}
}

func TestTheGateAndTheLastRefusalAreShownAndDoNotMoveTheStage(t *testing.T) {
	r := stageRow(ClaimRow{
		Stage:   string(core.StageReviewing),
		Gate:    &GateInfo{Reason: "check-not-passed", Detail: "build failed on the latest head", Resume: "re-run on the new head", Head: head},
		Refusal: &RefusalInfo{Reason: "not-issue-author", Detail: "edit refused"},
	})
	if r.Stage.Value != "REVIEWING" {
		t.Errorf("stage = %q", r.Stage.Value)
	}
	g, f := r.Gate.Value, r.Refusal.Value
	if g == nil || g.Reason != "check-not-passed" || g.Head != head || g.Resume == "" {
		t.Errorf("gate = %+v", g)
	}
	if f == nil || f.Reason != "not-issue-author" {
		t.Errorf("refusal = %+v", f)
	}
}

func TestOverlayGateAndRefusalTextIsMasked(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	r := stageRow(ClaimRow{
		Stage:   string(core.StageMerging),
		Park:    &ParkInfo{Reason: "auth-failed", Evidence: "token " + secret + " rejected", Resume: "rotate the credential"},
		Gate:    &GateInfo{Reason: "check-not-passed", Detail: "GITHUB_TOKEN=abc123def456ghi789", Resume: "fix"},
		Refusal: &RefusalInfo{Reason: "permission-denied", Detail: "Authorization: Bearer abc.def.ghi"},
	})
	for _, text := range []string{r.Park.Value.Evidence, r.Gate.Value.Detail, r.Refusal.Value.Detail} {
		for _, bad := range []string{secret, "abc123def456ghi789", "abc.def.ghi"} {
			if strings.Contains(text, bad) {
				t.Errorf("secret %q survived in %q", bad, text)
			}
		}
	}
}

func TestAnUnreadableClaimLeavesTheNewFieldsUnconfirmed(t *testing.T) {
	c := collector(fakeClaims{err: errors.New("store unreadable")}, fakeRemote{}, fakeLedger{})
	r := c.Collect(context.Background(), []int{7})[0]
	if r.SubStage.Confirmed || r.Park.Confirmed || r.Gate.Confirmed || r.Refusal.Confirmed || r.Park.Value != nil {
		t.Errorf("unconfirmed expected: %+v %+v %+v %+v", r.SubStage, r.Park, r.Gate, r.Refusal)
	}
}

// Completed follows the S46 stage, not the old state string.
func TestCompletedNeedsTheDoneStageNotTheOldStateString(t *testing.T) {
	merged := PRInfo{Number: 55, Merged: true, IssueClosed: true, HeadSHA: head}
	for name, cr := range map[string]ClaimRow{
		"old state done but stage still merging": {State: "done", Stage: string(core.StageMerging)},
		"stage done":                             {State: "landing", Stage: string(core.StageDone)},
	} {
		cr.Issue, cr.HeadSHA = 7, head
		c := collector(fakeClaims{rows: map[int]ClaimRow{7: cr}}, fakeRemote{}, fakeLedger{})
		c.PRs = fakePRs{info: map[int]PRInfo{7: merged}}
		got := c.Collect(context.Background(), []int{7})[0].Completed
		if want := name == "stage done"; got != want {
			t.Errorf("%s: completed = %v, want %v", name, got, want)
		}
	}
}
