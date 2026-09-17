package kanban

import (
	"context"
	"testing"
)

func TestCaptureGTDItem(t *testing.T) {
	if err := MigrateGTDSchema(nil); err == nil {
		t.Fatal("nil GTD store migrated")
	}
	store := newGTDTestStore(t)
	input := CaptureInput{Content: "review queue", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "capture-1"}
	first, err := CaptureGTDItem(context.Background(), store, input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := CaptureGTDItem(context.Background(), store, input)
	if err != nil || second.ItemID != first.ItemID {
		t.Fatalf("duplicate capture = %+v err=%v, want same item %s", second, err, first.ItemID)
	}
	rec, err := store.Load()
	if err != nil || len(rec.Items) != 0 {
		t.Fatalf("capture created execution card: items=%d err=%v", len(rec.Items), err)
	}
	for _, denied := range []CaptureInput{
		{Content: "x", Source: "feed", SourceAllowed: false, Sensitivity: SensitivityPrivate, EventID: "capture-2"},
		{Content: "x", Source: "user", SourceAllowed: true, Sensitivity: SensitivityUnknown, EventID: "capture-3"},
	} {
		if _, err := CaptureGTDItem(context.Background(), store, denied); err == nil {
			t.Fatalf("CaptureGTDItem(%+v) unexpectedly allowed", denied)
		}
	}
	if _, err := CaptureGTDItem(context.Background(), store, CaptureInput{Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "incomplete"}); err == nil {
		t.Fatal("incomplete capture allowed")
	}
}

func TestClarifyGTDItem(t *testing.T) {
	store := newGTDTestStore(t)
	item, err := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "ship feature", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "clarify-1"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ClarifyGTDItem(context.Background(), store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "feature shipped", CompletionEvidence: "CI and landed SHA", Authority: "code,commit", SourceTrusted: true})
	if err != nil || !result.Publishable {
		t.Fatalf("clarified action = %+v err=%v", result, err)
	}
	held, err := ClarifyGTDItem(context.Background(), store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "", CompletionEvidence: "", Authority: "", SourceTrusted: false})
	if err != nil || held.Publishable || len(held.HoldReasons) == 0 {
		t.Fatalf("ambiguous clarification = %+v err=%v", held, err)
	}
	if _, err := ClarifyGTDItem(context.Background(), store, ClarifyInput{}); err == nil {
		t.Fatal("missing clarify item allowed")
	}
	if _, err := ClarifyGTDItem(context.Background(), store, ClarifyInput{ItemID: "missing", Disposition: DispositionReference}); err == nil {
		t.Fatal("unknown clarify item allowed")
	}
}

func TestOrganizeGTDItem(t *testing.T) {
	store := newGTDTestStore(t)
	item, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "next action", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "organize-1"})
	_, _ = ClarifyGTDItem(context.Background(), store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "done", CompletionEvidence: "test", Authority: "queue", SourceTrusted: true})
	organized, err := OrganizeGTDItem(context.Background(), store, OrganizeInput{ItemID: item.ItemID, Class: ClassAction, Context: "computer", ReviewAt: "2026-09-16T00:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if organized.Class != ClassAction || organized.Context != "computer" || organized.Status != GTDStatusOrganized {
		t.Fatalf("organized item = %+v", organized)
	}
	if _, err := OrganizeGTDItem(context.Background(), store, OrganizeInput{ItemID: "missing", Class: ClassAction}); err == nil {
		t.Fatal("missing organize target allowed")
	}
	if _, err := OrganizeGTDItem(context.Background(), store, OrganizeInput{}); err == nil {
		t.Fatal("incomplete organize allowed")
	}
	unclarified, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "raw", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "organize-raw"})
	if _, err := OrganizeGTDItem(context.Background(), store, OrganizeInput{ItemID: unclarified.ItemID, Class: ClassAction}); err == nil {
		t.Fatal("unclarified item organized")
	}
}

func TestGTDEnumsAndOrganizeRelationsAreAtomic(t *testing.T) {
	ctx := context.Background()
	store := newGTDTestStore(t)
	item, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "atomic", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "atomic-organize"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, Disposition: GTDDisposition("executable_magic"), SourceTrusted: true}); err == nil {
		t.Fatal("unknown disposition persisted")
	}
	clarified, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "done", CompletionEvidence: "CI", Authority: "queue", SourceTrusted: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: item.ItemID, Class: GTDClass("mystery")}); err == nil {
		t.Fatal("unknown class persisted")
	}
	revisionBefore, _, err := GTDProjectionSource(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	_, err = OrganizeGTDItemWithRelations(ctx, store, OrganizeInput{ItemID: item.ItemID, Class: ClassAction}, []GTDRelation{{SubjectID: item.ItemID, ObjectID: "gtd-0000000000000000", Kind: RelationPartOf, Source: "operator", AssertionStatus: "confirmed", SourceRevision: clarified.Item.SourceRevision, PolicyVersion: "gtd-v1"}})
	if err == nil {
		t.Fatal("missing relation target accepted")
	}
	after, err := LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	revisionAfter, relations, err := GTDProjectionSource(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != GTDStatusClarified || after.SourceRevision != clarified.Item.SourceRevision || revisionAfter != revisionBefore || len(relations) != 0 {
		t.Fatalf("partial organize: item=%+v logical=%d/%d relations=%v", after, revisionBefore, revisionAfter, relations)
	}
}

func TestOrganizeGTDItemWithRelationsSuccessDefaultsAndCancellation(t *testing.T) {
	ctx := context.Background()
	store := newGTDTestStore(t)
	makeClarified := func(event string) GTDItem {
		item, err := CaptureGTDItem(ctx, store, CaptureInput{Content: event, Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: event})
		if err != nil {
			t.Fatal(err)
		}
		result, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "done", CompletionEvidence: "CI", Authority: "queue", SourceTrusted: true})
		if err != nil {
			t.Fatal(err)
		}
		return result.Item
	}
	subject, object := makeClarified("organize-subject"), makeClarified("organize-object")
	organized, err := OrganizeGTDItemWithRelations(ctx, store, OrganizeInput{ItemID: subject.ItemID, Class: ClassAction, Context: "@code", ReviewAt: "weekly"}, []GTDRelation{{ObjectID: object.ItemID, Kind: RelationDependsOn, Source: "operator", AssertionStatus: "confirmed", PolicyVersion: "gtd-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	if organized.Status != GTDStatusOrganized || organized.SourceRevision != subject.SourceRevision+1 {
		t.Fatalf("organized=%+v", organized)
	}
	db, err := openGTDDB(store)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	relations, err := listGTDRelations(ctx, db)
	if err != nil || len(relations) != 1 || relations[0].SubjectID != subject.ItemID || relations[0].SourceRevision != subject.SourceRevision+1 || len(relations[0].Note) != 0 {
		t.Fatalf("relations=%+v err=%v", relations, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := OrganizeGTDItemWithRelations(cancelled, store, OrganizeInput{ItemID: object.ItemID, Class: ClassAction}, nil); err == nil {
		t.Fatal("cancelled organize accepted")
	}
}

func TestReflectGTDState(t *testing.T) {
	state := ReflectInput{
		Items: []ReflectItem{
			{ItemID: "done", Completed: true},
			{ItemID: "cancelled", Cancelled: true},
			{ItemID: "successor", DependsOn: []string{"cancelled"}},
			{ItemID: "project-empty", Class: ClassProject},
			{ItemID: "project-active", Class: ClassProject},
			{ItemID: "project-finished", Class: ClassProject},
			{ItemID: "project-cancelled", Class: ClassProject},
			{ItemID: "project-reopened", Class: ClassProject},
			{ItemID: "action-active", Class: ClassAction},
			{ItemID: "action-finished", Class: ClassAction, Completed: true},
			{ItemID: "action-cancelled", Class: ClassAction, Cancelled: true},
			{ItemID: "action-reopened", Class: ClassAction},
			{ItemID: "evidence", EvidenceRevision: 1, CurrentEvidenceRevision: 2},
		},
		Relations: []GTDRelation{
			{SubjectID: "action-active", ObjectID: "project-active", Kind: RelationPartOf},
			{SubjectID: "action-finished", ObjectID: "project-finished", Kind: RelationPartOf},
			{SubjectID: "action-cancelled", ObjectID: "project-cancelled", Kind: RelationPartOf},
			{SubjectID: "action-reopened", ObjectID: "project-reopened", Kind: RelationPartOf},
		},
	}
	result := ReflectGTDState(state)
	if !result.Blocked["successor"] || !result.NeedsNextAction["project-empty"] || !result.Stale["evidence"] {
		t.Fatalf("reflect result = %+v", result)
	}
	if result.NeedsNextAction["project-active"] || result.NeedsNextAction["project-reopened"] {
		t.Fatalf("active or reopened action ignored: %+v", result.NeedsNextAction)
	}
	if !result.NeedsNextAction["project-finished"] || !result.NeedsNextAction["project-cancelled"] {
		t.Fatalf("finished-only or cancelled action treated active: %+v", result.NeedsNextAction)
	}
}

func TestReflectGTDStoreDetectsStaleRelationRevision(t *testing.T) {
	ctx := context.Background()
	store := newGTDTestStore(t)
	project, _ := CaptureGTDItem(ctx, store, CaptureInput{Content: "project", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "reflect-project"})
	action, _ := CaptureGTDItem(ctx, store, CaptureInput{Content: "action", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "reflect-action"})
	for _, item := range []GTDItem{project, action} {
		_, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "done", CompletionEvidence: "CI", Authority: "queue", SourceTrusted: true})
		if err != nil {
			t.Fatal(err)
		}
		class := ClassAction
		if item.ItemID == project.ItemID {
			class = ClassProject
		}
		if _, err := OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: item.ItemID, Class: class}); err != nil {
			t.Fatal(err)
		}
	}
	if err := PutGTDRelation(ctx, store, GTDRelation{SubjectID: action.ItemID, ObjectID: project.ItemID, Kind: RelationPartOf, Source: "test", AssertionStatus: "confirmed", SourceRevision: 1, PolicyVersion: "gtd-v1"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	view, err := ReflectGTDStore(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	if !view.Result.Stale[action.ItemID] {
		t.Fatalf("stale relation revision not detected: %+v", view.Result)
	}
}

func TestEngageGTDItem(t *testing.T) {
	store := newGTDTestStore(t)
	item, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "publish me", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "engage-1"})
	_, _ = ClarifyGTDItem(context.Background(), store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "done", CompletionEvidence: "test", Authority: "queue", SourceTrusted: true})
	_, _ = OrganizeGTDItem(context.Background(), store, OrganizeInput{ItemID: item.ItemID, Class: ClassAction, Context: "computer"})
	blocked, err := EngageGTDItem(context.Background(), store, EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || len(blocked.Reasons) != 1 || blocked.Reasons[0] != "dependencies_blocked" {
		t.Fatalf("dependency gate = %+v err=%v", blocked, err)
	}
	dry, err := EngageGTDItem(context.Background(), store, EngageInput{ItemID: item.ItemID, DryRun: true, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || dry.CardID != "" || !dry.Actionable {
		t.Fatalf("dry engage = %+v err=%v", dry, err)
	}
	rec, _ := store.Load()
	if len(rec.Items) != 0 {
		t.Fatalf("dry run mutated queue: %+v", rec.Items)
	}
	applied, err := EngageGTDItem(context.Background(), store, EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || applied.CardID == "" {
		t.Fatalf("applied engage = %+v err=%v", applied, err)
	}
	again, err := EngageGTDItem(context.Background(), store, EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || again.CardID != applied.CardID {
		t.Fatalf("idempotent engage = %+v err=%v", again, err)
	}
	held, _ := CaptureGTDItem(context.Background(), store, CaptureInput{Content: "held", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "engage-held"})
	denied, err := EngageGTDItem(context.Background(), store, EngageInput{ItemID: held.ItemID})
	if err != nil || len(denied.Reasons) < 5 {
		t.Fatalf("fail-closed engage=%+v err=%v", denied, err)
	}
}

func TestEngageGTDItemRecoversPublishedCardBeforeLink(t *testing.T) {
	ctx := context.Background()
	store := newGTDTestStore(t)
	item, err := CaptureGTDItem(ctx, store, CaptureInput{Content: "recover published card", Source: "user", SourceAllowed: true, Sensitivity: SensitivityPrivate, EventID: "engage-crash"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ClarifyGTDItem(ctx, store, ClarifyInput{ItemID: item.ItemID, Disposition: DispositionAction, DesiredOutcome: "landed", CompletionEvidence: "ci", Authority: "queue", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := OrganizeGTDItem(ctx, store, OrganizeInput{ItemID: item.ItemID, Class: ClassAction}); err != nil {
		t.Fatal(err)
	}

	identity := gtdCardUUID(item.ItemID)
	card, _, err := store.addWithCardUUID(item.Content, &identity)
	if err != nil {
		t.Fatal(err)
	}
	result, err := EngageGTDItem(ctx, store, EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || result.CardID != card.ID {
		t.Fatalf("crash recovery=%+v want card=%s err=%v", result, card.ID, err)
	}
	if linked, err := linkGTDCard(ctx, store, item.ItemID, card.ID); err != nil || linked != card.ID {
		t.Fatalf("same-card concurrent link=%s err=%v", linked, err)
	}
	record, err := store.Load()
	if err != nil || len(record.Items) != 1 {
		t.Fatalf("duplicate publish after recovery: items=%d err=%v", len(record.Items), err)
	}
	if err := store.Mutate(func(record *BacklogRecord) error { return record.ArchiveCard(card.ID) }); err != nil {
		t.Fatal(err)
	}
	archived, err := findCardByUUID(store, identity)
	if err != nil || archived == nil || archived.ID != card.ID {
		t.Fatalf("archived identity readback=%+v err=%v", archived, err)
	}
	if _, err := EngageGTDItem(ctx, store, EngageInput{ItemID: "missing", Authorized: true}); err == nil {
		t.Fatal("missing GTD item engaged")
	}
}
