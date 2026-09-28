// backlog_hold_test.go — SPEC-TODO-HOLD-STATE-001 M4 (AC-THS-016): the
// statusline's backlog counts count a held card in NEITHER the picked nor
// the queued count. The positive aggregates already give this by
// construction; the test pins it, and its positive control proves the count
// can see hold by flipping the card to queued and watching the count move.
package statusline

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestBacklogCountsForRootExcludesHeldCards(t *testing.T) {
	root := t.TempDir()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRootAdopting(root))
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		states := []kanban.BacklogState{
			kanban.BacklogStatePicked,
			kanban.BacklogStateQueued,
			kanban.BacklogStateQueued,
			kanban.BacklogStateHold,
		}
		for _, state := range states {
			rec.LastSeq++
			rec.Items = append(rec.Items, kanban.BacklogItem{
				ID:      "t" + itoa(rec.LastSeq),
				Text:    "hold-count fixture card",
				AddedAt: "2026-09-29T00:00:00Z",
				State:   state,
			})
		}
		return nil
	}); err != nil {
		t.Fatalf("seed queue: %v", err)
	}

	counts := kanban.BacklogCountsForRoot(root)
	if !counts.Available {
		t.Fatal("queue unreadable — the fixture is broken, not the contract")
	}
	if counts.Picked != 1 || counts.Queued != 2 {
		t.Errorf("counts = picked %d / queued %d, want 1 / 2 — the held card lands in neither count",
			counts.Picked, counts.Queued)
	}

	// Positive control (mutation): the same card in queued state moves the
	// queued count to 3 — the aggregate sees a held row, it does not ignore
	// the queue.
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].State == kanban.BacklogStateHold {
				rec.Items[i].State = kanban.BacklogStateQueued
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("flip hold to queued: %v", err)
	}
	counts = kanban.BacklogCountsForRoot(root)
	if counts.Queued != 3 {
		t.Errorf("positive control: queued = %d, want 3 once the held card is queued", counts.Queued)
	}
}
