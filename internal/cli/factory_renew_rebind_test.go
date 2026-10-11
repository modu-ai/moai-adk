// factory_renew_rebind_test.go — SPEC-FACTORY-LEASE-RENEW-001 roster re-check inside the renew write
// (sync-audit F14, card t1663; codex P1). The renewal resolves its run from the lane's roster binding, then
// the lease write runs in a later transaction. A re-registration of the label between those two steps must
// refuse the write, naming --run, with the lease unchanged. The interleaving runs through
// factoryRenewBeforeWrite, a nil-by-default seam these tests set, so the order is fixed without sleeps or
// goroutines.
package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// setFactoryRenewBeforeWrite installs fn as the pre-write seam for the rest of the test.
func setFactoryRenewBeforeWrite(t *testing.T, fn func()) {
	t.Helper()
	prev := factoryRenewBeforeWrite
	factoryRenewBeforeWrite = fn
	t.Cleanup(func() { factoryRenewBeforeWrite = prev })
}

// assertRebindRefused checks the F14 contract for one surface: the renewal refused with a detail that names
// --run (the verb also exits non-zero), and run-a's lease expiry did not move.
func assertRebindRefused(t *testing.T, surface string, wire factoryRenewWire, callErr error, before, after string) {
	t.Helper()
	if wire.Result == "renewed" {
		t.Errorf("renew with no --run renewed (expires %s) although the roster was re-registered to run-b before the write", wire.ExpiresAt)
	} else {
		if wire.Result != "error" || !strings.Contains(wire.Detail, "--run") {
			t.Errorf("refusal does not name --run: result=%s detail=%q err=%v", wire.Result, wire.Detail, callErr)
		}
		if surface == "verb" && callErr == nil {
			t.Errorf("verb refusal exited zero, want a non-zero exit")
		}
	}
	if after != before {
		t.Errorf("run-a lease expiry moved %s → %s; a refused renewal must not write", before, after)
	}
}

// TestRenewRebindBeforeWriteAmbiguousRefuses is sync-audit F14, the ambiguous set. Two runs are active. The
// lane is bound to run-a and holds card t1 there, so the resolver picks run-a. Before the write, the label is
// re-registered to run-b. The renewal must refuse, name --run, and leave run-a's lease unchanged.
func TestRenewRebindBeforeWriteAmbiguousRefuses(t *testing.T) {
	for _, surface := range []string{"verb", "MCP tool"} {
		t.Run(surface, func(t *testing.T) {
			root, _ := sdMoaiFixture(t)
			seedLiveRun(t, root, "run-a")
			seedLiveRun(t, root, "run-b")
			sdRegisterLane(t, root, "lane-1")
			factoryRenewBindLane(t, root, "lane-1", "run-a")
			live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
			fcPlace(t, root,
				homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
					OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
				homestate.Card{RunID: "run-b", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
					OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live, WorktreePath: root},
			)
			sdLaneEnv(t, "lane-1", "")
			setFactoryRenewBeforeWrite(t, func() { factoryRenewBindLane(t, root, "lane-1", "run-b") })
			before := factoryRenewCardIn(t, root, "run-a", "t1").LeaseExpiresAt
			wire, callErr := factoryRenewCardNoRun(t, root, surface, "t1")
			after := factoryRenewCardIn(t, root, "run-a", "t1").LeaseExpiresAt
			assertRebindRefused(t, surface, wire, callErr, before, after)
		})
	}
}

// TestRenewRebindBeforeWriteSoleRunRefuses is sync-audit F14, the sole-run branch. One run, run-a, is active,
// and the lane is unbound or bound to run-a when the renewal resolves run-a. Before the write, run-b starts
// and the label is re-registered to run-b. The renewal must refuse, name --run, and leave run-a's lease
// unchanged.
func TestRenewRebindBeforeWriteSoleRunRefuses(t *testing.T) {
	for _, start := range []struct{ name, bound string }{
		{"unbound at resolution", ""},
		{"bound to run-a at resolution", "run-a"},
	} {
		for _, surface := range []string{"verb", "MCP tool"} {
			t.Run(start.name+"/"+surface, func(t *testing.T) {
				root, _ := sdMoaiFixture(t)
				seedLiveRun(t, root, "run-a")
				sdRegisterLane(t, root, "lane-1")
				if start.bound != "" {
					factoryRenewBindLane(t, root, "lane-1", start.bound)
				}
				live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
				fcPlace(t, root,
					homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
						OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
				)
				sdLaneEnv(t, "lane-1", "")
				setFactoryRenewBeforeWrite(t, func() {
					seedLiveRun(t, root, "run-b")
					factoryRenewBindLane(t, root, "lane-1", "run-b")
				})
				before := factoryRenewCardIn(t, root, "run-a", "t1").LeaseExpiresAt
				wire, callErr := factoryRenewCardNoRun(t, root, surface, "t1")
				after := factoryRenewCardIn(t, root, "run-a", "t1").LeaseExpiresAt
				assertRebindRefused(t, surface, wire, callErr, before, after)
			})
		}
	}
}
