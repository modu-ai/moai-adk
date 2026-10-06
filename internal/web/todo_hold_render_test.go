// todo_hold_render_test.go — SPEC-TODO-HOLD-STATE-001 M4 (AC-THS-017): the
// web console renders a held card's state as the literal "hold" — the state
// passthrough (read seam → TodoItemVM → todoStateBadge) needs no change; the
// test pins that it renders truthfully.
package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestTodoQueueRendersHeldCardState(t *testing.T) {
	a := newTestApp(t)
	root := a.cfg.ProjectRoot

	store := factory.NewBacklogStore(factory.BacklogPathForRootAdopting(root))
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.LastSeq++
		rec.Items = append(rec.Items, factory.BacklogItem{
			ID:      "t" + itoaForWebTest(rec.LastSeq),
			Text:    "parked pending decision",
			AddedAt: "2026-09-29T00:00:00Z",
			State:   factory.BacklogStateHold,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed held card: %v", err)
	}

	// The view model carries the literal state through the read seam.
	vm := readTodoQueue(root)
	if len(vm.Items) != 1 {
		t.Fatalf("vm items = %d, want 1", len(vm.Items))
	}
	if vm.Items[0].State != "hold" {
		t.Errorf("TodoItemVM.State = %q, want the literal %q", vm.Items[0].State, "hold")
	}

	// The rendered screen markup carries the value.
	rec := serveGet(t, a.routes(), "/todo")
	if rec.Code != 200 {
		t.Fatalf("GET /todo status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-todo-state="hold"`) {
		t.Errorf("rendered /todo markup lacks data-todo-state=\"hold\" — the held card's state must be visible")
	}
	if !strings.Contains(body, "parked pending decision") {
		t.Errorf("rendered /todo markup does not show the held card's text")
	}
}

func itoaForWebTest(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
