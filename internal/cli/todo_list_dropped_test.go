// todo_list_dropped_test.go — card t384: the default list render hides
// dropped cards.
//
// The defect: `list` (and the bare `moai todo` glance) rendered every
// dropped card with its full `[DROPPED — <reason>]` text, forever — a
// decided card never leaves rec.Items, so the rendered list length diverged
// from the queue's actual load (the live queue rendered 40KB+ of mostly
// dropped cards). Every other live-queue reader already excludes them: bare
// `next` lists queued cards only, `analyze` skips them, and the counts count
// picked+queued only. The list was the last reader still rendering them.
//
// The repair is render-surface only — the store keeps the card and its
// reason (t153's exact-reversal contract is untouched): the default view
// renders live cards plus one count line naming the recovery path, and
// `list --dropped` renders the discarded set. `--json` stays the structured
// truth and keeps carrying every card.
package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// assertFixtureStates reads the seeded queue back through its own store and
// verifies the exact state mix before any render assertion runs — the
// render tests below are only meaningful when the fixture is what they say.
func assertFixtureStates(t *testing.T, store *kanban.BacklogStore, want map[string]kanban.BacklogState) {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load seeded queue: %v", err)
	}
	if len(rec.Items) != len(want) {
		t.Fatalf("seeded %d items, want %d: %+v", len(rec.Items), len(want), rec.Items)
	}
	for _, it := range rec.Items {
		if want[it.ID] != it.State {
			t.Fatalf("item %s state = %s, want %s", it.ID, it.State, want[it.ID])
		}
	}
}

// seedDroppedFixture builds a queue with one picked, one queued, and one
// dropped card — the smallest mix that discriminates the three render
// behaviors under test.
func seedDroppedFixture(t *testing.T) *kanban.BacklogStore {
	t.Helper()
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "alpha live work"); err != nil {
		t.Fatalf("seed add alpha: %v", err)
	}
	if _, _, err := runTodo(t, "add", "beta live work"); err != nil {
		t.Fatalf("seed add beta: %v", err)
	}
	if _, _, err := runTodo(t, "add", "gamma discard me"); err != nil {
		t.Fatalf("seed add gamma: %v", err)
	}
	if _, _, err := runTodo(t, "next", "t1"); err != nil {
		t.Fatalf("seed pick t1: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "t3", "superseded by another card"); err != nil {
		t.Fatalf("seed drop t3: %v", err)
	}
	assertFixtureStates(t, store, map[string]kanban.BacklogState{
		"t1": kanban.BacklogStatePicked,
		"t2": kanban.BacklogStateQueued,
		"t3": kanban.BacklogStateDropped,
	})
	return store
}

// The default view — both entry points — must render the live cards and
// hide the dropped one behind a count line naming the recovery path.
func TestTodoList_DroppedHiddenByDefault(t *testing.T) {
	seedDroppedFixture(t)

	for _, entry := range []struct {
		name string
		args []string
	}{{"bare invocation", nil}, {"list", []string{"list"}}} {
		t.Run(entry.name, func(t *testing.T) {
			out, _, err := runTodo(t, entry.args...)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if !strings.Contains(out, "t1") || !strings.Contains(out, "t2") {
				t.Errorf("live cards missing from the default render:\n%s", out)
			}
			if strings.Contains(out, "t3") {
				t.Errorf("dropped card t3 rendered in the default view:\n%s", out)
			}
			if !strings.Contains(out, "1 dropped") {
				t.Errorf("dropped count line missing from the default render:\n%s", out)
			}
			if !strings.Contains(out, "--dropped") {
				t.Errorf("recovery path not named in the default render:\n%s", out)
			}
		})
	}
}

// `list --dropped` is the recovery surface undrop needs: the discarded set
// with its markers, and nothing else.
func TestTodoList_DroppedFlagRendersDroppedOnly(t *testing.T) {
	seedDroppedFixture(t)

	out, _, err := runTodo(t, "list", "--dropped")
	if err != nil {
		t.Fatalf("list --dropped: %v", err)
	}
	if !strings.Contains(out, "t3") || !strings.Contains(out, "[DROPPED — superseded by another card]") {
		t.Errorf("dropped card or marker missing from --dropped render:\n%s", out)
	}
	if strings.Contains(out, "t1") || strings.Contains(out, "t2") {
		t.Errorf("live cards rendered in the --dropped view:\n%s", out)
	}
}

// A queue with no dropped cards says so explicitly under --dropped, rather
// than printing nothing.
func TestTodoList_DroppedFlagEmptySaysSo(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "only card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	assertFixtureStates(t, store, map[string]kanban.BacklogState{
		"t1": kanban.BacklogStateQueued,
	})

	out, _, err := runTodo(t, "list", "--dropped")
	if err != nil {
		t.Fatalf("list --dropped: %v", err)
	}
	if !strings.Contains(out, "no dropped cards") {
		t.Errorf("--dropped on a queue with no dropped cards printed:\n%s", out)
	}
}

// A queue whose every card is dropped renders just the count line in the
// default view — no live rows, and an honest glance.
func TestTodoList_AllDroppedDefaultView(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "only card"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "no longer wanted"); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	assertFixtureStates(t, store, map[string]kanban.BacklogState{
		"t1": kanban.BacklogStateDropped,
	})

	out, _, err := runTodo(t)
	if err != nil {
		t.Fatalf("bare list: %v", err)
	}
	if strings.Contains(out, "t1") {
		t.Errorf("dropped-only queue rendered the card in the default view:\n%s", out)
	}
	if !strings.Contains(out, "1 dropped") {
		t.Errorf("dropped count line missing:\n%s", out)
	}
}

// --json stays the structured truth: every card, dropped included, so a
// machine consumer filters by the state field rather than by absence.
func TestTodoList_JSONKeepsDroppedCards(t *testing.T) {
	seedDroppedFixture(t)

	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rec.Items) != 3 {
		t.Fatalf("json items = %d, want 3 (the dropped card stays in the structured record)", len(rec.Items))
	}
	var dropped int
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStateDropped {
			dropped++
		}
	}
	if dropped != 1 {
		t.Errorf("json dropped cards = %d, want 1", dropped)
	}
}
