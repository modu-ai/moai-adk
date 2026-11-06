package cli

// Card t2030: a run-less renew whose run cannot be decided refuses, names --run, and writes nothing.
// "Cannot be decided" is two active runs whose owners are both live and a lane with no roster binding: no
// dead owner to set aside, no binding to follow. The refusal text is the renew verb's own (the resolver in
// factorymsg only reports the candidates); the snapshot helpers are those of the dead-owner contract.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestRenewWithoutRunUndecidedRefusesNamingRunAndWritesNothing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root, _ := fcFixture(t)
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	renewMoveAssertIsolated(t, root)
	seedLiveRun(t, root, "run-a")
	seedLiveRun(t, root, "run-b")
	sdRegisterLane(t, root, "lane-1")
	live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
	fcPlace(t, root, homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root})
	sdLaneEnv(t, "lane-1", "")

	before := renewMoveSnapshot(t, root)
	out, _, err := runFactory(t, "renew", "t1", "--json")
	after := renewMoveSnapshot(t, root)
	renewMoveAssertIsolated(t, root)

	if err == nil {
		t.Fatalf("renew was expected to be refused, got stdout %q", strings.TrimSpace(out))
	}
	if !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") || !strings.Contains(err.Error(), "pass --run") {
		t.Errorf("refusal %q does not name the ambiguity and --run", err)
	}
	if diff := renewMoveAnyDiff(before, after); len(diff) > 0 {
		t.Errorf("a refused bare renew changed %d rows:\n  %s", len(diff), strings.Join(renewMoveBrief2(diff), "\n  "))
	}
}
