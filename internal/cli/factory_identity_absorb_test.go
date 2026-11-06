// factory_identity_absorb_test.go — SPEC-MCP-NEUTRALITY-001 M2 Part A (card
// t1632): the M1 identity contract applied to the CLI verb layer
// (REQ-MN-003's CLI side). The decide gate requires the POSITIVE operator
// mark on both surfaces; the lane verbs accept an explicit --actor flag with
// the spawn environment as the default only; an explicit run whose row is not
// active refuses stale before any expired row is read.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryDecideGate_PositiveMark pins the CLI decide gate: an unmarked
// session is unknown attribution (env absence grants nothing), a lane session
// keeps the REQ-SD-016 boundary refusal, and MOAI_FACTORY_ROLE=operator
// proceeds. The MCP factory_decide handler runs the same predicate (M1), so
// the two surfaces cannot drift.
func TestFactoryDecideGate_PositiveMark(t *testing.T) {
	build := func(t *testing.T) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
		return root
	}

	t.Run("unmarked session is unknown attribution", func(t *testing.T) {
		sdClearLaneEnv(t)
		root := build(t)
		t.Chdir(root)
		_, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
		if err == nil {
			t.Fatal("cli decide accepted an unmarked session — env absence still grants operator")
		}
		if !strings.Contains(err.Error(), "unknown attribution") || !strings.Contains(err.Error(), "operator") {
			t.Errorf("decide refusal = %v, want the unknown-attribution text naming the operator mark", err)
		}
	})

	t.Run("lane session keeps the boundary refusal", func(t *testing.T) {
		sdClearLaneEnv(t)
		root := build(t)
		t.Chdir(root)
		sdLaneEnv(t, "lane-1", "")
		_, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "lane boundary") {
			t.Fatalf("cli lane decide: err = %v, want the boundary refusal", err)
		}
	})

	t.Run("the operator mark proceeds", func(t *testing.T) {
		sdClearLaneEnv(t)
		root := build(t)
		t.Chdir(root)
		t.Setenv(config.EnvFactoryRole, "operator")
		if _, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun); err != nil {
			t.Fatalf("cli decide under the operator mark: %v", err)
		}
	})
}

// TestFactoryActorFlag pins the CLI actor argument (REQ-MN-003): the explicit
// flag decides whose label acts, the environment is the default only. The
// refusal texts stay the CLI ones the MCP parity tests pin.
func TestFactoryActorFlag(t *testing.T) {
	build := func(t *testing.T) (string, string) {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		sdSeedRun(t, root, fcRun, "active")
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "-A")
		fcGit(t, root, "commit", "-q", "-m", "plan artifact")
		sha := fcGit(t, root, "rev-parse", "HEAD")
		fcPlace(t, root, homestate.Card{
			CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
			WorktreePath:   root,
		})
		return root, sha
	}

	t.Run("stage --actor lane-2 on a lane-1 card is refused by the holder edge", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "") // the spawn is recorded as lane-1
		t.Chdir(root)
		_, _, err := runFactory(t, "stage", "t1", "plan-audit", sha+":plan.md", "--run", fcRun, "--actor", "lane-2")
		if err == nil {
			t.Fatal("cli stage --actor lane-2 was accepted — the actor flag is ignored")
		}
		if !strings.Contains(err.Error(), "does not hold the lease") || !strings.Contains(err.Error(), "lane-2") {
			t.Errorf("cli stage --actor lane-2: err = %v, want the F1 holder refusal naming lane-2", err)
		}
	})

	t.Run("stage --actor lane-1 proceeds", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if _, _, err := runFactory(t, "stage", "t1", "plan-audit", sha+":plan.md", "--run", fcRun, "--actor", "lane-1"); err != nil {
			t.Fatalf("cli stage --actor lane-1: %v", err)
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardPlanAudit {
			t.Errorf("t1 = %s, want plan-audit", c.State)
		}
	})
}

// TestFactoryStaleRunRefused pins the CLI stale-run refusal (REQ-MN-003): a
// run whose row exists but is not active is refused before any lease, stage,
// or evidence state is read from its expired rows.
func TestFactoryStaleRunRefused(t *testing.T) {
	root, _ := sdMoaiFixture(t)
	sdSeedRun(t, root, fcRun, "active")
	sdSeedRun(t, root, "run-done", "completed")
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root, homestate.Card{
		RunID: "run-done", CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath:   root,
	})
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	_, _, err := runFactory(t, "stage", "t1", "plan-audit", "--run", "run-done")
	if err == nil {
		t.Fatal("cli stage on a completed run was accepted — expired rows are being read")
	}
	if !strings.Contains(err.Error(), "not the currently active factory run") {
		t.Errorf("cli stale-run refusal = %v, want the not-active refusal", err)
	}
}
