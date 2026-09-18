package kanban

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGTDPersistenceOptInBackupRestoreExportImportAndLifecycle(t *testing.T) {
	ctx := context.Background()
	source := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	a, _ := CaptureGTDItem(ctx, source, CaptureInput{Content: "a", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "persist-a"})
	b, _ := CaptureGTDItem(ctx, source, CaptureInput{Content: "b", Source: "user", SourceAllowed: true, Sensitivity: SensitivitySecret, EventID: "persist-b"})
	rel := GTDRelation{SubjectID: a.ItemID, ObjectID: b.ItemID, Kind: RelationRelatedTo, Source: "user", AssertionStatus: "confirmed", SourceRevision: 1, PolicyVersion: "p1"}
	if err := PutGTDRelation(ctx, source, rel, nil, nil); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "backup.db")
	if err := BackupGTDStore(ctx, source, backup, false); err == nil {
		t.Fatal("default backup exposed private GTD")
	}
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatalf("denied backup exists: %v", err)
	}
	if err := BackupGTDStore(ctx, source, backup, true); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(backup); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("backup mode=%v err=%v", info, err)
	}
	restored := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := RestoreGTDStore(ctx, backup, restored, true); err != nil {
		t.Fatal(err)
	}
	rev, relations, err := GTDProjectionSource(ctx, restored)
	if err != nil || rev != 3 || len(relations) != 1 {
		t.Fatalf("restored rev=%d relations=%v err=%v", rev, relations, err)
	}
	if _, err := ExportGTD(ctx, source, false); err == nil {
		t.Fatal("default export exposed private GTD")
	}
	raw, err := ExportGTD(ctx, source, true)
	if err != nil {
		t.Fatal(err)
	}
	imported := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := ImportGTD(ctx, imported, raw, true); err != nil {
		t.Fatal(err)
	}
	if rev, rels, err := GTDProjectionSource(ctx, imported); err != nil || rev != 3 || len(rels) != 1 {
		t.Fatalf("imported rev=%d rels=%v err=%v", rev, rels, err)
	}
	if err := SetGTDItemCancelled(ctx, imported, b.ItemID, true); err != nil {
		t.Fatal(err)
	}
	if item, _ := LoadGTDItem(ctx, imported, b.ItemID); !item.Cancelled {
		t.Fatal("cancel not persisted")
	}
	if err := SetGTDItemCancelled(ctx, imported, b.ItemID, false); err != nil {
		t.Fatal(err)
	}
	if item, _ := LoadGTDItem(ctx, imported, b.ItemID); item.Cancelled {
		t.Fatal("reopen not persisted")
	}
	if err := DeleteGTDItem(ctx, imported, b.ItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGTDItem(ctx, imported, b.ItemID); err == nil {
		t.Fatal("deleted item loaded")
	}
	if _, rels, err := GTDProjectionSource(ctx, imported); err != nil || len(rels) != 0 {
		t.Fatalf("delete left relations=%v err=%v", rels, err)
	}
}

func TestGTDPersistenceRefusalBranches(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := BackupGTDStore(ctx, store, "relative.db", true); err == nil {
		t.Fatal("relative backup allowed")
	}
	if err := RestoreGTDStore(ctx, "relative.db", store, true); err == nil {
		t.Fatal("relative restore allowed")
	}
	if err := RestoreGTDStore(ctx, "/missing/gtd.db", store, true); err == nil {
		t.Fatal("missing restore allowed")
	}
	if err := ImportGTD(ctx, store, []byte(`{}`), true); err == nil {
		t.Fatal("invalid import allowed")
	}
	if err := ImportGTD(ctx, store, nil, false); err == nil {
		t.Fatal("default import allowed")
	}
	if err := SetGTDItemCancelled(ctx, store, "missing", true); err == nil {
		t.Fatal("missing cancel allowed")
	}
	if err := DeleteGTDItem(ctx, store, "missing"); err == nil {
		t.Fatal("missing delete allowed")
	}
	if _, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "x", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "refusal"}); err != nil {
		t.Fatal(err)
	}
	raw, err := ExportGTD(ctx, store, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := ImportGTD(ctx, store, raw, true); err == nil {
		t.Fatal("nonempty import allowed")
	}
	backup := filepath.Join(t.TempDir(), "existing.db")
	if err := os.WriteFile(backup, []byte("occupied"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := BackupGTDStore(ctx, store, backup, true); err == nil {
		t.Fatal("existing backup destination overwritten")
	}
	if err := RestoreGTDStore(ctx, backup, NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json")), false); err == nil {
		t.Fatal("restore without opt-in allowed")
	}
	corruptStore := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := RestoreGTDStore(ctx, backup, corruptStore, true); err == nil {
		t.Fatal("corrupt SQLite backup restored")
	}
	validBackup := filepath.Join(t.TempDir(), "valid.db")
	if err := BackupGTDStore(ctx, store, validBackup, true); err != nil {
		t.Fatal(err)
	}
	existingStore := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if _, err := CaptureGTDItem(ctx, existingStore, CaptureInput{Content: "existing", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "existing"}); err != nil {
		t.Fatal(err)
	}
	if err := RestoreGTDStore(ctx, validBackup, existingStore, true); err == nil {
		t.Fatal("restore overwrote existing database")
	}
}

func TestGTDExportImportPreservesGovernanceState(t *testing.T) {
	ctx := context.Background()
	source := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := MigrateGTDSchema(source); err != nil {
		t.Fatal(err)
	}
	db, err := openGTDDB(source)
	if err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`INSERT INTO gtd_contracts VALUES('m1','p1','h1',X'7B7D',1)`,
		`INSERT INTO gtd_missions VALUES('m1','auto','running','h1','s1','owner',2)`,
		`INSERT INTO gtd_events VALUES('e1','m1','approved','ph','2026-09-15T00:00:00Z')`,
		`INSERT INTO gtd_operations VALUES('op1','m1','publish','gtd:x','reconciled',X'7B7D','s1','2026-09-15T00:00:00Z')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()
	raw, err := ExportGTD(ctx, source, true)
	if err != nil {
		t.Fatal(err)
	}
	destination := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	if err := ImportGTD(ctx, destination, raw, true); err != nil {
		t.Fatal(err)
	}
	db, err = openGTDDB(destination)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	for _, table := range []string{"gtd_contracts", "gtd_missions", "gtd_events", "gtd_operations"} {
		var count int
		if err := db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("%s count=%d err=%v", table, count, err)
		}
	}
}

func TestReflectGTDStoreUsesPersistedRelationsAndLifecycle(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	project, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "project", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "reflect-project"})
	if err != nil {
		t.Fatal(err)
	}
	action, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "action", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "reflect-action"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: project.ItemID, DesiredOutcome: "done", CompletionEvidence: "proof", Disposition: DispositionAction, Authority: "delegated", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: project.ItemID, Class: ClassProject}); err != nil {
		t.Fatal(err)
	}
	if _, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: action.ItemID, DesiredOutcome: "done", CompletionEvidence: "proof", Disposition: DispositionAction, Authority: "delegated", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: action.ItemID, Class: ClassAction}); err != nil {
		t.Fatal(err)
	}
	if err := PutGTDRelation(ctx, store, GTDRelation{SubjectID: action.ItemID, ObjectID: project.ItemID, Kind: RelationPartOf, Source: "user", AssertionStatus: "confirmed", SourceRevision: 3, PolicyVersion: "p1"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	reflection, err := ReflectGTDStore(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if len(reflection.Items) != 2 || len(reflection.Result.NeedsNextAction) != 0 {
		t.Fatalf("reflection=%+v", reflection)
	}
	if err := SetGTDItemCancelled(ctx, store, action.ItemID, true); err != nil {
		t.Fatal(err)
	}
	reflection, err = ReflectGTDStore(ctx, store)
	if err != nil || !reflection.Result.NeedsNextAction[project.ItemID] {
		t.Fatalf("cancelled reflection=%+v err=%v", reflection, err)
	}
}

func TestReflectGTDStoreArchiveReopenCancelAndStaleRelationMutants(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	seed := func(event, content string, class GTDClass) GTDItem {
		t.Helper()
		item, err := CaptureGTDItem(ctx, store, CaptureInput{Content: content, Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: event})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, DesiredOutcome: "done", CompletionEvidence: "archive", Disposition: DispositionAction, Authority: "delegated", SourceTrusted: true}); err != nil {
			t.Fatal(err)
		}
		item, err = OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: item.ItemID, Class: class})
		if err != nil {
			t.Fatal(err)
		}
		return item
	}
	project := seed("reflect-archive-project", "project", ClassProject)
	action := seed("reflect-archive-action", "action", ClassAction)
	successor := seed("reflect-archive-successor", "successor", ClassAction)
	if err := PutGTDRelation(ctx, store, GTDRelation{SubjectID: action.ItemID, ObjectID: project.ItemID, Kind: RelationPartOf, Source: "user", AssertionStatus: "confirmed", SourceRevision: action.SourceRevision, PolicyVersion: "p1"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := PutGTDRelation(ctx, store, GTDRelation{SubjectID: successor.ItemID, ObjectID: action.ItemID, Kind: RelationDependsOn, Source: "user", AssertionStatus: "confirmed", SourceRevision: successor.SourceRevision, PolicyVersion: "p1"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	engaged, err := EngageGTDItem(ctx, store, EngageInput{ItemID: action.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || engaged.CardID == "" {
		t.Fatalf("engage=%+v err=%v", engaged, err)
	}
	view, err := ReflectGTDStore(ctx, store)
	if err != nil || !view.Result.Stale[action.ItemID] || view.Result.NeedsNextAction[project.ItemID] || !view.Result.Blocked[successor.ItemID] {
		t.Fatalf("live stale view=%+v err=%v", view.Result, err)
	}
	if err := store.Mutate(func(record *BacklogRecord) error { return record.ArchiveCard(engaged.CardID) }); err != nil {
		t.Fatal(err)
	}
	view, err = ReflectGTDStore(ctx, store)
	if err != nil || !view.Result.NeedsNextAction[project.ItemID] || view.Result.Blocked[successor.ItemID] {
		t.Fatalf("archived view=%+v err=%v", view.Result, err)
	}
	if err := store.Mutate(func(record *BacklogRecord) error { return record.RestoreCard(engaged.CardID) }); err != nil {
		t.Fatal(err)
	}
	view, err = ReflectGTDStore(ctx, store)
	if err != nil || view.Result.NeedsNextAction[project.ItemID] || !view.Result.Blocked[successor.ItemID] {
		t.Fatalf("reopened view=%+v err=%v", view.Result, err)
	}
	if err := SetGTDItemCancelled(ctx, store, action.ItemID, true); err != nil {
		t.Fatal(err)
	}
	view, err = ReflectGTDStore(ctx, store)
	if err != nil || !view.Result.NeedsNextAction[project.ItemID] || !view.Result.Blocked[successor.ItemID] {
		t.Fatalf("cancelled view=%+v err=%v", view.Result, err)
	}
}

func TestGTDPersistenceCancelledAndCorruptStateRefusals(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	item, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "x", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "cancelled-ops"})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := ExportGTD(cancelled, store, true); err == nil {
		t.Fatal("cancelled export succeeded")
	}
	if err := BackupGTDStore(cancelled, store, filepath.Join(t.TempDir(), "cancelled.db"), true); err == nil {
		t.Fatal("cancelled backup succeeded")
	}
	if err := ImportGTD(cancelled, NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json")), []byte(`{"version":1,"logical_revision":0}`), true); err == nil {
		t.Fatal("cancelled import succeeded")
	}
	if err := SetGTDItemCancelled(cancelled, store, item.ItemID, true); err == nil {
		t.Fatal("cancelled lifecycle update succeeded")
	}
	if err := DeleteGTDItem(cancelled, store, item.ItemID); err == nil {
		t.Fatal("cancelled delete succeeded")
	}
	if _, _, err := GTDProjectionSource(cancelled, store); err == nil {
		t.Fatal("cancelled projection query succeeded")
	}
	if _, err := ReflectGTDStore(cancelled, store); err == nil {
		t.Fatal("cancelled reflection succeeded")
	}

	db, err := openGTDDB(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE gtd_meta SET value='not-an-integer' WHERE key='logical_revision'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := ExportGTD(ctx, store, true); err == nil {
		t.Fatal("corrupt export revision accepted")
	}
	if _, _, err := GTDProjectionSource(ctx, store); err == nil {
		t.Fatal("corrupt projection revision accepted")
	}
}

func TestImportGTDPreservesTrustedCancelledAndLinkedFields(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	payload := gtdExport{Version: 1, LogicalRevision: 9, Items: []GTDItem{{ItemID: "gtd-1111111111111111", Content: "linked", Source: "user", Sensitivity: SensitivityPrivate, EventID: "import-linked", SourceTrusted: true, Cancelled: true, CardID: "t1", Status: GTDStatusOrganized, SourceRevision: 4}}}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := ImportGTD(ctx, store, raw, true); err != nil {
		t.Fatal(err)
	}
	item, err := LoadGTDItem(ctx, store, "gtd-1111111111111111")
	if err != nil || !item.SourceTrusted || !item.Cancelled || item.CardID != "t1" {
		t.Fatalf("imported item=%+v err=%v", item, err)
	}
}

func TestGTDWorkflowFailsClosedOnCancelledPersistenceContext(t *testing.T) {
	ctx := context.Background()
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	seed, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "seed", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "cancelled-workflow"})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := CaptureGTDItem(cancelled, store, CaptureInput{Content: "new", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "cancelled-new"}); err == nil {
		t.Fatal("capture ignored cancellation")
	}
	if _, err := ClarifyGTDItem(cancelled, store, ClarifyInput{ItemID: seed.ItemID, Disposition: DispositionReference}); err == nil {
		t.Fatal("clarify ignored cancellation")
	}
	if _, err := OrganizeGTDItem(cancelled, store, OrganizeInput{ItemID: seed.ItemID, Class: ClassReference}); err == nil {
		t.Fatal("organize ignored cancellation")
	}
	if _, err := EngageGTDItem(cancelled, store, EngageInput{ItemID: seed.ItemID}); err == nil {
		t.Fatal("engage ignored cancellation")
	}
	if err := PutGTDRelation(cancelled, store, GTDRelation{SubjectID: seed.ItemID, ObjectID: "missing", Kind: RelationRelatedTo}, nil, nil); err == nil {
		t.Fatal("relation write ignored cancellation")
	}
	op := GTDOperation{OperationID: "cancelled-op", MissionID: "mission", Action: "publish", Target: seed.ItemID, SnapshotHash: "snapshot", ReceiptJSON: []byte(`{}`)}
	if _, _, err := PrepareGTDOperation(cancelled, store, op); err == nil {
		t.Fatal("operation prepare ignored cancellation")
	}
}

func TestGTDPublicStoreOperationsRejectUnopenableEngine(t *testing.T) {
	ctx := context.Background()
	blocked := filepath.Join(t.TempDir(), "regular-file")
	if err := os.WriteFile(blocked, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	store := NewBacklogStore(filepath.Join(blocked, "backlog.json"))
	if err := MigrateGTDSchema(store); err == nil {
		t.Fatal("migration opened engine below regular file")
	}
	if _, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "x", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "blocked"}); err == nil {
		t.Fatal("capture opened invalid engine")
	}
	if _, err := LoadGTDItem(ctx, store, "gtd-1111111111111111"); err == nil {
		t.Fatal("load opened invalid engine")
	}
	if _, err := ExportGTD(ctx, store, true); err == nil {
		t.Fatal("export opened invalid engine")
	}
	if _, _, err := GTDProjectionSource(ctx, store); err == nil {
		t.Fatal("projection opened invalid engine")
	}
	if _, err := ReflectGTDStore(ctx, store); err == nil {
		t.Fatal("reflect opened invalid engine")
	}
	if err := SetGTDItemCancelled(ctx, store, "gtd-1111111111111111", true); err == nil {
		t.Fatal("lifecycle opened invalid engine")
	}
	if err := DeleteGTDItem(ctx, store, "gtd-1111111111111111"); err == nil {
		t.Fatal("delete opened invalid engine")
	}
}
