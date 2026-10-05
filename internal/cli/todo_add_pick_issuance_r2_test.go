package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// card t1454 card-review r2 finding 7: `add --pick` carries the issuance
// flags too. The early pick return skipped the attribute resolution
// entirely — --origin/--parent/--size-lines/--files went unvalidated and
// unsaved on the picked card.
func TestTodoAddPickValidatesAndCarriesIssuanceFlags(t *testing.T) {
	_, store := todoFixture(t)

	// (a) an out-of-set origin is refused, nothing is written.
	if _, _, err := runTodo(t, "add", "pick with a bad origin", "--pick", "--origin", "bogus"); err == nil {
		t.Fatal("--pick accepted an out-of-set --origin")
	}
	// (b) a parent that names no card is refused.
	if _, _, err := runTodo(t, "add", "pick with a missing parent", "--pick", "--parent", "t99"); err == nil {
		t.Fatal("--pick accepted a --parent that names no card")
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("the refused picks wrote %d card(s); want none", len(rec.Items))
	}

	// (c) valid attributes are validated AND recorded on the picked card.
	if _, _, err := runTodo(t, "add", "the parent card"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "pick with attributes", "--pick",
		"--origin", "follow-up", "--parent", "t1", "--size-lines", "12", "--files", "a.go,b.go"); err != nil {
		t.Fatalf("add --pick with issuance flags: %v", err)
	}
	rec, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	var iss *factory.BacklogIssuance
	for i := range rec.Items {
		if rec.Items[i].State == factory.BacklogStatePicked {
			iss = rec.Items[i].Issuance
		}
	}
	if iss == nil {
		t.Fatal("the picked card carries no issuance record — --pick dropped the attributes")
	}
	if iss.Origin != "follow-up" || iss.SpawnedBy != "t1" {
		t.Fatalf("issuance origin/parent = %q/%q, want follow-up/t1", iss.Origin, iss.SpawnedBy)
	}
	if iss.SizeLines == nil || *iss.SizeLines != 12 {
		t.Fatalf("issuance size_lines = %v, want 12", iss.SizeLines)
	}
	if len(iss.Files) != 2 || iss.Files[0] != "a.go" || iss.Files[1] != "b.go" {
		t.Fatalf("issuance files = %v, want [a.go b.go]", iss.Files)
	}
}
