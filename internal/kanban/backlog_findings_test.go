// backlog_findings_test.go — the findings helper family
// (SPEC-TODO-SQLITE-001 C-1, M6).
//
// These helpers encode queue semantics that nothing else enforces: which
// findings a card "names", when two findings describe the same relation, and
// the 1-based index `todo unrelate` addresses. They were reachable only from
// the CLI package's tests, so the rules below were asserted through a command
// surface rather than directly — which meant a change to the rule and a
// compensating change to the command could pass together.
package kanban

import (
	"testing"
)

func finding(subject, related, relation, source string) BacklogFinding {
	return BacklogFinding{SubjectID: subject, RelatedID: related, Relation: relation, Source: source, At: "T"}
}

// Names is positional-agnostic: a relation names a card whether the card sits
// on the subject side or the related side.
func TestBacklogFindingNames(t *testing.T) {
	t.Parallel()
	f := finding("t1", "t2", BacklogRelationContains, BacklogSourceAgent)
	for _, tc := range []struct {
		id   string
		want bool
	}{{"t1", true}, {"t2", true}, {"t3", false}, {"", false}} {
		if got := f.Names(tc.id); got != tc.want {
			t.Errorf("Names(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

// SamePairAs is UNORDERED: a relation between two cards is a property of the
// pair, not of the direction it happened to be written in. An agent recording
// {b, a} is answering a mechanical finding about {a, b}, and a direction-
// sensitive comparison would let both sit in the listing as if they were
// separate observations.
func TestBacklogFindingSamePairAs(t *testing.T) {
	t.Parallel()
	base := finding("t1", "t2", BacklogRelationNearDuplicate, BacklogSourceMechanical)

	if !base.SamePairAs(finding("t2", "t1", BacklogRelationContains, BacklogSourceAgent)) {
		t.Error("SamePairAs({t2,t1}) = false, want true — the pair is unordered")
	}
	if !base.SamePairAs(finding("t1", "t2", BacklogRelationReplaces, BacklogSourceAgent)) {
		t.Error("SamePairAs(same order, different relation) = false, want true — only the pair is compared")
	}
	if base.SamePairAs(finding("t1", "t3", BacklogRelationContains, BacklogSourceAgent)) {
		t.Error("SamePairAs({t1,t3}) = true, want false")
	}
}

// The tuple key is {subject, related, relation, source} — deliberately NOT the
// timestamp or the note. Re-running the analyser must not stack a second copy
// of a relation it already recorded, or the listing fills with duplicates of
// one measurement and stops being read.
func TestBacklogRecordFindingTupleAndAppendOnce(t *testing.T) {
	t.Parallel()
	rec := &BacklogRecord{}
	f := finding("t1", "t2", BacklogRelationNearDuplicate, BacklogSourceMechanical)

	if rec.HasFindingTuple(f) {
		t.Error("HasFindingTuple on an empty record = true, want false")
	}
	if !rec.AppendFindingOnce(f) {
		t.Fatal("first AppendFindingOnce = false, want true")
	}
	if !rec.HasFindingTuple(f) {
		t.Error("HasFindingTuple after append = false, want true")
	}

	// Same tuple, different timestamp and note: still a duplicate.
	again := f
	again.At = "LATER"
	again.Note = "re-measured"
	if rec.AppendFindingOnce(again) {
		t.Error("AppendFindingOnce(same tuple, new timestamp) = true, want false")
	}
	if len(rec.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(rec.Findings))
	}

	// A different SOURCE is a different record: an agent's judgment about a
	// pair does not collide with the machine's measurement of it.
	agentSide := finding("t1", "t2", BacklogRelationNearDuplicate, BacklogSourceAgent)
	if !rec.AppendFindingOnce(agentSide) {
		t.Error("AppendFindingOnce(different source) = false, want true")
	}
	if len(rec.Findings) != 2 {
		t.Fatalf("findings = %d, want 2", len(rec.Findings))
	}
}

// FindingsNaming returns 1-BASED indexes into the record — the addresses
// `todo unrelate` takes. Off-by-one here silently unrelates the wrong finding.
func TestBacklogRecordFindingsNaming(t *testing.T) {
	t.Parallel()
	rec := &BacklogRecord{Findings: []BacklogFinding{
		finding("t1", "t2", BacklogRelationContains, BacklogSourceAgent),
		finding("t3", "t4", BacklogRelationReplaces, BacklogSourceAgent),
		finding("t2", "t5", BacklogRelationConflicts, BacklogSourceAgent),
	}}

	found, indexes := rec.FindingsNaming("t2")
	if len(found) != 2 || len(indexes) != 2 {
		t.Fatalf("FindingsNaming(t2) = %d findings / %d indexes, want 2 / 2", len(found), len(indexes))
	}
	if indexes[0] != 1 || indexes[1] != 3 {
		t.Fatalf("indexes = %v, want [1 3] (1-based positions in the record)", indexes)
	}
	if found[0].RelatedID != "t2" || found[1].SubjectID != "t2" {
		t.Errorf("FindingsNaming returned the wrong rows: %+v", found)
	}

	if got, idx := rec.FindingsNaming("t99"); got != nil || idx != nil {
		t.Errorf("FindingsNaming(absent) = %v / %v, want nil / nil", got, idx)
	}
}

// RemoveFindingsNaming runs when a card leaves the queue: a finding that
// outlives its subject points at nothing, and the listing would render a
// relation to a card the operator can no longer see.
func TestBacklogRecordRemoveFindingsNaming(t *testing.T) {
	t.Parallel()
	rec := &BacklogRecord{Findings: []BacklogFinding{
		finding("t1", "t2", BacklogRelationContains, BacklogSourceAgent),
		finding("t3", "t4", BacklogRelationReplaces, BacklogSourceAgent),
		finding("t2", "t5", BacklogRelationConflicts, BacklogSourceAgent),
	}}

	if removed := rec.RemoveFindingsNaming("t2"); removed != 2 {
		t.Fatalf("RemoveFindingsNaming(t2) = %d, want 2", removed)
	}
	if len(rec.Findings) != 1 || rec.Findings[0].SubjectID != "t3" {
		t.Fatalf("surviving findings = %+v, want only the {t3,t4} row", rec.Findings)
	}
	if removed := rec.RemoveFindingsNaming("t99"); removed != 0 {
		t.Errorf("RemoveFindingsNaming(absent) = %d, want 0", removed)
	}
	if rec.Findings == nil {
		t.Error("RemoveFindingsNaming left a nil slice; the record always renders findings as an array")
	}
}

// HasAgentFindingForPair is the predicate behind the `machine-only` mark. It
// records the ABSENCE of an agent-sourced record for a pair — never that any
// review took place — so it must match the pair UNORDERED and must ignore
// mechanical rows entirely.
func TestBacklogRecordHasAgentFindingForPair(t *testing.T) {
	t.Parallel()
	mechanical := finding("t1", "t2", BacklogRelationNearDuplicate, BacklogSourceMechanical)

	rec := &BacklogRecord{Findings: []BacklogFinding{mechanical}}
	if rec.HasAgentFindingForPair(mechanical) {
		t.Error("a mechanical row satisfied the agent predicate — the mark would claim a review that never happened")
	}

	rec.Findings = append(rec.Findings, finding("t2", "t1", BacklogRelationAbsorbs, BacklogSourceAgent))
	if !rec.HasAgentFindingForPair(mechanical) {
		t.Error("an agent row on the reversed pair did not satisfy the predicate — the pair is unordered")
	}

	other := finding("t1", "t9", BacklogRelationNearDuplicate, BacklogSourceMechanical)
	if rec.HasAgentFindingForPair(other) {
		t.Error("an unrelated pair satisfied the predicate")
	}
}
