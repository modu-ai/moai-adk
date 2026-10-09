package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The selection paths that re-point the dispatch binding to the session's
// factory run (`todo next <card>`, `todo claim`, `todo --auto`) refresh the
// queue's current-dispatch record (turn-end gate relay #3, card t1538): a
// binding that moved ahead of the record would let a retried older dispatch
// operation read the stale record as the card's current engagement. These
// paths claim no owner, so the record keeps the owner it holds for the same
// run and carries none onto another (seedOlderDispatch lives with the
// factory assign tests).

func TestTodoPickRefreshesCurrentDispatch(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "re-selected card")
	fcPlaceFactoryCard(t, root, cardID, 1, "sha-new", "2026-09-26T02:00:00Z")
	seedOlderDispatch(t, root, store, cardID)

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "next", cardID); err != nil {
		t.Fatalf("next: %v", err)
	}
	fcWantCurrent(t, store, cardID, fcRun, "")
}

func TestTodoClaimRefreshesCurrentDispatch(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "claimed card")
	fcPlaceRun(t, root, fcRun, "active", "2026-09-25T00:00:00Z")
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: string(factory.BacklogStateQueued), OwnerLabel: "worker-1", Version: 1})
	seedOlderDispatch(t, root, store, cardID)

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "claim"); err != nil {
		t.Fatalf("claim: %v", err)
	}
	fcWantCurrent(t, store, cardID, fcRun, "")
}
