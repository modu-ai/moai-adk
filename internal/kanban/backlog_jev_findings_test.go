// backlog_jev_findings_test.go — SPEC-JEV-CONSUMERS-001 M4 (Consumer C):
// the third finding source and the source-agnostic unordered predicate the
// settled precedence rule needs.
//
// Two claims are separated here on purpose, because one is satisfiable by a
// mistake the other is not: AC-JEVN-001 asserts HasAgentFindingForPair stays
// FALSE on a Jev-only pair, which a predicate that always returned false would
// also satisfy — so it carries a positive control. AC-JEVN-003 asserts the
// precedence rule in BOTH arrival directions, because the direction whose
// failure is silent (a wrongly-suppressed measurement) leaves no trace at any
// surface a reader looks at.
package kanban

import "testing"

// jevFinding builds a Jev-sourced finding for the given ordered pair.
func jevFinding(subject, related, relation string) BacklogFinding {
	return BacklogFinding{
		SubjectID: subject, RelatedID: related,
		Relation: relation, Source: BacklogSourceJev,
		Score: 0.87, At: "2026-01-01T00:00:00Z",
	}
}

// TestBacklogSourceJev_IsAThirdConstant — AC-JEVN-002 (REQ-JEVN-002): the
// Jev source is a THIRD value, distinct from both existing ones. Asserting
// only "it is not mechanical" would pass for a constant accidentally set to
// "agent", which is the exact substitution REQ-JEVN-003 forbids.
func TestBacklogSourceJev_IsAThirdConstant(t *testing.T) {
	if BacklogSourceJev == BacklogSourceMechanical {
		t.Errorf("BacklogSourceJev == BacklogSourceMechanical (%q)", BacklogSourceJev)
	}
	if BacklogSourceJev == BacklogSourceAgent {
		t.Errorf("BacklogSourceJev == BacklogSourceAgent (%q)", BacklogSourceJev)
	}
	if BacklogSourceJev == "" {
		t.Error("BacklogSourceJev is empty — an empty source is indistinguishable from an unset field")
	}
}

// TestHasAgentFindingForPair_FalseOnJevOnlyPair — AC-JEVN-001 (REQ-JEVN-004,
// and the predicate half of REQ-JEVN-003). The positive control is the point:
// a predicate that always returned false would satisfy the absence assertion
// alone, so the same record shape is measured with an agent-sourced finding
// and must return true.
func TestHasAgentFindingForPair_FalseOnJevOnlyPair(t *testing.T) {
	probe := jevFinding("t1", "t2", BacklogRelationNearDuplicate)

	jevOnly := &BacklogRecord{Findings: []BacklogFinding{
		jevFinding("t1", "t2", BacklogRelationNearDuplicate),
		jevFinding("t2", "t1", BacklogRelationContains),
	}}
	if jevOnly.HasAgentFindingForPair(probe) {
		t.Error("HasAgentFindingForPair = true on a pair whose only findings are Jev-sourced — " +
			"the machine-only mark would be cleared for a pair nobody reviewed")
	}

	// Positive control: the same call on an agent-sourced record must fire.
	withAgent := &BacklogRecord{Findings: []BacklogFinding{
		jevFinding("t1", "t2", BacklogRelationNearDuplicate),
		{SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationContains, Source: BacklogSourceAgent},
	}}
	if !withAgent.HasAgentFindingForPair(probe) {
		t.Fatal("positive control failed: HasAgentFindingForPair = false on a pair carrying an " +
			"agent-sourced finding, so the absence measured above is unattributable")
	}
}

// TestHasFindingForPairAnySource — REQ-JEVN-006 half (a) needs a predicate
// keyed on the UNORDERED pair plus the relation and NOT on Source. Each
// dimension is exercised separately so a predicate that ignored one of them
// cannot pass: source-agnostic (all three sources match), unordered (the
// reversed pair matches), relation-keyed (a different relation does not), and
// pair-keyed (a different pair does not).
func TestHasFindingForPairAnySource(t *testing.T) {
	probe := jevFinding("t1", "t2", BacklogRelationNearDuplicate)

	for _, source := range []string{BacklogSourceMechanical, BacklogSourceAgent, BacklogSourceJev} {
		rec := &BacklogRecord{Findings: []BacklogFinding{
			{SubjectID: "t1", RelatedID: "t2", Relation: BacklogRelationNearDuplicate, Source: source},
		}}
		if !rec.HasFindingForPairAnySource(probe) {
			t.Errorf("HasFindingForPairAnySource = false for an existing %s finding — the predicate "+
				"must not key on Source", source)
		}
	}

	reversed := &BacklogRecord{Findings: []BacklogFinding{
		{SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationNearDuplicate, Source: BacklogSourceMechanical},
	}}
	if !reversed.HasFindingForPairAnySource(probe) {
		t.Error("HasFindingForPairAnySource = false for the reversed pair — the comparison must be unordered")
	}

	otherRelation := &BacklogRecord{Findings: []BacklogFinding{
		{SubjectID: "t1", RelatedID: "t2", Relation: BacklogRelationContains, Source: BacklogSourceMechanical},
	}}
	if otherRelation.HasFindingForPairAnySource(probe) {
		t.Error("HasFindingForPairAnySource = true for a DIFFERENT relation on the same pair — " +
			"the relation is part of the key")
	}

	otherPair := &BacklogRecord{Findings: []BacklogFinding{
		{SubjectID: "t1", RelatedID: "t3", Relation: BacklogRelationNearDuplicate, Source: BacklogSourceMechanical},
	}}
	if otherPair.HasFindingForPairAnySource(probe) {
		t.Error("HasFindingForPairAnySource = true for a different pair")
	}

	empty := &BacklogRecord{}
	if empty.HasFindingForPairAnySource(probe) {
		t.Error("HasFindingForPairAnySource = true on an empty record")
	}
}

// TestJevPrecedence_HalfB_LaterFindingsLandAlongside — AC-JEVN-003 sub-cases
// 3 and 4 (REQ-JEVN-006 half (b)). This half is delivered by changing
// NOTHING: AppendFindingOnce's key includes Source, so a mechanical or agent
// finding never matches an existing Jev tuple. The assertion exists so that a
// future change to the dedup key is recognised as BREAKING this behaviour
// rather than as a refactor — and it asserts both that the arriving finding
// landed and that the Jev finding survived, because a wrongly-suppressed
// measurement leaves no trace.
func TestJevPrecedence_HalfB_LaterFindingsLandAlongside(t *testing.T) {
	for _, arriving := range []string{BacklogSourceMechanical, BacklogSourceAgent} {
		rec := &BacklogRecord{Findings: []BacklogFinding{
			jevFinding("t1", "t2", BacklogRelationNearDuplicate),
		}}
		f := BacklogFinding{
			SubjectID: "t1", RelatedID: "t2",
			Relation: BacklogRelationNearDuplicate, Source: arriving, Score: 0.91,
		}
		if !rec.AppendFindingOnce(f) {
			t.Errorf("arriving %s finding reported not appended — half (b) requires it to land alongside", arriving)
		}
		if len(rec.Findings) != 2 {
			t.Fatalf("findings after arriving %s = %d, want 2: %+v", arriving, len(rec.Findings), rec.Findings)
		}
		var sawJev, sawArriving bool
		for _, got := range rec.Findings {
			switch got.Source {
			case BacklogSourceJev:
				sawJev = true
			case arriving:
				sawArriving = true
			}
		}
		if !sawJev {
			t.Errorf("the pre-existing Jev finding did not survive an arriving %s finding", arriving)
		}
		if !sawArriving {
			t.Errorf("the arriving %s finding is absent from the list", arriving)
		}
	}
}
