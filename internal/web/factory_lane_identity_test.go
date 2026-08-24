package web

import (
	"strings"
	"testing"
)

// TestLaneRowKeepsRegistryLaneAsIdentity pins REQ-WC15-047's premise on the one
// input the earlier code let through: a record whose own lane number disagrees
// with the registry label it was joined from.
//
// The record's number used to win, so a bad join could relabel a row — two lanes
// rendering the same number while the registered lane vanished from the page.
// That is the misattribution the lane section exists to keep off the screen, so
// a disagreement is resolved as unresolved rather than as a winner.
func TestLaneRowKeepsRegistryLaneAsIdentity(t *testing.T) {
	const sessionID = "s-lane-clash"

	sessions := map[string]SessionVM{sessionID: {ID: sessionID, PID: 4242, State: StateLive}}
	records := []KanbanRecord{{SessionID: sessionID, Role: "lane", Lane: 3, CardID: "t999", SpecID: "SPEC-X-001"}}

	row := resolveLane(
		5, 4242,
		map[int]int{4242: 1},
		map[int][]string{4242: {sessionID}},
		sessions,
		map[string]KanbanRecord{sessionID: records[0]},
	)

	if row.Lane != 5 {
		t.Errorf("row.Lane = %d, want 5 — the registry label is the row's identity", row.Lane)
	}
	if !row.Unresolved {
		t.Error("a record naming a different lane must render unresolved, not present")
	}
	if row.UnresolvedReason != LaneUnresolvedLaneClash {
		t.Errorf("UnresolvedReason = %q, want %q", row.UnresolvedReason, LaneUnresolvedLaneClash)
	}
	if row.CardID != "" || row.SpecID != "" {
		t.Errorf("an unresolved row carries no join values, got card=%q spec=%q", row.CardID, row.SpecID)
	}
}

// TestChainIsAbsentWhenOnlyFactoryLanesHaveRecords covers a supported
// configuration the chain builder answered wrongly: a project running factory
// lanes and no chain.
//
// buildChain reads "any record exists" as proof the chain is present but renders
// only ChainRoles, so a lane's record made such a project show a present chain
// stopped at an idle `lead` — a confident wrong answer, which is the failure
// class this SPEC was written against.
func TestChainIsAbsentWhenOnlyFactoryLanesHaveRecords(t *testing.T) {
	records := []KanbanRecord{
		{SessionID: "s1", Role: "lane", Lane: 1},
		{SessionID: "s2", Role: "lane", Lane: 2},
	}

	kept := chainRoleRecords(records)
	if len(kept) != 0 {
		t.Fatalf("chainRoleRecords kept %d lane record(s); the chain takes chain roles only", len(kept))
	}

	chain := buildChain(t.TempDir(), kept, map[string]SessionVM{}, "")
	if chain.Present {
		t.Error("a lanes-only project has no chain; Present must be false")
	}

	// And the mixed case still finds the chain it does have — carried through
	// buildChain, not stopped at the filter. A correct filter feeding a consumer
	// that then reads it wrongly is exactly the shape this test exists to catch,
	// so asserting only chainRoleRecords' output would leave that half unproven.
	mixed := append(records, KanbanRecord{SessionID: "s3", Role: "run"})
	keptMixed := chainRoleRecords(mixed)
	if len(keptMixed) != 1 || strings.ToLower(keptMixed[0].Role) != "run" {
		t.Fatalf("chainRoleRecords(mixed) = %+v, want exactly the run record", keptMixed)
	}

	mixedChain := buildChain(t.TempDir(), keptMixed,
		map[string]SessionVM{"s3": {ID: "s3", State: StateLive}}, "")
	if !mixedChain.Present {
		t.Error("a project with one chain-role record HAS a chain; Present must be true")
	}
	var runRole *RoleVM
	for i := range mixedChain.Roles {
		if mixedChain.Roles[i].Role == "run" {
			runRole = &mixedChain.Roles[i]
		}
	}
	if runRole == nil {
		t.Fatalf("the chain does not carry a run role: %+v", mixedChain.Roles)
	}
	if runRole.Session != "s3" {
		t.Errorf("run role session = %q, want the record's session s3", runRole.Session)
	}
	if mixedChain.IdleRole == "run" {
		t.Error("run has a record and a live session; it must not be reported idle")
	}
}
