package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// card t1454 card-review r2 finding 11: `todo trace` reads the queue's whole
// relation surface through the common resolver — the stored findings AND the
// issuance spawned_by projections. The findings-only walk missed the
// spawned_by parent/follow-up edges entirely.
func TestTodoTraceCoversSpawnedByProjections(t *testing.T) {
	_, store := todoFixture(t)
	_ = store
	if _, _, err := runTodo(t, "add", "the origin card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "the follow-up card", "--origin", "follow-up", "--parent", "t1"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runTodo(t, "trace", "t2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "follow-up-of") || !strings.Contains(out, "t1") {
		t.Fatalf("trace t2 = %q, want the spawned_by follow-up edge to t1", out)
	}
}

// card t1454 card-review r2 finding 14: `todo why` renders the follow-up
// projection child → origin — "t2 follow-up-of t1" — not the semantic
// reverse.
func TestTodoWhyShowsFollowUpChildToOrigin(t *testing.T) {
	_, store := todoFixture(t)
	_ = store
	if _, _, err := runTodo(t, "add", "the origin card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "the follow-up card", "--origin", "follow-up", "--parent", "t1"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runTodo(t, "why", "t2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "follow-up-of t2 → t1") {
		t.Fatalf("why t2 = %q, want the child → origin direction", out)
	}
}

// card t1454 card-review r2c: the supersedes cycle guard keys on the
// MAPPED kind. `replaces` is the legacy spelling of supersedes and rides
// the same edges; checking the input name alone let a replaces input close
// a supersedes cycle without the guard firing.
func TestTodoRelateReplacesRunsTheSupersedesCycleGuard(t *testing.T) {
	_, store := todoFixture(t)
	_ = store
	if _, _, err := runTodo(t, "add", "the first card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "the second card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "replaces"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "relate", "t2", "t1", "--relation", "replaces"); err == nil {
		t.Fatal("the opposite replaces input closed a supersedes cycle without the guard firing")
	}
}

// card t1454 card-review r2 finding 15, the CLI half: `todo relate`'s
// duplicate check normalizes the STORED rows — a pair an older writer
// recorded in the opposite order still maps onto the first record.
func TestTodoRelateNormalizesStoredRowsForDedup(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "first card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "second card"); err != nil {
		t.Fatal(err)
	}
	// Seed the pair unnormalized, the shape a pre-normalization writer left.
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Findings = append(rec.Findings, factory.BacklogFinding{
			SubjectID: "t2", RelatedID: "t1", Relation: "relates-to", Source: factory.BacklogSourceAgent,
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "relate", "t1", "t2", "--relation", "relates-to"); err == nil {
		t.Fatal("the opposite-order re-record of the stored pair was accepted")
	}
	_ = root
}
