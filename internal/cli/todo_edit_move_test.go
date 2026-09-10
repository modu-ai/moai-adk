// todo_edit_move_test.go — t119: acceptance tests for the two correction
// verbs the queue lacked, `edit` (card text) and `move` (queue order).
//
// Both verbs are OPERATOR acts: nothing here infers what to correct or where
// a card belongs. The tests bind the properties that make a mis-correction
// recoverable — identity fields survive an edit, the prior text is printed
// so the edit can be reversed by hand, no item is lost or duplicated by a
// move, and every refusal leaves the file byte-identical.
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedTodo appends cards through the CLI and fails the test on any error.
func seedTodo(t *testing.T, texts ...string) {
	t.Helper()
	for _, text := range texts {
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("seed add %q: %v", text, err)
		}
	}
}

// todoOrder returns the item ids in queue-file order.
func todoOrder(t *testing.T, store *kanban.BacklogStore) []string {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ids := make([]string, 0, len(rec.Items))
	for _, it := range rec.Items {
		ids = append(ids, it.ID)
	}
	return ids
}

// readBacklogBytes returns the queue's stored record in canonical document
// form, for invariance assertions on refused mutations.
func readBacklogBytes(t *testing.T, root string) []byte {
	t.Helper()
	return queueStateBytes(t, todoBacklogPath(root))
}

func TestTodoEdit_RewritesTextPreservingIdentity(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "fix teh flaky gate")
	if _, _, err := runTodo(t, "next", "1", "--spec", "SPEC-X-001"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}
	before, err := store.Load()
	if err != nil {
		t.Fatalf("load before: %v", err)
	}

	out, _, err := runTodo(t, "edit", "1", "fix the flaky gate")
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !strings.HasPrefix(out, "edited t1 ") {
		t.Errorf("edit output = %q, want it to start with %q", out, "edited t1 ")
	}

	after, err := store.Load()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	if len(after.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(after.Items))
	}
	got, want := after.Items[0], before.Items[0]
	if got.Text != "fix the flaky gate" {
		t.Errorf("text = %q, want the corrected text", got.Text)
	}
	if got.ID != want.ID || got.AddedAt != want.AddedAt || got.State != want.State {
		t.Errorf("identity drifted: %+v, want id/added_at/state of %+v", got, want)
	}
	if got.SpecID == nil || *got.SpecID != "SPEC-X-001" {
		t.Errorf("spec_id = %v, want it preserved", got.SpecID)
	}
}

func TestTodoEdit_PrintsPriorTextSoTheEditIsReversible(t *testing.T) {
	_, _ = todoFixture(t)
	seedTodo(t, "original wording")

	out, _, err := runTodo(t, "edit", "t1", "replacement wording")
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !strings.Contains(out, "original wording") {
		t.Errorf("edit output = %q, want it to carry the prior text", out)
	}
	if !strings.Contains(out, "replacement wording") {
		t.Errorf("edit output = %q, want it to carry the new text", out)
	}
}

func TestTodoEdit_EmptyText_RefusedNoWrite(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "keep me")
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "edit", "1", "   "); err == nil {
		t.Fatal("empty-text edit must fail")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused edit must leave the queue file byte-identical")
	}
}

func TestTodoEdit_UnknownID_RefusedNoWrite(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "keep me")
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "edit", "99", "new text"); err == nil {
		t.Fatal("edit of an absent id must fail")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused edit must leave the queue file byte-identical")
	}
}

func TestTodoEdit_ExpectMismatch_RefusedNoWrite(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha card")
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "edit", "1", "rewritten", "--expect", "beta"); err == nil {
		t.Fatal("edit whose --expect does not match must fail")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused edit must leave the queue file byte-identical")
	}

	if _, _, err := runTodo(t, "edit", "1", "rewritten", "--expect", "alpha"); err != nil {
		t.Fatalf("edit whose --expect matches must succeed: %v", err)
	}
}

func TestTodoMove_TopAndBottom(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "a", "b", "c")

	out, _, err := runTodo(t, "move", "3", "--top")
	if err != nil {
		t.Fatalf("move --top: %v", err)
	}
	if !strings.HasPrefix(out, "moved t3 1 ") {
		t.Errorf("move output = %q, want it to start with %q", out, "moved t3 1 ")
	}
	if got := todoOrder(t, store); strings.Join(got, ",") != "t3,t1,t2" {
		t.Errorf("order after --top = %v, want [t3 t1 t2]", got)
	}

	if _, _, err := runTodo(t, "move", "t3", "--bottom"); err != nil {
		t.Fatalf("move --bottom: %v", err)
	}
	if got := todoOrder(t, store); strings.Join(got, ",") != "t1,t2,t3" {
		t.Errorf("order after --bottom = %v, want [t1 t2 t3]", got)
	}
}

func TestTodoMove_BeforeAndAfter(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "a", "b", "c", "d")

	if _, _, err := runTodo(t, "move", "4", "--before", "2"); err != nil {
		t.Fatalf("move --before: %v", err)
	}
	if got := todoOrder(t, store); strings.Join(got, ",") != "t1,t4,t2,t3" {
		t.Errorf("order after --before t2 = %v, want [t1 t4 t2 t3]", got)
	}

	if _, _, err := runTodo(t, "move", "t1", "--after", "t3"); err != nil {
		t.Fatalf("move --after: %v", err)
	}
	if got := todoOrder(t, store); strings.Join(got, ",") != "t4,t2,t3,t1" {
		t.Errorf("order after --after t3 = %v, want [t4 t2 t3 t1]", got)
	}
}

func TestTodoMove_PreservesEveryItem(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "a", "b", "c")
	before, err := store.Load()
	if err != nil {
		t.Fatalf("load before: %v", err)
	}

	if _, _, err := runTodo(t, "move", "2", "--top"); err != nil {
		t.Fatalf("move: %v", err)
	}

	after, err := store.Load()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	if len(after.Items) != len(before.Items) {
		t.Fatalf("item count = %d, want %d — a move must neither drop nor duplicate a card",
			len(after.Items), len(before.Items))
	}
	byID := map[string]kanban.BacklogItem{}
	for _, it := range after.Items {
		byID[it.ID] = it
	}
	for _, want := range before.Items {
		got, ok := byID[want.ID]
		if !ok {
			t.Fatalf("item %s lost by the move", want.ID)
		}
		if got != want {
			t.Errorf("item %s changed by the move: %+v, want %+v", want.ID, got, want)
		}
	}
}

func TestTodoMove_PositionFlagContract(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "a", "b")
	before := readBacklogBytes(t, root)

	cases := [][]string{
		{"move", "1"},                      // no position given
		{"move", "1", "--top", "--bottom"}, // two positions given
		{"move", "1", "--before", "1"},     // relative to itself
		{"move", "1", "--after", "99"},     // absent anchor
		{"move", "99", "--top"},            // absent card
	}
	for _, args := range cases {
		if _, _, err := runTodo(t, args...); err == nil {
			t.Errorf("%v must fail", args)
		}
		if got := readBacklogBytes(t, root); string(got) != string(before) {
			t.Fatalf("%v: refused move must leave the queue file byte-identical", args)
		}
	}
}
