package kanban

import (
	"context"
	"testing"
)

func TestValidateGTDRelationCompatibility(t *testing.T) {
	legacyNote := []byte{0, 1, 2, 255}
	legacy := GTDRelation{SubjectID: "a", ObjectID: "b", Kind: RelationContains, Note: legacyNote}
	if err := ValidateGTDRelation(legacy, []GTDRelation{legacy}, map[string]bool{"a": true, "b": true}); err != nil {
		t.Fatalf("legacy relation rejected: %v", err)
	}
	if string(legacy.Note) != string(legacyNote) || legacy.SubjectID != "a" || legacy.ObjectID != "b" {
		t.Fatalf("legacy relation mutated: %+v", legacy)
	}
	for _, kind := range []GTDRelationKind{RelationPartOf, RelationSupportedBy, RelationRelatedTo, RelationSupersedes} {
		if GTDRelationBlocks(kind, false) {
			t.Fatalf("reference relation %s blocks execution", kind)
		}
	}
	if !GTDRelationBlocks(RelationDependsOn, false) || GTDRelationBlocks(RelationDependsOn, true) {
		t.Fatal("depends_on completion semantics are wrong")
	}
	missing := GTDRelation{SubjectID: "a", ObjectID: "missing", Kind: RelationDependsOn}
	if err := ValidateGTDRelation(missing, nil, map[string]bool{"a": true}); err == nil {
		t.Fatal("missing target allowed")
	}
	cycleBase := []GTDRelation{{SubjectID: "a", ObjectID: "b", Kind: RelationDependsOn}, {SubjectID: "b", ObjectID: "c", Kind: RelationDependsOn}}
	if err := ValidateGTDRelation(GTDRelation{SubjectID: "c", ObjectID: "a", Kind: RelationDependsOn}, cycleBase, map[string]bool{"a": true, "b": true, "c": true}); err == nil {
		t.Fatal("depends_on cycle allowed")
	}
}

func TestPutGTDRelationAndValidationRefusals(t *testing.T) {
	targets := map[string]bool{"a": true, "b": true}
	for _, invalid := range []GTDRelation{
		{},
		{SubjectID: "a", ObjectID: "a", Kind: RelationRelatedTo},
		{SubjectID: "a", ObjectID: "b", Kind: "invented"},
		{SubjectID: "a", ObjectID: "b", Kind: RelationRelatedTo, SourceRevision: -1},
	} {
		if err := ValidateGTDRelation(invalid, nil, targets); err == nil {
			t.Fatalf("invalid relation allowed: %+v", invalid)
		}
	}
	store := newGTDTestStore(t)
	a, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "a", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "relation-a"})
	b, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "b", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "relation-b"})
	relation := GTDRelation{SubjectID: a.ItemID, ObjectID: b.ItemID, Kind: RelationSupportedBy, Note: []byte{0, 255}, Source: "user", AssertionStatus: "confirmed", SourceRevision: 1, PolicyVersion: "p1"}
	if err := PutGTDRelation(context.Background(), store, relation, nil, targets); err != nil {
		t.Fatal(err)
	}
	relation.Note = []byte{1, 2, 3}
	if err := PutGTDRelation(context.Background(), store, relation, nil, targets); err != nil {
		t.Fatal(err)
	}
}
