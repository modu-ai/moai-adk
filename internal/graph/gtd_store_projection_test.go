package graph

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestBuildPrivateGTDProjectionFromStoreRevisionAndRelations(t *testing.T) {
	ctx := context.Background()
	store := kanban.NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	a, _ := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "a", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "a"})
	b, _ := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "b", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivitySecret, EventID: "b"})
	rel := kanban.GTDRelation{SubjectID: a.ItemID, ObjectID: b.ItemID, Kind: kanban.RelationDependsOn, Source: "user", AssertionStatus: "confirmed", SourceRevision: 1, PolicyVersion: "p1"}
	if err := kanban.PutGTDRelation(ctx, store, rel, nil, map[string]bool{a.ItemID: true, b.ItemID: true}); err != nil {
		t.Fatal(err)
	}
	projection, err := BuildPrivateGTDProjectionFromStore(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if projection.SourceRevision != 3 || projection.EdgeCount != 1 || !PrivateGTDProjectionFresh(projection.MetaPath, 3) {
		t.Fatalf("projection=%+v", projection)
	}
	if err := kanban.PutGTDRelation(ctx, store, kanban.GTDRelation{SubjectID: a.ItemID, ObjectID: b.ItemID, Kind: kanban.RelationRelatedTo, Source: "user", AssertionStatus: "confirmed", SourceRevision: 2, PolicyVersion: "p1"}, []kanban.GTDRelation{rel}, map[string]bool{a.ItemID: true, b.ItemID: true}); err != nil {
		t.Fatal(err)
	}
	if PrivateGTDProjectionFresh(projection.MetaPath, 4) {
		t.Fatal("old projection accepted after relation revision")
	}
}

func TestBuildPrivateGTDProjectionFromStoreHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := kanban.NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if _, err := BuildPrivateGTDProjectionFromStore(ctx, store); err == nil {
		t.Fatal("cancelled projection source accepted")
	}
}
