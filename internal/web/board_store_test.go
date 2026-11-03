package web

// board_store_test.go — the board over a real ledger.Store, not a fake (Issue
// #1948, local-only redesign). A card is added, claimed through the Store's own
// Claim, and the board is built from that Store, so the adapter is checked against
// the types and rules the ledger actually produces: Position, the lease and its
// conflict keys, and the database clock.
//
// Defect answer: these tests catch an adapter that reads the fake's shape but not
// the real one, a claimed card shown under the wrong stage or without its
// sub-stage, a lease without its session, epoch or keys, a conflict key without the
// card that holds it, and a flow figure that disagrees with the ledger.

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestBoard_RealStoreClaimedCardShowsStageLeaseKeysAndFlow(t *testing.T) {
	ctx := context.Background()
	store, err := ledger.Open(ctx, filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatalf("open the store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	add := func(id string, prio int, class string, paths ...string) {
		t.Helper()
		err := store.AddCard(ctx, ledger.Card{
			ID: id, Title: "card " + id, Position: core.Position{Stage: core.StageReady}, Priority: prio,
			Spec: ledger.Spec{Acceptance: []string{"TestX"}, Paths: paths, Class: class},
		})
		if err != nil {
			t.Fatalf("add %s: %v", id, err)
		}
	}
	add("a1", 9, "A", "internal/web/board.go") // claimed first: highest priority, class A goes to IMPLEMENT
	add("c2", 1, "C", "internal/cli/web.go")   // stays READY

	claimed, err := store.Claim(ctx, ledger.ClaimRequest{SessionID: "sess-real", Worktree: "/wt/real", PID: 4242, PIDStart: "start-real"})
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if claimed.CardID != "a1" {
		t.Fatalf("claimed %q, want a1", claimed.CardID)
	}
	wantKeys, err := ledger.KeysFor([]string{"internal/web/board.go"})
	if err != nil || len(wantKeys) == 0 {
		t.Fatalf("keys for the path: %v %v", wantKeys, err)
	}

	a := newApp(Config{ProjectRoot: t.TempDir(), Ledger: store})
	a.recordLastProfile = func(string) error { return nil }
	vm := a.buildBoard(ctx, claimed.ExpiresAt.Add(-1)) // just before the lease runs out
	if !vm.Available {
		t.Fatalf("a readable store gives an available board: %+v", vm)
	}

	stages := map[string][]BoardCardVM{}
	for _, g := range vm.Stages {
		stages[g.Stage] = g.Cards
	}
	if c := stages[string(core.StageBuilding)]; len(c) != 1 || c[0].ID != "a1" || c[0].Sub != string(core.SubImplement) || c[0].Parked {
		t.Errorf("BUILDING = %+v, want a1 at IMPLEMENT, not parked", c)
	}
	if c := stages[string(core.StageReady)]; len(c) != 1 || c[0].ID != "c2" {
		t.Errorf("READY = %+v, want c2", c)
	}

	if len(vm.Leases) != 1 || vm.Leases[0].CardID != "a1" || vm.Leases[0].Session != "sess-real" || vm.Leases[0].Epoch != "1" || vm.Leases[0].Expired {
		t.Errorf("leases = %+v, want one live lease of a1 by sess-real at epoch 1", vm.Leases)
	}
	held := map[string]string{}
	for _, k := range vm.Keys {
		held[k.Key] = k.CardID
	}
	for _, k := range wantKeys {
		if held[k] != "a1" {
			t.Errorf("conflict key %q held by %q, want a1 (keys: %v)", k, held[k], held)
		}
	}

	if !vm.Flow.Known || vm.Flow.WIP != "1" {
		t.Errorf("flow = %+v, want measured with WIP 1 (a1 building; c2 is waiting)", vm.Flow)
	}
	if w := groupBy(vm.Flow.Waiting); w["ready"].Count != "1" {
		t.Errorf("waiting = %+v, want one ready card", vm.Flow.Waiting)
	}

	// The same store through the routes: the page and the JSON both carry the card.
	rec := serveGet(t, a.routes(), "/board")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "sess-real") || !strings.Contains(rec.Body.String(), "a1") {
		t.Errorf("GET /board = %d, sess-real/a1 on the page: %v", rec.Code, strings.Contains(rec.Body.String(), "sess-real"))
	}
	if api := serveGet(t, a.routes(), "/api/board"); api.Code != http.StatusOK || !strings.Contains(api.Body.String(), `"a1"`) {
		t.Errorf("GET /api/board = %d", api.Code)
	}
}
