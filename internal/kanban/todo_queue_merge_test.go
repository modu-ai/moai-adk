// todo_queue_merge_test.go — the pure two-store merge core
// (SPEC-TODO-QUEUE-HOME-MERGE-001 M1, plan.md §F).
//
// Every test runs against in-memory BacklogRecord fixtures; no test touches a
// real queue. The fixture shape mirrors what LoadPure returns: items, findings,
// archive entries, runtime, and identity UUIDs projected onto each card.
package kanban

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mergeFixture builds a minimal record with the given items.
func mergeFixture(items []BacklogItem, lastSeq int) *BacklogRecord {
	rec := &BacklogRecord{
		Version:  backlogVersion,
		LastSeq:  lastSeq,
		Items:    items,
		Findings: []BacklogFinding{},
		Archived: []BacklogArchiveEntry{},
		Runtime:  TodoRuntime{Runs: []TodoRuntimeRun{}, Assignments: []TodoRuntimeAssignment{}},
	}
	return rec
}

func mergeItem(id, text string, state BacklogState) BacklogItem {
	return BacklogItem{ID: id, Text: text, AddedAt: "2026-09-13T00:00:00Z", State: state}
}

// mergeIdentity issues a real UUIDv7 for fixture identity fields.
func mergeIdentity(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("issue fixture UUIDv7: %v", err)
	}
	return id.String()
}

func idsOf(items []BacklogItem) map[string]bool {
	out := map[string]bool{}
	for _, it := range items {
		out[it.ID] = true
	}
	return out
}

func archivedIDsOf(entries []BacklogArchiveEntry) map[string]bool {
	out := map[string]bool{}
	for _, e := range entries {
		out[e.Item.ID] = true
	}
	return out
}

// TestMergeBacklogRecordsZeroLoss proves every card in either store appears in
// the merged record under its original id or a mapping row (AC-TQM-002 shape).
func TestMergeBacklogRecordsZeroLoss(t *testing.T) {
	home := mergeFixture([]BacklogItem{
		mergeItem("t1", "home only", BacklogStateQueued),
		mergeItem("t5", "shared identical", BacklogStateQueued),
	}, 5)
	project := mergeFixture([]BacklogItem{
		mergeItem("t5", "shared identical", BacklogStateQueued),
		mergeItem("t7", "project only", BacklogStateQueued),
		mergeItem("t9", "project dropped", BacklogStateDropped),
	}, 9)

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	got := idsOf(merged.Items)
	for _, want := range []string{"t1", "t7", "t9"} {
		if !got[want] {
			t.Errorf("merged items missing %s: %v", want, got)
		}
	}
	if len(report.Duplicates) != 1 || report.Duplicates[0].ID != "t5" {
		t.Errorf("want one resolved duplicate t5, got %+v", report.Duplicates)
	}
	if report.HighWater != 9 {
		t.Errorf("HighWater = %d, want 9", report.HighWater)
	}
}

// TestMergeBacklogRecordsCollisionRenumber proves a shared number with
// DIFFERENT content keeps the home card and reissues the project variant above
// the merged high-water (REQ-TQM-005/007).
func TestMergeBacklogRecordsCollisionRenumber(t *testing.T) {
	home := mergeFixture([]BacklogItem{
		mergeItem("t10", "home card ten", BacklogStateQueued),
	}, 10)
	project := mergeFixture([]BacklogItem{
		mergeItem("t10", "a DIFFERENT card that happens to share t10", BacklogStateQueued),
		mergeItem("t3", "low project card", BacklogStateQueued),
	}, 10)

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	// Home keeps t10 unchanged.
	for _, it := range merged.Items {
		if it.ID == "t10" && it.Text != "home card ten" {
			t.Fatalf("home t10 overwritten: %q", it.Text)
		}
	}
	if len(report.Renumbered) != 1 {
		t.Fatalf("want exactly one renumbered row, got %+v", report.Renumbered)
	}
	row := report.Renumbered[0]
	if row.OldID != "t10" {
		t.Errorf("mapping OldID = %q, want t10", row.OldID)
	}
	if row.NewID != "t11" {
		t.Errorf("mapping NewID = %q, want t11 (high-water 10 + 1)", row.NewID)
	}
	got := idsOf(merged.Items)
	if !got["t11"] {
		t.Errorf("renumbered card absent under t11: %v", got)
	}
	if !got["t3"] {
		t.Errorf("project-only t3 not migrated: %v", got)
	}
	if got["t3"] && merged.Items[mergedLastIndex(merged.Items, "t3")].Text != "low project card" {
		t.Errorf("migrated t3 content changed")
	}
}

func mergedLastIndex(items []BacklogItem, id string) int {
	at := -1
	for i, it := range items {
		if it.ID == id {
			at = i
		}
	}
	return at
}

// TestMergeBacklogRecordsDuplicateKeepsHome proves the content-identical pair
// leaves the home copy byte-unchanged and reports the pair (REQ-TQM-006).
func TestMergeBacklogRecordsDuplicateKeepsHome(t *testing.T) {
	home := mergeFixture([]BacklogItem{
		mergeItem("t4", "identical card", BacklogStatePicked),
	}, 4)
	project := mergeFixture([]BacklogItem{
		mergeItem("t4", "identical card", BacklogStatePicked),
	}, 4)

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	count := 0
	for _, it := range merged.Items {
		if it.ID == "t4" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("t4 appears %d times in merged items, want 1", count)
	}
	if len(report.Duplicates) != 1 || report.Duplicates[0].ID != "t4" {
		t.Errorf("want one duplicate row for t4, got %+v", report.Duplicates)
	}
}

// TestMergeBacklogRecordsFindingRewriteTokenBoundary proves a renumbered id is
// rewritten in structured finding fields AND in free-text notes, and that the
// rewrite is token-boundary aware (t642 must not corrupt t6420) (AC-TQM-004).
func TestMergeBacklogRecordsFindingRewriteTokenBoundary(t *testing.T) {
	home := mergeFixture([]BacklogItem{
		mergeItem("t642", "home six-forty-two", BacklogStateQueued),
		mergeItem("t6420", "unrelated big number", BacklogStateQueued),
		mergeItem("t2", "anchor", BacklogStateQueued),
	}, 6420)
	project := mergeFixture([]BacklogItem{
		mergeItem("t642", "PROJECT six-forty-two (different card)", BacklogStateQueued),
	}, 6420)
	project.Findings = []BacklogFinding{
		{SubjectID: "t642", RelatedID: "t2", Relation: BacklogRelationContains, Source: BacklogSourceAgent,
			Note: "project t642 contains t2; measured alongside t6420 which is unrelated"},
	}
	project.Runtime.Assignments = []TodoRuntimeAssignment{
		{RunID: "run-p1", CardID: "t642", OwnerLabel: "lane-1", ReportedState: "picked", EventKind: "assign"},
	}

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	row := report.Renumbered[0]
	newID := row.NewID

	// Structured rewrite.
	foundStructured := false
	for _, f := range merged.Findings {
		if f.SubjectID == "t642" && f.RelatedID == "t2" {
			t.Errorf("stale old id in structured finding fields: %+v", f)
		}
		if f.SubjectID == newID && f.RelatedID == "t2" {
			foundStructured = true
			// Boundary: t6420 must survive untouched inside the note, t642 rewritten.
			if !strings.Contains(f.Note, newID+" contains t2") {
				t.Errorf("note missing rewritten subject reference: %q", f.Note)
			}
			if !strings.Contains(f.Note, "t6420 which is unrelated") {
				t.Errorf("note corrupted the t6420 token: %q", f.Note)
			}
			if strings.Contains(f.Note, "t642 ") && !strings.Contains(f.Note, newID) {
				t.Errorf("note still carries the bare old token: %q", f.Note)
			}
		}
	}
	if !foundStructured {
		t.Errorf("rewritten finding missing under subject %s", newID)
	}

	// Runtime assignment rewrite.
	foundAssign := false
	for _, a := range merged.Runtime.Assignments {
		if a.CardID == "t642" {
			t.Errorf("stale old id in runtime assignment CardID: %+v", a)
		}
		if a.RunID == "run-p1" && a.CardID == newID {
			foundAssign = true
		}
	}
	if !foundAssign {
		t.Errorf("rewritten runtime assignment missing under CardID %s", newID)
	}

	// The home card t6420's own id must not have been touched.
	if !idsOf(merged.Items)["t6420"] {
		t.Errorf("t6420 lost from merged items")
	}
}

// TestMergeBacklogRecordsLastSeqLift proves the merged high-water is
// max(last_seq_home, last_seq_project, max id home, max id project) across LIVE
// and ARCHIVED populations (acceptance.md §D.5 — the merge must not itself
// mint a colliding id).
func TestMergeBacklogRecordsLastSeqLift(t *testing.T) {
	home := mergeFixture([]BacklogItem{mergeItem("t5", "home", BacklogStateQueued)}, 5)
	home.Archived = []BacklogArchiveEntry{{Item: mergeItem("t700", "archived big", BacklogStateQueued)}}
	project := mergeFixture([]BacklogItem{mergeItem("t6", "project", BacklogStateQueued)}, 6)
	project.Archived = []BacklogArchiveEntry{{Item: mergeItem("t900", "project archived big", BacklogStateQueued)}}

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if report.HighWater != 900 {
		t.Errorf("HighWater = %d, want 900 (max archived id across both stores)", report.HighWater)
	}
	if merged.LastSeq != 900 {
		t.Errorf("merged LastSeq = %d, want 900", merged.LastSeq)
	}
}

// TestMergeBacklogRecordsPopulationsPreserved proves dropped and archived
// project populations migrate (REQ-TQM-004 across queued/picked/dropped/archived).
func TestMergeBacklogRecordsPopulationsPreserved(t *testing.T) {
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	home.Archived = []BacklogArchiveEntry{{Item: mergeItem("t2", "home archived", BacklogStateQueued)}}
	project := mergeFixture([]BacklogItem{
		mergeItem("t3", "project dropped", BacklogStateDropped),
		mergeItem("t4", "project picked", BacklogStatePicked),
	}, 4)
	project.Archived = []BacklogArchiveEntry{{
		Item:     mergeItem("t5", "project archived", BacklogStateQueued),
		Findings: []BacklogArchivedFinding{},
	}}

	merged, _, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	got := idsOf(merged.Items)
	if !got["t3"] || !got["t4"] {
		t.Errorf("project dropped/picked cards missing: %v", got)
	}
	arch := archivedIDsOf(merged.Archived)
	for _, want := range []string{"t2", "t5"} {
		if !arch[want] {
			t.Errorf("archived population missing %s: %v", want, arch)
		}
	}
}

// TestMergeBacklogRecordsIdentityUUIDCollision proves the acceptance §D.5 edge:
// a project card carrying an identity UUID the home store already holds merges
// with a FRESH identity plus a reconciliation row — never an error and never a
// silent reuse. RED-FIRST test (SPEC delegation: observed failing before
// GREEN on the pre-implementation tree).
func TestMergeBacklogRecordsIdentityUUIDCollision(t *testing.T) {
	shared := mergeIdentity(t)
	home := mergeFixture([]BacklogItem{mergeItem("t8", "home card", BacklogStateQueued)}, 8)
	home.Items[0].CardUUID = &shared
	project := mergeFixture([]BacklogItem{mergeItem("t8", "DIFFERENT project card", BacklogStateQueued)}, 8)
	project.Items[0].CardUUID = &shared // same UUID, different card — the collision

	fresh := mergeIdentity(t)
	issued := 0
	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{
		NewUUID: func() (string, error) {
			issued++
			return fresh, nil
		},
	})
	if err != nil {
		t.Fatalf("merge refused an identity collision: %v", err)
	}
	if issued != 1 {
		t.Errorf("fresh UUID issued %d times, want exactly 1", issued)
	}
	// The renumbered project card must carry the FRESH identity, never the
	// home card's.
	for _, it := range merged.Items {
		if it.ID == "t8" && it.Text == "home card" {
			if it.CardUUID == nil || *it.CardUUID != shared {
				t.Errorf("home card identity disturbed: %+v", it.CardUUID)
			}
		}
	}
	row := report.Renumbered[0]
	renumbered := merged.Items[mergedLastIndex(merged.Items, row.NewID)]
	if renumbered.CardUUID == nil || *renumbered.CardUUID != fresh {
		t.Errorf("colliding card kept the contested UUID %v, want fresh %s", renumbered.CardUUID, fresh)
	}
	found := false
	for _, r := range report.Reconciliation {
		if r.CardID == "t8" && r.Kind == MergeReconcileIdentityCollision {
			found = true
		}
	}
	if !found {
		t.Errorf("no identity-collision reconciliation row: %+v", report.Reconciliation)
	}
}

// TestMergeBacklogRecordsArchivedIdentityCollision proves the same collision
// resolution reaches the archived population's identity fields.
func TestMergeBacklogRecordsArchivedIdentityCollision(t *testing.T) {
	shared := mergeIdentity(t)
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	home.Archived = []BacklogArchiveEntry{{Item: mergeItem("t2", "home archived", BacklogStateQueued)}}
	home.Archived[0].Item.CardUUID = &shared
	project := mergeFixture([]BacklogItem{mergeItem("t2", "DIFFERENT project card", BacklogStateQueued)}, 2)
	project.Items[0].CardUUID = &shared

	fresh := mergeIdentity(t)
	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{
		NewUUID: func() (string, error) { return fresh, nil },
	})
	if err != nil {
		t.Fatalf("merge refused an archived identity collision: %v", err)
	}
	row := report.Renumbered[0]
	renumbered := merged.Items[mergedLastIndex(merged.Items, row.NewID)]
	if renumbered.CardUUID == nil || *renumbered.CardUUID != fresh {
		t.Errorf("archived-colliding card kept the contested UUID %v", renumbered.CardUUID)
	}
}

// TestMergeBacklogRecordsIdentityCollisionIssuerError proves an issuer failure
// aborts the merge with an error rather than reusing the contested UUID.
func TestMergeBacklogRecordsIdentityCollisionIssuerError(t *testing.T) {
	shared := mergeIdentity(t)
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	home.Items[0].CardUUID = &shared
	project := mergeFixture([]BacklogItem{mergeItem("t1", "different project", BacklogStateQueued)}, 1)
	project.Items[0].CardUUID = &shared

	_, _, err := MergeBacklogRecords(home, project, MergeOptions{
		NewUUID: func() (string, error) { return "", errMergeTestIssuer },
	})
	if err == nil {
		t.Fatal("merge succeeded despite a failing UUID issuer")
	}
}

var errMergeTestIssuer = errorString("issuer offline")

type errorString string

func (e errorString) Error() string { return string(e) }

// TestMergeBacklogRecordsInputsUntouched proves the merge is PURE: neither
// input record is mutated (plan.md M1 "pure merge function").
func TestMergeBacklogRecordsInputsUntouched(t *testing.T) {
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	project := mergeFixture([]BacklogItem{
		mergeItem("t1", "different project card", BacklogStateQueued),
		mergeItem("t2", "project only", BacklogStateQueued),
	}, 2)
	homeBefore := home.Items[0].Text
	projectLen := len(project.Items)

	if _, _, err := MergeBacklogRecords(home, project, MergeOptions{}); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if home.Items[0].Text != homeBefore || len(home.Items) != 1 || home.LastSeq != 1 {
		t.Errorf("home input mutated: %+v", home)
	}
	if len(project.Items) != projectLen || project.LastSeq != 2 {
		t.Errorf("project input mutated: %+v", project)
	}
}

// TestMergeBacklogRecordsCollisionWithArchive proves a project LIVE card whose
// number collides with a home ARCHIVED entry is renumbered, not silently
// stacked onto the archived id-space.
func TestMergeBacklogRecordsCollisionWithArchive(t *testing.T) {
	home := mergeFixture([]BacklogItem{mergeItem("t1", "home", BacklogStateQueued)}, 1)
	home.Archived = []BacklogArchiveEntry{{Item: mergeItem("t3", "home archived three", BacklogStateQueued)}}
	project := mergeFixture([]BacklogItem{mergeItem("t3", "project live three (different)", BacklogStateQueued)}, 3)

	merged, report, err := MergeBacklogRecords(home, project, MergeOptions{})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(report.Renumbered) != 1 || report.Renumbered[0].NewID != "t4" {
		t.Fatalf("want renumber of project t3 to t4, got %+v", report.Renumbered)
	}
	if archivedIDsOf(merged.Archived)["t4"] {
		t.Errorf("project live card landed in the archive population")
	}
	if !idsOf(merged.Items)["t4"] {
		t.Errorf("renumbered card missing from live items")
	}
}
