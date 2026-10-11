package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The selection paths that re-point the binding (`todo next <card>`,
// `todo claim`, `todo --auto`) write the queue's current-dispatch record BEFORE
// the binding (turn-end gate, card t1538): their factory write has no guard
// that can refuse after it, so a record that cannot be written has to stop the
// selection before the binding moves — otherwise the binding sits ahead of a
// stale record, where a retried older dispatch reads the record as current and
// drags the binding back.

// fcBindingRunRaw reads the card's dispatch binding row itself (the run it
// names need not hold a card row).
func fcBindingRunRaw(t *testing.T, root, cardID string) string {
	t.Helper()
	var run string
	if err := fcOpen(t, root).DB.QueryRow(`SELECT run_id FROM card_dispatch WHERE card_id=?`, cardID).Scan(&run); err != nil {
		t.Fatalf("binding row of %s: %v", cardID, err)
	}
	return run
}

func TestReviewTodoPickRefreshFailureLeavesBindingUnchanged(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "re-selected card")
	fcPlaceFactoryCard(t, root, cardID, 1, "sha-new", "2026-09-26T02:00:00Z")
	seedOlderDispatch(t, root, store, cardID)
	if run := fcBindingRunRaw(t, root, cardID); run != "run-old" {
		t.Fatalf("fixture binding = %q, want run-old", run)
	}
	abortCurrentDispatchRefresh(t, store)

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "next", cardID); err == nil {
		t.Fatal("next with an unwritable current-dispatch record succeeded")
	}
	if run := fcBindingRunRaw(t, root, cardID); run != "run-old" {
		t.Fatalf("binding after the failed selection = %q, want it unmoved at run-old", run)
	}
}

func TestReviewTodoClaimRefreshFailureLeavesBindingUnchanged(t *testing.T) {
	root, store := fcFixture(t)
	cardID := addShapeCard(t, "claimed card")
	fcPlaceRun(t, root, fcRun, "active", "2026-09-25T00:00:00Z")
	fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: string(factory.BacklogStateQueued), OwnerLabel: "worker-1", Version: 1})
	seedOlderDispatch(t, root, store, cardID)
	if run := fcBindingRunRaw(t, root, cardID); run != "run-old" {
		t.Fatalf("fixture binding = %q, want run-old", run)
	}
	abortCurrentDispatchRefresh(t, store)

	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	if _, _, err := runTodo(t, "claim"); err == nil {
		t.Fatal("claim with an unwritable current-dispatch record succeeded")
	}
	if run := fcBindingRunRaw(t, root, cardID); run != "run-old" {
		t.Fatalf("binding after the failed claim = %q, want it unmoved at run-old", run)
	}
}

func TestReviewAutoReselectRefreshFailureLeavesBindingUnchanged(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t970", "auto reselect card", factory.BacklogStateQueued)
	fcPlace(t, root, homestate.Card{CardID: "t970", RunID: "run-old", State: homestate.CardDone, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T01:00:00Z"})
	fcPlaceFactoryCard(t, root, "t970", 1, "sha-new", "2026-09-26T02:00:00Z")
	seedOlderDispatch(t, root, store, "t970")
	if run := fcBindingRunRaw(t, root, "t970"); run != "run-old" {
		t.Fatalf("fixture binding = %q, want run-old", run)
	}
	abortCurrentDispatchRefresh(t, store)

	tick := 0
	opts := autoOptions{
		wait:      5 * time.Minute,
		liveness:  autoTestLiveness(root, "t970", true, true, nil),
		sessionID: "operator-session-fixture",
		sleep: func(time.Duration) {
			tick++
			path := filepath.Join(root, ".moai", "reports", "t970", "evidence.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# evidence\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		now: func() time.Time { return time.Unix(0, 0).Add(time.Duration(tick) * time.Minute) },
	}
	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	var out strings.Builder
	_ = runAutoCycle(&out, store, root, opts)
	if run := fcBindingRunRaw(t, root, "t970"); run != "run-old" {
		t.Fatalf("binding after the failed re-selection = %q, want it unmoved at run-old\n%s", run, out.String())
	}
}
