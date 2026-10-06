package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// card t1454 card-review r2 finding 9: a card leased through the factory
// (`factory next`) holds its lane in the factory RECORD, and the queue's
// runtime assignments carry no row for it — the presentation read only the
// assignments, so the lease's in-flight card never became an overlap probe
// target.
func TestTodoIssuancePresentationIncludesFactoryLeasedCards(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "in-flight card touching internal/cli/todo.go"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStatePicked
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// The lease lives only in the factory record — the shape a factory-next
	// lease leaves behind.
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, LeaseHolder: "lane-1"})

	p := todoIssuancePresentation(root, "also touches internal/cli/todo.go")
	if len(p.Overlap.Items) == 0 {
		t.Fatalf("overlap = %+v, want the factory-leased card among the in-flight targets", p.Overlap)
	}
	for _, it := range p.Overlap.Items {
		if it.CardID == "t1" && it.Lane == "lane-1" {
			return
		}
	}
	t.Fatalf("overlap items %+v do not name t1 on lane-1", p.Overlap.Items)
}
