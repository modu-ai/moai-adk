// todo_hold_test.go — SPEC-TODO-HOLD-STATE-001 M2 acceptance tests:
// `todo hold` / `todo unhold`, the operator verbs that park a card out of the
// queue without touching its text, and the pick refusal on a held card.
//
// The verbs are drop-isomorphic by contract: refusals return from inside
// Mutate's callback so the queue file stays byte-identical, `--expect`
// guards against an id typed from a stale listing, and the STATE is the
// authority — no `[HOLD]` marker ever enters the text.
package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-THS-006 — hold on a queued card: state becomes hold, every other field
// is preserved, one confirmation line carries the id and text prefix.
func TestTodoHold_MovesQueuedCardToHold(t *testing.T) {
	_, store := todoFixture(t)
	seedTodo(t, "raise the coverage floor")

	out, _, err := runTodo(t, "hold", "1")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if !strings.HasPrefix(out, "held t1 ") {
		t.Errorf("hold output = %q, want it to start with %q", out, "held t1 ")
	}
	if !strings.Contains(out, "raise the coverage floor") {
		t.Errorf("hold output = %q, want it to carry the text prefix", out)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	got := rec.Items[0]
	if got.State != kanban.BacklogStateHold {
		t.Errorf("state = %q, want %q", got.State, kanban.BacklogStateHold)
	}
	if got.Text != "raise the coverage floor" || got.AddedAt == "" {
		t.Errorf("held card text/added_at = %q/%q, want them untouched/present", got.Text, got.AddedAt)
	}
}

// AC-THS-006 (absorbed AC-THS-007) — hold on a non-queued card is refused
// with the current state and the recovery verb named, writing nothing.
func TestTodoHold_RefusesNonQueuedStates(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "alpha card", "beta card", "gamma card", "delta card")
	if _, _, err := runTodo(t, "next", "2"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "3", "discard"); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	if _, _, err := runTodo(t, "hold", "4"); err != nil {
		t.Fatalf("seed hold: %v", err)
	}
	before := readBacklogBytes(t, root)

	for _, tc := range []struct {
		name, id, wantInErr string
	}{
		{"picked card", "2", "unpick"},
		{"dropped card", "3", "undrop"},
		{"already held card", "4", "unhold"},
	} {
		_, _, err := runTodo(t, "hold", tc.id)
		if err == nil {
			t.Errorf("%s: hold must be refused", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.wantInErr) {
			t.Errorf("%s: refusal = %q, want it to name the recovery verb %q", tc.name, err, tc.wantInErr)
		}
		if got := readBacklogBytes(t, root); string(got) != string(before) {
			t.Fatalf("%s: refused hold must leave the queue file byte-identical", tc.name)
		}
	}
	_ = store
}

// AC-THS-006 — `--expect` mismatch refuses the hold, writing nothing (the
// drop/undrop Mutate contract).
func TestTodoHold_ExpectMismatchRefuses(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "alpha card")
	before := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "hold", "1", "--expect", "beta"); err == nil {
		t.Fatal("hold with a mismatched --expect must fail")
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused --expect hold must leave the queue file byte-identical")
	}

	if _, _, err := runTodo(t, "hold", "1", "--expect", "alpha"); err != nil {
		t.Fatalf("hold whose --expect matches must succeed: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStateHold {
		t.Errorf("state = %q, want %q", rec.Items[0].State, kanban.BacklogStateHold)
	}
}

// AC-THS-008 — unhold on a held card: back to queued, text byte-identical
// (no marker exists to strip), one confirmation line, and the pair converges
// over repeated cycles.
func TestTodoUnhold_ReturnsHeldCardToQueued(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "parked pending decision")
	original := readBacklogBytes(t, root)

	if _, _, err := runTodo(t, "hold", "1"); err != nil {
		t.Fatalf("hold: %v", err)
	}
	for cycle := 1; cycle <= 2; cycle++ {
		out, _, err := runTodo(t, "unhold", "1")
		if err != nil {
			t.Fatalf("unhold cycle %d: %v", cycle, err)
		}
		if !strings.HasPrefix(out, "unheld t1 ") {
			t.Errorf("unhold output = %q, want it to start with %q", out, "unheld t1 ")
		}
		rec, err := store.Load()
		if err != nil {
			t.Fatalf("load cycle %d: %v", cycle, err)
		}
		if rec.Items[0].State != kanban.BacklogStateQueued {
			t.Errorf("cycle %d: state = %q, want %q", cycle, rec.Items[0].State, kanban.BacklogStateQueued)
		}
		if rec.Items[0].Text != "parked pending decision" {
			t.Errorf("cycle %d: text = %q, want it byte-identical (no marker, no strip)", cycle, rec.Items[0].Text)
		}
		if _, _, err := runTodo(t, "hold", "1"); err != nil {
			t.Fatalf("re-hold cycle %d: %v", cycle, err)
		}
	}
	// Final unhold returns the card to queued; the queue bytes equal the
	// original layout (the hold state is not visible in the JSON body).
	if _, _, err := runTodo(t, "unhold", "1"); err != nil {
		t.Fatalf("final unhold: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("final load: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStateQueued || rec.Items[0].Text != "parked pending decision" {
		t.Errorf("final card = %+v, want the original queued card", rec.Items[0])
	}
	_ = original
}

// AC-THS-008 (absorbed AC-THS-009) — unhold on a non-held card is refused,
// writing nothing.
func TestTodoUnhold_RefusesNonHeldStates(t *testing.T) {
	root, _ := todoFixture(t)
	seedTodo(t, "alpha card", "beta card", "gamma card")
	if _, _, err := runTodo(t, "next", "2"); err != nil {
		t.Fatalf("seed pick: %v", err)
	}
	if _, _, err := runTodo(t, "drop", "3", "discard"); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	before := readBacklogBytes(t, root)

	for _, tc := range []struct {
		name, id string
	}{
		{"queued card", "1"},
		{"picked card", "2"},
		{"dropped card", "3"},
		{"absent card", "99"},
	} {
		if _, _, err := runTodo(t, "unhold", tc.id); err == nil {
			t.Errorf("%s: unhold must be refused", tc.name)
		}
		if got := readBacklogBytes(t, root); string(got) != string(before) {
			t.Fatalf("%s: refused unhold must leave the queue file byte-identical", tc.name)
		}
	}
}

// AC-THS-013 — the pick refusal on a held card. This is the SPEC's one
// behavioral red-now: the pick gate refused only `dropped`, so a held card
// (and any future state) was pickable. Positive control: after unhold, the
// same pick succeeds.
func TestTodoNext_PickOnHeldCardRefused(t *testing.T) {
	root, store := todoFixture(t)
	seedTodo(t, "held card")
	// Seeded directly through the store so this test isolates the PICK GATE's
	// behavior (the SPEC's one live red-now) from the hold verb's existence.
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items[0].State = kanban.BacklogStateHold
		return nil
	}); err != nil {
		t.Fatalf("seed hold state: %v", err)
	}
	before := readBacklogBytes(t, root)

	_, stderr, err := runTodo(t, "next", "1")
	if err == nil {
		t.Fatal("picking a held card must be refused — the held card is not a pick candidate")
	}
	if !strings.Contains(err.Error(), "unhold") && !strings.Contains(stderr, "unhold") {
		t.Errorf("pick refusal = %q / %q, want it to name unhold as the recovery verb", err, stderr)
	}
	if got := readBacklogBytes(t, root); string(got) != string(before) {
		t.Error("refused pick must leave the queue file byte-identical")
	}

	// The positive control: the same card back in queued state picks fine.
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items[0].State = kanban.BacklogStateQueued
		return nil
	}); err != nil {
		t.Fatalf("reset to queued: %v", err)
	}
	if _, _, err := runTodo(t, "next", "1"); err != nil {
		t.Fatalf("pick of a queued card must succeed: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStatePicked {
		t.Errorf("state after pick = %q, want %q", rec.Items[0].State, kanban.BacklogStatePicked)
	}
}

// AC-THS-010 — the actor boundary: hold/unhold exist ONLY as the operator's
// top-level todo verbs. The lease and lane paths (factory assign, the goal
// auto-mission pick, the auto-done scan, the gtd engage pick) carry no
// statement that sets or clears BacklogStateHold, and no hold/unhold flag
// reaches any lease-path command. Positive control: the hold verb's own
// implementation file carries the assignment.
func TestTodoHold_ActorBoundaryLeasePathsCannotHold(t *testing.T) {
	leasePathFiles := []string{
		"factory_card.go",
		"factory_mirror.go",
		"goal.go",
		"gtd.go",
		"todo_autodone.go",
	}
	assignment := "= kanban.BacklogStateHold"
	holdImplementations := 0
	for _, name := range leasePathFiles {
		body := readCliSource(t, name)
		for _, line := range strings.Split(string(body), "\n") {
			if strings.Contains(line, assignment) {
				t.Errorf("%s carries a hold assignment on a lease/lane path: %s", name, strings.TrimSpace(line))
			}
		}
	}
	// Positive control: the verbs' implementation names the state.
	for _, name := range []string{"todo_hold.go"} {
		body := readCliSource(t, name)
		found := false
		for _, line := range strings.Split(string(body), "\n") {
			if strings.Contains(line, assignment) || strings.Contains(line, "BacklogStateHold") {
				found = true
			}
		}
		if !found {
			t.Errorf("%s carries no BacklogStateHold reference — the positive control requires the verb's own implementation to name it", name)
		}
		holdImplementations++
	}
	if holdImplementations != 1 {
		t.Errorf("hold verb implementation files = %d, want exactly 1", holdImplementations)
	}
}

// readCliSource reads a source file of this package (the seam for the
// source-scan ACs, which are file-content judgments).
func readCliSource(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read source %s: %v", name, err)
	}
	return raw
}
