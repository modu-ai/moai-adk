package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-FDA-016 — a lane may run exactly `factory decide <card> --gate kickoff
// --choice approve --decider audit` for a card whose record owner is the
// lane; every other lane decide stays refused at the lane boundary.
func TestFDA_LaneAuditDecideAdmission(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardKickoff, OwnerLabel: "lane-1", DecisionGate: homestate.DecisionGateKickoff},
		homestate.Card{CardID: "t2", State: homestate.CardKickoff, OwnerLabel: "lane-2", DecisionGate: homestate.DecisionGateKickoff},
	)
	sdLaneEnv(t, "lane-1", "")

	out, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", fcRun)
	if err != nil && strings.Contains(err.Error(), factoryLaneBoundarySentinel) {
		t.Fatalf("own-card audit decide hit the lane boundary: %v", err)
	}
	if !strings.Contains(out, "audit kickoff refused") {
		t.Fatalf("own-card audit decide did not reach the evidence check: out=%q err=%v", out, err)
	}
	for name, args := range map[string][]string{
		"another lane's card": {"decide", "t2", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", fcRun},
		"human decider":       {"decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun},
		"audit reject":        {"decide", "t1", "--gate", "kickoff", "--choice", "reject", "--decider", "audit", "--run", fcRun},
		"push":                {"decide", "t1", "--gate", "push", "--decider", "audit", "--run", fcRun},
		"resume":              {"decide", "t1", "--choice", "resume", "--run", fcRun},
	} {
		if _, _, err := runFactory(t, args...); err == nil || !strings.Contains(err.Error(), "refused") {
			t.Errorf("%s: err=%v, want a refusal", name, err)
		}
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardKickoff {
		t.Fatalf("t1 moved to %s", c.State)
	}
}

// AC-FDA-014 — the queue hold arm is read at the decision point.
func TestFDA_AuditDecideReadsTheQueueHold(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateHold)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, OwnerLabel: "worker-1", DecisionGate: homestate.DecisionGateKickoff})
	sdClearLaneEnv(t)
	out, _, _ := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--decider", "audit", "--run", fcRun)
	if !strings.Contains(out, "on hold") {
		t.Fatalf("held queue item: out=%q, want a hold refusal", out)
	}
}
