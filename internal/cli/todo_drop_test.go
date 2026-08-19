// todo_drop_test.go — t153: acceptance tests for `drop` and `undrop`, the
// operator's discard verb and its exact inverse.
//
// The load-bearing test here is the round-trip: drop followed by undrop must
// leave the queue file BYTE-IDENTICAL. A discard verb whose reversal is
// approximate is how a mis-judged card gets silently swallowed, which is the
// hazard this pair exists to bound.
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestTodoDrop_MarksDroppedAndRecordsTheReason(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "raise the coverage floor")

	out, _, err := runTodo(t, "drop", "1", "전제 반증")
	if err != nil {
		t.Fatalf("drop: %v", err)
	}
	if !strings.HasPrefix(out, "dropped t1 ") {
		t.Errorf("drop output = %q, want it to start with %q", out, "dropped t1 ")
	}
	if !strings.Contains(out, "전제 반증") {
		t.Errorf("drop output = %q, want it to carry the reason", out)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("item count = %d, want 1 — drop keeps the card in the file", len(rec.Items))
	}
	got := rec.Items[0]
	if got.State != kanban.BacklogStateDropped {
		t.Errorf("state = %q, want %q", got.State, kanban.BacklogStateDropped)
	}
	if want := "[DROPPED — 전제 반증] raise the coverage floor"; got.Text != want {
		t.Errorf("text = %q, want %q", got.Text, want)
	}
}

func TestTodoDrop_UndropIsAnExactReversal(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha", "beta", "gamma")
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "drop", "2", "t20으로 흡수"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if same := readBacklogBytes(t, root); string(same) == string(before) {
		t.Fatal("drop must actually change the file")
	}

	out, _, err := runTodo(t, "undrop", "2")
	if err != nil {
		t.Fatalf("undrop: %v", err)
	}
	if !strings.HasPrefix(out, "undropped t2 ") {
		t.Errorf("undrop output = %q, want it to start with %q", out, "undropped t2 ")
	}

	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Errorf("drop+undrop is not an exact reversal.\n got: %s\nwant: %s", got, before)
	}
}

func TestTodoDrop_UndropRestoresAHandWrittenDroppedCard(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "hand-edited card")
	// The four dropped cards on the live queue were written into the file by
	// hand, before any CLI verb existed. undrop reads the STATE as the
	// authority and strips the marker only when it is there, so those cards
	// are recoverable too.
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items[0].State = kanban.BacklogStateDropped
		return nil
	}); err != nil {
		t.Fatalf("seed hand-edit: %v", err)
	}

	if _, _, err := runTodo(t, "undrop", "1"); err != nil {
		t.Fatalf("undrop of an unmarked dropped card: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStateQueued {
		t.Errorf("state = %q, want %q", rec.Items[0].State, kanban.BacklogStateQueued)
	}
	if rec.Items[0].Text != "hand-edited card" {
		t.Errorf("text = %q, want it untouched", rec.Items[0].Text)
	}
}

func TestTodoDrop_RefusalsLeaveTheFileByteIdentical(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha card", "beta card")
	if _, _, err := runTodo(t, "next", "2"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}
	before := readBacklogBytes(t, root)

	cases := []struct {
		name string
		args []string
	}{
		{"empty reason", []string{"drop", "1", "   "}},
		{"reason carrying the closing bracket", []string{"drop", "1", "see [x] instead"}},
		{"absent card", []string{"drop", "99", "gone"}},
		{"picked card", []string{"drop", "2", "not while in flight"}},
		{"expect mismatch", []string{"drop", "1", "reason", "--expect", "beta"}},
		{"undrop of a queued card", []string{"undrop", "1"}},
		{"undrop of an absent card", []string{"undrop", "99"}},
		{"undrop expect mismatch", []string{"undrop", "1", "--expect", "beta"}},
	}
	for _, tc := range cases {
		if _, _, err := runTodo(t, tc.args...); err == nil {
			t.Errorf("%s: %v must fail", tc.name, tc.args)
		}
		if got := readBacklogBytes(t, root); string(got) != string(before) {
			t.Fatalf("%s: refused command must leave the queue file byte-identical", tc.name)
		}
	}
}

func TestTodoDrop_AlreadyDropped_Refused(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha")
	if _, _, err := runTodo(t, "drop", "1", "first reason"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	after := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "drop", "1", "second reason"); err == nil {
		t.Fatal("dropping an already-dropped card must fail — a second marker would break the reversal")
	}
	if got := readBacklogBytes(t, root); string(got) != string(after) {
		t.Error("refused re-drop must leave the queue file byte-identical")
	}
}

func TestTodoDrop_DroppedCardIsNotAPickCandidate(t *testing.T) {
	_, _ = todoFixture(t)
	seedTodo(t, "alpha", "beta")
	if _, _, err := runTodo(t, "drop", "1", "전제 반증"); err != nil {
		t.Fatalf("drop: %v", err)
	}

	out, _, err := runTodo(t, "next")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if strings.Contains(out, "t1") {
		t.Errorf("bare next = %q, want the dropped card excluded from the candidates", out)
	}
	if !strings.Contains(out, "t2") {
		t.Errorf("bare next = %q, want the queued card still listed", out)
	}
}

func TestTodoDrop_ExpectMatchAllowsTheDrop(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "alpha card")

	if _, _, err := runTodo(t, "drop", "1", "reason", "--expect", "alpha"); err != nil {
		t.Fatalf("drop whose --expect matches must succeed: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStateDropped {
		t.Errorf("state = %q, want %q", rec.Items[0].State, kanban.BacklogStateDropped)
	}
}
