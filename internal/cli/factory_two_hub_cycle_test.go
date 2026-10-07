package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Explicit t1->t3 and inferred t2->t1, t3->t2 must not close a wait cycle.
func TestReviewTwoHubCombinedCycle(t *testing.T) {
	for _, nominated := range []bool{false, true} {
		name := "normal"
		if nominated {
			name = "nominated"
		}
		t.Run(name, func(t *testing.T) {
			root, store := fcFixture(t)
			fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
			for _, id := range []string{"t1", "t2", "t3"} {
				fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
			}
			fbSeedFiles(t, store, "t1", "internal/template/catalog.yaml")
			fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml", "internal/config/defaults.go")
			fbSeedFiles(t, store, "t3", "internal/config/defaults.go")
			sdRegisterLane(t, root, "lane-1")
			t.Chdir(root)
			for _, assignment := range []struct{ id, after string }{{"t2", ""}, {"t3", ""}, {"t1", "t3"}} {
				if _, _, err := runFactory(t, "assign", assignment.id, "--after", assignment.after, "--run", fcRun); err != nil {
					t.Fatalf("assign %s: %v", assignment.id, err)
				}
			}
			for _, want := range []string{"t3", "t1", "t2"} {
				sdLaneEnv(t, "lane-1", "")
				if nominated {
					if _, _, err := runFactory(t, "next", "--card", want, "--run", fcRun); err != nil {
						t.Fatalf("nominated %s: %v", want, err)
					}
				} else if got := fbLeasedCard(t, root, "lane-1"); got != want {
					t.Fatalf("normal lease=%q, want %s; combined hub/after order must make progress", got, want)
				}
				if c := fcCard(t, root, want); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
					t.Fatalf("%s state=%s holder=%q", want, c.State, c.LeaseHolder)
				}
				fcSetCardState(t, root, want, homestate.CardMergedLocal)
			}
		})
	}
}

// Generation and selection must agree even before the leading card has a row.
func TestReviewTwoHubGenerationAndInflight(t *testing.T) {
	_, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA, t1533HubB)
	fbSeedFiles(t, store, "t3", t1533HubB)
	queue, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	rows := []homestate.Card{
		{CardID: "t1", State: homestate.CardPicked, HintAfter: "t3"},
		{CardID: "t2", State: homestate.CardPicked},
	}
	if fields := factoryHubChainFields(queue, rows, "t3", nil); fields.HintAfter != nil {
		t.Fatalf("generation closes combined cycle: t3 after %s", *fields.HintAfter)
	}
	// Actual in-flight t2 must still hold t3, despite its earlier queue
	// position being behind the explicit follower t1.
	rows[1].State = homestate.CardLeased
	rows = append(rows, homestate.Card{CardID: "t3", State: homestate.CardPicked})
	if pred, wait := factoryHubWaitUnmerged(queue, rows, factoryMergedCards(rows), "t3"); !wait || pred != "t2" {
		t.Fatalf("in-flight predecessor lost: wait=(%q,%v), want (t2,true)", pred, wait)
	}
	rows[1].State = homestate.CardMergedLocal
	if pred, wait := factoryHubWaitUnmerged(queue, rows, factoryMergedCards(rows), "t3"); wait {
		t.Fatalf("merged predecessor still blocks: %s", pred)
	}
}
