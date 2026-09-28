package web

// role_naming_m3_legacy_leader_test.go — SPEC-ROLE-NAMING-CODE-001 M3,
// AC-RNC-013 web clause: a declared chain whose persisted leader role is
// `leader` renders the leader slot present with the label `leader`; one whose
// persisted leader role is the legacy `lead` reports NO present leader and
// its role label reads exactly "legacy run: relaunch required".

import "testing"

func TestLeaderRecordRendersPresentLeaderSlot(t *testing.T) {
	records := []KanbanRecord{{SessionID: "s1", Role: "leader"}}
	kept := chainRoleRecords(records)
	if len(kept) != 1 {
		t.Fatalf("chainRoleRecords kept %d records, want the leader record", len(kept))
	}
	chain := buildChain(t.TempDir(), kept, map[string]SessionVM{}, "")
	if !chain.Present {
		t.Fatal("a leader record proves a chain is present; Present = false")
	}
	if len(chain.Roles) == 0 || chain.Roles[0].Role != "leader" {
		t.Fatalf("leader slot role label = %v, want the first slot labelled leader", chain.Roles)
	}
	if chain.Roles[0].State == StateIdle {
		t.Errorf("leader slot State = %q, want a non-idle state for the present leader", chain.Roles[0].State)
	}
}

func TestLegacyLeadRecordRendersRelaunchLabel(t *testing.T) {
	records := []KanbanRecord{{SessionID: "s1", Role: "lead"}}
	kept := chainRoleRecords(records)
	if len(kept) != 1 {
		t.Fatalf("chainRoleRecords kept %d records, want the legacy lead record (leader evidence)", len(kept))
	}
	chain := buildChain(t.TempDir(), kept, map[string]SessionVM{}, "")
	if !chain.Present {
		t.Fatal("a legacy lead record proves a run exists; Present = false")
	}
	if len(chain.Roles) == 0 {
		t.Fatal("chain renders no role slots")
	}
	leader := chain.Roles[0]
	if leader.Role != "legacy run: relaunch required" {
		t.Errorf("leader slot role label = %q, want exactly %q", leader.Role, "legacy run: relaunch required")
	}
	if leader.State != StateIdle {
		t.Errorf("leader slot State = %q, want idle — a legacy run presents no present leader", leader.State)
	}
	if leader.Session != "" {
		t.Errorf("leader slot Session = %q, want empty — the legacy record is not adopted as the leader", leader.Session)
	}
	// IdleRole stays the plain chain role name: it feeds the attention line's
	// sentence, which names the role, not the relaunch label.
	if chain.IdleRole != "leader" {
		t.Errorf("IdleRole = %q, want %q", chain.IdleRole, "leader")
	}
}

// TestLegacyLeadRecordKeepsLaneRecordsOutOfTheChain re-checks the factory-lane
// exclusion the M1 test pins: widening the filter for the legacy leader must
// not re-admit lane records.
func TestLegacyLeadRecordKeepsLaneRecordsOutOfTheChain(t *testing.T) {
	records := []KanbanRecord{
		{SessionID: "s1", Role: "lead"},
		{SessionID: "s2", Role: "lane", Lane: 1},
	}
	kept := chainRoleRecords(records)
	if len(kept) != 1 || kept[0].Role != "lead" {
		t.Fatalf("chainRoleRecords = %+v, want exactly the legacy lead record", kept)
	}
}
