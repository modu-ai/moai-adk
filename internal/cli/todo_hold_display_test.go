// todo_hold_display_test.go — SPEC-TODO-HOLD-STATE-001 M4 acceptance tests:
// the display surfaces are TRUTHFUL about a held card — list renders its
// literal state, --json carries "state":"hold" for machine filtering, and
// nothing disguises a held card as queued (REQ-THS-016).
package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// AC-THS-015 — the text view renders the held card with its literal state
// (disclosed, never hidden, never disguised as queued), and the --json
// records carry "state":"hold".
func TestTodoListRendersHeldCardTruthfully(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "parked pending decision", "ordinary queued work")
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStateHold
		return nil
	}); err != nil {
		t.Fatalf("seed hold: %v", err)
	}

	textOut, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(textOut, "t1\thold\t") {
		t.Errorf("text list = %q, want the held card rendered with its literal hold state", textOut)
	}
	if strings.Contains(textOut, "t1\tqueued") {
		t.Errorf("text list = %q, want the held card NOT disguised as queued", textOut)
	}

	jsonOut, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	var rec factory.BacklogRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonOut)), &rec); err != nil {
		t.Fatalf("decode --json: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("json items = %d, want 2", len(rec.Items))
	}
	if rec.Items[0].State != factory.BacklogStateHold {
		t.Errorf("json state = %q, want %q — the state field is the machine-filtering contract",
			rec.Items[0].State, factory.BacklogStateHold)
	}
	if rec.Items[1].State != factory.BacklogStateQueued {
		t.Errorf("json state = %q, want %q", rec.Items[1].State, factory.BacklogStateQueued)
	}
	// The raw JSON text carries the literal value a machine filters on.
	if !strings.Contains(jsonOut, `"state":"hold"`) {
		t.Errorf("raw --json output lacks \"state\":\"hold\": %s", jsonOut)
	}
}
