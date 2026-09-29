// todo_relation_filter_test.go — SPEC-RELATION-PICKUP-FILTER-001 (card t1343):
// the blocks/depends relations consumed as a self-dispatch pickup filter and
// the todo-relate cycle guard.
//
// Every test runs on a t.TempDir() fixture; the operator's live queue is
// never a subject. RED-first per milestone — each test's failing state was
// observed on the pre-implementation tree (HEAD 28e672b39) before the
// implementation that flips it.
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestAutoPickTargetsRelationBlocked — AC-RPF-001/002 (REQ-RPF-001/002, M1):
// a queued card named on the BLOCKED side of a recorded sequencing finding is
// not a pickup candidate; only its predecessor remains. Direction semantics
// per REQ-RPF-002: `depends` blocks its SUBJECT, `blocks` blocks its
// RELATED. A single-direction implementation must fail one arm of this
// table (plan.md E-2).
func TestAutoPickTargetsRelationBlocked(t *testing.T) {
	cases := []struct {
		name        string
		subject     string
		relation    string
		related     string
		wantBlocked string
		wantKept    string
	}{
		{"depends blocks its subject", "t1", kanban.BacklogRelationDepends, "t2", "t1", "t2"},
		{"blocks blocks its related", "t1", kanban.BacklogRelationBlocks, "t2", "t2", "t1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := todoFixture(t)
			seedItems(t, store, "predecessor card", "successor card")
			seedFindings(t, store, kanban.BacklogFinding{
				SubjectID: tc.subject, RelatedID: tc.related,
				Relation: tc.relation, Source: kanban.BacklogSourceAgent,
			})
			rec, err := store.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			targets, _, err := autoPickTargets(rec, autoTestLiveness(root, "t1", true, true, nil), root)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, it := range targets {
				got = append(got, it.ID)
			}
			if strings.Join(got, ",") != tc.wantKept {
				t.Fatalf("pickup = [%s], want [%s] — %s must be excluded while %s stays a candidate",
					strings.Join(got, ","), tc.wantKept, tc.wantBlocked, tc.wantKept)
			}
		})
	}
}

// TestAutoPickTargetsReturnsAfterDone — AC-RPF-003 (REQ-RPF-003, M1): the
// predecessor's done moves the finding into the archive (existing
// ArchiveCard behavior), and the previously blocked card is a candidate
// again with no relation bookkeeping. A state-scan implementation re-opens
// the dropped/hold semantics this SPEC deliberately does not touch — this
// test catches it (spec.md B.1).
//
// Given: `t2 depends t1` — t2 (the subject) waits on t1, so t2 is the
// blocked side; when t1 (the related predecessor) is done, t2 returns.
func TestAutoPickTargetsReturnsAfterDone(t *testing.T) {
	root, store := todoFixture(t)
	seedItems(t, store, "predecessor card", "successor card")
	seedFindings(t, store, kanban.BacklogFinding{
		SubjectID: "t2", RelatedID: "t1",
		Relation: kanban.BacklogRelationDepends, Source: kanban.BacklogSourceAgent,
	})
	lv := autoTestLiveness(root, "t1", true, true, nil)

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	targets, _, err := autoPickTargets(rec, lv, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range targets {
		if it.ID == "t2" {
			t.Fatalf("blocked card t2 was a candidate before the predecessor's done")
		}
	}

	// done the predecessor the way `done` does: ArchiveCard moves every
	// finding naming it into the archive entry.
	if err := store.Mutate(func(r *kanban.BacklogRecord) error {
		return r.ArchiveCard("t1")
	}); err != nil {
		t.Fatal(err)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Findings) != 0 {
		t.Fatalf("findings survived the archive move: %+v", rec.Findings)
	}
	targets, _, err = autoPickTargets(rec, lv, root)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range targets {
		got = append(got, it.ID)
	}
	if len(got) != 1 || got[0] != "t2" {
		t.Fatalf("after the predecessor's done the successor is not the only candidate: [%s]",
			strings.Join(got, ","))
	}
}
