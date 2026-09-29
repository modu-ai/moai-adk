// todo_queue_sorted_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 AC-TCD-006
// (web half): the web queue read renders rows in the store's (sorted) order.
// The sort lives in the locked write (REQ-TCD-005), so the console needs no
// ordering logic of its own — this test pins that the read seam passes the
// stored order through untouched, so every consumer that reads position
// reads priority.
package web

import (
	"fmt"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestWebQueueReadRendersSortedOrder(t *testing.T) {
	root := t.TempDir()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
	ids := map[rune]string{}
	for _, name := range []string{"A", "B", "C", "D"} {
		it, _, err := store.Add("card " + name)
		if err != nil {
			t.Fatalf("add %s: %v", name, err)
		}
		ids[rune(name[0])] = it.ID
	}
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		cls := func(prio string, blocked bool) *kanban.CardClassification {
			return &kanban.CardClassification{Priority: prio, Blocked: blocked, Mode: kanban.ClassModeSerial, Decider: kanban.DeciderIdentityLLM}
		}
		for i := range rec.Items {
			switch rec.Items[i].ID {
			case ids['A']:
				rec.Items[i].Classification = cls(kanban.ClassPriorityNormal, false)
			case ids['B']:
				rec.Items[i].Classification = cls(kanban.ClassPriorityHigh, false)
			case ids['C']:
				rec.Items[i].Classification = cls(kanban.ClassPriorityLow, false)
			case ids['D']:
				rec.Items[i].Classification = cls(kanban.ClassPriorityNormal, true)
			}
		}
		// The add path re-sorts inside its locked write (REQ-TCD-005); the
		// read never re-sorts (plan G5). This fixture replays that same write
		// shape: classify AND sort in one mutation, then read.
		rec.SortByClassification()
		return nil
	}); err != nil {
		t.Fatalf("classify: %v", err)
	}

	vm := readTodoQueue(root)
	var got []string
	for _, it := range vm.Items {
		got = append(got, it.ID)
	}
	want := []string{ids['B'], ids['A'], ids['C'], ids['D']}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("web queue rows = %v, want the sorted order %v (blocked sinks, priority orders, insertion stable)", got, want)
	}
}
