package cli

// Card t2104 (GitHub #1938): a card left behind in a dead run with an EXPIRED
// lease has a formal path into a live run — `moai factory runs --retire <dead>
// --rebind-to <live>` — and `assign --reassign` is deliberately scoped to the
// card's current dispatch run. These tests pin both halves so the documented
// route cannot regress into a gap: (a) the expired leased row moves intact, (b)
// the earlier-assign workaround (a second row in the live run) is refused by
// the consolidation with the diverging columns named, and leaves the source row
// where it was. Fixtures are isolated sandboxes.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func expiredLeaseRebindFixture(t *testing.T, assignedInLive bool) string {
	t.Helper()
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "run-src", deadPID, deadStart)
	seedRun(t, root, "run-tgt", os.Getpid(), homestate.CurrentProcessFingerprint())
	fcPlace(t, root, homestate.Card{
		RunID: "run-src", CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-9",
		LeaseHolder: "lane-9", LeaseExpiresAt: "2026-09-20T01:00:00Z", Version: 3,
	})
	if assignedInLive {
		fcPlace(t, root, homestate.Card{RunID: "run-tgt", CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 1})
	}
	return root
}

func TestFactoryRunsRebindMovesExpiredLeasedRow(t *testing.T) {
	root := expiredLeaseRebindFixture(t, false)
	out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
	if err != nil {
		t.Fatalf("rebind of an expired leased row: %v\n%s", err, out)
	}
	if !strings.Contains(out, "moved cards: t1") {
		t.Errorf("output does not report t1 as moved:\n%s", out)
	}
	if got := lcQuery(t, root, `SELECT group_concat(run_id||':'||state||':'||owner_label, ' ; ') FROM cards WHERE card_id='t1'`); got != "run-tgt:leased:lane-9" {
		t.Errorf("card rows = %q, want the single row run-tgt:leased:lane-9", got)
	}
	if got := runStatusOf(t, root, "run-src"); got != "retired" {
		t.Errorf("source status = %q, want retired", got)
	}
}

func TestFactoryRunsRebindRefusesCardAlreadyAssignedInLiveRun(t *testing.T) {
	root := expiredLeaseRebindFixture(t, true)
	before := lcDump(t, root)
	out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
	if err == nil {
		t.Fatalf("consolidation accepted a diverging live row:\n%s", out)
	}
	for _, want := range []string{"t1", "run-tgt", "lease_holder"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name %q", err.Error(), want)
		}
	}
	if after := lcDump(t, root); after != before {
		t.Errorf("the refusal changed state")
	}
}
