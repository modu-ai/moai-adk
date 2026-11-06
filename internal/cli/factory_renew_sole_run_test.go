// factory_renew_sole_run_test.go — SPEC-FACTORY-LEASE-RENEW-001 run resolution of the renew verb
// (card t1663). It pins the sole-run answer against the roster (sync-audit F11) and every return of
// resolveFactoryRenewRun with the exact message it produces (sync-audit F12). The verb and the
// factory_renew tool share resolveFactoryRenewRun, so the sole-run cases run on both surfaces.
package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryRenewSoleRunExplicitMismatchRefuses is sync-audit F11 (card t1663; codex P1 in sync audit
// FAIL #3). With one active run, resolveFactoryRenewRun answered from the run list alone and never read
// the roster, so a lane that the roster binds to another run renewed its lease in the sole active run. A
// renewal with no --run refuses, naming --run, when the roster binds the lane to a different run, active
// or not, and writes nothing. The other arms are positive controls that keep the sole-run answer: a lane
// with an empty roster binding, a lane with no roster row, and a lane bound to the sole run.
func TestFactoryRenewSoleRunExplicitMismatchRefuses(t *testing.T) {
	// fixture: run-a is the one active run; run-b is recorded retired, so it is not active. Lane-1 holds
	// card t1 in run-a and in run-b, and the roster binds lane-1 to bound. Lane-3 holds card t3 in run-a
	// and has no roster row: it is never registered.
	fixture := func(t *testing.T, bound string) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		seedLiveRun(t, root, "run-a")
		fcPlaceRun(t, root, "run-b", "retired", "2026-09-25T00:00:00Z")
		sdRegisterLane(t, root, "lane-1")
		if bound != "" {
			factoryRenewBindLane(t, root, "lane-1", bound)
		}
		live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
		fcPlace(t, root,
			homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
			homestate.Card{RunID: "run-b", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
			homestate.Card{RunID: "run-a", CardID: "t3", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: "lane-3", LeaseHolder: "lane-3", LeaseExpiresAt: live, WorktreePath: root},
		)
		return root
	}
	want := fcNow.Add(15 * time.Minute).Format(time.RFC3339Nano)
	mismatch := func(bound string) string {
		return "no --run given: lane lane-1 is bound to run " + bound + ", not to the single active run run-a; pass --run"
	}
	for _, arm := range []struct {
		name, bound, lane, card, refuse string // refuse is the exact refusal detail; empty when the renewal proceeds
	}{
		{"bound to a retired run refuses and names --run", "run-b", "lane-1", "t1", mismatch("run-b")},
		{"bound to a run the record does not hold refuses and names --run", "run-ghost", "lane-1", "t1", mismatch("run-ghost")},
		{"unbound lane keeps the sole-run renewal", "", "lane-1", "t1", ""},
		{"lane with no roster row keeps the sole-run renewal", "", "lane-3", "t3", ""},
		{"lane bound to the sole run keeps the sole-run renewal", "run-a", "lane-1", "t1", ""},
	} {
		for _, surface := range []string{"verb", "MCP tool"} {
			t.Run(arm.name+"/"+surface, func(t *testing.T) {
				root := fixture(t, arm.bound)
				sdLaneEnv(t, arm.lane, "")
				beforeA := factoryRenewCardIn(t, root, "run-a", arm.card).LeaseExpiresAt
				beforeB := factoryRenewCardIn(t, root, "run-b", "t1").LeaseExpiresAt
				wire, callErr := factoryRenewCardNoRun(t, root, surface, arm.card)
				afterA := factoryRenewCardIn(t, root, "run-a", arm.card).LeaseExpiresAt
				afterB := factoryRenewCardIn(t, root, "run-b", "t1").LeaseExpiresAt
				if afterB != beforeB {
					t.Errorf("run-b lease expiry moved %s → %s; no renewal may touch the retired run", beforeB, afterB)
				}
				if arm.refuse != "" {
					if surface == "verb" && callErr == nil {
						t.Errorf("verb refusal exited zero, want a non-zero exit")
					}
					if wire.Result != "error" || wire.Detail != arm.refuse {
						t.Errorf("renew with no --run: result=%s detail=%q, want error with detail %q", wire.Result, wire.Detail, arm.refuse)
					}
					if afterA != beforeA {
						t.Errorf("refused renewal moved the run-a lease %s → %s", beforeA, afterA)
					}
					return
				}
				if callErr != nil || wire.Result != "renewed" || wire.ExpiresAt != want {
					t.Errorf("renew with no --run: err=%v result=%s expires=%s, want renewed to %s", callErr, wire.Result, wire.ExpiresAt, want)
				}
				if afterA != want {
					t.Errorf("run-a lease expiry = %s, want %s", afterA, want)
				}
			})
		}
	}
}

// TestResolveFactoryRenewRunRefusalArms pins every return of resolveFactoryRenewRun with the exact message
// it produces (sync-audit F12, card t1663). An ambiguous-set refusal names --run and ends with its arm's
// detail followed by "; pass --run"; the ambiguous outcome itself carries process-specific owner details,
// so the head and the arm's suffix are compared. The sole-run refusal and the no-active refusal answer
// without that wrapper, so their whole message is compared. Two refusal arms have no fixture and are not
// pinned here: the OpenFactory refusal (ResolveActiveRun opens the same record first and returns on its
// error before any ambiguous outcome exists) and the roster read refusal (no seam makes the SELECT fail
// between a successful ambiguous resolution and the read).
func TestResolveFactoryRenewRunRefusalArms(t *testing.T) {
	ctx := context.Background()
	const ambiguousHead = "no --run given and no single active factory run (AMBIGUOUS_FACTORY: "
	// placeT1 leases card t1 of runID to holder until five minutes past the test clock.
	placeT1 := func(t *testing.T, root, runID, holder string) {
		t.Helper()
		fcPlace(t, root, homestate.Card{RunID: runID, CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
			OwnerLabel: holder, LeaseHolder: holder, LeaseExpiresAt: fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano), WorktreePath: root})
	}
	twoRuns := []string{"run-a", "run-b"}
	for _, arm := range []struct {
		name   string
		live   []string                        // runs seeded owner-live
		setup  func(t *testing.T, root string) // roster binding, run rows and card placements of the arm
		lane   string
		run    string // the run the call returns; empty when it refuses
		detail string // the arm's detail, when the ambiguous set refuses
		exact  string // the whole message, when the call refuses outside the ambiguous wrapper
	}{
		{name: "no roster binding refuses and names --run", live: twoRuns, lane: "lane-1",
			detail: "lane lane-1 has no roster binding"},
		{name: "worker row missing where a binding is required refuses and names --run", live: twoRuns, lane: "lane-9",
			detail: "lane lane-9 has no roster binding"},
		{name: "bound run that is not active refuses and names --run", live: twoRuns, lane: "lane-1",
			setup: func(t *testing.T, root string) {
				fcPlaceRun(t, root, "run-c", "retired", "2026-09-25T00:00:00Z")
				factoryRenewBindLane(t, root, "lane-1", "run-c")
			},
			detail: "lane lane-1 is bound to run run-c, which is not an active run"},
		{name: "card missing from the bound run refuses and names --run", live: twoRuns, lane: "lane-1",
			setup: func(t *testing.T, root string) {
				factoryRenewBindLane(t, root, "lane-1", "run-b")
				placeT1(t, root, "run-a", "lane-1")
			},
			detail: "card t1 in its bound run run-b: factory card not found"},
		{name: "card not held by the lane in its bound run refuses and names --run", live: twoRuns, lane: "lane-1",
			setup: func(t *testing.T, root string) {
				factoryRenewBindLane(t, root, "lane-1", "run-b")
				placeT1(t, root, "run-b", "lane-2")
			},
			detail: "card t1 is not held by lane lane-1 in its bound run run-b"},
		{name: "card held by the lane in its bound run returns that run", live: twoRuns, lane: "lane-1", run: "run-b",
			setup: func(t *testing.T, root string) {
				factoryRenewBindLane(t, root, "lane-1", "run-b")
				placeT1(t, root, "run-b", "lane-1")
			}},
		{name: "sole active run bound to another run refuses with the exact message", live: []string{"run-a"}, lane: "lane-1",
			setup: func(t *testing.T, root string) {
				fcPlaceRun(t, root, "run-b", "retired", "2026-09-25T00:00:00Z")
				factoryRenewBindLane(t, root, "lane-1", "run-b")
			},
			exact: "no --run given: lane lane-1 is bound to run run-b, not to the single active run run-a; pass --run"},
		{name: "sole active run with an empty roster binding returns the sole run", live: []string{"run-a"}, lane: "lane-1", run: "run-a"},
		{name: "sole active run with no roster row returns the sole run", live: []string{"run-a"}, lane: "lane-9", run: "run-a"},
		{name: "sole active run bound to that run returns it", live: []string{"run-a"}, lane: "lane-1", run: "run-a",
			setup: func(t *testing.T, root string) { factoryRenewBindLane(t, root, "lane-1", "run-a") }},
		{name: "no active run refuses with the exact message", lane: "lane-1",
			setup: func(t *testing.T, root string) {
				// sdMoaiFixture seeds fcRun active for the MCP stale-run
				// validation (SPEC-MCP-NEUTRALITY-001 M1); retire it so this
				// arm's no-active-run premise holds.
				db := fcOpen(t, root)
				defer func() { _ = db.Close() }()
				if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, fcRun); err != nil {
					t.Fatal(err)
				}
			},
			exact: "no --run given and no single active factory run (NO_ACTIVE_FACTORY)"},
	} {
		t.Run(arm.name, func(t *testing.T) {
			root, _ := sdMoaiFixture(t)
			for _, run := range arm.live {
				seedLiveRun(t, root, run)
			}
			sdRegisterLane(t, root, "lane-1")
			if arm.setup != nil {
				arm.setup(t, root)
			}
			got, err := resolveFactoryRenewRun(ctx, root, "", arm.lane, "t1")
			switch {
			case arm.exact != "":
				if err == nil || err.Error() != arm.exact || got != "" {
					t.Errorf("resolveFactoryRenewRun = (%q, %v), want the refusal %q", got, err, arm.exact)
				}
			case arm.detail != "":
				suffix := ": " + arm.detail + "; pass --run"
				if err == nil || !strings.HasPrefix(err.Error(), ambiguousHead) || !strings.HasSuffix(err.Error(), suffix) || got != "" {
					t.Errorf("resolveFactoryRenewRun = (%q, %v), want an ambiguous-set refusal ending %q", got, err, suffix)
				}
			default:
				if err != nil || got != arm.run {
					t.Errorf("resolveFactoryRenewRun = (%q, %v), want (%q, nil)", got, err, arm.run)
				}
			}
		})
	}
}
