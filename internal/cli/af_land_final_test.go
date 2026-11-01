package cli

// Final audit of 5d90e66fe7 (F1, F5) and lane-19's help-text points, on a real ledger and a real git repository: a
// card under a park does not land; a gate that prints more than the capture limit and exits 0 lands, and one that
// fails keeps its last line; the help and the refusal message say what an operator needs.

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFLand_ACardUnderAParkIsNotLanded(t *testing.T) {
	ctx := context.Background()
	db, repo, cand, st, epoch := afQueuedCard(t)
	before := afGit(t, repo, "rev-parse", "refs/heads/af/trunk")
	park := core.Position{Stage: core.StageMerging, Park: &core.Park{Reason: core.ReasonOperatorHold, Evidence: "operator", Resume: "operator lifts it"}}
	if err := st.Move(ctx, "t1", epoch, park, "operator", "hold",
		ledger.Evidence{Kind: ledger.KindReview, TreeSHA: "t", SubjectSHA: cand, Result: ledger.ResultPass, Guards: []string{core.GuardOperatorHold}}); err != nil {
		t.Skipf("the ledger does not take an operator hold in MERGING: %v", err)
	}
	cmd := newAFCmd()
	cmd.SetArgs([]string{"land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "parked") {
		t.Fatalf("landing a parked card: %v, want a refusal saying it is parked", err)
	}
	if got := afGit(t, repo, "rev-parse", "refs/heads/af/trunk"); got != before {
		t.Errorf("a card under an operator hold reached the trunk: %s -> %s", before, got)
	}
}

func TestAFLand_ALoudPassingGateStillLands(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk",
		"--gate", "head -c 9000000 /dev/zero | tr '\\0' x; exit 0")
	if outcome != "landed" {
		t.Errorf("outcome %q reason %.200q: a gate that exits 0 and prints 9 MB did not land", outcome, reason)
	}
}

func TestAFLand_ALoudFailingGateKeepsItsLastLine(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk",
		"--gate", "head -c 9000000 /dev/zero | tr '\\0' x; echo REAL-\"FAILURE\"-LINE; exit 1")
	if outcome != "failed" {
		t.Fatalf("outcome %q, want failed", outcome)
	}
	// D1: the output is not in the reason; the last line is the end of the log file the reason names
	if strings.Contains(reason, "REAL-FAILURE-LINE") {
		t.Errorf("the gate's output is in the reason (%d bytes): %.200q", len(reason), reason)
	}
	m := afD1LogPath.FindStringSubmatch(reason)
	if m == nil {
		t.Fatalf("the reason names no log file: %.300q", reason)
	}
	if log, err := os.ReadFile(m[1]); err != nil || !strings.Contains(string(log), "REAL-FAILURE-LINE") {
		t.Errorf("the gate's last line is not in the log %s: %v", m[1], err)
	}
}

func TestAFLand_HelpAndRefusalPointTheOperatorOnward(t *testing.T) {
	for _, c := range newAFCmd().Commands() {
		if c.Name() != "land" {
			continue
		}
		for _, want := range []string{
			"Git Bash",                         // what Git Bash does and does not give the gate
			"background process the gate left", // an exit-0 gate's leftovers are killed, not waited for
		} {
			if !strings.Contains(c.Long, want) {
				t.Errorf("the help of `af land` lacks %q:\n%s", want, c.Long)
			}
		}
		if !strings.Contains(c.Long, "does not kill") {
			t.Errorf("the help does not say that Git Bash has no process-group kill:\n%s", c.Long)
		}
		return
	}
	t.Fatal("no `af land` verb")
}

func TestAFLand_TheRefusalOfAWithdrawnCardNamesReconcile(t *testing.T) {
	db, repo, _, _, epoch := afQueuedCard(t)
	runAF(t, "move", "t1", "--db", db, "--epoch", strconv.FormatInt(epoch, 10), "--stage", "REVIEWING", "--actor", "lane", "--reason", "withdrawn")
	cmd := newAFCmd()
	cmd.SetArgs([]string{"land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "moai af reconcile") {
		t.Errorf("the refusal of a withdrawn card: %v, want it to point at `moai af reconcile`", err)
	}
}
