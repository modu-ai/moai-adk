package factory

import (
	"strings"
	"testing"
)

// card t1454 card-review r2 finding 12: merged-into is a CURRENT vocabulary
// kind todo merge writes, but MapLegacyRelation's default branch remapped it
// to relates-to with a qualifier — so todo merge's own cycle guard, which
// reads the mapped kind, saw no merged-into edges at all.
func TestMapLegacyRelationKeepsMergedInto(t *testing.T) {
	kind, qualifier := MapLegacyRelation("merged-into")
	if kind != CardRelationMergedInto {
		t.Fatalf("merged-into mapped to %q (qualifier %q), want the merged-into kind preserved", kind, qualifier)
	}
	rec := &BacklogRecord{Findings: []BacklogFinding{{SubjectID: "t2", RelatedID: "t1", Relation: "merged-into"}}}
	if !rec.RelationKindClosesCycle("t1", "t2", string(CardRelationMergedInto)) {
		t.Fatal("t1 → t2 over a stored t2 merged-into t1 did not read as a merged-into cycle — the guard is dead")
	}
}

// card t1454 card-review r2 finding 13: the trace walk renders the STORED
// direction. t1 blocks t2 traced from t2 must read "t1 → t2", not the walk
// order "t2 → t1" that reads as the semantic reverse.
func TestTraceCardRelationsPreservesStoredDirection(t *testing.T) {
	rec := &BacklogRecord{Findings: []BacklogFinding{{SubjectID: "t1", RelatedID: "t2", Relation: "blocks"}}}
	lines := TraceCardRelations(rec, nil, "t2", nil, 0)
	if len(lines) == 0 {
		t.Fatal("trace from t2 found nothing")
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "blocks  t1 → t2") {
		t.Fatalf("trace from t2 rendered the walk order, not the stored direction:\n%s", joined)
	}
}

// card t1454 card-review r2 finding 15: the relate dedup normalizes the
// STORED rows too — a pair an older writer recorded in the opposite order
// still maps onto the first record.
func TestRecordRelationNormalizesStoredRowsForDedup(t *testing.T) {
	rec := &BacklogRecord{Findings: []BacklogFinding{{SubjectID: "t9", RelatedID: "t1", Relation: "relates-to"}}}
	if _, err := RecordRelation(rec, "t1", "t9", "relates-to"); err == nil {
		t.Fatal("the opposite-order re-record of an unnormalized stored pair was accepted")
	}
}

// card t1454 card-review r2 finding 11, the factory half of the common
// resolver: the GTD relations the caller resolved onto card ids join the
// findings and the spawned_by projections, and the walk renders the GTD
// edge with its own kind.
func TestTraceCardRelationsCoversGTDEdges(t *testing.T) {
	rec := &BacklogRecord{
		Findings: []BacklogFinding{{SubjectID: "t1", RelatedID: "t2", Relation: "blocks"}},
		Items: []BacklogItem{{
			ID: "t3", State: BacklogStatePicked,
			Issuance: &BacklogIssuance{SpawnedBy: "t1", Origin: "follow-up"},
		}},
	}
	gtd := []GTDCardRelation{{From: "t4", To: "t2", Kind: "depends_on", Source: "gtd"}}
	lines := TraceCardRelations(rec, gtd, "t2", nil, 0)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"blocks  t1 → t2", "follow-up-of  t3 → t1", "depends_on  t4 → t2"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("trace t2 missing %q in:\n%s", want, joined)
		}
	}
}

// The follow-up projection reads child → origin through the resolver, the
// parent projection parent → child (card t1454 card-review r2 finding 14).
func TestResolveCardEdgesProjectionDirections(t *testing.T) {
	rec := &BacklogRecord{Items: []BacklogItem{
		{ID: "t5", State: BacklogStatePicked, Issuance: &BacklogIssuance{SpawnedBy: "t1", Origin: "follow-up"}},
		{ID: "t6", State: BacklogStatePicked, Issuance: &BacklogIssuance{SpawnedBy: "t1", Origin: "split"}},
	}}
	follow := ResolveCardEdges(rec, nil, "t5")
	if len(follow) != 1 || follow[0].Kind != string(CardRelationFollowUpOf) || follow[0].From != "t5" || follow[0].To != "t1" {
		t.Fatalf("follow-up edge = %+v, want t5 follow-up-of t1 (child → origin)", follow)
	}
	parent := ResolveCardEdges(rec, nil, "t6")
	if len(parent) != 1 || parent[0].Kind != string(CardRelationParentOf) || parent[0].From != "t1" || parent[0].To != "t6" {
		t.Fatalf("parent edge = %+v, want t1 parent-of t6 (parent → child)", parent)
	}
}
